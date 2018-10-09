// Copyright 2018 Anapaya Systems

package dsconfig

import (
	"time"

	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/lib/util"
)

const (
	ZooQueryInterval = 5 * time.Second
	ZooTimeout       = 10 * time.Second
)

type Config struct {
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
	// Zoo holds the zookeeper config.
	Zoo ZooConfig
}

func (c *Config) InitDefaults() {
	c.Zoo.InitDefaults()
}

func (c *Config) Validate() error {
	if c.ACL == "" {
		return common.NewBasicError("No ACL file specified", nil)
	}
	if c.ListenAddr == "" {
		return common.NewBasicError("No address to listen on specified", nil)
	}

	if err := c.Zoo.Validate(); err != nil {
		return err
	}
	return nil
}

type ZooConfig struct {
	// Instances is a list of Zookeeper instances formated as "host:port".
	Instances []string
	// QueryInterval is the time between querying Zookeeper.
	QueryInterval util.DurWrap
	// Timeout is the timeout for connecting to Zookeeper.
	Timeout util.DurWrap
}

func (c *ZooConfig) InitDefaults() {
	if c.QueryInterval.Duration == 0 {
		c.QueryInterval.Duration = ZooQueryInterval
	}
	if c.Timeout.Duration == 0 {
		c.Timeout.Duration = ZooTimeout
	}
}

func (c *ZooConfig) Validate() error {
	if len(c.Instances) <= 0 {
		return common.NewBasicError("No Zookeeper specified", nil)
	}
	return nil
}
