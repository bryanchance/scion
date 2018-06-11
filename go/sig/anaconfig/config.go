// Copyright 2018 Anapaya Systems

// Package config is responsible for parsing the SIG json config file into a
// set of simple intermediate data-structures.
package config

import (
	"encoding/json"
	"io/ioutil"

	"github.com/scionproto/scion/go/lib/addr"
	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/lib/pktcls"
	"github.com/scionproto/scion/go/sig/config"
	"github.com/scionproto/scion/go/sig/mgmt"
)

// Alias to open source types to allow client packages to switch between open
// source/anapaya versions via a single import change.

type IPNet = config.IPNet
type SIG = config.SIG
type SIGSet = config.SIGSet

// Cfg is a direct Go representation of the JSON file format.
type Cfg struct {
	ASes          map[addr.IA]*ASEntry
	Classes       pktcls.ClassMap
	Actions       pktcls.ActionMap
	ConfigVersion uint64
}

// Load a JSON config file from path and parse it into a Cfg struct.
func LoadFromFile(path string) (*Cfg, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, common.NewBasicError("Unable to open SIG config", err)
	}
	cfg := &Cfg{}
	if err := json.Unmarshal(b, cfg); err != nil {
		return nil, common.NewBasicError("Unable to parse SIG config", err)
	}
	if err := cfg.postprocess(); err != nil {
		return nil, err
	}
	return cfg, nil
}

// postprocess fills in sig IDs in cfg, and checks that action and class names
// reference valid objects.
func (cfg *Cfg) postprocess() error {
	for ia, ae := range cfg.ASes {
		for id, sig := range ae.Sigs {
			sig.Id = id
		}

		// Check that each session references an action that exists
		for sessId, actName := range ae.Sessions {
			if actName == "" {
				continue
			}
			if _, ok := cfg.Actions[actName]; !ok {
				return common.NewBasicError("Unknown action name", nil,
					"ia", ia, "sessId", sessId, "action", actName)
			}
		}
		// check that each policy references a class that exists
		for i, pol := range ae.PktPolicies {
			if _, ok := cfg.Classes[pol.ClassName]; !ok {
				return common.NewBasicError("Unknown class name", nil,
					"ia", ia, "polIdx", i, "class", pol.ClassName)
			}
			// check that each session ID in the policy references a session that exists
			for _, sessId := range pol.SessIds {
				if _, ok := ae.Sessions[sessId]; !ok {
					return common.NewBasicError("Unknown session id", nil,
						"ia", ia, "polIdx", i, "class", pol.ClassName, "sessId", sessId)
				}
			}
		}
	}

	return nil
}

type ASEntry struct {
	// Common with open source ASEntry

	Nets []*config.IPNet `json:",omitempty"`
	Sigs config.SIGSet   `json:",omitempty"`

	// Anapaya specific

	Sessions    SessionMap   `json:",omitempty"`
	PktPolicies []*PktPolicy `json:",omitempty"`
}

type SessionMap map[mgmt.SessionType]string

type PktPolicy struct {
	Name      string
	ClassName string
	SessIds   []mgmt.SessionType
}

type Session struct {
	ID      mgmt.SessionType
	PolName string
	Pred    *pktcls.ActionFilterPaths
}

type SessionSet map[mgmt.SessionType]*Session

func BuildSessions(cfgSessMap SessionMap, actions pktcls.ActionMap) (SessionSet, error) {
	set := make(SessionSet)
	for sessId, actName := range cfgSessMap {
		act := actions[actName]
		afp, ok := act.(*pktcls.ActionFilterPaths)
		if actName != "" && !ok {
			// Unable to find the Action the session is referencing
			return nil, common.NewBasicError("Unable to find referenced Action", nil,
				"sessionID", sessId, "action", actName)
		}
		set[sessId] = &Session{
			ID:      sessId,
			PolName: actName,
			Pred:    afp,
		}
	}
	return set, nil
}
