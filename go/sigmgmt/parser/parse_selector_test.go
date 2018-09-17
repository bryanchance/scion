// Copyright 2018 Anapaya Systems

package parser

import (
	"flag"
	"os"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/scionproto/scion/go/lib/log"
	"github.com/scionproto/scion/go/lib/pktcls"
	"github.com/scionproto/scion/go/lib/spath/spathmeta"
	"github.com/scionproto/scion/go/lib/xtest"
)

func TestPredicateValidation(t *testing.T) {
	testCases := []struct {
		Name      string
		Predicate string
		Valid     bool
	}{
		{
			Name:      "wildcard selector",
			Predicate: "0-0#0",
			Valid:     true,
		},
		{
			Name:      "normal selector",
			Predicate: "1-ff0:ff1:f2#5",
			Valid:     true,
		},
		{
			Name:      "BGP AS number",
			Predicate: "1-34523#5",
			Valid:     true,
		},
		{
			Name:      "bad selector, bad hex",
			Predicate: "1-ffxx:ff1:f2#5",
			Valid:     false,
		},
		{
			Name:      "bad selector, missing colon",
			Predicate: "1-ff1:f2#5",
			Valid:     false,
		},
		{
			Name:      "NOT predicate",
			Predicate: "NOT(1-ff0:ff1:f2#5)",
			Valid:     true,
		},
		{
			Name:      "bad NOT",
			Predicate: "NoT(1-ff0:ff1:f2#5)",
			Valid:     false,
		},
		{
			Name:      "bad selector in NOT",
			Predicate: "NOT(1-ff1:f2#5)",
			Valid:     false,
		},
		{
			Name:      "single ALL predicate",
			Predicate: "ALL(1-ff0:ff1:f2#5)",
			Valid:     true,
		},
		{
			Name:      "double ALL predicate",
			Predicate: "ALL(1-ff0:ff1:f2#5,0-0#0)",
			Valid:     true,
		},
		{
			Name:      "double NOT predicate",
			Predicate: "NOT(1-ff0:ff1:f2#5,0-0#0)",
			Valid:     false,
		},
		{
			Name:      "ANY predicate",
			Predicate: "ANY(1-ff0:ff1:f2#5,0-0#0)",
			Valid:     true,
		},
		{
			Name:      "large any predicate",
			Predicate: "any(1-ff0:ff1:f2#5,0-0#0,1-ff0:ff1:f2#5,0-0#0,1-ff0:ff1:f2#5,0-0#0)",
			Valid:     true,
		},
		{
			Name: "nested any & all predicate",
			Predicate: "any(1-ff0:ff1:f2#5,all(0-0#0,1-ff0:ff1:f2#5,0-0#0)," +
				"all(1-ff0:ff1:f2#5,0-0#0))",
			Valid: true,
		},
		{
			Name: "bad nested any & all predicate",
			Predicate: "any(1-ff0:ff1:f2#5,all(0-0#0,all(0-0#0,1-ff0:ff1:f2#5,0-0#0)," +
				"all(1-ff0:ff1:f2#5,0-0#0))",
			Valid: false,
		},
		{
			Name: "nested not, any & all predicate",
			Predicate: "any(1-ff0:ff1:f2#5,all(0-0#0,1-ff0:ff1:f2#5,0-0#0)," +
				"all(1-ff0:ff1:f2#5,not(1-0#1)))",
			Valid: true,
		},
	}

	Convey("TestPredicateValidation", t, func() {
		for _, tc := range testCases {
			Convey(tc.Name, func() {
				err := ValidatePredicate(tc.Predicate)
				if tc.Valid {
					SoMsg("err", err, ShouldBeNil)
				} else {
					SoMsg("err", err, ShouldNotBeNil)
				}
			})
		}
	})
}

func TestPredicateTree(t *testing.T) {
	testCases := []struct {
		Name      string
		Predicate string
		Tree      pktcls.Cond
	}{
		{
			Name:      "wildcard selector",
			Predicate: "0-0#0",
			Tree:      mustCondPathPredicate(t, "0-0#0"),
		},
		{
			Name:      "normal selector",
			Predicate: "1-ff0:ff1:f2#5",
			Tree:      mustCondPathPredicate(t, "1-ff0:ff1:f2#5"),
		},
		{
			Name:      "NOT predicate",
			Predicate: "NOT(1-ff0:ff1:f2#5)",
			Tree: pktcls.CondNot{
				Operand: mustCondPathPredicate(t, "1-ff0:ff1:f2#5"),
			},
		},
		{
			Name:      "single ALL predicate",
			Predicate: "ALL(1-ff0:ff1:f2#5)",
			Tree: pktcls.CondAllOf{
				mustCondPathPredicate(t, "1-ff0:ff1:f2#5"),
			},
		},
		{
			Name:      "double ALL predicate",
			Predicate: "ALL(1-ff0:ff1:f2#5,0-0#0)",
			Tree: pktcls.CondAllOf{
				mustCondPathPredicate(t, "1-ff0:ff1:f2#5"),
				mustCondPathPredicate(t, "0-0#0"),
			},
		},
		{
			Name:      "ANY predicate",
			Predicate: "ANY(1-ff0:ff1:f2#5,0-0#0)",
			Tree: pktcls.CondAnyOf{
				mustCondPathPredicate(t, "1-ff0:ff1:f2#5"),
				mustCondPathPredicate(t, "0-0#0"),
			},
		},
		{
			Name:      "large any predicate",
			Predicate: "any(1-ff0:ff1:f2#5,0-0#0,1-ff0:ff1:f2#5,0-0#0,1-ff0:ff1:f2#5,0-0#0)",
			Tree: pktcls.CondAnyOf{
				mustCondPathPredicate(t, "1-ff0:ff1:f2#5"),
				mustCondPathPredicate(t, "0-0#0"),
				mustCondPathPredicate(t, "1-ff0:ff1:f2#5"),
				mustCondPathPredicate(t, "0-0#0"),
				mustCondPathPredicate(t, "1-ff0:ff1:f2#5"),
				mustCondPathPredicate(t, "0-0#0"),
			},
		},
		{
			Name: "nested any & all predicate",
			Predicate: "any(1-ff0:ff1:f2#5,all(0-0#0,1-ff0:ff1:f2#5,0-0#0)," +
				"all(1-ff0:ff1:f2#5,0-0#0))",
			Tree: pktcls.CondAnyOf{
				mustCondPathPredicate(t, "1-ff0:ff1:f2#5"),
				pktcls.CondAllOf{
					mustCondPathPredicate(t, "0-0#0"),
					mustCondPathPredicate(t, "1-ff0:ff1:f2#5"),
					mustCondPathPredicate(t, "0-0#0"),
				},
				pktcls.CondAllOf{
					mustCondPathPredicate(t, "1-ff0:ff1:f2#5"),
					mustCondPathPredicate(t, "0-0#0"),
				},
			},
		},
		{
			Name: "nested not, any & all predicate",
			Predicate: "any(1-ff0:ff1:f2#5,all(0-0#0,1-ff0:ff1:f2#5,0-0#0)," +
				"all(1-ff0:ff1:f2#5,not(1-0#1)))",
			Tree: pktcls.CondAnyOf{
				mustCondPathPredicate(t, "1-ff0:ff1:f2#5"),
				pktcls.CondAllOf{
					mustCondPathPredicate(t, "0-0#0"),
					mustCondPathPredicate(t, "1-ff0:ff1:f2#5"),
					mustCondPathPredicate(t, "0-0#0"),
				},
				pktcls.CondAllOf{
					mustCondPathPredicate(t, "1-ff0:ff1:f2#5"),
					pktcls.CondNot{
						Operand: mustCondPathPredicate(t, "1-0#1"),
					},
				},
			},
		},
	}

	Convey("TestPredicateTree", t, func() {
		for _, tc := range testCases {
			Convey(tc.Name, func() {
				tree, err := BuildPredicateTree(tc.Predicate)
				SoMsg("err", err, ShouldBeNil)
				SoMsg("tree", tree, ShouldResemble, tc.Tree)
			})
		}
	})
}

func TestMain(m *testing.M) {
	flag.Parse()
	if !testing.Verbose() {
		log.Root().SetHandler(log.DiscardHandler())
	}
	os.Exit(m.Run())
}

func mustCondPathPredicate(t *testing.T, str string) *pktcls.CondPathPredicate {
	t.Helper()

	pp, err := spathmeta.NewPathPredicate(str)
	xtest.FailOnErr(t, err)
	return pktcls.NewCondPathPredicate(pp)
}
