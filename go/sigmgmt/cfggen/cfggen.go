// Copyright 2018 Anapaya Systems

// Package cfggen can be used to populate a SIG configuration with classes and
// packet policies derived from text policy rules.
//
// One policy rule is supported:
//   name=NAME [ src=NETWORK ]  [ dst=NETWORK ] [ dscp=DSCP ] sessions=SESSIONS
//
// See the static markdown documentation for more information about the fields.
package cfggen

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/scionproto/scion/go/lib/addr"
	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/lib/pktcls"
	"github.com/scionproto/scion/go/sig/config"
	"github.com/scionproto/scion/go/sig/mgmt"
	"github.com/scionproto/scion/go/sigmgmt/db"
	"github.com/scionproto/scion/go/sigmgmt/util"
)

// Compile inserts the rule specified by command into cfg. Any overwrites
// happen silently. classID is used to derive a class name if one isn't found
// in parameter command.
func Compile(cfg *config.Cfg, policies map[addr.IA]string,
	aliases map[addr.IA]db.SessionAliasMap) error {

	if cfg.Classes == nil {
		cfg.Classes = make(pktcls.ClassMap)
	}
	for ia, asEntry := range cfg.ASes {
		policy := policies[ia]
		for i, command := range strings.Split(policy, "\n") {
			err := compileCommand(cfg, asEntry, aliases[ia], command,
				fmt.Sprintf("class-%v-%d", ia, i))
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func compileCommand(cfg *config.Cfg, asEntry *config.ASEntry, aliasMap db.SessionAliasMap,
	command string, className string) error {

	var (
		conditions    []pktcls.Cond
		sessionTokens []string
	)
	for _, kvToken := range strings.Fields(command) {
		key, value, err := kvSplit(kvToken)
		if err != nil {
			return err
		}
		switch key {
		case "name":
			className = value
			if err := util.ValidateIdentifier(className); err != nil {
				return err
			}
		case "src":
			msrc := &pktcls.IPv4MatchSource{}
			_, msrc.Net, err = net.ParseCIDR(value)
			if err != nil {
				return err
			}
			conditions = append(conditions, pktcls.NewCondIPv4(msrc))
		case "dst":
			mdst := &pktcls.IPv4MatchDestination{}
			_, mdst.Net, err = net.ParseCIDR(value)
			if err != nil {
				return err
			}
			conditions = append(conditions, pktcls.NewCondIPv4(mdst))
		case "tos":
			mtos := &pktcls.IPv4MatchToS{}
			tos, err := strconv.ParseUint(value, 0, 8)
			if err != nil {
				return err
			}
			mtos.TOS = uint8(tos)
			conditions = append(conditions, pktcls.NewCondIPv4(mtos))
		case "dscp":
			mdscp := &pktcls.IPv4MatchDSCP{}
			dscp, err := strconv.ParseUint(value, 0, 8)
			if err != nil {
				return err
			}
			mdscp.DSCP = uint8(dscp)
			conditions = append(conditions, pktcls.NewCondIPv4(mdscp))
		case "sessions":
			sessionTokens = strings.Split(value, ",")
		default:
			return common.NewBasicError("unknown key", nil, "key", key)
		}
	}
	cfg.Classes[className] = pktcls.NewClass(className, pktcls.NewCondAllOf(conditions...))

	var sessionIDs []mgmt.SessionType
	// Resolve session aliases to session IDs, and check that we're not using
	// undefined session IDs.
	for _, token := range sessionTokens {
		var err error
		if tokenIsAliasName(token) {
			if sessionIDs, err = appendByAlias(sessionIDs, asEntry, aliasMap[token]); err != nil {
				return common.NewBasicError("alias error", err, "name", token)
			}
		} else {
			// Token is raw session ID
			if sessionIDs, err = appendByRaw(sessionIDs, asEntry, token); err != nil {
				return common.NewBasicError("session id error", err)
			}
		}
	}
	policy := &config.PktPolicy{
		ClassName: className,
		SessIds:   sessionIDs,
	}
	asEntry.PktPolicies = append(asEntry.PktPolicies, policy)
	return nil
}

func tokenIsAliasName(token string) bool {
	r, _ := utf8.DecodeRuneInString(token)
	return unicode.IsLetter(r)
}

// appendByAlias appends to sessionIDs the session IDs contained in string
// sessions.
func appendByAlias(sessionIDs []mgmt.SessionType, asEntry *config.ASEntry,
	sessions string) ([]mgmt.SessionType, error) {

	if sessions == "" {
		return nil, common.NewBasicError("alias not found or empty", nil)
	}
	for _, token := range strings.Split(sessions, ",") {
		var err error
		sessionIDs, err = appendByRaw(sessionIDs, asEntry, token)
		if err != nil {
			return nil, err
		}
	}
	return sessionIDs, nil
}

// appendByRaw appends to sessionIDs the session ID contained in string
// session. appendByRaw also checks that the session ID is defined for the
// specified asEntry.
func appendByRaw(sessionIDs []mgmt.SessionType, asEntry *config.ASEntry,
	session string) ([]mgmt.SessionType, error) {

	id, err := parseSessionID(session)
	if err != nil {
		return nil, err
	}
	if _, ok := asEntry.Sessions[id]; !ok {
		return nil, common.NewBasicError("session does not exist", nil, "id", id)
	}
	return append(sessionIDs, id), nil
}

func parseSessionID(input string) (mgmt.SessionType, error) {
	i, err := strconv.ParseUint(input, 0, 8)
	return mgmt.SessionType(i), err
}

// kvSplit splits a "key=value" formatted token into its components.
func kvSplit(token string) (key, value string, err error) {
	items := strings.Split(token, "=")
	if len(items) != 2 {
		return "", "", common.NewBasicError("bad token", nil, "token", token)
	}
	return items[0], items[1], nil
}
