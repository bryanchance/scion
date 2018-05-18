// Copyright 2018 Anapaya Systems

package ctrl

import (
	"encoding/json"
	"net/http"

	"github.com/scionproto/scion/go/lib/log"
	sigcfg "github.com/scionproto/scion/go/sig/anaconfig"
)

var (
	JSONDecodeError = "Error decoding JSON"
)

type (
	Policy struct {
		Policy string
	}

	CIDR struct {
		ID   int
		CIDR string
	}

	SIG struct {
		ID        string
		Addr      string
		EncapPort uint16
		CtrlPort  uint16
	}
)

func SIGFromSIGCfg(sig sigcfg.SIG) *SIG {
	return &SIG{ID: string(sig.Id), Addr: sig.Addr.String(),
		EncapPort: sig.EncapPort, CtrlPort: sig.CtrlPort}
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
