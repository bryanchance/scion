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
	DefaultAgent            = "127.0.0.1:8500"
	DefaultHealthTTL        = 10 * time.Second
	DefaultHealthInterval   = 5 * time.Second
	DefaultHealthTimeout    = 1 * time.Second
	DefaultHealthDeregister = 1 * time.Hour
	DefaultInitConnPeriod   = 5 * time.Second
)

type HealthCheck struct {
	// Name is the name of the health check.
	Name string
	// TTL is the TTL of the health check.
	TTL util.DurWrap
	// Interval is the time between setting check status.
	Interval util.DurWrap
	// Timeout is the timeout for setting check status.
	Timeout util.DurWrap
	// DeregisterCriticalServiceAfter specifies the time after which the service
	// associated with this check is deregistered.
	DeregisterCriticalServiceAfter util.DurWrap
}

func (h *HealthCheck) InitDefaults() {
	if h.TTL.Duration == 0 {
		h.TTL.Duration = DefaultHealthTTL
	}
	if h.Interval.Duration == 0 {
		h.Interval.Duration = DefaultHealthInterval
	}
	if h.Timeout.Duration == 0 {
		h.Timeout.Duration = DefaultHealthTimeout
	}
	if h.DeregisterCriticalServiceAfter.Duration == 0 {
		h.DeregisterCriticalServiceAfter.Duration = DefaultHealthDeregister
	}
}

type Config struct {
	Enabled        bool
	Agent          string
	Prefix         string
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
	// Name is the name of the service. Used as the session name in consul.
	Name string
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
