// Copyright 2018 Anapaya Systems

package ctrl

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"time"

	"github.com/gorilla/mux"
	log "github.com/inconshreveable/log15"

	"github.com/scionproto/scion/go/lib/addr"
	sigcfg "github.com/scionproto/scion/go/sig/anaconfig"
	"github.com/scionproto/scion/go/sigmgmt/cfggen"
	"github.com/scionproto/scion/go/sigmgmt/db"
	"github.com/scionproto/scion/go/sigmgmt/netcopy"
	"github.com/scionproto/scion/go/sigmgmt/util"
)

const (
	// DefaultTimeout is the total timeout for pushes (config generation, saving, copying,
	// reloading, verification)
	DefaultTimeout = 5 * time.Second
	// IAParseError error string
	IAParseError    = "Unable to parse ISD-AS string"
	jsonSaveError   = "Error saving JSON config to file"
	configCopyError = "Unable to copy configuration to site"
)

func (sc *SiteController) getAndValidateSiteCfg(name string) (*db.Site, *sigcfg.Cfg, string,
	error) {
	site, err := sc.dbase.GetSiteWithHosts(name)
	if err != nil {
		return nil, nil, "Unable to get site", err
	}
	siteCfg, err := sc.dbase.GetSiteConfig(site.Name)
	if err != nil {
		return nil, nil, "Unable to read site config from database", err
	}
	policies, err := sc.dbase.GetPolicies(site.Name)
	if err != nil {
		return nil, nil, "Error fetching policies", err
	}
	if site == nil || siteCfg == nil || policies == nil {
		return nil, nil, "Not all necessary values are present",
			errors.New("Empty site, siteCfg or policies")
	}
	if err = cfggen.Compile(siteCfg, policies); err != nil {
		return nil, nil, "Config compiler error", err
	}
	siteCfg.ConfigVersion = uint64(time.Now().Unix())
	return site, siteCfg, "", nil
}

func (sc *SiteController) writeConfig(site *db.Site, siteCfg *sigcfg.Cfg) (string, error) {
	ctx, cancelF := context.WithTimeout(context.Background(), DefaultTimeout)
	defer cancelF()

	jsonStrNew, err := json.MarshalIndent(siteCfg, "", "    ")
	if err != nil {
		return "Error building json config file", err
	}
	// writeJSONFile
	fname := fmt.Sprintf("/sig-config-%d.json", siteCfg.ConfigVersion)
	path := filepath.Join(sc.cfg.OutputDir, fname)
	if err := ioutil.WriteFile(path, jsonStrNew, 0644); err != nil {
		log.Error(jsonSaveError, "file", path, "err", err)
		return jsonSaveError, err
	}
	log.Info("Generated JSON config", "file", path)

	if err := netcopy.CopyFileToSite(ctx, path, site, sc.cfg.SIGCfgPath, log.Root()); err != nil {
		log.Error(configCopyError, "site", site.Name, "err", err)
		return configCopyError, err
	}
	log.Info("SIG config copy to remote site successful", "site", site.Name)
	if err := netcopy.ReloadSite(ctx, site, log.Root()); err != nil {
		return "Unable to reload SIG on remote site", err
	}
	log.Debug("SIG configuration reload triggered successfully", "vhost", site.VHost)
	if err := netcopy.VerifyConfigVersion(ctx, site, siteCfg.ConfigVersion,
		log.Root()); err != nil {
		return "Unable to verify version of reloaded config", err
	}
	log.Info("SIG Config version verification successful", "vhost", site.VHost,
		"version", siteCfg.ConfigVersion)
	return "", nil
}

func (ac *ASController) validatePolicies(site string,
	policies map[addr.IA]string) (string, error) {
	siteConfig, err := ac.dbase.GetSiteConfig(site)
	if err != nil || siteConfig == nil {
		return "Unable to read site config from database", err
	}
	if err = cfggen.Compile(siteConfig, policies); err != nil {
		return err.Error(), err
	}
	return "", nil
}

func getIA(r *http.Request) (*addr.IA, string, error) {
	var err error
	var ia addr.IA
	if ia, err = addr.IAFromString(mux.Vars(r)["ia"]); err != nil {
		return &addr.IA{}, IAParseError, err
	}
	return &ia, "", nil
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
	err := parseHosts(site.Hosts)
	if err != nil {
		return "Bad hosts", err
	}
	return "", nil
}
