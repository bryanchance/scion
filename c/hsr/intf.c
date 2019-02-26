/*
 * Copyright (c) 2019 Anapaya Systems
 */
/*
 * SCION virtual interfaces
 */
#include <vppinfra/string.h>
#include <vnet/ip/ip.h>
#include "scion.h"
#include "intf.h"

static u8 *
format_scion_name (u8 * s, va_list * args)
{
    u32 intf_index = va_arg (*args, u32);
    scion_main_t *scm = &scion_main;
    scion_intf_t *t;

    if (intf_index == ~0) {
        return format (s, "<cached-unused>");
    }

    if (intf_index >= vec_len (scm->intfs)) {
        return format (s, "<improperly-referenced>");
    }

    t = pool_elt_at_index (scm->intfs, intf_index);

    if (t->ifid) {
        return format (s, "scion%d", t->ifid);
    }
    return format (s, "scion%d%s", t->ifid, ip46_address_is_ip4 (&t->local) ? "v4" : "v6");
}

/* TODO investigate effect of interface up/down on processin.
 * It is likely that the current behavior has no effect
 */
static clib_error_t *
scion_interface_admin_up_down (vnet_main_t * vnm, u32 hw_if_index, u32 flags)
{
    u32 hw_flags = (flags & VNET_SW_INTERFACE_FLAG_ADMIN_UP) ? VNET_HW_INTERFACE_FLAG_LINK_UP : 0;

    vnet_hw_interface_set_flags (vnm, hw_if_index, hw_flags);

    return /* no error */ 0;
}

/* *INDENT-OFF* */
VNET_DEVICE_CLASS (scion_device_class, static) = {
  .name = "SCION",
  .format_device_name = format_scion_name,
  .admin_up_down_function = scion_interface_admin_up_down,
};
/* *INDENT-ON* */

static u8 *
format_scion_header_with_length (u8 * s, va_list * args)
{
    u32 intf_index = va_arg (*args, u32);

    s = format (s, "unimplemented dev %u", intf_index);

    return s;
}

/* *INDENT-OFF* */
VNET_HW_INTERFACE_CLASS (scion_hw_class) = {
  .name = "SCION",
  .format_header = format_scion_header_with_length,
  .build_rewrite = default_build_rewrite,
};
/* *INDENT-ON* */

static u8 *
format_scion_linkto (u8 * s, va_list * args)
{
    scion_linkto_t l = va_arg (*args, scion_linkto_t);

    s = format (s, "%s", scion_linkto_str[l]);

    return s;
}

static uword
unformat_scion_linkto (unformat_input_t * input, va_list * args)
{
    scion_linkto_t *linkto = va_arg (*args, scion_linkto_t *);
    u8 *linkto_str = 0;
    uword rv = 0;

    if (unformat (input, "%s", &linkto_str)) {
#define _(sym) \
        if (!strcmp (scion_linkto_str[LINK_TO_##sym], (const char *)linkto_str)) { \
            *linkto = LINK_TO_##sym; \
            rv = 1; \
            goto done; \
        }
        foreach_scion_linkto;
#undef _
    }
 done:
    vec_free (linkto_str);
    return rv;
}

static u8 *
format_scion_intf (u8 * s, va_list * args)
{
    scion_main_t *scm = &scion_main;
    scion_intf_t *t = va_arg (*args, scion_intf_t *);
    u32 index = t - scm->intfs;

    s = format (s, "[%u] ifid %u local %U local_port %u sw-if-index %u",
                index, t->ifid, format_ip46_address, &t->local, IP46_TYPE_ANY,
                t->local_port, t->sw_if_index);
    if (t->ifid) {
        s = format (s, "\n    remote %U remote-port %u link-to %U isd-as %U",
                    format_ip46_address, &t->remote, IP46_TYPE_ANY, t->remote_port,
                    format_scion_linkto, t->linkto, format_scion_isdas, t->isdas);
    }
    return s;
}

always_inline u32
find_scion_intf_by_key (scion_main_t * scm, ip46_address_t * ip, u16 port,
                        scion_intf_key4_t * key4, scion_intf_key6_t * key6, u8 is_ip4)
{
    u32 found_by_key;

    /* Try to find interface by key local_ip + local_port */
    if (is_ip4) {
        scion_key4_pack (key4, ip->ip4, port);
        found_by_key = !clib_bihash_search_inline_8_8 (&scm->scion4_intf_by_key, key4);
    } else {
        ip6_address_t ip6;
        memcpy (&ip6, &ip->ip6, sizeof (ip6));
        scion_key6_pack (key6, &ip6, port);
        found_by_key = !clib_bihash_search_inline_24_8 (&scm->scion6_intf_by_key, key6);
    }

    return found_by_key;
}

always_inline u32
find_scion_intf_by_ifid (scion_main_t * scm, scion_ifid_t ifid, u8 is_ip4)
{
    u32 intf_index = ~0;

    if (ifid) {
        uword *v = hash_get (scm->intf_index_by_ifid, ifid);
        intf_index = v ? *v : intf_index;
    } else {
        /* ifid 0 is special for internal interface and can be IPv4 or IPv6 */
        intf_index = is_ip4 ? scm->int4_intf_index : scm->int6_intf_index;
    }

    return intf_index;
}

int
scion_add_intf (scion_add_intf_args_t * a, u32 * sw_if_indexp)
{
    scion_main_t *scm = &scion_main;
    vnet_main_t *vnm = scm->vnet_main;
    scion_intf_t *t = 0;
    scion_intf_key4_t key4;
    scion_intf_key6_t key6;
    u32 intf_index = ~0, sw_if_index;
    u32 is_ip4 = ip46_address_is_ip4 (&a->local);

    intf_index = find_scion_intf_by_ifid (scm, a->ifid, is_ip4);
    /* intf must not already exist */
    if (intf_index != ~0) {
        return VNET_API_ERROR_IF_ALREADY_EXISTS;
    }
    /* Use network byte order for the key */
    u16 local_port = clib_host_to_net_u16 (a->local_port);
    if (find_scion_intf_by_key (scm, &a->local, local_port, &key4, &key6, is_ip4)) {
        return VNET_API_ERROR_ADDRESS_IN_USE;
    }
    pool_get_aligned (scm->intfs, t, CLIB_CACHE_LINE_BYTES);
    clib_memset_u64 (t, 0, sizeof (t[0]) / CLIB_CACHE_LINE_BYTES);
    intf_index = t - scm->intfs;

    /* copy from arg structure */
    t->ifid = a->ifid;
    t->local = a->local;
    t->local_port = a->local_port;
    if (a->ifid != 0) {
        t->remote = a->remote;
        t->remote_port = a->remote_port;
        t->linkto = a->linkto;
        t->isdas = a->isdas;
    }

    t->hw_if_index = vnet_register_interface (vnm, scion_device_class.index, intf_index,
                                              scion_hw_class.index, intf_index);
    vnet_hw_interface_t *hi = vnet_get_hw_interface (vnm, t->hw_if_index);

    t->sw_if_index = sw_if_index = hi->sw_if_index;

    /* copy the key */
    int add_failed;
    if (is_ip4) {
        key4.value = (u64) intf_index;
        add_failed = clib_bihash_add_del_8_8 (&scm->scion4_intf_by_key, &key4, 1 /*add */ );
    } else {
        key6.value = (u64) intf_index;
        add_failed = clib_bihash_add_del_24_8 (&scm->scion6_intf_by_key, &key6, 1 /*add */ );
    }

    if (add_failed) {
        vnet_delete_hw_interface (vnm, t->hw_if_index);
        pool_put (scm->intfs, t);
        return VNET_API_ERROR_INVALID_REGISTRATION;
    }

    vec_validate_init_empty (scm->intf_index_by_sw_if_index, sw_if_index, ~0);
    scm->intf_index_by_sw_if_index[sw_if_index] = intf_index;

    if (t->ifid != 0) {
        hash_set (scm->intf_index_by_ifid, t->ifid, intf_index);
    } else if (is_ip4) {
        scm->int4_intf_index = intf_index;
    } else {
        scm->int6_intf_index = intf_index;
    }

    vnet_sw_interface_t *si = vnet_get_sw_interface (vnm, sw_if_index);
    si->flags &= ~VNET_SW_INTERFACE_FLAG_HIDDEN;
    vnet_sw_interface_set_flags (vnm, sw_if_index, VNET_SW_INTERFACE_FLAG_ADMIN_UP);

    if (sw_if_indexp) {
        *sw_if_indexp = sw_if_index;
    }

    return 0;
}

int
scion_del_intf (scion_del_intf_args_t * a)
{
    scion_main_t *scm = &scion_main;
    vnet_main_t *vnm = scm->vnet_main;
    scion_intf_t *t = 0;
    scion_intf_key4_t key4;
    scion_intf_key6_t key6;
    u32 intf_index = ~0, sw_if_index;
    u32 is_ip4 = ip46_address_is_ip4 (&a->local);

    intf_index = find_scion_intf_by_ifid (scm, a->ifid, is_ip4);
    /* deleting a intf: intf must exist */
    if (intf_index == ~0) {
        return VNET_API_ERROR_NO_SUCH_ENTRY;
    }
    /* Use network byte order for the key */
    u16 local_port = clib_host_to_net_u16 (a->local_port);
    if (!find_scion_intf_by_key (scm, &a->local, local_port, &key4, &key6, is_ip4)) {
        return VNET_API_ERROR_ADDRESS_NOT_IN_USE;
    }
    ASSERT (intf_index < vec_len (scm->intfs));

    t = pool_elt_at_index (scm->intfs, intf_index);

    ASSERT (intf_index == (ip46_address_is_ip4 (&t->local) ? key4.value : key6.value));

    sw_if_index = t->sw_if_index;
    vnet_sw_interface_set_flags (vnm, sw_if_index, 0 /* down */ );

    ASSERT (scm->intf_index_by_sw_if_index[sw_if_index] == intf_index);
    scm->intf_index_by_sw_if_index[sw_if_index] = ~0;

    if (is_ip4) {
        ASSERT ((u32) key4.value == intf_index);
        clib_bihash_add_del_8_8 (&scm->scion4_intf_by_key, &key4, 0 /*del */ );
    } else {
        ASSERT ((u32) key6.value == intf_index);
        clib_bihash_add_del_24_8 (&scm->scion6_intf_by_key, &key6, 0 /*del */ );
    }

    vnet_delete_hw_interface (vnm, t->hw_if_index);

    pool_put (scm->intfs, t);

    return 0;
}

static clib_error_t *
scion_add_intf_command_fn (vlib_main_t * vm, unformat_input_t * input, vlib_cli_command_t * cmd)
{
    unformat_input_t _line_input, *line_input = &_line_input;
    ip46_address_t local = ip46_address_initializer;
    ip46_address_t remote = ip46_address_initializer;
    u32 local_port = 0, remote_port = 0;
    scion_ifid_t ifid = 0;
    scion_linkto_t linkto;
    scion_isdas_t isdas;
    u8 ifid_set = 0, local_set = 0, remote_set = 0, local_port_set = 0, remote_port_set = 0;
    u8 linkto_set = 0, isdas_set = 0;

    /* Get a line of input. */
    if (!unformat_user (input, unformat_line_input, line_input)) {
        return 0;
    }

    while (unformat_check_input (line_input) != UNFORMAT_END_OF_INPUT) {
        if (unformat (line_input, "id %d", &ifid)) {
            ifid_set = 1;
        } else if (unformat (line_input, "local %U", unformat_ip46_address, &local, IP46_TYPE_ANY)) {
            local_set = 1;
        } else if (unformat (line_input, "local-port %u", &local_port)) {
            local_port_set = 1;
        } else if (unformat (line_input, "remote %U",
                             unformat_ip46_address, &remote, IP46_TYPE_ANY)) {
            remote_set = 1;
        } else if (unformat (line_input, "remote-port %u", &remote_port)) {
            remote_port_set = 1;
        } else if (unformat (line_input, "link-to %U", unformat_scion_linkto, &linkto)) {
            linkto_set = 1;
        } else if (unformat (line_input, "isd-as %U", unformat_scion_isdas, &isdas)) {
            isdas_set = 1;
        } else {
            return clib_error_return (0, "parse error: '%U'", format_unformat_error, line_input);
        }
    }
    unformat_free (line_input);

    if (!ifid_set) {
        return clib_error_return (0, "id not specified");
    }
    if (!local_set) {
        return clib_error_return (0, "local not specified");
    }
    if (!local_port_set) {
        return clib_error_return (0, "local_port not specified");
    }
    if (local_port >> 16) {
        return clib_error_return (0, "local_port %u out of range", local_port);
    }
    if (ifid != 0) {
        if (!remote_set) {
            return clib_error_return (0, "remote not specified");
        }
        if (!ip46_address_cmp (&local, &remote)) {
            return clib_error_return (0, "local and remote addresses are identical");
        }
        if (ip46_address_is_ip4 (&local) != ip46_address_is_ip4 (&remote)) {
            return clib_error_return (0, "both IPv4 and IPv6 addresses specified");
        }
        if (!remote_port_set) {
            return clib_error_return (0, "remote_port not specified");
        }
        if (remote_port >> 16) {
            return clib_error_return (0, "remote_port %u out of range", remote_port);
        }
        if (!linkto_set) {
            return clib_error_return (0, "linkto not specified");
        }
        if (!isdas_set) {
            return clib_error_return (0, "isdas not specified");
        }
    }

    scion_add_intf_args_t a = {
        .ifid = ifid,
        .local = local,
        .remote = remote,
        .local_port = local_port,
        .remote_port = remote_port,
        .linkto = linkto,
        .isdas = isdas,
    };

    u32 intf_sw_if_index;
    int rv = scion_add_intf (&a, &intf_sw_if_index);

    switch (rv) {
    case 0:
        vlib_cli_output (vm, "%U\n", format_vnet_sw_if_index_name,
                         vnet_get_main (), intf_sw_if_index);
        break;
    case VNET_API_ERROR_IF_ALREADY_EXISTS:
        return clib_error_return (0, "intf already exists...");
    case VNET_API_ERROR_ADDRESS_IN_USE:
        return clib_error_return (0, "intf address already in use...");
    default:
        return clib_error_return (0, "scion_add_intf returned %d", rv);
    }
    return 0;
}

/*?
 * Create a SCION interface.
 *
 * Interface ID 0 is special and is used to represent the internal interface.
 * When ID 0 is specified only the id, local and local-port arguments are required.
 ?*/
/* *INDENT-OFF* */
VLIB_CLI_COMMAND (create_scion_intf_command, static) = {
  .path = "create scion intf",
  .short_help = "create scion intf id <u64> local <ip46> local-port <u16> "
      "[remote <ip46> remote-port <u16> link-to <linkto> isd-as <isdas>]",
  .function = scion_add_intf_command_fn,
};
/* *INDENT-ON* */

static clib_error_t *
scion_del_intf_command_fn (vlib_main_t * vm, unformat_input_t * input, vlib_cli_command_t * cmd)
{
    unformat_input_t _line_input, *line_input = &_line_input;
    ip46_address_t local = ip46_address_initializer;
    u32 local_port = 0;
    scion_ifid_t ifid = 0;
    u8 ifid_set = 0, local_set = 0, local_port_set = 0;

    /* Get a line of input. */
    if (!unformat_user (input, unformat_line_input, line_input)) {
        return 0;
    }

    while (unformat_check_input (line_input) != UNFORMAT_END_OF_INPUT) {
        if (unformat (line_input, "id %d", &ifid)) {
            ifid_set = 1;
        } else if (unformat (line_input, "local %U", unformat_ip46_address, &local, IP46_TYPE_ANY)) {
            local_set = 1;
        } else if (unformat (line_input, "local-port %u", &local_port)) {
            local_port_set = 1;
        } else {
            return clib_error_return (0, "parse error: '%U'", format_unformat_error, line_input);
        }
    }
    unformat_free (line_input);

    if (!ifid_set) {
        return clib_error_return (0, "id not specified");
    }
    if (!local_set) {
        return clib_error_return (0, "local not specified");
    }
    if (!local_port_set) {
        return clib_error_return (0, "local_port not specified");
    }
    if (local_port >> 16) {
        return clib_error_return (0, "local_port %u out of range", local_port);
    }

    scion_del_intf_args_t a = {
        .ifid = ifid,
        .local = local,
        .local_port = local_port,
    };

    int rv = scion_del_intf (&a);

    switch (rv) {
    case 0:
        /* success, nothing to print */
        break;
    case VNET_API_ERROR_NO_SUCH_ENTRY:
        return clib_error_return (0, "intf does not exist...");

    case VNET_API_ERROR_ADDRESS_NOT_IN_USE:
        return clib_error_return (0, "intf address not in use...");

    default:
        return clib_error_return (0, "scion_del_intf returned %d", rv);
    }
    return 0;
}

/*?
 * Delete a SCION interface.
 ?*/
/* *INDENT-OFF* */
VLIB_CLI_COMMAND (delete_scion_intf_command, static) = {
    .path = "delete scion intf",
    .short_help = "delete scion intf id <id> local <ip46> local-port <u16>",
    .function = scion_del_intf_command_fn,
};
/* *INDENT-ON* */

/* *INDENT-OFF* */
static clib_error_t *
show_scion_intf_command_fn (vlib_main_t * vm, unformat_input_t * input, vlib_cli_command_t * cmd)
{
    scion_main_t *scm = &scion_main;
    scion_intf_t *t;
    u8 raw = 0;

    if (unformat (input, "raw")) {
        raw = 1;
    }

    if (pool_elts (scm->intfs) == 0) {
        vlib_cli_output (vm, "No scion intfs configured...");
    }

    pool_foreach (t, scm->intfs, ({
        vlib_cli_output (vm, "%U", format_scion_intf, t);
    }));

    if (raw) {
        vlib_cli_output (vm, "Raw IPv4 Hash Table:\n%U\n",
                         format_bihash_8_8, &scm->scion4_intf_by_key, 1 /* verbose */ );
        vlib_cli_output (vm, "Raw IPv6 Hash Table:\n%U\n",
                         format_bihash_24_8, &scm->scion6_intf_by_key, 1 /* verbose */ );
    }

    return 0;
}
/* *INDENT-ON* */

/*?
 * Display all the SCION interfaces.
 *
 * @cliexpar
 * Example of how to display the SCION interface entries:
 * @cliexstart{show scion intf}
 * [0] addr 10.0.3.1 port 30045 sw_if_index 5
 * [96] addr 10.0.3.1 port 30045 sw_if_index 5
 * @cliexend
 ?*/
/* *INDENT-OFF* */
VLIB_CLI_COMMAND (show_scion_intf_command, static) = {
    .path = "show scion intf",
    .short_help = "show scion intf [raw]",
    .function = show_scion_intf_command_fn,
};
/* *INDENT-ON* */
