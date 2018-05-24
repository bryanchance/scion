// Copyright 2018 Anapaya Systems

package ctrl

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"

	"github.com/scionproto/scion/go/lib/spath/spathmeta"
	"github.com/scionproto/scion/go/sigmgmt/config"
	"github.com/scionproto/scion/go/sigmgmt/db"
)

type Controller struct {
	db  *gorm.DB
	cfg *config.Global
}

func NewController(db *gorm.DB, cfg *config.Global) *Controller {
	return &Controller{
		db:  db,
		cfg: cfg,
	}
}

func (c *Controller) GetSite(w http.ResponseWriter, r *http.Request, _ http.HandlerFunc) {
	var site db.Site
	err := c.db.Preload("Hosts").First(&site, mux.Vars(r)["site"]).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			respondNotFound(w)
			return
		}
		respondError(w, err, DBFindError, http.StatusBadRequest)
		return
	}
	respondJSON(w, site)
}

func (c *Controller) GetSites(w http.ResponseWriter, r *http.Request, _ http.HandlerFunc) {
	var sites []*db.Site
	if err := c.db.Find(&sites).Error; err != nil {
		respondError(w, err, DBFindError, http.StatusInternalServerError)
		return
	}
	respondJSON(w, sites)
}

func (c *Controller) PostSite(w http.ResponseWriter, r *http.Request, _ http.HandlerFunc) {
	site := db.Site{}
	if err := json.NewDecoder(r.Body).Decode(&site); err != nil {
		respondError(w, err, JSONDecodeError, http.StatusBadRequest)
		return
	}
	if msg, err := validateSite(&site); err != nil {
		respondError(w, err, msg, http.StatusBadRequest)
		return
	}
	if !c.db.NewRecord(&site) {
		respondError(w, errors.New("new record failed"), "DB Error! Name must be unique!",
			http.StatusBadRequest)
		return
	}
	if !c.createOne(w, &site) {
		return
	}
	selector := db.PathSelector{Name: "any", SiteID: site.ID, Filter: "0-0#0"}
	if !c.createOne(w, &selector) {
		return
	}
	respondJSON(w, &site)
}

func (c *Controller) PutSite(w http.ResponseWriter, r *http.Request, h http.HandlerFunc) {
	site := db.Site{}
	if err := json.NewDecoder(r.Body).Decode(&site); err != nil {
		respondError(w, err, JSONDecodeError, http.StatusBadRequest)
		return
	}
	if msg, err := validateSite(&site); err != nil {
		respondError(w, err, msg, http.StatusBadRequest)
		return
	}
	id, err := strconv.Atoi(mux.Vars(r)["site"])
	if err != nil || int(site.ID) != id {
		respondError(w, nil, IDChangeError, http.StatusBadRequest)
		return
	}
	// Update hosts
	if err := c.updateHosts(site.Hosts, site.ID); err != nil {
		respondError(w, err, "Could not update hosts", http.StatusBadRequest)
		return
	}
	// Update fields
	err = c.db.Model(&site).Updates(
		map[string]interface{}{
			"Name":        site.Name,
			"VHost":       site.VHost,
			"MetricsPort": site.MetricsPort,
		}).Error
	if err != nil {
		respondError(w, err, DBUpdateError, http.StatusBadRequest)
		return
	}
	c.GetSite(w, r, h)
}

func (c *Controller) DeleteSite(w http.ResponseWriter, r *http.Request, _ http.HandlerFunc) {
	if err := c.db.Delete(&db.Site{}, mux.Vars(r)["site"]).Error; err != nil {
		respondError(w, err, DBDeleteError, http.StatusBadRequest)
		return
	}
	respondEmpty(w)
}

func (c *Controller) ReloadConfig(w http.ResponseWriter, r *http.Request,
	_ http.HandlerFunc) {
	siteID, err := strconv.Atoi(mux.Vars(r)["site"])
	if err != nil {
		respondError(w, err, "Bad SiteID", http.StatusBadRequest)
	}
	siteCfg, msg, err := c.getAndValidateSiteCfg(uint(siteID))
	if err != nil {
		respondError(w, err, msg, http.StatusBadRequest)
		return
	}
	if msg, err := c.writeConfig(uint(siteID), siteCfg); err != nil {
		respondError(w, err, msg, http.StatusInternalServerError)
		return
	}
}

func (c *Controller) GetPathSelectors(w http.ResponseWriter, r *http.Request,
	_ http.HandlerFunc) {
	var selectors []db.PathSelector
	if err := c.db.Where("site_id = ?", mux.Vars(r)["site"]).Find(&selectors).Error; err != nil {
		respondError(w, err, DBFindError, http.StatusBadRequest)
		return
	}
	respondJSON(w, selectors)
}

func (c *Controller) PostPathSelector(w http.ResponseWriter, r *http.Request,
	_ http.HandlerFunc) {
	selector := db.PathSelector{}
	if err := json.NewDecoder(r.Body).Decode(&selector); err != nil {
		respondError(w, err, JSONDecodeError, http.StatusBadRequest)
		return
	}
	selector.Filter = strings.TrimSpace(selector.Filter)
	_, err := spathmeta.NewPathPredicate(selector.Filter)
	if err != nil {
		respondError(w, err, PathPredicateError, 400)
		return
	}
	var site db.Site
	if !c.findOne(w, mux.Vars(r)["site"], &site) {
		return
	}
	selector.SiteID = site.ID
	if !c.createOne(w, &selector) {
		return
	}
	respondJSON(w, &selector)
}

func (c *Controller) PutPathSelector(w http.ResponseWriter, r *http.Request,
	_ http.HandlerFunc) {
	selector := db.PathSelector{}
	if err := json.NewDecoder(r.Body).Decode(&selector); err != nil {
		respondError(w, err, JSONDecodeError, http.StatusBadRequest)
		return
	}
	selector.Filter = strings.TrimSpace(selector.Filter)
	_, err := spathmeta.NewPathPredicate(selector.Filter)
	if err != nil {
		respondError(w, err, PathPredicateError, 400)
		return
	}
	id, err := strconv.Atoi(mux.Vars(r)["selector"])
	if err != nil || int(selector.ID) != id {
		respondError(w, nil, IDChangeError, http.StatusBadRequest)
		return
	}
	err = c.db.Model(&selector).Updates(
		map[string]interface{}{
			"Name":   selector.Name,
			"Filter": selector.Filter,
		}).Error
	if err != nil {
		respondError(w, err, DBUpdateError, http.StatusBadRequest)
		return
	}
	respondEmpty(w)
}

func (c *Controller) DeletePathSelector(w http.ResponseWriter, r *http.Request,
	_ http.HandlerFunc) {
	if err := c.db.Delete(&db.PathSelector{}, mux.Vars(r)["selector"]).Error; err != nil {
		respondError(w, err, DBDeleteError, http.StatusBadRequest)
		return
	}
	respondEmpty(w)
}
