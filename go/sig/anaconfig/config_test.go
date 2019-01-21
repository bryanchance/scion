// Copyright 2018 Anapaya Systems

package config

import (
	"flag"
	"net"
	"path/filepath"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/scionproto/scion/go/lib/addr"
	"github.com/scionproto/scion/go/lib/pathpol"
	"github.com/scionproto/scion/go/lib/pktcls"
	"github.com/scionproto/scion/go/lib/xtest"
	"github.com/scionproto/scion/go/sig/mgmt"
)

var (
	update = flag.Bool("update", false, "set to true to update reference testdata files")
)

func TestCompatible(t *testing.T) {
	// Sanity check to test that the open source reference JSON config files
	// are still compatible with the Anapaya unmarshaler.

	// This MUST load from the open source reference files, because those might
	// get updated by a programmer that is unaware of the closed source
	// version. This MUST NEVER update the open source reference files (because
	// some fields do not make sense in that version).
	testCases := []struct {
		Name     string
		FileName string
		Config   Cfg
	}{
		{
			Name:     "simple",
			FileName: "../config/testdata/01-loadfromfile.json",
			// Config mirrors open source test in config/config_test.go, but
			// some types (e.g., ASEntry) are different.
			Config: Cfg{
				ASes: map[addr.IA]*ASEntry{
					xtest.MustParseIA("1-ff00:0:1"): {
						Nets: []*IPNet{
							{
								IP:   net.IP{192, 0, 2, 0},
								Mask: net.CIDRMask(24, 8*net.IPv4len),
							},
							{
								IP:   net.ParseIP("2001:DB8::"),
								Mask: net.CIDRMask(48, 8*net.IPv6len),
							},
						},
					},
					xtest.MustParseIA("1-ff00:0:2"): {
						Nets: []*IPNet{
							{
								IP:   net.IP{203, 0, 113, 0},
								Mask: net.CIDRMask(24, 8*net.IPv4len),
							},
						},
					},
					xtest.MustParseIA("1-ff00:0:3"): {
						Nets: []*IPNet{},
					},
					xtest.MustParseIA("1-ff00:0:4"): {
						Nets: []*IPNet{},
					},
				},
				ConfigVersion: 9001,
			},
		},
	}

	Convey("Test open source/Anapaya config file compatibility", t, func() {
		for _, tc := range testCases {
			cfg, err := LoadFromFile(tc.FileName)
			SoMsg("err", err, ShouldBeNil)
			SoMsg("cfg", *cfg, ShouldResemble, tc.Config)
		}
	})

}

func TestLoadFromFile(t *testing.T) {
	testCases := []struct {
		Name     string
		FileName string
		Config   Cfg
	}{
		{
			Name:     "unmarshal basic config created from anapaya marshal",
			FileName: "01-loadfromfile.json",
			Config: Cfg{
				ASes: map[addr.IA]*ASEntry{
					xtest.MustParseIA("1-ff00:0:1"): {
						Nets: []*IPNet{
							{
								IP:   net.IP{192, 168, 1, 0},
								Mask: net.CIDRMask(24, 8*net.IPv4len),
							},
							{
								IP:   net.ParseIP("2001:DB8::"),
								Mask: net.CIDRMask(48, 8*net.IPv6len),
							},
						},
					},
					xtest.MustParseIA("1-ff00:0:2"): {
						Nets: []*IPNet{
							{
								IP:   net.IP{2, 0, 0, 0},
								Mask: net.CIDRMask(16, 8*net.IPv4len),
							},
						},
					},
					xtest.MustParseIA("1-ff00:0:3"): {
						Nets: []*IPNet{},
					},
					xtest.MustParseIA("1-ff00:0:4"): {
						Nets: []*IPNet{},
					},
				},
				ConfigVersion: 0,
			},
		},
		{
			Name:     "unmarshal config with classifiers",
			FileName: "02-loadfromfile.json",
			Config: Cfg{
				ASes: map[addr.IA]*ASEntry{
					xtest.MustParseIA("1-ff00:0:1"): {
						Nets: []*IPNet{
							{
								IP:   net.IP{192, 168, 1, 0},
								Mask: net.CIDRMask(24, 8*net.IPv4len),
							},
							{
								IP:   net.ParseIP("2001:DB8::"),
								Mask: net.CIDRMask(48, 8*net.IPv6len),
							},
						},
						Sessions: SessionMap{
							0x03: "policy-A",
							0x04: "policy-B",
						},
						PktPolicies: []*PktPolicy{
							{
								Name:      "policy-transit-1",
								ClassName: "transit-isd-1",
								SessIds:   []mgmt.SessionType{0x03},
							},
							{
								ClassName: "transit-isd-2",
								SessIds:   []mgmt.SessionType{0x03, 0x04},
							},
						},
					},
					xtest.MustParseIA("1-ff00:0:2"): {
						Nets: []*IPNet{
							{
								IP:   net.IP{2, 0, 0, 0},
								Mask: net.CIDRMask(16, 8*net.IPv4len),
							},
						},
					},
					xtest.MustParseIA("1-ff00:0:3"): {
						Nets: []*IPNet{},
					},
					xtest.MustParseIA("1-ff00:0:4"): {
						Nets: []*IPNet{},
						Sessions: SessionMap{
							0x00: "",
						},
						PktPolicies: []*PktPolicy{
							{
								ClassName: "transit-isd-1",
								SessIds:   []mgmt.SessionType{0x00},
							},
						},
					},
				},
				PktClasses: pktcls.ClassMap{
					"transit-isd-1": pktcls.NewClass(
						"transit-isd-1",
						pktcls.NewCondAllOf(
							pktcls.NewCondIPv4(&pktcls.IPv4MatchToS{TOS: 0x80}),
							pktcls.NewCondIPv4(&pktcls.IPv4MatchDestination{
								Net: &net.IPNet{
									IP:   net.IP{192, 168, 1, 0},
									Mask: net.IPv4Mask(255, 255, 255, 0),
								},
							}),
						),
					),
					"transit-isd-2": pktcls.NewClass(
						"transit-isd-2",
						pktcls.NewCondAnyOf(
							pktcls.NewCondIPv4(&pktcls.IPv4MatchToS{TOS: 0x0}),
							pktcls.NewCondIPv4(&pktcls.IPv4MatchSource{
								Net: &net.IPNet{
									IP:   net.IP{10, 0, 0, 0},
									Mask: net.IPv4Mask(255, 0, 0, 0),
								},
							}),
						),
					),
					"classC": pktcls.NewClass(
						"classC",
						pktcls.NewCondAllOf(),
					),
				},
				PathPolicies: pathpol.PolicyMap{
					"policy-A": &pathpol.ExtPolicy{Policy: pathpol.NewPolicy("", nil,
						mustSequence(t, "0-123#0 0-234#0"),
						nil,
					)},
					"policy-B": &pathpol.ExtPolicy{Policy: pathpol.NewPolicy("", nil,
						mustSequence(t, "0-123#0 0-134#0"), nil,
					)},
					"policy-C": &pathpol.ExtPolicy{Policy: pathpol.NewPolicy("",
						mustACL(t,
							&pathpol.ACLEntry{
								Action: pathpol.Allow,
								Rule:   mustHopPredicate(t, "0-123"),
							},
							&pathpol.ACLEntry{
								Action: pathpol.Allow,
								Rule:   mustHopPredicate(t, "0-124"),
							},
							&pathpol.ACLEntry{
								Action: pathpol.Deny,
								Rule:   mustHopPredicate(t, "0"),
							},
						),
						mustSequence(t, "0-123#0 0-134#0"),
						[]pathpol.Option{
							{
								Weight: 1,
								Policy: pathpol.NewPolicy("", nil,
									mustSequence(t, "0-123#0 0-134#0"), nil,
								),
							},
							{
								Weight: 2,
								Policy: pathpol.NewPolicy("", nil,
									mustSequence(t, "0-134#0 0-145#0"), nil,
								),
							},
						},
					)},
				},
				ConfigVersion: 0,
			},
		},
	}

	Convey("Test SIG config marshal/unmarshal", t, func() {
		for _, tc := range testCases {
			Convey(tc.Name, func() {
				if *update {
					xtest.MustMarshalJSONToFile(t, tc.Config, tc.FileName)
				}
				cfg, err := LoadFromFile(filepath.Join("testdata", tc.FileName))
				SoMsg("err", err, ShouldBeNil)
				SoMsg("cfg", *cfg, ShouldResemble, tc.Config)
			})
		}
	})
}

func TestBuildSessions(t *testing.T) {
	Convey("Build sessions", t, func() {
		pol1 := &pathpol.Policy{Sequence: mustSequence(t, "0")}
		pol2 := &pathpol.Policy{}
		sessionMap := SessionMap{1: "policy-1", 2: "policy-2"}
		policies := pathpol.PolicyMap{"policy-1": {Policy: pol1}, "policy-2": {Policy: pol2}}
		sessionSet, err := BuildSessions(sessionMap, policies)
		SoMsg("err", err, ShouldBeNil)
		resSet := SessionSet{
			1: &Session{ID: 1, PolName: "policy-1", Policy: pol1},
			2: &Session{ID: 2, PolName: "policy-2", Policy: pol2},
		}
		SoMsg("sessionSet", sessionSet, ShouldResemble, resSet)
	})
	Convey("Build sessions fail", t, func() {
		pol1 := &pathpol.Policy{Sequence: mustSequence(t, "0")}
		pol2 := &pathpol.Policy{}
		sessionMap := SessionMap{1: "policy-10", 2: "policy-2"}
		policies := pathpol.PolicyMap{"policy-1": {Policy: pol1}, "policy-2": {Policy: pol2}}
		_, err := BuildSessions(sessionMap, policies)
		SoMsg("err", err, ShouldNotBeNil)
	})
	Convey("Build sessions, only 0 session", t, func() {
		sessionMap := SessionMap{0: ""}
		policies := pathpol.PolicyMap{}
		sessionSet, err := BuildSessions(sessionMap, policies)
		SoMsg("err", err, ShouldBeNil)
		resSet := SessionSet{
			0: &Session{ID: 0, PolName: "", Policy: nil},
		}
		SoMsg("sessionSet", sessionSet, ShouldResemble, resSet)
	})
	Convey("Build session from extended policy", t, func() {
		pol1 := &pathpol.Policy{Name: "policy-1"}
		pol2 := &pathpol.Policy{
			Name:     "policy-2",
			Sequence: mustSequence(t, "0 1 2")}
		exPol1 := &pathpol.ExtPolicy{Policy: pol1, Extends: []string{"policy-2"}}
		exPol2 := &pathpol.ExtPolicy{Policy: pol2}
		sessionMap := SessionMap{1: "policy-1"}
		policies := pathpol.PolicyMap{exPol1.Policy.Name: exPol1, exPol2.Policy.Name: exPol2}
		sessionSet, err := BuildSessions(sessionMap, policies)
		SoMsg("err", err, ShouldBeNil)
		resSet := SessionSet{
			1: &Session{ID: 1, PolName: exPol1.Policy.Name,
				Policy: &pathpol.Policy{
					Name:     exPol1.Policy.Name,
					Sequence: mustSequence(t, "0 1 2")}},
		}
		SoMsg("sessionSet", sessionSet, ShouldResemble, resSet)
	})
	Convey("Build sessions from ext Policy fails", t, func() {
		pol1 := &pathpol.Policy{Name: "policy-1"}
		exPol1 := &pathpol.ExtPolicy{Policy: pol1, Extends: []string{"policy-2"}}
		sessionMap := SessionMap{1: "policy-1"}
		policies := pathpol.PolicyMap{exPol1.Policy.Name: exPol1}
		_, err := BuildSessions(sessionMap, policies)
		SoMsg("err", err, ShouldNotBeNil)
	})
}

func mustACL(t *testing.T, entries ...*pathpol.ACLEntry) *pathpol.ACL {
	acl, err := pathpol.NewACL(entries...)
	xtest.FailOnErr(t, err)
	return acl
}

func mustSequence(t *testing.T, str string) *pathpol.Sequence {
	t.Helper()

	seq, err := pathpol.NewSequence(str)
	xtest.FailOnErr(t, err)
	return seq
}

func mustHopPredicate(t *testing.T, str string) *pathpol.HopPredicate {
	hp, err := pathpol.HopPredicateFromString(str)
	xtest.FailOnErr(t, err)
	return hp
}
