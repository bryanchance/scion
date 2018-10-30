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

package core

import (
	"net"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/scionproto/scion/go/lib/addr"
	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/lib/log"
	"github.com/scionproto/scion/go/lib/pktcls"
	"github.com/scionproto/scion/go/lib/ringbuf"
	"github.com/scionproto/scion/go/sig/anaconfig"
	"github.com/scionproto/scion/go/sig/base"
	"github.com/scionproto/scion/go/sig/egress"
	"github.com/scionproto/scion/go/sig/egress/anapaya/policypathpool"
	"github.com/scionproto/scion/go/sig/egress/anapaya/sessselector"
	"github.com/scionproto/scion/go/sig/egress/dispatcher"
	"github.com/scionproto/scion/go/sig/egress/router"
	"github.com/scionproto/scion/go/sig/egress/session"
	"github.com/scionproto/scion/go/sig/egress/worker"
	"github.com/scionproto/scion/go/sig/mgmt"
	"github.com/scionproto/scion/go/sig/sigcmn"
	"github.com/scionproto/scion/go/sig/siginfo"
)

const (
	sigMgrTick        = 10 * time.Second
	healthMonitorTick = 5 * time.Second
	DefSessId         = mgmt.SessionType(0)
)

// ASEntry contains all of the information required to interact with a remote AS.
type ASEntry struct {
	sync.RWMutex
	Nets              map[string]*net.IPNet
	Sigs              *siginfo.SigMap
	IA                addr.IA
	IAString          string
	egressRing        *ringbuf.Ring
	sigMgrStop        chan struct{}
	healthMonitorStop chan struct{}
	version           uint64 // used to track certain changes made to ASEntry
	log.Logger

	PktPolicies *sessselector.SyncPktPols
	Sessions    egress.SessionSet
}

func newASEntry(ia addr.IA) (*ASEntry, error) {
	ae := &ASEntry{
		Logger:   log.New("ia", ia),
		IA:       ia,
		IAString: ia.String(),
		Nets:     make(map[string]*net.IPNet),
		Sigs:     &siginfo.SigMap{},
	}
	ae.PktPolicies = sessselector.NewSyncPktPols()
	ae.Sessions = make(egress.SessionSet)
	return ae, nil
}

func (ae *ASEntry) ReloadConfig(cfg *config.ASEntry, classes pktcls.ClassMap,
	actions pktcls.ActionMap) bool {
	ae.Lock()
	defer ae.Unlock()
	// Method calls first to prevent skips due to logical short-circuit
	s := ae.addNewSIGS(cfg.Sigs)
	s = ae.delOldSIGS(cfg.Sigs) && s
	s = ae.addNewNets(cfg.Nets) && s
	s = ae.delOldNets(cfg.Nets) && s

	sessionCfgs, err := config.BuildSessions(cfg.Sessions, actions)
	if err != nil {
		ae.Error("Unable to update sessions", "err", err)
		s = false
	}
	s = ae.addNewSessions(sessionCfgs) && s
	deletedSessions := ae.delOldSessions(sessionCfgs)
	// Build new packet policies on top of the updated sessions list, and then
	// completely overwrite the old policies.
	newPktPolicies := ae.buildNewPktPolicies(cfg.PktPolicies, classes)
	ae.PktPolicies.Store(newPktPolicies)
	// Clean up old sessions
	for _, session := range deletedSessions {
		session.Cleanup()
	}
	return s
}

// addNewNets adds the networks in ipnets that are not currently configured.
func (ae *ASEntry) addNewNets(ipnets []*config.IPNet) bool {
	s := true
	for _, ipnet := range ipnets {
		err := ae.addNet(ipnet.IPNet())
		if err != nil {
			ae.Error("Unable to add network", "net", ipnet, "err", err)
			s = false
		}
	}
	return s
}

// delOldNets deletes currently configured networks that are not in ipnets.
func (ae *ASEntry) delOldNets(ipnets []*config.IPNet) bool {
	s := true
Top:
	for k, v := range ae.Nets {
		for _, ipnet := range ipnets {
			if k == ipnet.IPNet().String() {
				continue Top
			}
		}
		err := ae.delNet(v)
		if err != nil {
			ae.Error("Unable to delete network", "net", k, "err", err)
			s = false
		}
	}
	return s
}

// AddNet idempotently adds a network for the remote IA.
func (ae *ASEntry) AddNet(ipnet *net.IPNet) error {
	ae.Lock()
	defer ae.Unlock()
	return ae.addNet(ipnet)
}

func (ae *ASEntry) addNet(ipnet *net.IPNet) error {
	if ae.egressRing == nil {
		// Ensure that the network setup is done
		ae.setupNet()
	}
	key := ipnet.String()
	if _, ok := ae.Nets[key]; ok {
		return nil
	}
	if err := router.NetMap.Add(ipnet, ae.IA, ae.egressRing); err != nil {
		return err
	}
	ae.Nets[key] = ipnet
	ae.version++
	// Generate NetworkChanged event
	params := base.NetworkChangedParams{
		RemoteIA: ae.IA,
		IpNet:    *ipnet,
		Healthy:  ae.checkHealth(),
		Added:    true,
	}
	base.NetworkChanged(params)
	if len(ae.Nets) == 1 {
		go dispatcher.NewDispatcher(ae.IA, ae.egressRing, ae.PktPolicies).Run()
		ae.sigMgrStop = make(chan struct{})
		go ae.sigMgr()
	}
	ae.Info("Added network", "net", ipnet)
	return nil
}

// DelNet removes a network for the remote IA.
func (ae *ASEntry) DelNet(ipnet *net.IPNet) error {
	ae.Lock()
	defer ae.Unlock()
	return ae.delNet(ipnet)
}

func (ae *ASEntry) delNet(ipnet *net.IPNet) error {
	key := ipnet.String()
	if _, ok := ae.Nets[key]; !ok {
		return common.NewBasicError("DelNet: no network found", nil, "ia", ae.IA, "net", ipnet)
	}
	if err := router.NetMap.Delete(ipnet); err != nil {
		return err
	}
	delete(ae.Nets, key)
	ae.version++
	// Generate NetworkChanged event
	params := base.NetworkChangedParams{
		RemoteIA: ae.IA,
		IpNet:    *ipnet,
		Healthy:  ae.checkHealth(),
		Added:    false,
	}
	base.NetworkChanged(params)
	if len(ae.Nets) == 0 {
		ae.egressRing.Close()
		ae.egressRing = nil
		ae.sigMgrStop <- struct{}{}
	}
	ae.Info("Removed network", "net", ipnet)
	return nil
}

// addNewSIGS adds the SIGs in sigs that are not currently configured.
func (ae *ASEntry) addNewSIGS(sigs config.SIGSet) bool {
	s := true
	for _, sig := range sigs {
		ctrlPort := int(sig.CtrlPort)
		if ctrlPort == 0 {
			ctrlPort = sigcmn.DefaultCtrlPort
		}
		encapPort := int(sig.EncapPort)
		if encapPort == 0 {
			encapPort = sigcmn.DefaultEncapPort
		}
		err := ae.AddSig(sig.Id, sig.Addr, ctrlPort, encapPort, true)
		if err != nil {
			ae.Error("Unable to add SIG", "sig", sig, "err", err)
			s = false
		}
	}
	return s
}

// delOldSIGS deletes the currently configured SIGs that are not in sigs.
func (ae *ASEntry) delOldSIGS(sigs config.SIGSet) bool {
	s := true
	ae.Sigs.Range(func(id siginfo.SigIdType, sig *siginfo.Sig) bool {
		if !sig.Static {
			return true
		}
		if _, ok := sigs[sig.Id]; !ok {
			err := ae.DelSig(sig.Id)
			if err != nil {
				ae.Error("Unable to delete SIG", "err", err)
				s = false
			}
		}
		return true
	})
	return s
}

// AddSig idempotently adds a SIG for the remote IA.
func (ae *ASEntry) AddSig(id siginfo.SigIdType, ip net.IP, ctrlPort, encapPort int,
	static bool) error {
	// ae.Sigs is thread safe, no master lock needed
	if len(id) == 0 {
		return common.NewBasicError("AddSig: SIG id empty", nil, "ia", ae.IA)
	}
	if ip == nil {
		return common.NewBasicError("AddSig: SIG address empty", nil, "ia", ae.IA)
	}
	if err := sigcmn.ValidatePort("remote ctrl", ctrlPort); err != nil {
		return common.NewBasicError("Remote ctrl port validation failed", err,
			"ia", ae.IA, "id", id)
	}
	if err := sigcmn.ValidatePort("remote encap", encapPort); err != nil {
		return common.NewBasicError("Remote encap port validation failed", err,
			"ia", ae.IA, "id", id)
	}
	if sig, ok := ae.Sigs.Load(id); ok {
		sig.Host = addr.HostFromIP(ip)
		sig.CtrlL4Port = ctrlPort
		sig.EncapL4Port = encapPort
		ae.Info("Updated SIG", "sig", sig)
	} else {
		sig := siginfo.NewSig(ae.IA, id, addr.HostFromIP(ip), ctrlPort, encapPort, static)
		ae.Sigs.Store(id, sig)
		ae.Info("Added SIG", "sig", sig)
	}
	return nil
}

// DelSIG removes an SIG for the remote IA.
func (ae *ASEntry) DelSig(id siginfo.SigIdType) error {
	// ae.Sigs is thread safe, no master lock needed
	se, ok := ae.Sigs.Load(id)
	if !ok {
		return common.NewBasicError("DelSig: no SIG found", nil, "ia", ae.IA, "id", id)
	}
	ae.Sigs.Delete(id)
	ae.Info("Removed SIG", "id", id)
	return se.Cleanup()
}

// addNewSessions adds the sessions in cfgs that are not currently configured.
// If a session is already configured, its state is updated to reflect the new
// config.
func (ae *ASEntry) addNewSessions(cfgs config.SessionSet) bool {
	s := true
	for _, cfg := range cfgs {
		if err := ae.addSession(cfg.ID, cfg.PolName, cfg.Pred); err != nil {
			ae.Error("Unable to add session", "id", cfg.ID, "err", err)
			s = false
			// Continue without rollback
			continue
		}
	}
	return s
}

// DelOldSessions removes every session whose session ID is not present in
// cfgs. A map containing the deleted sessions is returned. No cleanup is
// performed on deleted sessions.
func (ae *ASEntry) delOldSessions(cfgs config.SessionSet) egress.SessionSet {
	deleted := make(egress.SessionSet)
	// Destroy existing session IDs that are no longer in the config
	for _, sess := range ae.Sessions {
		// We're also walking newly added sessions that are surely in the
		// config, but since the number of sessions will usually be small
		// we don't need to optimize this yet.
		if _, ok := cfgs[sess.ID()]; !ok {
			session := ae.Sessions[sess.ID()]
			delete(ae.Sessions, sess.ID())
			deleted[session.ID()] = session
		}
	}
	return deleted
}

// AddSession idempotently adds a Session for the remote IA.
func (ae *ASEntry) AddSession(sessId mgmt.SessionType, polName string,
	afp *pktcls.ActionFilterPaths) error {
	ae.Lock()
	defer ae.Unlock()
	return ae.addSession(sessId, polName, afp)
}

func (ae *ASEntry) addSession(sessId mgmt.SessionType, polName string,
	afp *pktcls.ActionFilterPaths) error {
	if s, ok := ae.Sessions[sessId]; !ok {
		pool, err := policypathpool.NewPool(ae.IA, polName, afp)
		if err != nil {
			return err
		}
		// Session does not exist, so we create a new one
		s, err := session.NewSession(ae.IA, sessId, ae.Sigs, ae.Logger, pool, worker.DefaultFactory)
		if err != nil {
			return err
		}
		ae.Sessions[s.SessId] = s
		s.Start()
	} else {
		// Session exists, update its path pool
		pool := s.PathPool().(*policypathpool.Pool)
		if err := pool.Update(polName, afp); err != nil {
			return err
		}
	}
	if len(ae.Sessions) == 1 {
		ae.healthMonitorStop = make(chan struct{})
		go ae.monitorHealth()
	}
	return nil
}

// TODO(kormat): add DelSession
// len(ae.Sessions) == 0 { ae.healthMonitorStop <- struct{}{} }

func (ae *ASEntry) buildNewPktPolicies(cfgPktPols []*config.PktPolicy,
	classes pktcls.ClassMap) []*sessselector.PktPolicy {
	var newPktPolicies []*sessselector.PktPolicy
	for _, pol := range cfgPktPols {
		cls := classes[pol.ClassName]
		// Packet policies are stateless, so we construct new ones
		pp, err := sessselector.NewPktPolicy(pol.ClassName, cls, pol.SessIds, ae.Sessions)
		if err != nil {
			log.Error("Unable to create packet policy", "policy", pol, "err", err)
		}
		newPktPolicies = append(newPktPolicies, pp)
	}
	return newPktPolicies
}

func (ae *ASEntry) AddPktPolicy(name string, cls *pktcls.Class,
	sessIds []mgmt.SessionType) error {
	ae.Lock()
	defer ae.Unlock()
	return ae.addPktPolicy(name, cls, sessIds)
}

func (ae *ASEntry) addPktPolicy(name string, cls *pktcls.Class,
	sessIds []mgmt.SessionType) error {
	ppols := ae.PktPolicies.Load()
	for _, p := range ppols {
		// FIXME(kormat): support updating classes.
		if p.ClassName == name {
			return nil
		}
	}
	p, err := sessselector.NewPktPolicy(name, cls, sessIds, ae.Sessions)
	if err != nil {
		return err
	}
	ppols = append(ppols, p)
	ae.PktPolicies.Store(ppols)
	return nil
}

// manage the Sig map
func (ae *ASEntry) sigMgr() {
	defer log.LogPanicAndExit()
	ticker := time.NewTicker(sigMgrTick)
	defer ticker.Stop()
	ae.Info("sigMgr starting")
Top:
	for {
		// TODO(kormat): handle adding new SIGs from discovery, and updating existing ones.
		select {
		case <-ae.sigMgrStop:
			break Top
		case <-ticker.C:
			ae.Sigs.Range(func(id siginfo.SigIdType, sig *siginfo.Sig) bool {
				sig.ExpireFails()
				return true
			})
		}
	}
	close(ae.sigMgrStop)
	ae.sigMgrStop = nil
	ae.Info("sigMgr stopping")
}

func (ae *ASEntry) monitorHealth() {
	defer log.LogPanicAndExit()
	ticker := time.NewTicker(healthMonitorTick)
	defer ticker.Stop()
	ae.Info("Health monitor starting")
	prevHealth := false
	prevVersion := uint64(0)
Top:
	for {
		select {
		case <-ae.healthMonitorStop:
			break Top
		case <-ticker.C:
			ae.performHealthCheck(&prevHealth, &prevVersion)
		}
	}
	close(ae.healthMonitorStop)
	ae.healthMonitorStop = nil
	ae.Info("Health monitor stopping")
}

func (ae *ASEntry) performHealthCheck(prevHealth *bool, prevVersion *uint64) {
	ae.RLock()
	defer ae.RUnlock()
	curHealth := ae.checkHealth()
	if curHealth != *prevHealth || ae.version != *prevVersion {
		// Generate slice of networks.
		// XXX: This could become a bottleneck, namely in case of a large number
		// of remote prefixes and flappy health.
		nets := make([]*net.IPNet, 0, len(ae.Nets))
		for _, n := range ae.Nets {
			nets = append(nets, n)
		}
		// Overall health has changed. Generate event.
		params := base.RemoteHealthChangedParams{
			RemoteIA: ae.IA,
			Nets:     nets,
			Healthy:  curHealth,
		}
		base.RemoteHealthChanged(params)
	}
	*prevHealth = curHealth
	*prevVersion = ae.version
}

func (ae *ASEntry) checkHealth() bool {
	// If session 0 exists, determine overall health based on that session.
	defSess, ok := ae.Sessions[DefSessId]
	if ok {
		return defSess.Healthy()
	}
	// Otherwise, overall health is unhealthy if at least one session is unhealthy.
	for _, sess := range ae.Sessions {
		if !sess.Healthy() {
			return false
		}
	}
	return true
}

func (ae *ASEntry) Cleanup() error {
	ae.Lock()
	defer ae.Unlock()
	// Clean up sigMgr goroutine.
	ae.sigMgrStop <- struct{}{}
	// Clean up health monitor
	ae.healthMonitorStop <- struct{}{}
	// Clean up NetMap entries
	for _, v := range ae.Nets {
		if err := ae.delNet(v); err != nil {
			ae.Error("Error removing networks during cleanup", "err", err)
		}
	}
	ae.egressRing.Close()
	// Clean up sessions, and associated workers.
	ae.cleanSessions()
	return nil
}

func (ae *ASEntry) cleanSessions() {
	for _, s := range ae.Sessions {
		if err := s.Cleanup(); err != nil {
			s.Error("Error cleaning up session", "err", err)
		}
	}
}

func (ae *ASEntry) setupNet() {
	ae.egressRing = ringbuf.New(egress.EgressRemotePkts, nil, "egress",
		prometheus.Labels{"ringId": ae.IAString, "sessId": ""})
	ae.Info("Network setup done")
}
