// Copyright 2017 Anapaya

package main

import (
	"time"

	log "github.com/inconshreveable/log15"

	"github.com/scionproto/scion/go/border/netconf"
	"github.com/scionproto/scion/go/border/rctx"
	"github.com/scionproto/scion/go/lib/addr"
	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/lib/ctrl"
	"github.com/scionproto/scion/go/lib/ctrl/extn"
	liblog "github.com/scionproto/scion/go/lib/log"
)

const (
	pushInterval = 5 * time.Second
	aclBufSize   = 1 << 10
)

func (r *Router) PeriodicPushACL() {
	defer liblog.LogPanicAndExit()
	for range time.Tick(pushInterval) {
		ctx := rctx.Get()
		for _, intf := range ctx.Conf.Net.IFs {
			r.genACLPkt(intf)
		}
	}
}

func (r *Router) genACLPkt(intf *netconf.Interface) {
	logger := log.New("ifid", intf.Id)
	aclMsgObj := extn.NewPushACLFromValues(intf.PushACL)
	extnDataMsgObj, err := extn.NewCtrlExtnDataFromValues(aclMsgObj, aclBufSize)
	if err != nil {
		logger.Error("Unable to construct CtrlExtnData object", "err", err)
		return
	}
	extnDataListMsgObj := extn.NewCtrlExtnDataListFromValues([]*extn.CtrlExtnData{extnDataMsgObj})
	scpld, err := ctrl.NewSignedPldFromUnion(extnDataListMsgObj)
	if err != nil {
		logger.Error("Unable to construct signed CtrlPld object", "err", err)
		return
	}
	srcAddr := intf.IFAddr.PublicAddrInfo(intf.IFAddr.Overlay)
	if err := r.genPkt(intf.RemoteIA, addr.HostFromIP(intf.RemoteAddr.IP),
		intf.RemoteAddr.L4Port, srcAddr, scpld); err != nil {
		cerr := err.(*common.CError)
		cerr.AddCtx("desc", cerr.Desc)
		logger.Error("Error generating ACL Push packet", cerr.Ctx...)
	}
}
