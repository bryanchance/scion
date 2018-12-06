// Copyright 2018 Anapaya Systems

package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	consulapi "github.com/hashicorp/consul/api"

	"github.com/scionproto/scion/go/lib/consul"
	"github.com/scionproto/scion/go/lib/env"
	"github.com/scionproto/scion/go/lib/log"
)

var (
	id    = flag.String("id", "", "Id to use")
	key   = flag.String("key", "", "Key to use in leader election")
	agent = flag.String("agent", "", "Consul agent")
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
	le := consul.StartLeaderElector(c, *key, consul.LeaderElectorConf{
		Interval: 100 * time.Millisecond,
		Timeout:  5 * time.Second,
		// Use a low lock delay for tests to have a faster handover.
		LockDelay:  1 * time.Second,
		SessionTTL: "1s",
	})
	log.Trace("election started")
	wasLeader := false
	runTicker := time.NewTicker(time.Second)
	for {
		if le.IsLeader() {
			if !wasLeader {
				log.Info("ISLEADER", "id", *id)
				wasLeader = true
			}
		} else {
			if wasLeader {
				log.Info("LOSTLEADER", "id", *id)
			}
			wasLeader = false
		}
		select {
		case <-environ.AppShutdownSignal:
			le.Stop()
			return 0
		case <-runTicker.C:
		}
	}
}
