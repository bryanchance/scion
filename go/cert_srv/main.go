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
	_ "net/http/pprof"
	"os"

	"github.com/BurntSushi/toml"

	"github.com/scionproto/scion/go/cert_srv/internal/csconfig"
	"github.com/scionproto/scion/go/cert_srv/internal/reiss"
	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/lib/consul"
	"github.com/scionproto/scion/go/lib/consul/consulconfig"
	"github.com/scionproto/scion/go/lib/env"
	"github.com/scionproto/scion/go/lib/fatal"
	"github.com/scionproto/scion/go/lib/infra/infraenv"
	"github.com/scionproto/scion/go/lib/infra/messenger"
	"github.com/scionproto/scion/go/lib/infra/modules/itopo"
	"github.com/scionproto/scion/go/lib/log"
	"github.com/scionproto/scion/go/lib/periodic"
	"github.com/scionproto/scion/go/lib/truststorage"
)

type Config struct {
	General env.General
	Sciond  env.SciondClient `toml:"sd_client"`
	Logging env.Logging
	Metrics env.Metrics
	TrustDB truststorage.TrustDBConf
	Infra   env.Infra
	CS      csconfig.Conf
	state   *csconfig.State

	Consul consulconfig.Config
}

var (
	config      Config
	environment *env.Env
	reissRunner *periodic.Runner
	msgr        *messenger.Messenger
)

var (
	ttlUpdater *periodic.Runner
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
	if v, ok := env.CheckFlags(csconfig.Sample); !ok {
		return v
	}
	if err := setupBasic(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	defer log.Flush()
	defer env.LogAppStopped(common.CS, config.General.ID)
	defer log.LogPanicAndExit()
	// Setup the state and the messenger
	if err := setup(); err != nil {
		log.Crit("Setup failed", "err", err)
		return 1
	}
	// Start the periodic reissuance task.
	startReissRunner()
	// Start the messenger.
	go func() {
		defer log.LogPanicAndExit()
		msgr.ListenAndServe()
	}()
	// Set environment to listen for signals.
	environment = infraenv.InitInfraEnvironmentFunc(config.General.TopologyPath, func() {
		if err := reload(); err != nil {
			log.Error("Unable to reload", "err", err)
		}
	})
	// Cleanup when the CS exits.
	defer stop()
	config.Metrics.StartPrometheus()
	// Start periodic health status setter
	go func() {
		defer log.LogPanicAndExit()
		var err error
		if ttlUpdater, err = startUpdateTTL(); err != nil {
			fatal.Fatal(err)
		}
	}()
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
	config.General.Topology = itopo.GetCurrentTopology()
	var newConf Config
	// Load new config to get the CS parameters.
	if _, err := toml.DecodeFile(env.ConfigFile(), &newConf); err != nil {
		return err
	}
	if err := newConf.CS.Init(config.General.ConfigDir); err != nil {
		return common.NewBasicError("Unable to initialize CS config", err)
	}
	config.CS = newConf.CS
	// Restart the periodic reissue task to respect the fresh parameters.
	stopReissRunner()
	startReissRunner()
	return nil
}

// startReissRunner starts a periodic reissuance task. Core starts self-issuer.
// Non-core starts a requester.
func startReissRunner() {
	if !config.CS.AutomaticRenewal {
		log.Info("Reissue disabled, not starting reiss task.")
		return
	}
	if config.General.Topology.Core {
		log.Info("Starting periodic reiss.Self task")
		reissRunner = periodic.StartPeriodicTask(
			&reiss.Self{
				Msgr:     msgr,
				State:    config.state,
				IA:       config.General.Topology.ISD_AS,
				IssTime:  config.CS.IssuerReissueLeadTime.Duration,
				LeafTime: config.CS.LeafReissueLeadTime.Duration,
			},
			periodic.NewTicker(config.CS.ReissueRate.Duration),
			config.CS.ReissueTimeout.Duration,
		)
		return
	}
	log.Info("Starting periodic reiss.Requester task")
	reissRunner = periodic.StartPeriodicTask(
		&reiss.Requester{
			Msgr:     msgr,
			State:    config.state,
			IA:       config.General.Topology.ISD_AS,
			LeafTime: config.CS.LeafReissueLeadTime.Duration,
		},
		periodic.NewTicker(config.CS.ReissueRate.Duration),
		config.CS.ReissueTimeout.Duration,
	)
}

func stopReissRunner() {
	if reissRunner != nil {
		reissRunner.Stop()
	}
}

func startUpdateTTL() (*periodic.Runner, error) {
	if !config.Consul.UpdateTTL {
		return nil, nil
	}
	config.Consul.InitDefaults()
	c, err := config.Consul.Client()
	if err != nil {
		return nil, err
	}
	svc := &consul.Service{
		Type:        consul.CS,
		Agent:       c.Agent(),
		Logger:      log.New("part", "UpdateTTL"),
		CheckParams: config.Consul.Health,
		Check: func() consul.CheckInfo {
			return consul.CheckInfo{
				Id:     config.General.ID,
				Status: consul.StatusPass,
			}
		},
	}
	return svc.StartUpdateTTL(config.Consul.InitConnPeriod.Duration)
}

func stop() {
	stopReissRunner()
	msgr.CloseServer()
}
