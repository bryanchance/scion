// Copyright Anapaya 2017

package rpkt

import (
	"github.com/prometheus/client_golang/prometheus"

	"github.com/scionproto/scion/go/border/extn/aclextn"
	"github.com/scionproto/scion/go/border/metrics"
	"github.com/scionproto/scion/go/border/rcmn"
	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/lib/ctrl/extn"
)

// processExtnACL processes Push ACL messages received from a neighboring router
func (rp *RtrPkt) processExtnACL(extnData common.RawBytes) (HookResult, error) {
	acl, err := extn.NewPushACLFromRaw(extnData)
	if err != nil {
		return HookError, common.NewBasicError("Unable to parse PushACL message", err)
	}
	if rp.DirFrom != rcmn.DirExternal {
		return HookError, common.NewBasicError("Bad packet direction", nil,
			"actual", rp.DirFrom, "expected", rcmn.DirExternal)
	}
	// Set ACL for the outgoing socket corresponding to the IFID on which we received it
	aclextn.Map().Store(rp.Ingress.IfID, acl.ACL())
	return HookFinish, nil
}

func (rp *RtrPkt) RegisterACLHook() {
	rp.hooks.Validate = append(rp.hooks.Validate, rp.validateACLHook)
}

func (rp *RtrPkt) validateACLHook() (HookResult, error) {
	if rp.DirTo != rcmn.DirExternal {
		// We only care about egress packets.
		return HookContinue, nil
	}
	// srcIA is not guaranteed to be set at this stage
	if _, err := rp.SrcIA(); err != nil {
		return HookError, common.NewBasicError("Unable to determine source IA for ACL", nil,
			"rpkt", rp)
	}
	if rp.srcIA.Eq(rp.Ctx.Conf.IA) {
		// Always allow packets from the local IA, which may not have a path header anyway.
		return HookContinue, nil
	}
	if rp.ifCurr == nil {
		return HookError, common.NewBasicError("Unable to determine current interface for ACL", nil,
			"rpkt", rp)
	}
	ifid := *rp.ifCurr
	if acl, _ := aclextn.Map().Load(ifid); len(acl) > 0 {
		// Filtering is enabled, only allow matching packets through
		if !acl.Match(rp.srcIA) {
			//rp.Debug("Packet dropped due to ACL", "srcIA", rp.srcIA, "acl", acl, "ifid", ifid)
			metrics.ACLDroppedPkts.With(prometheus.Labels{"ifid": ifid.String()}).Inc()
			return HookError, nil
		}
	}
	return HookContinue, nil
}
