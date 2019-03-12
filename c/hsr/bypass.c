/*
 * Copyright (c) 2019 Anapaya Systems
 */
/*
 * The SCION bypass node checks whether an IP/UDP packet matches a SCION interface, otherwise
 * it keeps processing the packet. The node is implemented as a feature and can therefore be
 * enabled/disabled on runtime per interface.
 * This file includes the node implementation, registration, trace support and cli commands.
 */
#include <vlib/vlib.h>
#include <vppinfra/string.h>
#include <vnet/ip/ip.h>
#include <vnet/udp/udp.h>
#include <vnet/feature/feature.h>
#include <vnet/plugin/plugin.h>
#include <vnet/vnet.h>
#include "scion.h"
#include "scion_packet.h"

vlib_node_registration_t ip4_scion_bypass_node;
vlib_node_registration_t ip6_scion_bypass_node;

#define foreach_scion_bypass_next \
_(DROP, "error-drop") \
_(SCION_INPUT, "scion-input")

typedef enum {
#define _(v, n) SCION_BYPASS_NEXT_##v,
    foreach_scion_bypass_next
#undef _
    SCION_BYPASS_NEXT_N
} ip_scion_bypass_next_t;

#define foreach_scion_error \
_(NONE, "no error") \
_(NO_INTF_MATCH, "no interface match") \
_(MIN_LENGTH, "packet length < minimum length (IPv4/IPv6 + UDP + SCION)") \
_(IP_HEADER, "bad ip header") \
_(UDP_CHECKSUM, "bad udp checksum") \
_(UDP_LENGTH, "bad udp length")

typedef enum {
#define _(f,s) SCION_BYPASS_ERROR_##f,
    foreach_scion_error
#undef _
    SCION_BYPASS_ERROR_N
} scion_error_t;

static char *scion_error_strings[] = {
#define _(n,s) s,
    foreach_scion_error
#undef _
};

/* *INDENT-OFF* */
VNET_FEATURE_INIT (ip4_scion_bypass, static) = {
  .arc_name = "ip4-unicast",
  .node_name = "ip4-scion-bypass",
  .runs_before = VNET_FEATURES ("ip4-not-enable"),
};

VNET_FEATURE_INIT (ip6_scion_bypass, static) = {
  .arc_name = "ip6-unicast",
  .node_name = "ip6-scion-bypass",
  .runs_before = VNET_FEATURES ("ip6-not-enable"),
};
/* *INDENT-ON* */

typedef struct {
    udp_header_t udp;
    u32 next_index;
    u32 intf_index;
    u32 error;
    union {
        ip4_header_t ip4;
        ip6_header_t ip6;
    };
} scn_udp_trace_t;

static u8 *
format_scn_udp_trace (u8 * s, va_list * args, u32 is_ip4)
{
    CLIB_UNUSED (vlib_main_t * vm) = va_arg (*args, vlib_main_t *);
    CLIB_UNUSED (vlib_node_t * node) = va_arg (*args, vlib_node_t *);
    scn_udp_trace_t *t = va_arg (*args, scn_udp_trace_t *);

    s = format (s, "intf-index %u, next %u, error %u", t->intf_index, t->next_index, t->error);
    if (is_ip4) {
        s = format (s, "\n  %U", format_ip4_header, &t->ip4, sizeof (t->ip4));
    } else {
        s = format (s, "\n  %U", format_ip6_header, &t->ip6, sizeof (t->ip6));
    }
    s = format (s, "\n  %U", format_udp_header, &t->udp, sizeof (t->udp));
    return (s);
}

static u8 *
format_scn_udp4_trace (u8 * s, va_list * args)
{
    return format_scn_udp_trace (s, args, 1);
}

static u8 *
format_scn_udp6_trace (u8 * s, va_list * args)
{
    return format_scn_udp_trace (s, args, 0);
}

static_always_inline void
scion_bypass_trace (vlib_main_t * vm, vlib_node_runtime_t * node, vlib_buffer_t * b,
                    u32 error, u32 next, u32 intf_index, u32 is_ip4)
{
    if (PREDICT_FALSE (b->flags & VLIB_BUFFER_IS_TRACED)) {
        scn_udp_trace_t *tr = vlib_add_trace (vm, node, b, sizeof (*tr));
        u8 *ip, *udp;
        tr->next_index = next;
        tr->error = error;
        tr->intf_index = intf_index;
        if (error != SCION_BYPASS_ERROR_NONE) {
            ip = vlib_buffer_get_current (b);
            udp = vlib_buffer_get_current (b)
                    + (is_ip4 ? sizeof (ip4_header_t) : sizeof (ip6_header_t));
        } else {
            ip = vlib_buffer_get_current (b) - sizeof (udp_header_t)
                    - (is_ip4 ? sizeof (ip4_header_t) : sizeof (ip6_header_t));
            udp = vlib_buffer_get_current (b) - sizeof (udp_header_t);
        }
        if (is_ip4) {
            clib_memcpy (&tr->ip4, ip, sizeof (ip4_header_t));
        } else {
            clib_memcpy (&tr->ip6, ip, sizeof (ip6_header_t));
        }
        clib_memcpy (&tr->udp, udp, sizeof (udp_header_t));
    }
}

/**
 * Check that the packet length is at least the minimum expected length for SCION packets,
 * which would be IPv4/IPv6 + UDP + SCION (common + dst/src ISDAS).
 *
 * @param b packet buffer
 * @param is_ip4 current header is IPv4
 *
 * @return 1 if the packet length is less than the minimum expected length.
 *         0 if the packet length is at least the minimum expected length.
 */
static_always_inline u8
scion_overlay_check_min_size (vlib_buffer_t * b, u32 is_ip4)
{
    u32 min_len = sizeof (udp_header_t) + sizeof (scion_fixed_hdr_t);

    if (is_ip4) {
        min_len += sizeof (ip4_header_t);
    } else {
        min_len += sizeof (ip6_header_t);
    }

    return (b->current_length < min_len);
}

/**
 * Check that the next protocol after IP4/IP6 is UDP.
 *
 * @param b packet buffer
 * @param is_ip4 current header is IPv4
 *
 * @return 1 if next protocol is not UDP
 *         0 if next protocol is UDP
 */
static_always_inline u8
scion_overlay_check_udp (vlib_buffer_t * b, u32 is_ip4)
{
    u8 proto;

    if (is_ip4) {
        ip4_header_t *ip4 = vlib_buffer_get_current (b);
        proto = ip4->protocol;
    } else {
        ip6_header_t *ip6 = vlib_buffer_get_current (b);
        proto = ip6->protocol;
    }

    /* IPv6 extensions are not supported */
    return (proto != IP_PROTOCOL_UDP);
}

/**
 * Check UDP length is valid.
 * IPv4/IPv6 length have been checked in their respective input nodes.
 * IPv4 packets with options are considered errors by the ip4-input node.
 *
 * @param b packet buffer
 * @param is_ip4 current header is IPv4
 *
 * @return 1 if the IPv4/IPv6 length does not match UDP length
 *         0 if IPv4/IPv6 and UDP lengths are valid
 */
static_always_inline u8
scion_overlay_check_ip_udp_len (vlib_buffer_t * b, u32 is_ip4)
{
    udp_header_t *udp;
    u16 ip_len, udp_len;

    if (is_ip4) {
        ip4_header_t *ip4 = vlib_buffer_get_current (b);
        udp = vlib_buffer_get_current (b) + sizeof (ip4_header_t);
        ip_len = clib_net_to_host_u16 (ip4->length) - sizeof (ip4_header_t);
        udp_len = clib_net_to_host_u16 (udp->length);
    } else {
        ip6_header_t *ip6 = vlib_buffer_get_current (b);
        udp = vlib_buffer_get_current (b) + sizeof (ip6_header_t);
        ip_len = clib_net_to_host_u16 (ip6->payload_length);
        udp_len = clib_net_to_host_u16 (udp->length);
    }

    return udp_len != ip_len;
}

/**
 * Validate UDP checksum
 *
 * @param vm vlib main
 * @param b packet buffer
 * @param is_ip4 current header is IPv4
 *
 * @return 1 if the checksum is invalid
 *         0 if the checksum is valid
 */
static_always_inline u8
scion_overlay_validate_udp_csum (vlib_main_t * vm, vlib_buffer_t * b, u32 is_ip4)
{
    u32 flags = b->flags;

    if ((flags & VNET_BUFFER_F_L4_CHECKSUM_COMPUTED) == 0) {
        /* Verify UDP checksum */
        if (is_ip4) {
            flags = ip4_tcp_udp_validate_checksum (vm, b);
        } else {
            flags = ip6_tcp_udp_icmp_validate_checksum (vm, b);
        }
    }

    return (flags & VNET_BUFFER_F_L4_CHECKSUM_CORRECT) == 0;
}

/**
 * Match the destination address/port of a packet to an internal/external SCION interface.
 *
 * @param b Packet buffer
 * @param is_ip4 Whether the packet is an IPv4 packet
 *
 * @return Interface index on success
 *         MAX u32 if not matched
 */
static_always_inline u32
scion_match_intf_by_overlay (vlib_buffer_t * b, u32 is_ip4)
{
    scion_main_t *scm = &scion_main;
    clib_bihash_kv_8_8_t key4;
    clib_bihash_kv_24_8_t key6;
    udp_header_t *udp;
    u8 not_found;

    if (is_ip4) {
        ip4_header_t *ip4 = vlib_buffer_get_current (b);
        /* Assumes no IPv4 options */
        udp = vlib_buffer_get_current (b) + sizeof (ip4_header_t);
        scion_key4_pack (&key4, ip4->dst_address, udp->dst_port);
        not_found = clib_bihash_search_inline_8_8 (&scm->scion4_intf_by_key, &key4);
    } else {
        ip6_header_t *ip6 = vlib_buffer_get_current (b);
        /* Assumes no IPv6 extensions */
        udp = vlib_buffer_get_current (b) + sizeof (ip6_header_t);
        scion_key6_pack (&key6, &ip6->dst_address, udp->dst_port);
        not_found = clib_bihash_search_inline_24_8 (&scm->scion6_intf_by_key, &key6);
    }
    if (not_found) {
        /* no interface match */
        return ~0;
    }
    return is_ip4 ? (u32) key4.value : (u32) key6.value;
}

/**
 * Check that the source address/port of a packet matches the remote address/port of a
 * SCION interface. For internal interfaces, no checks are done.
 *
 * @param b Packet buffer
 * @param intf_index interface index
 * @param is_ip4 Whether the packet is an IPv4 packet
 *
 * @return 1 if the interface index is MAX u32, meaning invalid interface index.
 *         0 if it was received in an internal interface or if it matches the source
 *           address/port of the interface
 */
static_always_inline u8
scion_overlay_validate_src (vlib_buffer_t * b, u32 intf_index, u32 is_ip4)
{
    scion_main_t *scm = &scion_main;
    scion_intf_t *intf;
    udp_header_t *udp;
    u8 ip_compare;

    if (intf_index == ~0) {
        return 1;
    }
    intf = pool_elt_at_index (scm->intfs, intf_index);

    if (intf->ifid == 0) {
        /* Internal interface, no source address/port check required */
        return 0;
    }

    if (is_ip4) {
        ip4_header_t *ip4 = vlib_buffer_get_current (b);
        udp = vlib_buffer_get_current (b) + sizeof (ip4_header_t);
        ip_compare = ip4_address_compare (&ip4->src_address, &intf->remote.ip4);
    } else {
        ip6_header_t *ip6 = vlib_buffer_get_current (b);
        udp = vlib_buffer_get_current (b) + sizeof (ip6_header_t);
        ip_compare = ip6_address_compare (&ip6->src_address, &intf->remote.ip6);
    }

    /* XXX We could store the remote_port in network order to avoid swap */
    return ip_compare || udp->src_port != clib_host_to_net_u16 (intf->remote_port);
}

/**
 * This function returns an error value depending on the multiple different error checks.
 *
 * @param ip_err IP error
 * @param udp_err UDP error
 * @param csum_err checksum error
 * @param match_err interface match error
 *
 * @return error value
 */
static_always_inline u8
scion_bypass_err_code (u8 ip_err, u8 udp_err, u8 csum_err, u8 validate_err)
{
    u8 error = SCION_BYPASS_ERROR_NONE;

    if (validate_err) {
        error = SCION_BYPASS_ERROR_NO_INTF_MATCH;
    }
    if (csum_err) {
        error = SCION_BYPASS_ERROR_UDP_CHECKSUM;
    }
    if (udp_err) {
        error = SCION_BYPASS_ERROR_UDP_LENGTH;
    }
    if (ip_err) {
        error = SCION_BYPASS_ERROR_IP_HEADER;
    }

    return error;
}

/**
 * Packets arrive at this node after they have passed the ip4/ip6-input nodes,
 * so it is expected that the IP header values are valid.
 *
 * Currently VPP considers packets with IPv4 options as errors and will be dropped.
 *
 */
static_always_inline uword
ip_scion_bypass_inline (vlib_main_t * vm, vlib_node_runtime_t * node,
                        vlib_frame_t * frame, u32 is_ip4)
{
    scion_main_t *scm = &scion_main;
    u32 *from, *to_next, n_left_from, n_left_to_next, next_index;

    from = vlib_frame_vector_args (frame);
    n_left_from = frame->n_vectors;
    next_index = node->cached_next_index;

    while (n_left_from > 0) {
        vlib_get_next_frame (vm, node, next_index, to_next, n_left_to_next);

        while (n_left_from > 0 && n_left_to_next > 0) {
            vlib_buffer_t *b0;
            u32 bi0, next0;
            u32 intf_index0 = ~0;
            u8 ip_err0 = 0, udp_err0 = 0, csum_err0 = 0, validate_err0 = 0;
            u8 error0 = SCION_BYPASS_ERROR_NONE;

            bi0 = to_next[0] = from[0];
            from += 1;
            n_left_from -= 1;
            to_next += 1;
            n_left_to_next -= 1;
            next0 = SCION_BYPASS_NEXT_SCION_INPUT;

            b0 = vlib_get_buffer (vm, bi0);
            if (PREDICT_FALSE (scion_overlay_check_min_size (b0, is_ip4))) {
                next0 = SCION_BYPASS_NEXT_DROP;
                error0 = SCION_BYPASS_ERROR_MIN_LENGTH;
                b0->error = node->errors[error0];
                goto trace;
            }

            ip_err0 = scion_overlay_check_udp (b0, is_ip4);
            udp_err0 = scion_overlay_check_ip_udp_len (b0, is_ip4);
            csum_err0 = scion_overlay_validate_udp_csum (vm, b0, is_ip4);
            intf_index0 = scion_match_intf_by_overlay (b0, is_ip4);
            validate_err0 = scion_overlay_validate_src (b0, intf_index0, is_ip4);

            if (PREDICT_FALSE (ip_err0 || udp_err0 || csum_err0 || validate_err0)) {
                next0 = SCION_BYPASS_NEXT_DROP;
                error0 = scion_bypass_err_code (ip_err0, udp_err0, csum_err0, validate_err0);
                b0->error = node->errors[error0];
                goto trace;
            }

            scion_intf_t *intf0 = pool_elt_at_index (scm->intfs, intf_index0);
            vnet_buffer (b0)->sw_if_index[VLIB_RX] = intf0->sw_if_index;

            /* scion-input node expect current at SCION header */
            if (is_ip4) {
                vlib_buffer_advance (b0, sizeof (ip4_header_t) + sizeof (udp_header_t));
            } else {
                vlib_buffer_advance (b0, sizeof (ip6_header_t) + sizeof (udp_header_t));
            }
 trace:
            scion_bypass_trace (vm, node, b0, error0, next0, intf_index0, is_ip4);

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

    .n_errors = SCION_BYPASS_ERROR_N,
    .error_strings = scion_error_strings,

    .n_next_nodes = SCION_BYPASS_NEXT_N,
    .next_nodes = {
#define _(v, n) [SCION_BYPASS_NEXT_##v] = n,
        foreach_scion_bypass_next
#undef _
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

    .n_errors = SCION_BYPASS_ERROR_N,
    .error_strings = scion_error_strings,

    .n_next_nodes = SCION_BYPASS_NEXT_N,
    .next_nodes = {
#define _(v, n) [SCION_BYPASS_NEXT_##v] = n,
        foreach_scion_bypass_next
#undef _
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

void
vnet_int_scion_bypass_mode (u32 sw_if_index, u8 is_ip6, u8 is_enable)
{
    // XXX This is a very dumb enable/disable without keeping track of interface state, etc.
    is_enable = ! !is_enable;

    if (is_ip6) {
        vnet_feature_enable_disable ("ip6-unicast", "ip6-scion-bypass",
                                     sw_if_index, is_enable, 0, 0);
    } else {
        vnet_feature_enable_disable ("ip4-unicast", "ip4-scion-bypass",
                                     sw_if_index, is_enable, 0, 0);
    }
}

static clib_error_t *
set_ip_scion_bypass (u32 is_ip6, unformat_input_t * input, vlib_cli_command_t * cmd)
{
    unformat_input_t _line_input, *line_input = &_line_input;
    vnet_main_t *vnm = vnet_get_main ();
    clib_error_t *error = 0;
    u32 sw_if_index, is_enable;

    sw_if_index = ~0;
    is_enable = 1;

    if (!unformat_user (input, unformat_line_input, line_input)) {
        return 0;
    }

    while (unformat_check_input (line_input) != UNFORMAT_END_OF_INPUT) {
        if (unformat_user (line_input, unformat_vnet_sw_interface, vnm, &sw_if_index)) {
            ;
        } else if (unformat (line_input, "del")) {
            is_enable = 0;
        } else {
            error = unformat_parse_error (line_input);
            break;
        }
    }
    unformat_free (line_input);
    if (error) {
        return error;
    }

    if (sw_if_index == ~0) {
        return clib_error_return (0, "unknown interface `%U'", format_unformat_error, line_input);
    }

    vlib_cli_output (vlib_get_main (), "sw_if_index: %u, is_enable: %u\n", sw_if_index, is_enable);

    vnet_int_scion_bypass_mode (sw_if_index, is_ip6, is_enable);

    return 0;
}

static clib_error_t *
set_ip4_scion_bypass (vlib_main_t * vm, unformat_input_t * input, vlib_cli_command_t * cmd)
{
    return set_ip_scion_bypass (0, input, cmd);
}

/*?
 * This command adds the 'ip4-scion-bypass' graph node for a given interface.
 * By adding the IPv4 scion-bypass graph node to an interface, the node checks
 *  for and validates input scion packets and bypasses ip4-lookup, ip4-local,
 * and ip4-udp-lookup nodes to speedup scion packet forwarding. This node will
 * cause extra overhead for non-scion packets which is kept at a minimum.
 *
 * @cliexpar
 * @parblock
 * Example of graph node before ip4-scion-bypass is enabled:
 * @cliexstart{show vlib graph ip4-scion-bypass}
 *            Name                      Next                    Previous
 * ip4-scion-bypass                error-drop [0]
 *                                scion4-input [1]
 *                                 ip4-lookup [2]
 * @cliexend
 *
 * Example of how to enable ip4-scion-bypass on an interface:
 * @cliexcmd{set interface ip scion-bypass GigabitEthernet2/0/0}
 *
 * Example of graph node after ip4-scion-bypass is enabled:
 * @cliexstart{show vlib graph ip4-scion-bypass}
 *            Name                      Next                    Previous
 * ip4-scion-bypass                error-drop [0]               ip4-input
 *                                scion4-input [1]        ip4-input-no-checksum
 *                                 ip4-lookup [2]
 * @cliexend
 *
 * Example of how to display the feature enabled on an interface:
 * @cliexstart{show ip interface features GigabitEthernet2/0/0}
 * IP feature paths configured on GigabitEthernet2/0/0...
 * ...
 * ipv4 unicast:
 *   ip4-scion-bypass
 *   ip4-lookup
 * ...
 * @cliexend
 *
 * Example of how to disable ip4-scion-bypass on an interface:
 * @cliexcmd{set interface ip scion-bypass GigabitEthernet2/0/0 del}
 * @endparblock
?*/
/* *INDENT-OFF* */
VLIB_CLI_COMMAND (set_interface_ip_scion_bypass_command, static) = {
  .path = "set interface ip scion-bypass",
  .function = set_ip4_scion_bypass,
  .short_help = "set interface ip scion-bypass <interface> [del]",
};
/* *INDENT-ON* */

static clib_error_t *
set_ip6_scion_bypass (vlib_main_t * vm, unformat_input_t * input, vlib_cli_command_t * cmd)
{
    return set_ip_scion_bypass (1, input, cmd);
}

/*?
 * This command adds the 'ip6-scion-bypass' graph node for a given interface.
 * By adding the IPv6 scion-bypass graph node to an interface, the node checks
 *  for and validate input scion packet and bypass ip6-lookup, ip6-local,
 * ip6-udp-lookup nodes to speedup scion packet forwarding. This node will
 * cause extra overhead to for non-scion packets which is kept at a minimum.
 *
 * @cliexpar
 * @parblock
 * Example of graph node before ip6-scion-bypass is enabled:
 * @cliexstart{show vlib graph ip6-scion-bypass}
 *            Name                      Next                    Previous
 * ip6-scion-bypass                error-drop [0]
 *                                scion6-input [1]
 *                                 ip6-lookup [2]
 * @cliexend
 *
 * Example of how to enable ip6-scion-bypass on an interface:
 * @cliexcmd{set interface ip6 scion-bypass GigabitEthernet2/0/0}
 *
 * Example of graph node after ip6-scion-bypass is enabled:
 * @cliexstart{show vlib graph ip6-scion-bypass}
 *            Name                      Next                    Previous
 * ip6-scion-bypass                error-drop [0]               ip6-input
 *                                scion6-input [1]        ip4-input-no-checksum
 *                                 ip6-lookup [2]
 * @cliexend
 *
 * Example of how to display the feature enabled on an interface:
 * @cliexstart{show ip interface features GigabitEthernet2/0/0}
 * IP feature paths configured on GigabitEthernet2/0/0...
 * ...
 * ipv6 unicast:
 *   ip6-scion-bypass
 *   ip6-lookup
 * ...
 * @cliexend
 *
 * Example of how to disable ip6-scion-bypass on an interface:
 * @cliexcmd{set interface ip6 scion-bypass GigabitEthernet2/0/0 del}
 * @endparblock
?*/
/* *INDENT-OFF* */
VLIB_CLI_COMMAND (set_interface_ip6_scion_bypass_command, static) = {
  .path = "set interface ip6 scion-bypass",
  .function = set_ip6_scion_bypass,
  .short_help = "set interface ip scion-bypass <interface> [del]",
};
/* *INDENT-ON* */
