/*
 * Copyright (c) 2019 Anapaya Systems
 */
/*
 * SCION plugin node registration and initialization header
 */
#ifndef __SCION_H__
#define __SCION_H__

#include <vppinfra/bihash_8_8.h>
#include <vppinfra/bihash_24_8.h>
#include <vnet/ip/ip.h>
#include "intf.h"

#define VPP_VER "18.10"

typedef clib_bihash_kv_8_8_t scion_intf_key4_t;
typedef clib_bihash_kv_24_8_t scion_intf_key6_t;

typedef struct {
    /* vector of intf instances */
    scion_intf_t *intfs;

    /* lookup intf by key (key in network byte order) */
    clib_bihash_8_8_t scion4_intf_by_key;       /* keyed on ipv4.dst + port */
    clib_bihash_24_8_t scion6_intf_by_key;      /* keyed on ipv6.dst + port */

    /* Mapping from sw_if_index to intf index */
    u32 *intf_index_by_sw_if_index;

    /* mapping from ifid to intf index. 0 is not an allowed ifid in the map */
    uword *intf_index_by_ifid;
    /* interface index of the IPv4 internal interface */
    u32 int4_intf_index;
    /* interface index of the IPv6 internal interface */
    u32 int6_intf_index;

    /* graph node state */
    uword *bm_ip4_scion_enabled_by_sw_if;
    uword *bm_ip6_scion_enabled_by_sw_if;

    /* convenience */
    vlib_main_t *vlib_main;
    vnet_main_t *vnet_main;
} scion_main_t;

extern scion_main_t scion_main;

u8 *format_scion_as (u8 * s, va_list * args);

uword unformat_scion_as (unformat_input_t * input, va_list * args);

u8 *format_scion_isdas (u8 * s, va_list * args);

uword unformat_scion_isdas (unformat_input_t * input, va_list * args);

always_inline void
scion_key4_pack (scion_intf_key4_t * key4, ip4_address_t addr, u16 port)
{
    key4->key = (((u64) addr.as_u32) << 32) | port;
    key4->value = ~0ULL;
}

always_inline void
scion_key6_pack (scion_intf_key6_t * key6, ip6_address_t * addr, u16 port)
{
    key6->key[0] = addr->as_u64[0];
    key6->key[1] = addr->as_u64[1];
    key6->key[2] = port;
    key6->value = ~0ULL;
}

#endif /* __SCION_H__ */
