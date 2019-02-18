// Copyright 2018 Anapaya Systems
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package config

const Sample = `[general]
  # The ID of the service. This is used to choose the relevant portion of the
  # topology file for some services.
  ID = "ps-1"

  # Directory for loading AS information, certs, keys, path policy, topology.
  ConfigDir = "/etc/scion"

  # Topology file. If not specified, topology.json is loaded from the config
  # directory.
  # Topology = "/etc/scion/topology.json"

  # ReconnectToDispatcher can be set to true to enable the snetproxy reconnecter.
  # ReconnectToDispatcher = true

[logging]
  [logging.file]
    # Location of the logging file.
    Path = "/var/log/scion/ps-1.log"

    # File logging level (trace|debug|info|warn|error|crit) (default debug)
    Level = "debug"

    # Max size of log file in MiB (default 50)
    # Size = 50

    # Max age of log file in days (default 7)
    # MaxAge = 7

    # MaxBackups is the maximum number of log files to retain (default 10)
    # MaxBackups = 10

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

[TrustDB]
  # The type of trustdb backend
  Backend = "sqlite"
  # Connection for the trust database
  Connection = "/var/lib/scion/spki/ps-1.trust.db"

[discovery]
  [discovery.static]
    # Enable periodic fetching of the static topology. (default false)
    Enable = false

    # Time between two consecutive static topology queries. (default 5m)
    Interval = "5m"

    # Timeout for querying the static topology. (default 1s)
    Timeout = "1s"

    # Require https connection. (default false)
    Https = false

    # Filename where the updated static topologies are written. In case of the
    # empty string, the updated topologies are not written. (default "")
    Filename = ""

    [discovery.static.connect]
      # Maximum time spent attempting to fetch the topology from the
      # discovery service on start. If no topology is successfully fetched
      # in this period, the FailAction is executed. (default 20s)
      InitialPeriod = "20s"

      # The action to take if no topology is successfully fetched in
      # the InitialPeriod.
      # - Fatal: Exit process.
      # - Continue: Log error and continue with execution.
      # (Fatal | Continue) (default Continue)
      FailAction = "Continue"

  [discovery.dynamic]
    # Enable periodic fetching of the dynamic topology. (default false)
    Enable = false

    # Time between two consecutive dynamic topology queries. (default 5s)
    Interval = "5s"

    # Timeout for querying the dynamic topology. (default 1s)
    Timeout = "1s"

    # Require https connection. (default false)
    Https = false

    [discovery.dynamic.connect]
      # Maximum time spent attempting to fetch the topology from the
      # discovery service on start. If no topology is successfully fetched
      # in this period, the FailAction is executed. (default 20s)
      InitialPeriod = "20s"

      # The action to take if no topology is successfully fetched in InitialPeriod.
      # - Fatal: Exit process.
      # - Continue: Log error and continue with execution.
      # (Fatal | Continue) (default Continue)
      FailAction = "Continue"

[ps]
  # Enable the "old" replication of down segments between cores using SegSync
  # messages (default false)
  SegSync = false

  # The time after which segments for a destination are refetched. (default 5m)
  QueryInterval = "5m"

  [ps.PathDB]
    # The type of pathdb backend
    Backend = "sqlite"
    # Path to the path database.
    Connection = "/var/lib/scion/pathdb/ps-1.path.db"

  [ps.RevCache]
    Backend = "mem"

[consul]
  # Enables consul. (default false)
  Enabled = false

  # The consul agent to connect to. (default 127.0.0.1:8500)
  Agent = "127.0.0.1:8500"

  # The prefix to use in front of the service name.
  # (e.g. "1-ff00:0:110/" in "1-ff00:0:110/PathService") (default "")
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
