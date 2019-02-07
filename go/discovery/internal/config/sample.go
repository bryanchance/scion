// Copyright 2018 Anapaya Systems

package config

const Sample = `[general]
  # The ID of the service.
  ID = "ds-1"

  # Directory for loading AS information. (reloadable)
  ConfigDir = "/etc/scion"

  # Topology file. If not specified, topology.json is loaded from the config
  # directory (reloadable).
  # Topology = "/etc/scion/topology.json"

[logging]
  [logging.file]
    # Location of the logging file.
    Path = "/var/log/scion/ds-1.log"

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

[metrics]
  # The address to export prometheus metrics on. If not set, metrics are not
  # exported.
  # Prometheus = "127.0.0.1:8000"

[ds]
  # Address to serve on (host:port or ip:port or :port).
  ListenAddr = "127.0.0.1:8080"

  # Replace the topo timestamp with the file modification time for the
  # static topo. (default false) (reloadable)
  UseFileModTime = false

  # Path to ACL (reloadable).
  ACL = "/etc/scion/ds-acl"

  # The certificate file for TLS. If unset, serve plain HTTP.
  Cert = "/etc/scion/tls/cert"

  # Key file to use for TLS. If unset, serve plain HTTP.
  Key = "/etc/scion/tls/key"

  [ds.dynamic]
    # Time between querying consul (default 1s) (reloadable)
    QueryInterval = "1s"

    # Timeout for query to consul (default 1s) (reloadable)
    Timeout = "1s"

    # The service prefix in the service name. (default "") (reloadable)
    ServicePrefix = ""

    # The TTL of the served dynamic topology. (default "60s") (reloadable)
    TTL = "60s"

[consul]
  # Enables consul. (default false)
  Enabled = false

  # The consul agent to connect to. (default 127.0.0.1:8500)
  Agent = "127.0.0.1:8500"

  # The prefix to use in front of the service name.
  # (e.g. "1-ff00:0:110/" in "1-ff00:0:110/DiscoveryService") (default "")
  Prefix = ""

  # The maximum time the initial connection to consul can take. (default 5s)
  InitialConnectPeriod = "5s"

  [consul.Health]
    # Optional name of the health check. The empty string is replace by
    # "Health Check: ID", where ID is general.ID. (default "")
    Name = ""

    # TTL is the TTL of the health check. (default 2s)
    TTL = "2s"

    # The interval at which the health status should be reported to consul. (default 500ms)
    Interval = "500ms"

    # The timeout for setting the health status. (default 500ms)
    Timeout = "500ms"

    # Deregister the service if the check is in critical state
    # for more than this time. (default 1h)
    DeregisterCriticalServiceAfter = "1h"
`
