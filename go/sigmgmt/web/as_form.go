// Copyright 2018 Anapaya Systems

package web

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	log "github.com/inconshreveable/log15"

	"github.com/scionproto/scion/go/lib/addr"
	"github.com/scionproto/scion/go/lib/pathmgr"
	"github.com/scionproto/scion/go/lib/pktcls"
	sigcfg "github.com/scionproto/scion/go/sig/config"
	"github.com/scionproto/scion/go/sig/mgmt"
	"github.com/scionproto/scion/go/sig/siginfo"
	"github.com/scionproto/scion/go/sigmgmt/cfggen"
	"github.com/scionproto/scion/go/sigmgmt/config"
	"github.com/scionproto/scion/go/sigmgmt/db"
	"github.com/scionproto/scion/go/sigmgmt/netcopy"
)

// validatedForm contains objects retrieved from the online form and from the
// database. GetXxx methods populate the relevant field (either by parsing form
// contents or querying the database) and return its value. On errors, the
// value is set to nil.
type validatedForm struct {
	dbase          *db.DB
	request        *http.Request
	site           *db.Site
	ia             *addr.ISD_AS
	siteCfg        *sigcfg.Cfg
	sig            *sigcfg.SIG
	cidr           *net.IPNet
	filter         *pktcls.ActionFilterPaths
	session        *sigcfg.Session
	policies       map[addr.ISD_AS]string
	sessionAliases map[addr.ISD_AS]db.SessionAliasMap
	policy         *string
	webAppCfg      *config.Global
	parsed         map[string]bool

	log log.Logger
}

func newValidatedForm(r *http.Request, dbase *db.DB, webAppCfg *config.Global,
	logger log.Logger) *validatedForm {

	if err := r.ParseForm(); err != nil {
		logger.Error("Error during form parsing", "err", err)
		return nil
	}
	vf := &validatedForm{
		request:   r,
		dbase:     dbase,
		parsed:    make(map[string]bool),
		webAppCfg: webAppCfg,
		log:       logger,
	}
	return vf
}

func (vf *validatedForm) GetSite() *db.Site {
	var site *db.Site
	if !vf.parsed["site"] {
		vf.parsed["site"] = true
		var err error
		siteName := vf.request.Form.Get("site")
		if site, err = vf.dbase.GetSite(siteName); err != nil {
			vf.log.Error("Unable to load site information", "name", siteName,
				"err", err)
			return nil
		}
		vf.site = site
	}
	return vf.site
}

func (vf *validatedForm) GetIA() *addr.ISD_AS {
	var ia *addr.ISD_AS
	if !vf.parsed["ia"] {
		vf.parsed["ia"] = true
		var err error
		var iaStr string
		if vf.request.Form.Get("ia") != "" {
			iaStr = vf.request.Form.Get("ia")
		} else {
			iaStr = fmt.Sprintf("%s-%s", vf.request.Form.Get("ISD"), vf.request.Form.Get("AS"))
		}
		if ia, err = addr.IAFromString(iaStr); err != nil {
			vf.log.Error("Unable to parse ISD-AS value", "ia", iaStr, "err", err)
			return nil
		}
		vf.ia = ia
	}
	return vf.ia
}

func (vf *validatedForm) GetCIDR() *net.IPNet {
	var cidr *net.IPNet
	if !vf.parsed["cidr"] {
		vf.parsed["cidr"] = true
		var err error
		_, cidr, err = net.ParseCIDR(vf.request.Form.Get("cidr"))
		if err != nil {
			vf.log.Error("Unable to parse CIDR address", "err", err)
			return nil
		}
		vf.cidr = cidr
	}
	return vf.cidr
}

func (vf *validatedForm) GetSIG() *sigcfg.SIG {
	sig := &sigcfg.SIG{}
	if !vf.parsed["sig"] {
		vf.parsed["sig"] = true
		var err error
		sig.Id = siginfo.SigIdType(vf.request.Form.Get("SigName"))
		if sig.Addr = net.ParseIP(vf.request.Form.Get("Addr")); sig.Addr == nil {
			vf.log.Error("String is not a valid IP address", "str", vf.request.Form.Get("Addr"))
			return nil
		}
		ctrlPort, err := strconv.ParseUint(vf.request.Form.Get("CtrlPort"), 10, 16)
		if err != nil {
			log.Error("Bad CtrlPort number", "err", err)
			return nil
		}
		sig.CtrlPort = uint16(ctrlPort)
		encapPort, err := strconv.ParseUint(vf.request.Form.Get("EncapPort"), 10, 16)
		if err != nil {
			vf.log.Error("Bad EncapPort number", "err", err)
			return nil
		}
		sig.EncapPort = uint16(encapPort)
		vf.sig = sig
	}
	return vf.sig
}

func (vf *validatedForm) GetFilter() *pktcls.ActionFilterPaths {
	if !vf.parsed["filter"] {
		vf.parsed["filter"] = true
		pp, err := pathmgr.NewPathPredicate(vf.request.Form.Get("pp"))
		if err != nil {
			vf.log.Error("Bad path selector string", "err", err)
			return nil
		}
		vf.filter = pktcls.NewActionFilterPaths(vf.request.Form.Get("name"), pp)
	}
	return vf.filter
}

func (vf *validatedForm) GetSession() *sigcfg.Session {
	session := &sigcfg.Session{}
	if !vf.parsed["session"] {
		vf.parsed["session"] = true
		var err error
		session.PolName = vf.request.Form.Get("pathselector")
		idStr := vf.request.Form.Get("sessionname")
		sessID, err := strconv.ParseUint(idStr, 0, 8)
		if err != nil {
			vf.log.Error("Bad session name (must be integer between 0 and 255)", "value", idStr)
			return nil
		}
		session.ID = mgmt.SessionType(sessID)
	}
	vf.session = session
	return vf.session
}

func (vf *validatedForm) GetSiteConfig() *sigcfg.Cfg {
	var err error
	site := vf.GetSite()
	if site == nil {
		return nil
	}
	if vf.siteCfg, err = vf.dbase.GetSiteConfig(site.Name); err != nil {
		vf.log.Error("Unable to read site config from database", "site", site.Name,
			"err", err)
		return nil
	}
	return vf.siteCfg
}

func (vf *validatedForm) GetPolicy() *string {
	if !vf.parsed["policy"] {
		policyStr := vf.request.Form.Get("policy")
		vf.policy = &policyStr
		vf.parsed["policy"] = true
	}
	return vf.policy
}

func (vf *validatedForm) GetPolicies() map[addr.ISD_AS]string {
	var err error
	site := vf.GetSite()
	if site == nil {
		return nil
	}
	if vf.policies, err = vf.dbase.GetPolicies(site.Name); err != nil {
		vf.log.Error("Error fetching policies", "err", err)
	}
	return vf.policies
}

func (vf *validatedForm) GetSessionAliases() map[addr.ISD_AS]db.SessionAliasMap {
	var err error
	site := vf.GetSite()
	if site == nil {
		return nil
	}
	if vf.sessionAliases, err = vf.dbase.GetSessionAliases(site.Name); err != nil {
		vf.log.Error("Error fetching session aliases", "err", err)
	}
	return vf.sessionAliases
}

func (vf *validatedForm) DeleteIA() {
	ia := vf.GetIA()
	site := vf.GetSite()
	if ia == nil || site == nil {
		return
	}
	if err := vf.dbase.DeleteAS(site.Name, *ia); err != nil {
		vf.log.Error("Unable to delete AS from database", "err", err)
	}
}

func (vf *validatedForm) AddIA() {
	site := vf.GetSite()
	ia := vf.GetIA()
	if site == nil || ia == nil {
		return
	}
	if err := vf.dbase.InsertAS(site.Name, *ia); err != nil {
		vf.log.Error("Unable to insert AS into database", "err", err)
	}
}

func (vf *validatedForm) AddNetwork() {
	site := vf.GetSite()
	ia := vf.GetIA()
	cidr := vf.GetCIDR()
	if site == nil || ia == nil || cidr == nil {
		return
	}
	if err := vf.dbase.InsertNetwork(site.Name, *ia, cidr); err != nil {
		vf.log.Error("Unable to insert network into database", "err", err)
	}
}

func (vf *validatedForm) DeleteNetwork() {
	site := vf.GetSite()
	ia := vf.GetIA()
	cidr := vf.GetCIDR()
	if site == nil || ia == nil || cidr == nil {
		return
	}
	if err := vf.dbase.DeleteNetwork(site.Name, *ia, cidr.String()); err != nil {
		vf.log.Error("Unable to delete network from database", "err", err)
	}
}

func (vf *validatedForm) AddSIG() {
	site := vf.GetSite()
	ia := vf.GetIA()
	sig := vf.GetSIG()
	if site == nil || ia == nil || sig == nil {
		return
	}
	if err := vf.dbase.InsertSIG(site.Name, *ia, sig); err != nil {
		vf.log.Error("Unable to insert SIG into database", "err", err)
	}
}

func (vf *validatedForm) DeleteSIG() {
	site := vf.request.Form.Get("site")
	sigName := vf.request.Form.Get("name")
	ia, err := addr.IAFromString(vf.request.Form.Get("ia"))
	if err != nil {
		vf.log.Error("Unable to parse ISD-AS value", "err", err)
		return
	}
	if err := vf.dbase.DeleteSIG(site, *ia, sigName); err != nil {
		vf.log.Error("Unable to delete SIG from database", "err", err)
		return
	}
}

func (vf *validatedForm) AddPathPredicate() {
	site := vf.GetSite()
	filter := vf.GetFilter()
	if site == nil || filter == nil {
		return
	}
	if err := vf.dbase.InsertFilter(site.Name, filter.Name, filter.Contains); err != nil {
		vf.log.Error("Unable to insert path predicate into database", "err", err)
	}
}

func (vf *validatedForm) DeletePathPredicate() {
	site := vf.GetSite()
	name := vf.request.Form.Get("name")
	if err := vf.dbase.DeleteFilter(site.Name, name); err != nil {
		vf.log.Error("Unable to delete Path selector from the database", "err", err)
		return
	}
}

func (vf *validatedForm) AddSession() {
	site := vf.GetSite()
	ia := vf.GetIA()
	session := vf.GetSession()
	if site == nil || ia == nil || session == nil {
		return
	}
	if err := vf.dbase.InsertSession(site.Name, *ia, uint8(session.ID), session.PolName); err != nil {
		vf.log.Error("Unable to insert Session into database", "err", err)
	}
}

func (vf *validatedForm) DeleteSession() {
	site := vf.GetSite()
	ia := vf.GetIA()
	session := vf.GetSession()
	if err := vf.dbase.DeleteSession(site.Name, *ia, uint8(session.ID)); err != nil {
		vf.log.Error("Unable to delete Session from database", "err", err)
	}
}

func (vf *validatedForm) SetPolicy() {
	site := vf.GetSite()
	ia := vf.GetIA()
	policy := vf.GetPolicy()
	// Compile the new policy rules to test for correctness prior to committing
	// to the database.
	if site == nil || ia == nil || policy == nil {
		return
	}
	policies := map[addr.ISD_AS]string{
		*ia: *policy,
	}
	siteConfig := vf.GetSiteConfig()
	aliases := vf.GetSessionAliases()
	if siteConfig == nil || aliases == nil {
		return
	}
	if err := cfggen.Compile(siteConfig, policies, aliases); err != nil {
		vf.log.Error("Config compiler error", "err", err)
		return
	}

	if err := vf.dbase.SetPolicy(site.Name, *ia, *policy); err != nil {
		vf.log.Error("Unable to insert policy into database", "err", err)
		return
	}
	vf.log.Info("Successfully saved policy")
}

func (vf *validatedForm) Push() {
	site := vf.GetSite()
	siteCfg := vf.GetSiteConfig()
	policies := vf.GetPolicies()
	aliases := vf.GetSessionAliases()
	if site == nil || siteCfg == nil || policies == nil || aliases == nil {
		return
	}
	siteCfg.ConfigVersion = uint64(time.Now().Unix())
	ctx, cancelF := context.WithTimeout(context.Background(), DefaultTimeout)
	defer cancelF()

	jsonStrNew, err := json.MarshalIndent(siteCfg, "", "    ")
	if err != nil {
		vf.log.Error("Error building json config file", "err", err)
		return
	}
	// writeJSONFile
	fname := fmt.Sprintf("/sig-config-%d.json", siteCfg.ConfigVersion)
	path := filepath.Join(vf.webAppCfg.OutputDir, fname)
	if err := ioutil.WriteFile(path, jsonStrNew, 0644); err != nil {
		vf.log.Error("Error saving JSON config to file", "file", path, "err", err)
		return
	}
	vf.log.Info("Generated JSON config", "file", path)

	if err := netcopy.CopyFileToSite(ctx, path, site, vf.webAppCfg.SIGCfgPath, vf.log); err != nil {
		vf.log.Error("Unable to copy configuration to site", "site", site.Name,
			"err", err)
		return
	}
	vf.log.Info("SIG config copy to remote site successful", "site", site.Name)
	if err := netcopy.ReloadSite(ctx, site, vf.log); err != nil {
		return
	}
	vf.log.Info("SIG configuration reload successful", "vhost", site.VHost)
	if err := netcopy.VerifyConfigVersion(ctx, site, siteCfg.ConfigVersion, vf.log); err != nil {
		vf.log.Error("Unable to verify version of reloaded config", "err", err)
		return
	}
	vf.log.Info("SIG Config version verification successful", "vhost", site.VHost,
		"version", siteCfg.ConfigVersion)
}
