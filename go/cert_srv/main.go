// Copyright 2017 ETH Zurich
// Copyright 2018 ETH Zurich, Anapaya Systems
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
	"fmt"
	"io"
	_ "net/http/pprof"
	"os"
	"sync"
	"time"

	"github.com/BurntSushi/toml"

	"github.com/scionproto/scion/go/cert_srv/internal/config"
	"github.com/scionproto/scion/go/cert_srv/internal/reiss"
	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/lib/consul"
	"github.com/scionproto/scion/go/lib/consul/consulconfig"
	"github.com/scionproto/scion/go/lib/discovery"
	"github.com/scionproto/scion/go/lib/env"
	"github.com/scionproto/scion/go/lib/fatal"
	"github.com/scionproto/scion/go/lib/infra"
	"github.com/scionproto/scion/go/lib/infra/infraenv"
	"github.com/scionproto/scion/go/lib/infra/modules/idiscovery"
	"github.com/scionproto/scion/go/lib/infra/modules/itopo"
	"github.com/scionproto/scion/go/lib/log"
	"github.com/scionproto/scion/go/lib/periodic"
)

var (
	cfg         config.Config
	state       *config.State
	environment *env.Env
	reissRunner *periodic.Runner
	discRunners idiscovery.Runners
	reissMtx    sync.Mutex
	corePusher  *periodic.Runner
	msgr        infra.Messenger
)

var (
	leaderElector consul.LeaderElector
)

func init() {
	flag.Usage = env.Usage
}

// main initializes the certificate server and starts the dispatcher.
func main() {
	os.Exit(realMain())
}

func realMain() int {
	fatal.Init()
	env.AddFlags()
	flag.Parse()
	if v, ok := env.CheckFlags(config.Sample); !ok {
		return v
	}
	if err := setupBasic(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	defer log.Flush()
	defer env.LogAppStopped(common.CS, cfg.General.ID)
	defer log.LogPanicAndExit()
	// Setup the state and the messenger
	if err := setup(); err != nil {
		log.Crit("Setup failed", "err", err)
		return 1
	}
	if cfg.Consul.Enabled {
		c, err := setupConsul()
		if err != nil {
			log.Crit("Setup consul failed", "err", err)
			return 1
		}
		defer c.Close()
	} else {
		startReissRunner()
	}
	// Start the periodic fetching from discovery service.
	startDiscovery()
	// Start the messenger.
	go func() {
		defer log.LogPanicAndExit()
		msgr.ListenAndServe()
	}()
	// Set environment to listen for signals.
	environment = infraenv.InitInfraEnvironmentFunc(cfg.General.TopologyPath, func() {
		if err := reload(); err != nil {
			log.Error("Unable to reload", "err", err)
		}
	})
	// Cleanup when the CS exits.
	defer stop()
	cfg.Metrics.StartPrometheus()
	select {
	case <-environment.AppShutdownSignal:
		// Whenever we receive a SIGINT or SIGTERM we exit without an error.
		return 0
	case <-fatal.Chan():
		return 1
	}
}

// reload reloads the topology and CS config.
func reload() error {
	// FIXME(roosd): KeyConf reloading is not yet supported.
	// https://github.com/scionproto/scion/issues/2077
	cfg.General.Topology = itopo.Get()
	var newConf config.Config
	// Load new config to get the CS parameters.
	if _, err := toml.DecodeFile(env.ConfigFile(), &newConf); err != nil {
		return err
	}
	if err := newConf.CS.Init(cfg.General.ConfigDir); err != nil {
		return common.NewBasicError("Unable to initialize CS config", err)
	}
	cfg.CS = newConf.CS
	if cfg.Consul.Enabled {
		stopLeaderElector()
		if err := startLeaderElector(); err != nil {
			err = common.NewBasicError("Failed to start leader election", err)
			fatal.Fatal(err)
		}
	} else {
		// Restart the periodic reissue task to respect the fresh parameters.
		stopReissRunner()
		startReissRunner()
	}
	return nil
}

// startReissRunner starts a periodic reissuance task. Core starts self-issuer.
// Non-core starts a requester.
func startReissRunner() {
	reissMtx.Lock()
	defer reissMtx.Unlock()
	corePusher = periodic.StartPeriodicTask(
		&reiss.CorePusher{
			LocalIA: cfg.General.Topology.ISD_AS,
			TrustDB: state.TrustDB,
			Msger:   msgr,
		},
		periodic.NewTicker(time.Hour),
		time.Minute,
	)
	corePusher.TriggerRun()
	if !cfg.CS.AutomaticRenewal {
		log.Info("Reissue disabled, not starting reiss task.")
		return
	}
	if cfg.General.Topology.Core {
		log.Info("Starting periodic reiss.Self task")
		reissRunner = periodic.StartPeriodicTask(
			&reiss.Self{
				Msgr:       msgr,
				State:      state,
				IA:         cfg.General.Topology.ISD_AS,
				IssTime:    cfg.CS.IssuerReissueLeadTime.Duration,
				LeafTime:   cfg.CS.LeafReissueLeadTime.Duration,
				CorePusher: corePusher,
			},
			periodic.NewTicker(cfg.CS.ReissueRate.Duration),
			cfg.CS.ReissueTimeout.Duration,
		)
		return
	}
	log.Info("Starting periodic reiss.Requester task")
	reissRunner = periodic.StartPeriodicTask(
		&reiss.Requester{
			Msgr:       msgr,
			State:      state,
			IA:         cfg.General.Topology.ISD_AS,
			LeafTime:   cfg.CS.LeafReissueLeadTime.Duration,
			CorePusher: corePusher,
		},
		periodic.NewTicker(cfg.CS.ReissueRate.Duration),
		cfg.CS.ReissueTimeout.Duration,
	)
}

func startDiscovery() {
	var err error
	discRunners, err = idiscovery.StartRunners(cfg.Discovery, discovery.Full,
		idiscovery.TopoHandlers{}, nil)
	if err != nil {
		fatal.Fatal(common.NewBasicError("Unable to start dynamic topology fetcher", err))
	}
}

func killReissRunner() {
	reissMtx.Lock()
	defer reissMtx.Unlock()
	if corePusher != nil {
		corePusher.Kill()
	}
	if reissRunner != nil {
		reissRunner.Kill()
		reissRunner = nil
	}
}

func stopReissRunner() {
	reissMtx.Lock()
	defer reissMtx.Unlock()
	if corePusher != nil {
		corePusher.Stop()
	}
	if reissRunner != nil {
		reissRunner.Stop()
		reissRunner = nil
	}
}

func stop() {
	if cfg.Consul.Enabled {
		stopLeaderElector()
	} else {
		stopReissRunner()
	}
	discRunners.Stop()
	msgr.CloseServer()
}

type closerFunc func() error

func (cf closerFunc) Close() error {
	return cf()
}

func setupConsul() (io.Closer, error) {
	fatal.Check()
	cfg.Consul.InitDefaults()
	c, err := cfg.Consul.Client()
	if err != nil {
		return nil, err
	}
	topoAddr := itopo.Get().CS.GetById(cfg.General.ID)
	svc := &consul.Service{
		Agent:  c.Agent(),
		ID:     cfg.General.ID,
		Prefix: cfg.Consul.Prefix,
		Addr:   topoAddr.PublicAddr(topoAddr.Overlay),
		Type:   consul.CS,
	}
	ttlUpdater, err := svc.Register(cfg.Consul.InitConnPeriod.Duration,
		checkHealth, cfg.Consul.Health)
	if err != nil {
		return nil, common.NewBasicError("Unable to register service with consul", err)
	}
	if err = startLeaderElector(); err != nil {
		ttlUpdater.Kill()
		svc.Deregister()
		return nil, err
	}
	return closerFunc(func() error {
		ttlUpdater.Stop()
		svc.Deregister()
		return nil
	}), nil
}

func startLeaderElector() error {
	cfg.Consul.InitDefaults()
	c, err := cfg.Consul.Client()
	if err != nil {
		return err
	}
	leaderKey := fmt.Sprintf("cert_srv/leader/%s", cfg.General.Topology.ISD_AS.FileFmt(false))
	leaderElector, err = consul.StartLeaderElector(c, leaderKey, consulconfig.LeaderElectorConf{
		Name:           cfg.General.ID,
		AcquiredLeader: startReissRunner,
		LostLeader:     killReissRunner,
	})
	return err
}

func stopLeaderElector() {
	leaderElector.Stop()
	leaderElector = nil
}

func checkHealth() consul.CheckInfo {
	return consul.CheckInfo{
		Status: consul.StatusPass,
	}
}
