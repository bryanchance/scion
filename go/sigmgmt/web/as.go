// Copyright 2018 Anapaya Systems

package web

import (
	"net/http"

	log "github.com/inconshreveable/log15"

	"github.com/scionproto/scion/go/lib/addr"
	sigcfg "github.com/scionproto/scion/go/sig/config"
	"github.com/scionproto/scion/go/sigmgmt/config"
	"github.com/scionproto/scion/go/sigmgmt/db"
	"github.com/scionproto/scion/go/sigmgmt/weblogger"
)

type ASFormData struct {
	Site              *db.Site
	Cfg               *sigcfg.Cfg
	Policies          map[addr.IA]string
	AllSessionAliases map[addr.IA]db.SessionAliasMap
	Log               log.Logger
	Features          config.FeatureLevel
}

func BuildASFormData(webAppCfg *config.Global, dbase *db.DB, r *http.Request) *ASFormData {
	htmlLogger := weblogger.New(log.Root())
	vf := newValidatedForm(r, dbase, webAppCfg, htmlLogger)
	switch r.Form.Get("action") {
	case "delete-ia":
		vf.DeleteIA()
	case "add-ia":
		vf.AddIA()
	case "delete-network":
		vf.DeleteNetwork()
	case "add-network":
		vf.AddNetwork()
	case "add-sig":
		vf.AddSIG()
	case "delete-sig":
		vf.DeleteSIG()
	case "add-pp":
		vf.AddPathPredicate()
	case "delete-pp":
		vf.DeletePathPredicate()
	case "add-session":
		vf.AddSession()
	case "delete-session":
		vf.DeleteSession()
	case "set-policy":
		vf.SetPolicy()
	case "push":
		vf.Push()
	case "":
		// do nothing
	default:
		htmlLogger.Error("Unknown action", "action", r.Form.Get("action"))
	}
	data := &ASFormData{
		Site:              vf.GetSite(),
		Cfg:               vf.GetSiteConfig(),
		Policies:          vf.GetPolicies(),
		AllSessionAliases: vf.GetSessionAliases(),
		Log:               htmlLogger,
		Features:          webAppCfg.Features,
	}
	return data
}

func getPolicyFromIA(policies map[addr.IA]string, ia addr.IA) string {
	return policies[ia]
}

func getSessionAliasesFromIA(sessionAliasesMap map[addr.IA]db.SessionAliasMap,
	ia addr.IA) db.SessionAliasMap {
	return sessionAliasesMap[ia]
}

func sortByIA(collection map[addr.IA]*sigcfg.ASEntry) map[addr.IAInt]interface{} {
	m := make(map[addr.IAInt]interface{})
	for k, v := range collection {
		m[k.IAInt()] = v
	}
	return m
}

func iaIntToIA(i addr.IAInt) addr.IA {
	return i.IA()
}
