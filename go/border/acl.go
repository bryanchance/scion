// Copyright 2017 Anapaya

package main

import (
	"time"

	"github.com/scionproto/scion/go/border/netconf"
	"github.com/scionproto/scion/go/border/rctx"
	"github.com/scionproto/scion/go/lib/addr"
	"github.com/scionproto/scion/go/lib/ctrl"
	"github.com/scionproto/scion/go/lib/ctrl/extn"
	"github.com/scionproto/scion/go/lib/log"
)

const (
	pushInterval = 5 * time.Second
	aclBufSize   = 1 << 10
)

func (r *Router) PeriodicPushACL() {
	defer log.LogPanicAndExit()
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
	cpld, err := ctrl.NewPld(extnDataListMsgObj, nil)
	if err != nil {
		logger.Error("Unable to construct CtrlPld object", "err", err)
		return
	}
	scpld, err := cpld.SignedPld(ctrl.NullSigner)
	if err != nil {
		logger.Error("Unable to construct signed CtrlPld object", "err", err)
		return
	}
	src := intf.IFAddr.PublicAddr(intf.IFAddr.Overlay)
	dst := &addr.AppAddr{L3: intf.RemoteAddr.L3(), L4: intf.RemoteAddr.L4()}
	if err := r.genPkt(intf.RemoteIA, dst, src, intf.RemoteAddr, scpld); err != nil {
		logger.Error("Error generating ACL Push packet", "err", err)
	}
}
