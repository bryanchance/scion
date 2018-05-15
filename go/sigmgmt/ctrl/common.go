// Copyright 2018 Anapaya Systems

package ctrl

import (
	"encoding/json"
	"fmt"
	"net/http"

	log "github.com/inconshreveable/log15"

	"github.com/scionproto/scion/go/lib/addr"
)

var (
	JSONDecodeError = "Error decoding JSON"
)

type (
	IA struct {
		ISD string
		AS  string
	}

	Policy struct {
		Policy string
	}

	CIDR struct {
		ID   int
		CIDR string
	}

	Session struct {
		ID         uint8
		FilterName string
	}

	SessionAlias struct {
		Name     string
		Sessions string
	}

	SIG struct {
		ID        string
		Addr      string
		EncapPort uint16
		CtrlPort  uint16
	}
)

func IAFromAddrIA(ia addr.IA) *IA {
	return &IA{ISD: fmt.Sprint(ia.I), AS: ia.A.String()}
}

func (ia *IA) ToAddrIA() (addr.IA, error) {
	iaStr := fmt.Sprintf("%s-%s", ia.ISD, ia.AS)
	return addr.IAFromString(iaStr)
}

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
