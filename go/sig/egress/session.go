// Copyright 2017 ETH Zurich
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package egress

import (
	"fmt"
	"sync"
	"sync/atomic"

	log "github.com/inconshreveable/log15"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/scionproto/scion/go/lib/addr"
	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/lib/pathmgr"
	"github.com/scionproto/scion/go/lib/pktdisp"
	"github.com/scionproto/scion/go/lib/ringbuf"
	"github.com/scionproto/scion/go/lib/snet"
	"github.com/scionproto/scion/go/sig/mgmt"
	"github.com/scionproto/scion/go/sig/sigcmn"
	"github.com/scionproto/scion/go/sig/siginfo"
)

type SessionSet map[mgmt.SessionType]*Session

// Session contains a pool of paths to the remote AS, metrics about those paths,
// as well as maintaining the currently favoured path and remote SIG to use.
// The Anapaya version uses packet classification and path predicates to all
// configurable routing.
type Session struct {
	log.Logger
	IA     *addr.ISD_AS
	SessId mgmt.SessionType

	// used when updating the path policy or its name
	policyLock sync.Mutex
	// used by pathmgr to filter the paths in the pool
	policy  *pathmgr.PathPredicate
	polName string

	// pool of paths, managed by pathmgr
	pool *AtomicSP
	// remote SIGs
	sigMap *siginfo.SigMap
	// *RemoteInfo
	currRemote atomic.Value
	// bool
	healthy        atomic.Value
	ring           *ringbuf.Ring
	conn           *snet.Conn
	sessMonStop    chan struct{}
	sessMonStopped chan struct{}
	workerStopped  chan struct{}
}

func NewSession(dstIA *addr.ISD_AS, sessId mgmt.SessionType,
	sigMap *siginfo.SigMap, logger log.Logger,
	polName string, policy *pathmgr.PathPredicate) (*Session, error) {
	var err error
	s := &Session{
		Logger:  logger.New("sessId", sessId),
		IA:      dstIA,
		SessId:  sessId,
		polName: polName,
		policy:  policy,
		pool:    NewAtomicSP(),
		sigMap:  sigMap,
	}
	pool, err := sigcmn.PathMgr.WatchFilter(sigcmn.IA, s.IA, s.policy)
	if err != nil {
		return nil, err
	}
	s.pool.UpdateSP(pool)
	s.currRemote.Store((*RemoteInfo)(nil))
	s.healthy.Store(false)
	s.ring = ringbuf.New(64, nil, "egress",
		prometheus.Labels{"ringId": dstIA.String(), "sessId": sessId.String()})
	// Not using a fixed local port, as this is for outgoing data only.
	s.conn, err = snet.ListenSCION("udp4", &snet.Addr{IA: sigcmn.IA, Host: sigcmn.Host})
	// spawn a PktDispatcher to log any unexpected messages received on a write-only connection.
	go pktdisp.PktDispatcher(s.conn, pktdisp.DispLogger)
	s.sessMonStop = make(chan struct{})
	s.sessMonStopped = make(chan struct{})
	s.workerStopped = make(chan struct{})
	return s, err
}

func (s *Session) Start() {
	go newSessMonitor(s).run()
	go NewWorker(s, s.Logger).Run()
}

func (s *Session) Cleanup() error {
	s.ring.Close()
	close(s.sessMonStop)
	s.Debug("egress.Session Cleanup: wait for worker")
	<-s.workerStopped
	s.Debug("egress.Session Cleanup: wait for session monitor")
	<-s.sessMonStopped
	s.Debug("egress.Session Cleanup: closing conn")
	if err := s.conn.Close(); err != nil {
		return common.NewCError("Unable to close conn", "err", err)
	}
	if err := sigcmn.PathMgr.Unwatch(sigcmn.IA, s.IA); err != nil {
		return common.NewCError("Unable to unwatch src-dst", "src", sigcmn.IA, "dst", s.IA,
			"err", err)
	}
	return nil
}

func (s *Session) Remote() *RemoteInfo {
	return s.currRemote.Load().(*RemoteInfo)
}

func (s *Session) Healthy() bool {
	// FIxME(kormat): export as metric.
	return s.healthy.Load().(bool)
}

func (s *Session) UpdatePolicy(name string, pred *pathmgr.PathPredicate) error {
	s.policyLock.Lock()
	defer s.policyLock.Unlock()

	pool, err := sigcmn.PathMgr.WatchFilter(sigcmn.IA, s.IA, pred)
	if err != nil {
		return common.NewCError("Unable to register watch", "err", err)
	}
	// Store old predicate so we can unwatch it later
	oldPred := s.policy
	s.polName = name
	s.policy = pred
	s.pool.UpdateSP(pool)
	if err := sigcmn.PathMgr.UnwatchFilter(sigcmn.IA, s.IA, oldPred); err != nil {
		return common.NewCError("Unable to unregister watch", "err", err)
	}
	return nil
}

type RemoteInfo struct {
	Sig      *siginfo.Sig
	sessPath *sessPath
}

func (r *RemoteInfo) String() string {
	return fmt.Sprintf("Sig: %s Path: %s", r.Sig, r.sessPath)
}
