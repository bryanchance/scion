// Copyright 2018 Anapaya Systems

package ctrl

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/gorilla/mux"

	"github.com/scionproto/scion/go/lib/spath/spathmeta"
	"github.com/scionproto/scion/go/sigmgmt/config"
	"github.com/scionproto/scion/go/sigmgmt/db"
)

type SiteController struct {
	dbase *db.DB
	cfg   *config.Global
}

func NewSiteController(dbase *db.DB, cfg *config.Global) *SiteController {
	return &SiteController{
		dbase: dbase,
		cfg:   cfg,
	}
}

func (sc *SiteController) GetSite(w http.ResponseWriter, r *http.Request, _ http.HandlerFunc) {
	var err error
	var site *db.Site
	name := mux.Vars(r)["site"]
	if site, err = sc.dbase.GetSiteWithHosts(name); err != nil {
		if site, err = sc.dbase.GetSite(name); err != nil {
			respondError(w, err, "Unable to get site", http.StatusNotFound)
			return
		}
	}
	respondJSON(w, site)
}

func (sc *SiteController) GetSites(w http.ResponseWriter, r *http.Request, _ http.HandlerFunc) {
	var err error
	var sites []*db.Site
	if sites, err = sc.dbase.QuerySites(); err != nil {
		respondError(w, err, "Unable to get sites", http.StatusNotFound)
		return
	}
	respondJSON(w, sites)
}

func (sc *SiteController) PostSite(w http.ResponseWriter, r *http.Request, _ http.HandlerFunc) {
	site := db.Site{}
	if err := json.NewDecoder(r.Body).Decode(&site); err != nil {
		respondError(w, err, JSONDecodeError, http.StatusBadRequest)
		return
	}
	if msg, err := validateSite(&site); err != nil {
		respondError(w, err, msg, http.StatusBadRequest)
		return
	}
	if err := sc.dbase.InsertSite(&site); err != nil {
		respondError(w, err, "DB Error! Is the name unique?", http.StatusBadRequest)
		return
	}
	insSite, err := sc.dbase.GetSite(site.Name)
	if err != nil {
		respondError(w, err, "Unable to get site", http.StatusBadRequest)
		return
	}
	respondJSON(w, &insSite)
}

func (sc *SiteController) PutSite(w http.ResponseWriter, r *http.Request, h http.HandlerFunc) {
	site := db.Site{}
	if err := json.NewDecoder(r.Body).Decode(&site); err != nil {
		respondError(w, err, JSONDecodeError, http.StatusBadRequest)
		return
	}
	if mux.Vars(r)["site"] != site.Name {
		respondError(w, errors.New("Names are not equal"), "Name cannot be changed.",
			http.StatusBadRequest)
		return
	}
	if msg, err := validateSite(&site); err != nil {
		respondError(w, err, msg, http.StatusBadRequest)
		return
	}
	if err := sc.dbase.UpdateSite(&site); err != nil {
		respondError(w, err, "DB Error! Unable to update site", http.StatusBadRequest)
		return
	}
	respondEmpty(w)
}

func (sc *SiteController) DeleteSite(w http.ResponseWriter, r *http.Request, _ http.HandlerFunc) {
	if err := sc.dbase.DeleteSite(mux.Vars(r)["site"]); err != nil {
		respondError(w, err, "Unable to delete site", http.StatusBadRequest)
		return
	}
	respondEmpty(w)
}

func (sc *SiteController) ReloadConfig(w http.ResponseWriter, r *http.Request,
	_ http.HandlerFunc) {
	site, siteCfg, msg, err := sc.getAndValidateSiteCfg(mux.Vars(r)["site"])
	if err != nil {
		respondError(w, err, msg, http.StatusBadRequest)
		return
	}
	if msg, err := sc.writeConfig(site, siteCfg); err != nil {
		respondError(w, err, msg, http.StatusInternalServerError)
		return
	}
}

func (sc *SiteController) GetPathPredicates(w http.ResponseWriter, r *http.Request,
	_ http.HandlerFunc) {
	actions, err := sc.dbase.QueryRawFilters(mux.Vars(r)["site"])
	if err != nil {
		respondError(w, err, "Unable to query path selectors", 400)
		return
	}
	respondJSON(w, actions)
}

func (sc *SiteController) AddPathPredicate(w http.ResponseWriter, r *http.Request,
	_ http.HandlerFunc) {
	filterJSON := db.Filter{}
	if err := json.NewDecoder(r.Body).Decode(&filterJSON); err != nil {
		respondError(w, err, JSONDecodeError, http.StatusBadRequest)
		return
	}
	pp, err := spathmeta.NewPathPredicate(strings.TrimSpace(filterJSON.PP))
	if err != nil {
		respondError(w, err, "Bad path selector string", 400)
		return
	}
	if err := sc.dbase.InsertFilter(mux.Vars(r)["site"], filterJSON.Name, pp); err != nil {
		respondError(w, err, "DB Error! Is the name unique?", 400)
		return
	}
	respondEmpty(w)
}

func (sc *SiteController) PutPathPredicate(w http.ResponseWriter, r *http.Request,
	_ http.HandlerFunc) {
	filterSct := db.Filter{}
	if err := json.NewDecoder(r.Body).Decode(&filterSct); err != nil {
		respondError(w, err, JSONDecodeError, http.StatusBadRequest)
		return
	}
	if mux.Vars(r)["path"] != filterSct.Name {
		respondError(w, errors.New("Names are not equal"), "Name cannot be changed.",
			http.StatusBadRequest)
		return
	}
	pp, err := spathmeta.NewPathPredicate(strings.TrimSpace(filterSct.PP))
	if err != nil {
		respondError(w, err, "Bad path selector string", 400)
		return
	}
	if err := sc.dbase.UpdateFilter(mux.Vars(r)["site"], filterSct.Name, pp); err != nil {
		respondError(w, err, "DB Error! Unable to update PathPredicate!", 400)
		return
	}
	respondEmpty(w)
}

func (sc *SiteController) DeletePathPredicate(w http.ResponseWriter, r *http.Request,
	_ http.HandlerFunc) {
	if err := sc.dbase.DeleteFilter(mux.Vars(r)["site"], mux.Vars(r)["path"]); err != nil {
		respondError(w, err, "Unable to delete Path selector from the database", 400)
		return
	}
	respondEmpty(w)
}
