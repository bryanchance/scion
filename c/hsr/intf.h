/*
 * Copyright (c) 2019 Anapaya Systems
 */
/*
 * SCION virtual interfaces
 */
#ifndef __SCION_INTF_H__
#define __SCION_INTF_H__

typedef u64 scion_ifid_t;

typedef u64 scion_isdas_t;

#define foreach_scion_linkto \
    _(CORE) \
    _(PARENT) \
    _(CHILD) \
    _(PEER)

typedef enum {
#define _(sym) LINK_TO_##sym,
    foreach_scion_linkto
#undef _
    SCION_LINKTO_N
} scion_linkto_t;

static const char *scion_linkto_str[] = {
#define _(sym) #sym,
    foreach_scion_linkto
#undef _
};

typedef struct {
    scion_ifid_t ifid;
    ip46_address_t local;
    ip46_address_t remote;
    u16 local_port;
    u16 remote_port;
    scion_linkto_t linkto;
    scion_isdas_t isdas;
} scion_add_intf_args_t;

typedef struct {
    scion_ifid_t ifid;
    ip46_address_t local;
    u16 local_port;
} scion_del_intf_args_t;

typedef struct {
    /* Required for pool_get_aligned */
    CLIB_CACHE_LINE_ALIGN_MARK (cacheline0);

    scion_ifid_t ifid;
    ip46_address_t local;
    ip46_address_t remote;
    u16 local_port;
    u16 remote_port;
    scion_linkto_t linkto;
    scion_isdas_t isdas;

    /* vnet intfc index */
    u32 sw_if_index;
    u32 hw_if_index;
} scion_intf_t;

int scion_add_intf (scion_add_intf_args_t * a, u32 * sw_if_indexp);

int scion_del_intf (scion_del_intf_args_t * a);

#endif /* __SCION_INTF_H__ */
