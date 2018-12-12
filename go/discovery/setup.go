// Copyright 2018 Anapaya Systems

package main

import (
	"github.com/BurntSushi/toml"

	"github.com/scionproto/scion/go/discovery/acl"
	"github.com/scionproto/scion/go/discovery/dynamic"
	"github.com/scionproto/scion/go/discovery/metrics"
	"github.com/scionproto/scion/go/discovery/static"
	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/lib/env"
	"github.com/scionproto/scion/go/lib/log"
	"github.com/scionproto/scion/go/lib/periodic"
)

func setupBasic() (*Config, error) {
	config := &Config{}
	if _, err := toml.DecodeFile(env.ConfigFile(), config); err != nil {
		return nil, err
	}
	if err := env.InitLogging(&config.Logging); err != nil {
		return nil, err
	}
	env.LogAppStarted(common.DS, config.General.ID)
	return config, nil
}

func setup(config *Config) error {
	if err := env.InitGeneral(&config.General); err != nil {
		return err
	}
	config.DS.InitDefaults()
	if err := validateConfig(config); err != nil {
		return common.NewBasicError("Unable to validate config", err)
	}
	ia = config.General.Topology.ISD_AS
	// metrics must be initialized before the files are loaded.
	metrics.Init(config.General.ID)
	if err := loadFiles(config); err != nil {
		return err
	}
	startDynUpdater(config)
	environment = env.SetupEnv(func() {
		if err := reload(env.ConfigFile()); err != nil {
			log.Error("Unable to reload config", "err", err)
		}
	})
	return nil
}

func reload(configName string) error {
	config := &Config{}
	if _, err := toml.DecodeFile(configName, config); err != nil {
		return err
	}
	if err := env.InitGeneral(&config.General); err != nil {
		return err
	}
	config.DS.InitDefaults()
	if err := validateConfig(config); err != nil {
		return common.NewBasicError("Unable to validate config", err)
	}
	// Reload relevant static configuration files
	if err := loadFiles(config); err != nil {
		return err
	}
	startDynUpdater(config)
	return nil
}

func validateConfig(conf *Config) error {
	if conf.General.ID == "" {
		return common.NewBasicError("No element ID specified", nil)
	}
	if err := conf.DS.Validate(); err != nil {
		return err
	}
	if !ia.IsZero() && !ia.Eq(conf.General.Topology.ISD_AS) {
		return common.NewBasicError("IA changed", nil,
			"newIA", conf.General.Topology.ISD_AS, "currIA", ia)
	}
	return nil
}

func loadFiles(conf *Config) error {
	log.Info("Loading ACL", "filename", conf.DS.ACL)
	if err := acl.Load(conf.DS.ACL); err != nil {
		return common.NewBasicError("ACL file load failed", err)
	}
	log.Info("Loading static topology file", "filename", conf.General.TopologyPath)
	if err := static.Load(conf.General.TopologyPath, conf.DS.UseFileModTime); err != nil {
		return common.NewBasicError("Static topology file load failed", err)
	}
	return nil
}

func startDynUpdater(config *Config) {
	if dynUpdater != nil {
		// Run at most one updater at the same time.
		dynUpdater.Stop()
	}
	dynUpdater = periodic.StartPeriodicTask(
		&dynamic.ZkTask{
			IA:        ia,
			Instances: config.DS.Zoo.Instances,
			Timeout:   config.DS.Zoo.Timeout.Duration,
		},
		periodic.NewTicker(config.DS.Zoo.QueryInterval.Duration),
		config.DS.Zoo.Timeout.Duration,
	)
}
