/*
 * Copyright (c) 2019 Anapaya Systems
 */
/*
 * SCION headers
 */
#ifndef __SCION_EXTENSIONS_H__
#define __SCION_EXTENSIONS_H__

#define SCION_EXT_HBH 0
#define SCION_EXT_E2E 222

#define foreach_scion_hbh_ext \
_(SCMP, "SCMP")  \
_(OHP, "One Hop Path")  \
_(SIBRA, "Sibra")

typedef enum {
#define _(v, s) SCION_HBH_EXT_##v,
    foreach_scion_hbh_ext
#undef _
    SCION_HBH_EXT_N
} scion_hbh_ext_t;

#define PATH_TRANSPORT 0
#define PATH_PROBE 1
#define SCION_PKT_SECURITY 2

#define foreach_scion_e2e_ext \
_(PATH_TRANSPORT, "Path Transport")  \
_(PATH_PROBE, "Path Probe")  \
_(PKT_SECURITY, "Packet Security")

typedef enum {
#define _(v, s) SCION_E2E_EXT_##v,
    foreach_scion_e2e_ext
#undef _
    SCION_E2E_EXT_N
} scion_e2e_ext_t;

#endif /* __SCION_EXTENSIONS_H__ */
