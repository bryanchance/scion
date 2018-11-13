// Copyright 2018 Anapaya Systems

package consulconfig

import (
	"time"

	consulapi "github.com/hashicorp/consul/api"

	"github.com/scionproto/scion/go/lib/util"
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
