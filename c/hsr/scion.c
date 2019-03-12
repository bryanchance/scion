/*
 * Copyright (c) 2019 Anapaya Systems
 */
/*
 * SCION plugin node registration and initialization
 */
#include <openssl/cmac.h>
#include <vnet/feature/feature.h>
#include <vnet/plugin/plugin.h>
#include <vnet/vnet.h>
#include "scion.h"
#include "scion_packet.h"

scion_main_t scion_main;

#define SCION_HASH_NUM_BUCKETS (2 * 1024)
#define SCION_HASH_MEMORY_SIZE (1 << 20)

int
scion_set_key (scion_set_key_args_t * a)
{
    scion_main_t *scm = &scion_main;

    if (a->index >= SCION_KEY_N) {
        return VNET_API_ERROR_INVALID_VALUE;
    }
    if (a->key.len > SCION_KEY_MAX_SIZE) {
        return VNET_API_ERROR_INVALID_VALUE_2;
    }

    memcpy (scm->sym_keys[a->index].val, a->key.val, a->key.len);
    scm->sym_keys[a->index].len = a->key.len;

    return 0;
}

static clib_error_t *
scion_set_key_command_fn (vlib_main_t * vm, unformat_input_t * input, vlib_cli_command_t * cmd)
{
    unformat_input_t _line_input, *line_input = &_line_input;
    clib_error_t *error = 0;
    u8 *key = 0;
    u32 key_index = ~0;

    /* Get a line of input. */
    if (!unformat_user (input, unformat_line_input, line_input)) {
        return 0;
    }

    while (unformat_check_input (line_input) != UNFORMAT_END_OF_INPUT) {
        if (unformat (line_input, "%u %U", &key_index, unformat_hex_string, &key)) {
            ;
        } else {
            error = unformat_parse_error (line_input);
            break;
        }
    }
    unformat_free (line_input);
    if (error) {
        return error;
    }

    if (vec_len (key) > SCION_KEY_MAX_SIZE) {
        return clib_error_return (0, "invalid key length");
    }

    scion_set_key_args_t a = {
        .index = key_index,
        .key.len = vec_len (key),
    };
    memcpy (&a.key.val, key, a.key.len);

    int rv = scion_set_key (&a);

    switch (rv) {
    case 0:
        /* success, nothing to print */
        break;
    case VNET_API_ERROR_INVALID_VALUE:
        return clib_error_return (0, "invalid index");

    case VNET_API_ERROR_INVALID_VALUE_2:
        return clib_error_return (0, "invalid key length");

    default:
        return clib_error_return (0, "scion_set_key returned %d", rv);
    }
    return 0;
}

/*?
 * Set SCION keys for Hop Field validation.
 ?*/
/* *INDENT-OFF* */
VLIB_CLI_COMMAND (set_scion_key_command, static) = {
  .path = "set scion key",
  .short_help = "set scion key 0 <key>",
  .function = scion_set_key_command_fn,
};
/* *INDENT-ON* */

static clib_error_t *
scion_show_keys (vlib_main_t * vm, unformat_input_t * input, vlib_cli_command_t * cmd)
{
    scion_main_t *scm = &scion_main;

    for (u32 i = 0; i < SCION_KEY_N; i += 1) {
        sym_key_t *k = &scm->sym_keys[i];
        vlib_cli_output (vm, "[%u] %U", i, format_hex_bytes, k->val, k->len);
    }
    return 0;
}

/*?
 * Show SCION keys for Hop Field validation.
 ?*/
/* *INDENT-OFF* */
VLIB_CLI_COMMAND (show_scion_key_command, static) = {
  .path = "show scion keys",
  .short_help = "show scion keys",
  .function = scion_show_keys,
};
/* *INDENT-ON* */

static_always_inline void
scion_per_thread_init (scion_main_t * scm)
{
    vlib_thread_main_t *tm = vlib_get_thread_main ();
    u32 thread_id;

    vec_validate_aligned (scm->per_thread_data, tm->n_vlib_mains - 1, CLIB_CACHE_LINE_BYTES);

    for (thread_id = 0; thread_id < tm->n_vlib_mains; thread_id++) {
        CMAC_CTX *ctx = CMAC_CTX_new ();
        CMAC_Init (ctx, NULL, 16, EVP_aes_128_cbc (), NULL);
        scm->per_thread_data[thread_id].cmac_ctx = ctx;
    }
}

clib_error_t *
scion_init (vlib_main_t * vm)
{
    scion_main_t *scm = &scion_main;

    memset (scm, 0, sizeof (scion_main_t));

    clib_bihash_init_8_8 (&scm->scion4_intf_by_key, "scion4",
                          SCION_HASH_NUM_BUCKETS, SCION_HASH_MEMORY_SIZE);
    clib_bihash_init_24_8 (&scm->scion6_intf_by_key, "scion6",
                           SCION_HASH_NUM_BUCKETS, SCION_HASH_MEMORY_SIZE);

    scm->bm_ip4_scion_enabled_by_sw_if = 0;
    scm->bm_ip6_scion_enabled_by_sw_if = 0;

    scm->int4_intf_index = ~0;
    scm->int6_intf_index = ~0;

    scion_per_thread_init (scm);

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
