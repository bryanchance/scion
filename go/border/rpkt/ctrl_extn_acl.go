// Copyright Anapaya 2017

package rpkt

import (
	"strconv"

	//log "github.com/inconshreveable/log15"
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
	if len(rp.Ingress.IfIDs) != 1 {
		return HookError, common.NewBasicError("Unexpected IFIDs count for external packet", nil,
			"IFIDs", rp.Ingress.IfIDs, "actual", len(rp.Ingress.IfIDs), "expected", 1)
	}
	// Set ACL for the outgoing socket corresponding to the IFID on which we received it
	ifid := rp.Ingress.IfIDs[0]
	aclextn.Map().Store(ifid, acl.ACL())
	return HookFinish, nil
}

func (rp *RtrPkt) RegisterACLHook() {
	rp.hooks.Validate = append(rp.hooks.Validate, rp.validateACLHook)
}

func (rp *RtrPkt) validateACLHook() (HookResult, error) {
	intf := *rp.ifCurr
	if acl, _ := aclextn.Map().Load(intf); len(acl) > 0 {
		// Filtering is enabled, only allow matching packets through
		// srcIA is not guaranteed to be set at this stage
		ia, err := rp.SrcIA()
		if err != nil {
			return HookError, common.NewBasicError("Unable to determine source IA for ACL", nil,
				"rpkt", rp)
		}
		if !acl.Match(ia) {
			//log.Debug("Packet dropped due to ACL", "srcIA", rp.srcIA, "acl", acl)
			metrics.ACLDroppedPkts.With(prometheus.Labels{"ifid": strconv.Itoa(int(intf))}).Inc()
			return HookError, nil
		}
	}
	return HookContinue, nil
}
