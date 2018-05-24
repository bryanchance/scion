// Copyright 2018 Anapaya Systems

package ctrl

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"time"

	"github.com/scionproto/scion/go/lib/addr"
	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/lib/log"
	"github.com/scionproto/scion/go/sig/anaconfig"
	"github.com/scionproto/scion/go/sigmgmt/cfggen"
	"github.com/scionproto/scion/go/sigmgmt/db"
	"github.com/scionproto/scion/go/sigmgmt/netcopy"
	"github.com/scionproto/scion/go/sigmgmt/util"
)

const (
	// defaultTimeout is the total timeout for pushes (config generation, saving, copying,
	// reloading, verification)
	defaultTimeout  = 5 * time.Second
	jsonSaveError   = "Error saving JSON config to file"
	configCopyError = "Unable to copy configuration to site"
)

func (c *Controller) getAndValidateSiteCfg(siteID uint) (*config.Cfg, string, error) {
	siteCfg, site, msg, err := c.getSiteConfig(siteID)
	if err != nil {
		return nil, "Unable to read site config from database: " + msg, err
	}
	if len(site.ASEntries) == 0 {
		return nil, "No remote AS configured", common.NewBasicError("Empty ASEntries", nil)
	}
	policies := make(map[addr.IA]string)
	for _, as := range site.ASEntries {
		ia, err := as.ToAddrIA()
		if err != nil {
			return nil, "Could not convert AS to addr.IA", err
		}
		if as.Policies != "" {
			policies[ia] = as.Policies
		}
	}
	if err = cfggen.Compile(siteCfg, policies); err != nil {
		return nil, "Config compiler error", err
	}
	siteCfg.ConfigVersion = uint64(time.Now().Unix())
	return siteCfg, "", nil
}

func (c *Controller) writeConfig(siteID uint, siteCfg *config.Cfg) (string, error) {
	ctx, cancelF := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancelF()

	jsonStrNew, err := json.MarshalIndent(siteCfg, "", "    ")
	if err != nil {
		return "Error building json config file", err
	}
	// writeJSONFile
	fname := fmt.Sprintf("/sig-config-%d.json", siteCfg.ConfigVersion)
	path := filepath.Join(c.cfg.OutputDir, fname)
	if err := ioutil.WriteFile(path, jsonStrNew, 0644); err != nil {
		log.Error(jsonSaveError, "file", path, "err", err)
		return jsonSaveError, err
	}
	log.Info("Generated JSON config", "file", path)
	var site db.Site
	if err = c.db.Preload("Hosts").Where("id = ?", siteID).Find(&site).Error; err != nil {
		return "Could not get site", err
	}
	if len(site.Hosts) == 0 {
		return "No hosts specified", common.NewBasicError("Empty hosts list", nil)
	}
	if err := netcopy.CopyFileToSite(ctx, path, &site, c.cfg.SIGCfgPath, log.Root()); err != nil {
		log.Error(configCopyError, "site", site.Name, "err", err)
		return configCopyError, err
	}
	log.Info("SIG config copy to remote site successful", "site", site.Name)
	if err := netcopy.ReloadSite(ctx, &site, log.Root()); err != nil {
		return "Unable to reload SIG on remote site", err
	}
	log.Debug("SIG configuration reload triggered successfully", "vhost", site.VHost)
	if err := netcopy.VerifyConfigVersion(ctx, &site, siteCfg.ConfigVersion,
		log.Root()); err != nil {
		return "Unable to verify version of reloaded config", err
	}
	log.Info("SIG Config version verification successful", "vhost", site.VHost,
		"version", siteCfg.ConfigVersion)
	return "", nil
}

func (c *Controller) getSiteConfig(siteID uint) (*config.Cfg, *db.Site, string, error) {
	cfg := &config.Cfg{}
	var site db.Site
	err := c.db.Preload("ASEntries").Preload("PathSelectors").Preload("ASEntries.Networks").
		Preload("ASEntries.SIGs").Where("id = ?", siteID).Find(&site).Error
	if err != nil {
		return nil, nil, "Could not get site", err
	}
	cfg.ASes = make(map[addr.IA]*config.ASEntry)
	for _, as := range site.ASEntries {
		ia, err := as.ToAddrIA()
		if err != nil {
			return nil, nil, "Could not convert AS to addr.IA", err
		}
		networks, err := db.IPNetsFromNetworks(as.Networks)
		if err != nil {
			return nil, nil, "Could not convert Networks", err
		}
		sigs, err := db.SIGSetFromSIGs(as.SIGs)
		if err != nil {
			return nil, nil, "Could not convert SIGs", err
		}
		cfg.ASes[ia] = &config.ASEntry{Nets: networks, Sigs: sigs}
	}
	actions, err := db.ActionMapFromSelectors(site.PathSelectors)
	if err != nil {
		return nil, nil, "Could not convert Networks", err
	}
	cfg.Actions = actions
	return cfg, &site, "", nil
}

func (c *Controller) validatePolicies(as *db.ASEntry) (string, error) {
	cfg, _, msg, err := c.getSiteConfig(as.SiteID)
	if err != nil {
		return "Unable to read site config from database: " + msg, err
	}
	currIA, _ := as.ToAddrIA()
	if err = cfggen.Compile(cfg, map[addr.IA]string{currIA: as.Policies}); err != nil {
		return err.Error(), err
	}
	return "", nil
}

func (c *Controller) updateHosts(newHosts []db.Host, siteID uint) error {
	oldHosts := []db.Host{}
	if err := c.db.Where("site_id = ?", siteID).Find(&oldHosts).Error; err != nil {
		return err
	}
	newHostMap := make(map[uint]db.Host, len(newHosts))
	for _, host := range newHosts {
		if host.ID != 0 {
			newHostMap[host.ID] = host
		}
	}
	// Delete removed hosts from DB
	for _, host := range oldHosts {
		if _, ok := newHostMap[host.ID]; !ok {
			if err := c.db.Delete(host).Error; err != nil {
				return err
			}
		}
	}
	// Add or update new hosts in DB
	for _, host := range newHosts {
		host.SiteID = siteID
		if host.ID != 0 {
			// Update host when ID present
			if err := c.db.Save(&host).Error; err != nil {
				return err
			}
		} else {
			if err := c.db.Create(&host).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

func parseHosts(hosts []db.Host) error {
	for _, host := range hosts {
		if err := util.ValidateIdentifier(host.Name); err != nil {
			return err
		}
		if err := util.ValidateUser(host.User); host.User != "" && err != nil {
			return err
		}
		if err := util.ValidateKey(host.Key); host.Key != "" && err != nil {
			return err
		}
	}
	return nil
}

func validateSite(site *db.Site) (string, error) {
	if err := util.ValidateIdentifier(site.VHost); err != nil {
		return "Bad VHost", err
	}
	if err := parseHosts(site.Hosts); err != nil {
		return "Bad hosts", err
	}
	return "", nil
}
