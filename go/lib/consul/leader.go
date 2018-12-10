// Copyright 2018 Anapaya Systems

package consul

import (
	consulapi "github.com/hashicorp/consul/api"

	"github.com/scionproto/scion/go/lib/consul/consulconfig"
	"github.com/scionproto/scion/go/lib/consul/internal/leader"
)

// LeaderElector handles leader election for a specific key.
type LeaderElector interface {
	// Stop releases the leader and stops the process.
	Stop()
	// Key returns the key for which this LeaderElector is running.
	Key() string
}

// StartLeaderElector starts a new LeaderElector for the given key.
func StartLeaderElector(c *consulapi.Client, key string,
	conf consulconfig.LeaderElectorConf) (LeaderElector, error) {

	if err := conf.Validate(); err != nil {
		return nil, err
	}
	return leader.StartSessMon(key, c, conf), nil
}
