/*
 * Copyright (c) 2019 Anapaya Systems
 */
/*
 * SCION plugin node registration and initialization
 */
#include <vnet/feature/feature.h>
#include <vnet/plugin/plugin.h>
#include <vnet/vnet.h>

#define VPP_VER "18.10"

clib_error_t *
scion_init (vlib_main_t * vm)
{
    // XXX Currently we do not need any setup

    return 0;
}

VLIB_INIT_FUNCTION (scion_init);

/* *INDENT-OFF* */
VLIB_PLUGIN_REGISTER () = {
    .version = VPP_VER,
    .description = "SCION",
};
/* *INDENT-ON* */
