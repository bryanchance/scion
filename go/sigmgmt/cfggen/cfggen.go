// Copyright 2018 Anapaya Systems

// Package cfggen can be used to populate a SIG configuration with classes and
// packet policies derived from text policy rules.
//
// See the static markdown documentation for more information about the fields.
package cfggen

import (
	"strconv"
	"strings"

	"github.com/scionproto/scion/go/lib/addr"
	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/lib/pktcls"
	"github.com/scionproto/scion/go/sig/anaconfig"
	"github.com/scionproto/scion/go/sig/mgmt"
	"github.com/scionproto/scion/go/sigmgmt/db"
	"github.com/scionproto/scion/go/sigmgmt/parser"
)

type Command struct {
	Name      string
	Condition pktcls.Cond
	Selectors []string
}

// Compile inserts the rules specified by policies into cfg. Any overwrites happen silently.
func Compile(cfg *config.Cfg, policies map[addr.IA][]db.Policy,
	trafficClasses map[uint]db.TrafficClass, selectors []db.PathSelector) error {
	if cfg.Classes == nil {
		cfg.Classes = make(pktcls.ClassMap)
	}
	actions, err := db.ActionMapFromSelectors(selectors)
	if err != nil {
		return err
	}
	cfg.Actions = actions
	// Create ID to name map
	selectorIDMap := map[string]string{}
	for _, selector := range selectors {
		selectorIDMap[strconv.Itoa(int(selector.ID))] = selector.Name
	}

	for ia, asEntry := range cfg.ASes {
		var commands []*Command
		for _, policy := range policies[ia] {
			if _, ok := trafficClasses[policy.TrafficClass]; !ok {
				return common.NewBasicError("Referenced TrafficClass cannot be found", nil,
					"class", policy.TrafficClass)
			}
			cond, err := parser.BuildClassTree(trafficClasses[policy.TrafficClass].CondStr)
			if err != nil {
				return err
			}
			selectorIDs := []string{}
			for _, selectorID := range strings.Split(policy.Selectors, ",") {
				selectorIDs = append(selectorIDs, selectorIDMap[selectorID])
			}
			commands = append(commands, &Command{Name: trafficClasses[policy.TrafficClass].Name,
				Condition: cond, Selectors: selectorIDs})
		}

		// Determine which path selectors are used by the current remote AS
		sessionMap := newSessionMapper(cfg.Actions)
		for _, command := range commands {
			if err := sessionMap.Add(command.Selectors); err != nil {
				return err
			}
		}

		for _, command := range commands {
			name := command.Name
			cfg.Classes[name] = pktcls.NewClass(name, command.Condition)
			// Add sessionIDs based on path selectors
			var sessIds []mgmt.SessionType
			for _, selector := range command.Selectors {
				sessIds = append(sessIds, sessionMap.sessions[selector])
			}
			asEntry.PktPolicies = append(
				asEntry.PktPolicies,
				&config.PktPolicy{
					ClassName: name,
					SessIds:   sessIds,
				})
			// Add the necessary sessions themselves
			sessions := make(config.SessionMap)
			for selector, id := range sessionMap.sessions {
				sessions[id] = selector
			}
			asEntry.Sessions = sessions
		}
	}
	return nil
}

type sessionMapper struct {
	nextSessionID  mgmt.SessionType
	sessions       map[string]mgmt.SessionType
	knownSelectors pktcls.ActionMap
}

func newSessionMapper(knownSelectors pktcls.ActionMap) *sessionMapper {
	return &sessionMapper{
		sessions:       make(map[string]mgmt.SessionType),
		knownSelectors: knownSelectors,
	}
}

func (s *sessionMapper) Add(references []string) error {
	for _, selector := range references {
		if _, ok := s.knownSelectors[selector]; !ok {
			return common.NewBasicError("Path selector does not exist:", nil, "selector", selector)
		}
		if _, ok := s.sessions[selector]; !ok {
			s.sessions[selector] = s.nextSessionID
			s.nextSessionID++
			if s.nextSessionID == 0 {
				// mgmt.SessionType overflowed
				return common.NewBasicError("Too many sessions", nil)
			}
		}
	}
	return nil
}
