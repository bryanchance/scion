// Copyright 2018 Anapaya Systems

package config

import (
	"flag"
	"net"
	"path/filepath"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/scionproto/scion/go/lib/addr"
	"github.com/scionproto/scion/go/lib/pktcls"
	"github.com/scionproto/scion/go/lib/spath/spathmeta"
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
						Sigs: SIGSet{
							"remote-1": &SIG{
								Id:        "remote-1",
								Addr:      net.ParseIP("192.0.2.1"),
								CtrlPort:  1234,
								EncapPort: 5678,
							},
							"remote-2": &SIG{
								Id:        "remote-2",
								Addr:      net.ParseIP("192.0.2.2"),
								CtrlPort:  65535,
								EncapPort: 0,
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
						Sigs: SIGSet{},
					},
					xtest.MustParseIA("1-ff00:0:3"): {
						Nets: []*IPNet{},
						Sigs: SIGSet{},
					},
					xtest.MustParseIA("1-ff00:0:4"): {
						Nets: []*IPNet{},
						Sigs: SIGSet{
							"remote-3": &SIG{
								Id:   "remote-3",
								Addr: net.ParseIP("2001:DB8::4"),
							},
						},
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
						Sigs: SIGSet{
							"remote-1": &SIG{
								Id:        "remote-1",
								Addr:      net.ParseIP("10.0.1.1"),
								CtrlPort:  1234,
								EncapPort: 5678,
							},
							"remote-2": &SIG{
								Id:        "remote-2",
								Addr:      net.ParseIP("10.0.1.2"),
								CtrlPort:  65535,
								EncapPort: 0,
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
						Sigs: SIGSet{},
					},
					xtest.MustParseIA("1-ff00:0:3"): {
						Nets: []*IPNet{},
						Sigs: SIGSet{},
					},
					xtest.MustParseIA("1-ff00:0:4"): {
						Nets: []*IPNet{},
						Sigs: SIGSet{
							"remote-3": &SIG{
								Id:   "remote-3",
								Addr: net.ParseIP("2001:DB8::4"),
							},
						},
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
						Sigs: SIGSet{
							"remote-1": &SIG{
								Id:        "remote-1",
								Addr:      net.ParseIP("10.0.1.1"),
								CtrlPort:  1234,
								EncapPort: 5678,
							},
							"remote-2": &SIG{
								Id:        "remote-2",
								Addr:      net.ParseIP("10.0.1.2"),
								CtrlPort:  65535,
								EncapPort: 0,
							},
						},
						Sessions: SessionMap{
							0x03: "filter-A",
							0x04: "filter-B",
						},
						PktPolicies: []*PktPolicy{
							{
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
						Sigs: SIGSet{},
					},
					xtest.MustParseIA("1-ff00:0:3"): {
						Nets: []*IPNet{},
						Sigs: SIGSet{},
					},
					xtest.MustParseIA("1-ff00:0:4"): {
						Nets: []*IPNet{},
						Sigs: SIGSet{
							"remote-3": &SIG{
								Id:   "remote-3",
								Addr: net.ParseIP("2001:DB8::4"),
							},
						},
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
				Classes: pktcls.ClassMap{
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
				Actions: pktcls.ActionMap{
					"filter-A": &pktcls.ActionFilterPaths{
						Name: "filter-A",
						Cond: pktcls.CondAnyOf{
							pktcls.CondAllOf{
								mustCondPathPredicate(t, "0-0#123"),
								mustCondPathPredicate(t, "0-0#234"),
							},
							pktcls.CondAllOf{
								mustCondPathPredicate(t, "0-0#345"),
								mustCondPathPredicate(t, "0-0#456"),
							},
						},
					},
					"filter-B": &pktcls.ActionFilterPaths{
						Name: "filter-B",
						Cond: pktcls.CondAnyOf{
							mustCondPathPredicate(t, "0-0#123"),
							mustCondPathPredicate(t, "0-0#134"),
						},
					},
				},
				ConfigVersion: 0,
			},
		},
	}

	Convey("Test SIG config marshal/unmarshal", t, func() {
		for _, tc := range testCases {
			if *update {
				xtest.MustMarshalJSONToFile(t, tc.Config, tc.FileName)
			}
			cfg, err := LoadFromFile(filepath.Join("testdata", tc.FileName))
			SoMsg("err", err, ShouldBeNil)
			SoMsg("cfg", *cfg, ShouldResemble, tc.Config)
		}
	})
}

func mustCondPathPredicate(t *testing.T, str string) *pktcls.CondPathPredicate {
	t.Helper()

	pp, err := spathmeta.NewPathPredicate(str)
	xtest.FailOnErr(t, err)
	return pktcls.NewCondPathPredicate(pp)
}
