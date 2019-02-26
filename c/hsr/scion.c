/*
 * Copyright (c) 2019 Anapaya Systems
 */
/*
 * SCION plugin node registration and initialization
 */
#include <vnet/feature/feature.h>
#include <vnet/plugin/plugin.h>
#include <vnet/vnet.h>
#include "scion.h"
#include "scion_packet.h"

scion_main_t scion_main;

#define SCION_HASH_NUM_BUCKETS (2 * 1024)
#define SCION_HASH_MEMORY_SIZE (1 << 20)

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

clib_error_t *
scion_init (vlib_main_t * vm)
{
    scion_main_t *scm = &scion_main;

    clib_bihash_init_8_8 (&scm->scion4_intf_by_key, "scion4",
                          SCION_HASH_NUM_BUCKETS, SCION_HASH_MEMORY_SIZE);
    clib_bihash_init_24_8 (&scm->scion6_intf_by_key, "scion6",
                           SCION_HASH_NUM_BUCKETS, SCION_HASH_MEMORY_SIZE);

    scm->bm_ip4_scion_enabled_by_sw_if = 0;
    scm->bm_ip6_scion_enabled_by_sw_if = 0;

    scm->int4_intf_index = ~0;
    scm->int6_intf_index = ~0;

    scm->vnet_main = vnet_get_main ();
    scm->vlib_main = vm;

    return 0;
}

VLIB_INIT_FUNCTION (scion_init);

/* *INDENT-OFF* */
VLIB_PLUGIN_REGISTER () = {
    .version = VPP_VER,
    .description = "SCION",
};
/* *INDENT-ON* */
