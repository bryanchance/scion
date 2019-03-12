/*
 * Copyright (c) 2019 Anapaya Systems
 */
/*
 * SCION headers
 */
#ifndef __SCION_PACKET_H__
#define __SCION_PACKET_H__

#include <vnet/vnet.h>

#define SCION_HBH_EXT 0
#define SCION_LINE_LEN 8
#define SCION_PATH_LEN 8

typedef u16 scion_isd_t;
typedef u64 scion_as_t;
typedef u64 scion_isdas_t;

static_always_inline scion_isd_t
isd_from_isdas (scion_isdas_t isdas)
{
    return (isdas >> 48);
}

static_always_inline scion_as_t
as_from_isdas (scion_isdas_t isdas)
{
    return (isdas & ((((scion_as_t) 1) << 48) - 1));
}

static_always_inline scion_isdas_t
isdas_from_isd_as (scion_isd_t isd, scion_as_t as)
{
    return ((scion_isdas_t) isd) << 48 | (as & ((((scion_as_t) 1) << 48) - 1));
}

/*                      1                   2                   3
 *  0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
 * +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
 * |Version|  DstType  |  SrcType  |           TotalLen            |
 * +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
 * |     HdrLen    |    CurrINF    |     CurrHF    |    NextHdr    |
 * +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
 * |                           DstISDAS                            |
 * +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
 * |                           SrcISDAS                            |
 * +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
 */
typedef struct {
    /** packet Type of the packet (version, dstType, srcType) */
    u16 ver_dst_src;
    /** total Length of the packet */
    u16 total_len;
    /** header length that includes the path */
    u8 header_len;
    /** offset of current Info opaque field*/
    u8 current_infof;
    /** offset of current Hop opaque field*/
    u8 current_hopf;
    /** next header type, shared with IP protocol number*/
    u8 next_header;
    /* */
    scion_isdas_t dst_isdas;
    /* */
    scion_isdas_t src_isdas;
} __attribute__ ((packed)) scion_fixed_hdr_t;

/* Address types and lengths */
#define foreach_scion_addr_type_len \
_(NONE, 0)  \
_(IPV4, 4)  \
_(IPV6, 16) \
_(SVC, 2)

typedef enum {
#define _(type, len) SCION_ADDR_TYPE_##type,
    foreach_scion_addr_type_len
#undef _
    SCION_ADDR_TYPE_N
} scion_addr_t;

static const u16 SCION_ADDR_LEN[] = {
#define _(type, len) len,
    foreach_scion_addr_type_len
#undef _
};

#define SCION_ADDR_TYPE_BITS 6
#define SCION_ADDR_TYPE_MASK ((1 << SCION_ADDR_TYPE_BITS) - 1)

static_always_inline scion_addr_t
scion_version (scion_fixed_hdr_t * sh)
{
    return clib_net_to_host_u16 (sh->ver_dst_src) >> 2 * SCION_ADDR_TYPE_BITS;
}

static_always_inline scion_addr_t
scion_src_addr_type (scion_fixed_hdr_t * sh)
{
    return (clib_net_to_host_u16 (sh->ver_dst_src) >> SCION_ADDR_TYPE_BITS) & SCION_ADDR_TYPE_MASK;
}

static_always_inline scion_addr_t
scion_dst_addr_type (scion_fixed_hdr_t * sh)
{
    return clib_net_to_host_u16 (sh->ver_dst_src) & SCION_ADDR_TYPE_MASK;
}

static_always_inline u16
scion_addr_type_len (scion_addr_t addr_type)
{
    u16 len = 0;
    if (PREDICT_TRUE (addr_type < SCION_ADDR_TYPE_N)) {
        len = SCION_ADDR_LEN[addr_type];
    }
    return len;
}

/**
 * Return the SCION source and destination address length, based on their types.
 * This is only for the variable portion of the SCION address header, thus it does not
 * include the source and destination ISD-AS.
 *
 * @param sh scion header
 *
 * @return src+dst length padded to a multiple of SCION_LINE_LEN
 *         0 if any of the address types is not supported (including SCION_ADDR_TYPE_NONE)
 */
static_always_inline u16
scion_addr_len (scion_fixed_hdr_t * sh)
{
    u32 src_type = scion_src_addr_type (sh);
    u32 dst_type = scion_dst_addr_type (sh);
    u16 len = 0;

    if (PREDICT_TRUE (src_type < SCION_ADDR_TYPE_N && src_type != SCION_ADDR_TYPE_NONE
                      && dst_type < SCION_ADDR_TYPE_N && dst_type != SCION_ADDR_TYPE_NONE)) {
        len = SCION_ADDR_LEN[src_type] + SCION_ADDR_LEN[dst_type];
    }
    /* pad to line length */
    return (len + (SCION_LINE_LEN - 1)) & ~(SCION_LINE_LEN - 1);
}

/*
 * SCION Addresses - variable length
 *
 *                      1                   2                   3
 *  0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
 * +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
 * |          SrcHostAddr          |                               |
 * +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+                               +
 * |                          DstHostAddr                          |
 * +                               +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
 * |                               |          Padding              |
 * +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
 */

typedef u16 scion_svc_t;

/* SVC addresses */
#define foreach_scion_svc_addr \
_(BEACON, "BS") \
_(PATH_MGMT, "PS") \
_(CERT_MGMT, "CS") \
_(SIBRA, "SB") \
_(SIG, "SIG")

typedef enum {
#define _(type, str) SCION_SVC_TYPE_##type,
    foreach_scion_svc_addr
#undef _
    SCION_SVC_TYPE_N
} scion_svc_type_t;

#define SCION_SVC_MULTICAST ((scion_svc_t)(1 << 15))

/*
 * SCION Path Info Field
 *
 *                      1                   2                   3
 *  0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
 * +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
 * |r r r r r P S C|                   Timestamp       ...         |
 * +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
 * |      ...      |              ISD              |    SegLen     |
 * +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
 */
typedef struct {
    u8 flags;
    u32 timestamp; /** timestamp is the Unix time seconds value */
    u16 isd;
    u8 hops; /** Number of hops in this segment */
} __attribute__ ((packed)) scion_infof_t;

/* SVC addresses */
#define foreach_scion_infof_flag \
_(0, CONS_DIR, "cons-dir") \
_(1, SHORTCUT, "shortcut") \
_(2, PEER, "peer")

enum {
#define _(bit, name, v) SCION_INFOF_##name = (1 << (bit)),
    foreach_scion_infof_flag
#undef _
};

/*
 * SCION Path Hop Field
 *
 *                      1                   2                   3
 *  0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
 * +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
 * |r r r r r r V X|    ExpTime    |        InIF           |  …    |
 * +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
 * |   … EgIF      |                      MAC                      |
 * +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
 */
typedef u64 scion_hopf_t;

/* SVC addresses */
#define foreach_scion_hopf_flag \
_(0, XOVER, "xover") \
_(1, VERIFY_ONLY, "verify-only")

enum {
#define _(bit, name, v) SCION_HOPF_##name = (1 << (bit)),
    foreach_scion_hopf_flag
#undef _
};

static_always_inline u8
hopf_flags (scion_hopf_t h)
{
    return h >> 56;
}

static_always_inline u8
hopf_exp_time (scion_hopf_t h)
{
    return (h >> 48) & 0xff;
}

static_always_inline scion_ifid_t
hopf_consdir_ingress (scion_hopf_t h)
{
    return (h >> 36) & 0xfff;
}

static_always_inline scion_ifid_t
hopf_consdir_egress (scion_hopf_t h)
{
    return (h >> 24) & 0xfff;
}

static_always_inline scion_ifid_t
hopf_ingress (scion_hopf_t h, u32 consdir)
{
    if (consdir) {
        return hopf_consdir_ingress (h);
    }
    return hopf_consdir_egress (h);
}

static_always_inline scion_ifid_t
hopf_egress (scion_hopf_t h, u32 consdir)
{
    if (consdir) {
        return hopf_consdir_egress (h);
    }
    return hopf_consdir_ingress (h);
}

static_always_inline u32
hopf_mac (scion_hopf_t h)
{
    return h & 0xffffff;
}

#endif /* __SCION_PACKET_H__ */
