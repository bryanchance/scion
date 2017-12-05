// Copyright 2017 ETH Zurich
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package config is responsible for parsing the SIG json config file into a
// set of simple intermediate data-structures.
package config

import (
	"encoding/json"
	"io/ioutil"
	"net"

	"github.com/scionproto/scion/go/lib/addr"
	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/lib/pathmgr"
	"github.com/scionproto/scion/go/lib/pktcls"
	"github.com/scionproto/scion/go/sig/mgmt"
	"github.com/scionproto/scion/go/sig/siginfo"
)

// Cfg is a direct Go representation of the JSON file format.
type Cfg struct {
	ASes    map[addr.ISD_AS]*ASEntry
	Classes pktcls.ClassMap
	Actions pktcls.ActionMap
}

// Load a JSON config file from path and parse it into a Cfg struct.
func LoadFromFile(path string) (*Cfg, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, common.NewCError("Unable to open SIG config", "err", err)
	}
	cfg, err := Parse(b)
	if err != nil {
		return nil, err
	}
	if len(cfg.ASes) == 0 {
		return nil, common.NewCError("Empty ASTable in config")
	}
	for ia, ae := range cfg.ASes {
		for sessId, actName := range ae.Sessions {
			if actName == "" {
				continue
			}
			if _, ok := cfg.Actions[actName]; !ok {
				return nil, common.NewCError("Unknown action name", "ia", ia, "sessId", sessId,
					"action", actName)
			}
		}
		for i, pol := range ae.PktPolicies {
			if _, ok := cfg.Classes[pol.ClassName]; !ok {
				return nil, common.NewCError("Unknown class name", "ia", ia, "polIdx", i,
					"class", pol.ClassName)
			}
			for _, sessId := range pol.SessIds {
				if _, ok := ae.Sessions[sessId]; !ok {
					return nil, common.NewCError("Unknown session id", "ia", ia, "polIdx", i,
						"class", pol.ClassName, "sessId", sessId)
				}
			}
		}
	}
	return cfg, nil
}

// Parse a JSON config from b into a Cfg struct.
func Parse(b common.RawBytes) (*Cfg, error) {
	cfg := &Cfg{}
	if err := json.Unmarshal(b, cfg); err != nil {
		return nil, common.NewCError("Unable to parse SIG config", "err", err)
	}
	// Populate IDs
	for _, as := range cfg.ASes {
		for id := range as.Sigs {
			sig := as.Sigs[id]
			sig.Id = id
		}
	}
	return cfg, nil
}

type SessionMap map[mgmt.SessionType]string

type ASEntry struct {
	Nets        []*IPNet
	Sigs        SIGSet
	Sessions    SessionMap
	PktPolicies []*PktPolicy
}

// IPNet is custom type of net.IPNet, to allow custom unmarshalling.
type IPNet net.IPNet

func (in *IPNet) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return common.NewCError("Unable to unmarshal IPnet from JSON", "raw", b, "err", err)
	}
	_, ipnet, err := net.ParseCIDR(s)
	if err != nil {
		return common.NewCError("Unable to parse IPnet string", "raw", s, "err", err)
	}
	*in = IPNet(*ipnet)
	return nil
}

func (in *IPNet) IPNet() *net.IPNet {
	return (*net.IPNet)(in)
}

// SIG represents a SIG in a remote IA.
type SIG struct {
	Id        siginfo.SigIdType
	Addr      net.IP
	CtrlPort  uint16
	EncapPort uint16
}

type SIGSet map[siginfo.SigIdType]*SIG

type PktPolicy struct {
	SessionId mgmt.SessionType
	ClassName string
	SessIds   []mgmt.SessionType
}

type Session struct {
	ID      mgmt.SessionType
	PolName string
	Pred    *pathmgr.PathPredicate
}

type SessionSet map[mgmt.SessionType]*Session

func BuildSessions(cfgSessMap SessionMap, actions pktcls.ActionMap) (SessionSet, error) {
	set := make(SessionSet)
	for sessId, actName := range cfgSessMap {
		act := actions[actName]
		var pred *pathmgr.PathPredicate
		if afp, ok := act.(*pktcls.ActionFilterPaths); ok {
			pred = afp.Contains
		}
		if actName != "" && pred == nil {
			// Unable to find the Action the session is referencing
			return nil, common.NewCError("Unable to find referenced Action", "sessionID", sessId,
				"action", actName)
		}
		set[sessId] = &Session{
			ID:      sessId,
			PolName: actName,
			Pred:    pred,
		}
	}
	return set, nil
}
