// Copyright 2018 Anapaya Systems

package ctrl

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/jinzhu/gorm"

	"github.com/scionproto/scion/go/lib/addr"
	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/lib/log"
	"github.com/scionproto/scion/go/lib/pathpol"
	"github.com/scionproto/scion/go/sig/anaconfig"
	"github.com/scionproto/scion/go/sigmgmt/cfggen"
	"github.com/scionproto/scion/go/sigmgmt/db"
	"github.com/scionproto/scion/go/sigmgmt/netcopy"
	"github.com/scionproto/scion/go/sigmgmt/parser"
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
	var trafficPolicies map[addr.IA][]db.TrafficPolicy
	var trafficClasses map[uint]db.TrafficClass
	var pathPolicies []*pathpol.ExtPolicy
	if trafficPolicies, msg, err = c.getTrafficPolicies(site); err != nil {
		return nil, msg, err
	}
	if trafficClasses, msg, err = c.getTrafficClasses(site); err != nil {
		return nil, msg, err
	}
	if pathPolicies, msg, err = c.getPathPolicies(siteID); err != nil {
		return nil, msg, err
	}
	if err = cfggen.Compile(siteCfg, trafficPolicies, trafficClasses, pathPolicies); err != nil {
		return nil, "Config compiler error", err
	}
	siteCfg.ConfigVersion = uint64(time.Now().Unix())
	return siteCfg, "", nil
}

func (c *Controller) writeConfig(siteID uint, siteCfg *config.Cfg) (string, error,
	map[string]error) {
	ctx, cancelF := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancelF()

	jsonStrNew, err := json.MarshalIndent(siteCfg, "", "    ")
	if err != nil {
		return "Error building json config file", err, nil
	}
	// writeJSONFile
	fname := fmt.Sprintf("/sig-config-%d.json", siteCfg.ConfigVersion)
	path := filepath.Join(c.cfg.OutputDir, fname)
	if err := ioutil.WriteFile(path, jsonStrNew, 0644); err != nil {
		log.Error(jsonSaveError, "file", path, "err", err)
		return jsonSaveError, err, nil
	}
	log.Info("Generated JSON config", "file", path)
	var site db.Site
	if err = c.db.Preload("Hosts").Where("id = ?", siteID).Find(&site).Error; err != nil {
		return "Could not get site", err, nil
	}
	if len(site.Hosts) == 0 {
		return "No hosts specified", common.NewBasicError("Empty hosts list", nil), nil
	}
	if ok, errs := netcopy.CopyFileToSite(ctx, path, &site, c.cfg.SIGCfgPath, log.Root()); !ok {
		log.Error(configCopyError, "site", site.Name, "errs", errs)
		return configCopyError, nil, errs
	}
	log.Info("SIG config copy to remote site successful", "site", site.Name)
	if ok, errs := netcopy.ReloadSite(ctx, &site, log.Root()); !ok {
		return "Unable to reload SIG on remote site", nil, errs
	}
	log.Debug("SIG configuration reload triggered successfully", "vhost", site.VHost,
		"name", site.Name)
	if ok, errs := netcopy.VerifyConfigVersion(ctx, &site, siteCfg.ConfigVersion,
		log.Root()); !ok {
		return "Unable to verify version of reloaded config", nil, errs
	}
	log.Info("SIG Config version verification successful", "vhost", site.VHost, "name", site.Name,
		"version", siteCfg.ConfigVersion)
	return "", nil, nil
}

func (c *Controller) getSiteConfig(siteID uint) (*config.Cfg, *db.Site, string, error) {
	cfg := &config.Cfg{}
	var site db.Site
	err := c.db.Preload("ASEntries").Preload("TrafficClasses").
		Preload("ASEntries.Networks").Preload("ASEntries.Policy").Preload("ASEntries.SIGs").
		Where("id = ?", siteID).Find(&site).Error
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
		cfg.ASes[ia] = &config.ASEntry{Nets: networks}
	}
	return cfg, &site, "", nil
}

// validateSubTrafficClasses checks if the traffic classes referenced
// in `cond` do exist in the database
func (c *Controller) validateSubTrafficClasses(cls *db.TrafficClass) error {
	re, _ := regexp.Compile("cls=([0-9]+)")
	for _, match := range re.FindAllString(cls.CondStr, -1) {
		id := match[4:]
		if id == strconv.Itoa(int(cls.ID)) {
			return common.NewBasicError("Cannot self-reference!", nil)
		}
		err := c.db.First(&db.TrafficClass{}, id).Error
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Controller) validateTrafficClass(class *db.TrafficClass) (string, error) {
	if err := parser.ValidateTrafficClass(class.CondStr); err != nil {
		return err.Error(), err
	}
	if err := c.validateSubTrafficClasses(class); err != nil {
		return TrafficClassError, err
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

func (c *Controller) getTrafficPolicies(site *db.Site) (map[addr.IA][]db.TrafficPolicy, string,
	error) {

	policies := make(map[addr.IA][]db.TrafficPolicy)
	for _, as := range site.ASEntries {
		ia, err := as.ToAddrIA()
		if err != nil {
			return nil, "Could not convert AS to addr.IA", err
		}
		policies[ia] = as.Policy
	}
	return policies, "", nil
}

func (c *Controller) getTrafficClasses(site *db.Site) (map[uint]db.TrafficClass, string, error) {
	trafficClasses := make(map[uint]db.TrafficClass)
	for _, tc := range site.TrafficClasses {
		if err := substituteTrafficClasses(&tc, c.db); err != nil {
			return nil, "Could not substitute traffic classes", err
		}
		trafficClasses[tc.ID] = tc
	}
	return trafficClasses, "", nil
}

func (c *Controller) getPathPolicies(siteID uint) ([]*pathpol.ExtPolicy, string, error) {
	var pathPolicyFiles []db.PathPolicyFile
	if err := c.db.Where("type = ? OR site_id = ?", db.GlobalPolicy, siteID).Find(
		&pathPolicyFiles).Error; err != nil {
		return nil, "No path policies found", err
	}
	pathPolicies := make([]*pathpol.ExtPolicy, 0)
	for _, ppF := range pathPolicyFiles {
		err := json.Unmarshal([]byte(ppF.CodeStr), &ppF.Code)
		if err != nil {
			return nil, "", err
		}
		err = c.validatePathPolicyFile(ppF)
		if err != nil {
			return nil, "", err
		}
		policies, err := ppF.GetExtPolicies()
		if err != nil {
			return nil, "Error getting extPolicies", err
		}
		for _, policy := range policies {
			// TODO that implies that policy names must be globally unique
			// maybe this should be changed to unique per file
			pathPolicies = append(pathPolicies, policy)
		}
	}
	return pathPolicies, "", nil
}

func (c *Controller) pathPoliciesExists(names []string, asID uint) error {
	var as db.ASEntry
	var err error
	if err = c.db.First(&as, asID).Error; err != nil {
		return common.NewBasicError("AS not found", nil, "AS", asID)
	}
	var pathPolicies []*pathpol.ExtPolicy
	if pathPolicies, _, err = c.getPathPolicies(as.SiteID); err != nil {
		return err
	}
	policyNames := make(map[string]struct{})
	for _, policy := range pathPolicies {
		policyNames[policy.Name] = struct{}{}
	}
	for _, name := range names {
		if _, ok := policyNames[name]; !ok {
			return common.NewBasicError("Could not find path policy", nil, "policy", name)
		}
	}
	return nil
}

// validatePathPolicyFile tests for exach policy in the file if it compiles
func (c *Controller) validatePathPolicyFile(pF db.PathPolicyFile) error {
	policies, err := pF.GetExtPolicies()
	if err != nil {
		return err
	}
	for _, policy := range policies {
		if len(policy.Extends) > 0 {
			_, err := pathpol.PolicyFromExtPolicy(policy, policies)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// substituteTrafficClasses replaces `cls=xx` by the corresponding class
func substituteTrafficClasses(tc *db.TrafficClass, dbase *gorm.DB) error {
	IDs := map[uint]struct{}{tc.ID: {}}
	re, _ := regexp.Compile("cls=([0-9]+)")
	match := re.FindString(tc.CondStr)
	for match != "" {
		id := match[4:]
		uID, _ := strconv.Atoi(id)
		// make sure there is no cicular dependency
		if _, ok := IDs[uint(uID)]; !ok {
			IDs[uint(uID)] = struct{}{}
		} else {
			return common.NewBasicError("Circle reference detected", nil, "ID", id, "IDs", IDs)
		}
		cls := db.TrafficClass{}
		if err := dbase.First(&cls, id).Error; err != nil {
			return common.NewBasicError("Traffic Class not found", nil, "ID", id)
		}
		// replace `cls=xx` by corresponding condStr
		tc.CondStr = strings.Replace(tc.CondStr, match, cls.CondStr, 1)
		match = re.FindString(tc.CondStr)
	}
	return nil
}

func validateIPAllocation(allocation *db.SiteNetwork) (string, error) {
	_, _, err := net.ParseCIDR(allocation.CIDR)
	if err != nil {
		return CIDRParseError, err
	}
	if len(allocation.ACL) == 0 {
		return "No ACL found", common.NewBasicError("Bad ACL", nil)
	}
	aclParts := strings.Split(allocation.ACL, " ")
	if len(aclParts) == 1 && aclParts[0] == "*" {
		return "", nil
	}
	for _, part := range aclParts {
		_, err := addr.IAFromString(part)
		if err != nil {
			return "Unable to parse IA in ACL", err
		}
	}
	return "", nil
}
