// Copyright 2018 Anapaya Systems

package ctrl

import (
	"encoding/json"
	"net"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/sig/sigcmn"
	"github.com/scionproto/scion/go/sigmgmt/db"
)

const (
	iaParseError   = "Unable to parse ISD-AS string"
	cidrParseError = "Unable to parse CIDR address"
)

func (c *Controller) GetASes(w http.ResponseWriter, r *http.Request, _ http.HandlerFunc) {
	var ases []db.ASEntry
	if err := c.db.Where("site_id = ?", mux.Vars(r)["site"]).Find(&ases).Error; err != nil {
		respondError(w, err, DBFindError, http.StatusBadRequest)
		return
	}
	respondJSON(w, ases)
}

func (c *Controller) PostAS(w http.ResponseWriter, r *http.Request, _ http.HandlerFunc) {
	var as *db.ASEntry
	if err := json.NewDecoder(r.Body).Decode(&as); err != nil {
		respondError(w, err, JSONDecodeError, http.StatusBadRequest)
		return
	}
	_, err := as.ToAddrIA()
	if err != nil {
		respondError(w, err, iaParseError, http.StatusBadRequest)
		return
	}
	var site db.Site
	if !c.findOne(w, mux.Vars(r)["site"], &site) {
		return
	}
	as.SiteID = site.ID
	if !c.createOne(w, &as) {
		return
	}
	respondJSON(w, as)
}

func (c *Controller) DeleteAS(w http.ResponseWriter, r *http.Request, _ http.HandlerFunc) {
	if err := c.db.Delete(&db.ASEntry{}, mux.Vars(r)["as"]).Error; err != nil {
		respondError(w, err, DBDeleteError, http.StatusBadRequest)
		return
	}
	respondEmpty(w)
}

func (c *Controller) GetAS(w http.ResponseWriter, r *http.Request, _ http.HandlerFunc) {
	var as db.ASEntry
	if !c.findOne(w, mux.Vars(r)["as"], &as) {
		return
	}
	respondJSON(w, as)
}

func (c *Controller) UpdateAS(w http.ResponseWriter, r *http.Request, _ http.HandlerFunc) {
	var as *db.ASEntry
	if err := json.NewDecoder(r.Body).Decode(&as); err != nil {
		respondError(w, err, JSONDecodeError, http.StatusBadRequest)
		return
	}
	id, err := strconv.Atoi(mux.Vars(r)["as"])
	if err != nil || int(as.ID) != id {
		respondError(w, nil, IDChangeError, http.StatusBadRequest)
		return
	}
	if err := c.db.Model(&as).Update("Name", as.Name).Error; err != nil {
		respondError(w, err, DBUpdateError, http.StatusBadRequest)
		return
	}
	respondEmpty(w)
}

func (c *Controller) UpdatePolicy(w http.ResponseWriter, r *http.Request, _ http.HandlerFunc) {
	var as, dbAS db.ASEntry
	if err := json.NewDecoder(r.Body).Decode(&as); err != nil {
		respondError(w, err, JSONDecodeError, http.StatusBadRequest)
		return
	}
	if !c.findOne(w, mux.Vars(r)["as"], &dbAS) {
		return
	}
	dbAS.Policies = as.Policies
	if str, err := c.validatePolicies(&dbAS); err != nil {
		respondError(w, err, str, http.StatusBadRequest)
		return
	}
	if err := c.db.Model(&as).Update("Policies", as.Policies).Error; err != nil {
		respondError(w, err, DBUpdateError, http.StatusBadRequest)
		return
	}
	respondEmpty(w)
}

func (c *Controller) GetNetworks(w http.ResponseWriter, r *http.Request, _ http.HandlerFunc) {
	var networks []db.Network
	if err := c.db.Where("as_entry_id = ?", mux.Vars(r)["as"]).Find(&networks).Error; err != nil {
		respondError(w, err, DBFindError, http.StatusBadRequest)
		return
	}
	respondJSON(w, networks)
}

func (c *Controller) PostNetwork(w http.ResponseWriter, r *http.Request, _ http.HandlerFunc) {
	var network db.Network
	if err := json.NewDecoder(r.Body).Decode(&network); err != nil {
		respondError(w, err, JSONDecodeError, http.StatusBadRequest)
		return
	}
	_, _, err := net.ParseCIDR(network.CIDR)
	if err != nil {
		respondError(w, err, cidrParseError, 400)
		return
	}
	var as db.ASEntry
	if !c.findOne(w, mux.Vars(r)["as"], &as) {
		return
	}
	network.ASEntryID = as.ID
	if !c.createOne(w, &network) {
		return
	}
	respondJSON(w, network)
}

func (c *Controller) DeleteNetwork(w http.ResponseWriter, r *http.Request, _ http.HandlerFunc) {
	if err := c.db.Delete(&db.Network{}, mux.Vars(r)["network"]).Error; err != nil {
		respondError(w, err, DBDeleteError, http.StatusBadRequest)
		return
	}
	respondEmpty(w)
}

func (c *Controller) GetSIGs(w http.ResponseWriter, r *http.Request, _ http.HandlerFunc) {
	var sigs []db.SIG
	if err := c.db.Where("as_entry_id = ?", mux.Vars(r)["as"]).Find(&sigs).Error; err != nil {
		respondError(w, err, DBFindError, http.StatusBadRequest)
		return
	}
	respondJSON(w, sigs)
}

func (c *Controller) GetDefaultSIG(w http.ResponseWriter, r *http.Request, _ http.HandlerFunc) {
	respondJSON(w, db.SIG{CtrlPort: sigcmn.DefaultCtrlPort, EncapPort: sigcmn.DefaultEncapPort})
}

func (c *Controller) PostSIG(w http.ResponseWriter, r *http.Request, _ http.HandlerFunc) {
	var sig *db.SIG
	if err := json.NewDecoder(r.Body).Decode(&sig); err != nil {
		respondError(w, err, JSONDecodeError, http.StatusBadRequest)
		return
	}
	if addr := net.ParseIP(sig.Address); addr == nil {
		respondError(
			w, common.NewBasicError(IPParseError, nil, "address", sig.Address), IPParseError, 400)
		return
	}
	var as db.ASEntry
	if !c.findOne(w, mux.Vars(r)["as"], &as) {
		return
	}
	sig.ASEntryID = as.ID
	if !c.createOne(w, &sig) {
		return
	}
	respondJSON(w, sig)
}

func (c *Controller) UpdateSIG(w http.ResponseWriter, r *http.Request, _ http.HandlerFunc) {
	var sig *db.SIG
	if err := json.NewDecoder(r.Body).Decode(&sig); err != nil {
		respondError(w, err, JSONDecodeError, http.StatusBadRequest)
		return
	}
	if addr := net.ParseIP(sig.Address); addr == nil {
		respondError(
			w, common.NewBasicError(IPParseError, nil, "address", sig.Address), IPParseError, 400)
		return
	}
	id, err := strconv.Atoi(mux.Vars(r)["sig"])
	if err != nil || int(sig.ID) != id {
		respondError(w, nil, IDChangeError, http.StatusBadRequest)
		return
	}
	err = c.db.Model(&sig).Updates(
		map[string]interface{}{
			"Name":      sig.Name,
			"Address":   sig.Address,
			"CtrlPort":  sig.CtrlPort,
			"EncapPort": sig.EncapPort,
		}).Error
	if err != nil {
		respondError(w, err, DBUpdateError, http.StatusBadRequest)
		return
	}
	respondEmpty(w)
}

func (c *Controller) DeleteSIG(w http.ResponseWriter, r *http.Request, _ http.HandlerFunc) {
	if err := c.db.Delete(&db.SIG{}, mux.Vars(r)["sig"]).Error; err != nil {
		respondError(w, err, DBDeleteError, http.StatusBadRequest)
		return
	}
	respondEmpty(w)
}
