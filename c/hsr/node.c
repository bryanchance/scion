/*
 * Copyright (c) 2019 Anapaya Systems
 */
/*
 * The SCION bypass node drops all the packets. The node is implemented as a feature and can therefore
 * be enabled/disabled on runtime per interface.
 * It includes node implementation and registration as well as packet trace support.
 */
#include <vlib/vlib.h>
#include <vppinfra/string.h>
#include <vnet/ip/ip.h>
#include <vnet/udp/udp.h>
#include "scion.h"

vlib_node_registration_t ip4_scion_bypass_node;
vlib_node_registration_t ip6_scion_bypass_node;

/* *INDENT-OFF* */
VNET_FEATURE_INIT (ip4_scion_bypass, static) = {
  .arc_name = "ip4-unicast",
  .node_name = "ip4-scion-bypass",
  .runs_before = VNET_FEATURES ("ip4-lookup"),
};

VNET_FEATURE_INIT (ip6_scion_bypass, static) = {
  .arc_name = "ip6-unicast",
  .node_name = "ip6-scion-bypass",
  .runs_before = VNET_FEATURES ("ip6-lookup"),
};
/* *INDENT-ON* */

typedef struct udp4_encap_trace_t_ {
    udp_header_t udp;
    ip4_header_t ip;
} scn_udp4_trace_t;

typedef struct udp6_encap_trace_t_ {
    udp_header_t udp;
    ip6_header_t ip;
} scn_udp6_trace_t;

static u8 *
format_scn_udp4_trace (u8 * s, va_list * args)
{
    CLIB_UNUSED (vlib_main_t * vm) = va_arg (*args, vlib_main_t *);
    CLIB_UNUSED (vlib_node_t * node) = va_arg (*args, vlib_node_t *);
    scn_udp4_trace_t *t;

    t = va_arg (*args, scn_udp4_trace_t *);

    s = format (s, "%U\n  %U",
                format_ip4_header, &t->ip, sizeof (t->ip),
                format_udp_header, &t->udp, sizeof (t->udp));
    return (s);
}

static u8 *
format_scn_udp6_trace (u8 * s, va_list * args)
{
    CLIB_UNUSED (vlib_main_t * vm) = va_arg (*args, vlib_main_t *);
    CLIB_UNUSED (vlib_node_t * node) = va_arg (*args, vlib_node_t *);
    scn_udp6_trace_t *t;

    t = va_arg (*args, scn_udp6_trace_t *);

    s = format (s, "%U\n  %U",
                format_ip6_header, &t->ip, sizeof (t->ip),
                format_udp_header, &t->udp, sizeof (t->udp));
    return (s);
}

typedef enum {
    IP_SCION_BYPASS_NEXT_DROP,
    IP_SCION_BYPASS_N_NEXT,
} ip_vxan_bypass_next_t;

always_inline uword
ip_scion_bypass_inline (vlib_main_t * vm, vlib_node_runtime_t * node,
                        vlib_frame_t * frame, u32 is_ip4)
{
    u32 *from, *to_next, n_left_from, n_left_to_next, next_index;

    from = vlib_frame_vector_args (frame);
    n_left_from = frame->n_vectors;
    next_index = node->cached_next_index;

    while (n_left_from > 0) {
        vlib_get_next_frame (vm, node, next_index, to_next, n_left_to_next);

        while (n_left_from > 0 && n_left_to_next > 0) {
            vlib_buffer_t *b0;
            ip4_header_t *ip40;
            ip6_header_t *ip60;
            udp_header_t *udp0;
            u32 bi0, next0;

            bi0 = to_next[0] = from[0];
            from += 1;
            n_left_from -= 1;
            to_next += 1;
            n_left_to_next -= 1;
            next0 = IP_SCION_BYPASS_NEXT_DROP;

            b0 = vlib_get_buffer (vm, bi0);
            // XXX currently no checks - assume ip/udp
            if (is_ip4) {
                ip40 = vlib_buffer_get_current (b0);
                udp0 = ip4_next_header (ip40);
            } else {
                ip60 = vlib_buffer_get_current (b0);
                udp0 = ip6_next_header (ip60);
            }

            if (PREDICT_FALSE (b0->flags & VLIB_BUFFER_IS_TRACED)) {
                if (is_ip4) {
                    scn_udp4_trace_t *tr = vlib_add_trace (vm, node, b0, sizeof (*tr));
                    clib_memcpy (&tr->udp, udp0, sizeof (udp_header_t));
                    clib_memcpy (&tr->ip, ip40, sizeof (ip4_header_t));
                } else {
                    scn_udp6_trace_t *tr = vlib_add_trace (vm, node, b0, sizeof (*tr));
                    clib_memcpy (&tr->udp, udp0, sizeof (udp_header_t));
                    clib_memcpy (&tr->ip, ip60, sizeof (ip6_header_t));
                }
            }

            vlib_validate_buffer_enqueue_x1 (vm, node, next_index,
                                             to_next, n_left_to_next, bi0, next0);
        }

        vlib_put_next_frame (vm, node, next_index, n_left_to_next);
    }

    return frame->n_vectors;
}

/* *INDENT-OFF* */
VLIB_NODE_FN (ip4_scion_bypass_node) (vlib_main_t * vm,
                                      vlib_node_runtime_t * node,
                                      vlib_frame_t * frame)
{
  return ip_scion_bypass_inline (vm, node, frame, /* is_ip4 */ 1);
}

VLIB_REGISTER_NODE (ip4_scion_bypass_node) =
{
  .name = "ip4-scion-bypass",
  .vector_size = sizeof (u32),
  .n_next_nodes = IP_SCION_BYPASS_N_NEXT,
  .next_nodes = {
      [IP_SCION_BYPASS_NEXT_DROP] = "error-drop",
  },
  .format_buffer = format_ip4_header,
  .format_trace = format_scn_udp4_trace,
};

/* *INDENT-ON* */

/* Dummy init function to get us linked in. */
static clib_error_t *
ip4_scion_bypass_init (vlib_main_t * vm)
{
    return 0;
}

VLIB_INIT_FUNCTION (ip4_scion_bypass_init);

/* *INDENT-OFF* */
VLIB_NODE_FN (ip6_scion_bypass_node) (vlib_main_t * vm,
                                      vlib_node_runtime_t * node,
                                      vlib_frame_t * frame)
{
  return ip_scion_bypass_inline (vm, node, frame, /* is_ip4 */ 0);
}

VLIB_REGISTER_NODE (ip6_scion_bypass_node) =
{
  .name = "ip6-scion-bypass",
  .vector_size = sizeof (u32),
  .n_next_nodes = IP_SCION_BYPASS_N_NEXT,
  .next_nodes = {
      [IP_SCION_BYPASS_NEXT_DROP] = "error-drop",
  },
  .format_buffer = format_ip6_header,
  .format_trace = format_scn_udp6_trace,
};

/* *INDENT-ON* */

/* Dummy init function to get us linked in. */
static clib_error_t *
ip6_scion_bypass_init (vlib_main_t * vm)
{
    return 0;
}

VLIB_INIT_FUNCTION (ip6_scion_bypass_init);
