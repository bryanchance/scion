// Copyright 2018 Anapaya Systems

// Package cfggen can be used to populate a SIG configuration with classes and
// packet policies derived from text policy rules.
//
// See the static markdown documentation for more information about the fields.
package cfggen

import (
	"github.com/scionproto/scion/go/lib/addr"
	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/lib/pathpol"
	"github.com/scionproto/scion/go/lib/pktcls"
	"github.com/scionproto/scion/go/sig/anaconfig"
	"github.com/scionproto/scion/go/sig/mgmt"
	"github.com/scionproto/scion/go/sigmgmt/db"
	"github.com/scionproto/scion/go/sigmgmt/parser"
)

type Command struct {
	Name      string
	Condition pktcls.Cond
	Policies  []string
}

// Compile inserts the rules specified by policies into cfg. Any overwrites happen silently.
func Compile(cfg *config.Cfg, policies map[addr.IA][]db.TrafficPolicy,
	trafficClasses map[uint]db.TrafficClass, pathPolicies []*pathpol.ExtPolicy) error {
	if cfg.PktClasses == nil {
		cfg.PktClasses = make(pktcls.ClassMap)
	}
	cfg.PathPolicies = make(pathpol.PolicyMap)
	polMap := make(pathpol.PolicyMap)
	for _, extPol := range pathPolicies {
		polMap[extPol.Policy.Name] = extPol
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
			commands = append(commands, &Command{Name: trafficClasses[policy.TrafficClass].Name,
				Condition: cond, Policies: policy.PathPolicies})
		}

		// Make sure there is a default policy
		if len(commands) == 0 {
			cmd, err := defaultPolicyCommand()
			if err != nil {
				return err
			}
			commands = []*Command{cmd}
		}

		// Determine which path policies are used by the current remote AS
		sessionMap := newSessionMapper(polMap)
		for _, command := range commands {
			if err := sessionMap.Add(command.Policies); err != nil {
				return err
			}
		}

		for _, command := range commands {
			name := command.Name
			cfg.PktClasses[name] = pktcls.NewClass(name, command.Condition)
			// Add sessionIDs based on path policies
			var sessIds []mgmt.SessionType
			for _, policy := range command.Policies {
				sessIds = append(sessIds, sessionMap.sessions[policy])
			}
			asEntry.PktPolicies = append(
				asEntry.PktPolicies,
				&config.PktPolicy{
					ClassName: name,
					SessIds:   sessIds,
				})
			// Add the necessary sessions themselves
			sessions := make(config.SessionMap)
			for policy, id := range sessionMap.sessions {
				sessions[id] = policy
			}
			// Add the necessary policies
			for _, name := range command.Policies {
				cfg.PathPolicies[name] = polMap[name]
			}
			asEntry.Sessions = sessions
		}
	}
	return nil
}

func defaultPolicyCommand() (*Command, error) {
	cond, err := parser.BuildClassTree("bool=true")
	if err != nil {
		return nil, err
	}
	return &Command{Name: "default-generated", Condition: cond}, nil
}

type sessionMapper struct {
	nextSessionID mgmt.SessionType
	sessions      map[string]mgmt.SessionType
	knownPolicies pathpol.PolicyMap
}

func newSessionMapper(knownPolicies pathpol.PolicyMap) *sessionMapper {
	return &sessionMapper{
		sessions:      make(map[string]mgmt.SessionType),
		knownPolicies: knownPolicies,
	}
}

func (s *sessionMapper) Add(references []string) error {
	for _, policy := range references {
		if _, ok := s.knownPolicies[policy]; !ok {
			return common.NewBasicError("Path policy does not exist:", nil, "policy", policy)
		}
		if _, ok := s.sessions[policy]; !ok {
			s.sessions[policy] = s.nextSessionID
			s.nextSessionID++
			if s.nextSessionID == 0 {
				// mgmt.SessionType overflowed
				return common.NewBasicError("Too many sessions", nil)
			}
		}
	}
	return nil
}
