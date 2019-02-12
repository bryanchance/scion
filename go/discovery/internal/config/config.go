// Copyright 2018 Anapaya Systems

package config

import (
	"strings"
	"time"

	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/lib/consul/consulconfig"
	"github.com/scionproto/scion/go/lib/env"
	"github.com/scionproto/scion/go/lib/util"
)

const (
	// DefaultDynamicQueryInterval is the default interval between querying consul.
	DefaultDynamicQueryInterval = 1 * time.Second
	// DefaultDynamicTimeout is the default timeout for querying consul.
	DefaultDynamicTimeout = 1 * time.Second
	// DefaultDynamicTTL is the default TTL of the dynamic topology.
	DefaultDynamicTTL = 1 * time.Minute
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
	// TTL is the TTL of the dynamic topology.
	TTL util.DurWrap
	// NoConsulConnAction indicates what action shall be taken if there is no
	// connection to the local consul agent when building the dynamic topology.
	NoConsulConnAction NoConsulConnAction
}

func (d *DynConfig) InitDefaults() {
	if d.QueryInterval.Duration == 0 {
		d.QueryInterval.Duration = DefaultDynamicQueryInterval
	}
	if d.Timeout.Duration == 0 {
		d.Timeout.Duration = DefaultDynamicTimeout
	}
	if d.TTL.Duration == 0 {
		d.TTL.Duration = DefaultDynamicTTL
	}
	if d.NoConsulConnAction != NoConsulConnError {
		d.NoConsulConnAction = NoConsulConnFallback
	}
}

// NoConsulConnAction indicates the action that shall be taken if there is
// no connection to the consul agent when building the dynamic topology.
type NoConsulConnAction string

const (
	// NoConsulConnError indicates that an error is served, instead of the dynamic
	// topology, when no connection to the consul agent is possible.
	NoConsulConnError NoConsulConnAction = "Error"
	// NoConsulConnFallback indicates that the static topology is served as the
	// dynamic topology when no connection to the consul agent is possible.
	NoConsulConnFallback NoConsulConnAction = "Fallback"
)

func (f *NoConsulConnAction) UnmarshalText(text []byte) error {
	switch strings.ToLower(string(text)) {
	case strings.ToLower(string(NoConsulConnError)):
		*f = NoConsulConnError
	case strings.ToLower(string(NoConsulConnFallback)):
		*f = NoConsulConnFallback
	default:
		return common.NewBasicError("Unknown FailAction", nil, "input", string(text))
	}
	return nil
}
