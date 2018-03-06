// Copyright 2018 Anapaya Systems

package web

import (
	"net/http"
	"strconv"
	"strings"

	log "github.com/inconshreveable/log15"

	"github.com/scionproto/scion/go/sigmgmt/config"
	"github.com/scionproto/scion/go/sigmgmt/db"
	"github.com/scionproto/scion/go/sigmgmt/util"
)

type sitesValidatedForm struct {
	dbase     *db.DB
	request   *http.Request
	sites     map[string]*db.Site
	site      *db.Site
	webAppCfg *config.Global
	log       log.Logger
}

func newSitesValidatedForm(r *http.Request, dbase *db.DB, webAppCfg *config.Global,
	logger log.Logger) *sitesValidatedForm {

	if err := r.ParseForm(); err != nil {
		logger.Error("Error during form parsing", "err", err)
		return nil
	}
	vf := &sitesValidatedForm{
		request:   r,
		dbase:     dbase,
		webAppCfg: webAppCfg,
		log:       logger,
	}
	return vf
}

func (vf *sitesValidatedForm) GetSites() map[string]*db.Site {
	var err error
	if vf.sites, err = vf.dbase.QuerySites(); err != nil {
		vf.log.Error("Unable to load sites", "err", err)
	}
	return vf.sites
}

func (vf *sitesValidatedForm) GetSite() *db.Site {
	vHost := vf.request.Form.Get("vhost")
	if err := util.ValidateIdentifier(vHost); err != nil {
		vf.log.Error("Bad VHost", "err", err)
		return nil
	}
	hosts, err := parseHosts(vf.request.Form.Get("hosts"))
	if err != nil {
		vf.log.Error("Bad hosts", "err", err)
		return nil
	}
	metricsPort, err := strconv.ParseUint(vf.request.Form.Get("port"), 10, 16)
	if err != nil {
		vf.log.Error("Bad port", "err", err)
		return nil
	}
	vf.site = &db.Site{
		Name:        vf.request.Form.Get("name"),
		VHost:       vHost,
		Hosts:       hosts,
		MetricsPort: uint16(metricsPort),
	}
	return vf.site
}

func (vf *sitesValidatedForm) DeleteSite() {
	site := vf.GetSite()
	if err := vf.dbase.DeleteSite(site.Name); err != nil {
		vf.log.Error("Unable to delete site", "name", site.Name)
	}
}

func (vf *sitesValidatedForm) AddSite() {
	site := vf.GetSite()
	if site == nil {
		return
	}
	if err := vf.dbase.InsertSite(site); err != nil {
		vf.log.Error("Unable to add site", "name", vf.request.Form.Get("name"),
			"err", err)
	}
}

func parseHosts(hostsStr string) ([]string, error) {
	hosts := strings.Split(hostsStr, ",")
	for _, host := range hosts {
		if err := util.ValidateIdentifier(host); err != nil {
			return nil, err
		}
	}
	return hosts, nil
}
