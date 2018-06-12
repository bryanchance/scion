// Copyright 2018 Anapaya Systems

package cfggen

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/scionproto/scion/go/lib/addr"
	"github.com/scionproto/scion/go/lib/xtest"
	"github.com/scionproto/scion/go/sig/anaconfig"
	"github.com/scionproto/scion/go/sigmgmt/db"
)

var (
	update = flag.Bool("update", false, "set to true to update reference testdata files")
)

func TestCompile(t *testing.T) {
	testCases := []struct {
		Name      string
		File      string
		Config    *config.Cfg // This mutates during the test
		Policies  map[addr.IA][]db.Policy
		Classes   map[uint]db.TrafficClass
		Selectors []db.PathSelector
	}{
		{
			Name: "one selector, one AS",
			File: "1",
			Config: &config.Cfg{
				ASes: map[addr.IA]*config.ASEntry{
					{I: 1, A: 1}: {},
				},
			},
			Policies: map[addr.IA][]db.Policy{
				{I: 1, A: 1}: {
					{Name: "audio", Selectors: "0", TrafficClass: 0},
				},
			},
			Classes: map[uint]db.TrafficClass{
				0: {ID: 0, Name: "class-1", CondStr: "src=1.1.1.0/24"},
			},
			Selectors: []db.PathSelector{
				{ID: 0, Name: "foo", Filter: "0-0#0"},
			},
		},
		{
			Name: "two selectors, AS 1 uses both, AS 2 uses one",
			File: "2",
			Config: &config.Cfg{
				ASes: map[addr.IA]*config.ASEntry{
					{I: 1, A: 1}: {},
					{I: 2, A: 2}: {},
				},
			},
			Policies: map[addr.IA][]db.Policy{
				{I: 1, A: 1}: {
					{Name: "audio", Selectors: "0,1", TrafficClass: 0},
				},
				{I: 2, A: 2}: {
					{Name: "video", Selectors: "1", TrafficClass: 1},
				},
			},
			Classes: map[uint]db.TrafficClass{
				0: {ID: 0, Name: "class-1", CondStr: "src=1.1.1.0/24"},
				1: {ID: 1, Name: "class-2", CondStr: "dst=2.2.2.0/24"},
			},
			Selectors: []db.PathSelector{
				{ID: 0, Name: "foo", Filter: "0-0#0"},
				{ID: 1, Name: "bar", Filter: "0-0#0"},
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
			Policies: map[addr.IA][]db.Policy{
				{I: 1, A: 1}: {
					{Name: "audio", Selectors: "0,1", TrafficClass: 0},
					{Name: "video", Selectors: "1,2", TrafficClass: 1},
				},
			},
			Classes: map[uint]db.TrafficClass{
				0: {Name: "class-1", CondStr: "dscp=0x12"},
				1: {Name: "class-2", CondStr: "tos=0x34"},
			},
			Selectors: []db.PathSelector{
				{ID: 0, Name: "foo", Filter: "1-1#1"},
				{ID: 1, Name: "bar", Filter: "2-2#2"},
				{ID: 2, Name: "baz", Filter: "3-3#3"},
				{ID: 3, Name: "bad", Filter: "4-4#4"},
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
			Policies: map[addr.IA][]db.Policy{
				{I: 1, A: 1}: {
					{Name: "audio", Selectors: "0,1", TrafficClass: 0},
					{Name: "video", Selectors: "1,2", TrafficClass: 1},
					{Name: "http", Selectors: "3,1", TrafficClass: 2},
				},
				{I: 2, A: 2}: {
					{Name: "audio", Selectors: "4,2", TrafficClass: 2},
					{Name: "video", Selectors: "1,0", TrafficClass: 4},
					{Name: "http", Selectors: "0,2", TrafficClass: 3},
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
			Selectors: []db.PathSelector{
				{ID: 0, Name: "foo", Filter: "1-1#1"},
				{ID: 1, Name: "bar", Filter: "2-2#2"},
				{ID: 2, Name: "baz", Filter: "3-3#3"},
				{ID: 3, Name: "bad", Filter: "4-4#4"},
				{ID: 4, Name: "bax", Filter: "5-4#4"},
			},
		},
	}

	Convey("TestCompile", t, func() {
		for _, tc := range testCases {
			Convey(tc.Name, func() {
				err := Compile(tc.Config, tc.Policies, tc.Classes, tc.Selectors)
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
}
