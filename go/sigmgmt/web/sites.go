// Copyright 2018 Anapaya Systems

package web

import (
	"net/http"

	log "github.com/inconshreveable/log15"

	"github.com/scionproto/scion/go/sigmgmt/config"
	"github.com/scionproto/scion/go/sigmgmt/db"
	"github.com/scionproto/scion/go/sigmgmt/weblogger"
)

type SiteFormData struct {
	Sites    map[string]*db.Site
	Log      log.Logger
	Features config.FeatureLevel
}

func BuildSiteFormData(webAppCfg *config.Global, dbase *db.DB, r *http.Request) *SiteFormData {
	htmlLogger := weblogger.New(log.Root())
	vf := newSitesValidatedForm(r, dbase, webAppCfg, htmlLogger)
	switch r.Form.Get("action") {
	case "delete":
		vf.DeleteSite()
	case "add":
		vf.AddSite()
	case "":
	default:
		htmlLogger.Error("Unknown action", "action", r.Form.Get("action"))
	}
	data := &SiteFormData{
		Sites:    vf.GetSites(),
		Features: webAppCfg.Features,
		Log:      htmlLogger,
	}
	return data
}
