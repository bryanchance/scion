// Copyright 2018 Anapaya Systems

package cfggen

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/scionproto/scion/go/lib/addr"
	"github.com/scionproto/scion/go/lib/pathpol"
	"github.com/scionproto/scion/go/lib/xtest"
	"github.com/scionproto/scion/go/sig/anaconfig"
	"github.com/scionproto/scion/go/sigmgmt/db"
)

var (
	update = flag.Bool("update", false, "set to true to update reference testdata files")
)

func TestCompile(t *testing.T) {
	testCases := []struct {
		Name         string
		File         string
		Config       *config.Cfg // This mutates during the test
		Policies     map[addr.IA][]db.TrafficPolicy
		Classes      map[uint]db.TrafficClass
		PathPolicies []*pathpol.ExtPolicy
	}{
		{
			Name: "one path policy, one AS",
			File: "1",
			Config: &config.Cfg{
				ASes: map[addr.IA]*config.ASEntry{
					{I: 1, A: 1}: {},
				},
			},
			Policies: map[addr.IA][]db.TrafficPolicy{
				{I: 1, A: 1}: {
					{Name: "audio", PathPolicies: "foo", TrafficClass: 0},
				},
			},
			Classes: map[uint]db.TrafficClass{
				0: {ID: 0, Name: "class-1", CondStr: "src=1.1.1.0/24"},
			},
			PathPolicies: []*pathpol.ExtPolicy{
				{
					Policy: &pathpol.Policy{
						Name: "foo",
						Sequence: newSequence(t,
							"1-ff00:0:133#1010 1-ff00:0:132#1910")}},
			},
		},
		{
			Name: "two path policies, AS 1 uses both, AS 2 uses one",
			File: "2",
			Config: &config.Cfg{
				ASes: map[addr.IA]*config.ASEntry{
					{I: 1, A: 1}: {},
					{I: 2, A: 2}: {},
				},
			},
			Policies: map[addr.IA][]db.TrafficPolicy{
				{I: 1, A: 1}: {
					{Name: "audio", PathPolicies: "any, foo", TrafficClass: 0},
				},
				{I: 2, A: 2}: {
					{Name: "video", PathPolicies: "foo", TrafficClass: 1},
				},
			},
			Classes: map[uint]db.TrafficClass{
				0: {ID: 0, Name: "class-1", CondStr: "src=1.1.1.0/24"},
				1: {ID: 1, Name: "class-2", CondStr: "dst=2.2.2.0/24"},
			},
			PathPolicies: []*pathpol.ExtPolicy{
				{
					Policy: &pathpol.Policy{Name: "any"}},
				{
					Policy: &pathpol.Policy{Name: "foo",
						Sequence: newSequence(t, "1-ff00:0:133#1010")}},
				{
					Policy: &pathpol.Policy{Name: "bar",
						Sequence: newSequence(t, "1-ff00:0:132#1910")}},
			},
		},
		{
			Name: "multiple classes in same AS",
			File: "3",
			Config: &config.Cfg{
				ASes: map[addr.IA]*config.ASEntry{
					{I: 1, A: 1}: {},
					{I: 2, A: 2}: {},
				},
			},
			Policies: map[addr.IA][]db.TrafficPolicy{
				{I: 1, A: 1}: {
					{Name: "audio", PathPolicies: "any, foo", TrafficClass: 0},
					{Name: "video", PathPolicies: "foo, bar", TrafficClass: 1},
				},
			},
			Classes: map[uint]db.TrafficClass{
				0: {Name: "class-1", CondStr: "dscp=0x12"},
				1: {Name: "class-2", CondStr: "tos=0x34"},
			},
			PathPolicies: []*pathpol.ExtPolicy{
				{
					Policy: &pathpol.Policy{Name: "any"}},
				{
					Policy: &pathpol.Policy{Name: "foo",
						Sequence: newSequence(t, "1-ff00:0:133#1010")}},
				{
					Policy: &pathpol.Policy{Name: "bar",
						Sequence: newSequence(t, "1-ff00:0:132#1910")}},
				{
					Policy: &pathpol.Policy{Name: "baz",
						Sequence: newSequence(t, "1-ff00:0:132#0")}},
			},
		},
		{
			Name: "complex policies",
			File: "4",
			Config: &config.Cfg{
				ASes: map[addr.IA]*config.ASEntry{
					{I: 1, A: 1}: {},
					{I: 2, A: 2}: {},
				},
			},
			Policies: map[addr.IA][]db.TrafficPolicy{
				{I: 1, A: 1}: {
					{Name: "audio", PathPolicies: "any, foo", TrafficClass: 0},
					{Name: "video", PathPolicies: "foo, bar", TrafficClass: 1},
					{Name: "http", PathPolicies: "baz, foo", TrafficClass: 2},
				},
				{I: 2, A: 2}: {
					{Name: "audio", PathPolicies: "bax, bar", TrafficClass: 2},
					{Name: "video", PathPolicies: "foo, any", TrafficClass: 4},
					{Name: "http", PathPolicies: "any, bar", TrafficClass: 3},
				},
			},
			Classes: map[uint]db.TrafficClass{
				0: {Name: "class-1", CondStr: "src=12.12.127.0/24"},
				1: {Name: "class-2", CondStr: "NOT(dst=124.124.153.0/24)"},
				2: {Name: "class-3", CondStr: "ALL(src=12.1.1.0/24, NOT(dst=124.124.153.0/24))"},
				3: {Name: "class-4", CondStr: "ANY(tos=0x34, src=1.1.1.0/28)"},
				4: {Name: "class-5",
					CondStr: "ALL(NOT(src=12.12.127.0/24),ANY(tos=0x34, src=1.1.1.0/28))"},
			},
			PathPolicies: []*pathpol.ExtPolicy{
				{
					Policy: &pathpol.Policy{Name: "any"}},
				{
					Policy: &pathpol.Policy{Name: "foo",
						Sequence: newSequence(t, "1-ff00:0:133#1010")}},
				{
					Policy: &pathpol.Policy{Name: "bar",
						Sequence: newSequence(t, "1-ff00:0:132#1910")}},
				{
					Policy: &pathpol.Policy{Name: "baz",
						Sequence: newSequence(t, "1-ff00:0:132#0")}},
				{
					Policy: &pathpol.Policy{Name: "bax",
						Sequence: newSequence(t, "1-0#0")}},
			},
		},
		{
			Name: "empty AS config",
			File: "5",
			Config: &config.Cfg{
				ASes: map[addr.IA]*config.ASEntry{
					{I: 1, A: 1}: {},
				},
			},
		},
		{
			Name: "extended policies",
			File: "6",
			Config: &config.Cfg{
				ASes: map[addr.IA]*config.ASEntry{
					{I: 1, A: 1}: {},
				},
			},
			Policies: map[addr.IA][]db.TrafficPolicy{
				{I: 1, A: 1}: {
					{Name: "audio", PathPolicies: "any", TrafficClass: 0},
					{Name: "audio", PathPolicies: "bar2", TrafficClass: 0},
				},
			},
			Classes: map[uint]db.TrafficClass{
				0: {Name: "class-1", CondStr: "src=12.12.127.0/24"},
			},
			PathPolicies: []*pathpol.ExtPolicy{
				{
					Policy:  &pathpol.Policy{Name: "any"},
					Extends: []string{"bar", "foo"},
				},
				{
					Policy: &pathpol.Policy{Name: "foo",
						Sequence: newSequence(t, "1-ff00:0:133#1010")}},
				{
					Policy: &pathpol.Policy{Name: "bar",
						ACL: mustACL(t,
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
						Sequence: newSequence(t, "1-ff00:0:132#1910")}},
				{
					Policy:  &pathpol.Policy{Name: "bar2"},
					Extends: []string{"foo2"},
				},
				{
					Policy:  &pathpol.Policy{Name: "foo2"},
					Extends: []string{"bar"},
				},
			},
		},
	}

	Convey("TestCompile", t, func() {
		for _, tc := range testCases {
			Convey(tc.Name, func() {
				err := Compile(tc.Config, tc.Policies, tc.Classes, tc.PathPolicies)
				SoMsg("err", err, ShouldBeNil)

				if *update {
					xtest.MustMarshalJSONToFile(t, tc.Config, tc.File+".json")
				}

				expStr, err := ioutil.ReadFile(xtest.ExpandPath(tc.File + ".json"))
				xtest.FailOnErr(t, err)
				// Remove EOF
				expStr = expStr[:len(expStr)-1]

				str, err := json.MarshalIndent(tc.Config, "", "    ")
				SoMsg("json", string(str), ShouldEqual, string(expStr))
			})
		}
	})

	testCases = []struct {
		Name         string
		File         string
		Config       *config.Cfg // This mutates during the test
		Policies     map[addr.IA][]db.TrafficPolicy
		Classes      map[uint]db.TrafficClass
		PathPolicies []*pathpol.ExtPolicy
	}{
		{
			Name: "missing traffic class",
			Config: &config.Cfg{
				ASes: map[addr.IA]*config.ASEntry{
					{I: 1, A: 1}: {},
				},
			},
			Policies: map[addr.IA][]db.TrafficPolicy{
				{I: 1, A: 1}: {
					{Name: "audio", PathPolicies: "foo", TrafficClass: 0},
				},
			},
		},
		{
			Name: "bad traffic class",
			Config: &config.Cfg{
				ASes: map[addr.IA]*config.ASEntry{
					{I: 1, A: 1}: {},
				},
			},
			Classes: map[uint]db.TrafficClass{
				0: {ID: 0, Name: "class-1", CondStr: "cls=1"},
			},
			Policies: map[addr.IA][]db.TrafficPolicy{
				{I: 1, A: 1}: {
					{Name: "audio", PathPolicies: "foo", TrafficClass: 0},
				},
			},
		},
		{
			Name: "missing path policy",
			Config: &config.Cfg{
				ASes: map[addr.IA]*config.ASEntry{
					{I: 1, A: 1}: {},
				},
			},
			Classes: map[uint]db.TrafficClass{
				0: {ID: 0, Name: "class-1", CondStr: "src=1.1.1.0/24"},
			},
			Policies: map[addr.IA][]db.TrafficPolicy{
				{I: 1, A: 1}: {
					{Name: "audio", PathPolicies: "foo"},
				},
			},
		},
	}

	Convey("TestCompile fail", t, func() {
		for _, tc := range testCases {
			Convey(tc.Name, func() {
				err := Compile(tc.Config, tc.Policies, tc.Classes, tc.PathPolicies)
				SoMsg("err", err, ShouldNotBeNil)
			})
		}
	})
}

func newSequence(t *testing.T, str string) *pathpol.Sequence {
	seq, err := pathpol.NewSequence(str)
	xtest.FailOnErr(t, err)
	return seq
}

func mustACL(t *testing.T, entries ...*pathpol.ACLEntry) *pathpol.ACL {
	acl, err := pathpol.NewACL(entries...)
	xtest.FailOnErr(t, err)
	return acl
}

func mustHopPredicate(t *testing.T, str string) *pathpol.HopPredicate {
	hp, err := pathpol.HopPredicateFromString(str)
	xtest.FailOnErr(t, err)
	return hp
}
