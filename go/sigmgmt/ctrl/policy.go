// Copyright 2018 Anapaya Systems

package ctrl

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"

	"github.com/scionproto/scion/go/sigmgmt/db"
)

// GetPathPolicy fetches a single PathPolicyFile
func (c *Controller) GetPathPolicy(w http.ResponseWriter, r *http.Request,
	_ http.HandlerFunc) {
	var policy db.PathPolicyFile
	err := c.db.First(&policy, mux.Vars(r)["policy"]).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			respondNotFound(w)
			return
		}
		respondError(w, err, DBFindError, http.StatusBadRequest)
		return
	}
	err = json.Unmarshal([]byte(policy.CodeStr), &policy.Code)
	if err != nil {
		respondError(w, err, JSONDecodeError, http.StatusBadRequest)
		return
	}
	respondJSON(w, policy)
}

// GetPathPolicies fetches all PathPolicyFiles
func (c *Controller) GetPathPolicies(w http.ResponseWriter, r *http.Request,
	_ http.HandlerFunc) {
	var policies []db.PathPolicyFile
	if err := c.db.Order("id asc").Find(&policies).Error; err != nil {
		respondError(w, err, DBFindError, http.StatusBadRequest)
		return
	}
	for i := range policies {
		err := json.Unmarshal([]byte(policies[i].CodeStr), &policies[i].Code)
		if err != nil {
			respondError(w, err, JSONDecodeError, http.StatusBadRequest)
			return
		}
	}
	respondJSON(w, policies)
}

// PostPathPolicy creates a PathPolicyFile
func (c *Controller) PostPathPolicy(w http.ResponseWriter, r *http.Request,
	_ http.HandlerFunc) {
	policy := db.PathPolicyFile{}
	if err := json.NewDecoder(r.Body).Decode(&policy); err != nil {
		respondError(w, err, JSONDecodeError, http.StatusBadRequest)
		return
	}
	str, err := json.Marshal(policy.Code)
	if err != nil {
		respondError(w, err, JSONDecodeError, http.StatusBadRequest)
		return
	}
	policy.CodeStr = string(str)
	if policy.Type == db.GlobalPolicy {
		policy.SiteID = nil
	}
	err = c.validatePathPolicyFile(policy)
	if err != nil {
		respondError(w, err, PathPolicyError, http.StatusBadRequest)
	}
	if !c.createOne(w, &policy) {
		return
	}
	respondJSON(w, &policy)
}

// PutPathPolicy updates a PathPolicyFile
func (c *Controller) PutPathPolicy(w http.ResponseWriter, r *http.Request,
	_ http.HandlerFunc) {
	policy := db.PathPolicyFile{}
	if err := json.NewDecoder(r.Body).Decode(&policy); err != nil {
		respondError(w, err, JSONDecodeError, http.StatusBadRequest)
		return
	}
	str, err := json.Marshal(policy.Code)
	if err != nil {
		respondError(w, err, JSONDecodeError, http.StatusBadRequest)
		return
	}
	err = c.validatePathPolicyFile(policy)
	if err != nil {
		respondError(w, err, PathPolicyError, http.StatusBadRequest)
	}
	id, err := strconv.Atoi(mux.Vars(r)["policy"])
	if err != nil || int(policy.ID) != id {
		respondError(w, nil, IDChangeError, http.StatusBadRequest)
		return
	}
	err = c.db.Model(&policy).Updates(
		map[string]interface{}{
			"Name":    policy.Name,
			"CodeStr": string(str),
		}).Error
	if err != nil {
		respondError(w, err, DBUpdateError, http.StatusBadRequest)
		return
	}
	respondEmpty(w)
}

// ValidatePathPolicy validate a PathPolicyFile
func (c *Controller) ValidatePathPolicy(w http.ResponseWriter, r *http.Request,
	_ http.HandlerFunc) {
	policyFile := db.PathPolicyFile{}
	if err := json.NewDecoder(r.Body).Decode(&policyFile); err != nil {
		respondError(w, err, JSONDecodeError, http.StatusBadRequest)
		return
	}
	_, err := json.Marshal(policyFile.Code)
	if err != nil {
		respondError(w, err, JSONDecodeError, http.StatusBadRequest)
		return
	}
	err = c.validatePathPolicyFile(policyFile)
	if err != nil {
		respondError(w, err, PathPolicyError, http.StatusBadRequest)
		return
	}
	respondJSON(w, &policyFile)
}

func (c *Controller) DeletePathPolicy(w http.ResponseWriter, r *http.Request, _ http.HandlerFunc) {
	if err := c.db.Delete(&db.PathPolicyFile{}, mux.Vars(r)["policy"]).Error; err != nil {
		respondError(w, err, DBDeleteError, http.StatusBadRequest)
		return
	}
	respondEmpty(w)
}
