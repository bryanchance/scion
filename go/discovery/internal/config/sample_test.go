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

import (
	"testing"

	"github.com/BurntSushi/toml"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/scionproto/scion/go/lib/consul/consulconfig"
)

func TestSampleCorrect(t *testing.T) {
	Convey("Load", t, func() {
		var cfg Config

		cfg.DS.UseFileModTime = true
		cfg.DS.Dynamic.ServicePrefix = "⠎⠉⠊⠕⠝"

		_, err := toml.Decode(Sample, &cfg)
		SoMsg("err", err, ShouldBeNil)

		// Non-dsconfig specific
		SoMsg("ID correct", cfg.General.ID, ShouldEqual, "ds-1")
		SoMsg("ConfigDir correct", cfg.General.ConfigDir, ShouldEqual, "/etc/scion")
		SoMsg("LogFile correct", cfg.Logging.File.Path, ShouldEqual, "/var/log/scion/ds-1.log")
		SoMsg("LogLvl correct", cfg.Logging.File.Level, ShouldEqual, "debug")
		SoMsg("LogFlush correct", *cfg.Logging.File.FlushInterval, ShouldEqual, 5)
		SoMsg("LogConsoleLvl correct", cfg.Logging.Console.Level, ShouldEqual, "crit")

		// dsconfig specific
		SoMsg("UseFileModTime correct", cfg.DS.UseFileModTime, ShouldBeFalse)
		SoMsg("ACL correct", cfg.DS.ACL, ShouldEqual, "/etc/scion/ds-acl")
		SoMsg("Cert correct", cfg.DS.Cert, ShouldEqual, "/etc/scion/tls/cert")
		SoMsg("Key correct", cfg.DS.Key, ShouldEqual, "/etc/scion/tls/key")
		SoMsg("ListenAddr correct", cfg.DS.ListenAddr, ShouldEqual, "127.0.0.1:8080")
		SoMsg("Dynamic.QueryInterval correct", cfg.DS.Dynamic.QueryInterval.Duration, ShouldEqual,
			DefaultDynamicQueryInterval)
		SoMsg("Dynamic.Timeout correct", cfg.DS.Dynamic.Timeout.Duration, ShouldEqual,
			DefaultDynamicTimeout)
		SoMsg("Dynamic.Service prefix correct", cfg.DS.Dynamic.ServicePrefix, ShouldBeEmpty)

		// Consul specific
		SoMsg("Consul.Enabled correct", cfg.Consul.Enabled, ShouldBeFalse)
		SoMsg("Consul.Agent correct", cfg.Consul.Agent, ShouldEqual, consulconfig.DefaultAgent)
		SoMsg("Consul.Health.Interval correct", cfg.Consul.Health.Interval.Duration,
			ShouldEqual, consulconfig.DefaultHealthInterval)
		SoMsg("Consul.Health.Timeout correct", cfg.Consul.Health.Timeout.Duration,
			ShouldEqual, consulconfig.DefaultHealthTimeout)
		SoMsg("Consul.InitConnPeriod correct", cfg.Consul.InitConnPeriod.Duration,
			ShouldEqual, consulconfig.DefaultInitConnPeriod)
	})
}
