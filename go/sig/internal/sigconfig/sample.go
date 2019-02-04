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

package sigconfig

const Sample = `
[sig]
  # ID of the SIG (required)
  ID = "sig4"

  # The SIG config json file. (required)
  SIGConfig = "/etc/scion/sig/sig.json"

  # The local IA (required)
  IA = "1-ff00:0:113"

  # The bind IP address (required)
  IP = "192.0.2.100"

  # Control data port, e.g. keepalives. (default 10081)
  CtrlPort = 10081

  # Encapsulation data port. (default 10080)
  EncapPort = 10080

  # SCION dispatcher path. (default "")
  Dispatcher = ""

  # Name of TUN device to create. (default DefaultTunName)
  Tun = "sig"

  # Id of the routing table (default 11)
  TunRTableId = 11

[sig.Quagga]
  # Whether to export SIG routes to Zebra. (default false)
  ExportRoutes = false

  # The path to the zserv socket. (default /var/run/quagga/zserv.api)
  ZServApi = "/var/run/quagga/zserv.api"

[sd_client]
  # Sciond path. It defaults to sciond.DefaultSCIONDPath.
  Path = "/run/shm/sciond/default.sock"

  # Maximum time spent attempting to connect to sciond on start. (default 20s)
  InitialConnectPeriod = "20s"

[logging]
[logging.file]
  # Location of the logging file.
  Path = "/var/log/scion/sig4.log"

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
# The address to export prometheus metrics on. (default 127.0.0.1:1281)
  Prometheus = "127.0.0.1:8000"

[consul]
  # Enables consul. (default: false)
  Enabled = false

  # The consul agent to connect to. (default: 127.0.0.1:8500)
  Agent = "127.0.0.1:8500"

  # The maximum time the initial connection to consul can take. (default 5s)
  InitialConnectPeriod = "5s"

  [consul.Health]
    # The interval at which the health status should be reported to consul. (default 5s)
    Interval = "5s"
    # The timeout for setting the health status. (default 1s)
    Timeout = "1s"
`
