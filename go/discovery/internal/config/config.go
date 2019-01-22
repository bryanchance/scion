// Copyright 2018 Anapaya Systems

package config

import (
	"time"

	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/lib/consul/consulconfig"
	"github.com/scionproto/scion/go/lib/env"
	"github.com/scionproto/scion/go/lib/util"
)

const (
	// DefaultDynamicQueryInterval is the default interval between querying consul.
	DefaultDynamicQueryInterval = 5 * time.Second
	// DefaultDynamicTimeout is the default timeout for querying consul.
	DefaultDynamicTimeout = 2 * time.Second
)

type Config struct {
	General env.General
	Logging env.Logging
	Metrics env.Metrics
	Infra   env.Infra
	DS      DSConfig
	Consul  consulconfig.Config
}

func (c *Config) Validate() error {
	if c.General.ID == "" {
		return common.NewBasicError("No element ID specified", nil)
	}
	if err := c.DS.Validate(); err != nil {
		return err
	}
	return nil
}

type DSConfig struct {
	// UseFileModTime indicates if the file modification time is used for static topo timestamp.
	UseFileModTime bool
	// ACL is the ACL file for accessing the full topology.
	ACL string
	// Cert is the certificate file for TLS. If unset, serve plain HTTP.
	Cert string
	// Key is the key file to use for TLS. If unset, serve plain HTTP.
	Key string
	// ListenAddr is the address to serve on (host:port or ip:port or :port).
	ListenAddr string
	// Dynamic holds the parameters for the dynamic config.
	Dynamic DynConfig
}

func (d *DSConfig) InitDefaults() {
	d.Dynamic.InitDefaults()
}

func (d *DSConfig) Validate() error {
	if d.ACL == "" {
		return common.NewBasicError("No ACL file specified", nil)
	}
	if d.ListenAddr == "" {
		return common.NewBasicError("No address to listen on specified", nil)
	}
	return nil
}

type DynConfig struct {
	// QueryInterval is the time between querying consul.
	QueryInterval util.DurWrap
	// Timeout is the timeout for connecting to consul.
	Timeout util.DurWrap
	// ServicePrefix is the service prefix.
	ServicePrefix string
}

func (d *DynConfig) InitDefaults() {
	if d.QueryInterval.Duration == 0 {
		d.QueryInterval.Duration = DefaultDynamicQueryInterval
	}
	if d.Timeout.Duration == 0 {
		d.Timeout.Duration = DefaultDynamicTimeout
	}
}
