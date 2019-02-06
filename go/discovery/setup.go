// Copyright 2018 Anapaya Systems

package main

import (
	"io"

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

type closerFunc func() error

func (cf closerFunc) Close() error {
	return cf()
}

func setup(cfg *config.Config) (io.Closer, error) {
	if err := env.InitGeneral(&cfg.General); err != nil {
		return nil, err
	}
	cfg.DS.InitDefaults()
	if err := validateConfig(cfg); err != nil {
		return nil, common.NewBasicError("Unable to validate config", err)
	}
	ia = cfg.General.Topology.ISD_AS
	// metrics must be initialized before the files are loaded.
	metrics.Init(cfg.General.ID)
	if err := loadFiles(cfg); err != nil {
		return nil, err
	}
	c, err := setupConsul(cfg)
	if err != nil {
		return nil, err
	}
	startDynUpdater(cfg)
	environment = env.SetupEnv(func() {
		if err := reload(env.ConfigFile()); err != nil {
			log.Error("Unable to reload config", "err", err)
		}
	})
	return c, nil
}

func setupConsul(cfg *config.Config) (io.Closer, error) {
	cfg.Consul.InitDefaults()
	var err error
	if consulClient, err = cfg.Consul.Client(); err != nil {
		return nil, err
	}
	topoAddr := cfg.General.Topology.DS.GetById(cfg.General.ID)
	svc := &consul.Service{
		Agent:  consulClient.Agent(),
		ID:     cfg.General.ID,
		Prefix: cfg.Consul.Prefix,
		Addr:   topoAddr.PublicAddr(topoAddr.Overlay),
		Type:   consul.DS,
	}
	ttlUpdater, err := svc.Register(cfg.Consul.InitConnPeriod.Duration, checkHealth,
		cfg.Consul.Health)
	if err != nil {
		return nil, common.NewBasicError("Unable to register service with consul", err)
	}
	return closerFunc(func() error {
		ttlUpdater.Stop()
		svc.Deregister()
		return nil
	}), nil
}

func checkHealth() consul.CheckInfo {
	return consul.CheckInfo{
		Status: consul.StatusPass,
	}
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
	if !ia.IsZero() && !ia.Equal(cfg.General.Topology.ISD_AS) {
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
