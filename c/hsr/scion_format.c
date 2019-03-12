/*
 * Copyright (c) 2019 Anapaya Systems
 */
/*
 * format/unformat functions regarding SCION headers, fields, types, etc.
 */
#include <vnet/vnet.h>
#include "scion.h"
#include "scion_packet.h"

static u8 *
format_unix_time (u8 * s, va_list * args)
{
    const u32 len = 32;
    time_t ts = va_arg (*args, time_t);
    struct tm t = { 0 };
    char ts_str[len];

    memset (ts_str, 0, len);
    if (gmtime_r (&ts, &t) == NULL) {
        return s;
    }
    /* time format: 2006-01-02 15:04:05-0700 */
    if (strftime (ts_str, len, "%F %T%z", &t)) {
        s = format (s, "%s", ts_str);
    } else {
        s = format (s, "secs since epoch: %u", ts);
    }
    return s;
}

u8 *
format_scion_as (u8 * s, va_list * args)
{
    scion_as_t as = va_arg (*args, scion_as_t);

    if (as >> 32) {
        s = format (s, "%x:%x:%x", (as >> 32) & 0xffff, (as >> 16) & 0xffff, (as >> 0) & 0xffff);
    } else {
        s = format (s, "%u", as);
    }

    return s;
}

uword
unformat_scion_as (unformat_input_t * input, va_list * args)
{
    scion_as_t *as = va_arg (*args, scion_as_t *);
    u16 a, b, c;
    u8 rv = 0;

    if (unformat (input, "%X:%X:%X", sizeof (a), &a, sizeof (b), &b, sizeof (c), &c)) {
        *as = ((scion_as_t) a) << 32 | ((scion_as_t) b) << 16 | c;
        rv = 1;
    } else if (unformat (input, "%u", as)) {
        rv = !(*as >> 32);
    }

    return rv;
}

u8 *
format_scion_isdas (u8 * s, va_list * args)
{
    scion_isdas_t isdas = va_arg (*args, scion_isdas_t);

    s = format (s, "%u-%U", isd_from_isdas (isdas), format_scion_as, as_from_isdas (isdas));

    return s;
}

uword
unformat_scion_isdas (unformat_input_t * input, va_list * args)
{
    scion_isdas_t *isdas = va_arg (*args, scion_isdas_t *);
    u32 isd;
    scion_as_t as;
    uword rv = 0;

    if (unformat (input, "%u-%U", &isd, unformat_scion_as, &as)) {
        *isdas = isdas_from_isd_as ((scion_isd_t) isd, as);
        rv = 1;
    }

    return rv;
}

static u8 *
format_scion_svc_address (u8 * s, va_list * args)
{
    scion_svc_t svc = (scion_svc_t) va_arg (*args, u32);
    char *tag = "A";

    if (svc & SCION_SVC_MULTICAST) {
        tag = "M";
    }
    switch (svc & ~SCION_SVC_MULTICAST) {
/* *INDENT-OFF* */
#define _(type, str) \
    case SCION_SVC_TYPE_##type: \
        s = format (s, "%s_%c", str, tag); \
        break;
    foreach_scion_svc_addr
#undef _
/* *INDENT-ON* */
    default:
        s = format (s, "Unknown(%u)", svc);
    }
    return s;
}

static u8 *
format_scion_addr (u8 * s, va_list * args)
{
    u8 *addr = va_arg (*args, u8 *);
    scion_addr_t type = va_arg (*args, scion_addr_t);

    switch (type) {
    case SCION_ADDR_TYPE_IPV4:
        s = format (s, "%U", format_ip4_address, addr);
        break;
    case SCION_ADDR_TYPE_IPV6:
        s = format (s, "%U", format_ip6_address, addr);
        break;
    case SCION_ADDR_TYPE_SVC:
        s = format (s, "%U", format_scion_svc_address, ((scion_svc_t *) addr)[0]);
        break;
    default:
        s = format (s, "Unknown(%u)", type);
    }
    return s;
}

static u8 *
format_scion_path_hopf (u8 * s, va_list * args)
{
    scion_hopf_t *hopf_p = va_arg (*args, scion_hopf_t *);
    scion_hopf_t hopf = clib_net_to_host_u64 (hopf_p[0]);
    u8 flags = hopf_flags (hopf);

    s = format (s, "HOP flags: ");
/* *INDENT-OFF* */
#define _(bit, name, v) \
    if (flags & SCION_HOPF_##name) { \
        s = format (s, "%s, ", v); \
    }
    foreach_scion_hopf_flag
#undef _
/* *INDENT-ON* */
    if (!flags) {
        s = format (s, "none, ");
    }
    s = format (s, "ExpTime %u, ConsIn %u, ConsEg %u, Mac %x", hopf_exp_time (hopf),
                hopf_ingress (hopf, 1), hopf_egress (hopf, 1), hopf_mac (hopf));
    return s;
}

static u8 *
format_scion_path_infof (u8 * s, va_list * args)
{
    scion_infof_t *infof = va_arg (*args, scion_infof_t *);

    s = format (s, "INFO flags: ");
/* *INDENT-OFF* */
#define _(bit, name, v) \
    if (infof->flags & SCION_INFOF_##name) { \
        s = format (s, "%s, ", v); \
    }
    foreach_scion_infof_flag
#undef _
/* *INDENT-ON* */
    if (!infof->flags) {
        s = format (s, "none, ");
    }
    s = format (s, "isd %u, hops %u, %U", clib_net_to_host_u16 (infof->isd), infof->hops,
                format_unix_time, clib_net_to_host_u32 (infof->timestamp));
    return s;
}

static u8 *
format_scion_path_segment (u8 * s, va_list * args)
{
    scion_infof_t *infof = va_arg (*args, scion_infof_t *);
    scion_hopf_t *hopf;
    u32 len = va_arg (*args, u32);
    u32 indent = format_get_indent (s);
    indent += 2;

    u32 seg_len = clib_min (infof->hops + 1, len);
    if (seg_len) {
        s = format (s, "%U", format_scion_path_infof, infof);
        hopf = (scion_hopf_t *) infof;
        for (u32 hop_offset = 1; hop_offset < seg_len; hop_offset += 1) {
            s = format (s, "\n%U%U", format_white_space, indent,
                        format_scion_path_hopf, hopf + hop_offset);
        }
    }
    return s;
}

static u8 *
format_scion_path (u8 * s, va_list * args)
{
    u8 *path = va_arg (*args, u8 *);
    u32 len = va_arg (*args, u32);
    scion_infof_t *infof;
    u32 indent = format_get_indent (s);

    len = len / SCION_PATH_LEN;

    infof = (scion_infof_t *) path;
    if (len) {
        s = format (s, "%U", format_scion_path_segment, infof, len);
        for (u32 offset = 1 + infof->hops; offset < len; offset += 1 + infof->hops) {
            s = format (s, "\n%U%U", format_white_space, indent,
                        format_scion_path_segment, infof + offset, len - offset);
        }
    }
    return s;
}

u8 *
format_scion_header (u8 * s, va_list * args)
{
    scion_fixed_hdr_t *scion = va_arg (*args, scion_fixed_hdr_t *);
    u32 len = va_arg (*args, u32);
    u32 path_offset = sizeof (scion[0]) + scion_addr_len (scion);

    scion_isdas_t dst_isdas = clib_net_to_host_u64 (scion->dst_isdas);
    scion_isdas_t src_isdas = clib_net_to_host_u64 (scion->src_isdas);

    scion_addr_t dst_type = scion_dst_addr_type (scion);
    scion_addr_t src_type = scion_src_addr_type (scion);

    u8 *dst_addr = (u8 *) (scion + 1);
    u8 *src_addr = dst_addr + scion_addr_type_len (dst_type);

    i32 path_len = len - path_offset;
    u32 indent = format_get_indent (s);
    indent += 2;

    ASSERT (len >= sizeof (scion[0]));

    s = format (s, "%U: %U,[%U] -> %U,[%U]", format_ip_protocol, scion->next_header,
                format_scion_isdas, src_isdas, format_scion_addr, src_addr, src_type,
                format_scion_isdas, dst_isdas, format_scion_addr, dst_addr, dst_type);
    s = format (s, "\n%Uversion %u, total-len %uB, header-len %u",
                format_white_space, indent, scion_version (scion),
                clib_net_to_host_u16 (scion->total_len), scion->header_len);
    s = format (s, "\n%Ucurrent-info %u, current-hop %u",
                format_white_space, indent, scion->current_infof, scion->current_hopf);
    if (path_len > 0) {
        s = format (s, "\n%U%U", format_white_space, indent - 2,
                    format_scion_path, ((u8 *) scion) + path_offset, path_len);
    }
    return s;
}
