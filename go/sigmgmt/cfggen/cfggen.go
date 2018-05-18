// Copyright 2018 Anapaya Systems

// Package cfggen can be used to populate a SIG configuration with classes and
// packet policies derived from text policy rules.
//
// One policy rule is supported:
//   name=NAME [ src=NETWORK ]  [ dst=NETWORK ] [ dscp=DSCP ] selectors=SELECTORS
//
// See the static markdown documentation for more information about the fields.
package cfggen

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/scionproto/scion/go/lib/addr"
	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/lib/pktcls"
	"github.com/scionproto/scion/go/sig/config"
	"github.com/scionproto/scion/go/sig/mgmt"
	"github.com/scionproto/scion/go/sigmgmt/util"
)

// Compile inserts the rules specified by policies into cfg. Any overwrites happen silently.
func Compile(cfg *config.Cfg, policies map[addr.IA]string) error {
	if cfg.Classes == nil {
		cfg.Classes = make(pktcls.ClassMap)
	}

	for ia, asEntry := range cfg.ASes {
		var commands []*Command
		for _, commandStr := range strings.Split(policies[ia], "\n") {
			command, err := NewCommand(commandStr)
			if err != nil {
				return err
			}
			commands = append(commands, command)
		}

		// Determine which path selectors are used by the current remote AS
		sessionMap := newSessionMapper(cfg.Actions)
		for _, command := range commands {
			if err := sessionMap.Add(command.Selectors); err != nil {
				return err
			}
		}

		for i, command := range commands {
			var className = fmt.Sprintf("class-%v-%d", ia, i)
			if command.Name != nil {
				className = *command.Name
			}

			cfg.Classes[className] = pktcls.NewClass(
				className,
				pktcls.NewCondAllOf(command.Conditions...))
			// Add sessionIDs based on path selectors
			var sessIds []mgmt.SessionType
			for _, selector := range command.Selectors {
				sessIds = append(sessIds, sessionMap.sessions[selector])
			}
			asEntry.PktPolicies = append(
				asEntry.PktPolicies,
				&config.PktPolicy{
					ClassName: className,
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

type Command struct {
	Name       *string
	Conditions []pktcls.Cond
	Selectors  []string
}

func NewCommand(command string) (*Command, error) {
	cmd := &Command{}
	for _, kvToken := range strings.Fields(command) {
		key, value, err := kvSplit(kvToken)
		if err != nil {
			return nil, err
		}
		switch key {
		case "name":
			if err := util.ValidateIdentifier(value); err != nil {
				return nil, err
			}
			otherValue := value
			cmd.Name = &otherValue
		case "src":
			msrc := &pktcls.IPv4MatchSource{}
			_, msrc.Net, err = net.ParseCIDR(value)
			if err != nil {
				return nil, err
			}
			cmd.Conditions = append(cmd.Conditions, pktcls.NewCondIPv4(msrc))
		case "dst":
			mdst := &pktcls.IPv4MatchDestination{}
			_, mdst.Net, err = net.ParseCIDR(value)
			if err != nil {
				return nil, err
			}
			cmd.Conditions = append(cmd.Conditions, pktcls.NewCondIPv4(mdst))
		case "tos":
			mtos := &pktcls.IPv4MatchToS{}
			tos, err := strconv.ParseUint(value, 0, 8)
			if err != nil {
				return nil, err
			}
			mtos.TOS = uint8(tos)
			cmd.Conditions = append(cmd.Conditions, pktcls.NewCondIPv4(mtos))
		case "dscp":
			mdscp := &pktcls.IPv4MatchDSCP{}
			dscp, err := strconv.ParseUint(value, 0, 8)
			if err != nil {
				return nil, err
			}
			mdscp.DSCP = uint8(dscp)
			cmd.Conditions = append(cmd.Conditions, pktcls.NewCondIPv4(mdscp))
		case "selectors":
			cmd.Selectors = strings.Split(value, ",")
		default:
			return nil, common.NewBasicError("unknown key", nil, "key", key)
		}
	}
	return cmd, nil
}

// kvSplit splits a "key=value" formatted token into its components.
func kvSplit(token string) (key, value string, err error) {
	items := strings.Split(token, "=")
	if len(items) != 2 {
		return "", "", common.NewBasicError("Bad token: ", nil, "token", token)
	}
	return items[0], items[1], nil
}
