// Copyright 2018 Anapaya Systems

// Package policypathpool implements path selection based on path policies.
package policypathpool

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/scionproto/scion/go/lib/addr"
	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/lib/pathmgr"
	"github.com/scionproto/scion/go/lib/pathpol"
	"github.com/scionproto/scion/go/lib/spath/spathmeta"
	"github.com/scionproto/scion/go/sig/sigcmn"
)

// Pool enhances the open source version with policy-aware path filtering.
type Pool struct {
	ia         addr.IA
	policyLock sync.Mutex
	polName    string
	pool       *atomicSP
}

func NewPool(dst addr.IA, name string,
	pp *pathpol.Policy) (*Pool, error) {

	pool, err := sigcmn.PathMgr.WatchFilter(context.TODO(), sigcmn.IA, dst, pp)
	if err != nil {
		return nil, common.NewBasicError("Unable to register watch", err)
	}
	atomicPool := NewAtomicSP()
	atomicPool.UpdateSP(pool)

	return &Pool{
		ia:      dst,
		polName: name,
		pool:    atomicPool,
	}, nil
}

func (ppp *Pool) Destroy() error {
	ppp.policyLock.Lock()
	defer ppp.policyLock.Unlock()

	oldPool := ppp.pool.GetSP()
	ppp.pool.UpdateSP(nil)
	oldPool.Destroy()
	return nil
}

func (ppp *Pool) Update(name string, pp *pathpol.Policy) error {
	ppp.policyLock.Lock()
	defer ppp.policyLock.Unlock()

	pool, err := sigcmn.PathMgr.WatchFilter(context.TODO(), sigcmn.IA, ppp.ia, pp)
	if err != nil {
		return common.NewBasicError("Unable to register watch", err)
	}
	oldPool := ppp.pool.GetSP()
	ppp.polName = name
	ppp.pool.UpdateSP(pool)
	oldPool.Destroy()

	return nil
}

// Paths returns a read-only snapshot of this pool's managed paths.
func (ppp *Pool) Paths() spathmeta.AppPathSet {
	return ppp.pool.GetSP().Load().APS
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

func (a *atomicSP) GetSP() *pathmgr.SyncPaths {
	return a.v.Load().(*pathmgr.SyncPaths)
}
