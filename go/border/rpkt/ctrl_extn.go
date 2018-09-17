// Copyright Anapaya 2017
// +build ignore

package rpkt

import (
	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/lib/ctrl/extn"
)

// processACL handles ACL messages from neighbors
func (rp *RtrPkt) processExtnList(edl *extn.CtrlExtnDataList) (HookResult, error) {
	for _, e := range edl.Items {
		switch e.TypeStr() {
		case "com.anapaya.pushacl":
			rp.processExtnACL(e.Data)
		default:
			return HookError, common.NewBasicError("Unsupported ctrl extension", nil,
				"type", e.Type)
		}
	}
	return HookContinue, nil
}
