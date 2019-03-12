/*
 * Copyright (c) 2019 Anapaya Systems
 */
/*
 * Common SCION header checks.
 */
#include <time.h>
#include <openssl/cmac.h>
#include <vnet/feature/feature.h>
#include <vnet/plugin/plugin.h>
#include <vnet/vnet.h>
#include "scion.h"
#include "scion_packet.h"
#include "scion_error.h"

#define SCION_MAX_TRACE 512

#define foreach_scion_input_next \
_(DROP, "error-drop") \
_(EXT_INPUT, "scion-ext-input") \
_(PATH_UPDATE, "scion-path-update") \
_(IP4_UDP_INT, "scion-ip4-udp-int") \
_(IP6_UDP_INT, "scion-ip6-udp-int")

typedef enum {
#define _(v, n) SCION_INPUT_NEXT_##v,
    foreach_scion_input_next
#undef _
    SCION_INPUT_NEXT_N
} scion_input_next_t;

typedef struct {
    u32 next_index;
    u32 error;
    u32 data_len;
    u8 data[SCION_MAX_TRACE];
} scion_trace_t;

static u8 *
format_scion_trace (u8 * s, va_list * args)
{
    CLIB_UNUSED (vlib_main_t * vm) = va_arg (*args, vlib_main_t *);
    CLIB_UNUSED (vlib_node_t * node) = va_arg (*args, vlib_node_t *);
    scion_trace_t *t = va_arg (*args, scion_trace_t *);
    scion_fixed_hdr_t *scion = (scion_fixed_hdr_t *) t->data;
    u32 data_len = t->data_len;
    u32 indent = format_get_indent (s);

    s = format (s, "next %u, error %u", t->next_index, t->error);
    if (sizeof (scion[0]) <= data_len) {
        s = format (s, "\n%U%U", format_white_space, indent, format_scion_header, scion, data_len);
    } else {
        s = format (s, "\n%U%U", format_white_space, indent, format_hex_bytes, scion, data_len);
    }

    return s;
}

static_always_inline void
scion_input_trace (vlib_main_t * vm, vlib_node_runtime_t * node, vlib_buffer_t * b,
                   u32 error, u32 next)
{
    if (PREDICT_FALSE (b->flags & VLIB_BUFFER_IS_TRACED)) {
        scion_trace_t *tr = vlib_add_trace (vm, node, b, sizeof (*tr));
        scion_fixed_hdr_t *scion0 = vlib_buffer_get_current (b);
        u32 data_len = clib_min (b->current_length, SCION_MAX_TRACE);
        tr->next_index = next;
        tr->error = error;

        if (sizeof (scion0[0]) <= data_len) {
            data_len = clib_min (data_len, scion0->header_len * SCION_LINE_LEN);
        }
        ASSERT (data_len <= SCION_MAX_TRACE);
        tr->data_len = data_len;
        clib_memcpy (tr->data, scion0, data_len);
    }
}

/* One day in seconds */
#define MAX_TTL (24 * 60 * 60)
/* Expired time unit ~5m38s */
#define EXP_TIME_UNIT (MAX_TTL / 256)

/**
 * Returns the hop field to mac against depending on the info field construction
 * direction flag.
 *
 * @param infof current info field
 * @param hopf current hop field
 *
 * @return hop field to mac against
 *         NULL if the current hop field was the first/last (depending on direction)
 */
static_always_inline scion_hopf_t *
scion_mac_hopf (scion_infof_t * infof, scion_hopf_t * hopf)
{
    scion_hopf_t *h = infof->flags & SCION_INFOF_CONS_DIR ? hopf - 1 : hopf + 1;
    scion_hopf_t *low = (scion_hopf_t *) infof;
    scion_hopf_t *up = low + infof->hops;

    /* TODO Segment change, cross-over and peer */

    ASSERT (low < hopf);
    ASSERT (hopf <= up);

    /* validate the mac hopf is a valid hop field in the segment */
    h = low < h ? h : 0;
    h = h <= up ? h : 0;

    return h;
}

typedef union {
    u8 as_u8[16];
    u32 as_u32[4];
    u64 as_u64[2];
} scion_pad16_t;

/**
 * Create the 16 Bytes message to authenticate with the format below.
 *
 * If the mac_hopf is NULL, it means that the current hop is the first/last hop of the
 * current segment. In such a case, PrevHopF is zeroed.
 *
 *   0                   1                   2                   3
 *   0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
 *  +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
 *  |                           Timestamp                           |
 *  +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
 *  |       0       |    ExpTime    |      ConsIngress      |  ...  |
 *  +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
 *  | ...ConsEgress |                                               |
 *  +-+-+-+-+-+-+-+-+                                               +
 *  |                           PrevHopF                            |
 *  +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
 */
static_always_inline void
scion_mac_data (scion_pad16_t * data, scion_hopf_t * hopf, scion_hopf_t * mac_hopf, u32 ts)
{
    memset (data->as_u8, 0, sizeof (data[0]));
    data->as_u32[0] = ts;
    /* set ExpTime, ConsIngres and  ConsEgress */
    memcpy (&data->as_u8[5], &((u8 *) hopf)[1], 4);
    if (mac_hopf) {
        /* set PrevHopF,  */
        memcpy (&data->as_u8[9], &((u8 *) mac_hopf)[1], 7);
    }
}

/**
 * Validate the hop field mac.
 *
 * @param vm vlib main
 * @param infof current info field
 * @param hopf current hop field
 * @param error existing error
 *
 * @return SCION_ERROR_HOPF_BAD_MAC if the mac validation failed,
 *           existing error otherwise
 */
static_always_inline scion_error_t
scion_validate_mac (vlib_main_t * vm, scion_infof_t * infof, scion_hopf_t * hopf,
                    scion_error_t error)
{
    scion_main_t *scm = &scion_main;
    scion_hopf_t *mac_hopf;
    scion_pad16_t data;
    scion_pad16_t mac;
    size_t maclen;
    int ret;
    u32 mactrunc;

    mac_hopf = scion_mac_hopf (infof, hopf);

    scion_mac_data (&data, hopf, mac_hopf, infof->timestamp);

    /* TODO support two keys based on flags bit, not supported in BR yet.
     * https://github.com/scionproto/scion/issues/1714
     */
    sym_key_t *key = &scm->sym_keys[0];

    CMAC_CTX *ctx = scm->per_thread_data[vm->thread_index].cmac_ctx;
    ret = CMAC_Init (ctx, key->val, key->len, 0, 0);
    ASSERT (ret == 1);
    ret = CMAC_Update (ctx, &data, sizeof (data));
    ASSERT (ret == 1);
    ret = CMAC_Final (ctx, &mac.as_u8[0], &maclen);
    ASSERT (ret == 1);
    /* compare to MAC's 3 LSB */
    mactrunc = mac.as_u32[3] & 0xffffff;
    if (mactrunc != hopf_mac (clib_net_to_host_u64 (hopf[0]))) {
        error = SCION_ERROR_HOPF_BAD_MAC;
    }

    return error;
}

/**
 * Check that the current hop field is valid.
 * This function assumes that the SCION common header is valid.
 *
 * @param vm vlib main
 * @param error_node error node
 * @param now current time in seconds
 * @param b packet buffer
 * @param scion scion header within the buffer
 * @param[out] next next node index
 *
 * @return SCION_ERROR_NONE on sucess
 *           specific error otherwise
 */
static_always_inline scion_error_t
scion_input_check_hopf (vlib_main_t * vm, vlib_node_runtime_t * error_node, const u32 now,
                        vlib_buffer_t * b, scion_fixed_hdr_t * scion, u32 * next)
{
    scion_main_t *scm = &scion_main;
    scion_infof_t *infof;
    scion_hopf_t *hopf_p, hopf;
    scion_intf_t *intf;
    u32 sw_if_index, intf_index;
    scion_error_t error;

    error = SCION_ERROR_NONE;
    sw_if_index = vnet_buffer (b)->sw_if_index[VLIB_RX];
    intf_index = vec_elt (scm->intf_index_by_sw_if_index, sw_if_index);
    intf = pool_elt_at_index (scm->intfs, intf_index);

    ASSERT (scion->current_infof * SCION_LINE_LEN < b->current_length);
    ASSERT (scion->current_hopf * SCION_LINE_LEN < b->current_length);
    infof = vlib_buffer_get_current (b) + (scion->current_infof * SCION_LINE_LEN);
    hopf_p = vlib_buffer_get_current (b) + (scion->current_hopf * SCION_LINE_LEN);
    hopf = clib_net_to_host_u64 (hopf_p[0]);

    /* check current hop is within the segment hops */
    if (scion->current_hopf > scion->current_infof + infof->hops) {
        error = SCION_ERROR_HOPF_NOT_IN_SEGMENT;
    }

    /* check expired time */
    u32 exp_time = clib_net_to_host_u32 (infof->timestamp);
    exp_time += (hopf_exp_time (hopf) + 1) * EXP_TIME_UNIT;
    if (exp_time < now) {
        error = SCION_ERROR_HOPF_EXPIRED;
    }
    /* TODO allow for some forward clock drift. https://github.com/scionproto/scion/issues/1980 */

    /* check ingress ifid for traffic from external interfaces */
    if (intf->ifid && hopf_ingress (hopf, infof->flags & SCION_INFOF_CONS_DIR) != intf->ifid) {
        error = SCION_ERROR_HOPF_BAD_INGRESS_INTF;
    }

    error = scion_validate_mac (vm, infof, hopf_p, error);

    if (PREDICT_FALSE (error != SCION_ERROR_NONE)) {
        /* TODO SCMP reply? */
        b->error = error_node->errors[error];
        *next = SCION_INPUT_NEXT_DROP;
    }
    return error;
}

/**
 * Check that the SCION header is valid.
 *
 * @param vm vlib main
 * @param error_node error node
 * @param now current time in seconds
 * @param b packet buffer
 * @param scion scion header within the buffer
 * @param[out] next next node index
 *
 * @return SCION_ERROR_NONE on sucess
 *           specific error otherwise
 */
static_always_inline scion_error_t
scion_input_check (vlib_main_t * vm, vlib_node_runtime_t * error_node, const u32 now,
                   vlib_buffer_t * b, scion_fixed_hdr_t * scion, u32 * next)
{
    scion_error_t error = SCION_ERROR_NONE;

    u32 common_addr_len = (sizeof (scion[0]) + scion_addr_len (scion)) / SCION_LINE_LEN;

    /* check current infof */
    error = common_addr_len <= scion->current_infof ? error : SCION_ERROR_BAD_CURRENT_INFOF;

    /* check current infof and hopf */
    error = scion->current_infof < scion->current_hopf ? error : SCION_ERROR_BAD_CURRENT_INFOF_HOPF;

    /* check current_hopf */
    error = scion->current_hopf < scion->header_len ? error : SCION_ERROR_BAD_CURRENT_HOPF;

    /* check path length for minimum possible path segment INFO+HOP+HOP (24 Bytes) */
    error = scion->header_len - common_addr_len >= 3 ? error : SCION_ERROR_BAD_PATH;

    /* check scion header length */
    u16 total_len = clib_net_to_host_u16 (scion->total_len);
    error = scion->header_len * SCION_LINE_LEN < total_len ? error : SCION_ERROR_BAD_HEADER_LENGTH;

    /* check scion total length */
    error = b->current_length == total_len ? error : SCION_ERROR_BAD_LENGTH;

    /* check version */
    error = scion_version (scion) == SCION_VERSION ? error : SCION_ERROR_VERSION;

    /* length must be at least minimal SCION header. */
    error = sizeof (scion[0]) < b->current_length ? error : SCION_ERROR_TOO_SHORT;

    if (PREDICT_FALSE (error != SCION_ERROR_NONE)) {
        /* TODO SCMP reply? */
        b->error = error_node->errors[error];
        *next = SCION_INPUT_NEXT_DROP;
        return error;
    }

    return scion_input_check_hopf (vm, error_node, now, b, scion, next);
}

/**
 * Returns the Unix time in seconds.
 *
 * @param c time object
 *
 * @return Unix time in seconds
 */
static_always_inline u32
scion_time_now_secs (clib_time_t * c)
{
    /* XXX Ideally have the master thread calling the vlib_time_now() from time to time to do
     * full time verification/update in case of variant TSC, and worker threads just use the
     * latest values.
     */
    /* clib_time_t keeps track of Unix time and TSC values, where last_verify_reference_time is
     * the Unix Time float value corresponding to the last_cpu_time cycles values */
    return ((u32) c->last_verify_reference_time)
            + ((u32) ((clib_cpu_time_now () - c->last_cpu_time) / c->clocks_per_second));
}

/**
 * Return the next node for a packet.
 *
 * @param b packet buffer
 * @param scion scion header within the buffer
 *
 * @return next node index
 */
static_always_inline u32
scion_set_next (vlib_buffer_t * b, scion_fixed_hdr_t * scion)
{
    scion_main_t *scm = &scion_main;
    u32 next = SCION_INPUT_NEXT_DROP;

    u16 dst_addr_type = scion_dst_addr_type (scion);

    if (scion->next_header == SCION_HBH_EXT) {
        next = SCION_INPUT_NEXT_EXT_INPUT;
    } else if (scion->dst_isdas != scm->isdas) {
        next = SCION_INPUT_NEXT_PATH_UPDATE;
    } else if (dst_addr_type == SCION_ADDR_TYPE_IPV4) {
        next = SCION_INPUT_NEXT_IP4_UDP_INT;
    } else if (dst_addr_type == SCION_ADDR_TYPE_IPV6) {
        next = SCION_INPUT_NEXT_IP6_UDP_INT;
    } else if (dst_addr_type == SCION_ADDR_TYPE_SVC) {
        /* TODO hsr support or punt packet */
    }

    return next;
}

/**
 * Common checks of the SCION headers.
 * By common checks we mean the all the checks that need to be performed on a packet
 * regardless of where they are being forwarded to.
 */
static_always_inline uword
scion_input_node_inline (vlib_main_t * vm, vlib_node_runtime_t * node, vlib_frame_t * frame)
{
    u32 *from, *to_next, n_left_from, n_left_to_next, next_index;
    const u32 now = scion_time_now_secs (&vm->clib_time);

    from = vlib_frame_vector_args (frame);
    n_left_from = frame->n_vectors;
    next_index = node->cached_next_index;

    while (n_left_from > 0) {
        vlib_get_next_frame (vm, node, next_index, to_next, n_left_to_next);

        while (n_left_from > 0 && n_left_to_next > 0) {
            vlib_buffer_t *b0;
            scion_fixed_hdr_t *scion0;
            u32 bi0, next0, error0;

            bi0 = to_next[0] = from[0];
            from += 1;
            n_left_from -= 1;
            to_next += 1;
            n_left_to_next -= 1;

            b0 = vlib_get_buffer (vm, bi0);
            scion0 = vlib_buffer_get_current (b0);

            next0 = scion_set_next (b0, scion0);
            error0 = scion_input_check (vm, node, now, b0, scion0, &next0);

            scion_input_trace (vm, node, b0, error0, next0);

            vlib_validate_buffer_enqueue_x1 (vm, node, next_index, to_next, n_left_to_next,
                                             bi0, next0);
        }

        vlib_put_next_frame (vm, node, next_index, n_left_to_next);
    }

    return frame->n_vectors;
}

/* *INDENT-OFF* */
VLIB_NODE_FN (scion_input_node) (vlib_main_t * vm,
                                 vlib_node_runtime_t * node,
                                 vlib_frame_t * frame) {
    return scion_input_node_inline (vm, node, frame);
}

VLIB_REGISTER_NODE (scion_input_node) =
{
    .name = "scion-input",
    .vector_size = sizeof (u32),

    .n_errors = SCION_ERROR_N,
    .error_strings = scion_error_strings,

    .n_next_nodes = SCION_INPUT_NEXT_N,
    .next_nodes = {
#define _(v, n) [SCION_INPUT_NEXT_##v] = n,
        foreach_scion_input_next
#undef _
    },
    .format_buffer = format_scion_header,
    .format_trace = format_scion_trace,
};
/* *INDENT-ON* */

/* Dummy init function to get us linked in. */
static clib_error_t *
scion_input_init (vlib_main_t * vm)
{
    return 0;
}

VLIB_INIT_FUNCTION (scion_input_init);

/*
 *
 * XXX This is plumbing until each node is implemented
 *
 */
static_always_inline uword
temp_inline (vlib_main_t * vm, vlib_node_runtime_t * node, vlib_frame_t * frame)
{
    u32 *from, *to_next, n_left_from, n_left_to_next, next_index;

    from = vlib_frame_vector_args (frame);
    n_left_from = frame->n_vectors;
    next_index = node->cached_next_index;

    while (n_left_from > 0) {
        vlib_get_next_frame (vm, node, next_index, to_next, n_left_to_next);

        while (n_left_from > 0 && n_left_to_next > 0) {
            u32 bi0, next0 = 0;

            bi0 = to_next[0] = from[0];
            from += 1;
            n_left_from -= 1;
            to_next += 1;
            n_left_to_next -= 1;

            vlib_validate_buffer_enqueue_x1 (vm, node, next_index,
                                             to_next, n_left_to_next, bi0, next0);
        }

        vlib_put_next_frame (vm, node, next_index, n_left_to_next);
    }

    return frame->n_vectors;
}

/* *INDENT-OFF* */
VLIB_NODE_FN (temp_node0) (vlib_main_t * vm, vlib_node_runtime_t * node, vlib_frame_t * frame)
{
    return temp_inline (vm, node, frame);
}
VLIB_REGISTER_NODE (temp_node0) =
{
    .name = "scion-ext-input",
    .vector_size = sizeof (u32),

    .n_next_nodes = 1,
    .next_nodes = {
        [0] = "error-drop",
    },
};

VLIB_NODE_FN (temp_node1) (vlib_main_t * vm, vlib_node_runtime_t * node, vlib_frame_t * frame)
{
    return temp_inline (vm, node, frame);
}
VLIB_REGISTER_NODE (temp_node1) =
{
    .name = "scion-path-update",
    .vector_size = sizeof (u32),

    .n_next_nodes = 1,
    .next_nodes = {
        [0] = "error-drop",
    },
};

VLIB_NODE_FN (temp_node2) (vlib_main_t * vm, vlib_node_runtime_t * node, vlib_frame_t * frame)
{
    return temp_inline (vm, node, frame);
}
VLIB_REGISTER_NODE (temp_node2) =
{
    .name = "scion-ip4-udp-int",
    .vector_size = sizeof (u32),

    .n_next_nodes = 1,
    .next_nodes = {
        [0] = "error-drop",
    },
};

VLIB_NODE_FN (temp_node3) (vlib_main_t * vm, vlib_node_runtime_t * node, vlib_frame_t * frame)
{
    return temp_inline (vm, node, frame);
}
VLIB_REGISTER_NODE (temp_node3) =
{
    .name = "scion-ip6-udp-int",
    .vector_size = sizeof (u32),

    .n_next_nodes = 1,
    .next_nodes = {
        [0] = "error-drop",
    },
};
/* *INDENT-ON* */
