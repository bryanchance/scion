// Copyright 2018 Anapaya Systems

package parser

import (
	"net"
	"testing"

	"github.com/scionproto/scion/go/lib/pktcls"

	. "github.com/smartystreets/goconvey/convey"
)

func TestTrafficClassValidation(t *testing.T) {
	testCases := []struct {
		Name  string
		Class string
		Valid bool
	}{
		{
			Name:  "src IPv4Cond",
			Class: "src=12.12.12.0/26",
			Valid: true,
		},
		{
			Name:  "dst IPv4Cond",
			Class: "dst=12.12.12.0/26",
			Valid: true,
		},
		{
			Name:  "bad dst IPv4Cond",
			Class: "dst=12.12.12.0",
			Valid: false,
		},
		{
			Name:  "dscp IPv4Cond",
			Class: "dscp=0x2",
			Valid: true,
		},
		{
			Name:  "bad dscp IPv4Cond",
			Class: "dscp=2",
			Valid: false,
		},
		{
			Name:  "NOT",
			Class: "NOT(dscp=0x2)",
			Valid: true,
		},
		{
			Name:  "bad NOT",
			Class: "Not(dscp=0x2)",
			Valid: false,
		},
		{
			Name:  "bad NOT ,",
			Class: "Not(dscp=0x2,)",
			Valid: false,
		},
		{
			Name:  "BOOL",
			Class: "BOOL=true",
			Valid: true,
		},
		{
			Name:  "bad BOOL",
			Class: "BOOL=True",
			Valid: false,
		},
		{
			Name:  "single ALL",
			Class: "ALL(dscp=0x2)",
			Valid: true,
		},
		{
			Name:  "double ALL",
			Class: "ALL(dscp=0x2,dst=12.12.12.0/24)",
			Valid: true,
		},
		{
			Name:  "single ANY",
			Class: "ANY(dscp=0x2)",
			Valid: true,
		},
		{
			Name:  "double ANY",
			Class: "ANY(dscp=0x2,dst=12.12.12.0/24)",
			Valid: true,
		},
		{
			Name:  "bad triple ANY",
			Class: "ANY(dscp=0x2,dst=12.12.12.0/24,)",
			Valid: false,
		},
		{
			Name:  "ANY ALL NOT src dst dscp",
			Class: "ANY(dscp=0x2,ALL(dst=12.12.12.0/24,dscp=0x2, NOT(src=2.2.2.0/28)))",
			Valid: true,
		},
	}

	Convey("TestTrafficClassValidation", t, func() {
		for _, tc := range testCases {
			Convey(tc.Name, func() {
				err := ValidateTrafficClass(tc.Class)
				if tc.Valid {
					SoMsg("err", err, ShouldBeNil)
				} else {
					SoMsg("err", err, ShouldNotBeNil)
				}
			})
		}
	})
}

func TestTrafficClassTree(t *testing.T) {
	_, net, _ := net.ParseCIDR("12.12.12.0/26")
	testCases := []struct {
		Name  string
		Class string
		Tree  pktcls.Cond
	}{
		{
			Name:  "src IPv4Cond",
			Class: "src=12.12.12.0/26",
			Tree: pktcls.NewCondIPv4(
				&pktcls.IPv4MatchSource{Net: net},
			),
		},
		{
			Name:  "dst IPv4Cond",
			Class: "dst=12.12.12.0/26",
			Tree: pktcls.NewCondIPv4(
				&pktcls.IPv4MatchDestination{Net: net},
			),
		},
		{
			Name:  "dscp IPv4Cond",
			Class: "dscp=0x2",
			Tree: pktcls.NewCondIPv4(
				&pktcls.IPv4MatchDSCP{DSCP: uint8(0x2)},
			),
		},
		{
			Name:  "NOT",
			Class: "NOT(dscp=0x2)",
			Tree: pktcls.CondNot{Operand: pktcls.NewCondIPv4(
				&pktcls.IPv4MatchDSCP{DSCP: uint8(0x2)},
			)},
		},
		{
			Name:  "BOOL",
			Class: "bool=true",
			Tree:  pktcls.CondBool(true),
		},
		{
			Name:  "single ALL",
			Class: "ALL(dscp=0x2)",
			Tree: pktcls.CondAllOf{pktcls.NewCondIPv4(
				&pktcls.IPv4MatchDSCP{DSCP: uint8(0x2)},
			)},
		},
		{
			Name:  "double ALL",
			Class: "ALL(dscp=0x2,dst=12.12.12.0/26)",
			Tree: pktcls.CondAllOf{
				pktcls.NewCondIPv4(&pktcls.IPv4MatchDSCP{DSCP: uint8(0x2)}),
				pktcls.NewCondIPv4(&pktcls.IPv4MatchDestination{Net: net})},
		},
		{
			Name:  "single ANY",
			Class: "ANY(dscp=0x2)",
			Tree: pktcls.CondAnyOf{pktcls.NewCondIPv4(
				&pktcls.IPv4MatchDSCP{DSCP: uint8(0x2)},
			)},
		},
		{
			Name:  "double ANY",
			Class: "ANY(dscp=0x2,dst=12.12.12.0/26)",
			Tree: pktcls.CondAnyOf{
				pktcls.NewCondIPv4(&pktcls.IPv4MatchDSCP{DSCP: uint8(0x2)}),
				pktcls.NewCondIPv4(&pktcls.IPv4MatchDestination{Net: net})},
		},
		{
			Name:  "ANY ALL NOT src dst dscp",
			Class: "ANY(dscp=0x2,ALL(dst=12.12.12.0/26,dscp=0x2, NOT(src=12.12.12.0/26)))",
			Tree: pktcls.CondAnyOf{
				pktcls.NewCondIPv4(&pktcls.IPv4MatchDSCP{DSCP: uint8(0x2)}),
				pktcls.CondAllOf{
					pktcls.NewCondIPv4(&pktcls.IPv4MatchDestination{Net: net}),
					pktcls.NewCondIPv4(&pktcls.IPv4MatchDSCP{DSCP: uint8(0x2)}),
					pktcls.CondNot{Operand: pktcls.NewCondIPv4(
						&pktcls.IPv4MatchSource{Net: net},
					)},
				},
			},
		},
	}

	Convey("TestTrafficClassTree", t, func() {
		for _, tc := range testCases {
			Convey(tc.Name, func() {
				tree, err := BuildClassTree(tc.Class)
				SoMsg("err", err, ShouldBeNil)
				SoMsg("tree", tree, ShouldResemble, tc.Tree)
			})
		}
	})
}
