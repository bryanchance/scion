/*
 * Copyright (c) 2019 Anapaya Systems
 */
/*
 * SCION plugin node registration and initialization header
 */
#ifndef __SCION_H__
#define __SCION_H__

#include <openssl/cmac.h>
#include <vppinfra/bihash_8_8.h>
#include <vppinfra/bihash_24_8.h>
#include <vnet/ip/ip.h>
#include "intf.h"

#define VPP_VER "18.10"

#define SCION_VERSION 0x0

#define SCION_KEY_N 1
#define SCION_KEY_MAX_SIZE 32

typedef struct {
    u32 len;
    u8 val[SCION_KEY_MAX_SIZE];
} sym_key_t;

typedef struct {
    u32 index;
    sym_key_t key;
} scion_set_key_args_t;

/* This struct contains SCION data owned by each thread, cache aligned so there is
 * no false sharing.
 */
typedef struct {
    CLIB_CACHE_LINE_ALIGN_MARK (cacheline0);
    CMAC_CTX *cmac_ctx;
} scion_per_thread_data_t;

/**
 * This structure contains the SCION configuration/context.
 *
 * It is used by multiple SCION nodes for different functionality, mostly regarding
 * SCION interfaces (internal and external).
 */
typedef struct {
    /* list (pool) of instances */
    scion_intf_t *intfs;

    /* XXX currently the remote port is used to match traffic to an interface.
     * One disadvantage is that with RSS all the traffic from an interface is going to be
     * processed by the same thread, given that it would be seen as the same flow.
     * To leverage RSS and distribute traffic from an interface over multiple threads,
     * the remote port could be randomized and should not be part of the interface setup.
     */
    /* mapping (bihash) intf index by IPv4 local/remote addr + udp port (network order) */
    clib_bihash_8_8_t scion4_intf_by_key;
    /* mapping (bihash) intf index by IPv6 local/remote addr + udp port (network order) */
    clib_bihash_24_8_t scion6_intf_by_key;

    /* mapping (vector) from sw_if_index to intf index */
    u32 *intf_index_by_sw_if_index;

    /* mapping (hash) from external ifid (network order) to intf index */
    uword *intf_index_by_ifid;
    /* IPv4 internal interface index */
    u32 int4_intf_index;
    /* IPv6 internal interface index */
    u32 int6_intf_index;

    /* graph node state */
    uword *bm_ip4_scion_enabled_by_sw_if;
    uword *bm_ip6_scion_enabled_by_sw_if;

    scion_isdas_t isdas;

    sym_key_t sym_keys[SCION_KEY_N];

    /* SCION specific data for each particular thread */
    scion_per_thread_data_t *per_thread_data;

    /* convenience */
    vlib_main_t *vlib_main;
    vnet_main_t *vnet_main;
} scion_main_t;

extern scion_main_t scion_main;

u8 *format_scion_as (u8 * s, va_list * args);

uword unformat_scion_as (unformat_input_t * input, va_list * args);

u8 *format_scion_isdas (u8 * s, va_list * args);

uword unformat_scion_isdas (unformat_input_t * input, va_list * args);

u8 *format_scion_header (u8 * s, va_list * args);

static_always_inline void
scion_key4_pack (clib_bihash_kv_8_8_t * key4, ip4_address_t addr, u16 port)
{
    key4->key = (((u64) addr.as_u32) << 32) | port;
    key4->value = ~0ULL;
}

static_always_inline void
scion_key6_pack (clib_bihash_kv_24_8_t * key6, ip6_address_t * addr, u16 port)
{
    key6->key[0] = addr->as_u64[0];
    key6->key[1] = addr->as_u64[1];
    key6->key[2] = port;
    key6->value = ~0ULL;
}

#endif /* __SCION_H__ */
