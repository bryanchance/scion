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

package base

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	log "github.com/inconshreveable/log15"
	"github.com/vishvananda/netlink"

	"github.com/scionproto/scion/go/lib/addr"
	"github.com/scionproto/scion/go/lib/common"
	liblog "github.com/scionproto/scion/go/lib/log"
	"github.com/scionproto/scion/go/lib/pathmgr"
	"github.com/scionproto/scion/go/lib/pktcls"
	"github.com/scionproto/scion/go/sig/config"
	"github.com/scionproto/scion/go/sig/egress"
	"github.com/scionproto/scion/go/sig/mgmt"
	"github.com/scionproto/scion/go/sig/sigcmn"
	"github.com/scionproto/scion/go/sig/siginfo"
	"github.com/scionproto/scion/go/sig/xnet"
)

const sigMgrTick = 10 * time.Second

// ASEntry contains all of the information required to interact with a remote AS.
type ASEntry struct {
	sync.RWMutex
	Nets        map[string]*NetEntry
	Sigs        *siginfo.SigMap
	IA          *addr.ISD_AS
	IAString    string
	PktPolicies *egress.SyncPktPols
	Sessions    egress.SessionSet
	DevName     string
	tunLink     netlink.Link
	tunIO       io.ReadWriteCloser
	sigMgrStop  chan struct{}
	log.Logger
}

func newASEntry(ia *addr.ISD_AS) (*ASEntry, error) {
	ae := &ASEntry{
		Logger:      log.New("ia", ia),
		IA:          ia,
		IAString:    ia.String(),
		Nets:        make(map[string]*NetEntry),
		Sigs:        &siginfo.SigMap{},
		PktPolicies: egress.NewSyncPktPols(),
		Sessions:    make(egress.SessionSet),
		DevName:     fmt.Sprintf("scion-%s", ia),
		sigMgrStop:  make(chan struct{}),
	}
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
		ae.Error("Unable to update sessions", "err", common.FmtError(err))
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
			ae.Error("Unable to add network", "net", ipnet, "err", common.FmtError(err))
			s = false
		}
	}
	return s
}

// delOldNets deletes currently configured networks that are not in ipnets.
func (ae *ASEntry) delOldNets(ipnets []*config.IPNet) bool {
	s := true
Top:
	for _, ne := range ae.Nets {
		for _, ipnet := range ipnets {
			if ne.Net.String() == ipnet.IPNet().String() {
				continue Top
			}
		}
		err := ae.delNet(ne.Net)
		if err != nil {
			ae.Error("Unable to delete network", "NetEntry", ne, "err", common.FmtError(err))
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
	if ae.tunLink == nil {
		// Ensure that the network setup is done, as otherwise route entries can't be added.
		if err := ae.setupNet(); err != nil {
			return err
		}
	}
	key := ipnet.String()
	if _, ok := ae.Nets[key]; ok {
		return nil
	}
	ne, err := newNetEntry(ae.tunLink, ipnet)
	if err != nil {
		return err
	}
	ae.Nets[key] = ne
	ae.Info("Added network", "net", ipnet)
	return nil
}

// DelIA removes a network for the remote IA.
func (ae *ASEntry) DelNet(ipnet *net.IPNet) error {
	ae.Lock()
	defer ae.Unlock()
	return ae.delNet(ipnet)
}

// DelIA removes a network for the remote IA.
func (ae *ASEntry) delNet(ipnet *net.IPNet) error {
	key := ipnet.String()
	ne, ok := ae.Nets[key]
	if !ok {
		ae.Unlock()
		return common.NewBasicError("DelNet: no network found", nil, "ia", ae.IA, "net", ipnet)
	}
	delete(ae.Nets, key)
	ae.Info("Removed network", "net", ipnet)
	return ne.Cleanup()
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
			ae.Error("Unable to add SIG", "sig", sig, "err", common.FmtError(err))
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
				ae.Error("Unable to delete SIG", "err", common.FmtError(err))
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
			ae.Error("Unable to add session", "id", cfg.ID, "err", common.FmtError(err))
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
		if _, ok := cfgs[sess.SessId]; !ok {
			session := ae.Sessions[sess.SessId]
			delete(ae.Sessions, sess.SessId)
			deleted[session.SessId] = session
		}
	}
	return deleted
}

// AddSession idempotently adds a Session for the remote IA.
func (ae *ASEntry) AddSession(sessId mgmt.SessionType, polName string,
	pathPred *pathmgr.PathPredicate) error {
	ae.Lock()
	defer ae.Unlock()
	return ae.addSession(sessId, polName, pathPred)
}

func (ae *ASEntry) addSession(sessId mgmt.SessionType, polName string,
	pathPred *pathmgr.PathPredicate) error {
	if s, ok := ae.Sessions[sessId]; !ok {
		// Session does not exist, so we create a new one
		s, err := egress.NewSession(ae.IA, sessId, ae.Sigs, ae.Logger, polName, pathPred)
		if err != nil {
			return err
		}
		ae.Sessions[s.SessId] = s
		s.Start()
	} else {
		// Session exists, update its information
		if err := s.UpdatePolicy(polName, pathPred); err != nil {
			return err
		}
	}
	if len(ae.Sessions) == 1 {
		go egress.NewDispatcher(ae.DevName, ae.tunIO, ae.PktPolicies).Run()
		go ae.sigMgr()
	}
	return nil
}

// TODO(kormat): add DelSession, and set tun device down there's no sessions left.

func (ae *ASEntry) buildNewPktPolicies(cfgPktPols []*config.PktPolicy, classes pktcls.ClassMap) []*egress.PktPolicy {
	var newPktPolicies []*egress.PktPolicy
	for _, pol := range cfgPktPols {
		cls := classes[pol.ClassName]
		// Packet policies are stateless, so we construct new ones
		pp, err := egress.NewPktPolicy(pol.ClassName, cls, pol.SessIds, ae.Sessions)
		if err != nil {
			log.Error("Unable to create packet policy", "policy", pol, "err", common.FmtError(err))
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
	p, err := egress.NewPktPolicy(name, cls, sessIds, ae.Sessions)
	if err != nil {
		return err
	}
	ppols = append(ppols, p)
	ae.PktPolicies.Store(ppols)
	return nil
}

// manage the Sig map
func (ae *ASEntry) sigMgr() {
	defer liblog.LogPanicAndExit()
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
	ae.Info("sigMgr stopping")
}

func (ae *ASEntry) Cleanup() error {
	ae.Lock()
	defer ae.Unlock()
	// Clean up sigMgr goroutine.
	ae.sigMgrStop <- struct{}{}
	// Clean up the egress dispatcher.
	if err := ae.tunIO.Close(); err != nil {
		ae.Error("Error closing TUN io", "dev", ae.DevName, "err", common.FmtError(err))
	}
	// Clean up sessions, and associated workers.
	ae.cleanSessions()
	// The operating system also removes the routes when deleting the link.
	if err := netlink.LinkDel(ae.tunLink); err != nil {
		// Only return this error, as it's the only critical one.
		return common.NewBasicError("Error removing TUN link", err,
			"ia", ae.IA, "dev", ae.DevName)
	}
	return nil
}

func (ae *ASEntry) cleanSessions() {
	for _, s := range ae.Sessions {
		if err := s.Cleanup(); err != nil {
			s.Error("Error cleaning up session", "err", common.FmtError(err))
		}
	}
}

func (ae *ASEntry) setupNet() error {
	var err error
	ae.tunLink, ae.tunIO, err = xnet.ConnectTun(ae.DevName)
	if err != nil {
		return err
	}
	ae.Info("Network setup done")
	return nil
}
