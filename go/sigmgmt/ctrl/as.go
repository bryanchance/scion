// Copyright 2018 Anapaya Systems

package ctrl

import (
	"encoding/json"
	"net"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/scionproto/scion/go/lib/addr"
	sigcfg "github.com/scionproto/scion/go/sig/config"
	"github.com/scionproto/scion/go/sig/siginfo"
	"github.com/scionproto/scion/go/sigmgmt/config"
	"github.com/scionproto/scion/go/sigmgmt/db"
)

type ASController struct {
	dbase *db.DB
	cfg   *config.Global
}

func NewASController(dbase *db.DB, cfg *config.Global) *ASController {
	return &ASController{
		dbase: dbase,
		cfg:   cfg,
	}
}

func (ac ASController) GetIAs(w http.ResponseWriter, r *http.Request, _ http.HandlerFunc) {
	var err error
	var ias []addr.IA
	if ias, err = ac.dbase.QueryASes(mux.Vars(r)["site"]); err != nil {
		respondError(w, err, "Unable to get IAs", 400)
		return
	}

	iaList := []IA{}
	for _, ia := range ias {
		iaList = append(iaList, *IAFromAddrIA(ia))
	}
	respondJSON(w, iaList)
}

func (ac ASController) PostIA(w http.ResponseWriter, r *http.Request, _ http.HandlerFunc) {
	var iaSct *IA
	if err := json.NewDecoder(r.Body).Decode(&iaSct); err != nil {
		respondError(w, err, JSONDecodeError, http.StatusBadRequest)
		return
	}
	ia, err := iaSct.ToAddrIA()
	if err != nil {
		respondError(w, err, IAParseError, http.StatusBadRequest)
		return
	}
	if err := ac.dbase.InsertAS(mux.Vars(r)["site"], ia); err != nil {
		respondError(w, err, "DB Error! IS the ISD-AS identifier unique?", 400)
		return
	}
	respondJSON(w, iaSct)
}

func (ac ASController) DeleteIA(w http.ResponseWriter, r *http.Request, _ http.HandlerFunc) {
	ia, msg, err := getIA(r)
	if err != nil {
		respondError(w, err, msg, http.StatusBadRequest)
		return
	}
	if err := ac.dbase.DeleteAS(mux.Vars(r)["site"], *ia); err != nil {
		respondError(w, err, "Unable to delete AS from database", 400)
		return
	}
	respondEmpty(w)
}

func (ac ASController) GetPolicy(w http.ResponseWriter, r *http.Request, _ http.HandlerFunc) {
	ia, msg, err := getIA(r)
	if err != nil {
		respondError(w, err, msg, 400)
		return
	}
	policy := Policy{}
	if policy.Policy, err = ac.dbase.GetPolicy(mux.Vars(r)["site"], *ia); err != nil {
		respondError(w, err, "Unable to get Policy", 400)
		return
	}
	respondJSON(w, policy)
}

func (ac ASController) UpdatePolicy(w http.ResponseWriter, r *http.Request, _ http.HandlerFunc) {
	site := mux.Vars(r)["site"]
	ia, msg, err := getIA(r)
	if err != nil {
		respondError(w, err, msg, http.StatusBadRequest)
		return
	}
	var policy Policy
	if err := json.NewDecoder(r.Body).Decode(&policy); err != nil {
		respondError(w, err, JSONDecodeError, http.StatusBadRequest)
		return
	}
	policies := map[addr.IA]string{
		*ia: policy.Policy,
	}
	if str, err := ac.validatePolicies(site, policies); err != nil {
		respondError(w, err, str, http.StatusBadRequest)
		return
	}
	if err = ac.dbase.SetPolicy(site, *ia, policy.Policy); err != nil {
		respondError(w, err, "DB Error! Unable to set Policy", 400)
		return
	}
	respondEmpty(w)
}

func (ac ASController) GetNetworks(w http.ResponseWriter, r *http.Request, _ http.HandlerFunc) {
	ia, msg, err := getIA(r)
	if err != nil {
		respondError(w, err, msg, http.StatusBadRequest)
		return
	}
	var networks map[int]*sigcfg.IPNet
	if networks, err = ac.dbase.QueryNetworks(mux.Vars(r)["site"], *ia); err != nil {
		respondError(w, err, "Unable to get networks", http.StatusBadRequest)
		return
	}
	cidr := []CIDR{}
	for id, network := range networks {
		cidr = append(cidr, CIDR{ID: id, CIDR: network.String()})
	}
	respondJSON(w, cidr)
}

func (ac ASController) PostNetwork(w http.ResponseWriter, r *http.Request, _ http.HandlerFunc) {
	ia, msg, err := getIA(r)
	if err != nil {
		respondError(w, err, msg, http.StatusBadRequest)
		return
	}
	var cidrSct *CIDR
	if err := json.NewDecoder(r.Body).Decode(&cidrSct); err != nil {
		respondError(w, err, JSONDecodeError, http.StatusBadRequest)
		return
	}
	_, cidr, err := net.ParseCIDR(cidrSct.CIDR)
	if err != nil {
		respondError(w, err, "Unable to parse CIDR address", 400)
		return
	}
	if err := ac.dbase.InsertNetwork(mux.Vars(r)["site"], *ia, cidr); err != nil {
		respondError(w, err, "DB Error!", 400)
		return
	}
	respondJSON(w, cidrSct)
}

func (ac ASController) DeleteNetwork(w http.ResponseWriter, r *http.Request, _ http.HandlerFunc) {
	ia, msg, err := getIA(r)
	if err != nil {
		respondError(w, err, msg, http.StatusBadRequest)
		return
	}
	network, err := strconv.ParseInt(mux.Vars(r)["network"], 10, 32)
	if err := ac.dbase.DeleteNetwork(mux.Vars(r)["site"], *ia, int(network)); err != nil {
		respondError(w, err, "Unable to delete network", http.StatusBadRequest)
		return
	}
	respondEmpty(w)
}

func (ac ASController) GetSIGs(w http.ResponseWriter, r *http.Request, _ http.HandlerFunc) {
	ia, msg, err := getIA(r)
	if err != nil {
		respondError(w, err, msg, http.StatusBadRequest)
		return
	}
	var sigMap sigcfg.SIGSet
	if sigMap, err = ac.dbase.QuerySIGs(mux.Vars(r)["site"], *ia); err != nil {
		respondError(w, err, "Unable to get SIGs", 400)
		return
	}
	sigs := []*SIG{}
	for _, sig := range sigMap {
		sigs = append(sigs, &SIG{ID: string(sig.Id), Addr: sig.Addr.String(),
			CtrlPort: sig.CtrlPort, EncapPort: sig.EncapPort})
	}
	respondJSON(w, sigs)
}

func (ac ASController) PostSIG(w http.ResponseWriter, r *http.Request, _ http.HandlerFunc) {
	ia, msg, err := getIA(r)
	if err != nil {
		respondError(w, err, msg, http.StatusBadRequest)
		return
	}
	var sigSct *SIG
	if err := json.NewDecoder(r.Body).Decode(&sigSct); err != nil {
		respondError(w, err, JSONDecodeError, http.StatusBadRequest)
		return
	}
	sig := sigcfg.SIG{}
	sig.Id = siginfo.SigIdType(sigSct.ID)
	if sig.Addr = net.ParseIP(sigSct.Addr); sig.Addr == nil {
		respondError(w, err, "IP Address is not valid!", 400)
		return
	}
	sig.CtrlPort = sigSct.CtrlPort
	sig.EncapPort = sigSct.EncapPort
	if err = ac.dbase.InsertSIG(mux.Vars(r)["site"], *ia, &sig); err != nil {
		respondError(w, err, "DB Error! Is the name unique?", 400)
		return
	}
	respondJSON(w, sigSct)
}

func (ac ASController) UpdateSIG(w http.ResponseWriter, r *http.Request, _ http.HandlerFunc) {
	ia, msg, err := getIA(r)
	if err != nil {
		respondError(w, err, msg, http.StatusBadRequest)
		return
	}
	var sigSct *SIG
	if err := json.NewDecoder(r.Body).Decode(&sigSct); err != nil {
		respondError(w, err, JSONDecodeError, http.StatusBadRequest)
		return
	}
	sig := sigcfg.SIG{}
	sig.Id = siginfo.SigIdType(sigSct.ID)
	if sig.Addr = net.ParseIP(sigSct.Addr); sig.Addr == nil {
		respondError(w, err, "IP Address is not valid!", 400)
		return
	}
	sig.CtrlPort = sigSct.CtrlPort
	sig.EncapPort = sigSct.EncapPort
	if err = ac.dbase.UpdateSIG(mux.Vars(r)["site"], *ia, &sig); err != nil {
		respondError(w, err, "DB Error! Unable to update SIG", 400)
		return
	}
	respondEmpty(w)
}

func (ac ASController) DeleteSIG(w http.ResponseWriter, r *http.Request, _ http.HandlerFunc) {
	ia, msg, err := getIA(r)
	if err != nil {
		respondError(w, err, msg, http.StatusBadRequest)
		return
	}
	if err = ac.dbase.DeleteSIG(mux.Vars(r)["site"], *ia, mux.Vars(r)["sig"]); err != nil {
		respondError(w, err, "Unable to get AS for site", 400)
		return
	}
	respondEmpty(w)
}

func (ac ASController) GetSessions(w http.ResponseWriter, r *http.Request, _ http.HandlerFunc) {
	ia, msg, err := getIA(r)
	if err != nil {
		respondError(w, err, msg, http.StatusBadRequest)
		return
	}
	var sessionMap sigcfg.SessionMap
	if sessionMap, err = ac.dbase.QuerySessions(mux.Vars(r)["site"], *ia); err != nil {
		respondError(w, err, "Unable to get sessions", 400)
		return
	}
	sessions := []Session{}
	for id, path := range sessionMap {
		sessions = append(sessions, Session{ID: uint8(id), FilterName: path})
	}
	respondJSON(w, sessions)
}

func (ac ASController) GetSessionAliases(w http.ResponseWriter, r *http.Request,
	_ http.HandlerFunc) {
	ia, msg, err := getIA(r)
	if err != nil {
		respondError(w, err, msg, http.StatusBadRequest)
		return
	}
	var sessionAliasMap map[addr.IA]db.SessionAliasMap
	if sessionAliasMap, err = ac.dbase.GetSessionAliases(mux.Vars(r)["site"]); err != nil {
		respondError(w, err, "Unable to get sessions aliases", 400)
		return
	}
	sessionAliases := []SessionAlias{}
	for name, sessions := range sessionAliasMap[*ia] {
		sessionAliases = append(sessionAliases, SessionAlias{Name: name, Sessions: sessions})
	}
	respondJSON(w, sessionAliases)
}

func (ac ASController) PostSession(w http.ResponseWriter, r *http.Request, _ http.HandlerFunc) {
	ia, msg, err := getIA(r)
	if err != nil {
		respondError(w, err, msg, http.StatusBadRequest)
		return
	}
	var session *Session
	if err := json.NewDecoder(r.Body).Decode(&session); err != nil {
		respondError(w, err, JSONDecodeError, http.StatusBadRequest)
		return
	}
	if err = ac.dbase.InsertSession(mux.Vars(r)["site"], *ia,
		session.ID, session.FilterName); err != nil {
		respondError(w, err, "DB Error! Is the ID unique?", 400)
		return
	}
	respondJSON(w, session)
}

func (ac ASController) DeleteSession(w http.ResponseWriter, r *http.Request, _ http.HandlerFunc) {
	ia, msg, err := getIA(r)
	if err != nil {
		respondError(w, err, msg, http.StatusBadRequest)
		return
	}
	sessionID, err := strconv.ParseUint(mux.Vars(r)["session"], 10, 8)
	if err = ac.dbase.DeleteSession(mux.Vars(r)["site"], *ia, uint8(sessionID)); err != nil {
		respondError(w, err, "Unable to delete session", 400)
		return
	}
	respondEmpty(w)
}
