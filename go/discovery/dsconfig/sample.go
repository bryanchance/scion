// Copyright 2018 Anapaya Systems

package dsconfig

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
    Level = "cirt"

[metrics]
  # The address to export prometheus metrics on. If not set, metrics are not
  # exported.
  # Prometheus = "127.0.0.1:8000"

[ds]
  # Address to serve on (host:port or ip:port or :port).
  ListenAddr = "127.0.0.1:8080"

  # Replace the topo timestamp with the file modification time for the
  # static topo. (default false) (reloadable)
  UseFileModTime = true

  # Path to ACL (reloadable).
  ACL = "/etc/scion/ds-acl"

  # The certificate file for TLS. If unset, serve plain HTTP.
  Cert = "/etc/scion/tls/cert"

  # Key file to use for TLS. If unset, serve plain HTTP.
  Key = "/etc/scion/tls/key"

  [ds.zoo]
    # Array of Zookeeper instances formated as "host:port". (reloadable)
    Instances = ["127.0.0.1:2181", "127.0.0.2:2181"]

    # Time between querying Zookeeper (default 5s) (reloadable).
    # QueryInterval = "5s"

    # Timeout for connecting to Zookeeper (default 10s) (reloadable).
    # Timeout = "10s"

`
