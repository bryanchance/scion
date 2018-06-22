// Copyright 2018 Anapaya Systems

// Package sessselector implements session selection based on traffic policies.
package sessselector

import (
	"sync/atomic"

	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/lib/pktcls"
	"github.com/scionproto/scion/go/sig/egress"
	"github.com/scionproto/scion/go/sig/mgmt"
)

// Must stay compatible with open source default session selector factory
func NewDefaultSessionSelector() egress.SessionSelector {
	return NewSyncPktPols()
}

type SyncPktPols struct {
	atomic.Value
}

func NewSyncPktPols() *SyncPktPols {
	spp := &SyncPktPols{}
	spp.Store(([]*PktPolicy)(nil))
	return spp
}

func (spp *SyncPktPols) Load() []*PktPolicy {
	return spp.Value.Load().([]*PktPolicy)
}

func (spp *SyncPktPols) ChooseSess(b common.RawBytes) egress.Session {
	var sess egress.Session
	ppols := spp.Load()
	clsPkt := pktcls.NewPacket(b)
	for _, ppol := range ppols {
		if ppol.Class.Eval(clsPkt) {
			for _, sess = range ppol.Sessions {
				if sess.Healthy() {
					return sess
				}
				// XXX(kormat): If all sessions are unhealthy, the last
				// one will be chosen.
			}
		}
	}
	return sess
}

type PktPolicy struct {
	ClassName string
	Class     *pktcls.Class
	Sessions  []egress.Session
}

func NewPktPolicy(name string, cls *pktcls.Class, sessIds []mgmt.SessionType,
	currSessions egress.SessionSet) (*PktPolicy, error) {
	ppol := &PktPolicy{ClassName: name, Class: cls, Sessions: make([]egress.Session, len(sessIds))}
Top:
	for i, sessId := range sessIds {
		for _, sess := range currSessions {
			if sessId == sess.ID() {
				ppol.Sessions[i] = sess
				continue Top
			}
		}
		return nil, common.NewBasicError("newPktPolicy: unknown session id", nil,
			"name", name, "sessId", sessId)
	}
	return ppol, nil
}
