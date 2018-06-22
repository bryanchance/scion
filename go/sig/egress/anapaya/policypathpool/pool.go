// Copyright 2018 Anapaya Systems

// Package policypathpool implements path selection based on path policies.
package policypathpool

import (
	"sync"
	"sync/atomic"

	"github.com/scionproto/scion/go/lib/addr"
	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/lib/pathmgr"
	"github.com/scionproto/scion/go/lib/pktcls"
	"github.com/scionproto/scion/go/lib/spath/spathmeta"
	"github.com/scionproto/scion/go/sig/sigcmn"
)

// Pool enhances the open source version with policy-aware path filtering.
type Pool struct {
	ia         addr.IA
	policyLock sync.Mutex
	// used by pathmgr to filter the paths in the pool
	policy  *pktcls.ActionFilterPaths
	polName string
	pool    *atomicSP
}

func NewPool(dst addr.IA, name string,
	afp *pktcls.ActionFilterPaths) (*Pool, error) {

	pool, err := sigcmn.PathMgr.WatchFilter(sigcmn.IA, dst, afp)
	if err != nil {
		return nil, common.NewBasicError("Unable to register watch", err)
	}
	atomicPool := NewAtomicSP()
	atomicPool.UpdateSP(pool)

	return &Pool{
		ia:      dst,
		polName: name,
		policy:  afp,
		pool:    atomicPool,
	}, nil
}

func (ppp *Pool) Destroy() error {
	ppp.policyLock.Lock()
	defer ppp.policyLock.Unlock()

	if err := sigcmn.PathMgr.UnwatchFilter(sigcmn.IA, ppp.ia, ppp.policy); err != nil {
		return common.NewBasicError("Unable to unregister watch", err)
	}
	return nil
}

func (ppp *Pool) Update(name string, afp *pktcls.ActionFilterPaths) error {
	ppp.policyLock.Lock()
	defer ppp.policyLock.Unlock()

	pool, err := sigcmn.PathMgr.WatchFilter(sigcmn.IA, ppp.ia, afp)
	if err != nil {
		return common.NewBasicError("Unable to register watch", err)
	}
	// Store old predicate so we can unwatch it later
	oldPolicy := ppp.policy
	ppp.polName = name
	ppp.policy = afp
	ppp.pool.UpdateSP(pool)
	if err := sigcmn.PathMgr.UnwatchFilter(sigcmn.IA, ppp.ia, oldPolicy); err != nil {
		return common.NewBasicError("Unable to unregister watch", err)
	}
	return nil
}

// Paths returns a read-only snapshot of this pool's managed paths.
func (ppp *Pool) Paths() spathmeta.AppPathSet {
	return ppp.pool.GetSPD().APS
}

// atomicSP contains a pointer to a SyncPaths object; the pointer itself can be
// changed atomically via method UpdateSP. Method GetSPD returns the paths
// within the current SyncPaths object.
type atomicSP struct {
	// v contains *pathmgr.SyncPaths
	v atomic.Value
}

func NewAtomicSP() *atomicSP {
	return &atomicSP{}
}

func (a *atomicSP) UpdateSP(sp *pathmgr.SyncPaths) {
	a.v.Store(sp)
}

func (a *atomicSP) GetSPD() *pathmgr.SyncPathsData {
	sp := a.v.Load().(*pathmgr.SyncPaths)
	return sp.Load()
}
