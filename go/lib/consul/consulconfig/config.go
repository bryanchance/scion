// Copyright 2018 Anapaya Systems

package consulconfig

import (
	"time"

	consulapi "github.com/hashicorp/consul/api"

	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/lib/util"
)

const (
	DefaultLETimeout = 5 * time.Minute
)

var (
	DefaultAgent          = "127.0.0.1:8500"
	DefaultHealthInterval = 5 * time.Second
	DefaultHealthTimeout  = 1 * time.Second
	DefaultInitConnPeriod = 5 * time.Second
)

type HealthCheck struct {
	// Interval is the time between setting check status.
	Interval util.DurWrap
	// Timeout is the timeout for setting check status.
	Timeout util.DurWrap
}

func (h *HealthCheck) InitDefaults() {
	if h.Interval.Duration == 0 {
		h.Interval.Duration = DefaultHealthInterval
	}
	if h.Timeout.Duration == 0 {
		h.Timeout.Duration = DefaultHealthTimeout
	}
}

type Config struct {
	Agent          string
	UpdateTTL      bool
	Health         HealthCheck
	InitConnPeriod util.DurWrap `toml:"InitialConnectPeriod"`
}

func (c *Config) InitDefaults() {
	if c.Agent == "" {
		c.Agent = DefaultAgent
	}
	if c.InitConnPeriod.Duration == 0 {
		c.InitConnPeriod.Duration = DefaultInitConnPeriod
	}
	c.Health.InitDefaults()
}

func (c *Config) Client() (*consulapi.Client, error) {
	cfg := consulapi.DefaultConfig()
	cfg.Address = c.Agent
	return consulapi.NewClient(cfg)
}

type LeaderElectorConf struct {
	// Timeout indicates how long the leader elector should blockingly wait for leader changes.
	// If not set the default DefaultLETimeout is used.
	Timeout time.Duration
	// LockDelay is the lock delay passed to the consul API.
	// If set to zero the default of the consul API is used.
	// See also: https://www.consul.io/api/session.html#lockdelay
	LockDelay time.Duration
	// SessionTTL is the TTL of the session. By default this should be 10s.
	// See https://www.consul.io/api/session.html#ttl
	SessionTTL string
	// AcquiredLeader is called if leadership is acquired.
	// AcquiredLeader should be non-blocking/short-running
	AcquiredLeader func()
	// LostLeader is called if leadership is lost. LostLeader is allowed to block.
	// As long as LostLeader blocks, leadership is not acquired.
	LostLeader func()
}

func (c *LeaderElectorConf) Validate() error {
	if c.AcquiredLeader == nil || c.LostLeader == nil {
		return common.NewBasicError("Acquired-/LostLeader has to be set", nil)
	}
	return nil
}

func (c *LeaderElectorConf) InitDefaults() {
	if c.Timeout == 0 {
		c.Timeout = DefaultLETimeout
	}
	if c.SessionTTL == "" {
		c.SessionTTL = "10s" // minimum (see https://www.consul.io/api/session.html#ttl)
	}
}
