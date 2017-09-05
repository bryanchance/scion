// Copyright 2017 Anapaya Systems

// Package acl implements a very simple IP/CIDR-based whitelisting ACL system.
//
// Syntax of the ACL file is simple:
//
// - Everything after a `#` or whitespace is dropped from further consideration
// - Empty lines are skipped
// - The rest is interpreted as an IPv4 or IPv6 address plus a mask (as
// understood by `net.ParseCIDR()`)
//
// If the call to `net.ParseCIDR` fails, parsing stops and an error is returned.
// Like all other error cases (e.g. unable to open file) an empty ACL is
// returned, which means that the ACL system fails closed (no access is
// allowed).
//
// When `IsAllowed()` is called, the ACL is examined in file sequence and on
// the first match, `true` is returned. If the end of the ACL is reached, the
// result is `false`. Thus, it is recommended to use IP networks instead of
// bare (`/32` or `/128` for v4 and v6 respectively, or omitting the mask
// entirely). Also, large, often-matching entries should be near the top.
//
// If the intention is to whitelist every IPv4 and IPv6 address, the entries in
// the ACL file should b `0.0.0.0/0` and `::/0`, respectively
package acl

import (
	"io/ioutil"
	"net"
	"strings"
	"sync/atomic"

	"github.com/gavv/monotime"

	log "github.com/inconshreveable/log15"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/netsec-ethz/scion/go/discovery/metrics"
	"github.com/netsec-ethz/scion/go/lib/common"
)

const (
	ERRREADFILE = "read-file-error"
	ERRPARSEACL = "parse-acl-error"
	SUCCESS     = "success"

	RESALLOWED = "allowed"
	RESDENIED  = "denied"
)

type atomicACL struct {
	atomic.Value
}

func (a *atomicACL) Load() []net.IPNet {
	return a.Value.Load().([]net.IPNet)
}

var fullTopoACL *atomicACL

func init() {
	fullTopoACL = &atomicACL{}
	fullTopoACL.Store([]net.IPNet{})
}

func Load(filename string) *common.Error {
	l := prometheus.Labels{"result": ""}
	defer func() {
		metrics.TotalACLLoads.With(l).Inc()
	}()
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		l["result"] = ERRREADFILE
		return common.NewError("Could not open ACL file", "filename", filename, "err", err)
	}
	newacl, cerr := makeACL(string(data))
	if cerr != nil {
		l["result"] = ERRPARSEACL
		return common.NewError("Could not parse ACL file", "filename", filename, "err", cerr)
	}
	fullTopoACL.Store(newacl)
	l["result"] = SUCCESS
	log.Info("Loaded ACL", "entries", len(newacl))
	return nil
}

func makeACL(iplist string) ([]net.IPNet, *common.Error) {
	var newacl []net.IPNet
	lines := strings.Split(iplist, "\n")
	for i, l := range lines {
		if len(l) == 0 {
			continue
		}
		if l[0] == '#' {
			continue
		}
		ipnet := strings.SplitN(l, " ", 2)[0]
		if ipnet == "" {
			// This can happen if a line has leading space and then a comment,
			// e.g. " # foo"
			continue
		}
		_, n, err := net.ParseCIDR(ipnet)
		if err != nil {
			return nil, common.NewError(
				"Could not convert string to IP network",
				"lineno", i+1, "string", ipnet, "err", err)
		}
		newacl = append(newacl, *n)
	}
	return newacl, nil
}

func IsAllowed(address net.IP) bool {
	l := prometheus.Labels{"result": ""}
	start := monotime.Now()
	curracl := fullTopoACL.Load()
	for _, n := range curracl {
		if n.Contains(address) {
			l["result"] = RESALLOWED
			metrics.TotalACLChecks.With(l).Inc()
			metrics.TotalACLCheckTime.With(l).Add(monotime.Since(start).Seconds())
			return true
		}
	}
	l["result"] = RESDENIED
	metrics.TotalACLChecks.With(l).Inc()
	metrics.TotalACLCheckTime.With(l).Add(monotime.Since(start).Seconds())
	return false
}
