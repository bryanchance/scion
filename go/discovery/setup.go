// Copyright 2018 Anapaya Systems

package main

import (
	"github.com/BurntSushi/toml"

	"github.com/scionproto/scion/go/discovery/internal/acl"
	"github.com/scionproto/scion/go/discovery/internal/config"
	"github.com/scionproto/scion/go/discovery/internal/dynamic"
	"github.com/scionproto/scion/go/discovery/internal/metrics"
	"github.com/scionproto/scion/go/discovery/internal/static"
	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/lib/consul"
	"github.com/scionproto/scion/go/lib/env"
	"github.com/scionproto/scion/go/lib/log"
	"github.com/scionproto/scion/go/lib/periodic"
)

func setupBasic() (*config.Config, error) {
	cfg := &config.Config{}
	if _, err := toml.DecodeFile(env.ConfigFile(), cfg); err != nil {
		return nil, err
	}
	if err := env.InitLogging(&cfg.Logging); err != nil {
		return nil, err
	}
	env.LogAppStarted(common.DS, cfg.General.ID)
	return cfg, nil
}

func setup(cfg *config.Config) error {
	if err := env.InitGeneral(&cfg.General); err != nil {
		return err
	}
	cfg.DS.InitDefaults()
	if err := validateConfig(cfg); err != nil {
		return common.NewBasicError("Unable to validate config", err)
	}
	ia = cfg.General.Topology.ISD_AS
	// metrics must be initialized before the files are loaded.
	metrics.Init(cfg.General.ID)
	if err := loadFiles(cfg); err != nil {
		return err
	}
	cfg.Consul.InitDefaults()
	if err := setupConsul(cfg); err != nil {
		return err
	}
	startDynUpdater(cfg)
	environment = env.SetupEnv(func() {
		if err := reload(env.ConfigFile()); err != nil {
			log.Error("Unable to reload config", "err", err)
		}
	})
	return nil
}

func setupConsul(cfg *config.Config) error {
	var err error
	if consulClient, err = cfg.Consul.Client(); err != nil {
		return err
	}
	if _, err = startUpdateTTL(cfg); err != nil {
		return err
	}
	return nil
}

func startUpdateTTL(cfg *config.Config) (*periodic.Runner, error) {
	svc := &consul.Service{
		Type:        consul.DS,
		Agent:       consulClient.Agent(),
		Logger:      log.New("part", "UpdateTTL"),
		CheckParams: cfg.Consul.Health,
		Check: func() consul.CheckInfo {
			return consul.CheckInfo{
				Id:     cfg.General.ID,
				Status: consul.StatusPass,
			}
		},
	}
	return svc.StartUpdateTTL(cfg.Consul.InitConnPeriod.Duration)
}

func reload(configName string) error {
	cfg := &config.Config{}
	if _, err := toml.DecodeFile(configName, cfg); err != nil {
		return err
	}
	if err := env.InitGeneral(&cfg.General); err != nil {
		return err
	}
	cfg.DS.InitDefaults()
	if err := validateConfig(cfg); err != nil {
		return common.NewBasicError("Unable to validate config", err)
	}
	// Reload relevant static configuration files
	if err := loadFiles(cfg); err != nil {
		return err
	}
	startDynUpdater(cfg)
	return nil
}

func validateConfig(cfg *config.Config) error {
	if err := cfg.Validate(); err != nil {
		return err
	}
	if !ia.IsZero() && !ia.Eq(cfg.General.Topology.ISD_AS) {
		return common.NewBasicError("IA changed", nil,
			"newIA", cfg.General.Topology.ISD_AS, "currIA", ia)
	}
	return nil
}

func loadFiles(cfg *config.Config) error {
	log.Info("Loading ACL", "filename", cfg.DS.ACL)
	if err := acl.Load(cfg.DS.ACL); err != nil {
		return common.NewBasicError("ACL file load failed", err)
	}
	log.Info("Loading static topology file", "filename", cfg.General.TopologyPath)
	if err := static.Load(cfg.General.TopologyPath, cfg.DS.UseFileModTime); err != nil {
		return common.NewBasicError("Static topology file load failed", err)
	}
	return nil
}

func startDynUpdater(cfg *config.Config) {
	if dynUpdater != nil {
		// Run at most one updater at the same time.
		dynUpdater.Stop()
	}
	dynUpdater = periodic.StartPeriodicTask(
		&dynamic.Updater{
			SvcPrefix: cfg.DS.Dynamic.ServicePrefix,
			Client:    consulClient,
		},
		periodic.NewTicker(cfg.DS.Dynamic.QueryInterval.Duration),
		cfg.DS.Dynamic.Timeout.Duration,
	)
}
