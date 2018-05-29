// Copyright 2018 Anapaya Systems

package ctrl

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/jinzhu/gorm"

	"github.com/scionproto/scion/go/lib/log"
)

const (
	JSONDecodeError    = "Error decoding JSON"
	IDChangeError      = "ID can not be changed"
	IPParseError       = "IP Address is not valid"
	DBFindError        = "Error finding object in DB"
	DBCreateError      = "Error creating object in DB"
	DBUniqueError      = "Unique Constraint Error"
	DBUpdateError      = "Error updating object in DB"
	DBDeleteError      = "Error deleting object in DB"
	PathPredicateError = "Bad Path Predicate"
)

func respond(w http.ResponseWriter, data []byte, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(data)
}

func respondJSON(w http.ResponseWriter, v interface{}) {
	data, err := json.Marshal(v)
	if err != nil {
		respondError(w, err, "Could not create JSON response", http.StatusInternalServerError)
		return
	}
	respond(w, data, http.StatusOK)
}

func respondEmpty(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

func respondNotFound(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNotFound)
}

func respondError(w http.ResponseWriter, err error, errStr string, status int) {
	log.Error("Error handling request", "err", err, "message", errStr)
	errorMsg := map[string]string{"error": errStr}
	data, err := json.Marshal(errorMsg)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Error("Could not create JSON error response", "err", err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(data)
}

func (c *Controller) findOne(w http.ResponseWriter, key string, entity interface{}) bool {
	err := c.db.First(entity, key).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			respondNotFound(w)
			return false
		}
		respondError(w, err, DBFindError, http.StatusBadRequest)
		return false
	}
	return true
}

func (c *Controller) createOne(w http.ResponseWriter, entity interface{}) bool {
	if err := c.db.Create(entity).Error; err != nil {
		if strings.HasPrefix(err.Error(), "UNIQUE") {
			respondError(w, err, DBUniqueError, http.StatusBadRequest)
			return false
		}
		respondError(w, err, DBCreateError, http.StatusBadRequest)
		return false
	}
	return true
}
