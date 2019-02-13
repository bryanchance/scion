// Copyright 2018 Anapaya Systems

package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	consulapi "github.com/hashicorp/consul/api"

	"github.com/scionproto/scion/go/lib/consul"
	"github.com/scionproto/scion/go/lib/consul/consulconfig"
	"github.com/scionproto/scion/go/lib/env"
	"github.com/scionproto/scion/go/lib/log"
)

var (
	id    = flag.String("id", "", "Id to use")
	key   = flag.String("key", "", "Key to use in leader election")
	agent = flag.String("agent", "", "Consul agent")

	noClusterLeader = flag.Bool("noClusterLeader", false, "If no cluster leader become leader.")
)

func main() {
	os.Exit(realMain())
}

func realMain() int {
	log.AddLogConsFlags()
	flag.Parse()
	log.SetupFromFlags(fmt.Sprintf("leader_acceptance_%s", *id))
	environ := env.SetupEnv(func() {})
	cfg := consulapi.DefaultConfig()
	cfg.Address = *agent
	c, err := consulapi.NewClient(cfg)
	if err != nil {
		log.Crit("Error during setup", "err", err)
		return 1
	}
	le, err := consul.StartLeaderElector(c, *key, consulconfig.LeaderElectorConf{
		Name:    fmt.Sprintf("leader_acceptance_%s", *id),
		Timeout: 5 * time.Second,
		// Use a low lock delay for tests to have a faster handover.
		LockDelay:               1 * time.Second,
		SessionTTL:              "1s",
		LeaderIfNoClusterLeader: *noClusterLeader,
		AcquiredLeader: func() {
			log.Info("ISLEADER", "id", *id)
		},
		LostLeader: func() {
			log.Info("LOSTLEADER", "id", *id)
		},
	})
	if err != nil {
		log.Crit("Error during start of leader elector", "err", err)
		return 1
	}
	log.Trace("election started")
	select {
	case <-environ.AppShutdownSignal:
		le.Stop()
		return 0
	}
}
