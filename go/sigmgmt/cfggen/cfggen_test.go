// Copyright 2018 Anapaya Systems

package cfggen

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/scionproto/scion/go/lib/addr"
	"github.com/scionproto/scion/go/lib/pktcls"
	"github.com/scionproto/scion/go/lib/spath/spathmeta"
	"github.com/scionproto/scion/go/lib/xtest"
	"github.com/scionproto/scion/go/sig/config"
)

var (
	update = flag.Bool("update", false, "set to true to update reference testdata files")
)

func TestCompile(t *testing.T) {
	testCases := []struct {
		Name     string
		File     string
		Config   *config.Cfg // This mutates during the test
		Policies map[addr.IA]string
	}{
		{
			Name: "one selector, one AS",
			File: "1",
			Config: &config.Cfg{
				ASes: map[addr.IA]*config.ASEntry{
					{I: 1, A: 1}: {},
				},
				Actions: pktcls.ActionMap{
					"foo": MustInitSimpleFilter(t, "foo", "0-0#0"),
				},
			},
			Policies: map[addr.IA]string{
				{I: 1, A: 1}: "name=audio src=1.1.1.0/24 selectors=foo",
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
				Actions: pktcls.ActionMap{
					"foo": MustInitSimpleFilter(t, "foo", "0-0#0"),
					"bar": MustInitSimpleFilter(t, "bar", "0-0#0"),
				},
			},
			Policies: map[addr.IA]string{
				{I: 1, A: 1}: "name=audio src=1.1.1.0/24 selectors=foo,bar",
				{I: 2, A: 2}: "name=video dst=2.2.2.0/24 selectors=bar",
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
				Actions: pktcls.ActionMap{
					"foo": MustInitSimpleFilter(t, "foo", "1-1#1"),
					"bar": MustInitSimpleFilter(t, "bar", "2-2#2"),
					"baz": MustInitSimpleFilter(t, "baz", "3-3#3"),
					"bad": MustInitSimpleFilter(t, "bad", "4-4#4"),
				},
			},
			Policies: map[addr.IA]string{
				{I: 1, A: 1}: strings.Join(
					[]string{
						"name=audio dscp=0x12 selectors=foo,bar",
						"name=video tos=0x34 selectors=bar,baz",
					},
					"\n"),
			},
		},
	}

	Convey("TestCompile", t, func() {
		for _, tc := range testCases {
			Convey(tc.Name, func() {
				err := Compile(tc.Config, tc.Policies)
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

func MustInitSimpleFilter(t *testing.T, name, predicate string) *pktcls.ActionFilterPaths {
	t.Helper()

	pp, err := spathmeta.NewPathPredicate(predicate)
	if err != nil {
		t.Fatal(err)
	}
	return &pktcls.ActionFilterPaths{
		Name: name,
		Cond: &pktcls.CondPathPredicate{
			PP: pp,
		},
	}
}
