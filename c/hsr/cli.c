/*
 * Copyright (c) 2019 Anapaya Systems
 */
/*
 * SCION cli commands
 */
#include <vnet/feature/feature.h>
#include <vnet/plugin/plugin.h>
#include <vnet/vnet.h>

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
            goto done;
        }
    }

    if (~0 == sw_if_index) {
        error = clib_error_return (0, "unknown interface `%U'", format_unformat_error, line_input);
        goto done;
    }

    vlib_cli_output (vlib_get_main (), "sw_if_index: %u, is_enable: %u\n", sw_if_index, is_enable);

    vnet_int_scion_bypass_mode (sw_if_index, is_ip6, is_enable);

 done:
    unformat_free (line_input);

    return error;
}

static clib_error_t *
set_ip4_scion_bypass (vlib_main_t * vm, unformat_input_t * input, vlib_cli_command_t * cmd)
{
    return set_ip_scion_bypass (0, input, cmd);
}

/*?
 * This command adds the 'ip4-scion-bypass' graph node for a given interface.
 * By adding the IPv4 scion-bypass graph node to an interface, the node checks
 *  for and validate input scion packet and bypass ip4-lookup, ip4-local,
 * ip4-udp-lookup nodes to speedup scion packet forwarding. This node will
 * cause extra overhead to for non-scion packets which is kept at a minimum.
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
