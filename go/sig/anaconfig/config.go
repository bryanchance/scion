// Copyright 2018 Anapaya Systems

// Package config is responsible for parsing the SIG json config file into a
// set of simple intermediate data-structures.
package config

import (
	"encoding/json"
	"io/ioutil"

	"github.com/scionproto/scion/go/lib/addr"
	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/lib/pathpol"
	"github.com/scionproto/scion/go/lib/pktcls"
	"github.com/scionproto/scion/go/sig/config"
	"github.com/scionproto/scion/go/sig/mgmt"
)

// Alias to open source types to allow client packages to switch between open
// source/anapaya versions via a single import change.

type IPNet = config.IPNet

// Cfg is a direct Go representation of the JSON file format.
type Cfg struct {
	ASes          map[addr.IA]*ASEntry
	PktClasses    pktcls.ClassMap   `json:",omitempty"`
	PathPolicies  pathpol.PolicyMap `json:",omitempty"`
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

// postprocess fills in sig IDs in cfg, and checks that policy and class names
// reference valid objects.
func (cfg *Cfg) postprocess() error {
	for ia, ae := range cfg.ASes {
		// Check that each session references an policy that exists
		for sessId, polName := range ae.Sessions {
			if polName == "" {
				continue
			}
			if _, ok := cfg.PathPolicies[polName]; !ok {
				return common.NewBasicError("Unknown policy name", nil,
					"ia", ia, "sessId", sessId, "policy", polName)
			}
		}
		// check that each policy references a class that exists
		for i, pol := range ae.PktPolicies {
			if _, ok := cfg.PktClasses[pol.ClassName]; !ok {
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
		if ae.Nets == nil {
			ae.Nets = []*config.IPNet{}
		}
	}
	return nil
}

type ASEntry struct {
	// Common with open source ASEntry
	Name string          `json:",omitempty"`
	Nets []*config.IPNet `json:",omitempty"`

	// Anapaya specific

	Sessions    SessionMap   `json:",omitempty"`
	PktPolicies []*PktPolicy `json:",omitempty"`
}

// SessionMap provides a mapping from session id to path policy name
type SessionMap map[mgmt.SessionType]string

type PktPolicy struct {
	Name      string `json:",omitempty"`
	ClassName string
	SessIds   []mgmt.SessionType
}

type Session struct {
	ID      mgmt.SessionType
	PolName string
	Policy  *pathpol.Policy
}

// SessionSet is a mapping from session id to Session
type SessionSet map[mgmt.SessionType]*Session

// BuildSessions creates a SessionSet from a SessionMap and the referenced Policies
func BuildSessions(cfgSessMap SessionMap, polMap pathpol.PolicyMap) (SessionSet, error) {
	set := make(SessionSet)
	for sessId, polName := range cfgSessMap {
		ep, ok := polMap[polName]
		if polName != "" && !ok {
			// Unable to find the Policy the session is referencing
			return nil, common.NewBasicError("Unable to find referenced Policy", nil,
				"sessionId", sessId, "policy", polName)
		}
		var pp *pathpol.Policy
		// Session has no policy specified, generate allow-all fallback
		if polName != "" {
			policies := make([]*pathpol.ExtPolicy, 0, len(polMap))
			for _, policy := range polMap {
				policies = append(policies, policy)
			}
			var err error
			if pp, err = pathpol.PolicyFromExtPolicy(ep, policies); err != nil {
				return nil, common.NewBasicError("Unable to create Policy from ExtPolicy", err)
			}
		}
		set[sessId] = &Session{
			ID:      sessId,
			PolName: polName,
			Policy:  pp,
		}
	}
	return set, nil
}
