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

package main

import (
	"flag"
	"net"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"

	log "github.com/inconshreveable/log15"

	"github.com/netsec-ethz/scion/go/lib/addr"
	"github.com/netsec-ethz/scion/go/lib/common"
	liblog "github.com/netsec-ethz/scion/go/lib/log"
	"github.com/netsec-ethz/scion/go/lib/pathmgr"
	"github.com/netsec-ethz/scion/go/lib/pktcls"
	"github.com/netsec-ethz/scion/go/sig/base"
	"github.com/netsec-ethz/scion/go/sig/config"
	"github.com/netsec-ethz/scion/go/sig/disp"
	"github.com/netsec-ethz/scion/go/sig/egress"
	"github.com/netsec-ethz/scion/go/sig/ingress"
	"github.com/netsec-ethz/scion/go/sig/metrics"
	"github.com/netsec-ethz/scion/go/sig/sigcmn"
)

var (
	id      = flag.String("id", "", "Element ID (Required. E.g. 'sig4-21-9')")
	cfgPath = flag.String("config", "", "Config file (Required)")
	isdas   = flag.String("ia", "", "Local AS (Required, e.g., 1-10)")
	ipStr   = flag.String("ip", "", "address to bind to (Required)")
)

func main() {
	flag.Parse()
	if *id == "" {
		log.Crit("No element ID specified")
		flag.Usage()
		os.Exit(1)
	}
	liblog.Setup(*id)
	defer liblog.LogPanicAndExit()
	setupSignals()

	// Export prometheus metrics.
	metrics.Init(*id)
	if err := metrics.Start(); err != nil {
		fatal("Unable to export prometheus metrics", "err", err)
	}
	// Parse basic flags
	ia, err := addr.IAFromString(*isdas)
	if err != nil {
		fatal("Unable to parse local ISD-AS", "ia", *isdas, "err", err)
	}
	ip := net.ParseIP(*ipStr)
	if ip == nil {
		fatal("unable to parse IP address", "addr", *ipStr)
	}
	if err = sigcmn.Init(ia, ip); err != nil {
		fatal("Error during initialization", "err", err)
	}
	egress.Init()
	disp.Init(sigcmn.CtrlConn)
	go base.PollReqHdlr()
	// Parse config
	if loadConfig(*cfgPath) != true {
		fatal("Unable to load config on startup")
	}

	// Spawn ingress Dispatcher.
	if err := ingress.Init(); err != nil {
		fatal("Unable to spawn ingress dispatcher", "err", err)
	}
}

func setupSignals() {
	sig := make(chan os.Signal, 2)
	signal.Notify(sig, os.Interrupt)
	signal.Notify(sig, syscall.SIGTERM)
	go func() {
		s := <-sig
		log.Info("Received signal, exiting...", "signal", s)
		liblog.Flush()
		os.Exit(1)
	}()
}

func loadConfig(path string) bool {
	cfg, err := config.LoadFromFile(path)
	if err != nil {
		cerr := err.(*common.CError)
		log.Error(cerr.Desc, cerr.Ctx...)
		return false
	}
	success := true
	for iaVal, cfgEntry := range cfg.ASes {
		ia := &iaVal
		ae, err := base.Map.AddIA(ia)
		if err != nil {
			cerr := err.(*common.CError)
			log.Error(cerr.Desc, cerr.Ctx...)
			success = false
			continue
		}
		// Add sigs before networks, so there's somewhere for packets to go.
		for id, sig := range cfgEntry.Sigs {
			ctrlPort := int(sig.CtrlPort)
			if ctrlPort == 0 {
				ctrlPort = sigcmn.DefaultCtrlPort
			}
			encapPort := int(sig.EncapPort)
			if encapPort == 0 {
				encapPort = sigcmn.DefaultEncapPort
			}
			if err := ae.AddSig(id, sig.Addr, ctrlPort, encapPort, true); err != nil {
				cerr := err.(*common.CError)
				log.Error(cerr.Desc, cerr.Ctx...)
				success = false
				continue
			}
		}
		for _, netw := range cfgEntry.Nets {
			if err := ae.AddNet(netw.IPNet()); err != nil {
				cerr := err.(*common.CError)
				log.Error(cerr.Desc, cerr.Ctx...)
				success = false
				continue
			}
		}
		if !loadSessions(ae, cfg.Actions, cfgEntry.Sessions) {
			success = false
			continue
		}
		if !loadPktPols(ae, cfg.Classes, cfgEntry.PktPolicies) {
			success = false
			continue
		}
	}
	return success
}

func loadSessions(ae *base.ASEntry, actions pktcls.ActionMap, cfgSessMap config.SessionMap) bool {
	success := true
	for sessId, actName := range cfgSessMap {
		act := actions[actName]
		var pred *pathmgr.PathPredicate
		if afp, ok := act.(*pktcls.ActionFilterPaths); ok {
			pred = afp.Contains
		}
		if err := ae.AddSession(sessId, actName, pred); err != nil {
			cerr := err.(*common.CError)
			log.Error(cerr.Desc, cerr.Ctx...)
			success = false
			continue
		}
	}
	return success
}

func loadPktPols(ae *base.ASEntry, classes pktcls.ClassMap, cfgPktPols []*config.PktPolicy) bool {
	success := true
	for _, pol := range cfgPktPols {
		cls := classes[pol.ClassName]
		if err := ae.AddPktPolicy(pol.ClassName, cls, pol.SessIds); err != nil {
			cerr := err.(*common.CError)
			log.Error(cerr.Desc, cerr.Ctx...)
			success = false
			continue
		}
	}
	return success
}

func fatal(msg string, args ...interface{}) {
	log.Crit(msg, args...)
	liblog.Flush()
	os.Exit(1)
}
