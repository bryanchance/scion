/*
 * Copyright (c) 2019 Anapaya Systems
 */
/*
 * SCION headers
 */
#ifndef __SCION_ERROR_H__
#define __SCION_ERROR_H__

#define foreach_scion_error \
_(NONE, "valid scion packets") \
_(TOO_SHORT, "scion length < 24 bytes") \
_(BAD_LENGTH, "scion length > l2 length") \
_(BAD_HEADER_LENGTH, "scion header length > scion length") \
_(BAD_CURRENT_INFOF, "scion current info < common + address header length") \
_(BAD_CURRENT_INFOF_HOPF, "scion current info field >= current hop field") \
_(BAD_CURRENT_HOPF, "scion current hopf >= scion header length") \
_(VERSION, "bad version") \
_(BAD_PATH, "scion path length < 24 bytes (min path)") \
_(HOPF_EXPIRED, "hop field expired time") \
_(HOPF_NOT_IN_SEGMENT, "current hop field > current infof + infof_hops ") \
_(HOPF_BAD_MAC, "hop field bad MAC") \
_(HOPF_BAD_INGRESS_INTF, "hop field ingress intf != received intf")

typedef enum {
#define _(sym,str) SCION_ERROR_##sym,
    foreach_scion_error
#undef _
    SCION_ERROR_N
} scion_error_t;

static char *scion_error_strings[] = {
#define _(sym,str) str,
    foreach_scion_error
#undef _
};

#endif /* __SCION_ERROR_H__ */
