// Copyright 2018 Anapaya Systems

// Package config contains the configuration of the IPProvider.
package config

import (
	"net"

	"github.com/scionproto/scion/go/lib/addr"
	"github.com/scionproto/scion/go/lib/env"
)

const Sample = `
[logging]
  [logging.file]
    # Location of the logging file.
    Path = "/var/log/scion/ipscraper.log"
    # File logging level (trace|debug|info|warn|error|crit) (default debug)
    Level = "debug"
    # Max size of log file in MiB (default 50)
    # Size = 50
    # Max age of log file in days (default 7)
    # MaxAge = 7
    # How frequently to flush to the log file, in seconds. If 0, all messages
    # are immediately flushed. If negative, messages are never flushed
    # automatically. (default 5)
    FlushInterval = 5
  [logging.console]
    # Console logging level (trace|debug|info|warn|error|crit) (default crit)
    Level = "crit"
[sd_client]
  # Sciond path. It defaults to sciond.DefaultSCIONDPath.
  Path = "/run/shm/sciond/default.sock"
  # Maximum time spent attempting to connect to sciond on start. (default 20s)
  InitialConnectPeriod = "20s"
[ipprovider]
  # Local ISD and AS.
  LocalIA = "64-2:0:0"
  # Local IP address.
  LocalAddr = "192.168.0.111"
  # Local port to bind to.
  LocalPort = 5555
  # Path to the sigmgmt's SQLite3 database.
  DBPath = "testdb"
`

type Config struct {
	Logging    env.Logging
	Sciond     env.SciondClient `toml:"sd_client"`
	IPProvider IPProviderConfig
}

type IPProviderConfig struct {
	// IA the local IA (required).
	LocalIA addr.IA
	// Local address to bind to (required).
	LocalAddr net.IP
	// Local port to bind to (required).
	LocalPort uint16
	// Path to the sigmgmt's SQLite3 database.
	DBPath string
}
