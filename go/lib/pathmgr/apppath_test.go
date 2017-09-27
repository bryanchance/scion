// Copyright 2017 ETH Zurich
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package pathmgr

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/netsec-ethz/scion/go/lib/addr"
	"github.com/netsec-ethz/scion/go/lib/sciond"
)

var (
	// First 3 paths are used in this test file, all 8 paths are used in revtable_test.go
	paths = map[string]*sciond.PathReply{
		"1-19.2-25": {
			ErrorCode: 0x0,
			Entries: []sciond.PathReplyEntry{
				{
					Path: sciond.FwdPathMeta{
						FwdPath: []uint8{0x1, 0x59, 0xae, 0xe3, 0xb4, 0x0, 0x1, 0x3, 0x0, 0x3f, 0x3,
							0xc0, 0x0, 0xf6, 0x15, 0x5b, 0x0, 0x3f, 0x1, 0x60, 0x26, 0x68, 0xee, 0x87,
							0x1, 0x3f, 0x0, 0x0, 0x17, 0x57, 0x14, 0x93, 0x1, 0x59, 0xae, 0xe3, 0xa9,
							0x0, 0x2, 0x3, 0x1, 0x3f, 0x2, 0xe0, 0x0, 0xca, 0x65, 0x4d, 0x0, 0x3f, 0x5,
							0x70, 0x12, 0x9, 0x22, 0xba, 0x1, 0x3f, 0x0, 0x0, 0x61, 0x7f, 0x57, 0xbd,
							0x0, 0x59, 0xae, 0xe3, 0xaf, 0x0, 0x2, 0x3, 0x1, 0x3f, 0x0, 0x0, 0x45, 0xf2,
							0x13, 0x18, 0x0, 0x3f, 0x3, 0x90, 0x42, 0x94, 0x5, 0xc, 0x0, 0x3f, 0x4,
							0xa0, 0x0, 0x13, 0x35, 0xeb},
						Mtu: 0x500,
						Interfaces: []sciond.PathInterface{
							{RawIsdas: IA("1-19"), IfID: 60}, {RawIsdas: IA("1-16"), IfID: 38},
							{RawIsdas: IA("1-16"), IfID: 22}, {RawIsdas: IA("1-13"), IfID: 23},
							{RawIsdas: IA("1-13"), IfID: 46}, {RawIsdas: IA("1-11"), IfID: 18},
							{RawIsdas: IA("1-11"), IfID: 87}, {RawIsdas: IA("2-21"), IfID: 97},
							{RawIsdas: IA("2-21"), IfID: 69}, {RawIsdas: IA("2-23"), IfID: 57},
							{RawIsdas: IA("2-23"), IfID: 66}, {RawIsdas: IA("2-25"), IfID: 74}}},
					HostInfo: sciond.HostInfo{
						Port: 0x756f,
						Addrs: struct {
							Ipv4 []uint8
							Ipv6 []uint8
						}{
							Ipv4: []uint8{0x7f, 0x0, 0x0, 0xb1},
							Ipv6: []uint8(nil),
						},
					},
				},
			},
		},
		"1-10.1-18": {
			ErrorCode: 0x0,
			Entries: []sciond.PathReplyEntry{
				{
					Path: sciond.FwdPathMeta{
						FwdPath: []uint8{0x7, 0x59, 0xaf, 0xe0, 0xe4, 0x0, 0x1, 0x5, 0x0, 0x3f, 0x3,
							0x30, 0x0, 0x13, 0x22, 0x7f, 0x0, 0x3f, 0x3, 0xc0, 0x31, 0x7d, 0x5c, 0x44,
							0x1, 0x3f, 0x1, 0x60, 0x26, 0xf8, 0xe5, 0x0, 0x1, 0x3f, 0x1, 0xe0, 0x26,
							0x7, 0x46, 0x27, 0x2, 0x3f, 0x0, 0x0, 0x17, 0x3e, 0x70, 0x74, 0x6, 0x59,
							0xaf, 0xe0, 0xe7, 0x0, 0x1, 0x4, 0x2, 0x3f, 0x0, 0x0, 0x23, 0x87, 0x1b,
							0xf2, 0x1, 0x3f, 0x2, 0x30, 0x54, 0x41, 0xfb, 0x89, 0x1, 0x3f, 0x4, 0x0,
							0x54, 0x42, 0x92, 0x6f, 0x0, 0x3f, 0x2, 0x80, 0x0, 0xd, 0x4f, 0x20},
						Mtu: 0x5c0,
						Interfaces: []sciond.PathInterface{
							{RawIsdas: IA("1-10"), IfID: 51}, {RawIsdas: IA("1-19"), IfID: 49},
							{RawIsdas: IA("1-19"), IfID: 60}, {RawIsdas: IA("1-16"), IfID: 38},
							{RawIsdas: IA("1-16"), IfID: 30}, {RawIsdas: IA("1-15"), IfID: 35},
							{RawIsdas: IA("1-15"), IfID: 84}, {RawIsdas: IA("1-18"), IfID: 40}}},
					HostInfo: sciond.HostInfo{
						Port: 0x7579,
						Addrs: struct {
							Ipv4 []uint8
							Ipv6 []uint8
						}{
							Ipv4: []uint8{0x7f, 0x0, 0x0, 0x31},
							Ipv6: []uint8(nil),
						},
					},
				},
			},
		},
		"2-24.1-17": {
			ErrorCode: 0x0,
			Entries: []sciond.PathReplyEntry{
				{
					Path: sciond.FwdPathMeta{
						FwdPath: []uint8{0x1, 0x59, 0xaf, 0xac, 0xd7, 0x0, 0x2, 0x2, 0x0, 0x3f, 0x4,
							0xb0, 0x0, 0x85, 0x25, 0xf2, 0x1, 0x3f, 0x0, 0x0, 0x4d, 0xed, 0xf, 0xf5,
							0x1, 0x59, 0xaf, 0xac, 0xd2, 0x0, 0x1, 0x3, 0x1, 0x3f, 0x3, 0x10, 0x0, 0x23,
							0x9a, 0xc4, 0x0, 0x3f, 0x5, 0xf0, 0x17, 0x25, 0x4, 0xad, 0x1, 0x3f, 0x0,
							0x0, 0x36, 0x5d, 0x8e, 0x22, 0x0, 0x59, 0xaf, 0xac, 0xd2, 0x0, 0x1, 0x3,
							0x1, 0x3f, 0x0, 0x0, 0x1b, 0x6b, 0xed, 0x2d, 0x0, 0x3f, 0x3, 0xa0, 0x31,
							0xe, 0x15, 0xbe, 0x0, 0x3f, 0x0, 0xe0, 0x0, 0x8f, 0x4b, 0xf5},
						Mtu: 0x578,
						Interfaces: []sciond.PathInterface{
							{RawIsdas: IA("2-26"), IfID: 75}, {RawIsdas: IA("2-22"), IfID: 77},
							{RawIsdas: IA("2-22"), IfID: 49}, {RawIsdas: IA("1-12"), IfID: 23},
							{RawIsdas: IA("1-12"), IfID: 95}, {RawIsdas: IA("1-11"), IfID: 54},
							{RawIsdas: IA("1-11"), IfID: 28}, {RawIsdas: IA("1-14"), IfID: 48},
							{RawIsdas: IA("1-14"), IfID: 49}, {RawIsdas: IA("1-17"), IfID: 14}}},
					HostInfo: sciond.HostInfo{
						Port: 0x7563,
						Addrs: struct {
							Ipv4 []uint8
							Ipv6 []uint8
						}{
							Ipv4: []uint8{0x7f, 0x0, 0x0, 0xf1},
							Ipv6: []uint8(nil),
						},
					},
				},
			},
		},
		"2-22.1-16": {
			ErrorCode: 0x0,
			Entries: []sciond.PathReplyEntry{
				{
					Path: sciond.FwdPathMeta{
						FwdPath: []uint8{0x1, 0x59, 0xaf, 0xad, 0x40, 0x0, 0x1, 0x3, 0x0, 0x3f, 0x3,
							0x10, 0x0, 0x35, 0xfd, 0x5f, 0x0, 0x3f, 0x4, 0x20, 0x17, 0xb6, 0x63, 0xc,
							0x1, 0x3f, 0x0, 0x0, 0x55, 0xd9, 0xfd, 0xc5, 0x0, 0x59, 0xaf, 0xad, 0x4a,
							0x0, 0x1, 0x2, 0x1, 0x3f, 0x0, 0x0, 0x17, 0x28, 0x64, 0xfa, 0x0, 0x3f, 0x1,
							0x60, 0x0, 0xb8, 0x5b, 0xd4},
						Mtu: 0x578,
						Interfaces: []sciond.PathInterface{
							{RawIsdas: IA("2-22"), IfID: 49}, {RawIsdas: IA("1-12"), IfID: 23},
							{RawIsdas: IA("1-12"), IfID: 66}, {RawIsdas: IA("1-13"), IfID: 85},
							{RawIsdas: IA("1-13"), IfID: 23}, {RawIsdas: IA("1-16"), IfID: 22}}},
					HostInfo: sciond.HostInfo{
						Port: 0x7574,
						Addrs: struct {
							Ipv4 []uint8
							Ipv6 []uint8
						}{
							Ipv4: []uint8{0x7f, 0x0, 0x0, 0xd1},
							Ipv6: []uint8(nil),
						},
					},
				},
			},
		},
		"1-18.2-25": {
			ErrorCode: 0x0,
			Entries: []sciond.PathReplyEntry{
				{
					Path: sciond.FwdPathMeta{
						FwdPath: []uint8{0x1, 0x59, 0xaf, 0xad, 0x6d, 0x0, 0x1, 0x3, 0x0, 0x3f, 0x2,
							0x80, 0x0, 0x69, 0xb6, 0xf0, 0x0, 0x3f, 0x4, 0x0, 0x54, 0xdc, 0xe9, 0x63,
							0x1, 0x3f, 0x0, 0x0, 0x23, 0x19, 0xe8, 0xaf, 0x1, 0x59, 0xaf, 0xad, 0x70,
							0x0, 0x2, 0x3, 0x1, 0x3f, 0x1, 0x70, 0x0, 0x19, 0x33, 0xae, 0x0, 0x3f, 0x0,
							0xa0, 0x31, 0x19, 0x26, 0xbf, 0x1, 0x3f, 0x0, 0x0, 0x16, 0xac, 0x87, 0xa2,
							0x0, 0x59, 0xaf, 0xad, 0x70, 0x0, 0x2, 0x3, 0x1, 0x3f, 0x0, 0x0, 0x45, 0x41,
							0x25, 0x1b, 0x0, 0x3f, 0x3, 0x90, 0x42, 0xb1, 0x7d, 0x53, 0x0, 0x3f, 0x4,
							0xa0, 0x0, 0xb4, 0x68, 0x31},
						Mtu: 0x500,
						Interfaces: []sciond.PathInterface{
							{RawIsdas: IA("1-18"), IfID: 40}, {RawIsdas: IA("1-15"), IfID: 84},
							{RawIsdas: IA("1-15"), IfID: 64}, {RawIsdas: IA("1-12"), IfID: 35},
							{RawIsdas: IA("1-12"), IfID: 23}, {RawIsdas: IA("2-22"), IfID: 49},
							{RawIsdas: IA("2-22"), IfID: 10}, {RawIsdas: IA("2-21"), IfID: 22},
							{RawIsdas: IA("2-21"), IfID: 69}, {RawIsdas: IA("2-23"), IfID: 57},
							{RawIsdas: IA("2-23"), IfID: 66}, {RawIsdas: IA("2-25"), IfID: 74}}},
					HostInfo: sciond.HostInfo{
						Port: 0x7574,
						Addrs: struct {
							Ipv4 []uint8
							Ipv6 []uint8
						}{
							Ipv4: []uint8{0x7f, 0x0, 0x0, 0xa1},
							Ipv6: []uint8(nil),
						},
					},
				},
			},
		},
		"2-21.2-26": {
			ErrorCode: 0x0,
			Entries: []sciond.PathReplyEntry{
				{
					Path: sciond.FwdPathMeta{
						FwdPath: []uint8{0x0, 0x59, 0xaf, 0xad, 0x90, 0x0, 0x2, 0x3, 0x0, 0x3f, 0x0,
							0x0, 0x45, 0x97, 0x85, 0x64, 0x0, 0x3f, 0x3, 0x90, 0x11, 0x1, 0xc7, 0x2a,
							0x0, 0x3f, 0x2, 0x20, 0x0, 0xbd, 0x8, 0xb1},
						Mtu: 0x500,
						Interfaces: []sciond.PathInterface{
							{RawIsdas: IA("2-21"), IfID: 69}, {RawIsdas: IA("2-23"), IfID: 57},
							{RawIsdas: IA("2-23"), IfID: 17}, {RawIsdas: IA("2-26"), IfID: 34}}},
					HostInfo: sciond.HostInfo{
						Port: 0x7581,
						Addrs: struct {
							Ipv4 []uint8
							Ipv6 []uint8
						}{
							Ipv4: []uint8{0x7f, 0x0, 0x0, 0xc3},
							Ipv6: []uint8(nil),
						},
					},
				},
			},
		},
		"1-11.2-23": {
			ErrorCode: 0x0,
			Entries: []sciond.PathReplyEntry{
				{
					Path: sciond.FwdPathMeta{
						FwdPath: []uint8{0x1, 0x59, 0xaf, 0xad, 0xc2, 0x0, 0x2, 0x2, 0x0, 0x3f, 0x5,
							0x70, 0x0, 0x7e, 0xcf, 0xc5, 0x1, 0x3f, 0x0, 0x0, 0x61, 0x3d, 0xe5, 0xde,
							0x0, 0x59, 0xaf, 0xad, 0xc2, 0x0, 0x2, 0x2, 0x1, 0x3f, 0x0, 0x0, 0x45, 0x67,
							0xf7, 0xe0, 0x0, 0x3f, 0x3, 0x90, 0x0, 0x63, 0x7e, 0x27},
						Mtu: 0x500,
						Interfaces: []sciond.PathInterface{
							{RawIsdas: IA("1-11"), IfID: 87}, {RawIsdas: IA("2-21"), IfID: 97},
							{RawIsdas: IA("2-21"), IfID: 69}, {RawIsdas: IA("2-23"), IfID: 57}}},
					HostInfo: sciond.HostInfo{
						Port: 0x7590,
						Addrs: struct {
							Ipv4 []uint8
							Ipv6 []uint8
						}{
							Ipv4: []uint8{0x7f, 0x0, 0x0, 0x43},
							Ipv6: []uint8(nil),
						},
					},
				},
			},
		},
		"1-13.1-18": {
			ErrorCode: 0x0,
			Entries: []sciond.PathReplyEntry{
				{
					Path: sciond.FwdPathMeta{
						FwdPath: []uint8{0x1, 0x59, 0xaf, 0xad, 0xd9, 0x0, 0x1, 0x3, 0x0, 0x3f, 0x2,
							0xe0, 0x0, 0x94, 0x6, 0x85, 0x0, 0x3f, 0x3, 0x60, 0x12, 0xd8, 0x22, 0x8b,
							0x1, 0x3f, 0x0, 0x0, 0x5f, 0x2b, 0xfc, 0x2f, 0x0, 0x59, 0xaf, 0xad, 0xd9,
							0x0, 0x1, 0x3, 0x1, 0x3f, 0x0, 0x0, 0x23, 0x3f, 0x86, 0x5, 0x0, 0x3f, 0x4,
							0x0, 0x54, 0x32, 0xb3, 0xac, 0x0, 0x3f, 0x2, 0x80, 0x0, 0xe8, 0xb0, 0x2e},
						Mtu: 0x5c0,
						Interfaces: []sciond.PathInterface{
							{RawIsdas: IA("1-13"), IfID: 46}, {RawIsdas: IA("1-11"), IfID: 18},
							{RawIsdas: IA("1-11"), IfID: 54}, {RawIsdas: IA("1-12"), IfID: 95},
							{RawIsdas: IA("1-12"), IfID: 35}, {RawIsdas: IA("1-15"), IfID: 64},
							{RawIsdas: IA("1-15"), IfID: 84}, {RawIsdas: IA("1-18"), IfID: 40}}},
					HostInfo: sciond.HostInfo{
						Port: 0x7562,
						Addrs: struct {
							Ipv4 []uint8
							Ipv6 []uint8
						}{
							Ipv4: []uint8{0x7f, 0x0, 0x0, 0x61},
							Ipv6: []uint8(nil),
						},
					},
				},
			},
		},
	}
)

func TestAppPathSets(t *testing.T) {
	Convey("Construct path set with one path", t, func() {
		aps := NewAppPathSet(paths["1-19.2-25"])
		SoMsg("len", len(aps), ShouldEqual, 1)

		Convey("API users can retrieve the path", func() {
			path := aps.GetAppPath()
			SoMsg("path", path, ShouldNotEqual, nil)
		})

		Convey("Revoke removes the path", func() {
			for _, v := range aps {
				v.revoke()
			}
			SoMsg("len", len(aps), ShouldEqual, 0)

			Convey("API users can no longer retrieve paths", func() {
				path := aps.GetAppPath()
				SoMsg("path", path, ShouldEqual, nil)
			})
		})

		Convey("Two more paths can be added", func() {
			// These paths are between different pairs of sources and destinations so they should
			// never appear in the same path set. This is done here for testing purposes only.
			aps.addChildAppPath(&paths["1-10.1-18"].Entries[0])
			aps.addChildAppPath(&paths["2-24.1-17"].Entries[0])
			SoMsg("len", len(aps), ShouldEqual, 3)

			Convey("Revoking one of them leaves the other two", func() {
				for _, v := range aps {
					v.revoke()
					break
				}
				SoMsg("len", len(aps), ShouldEqual, 2)
			})
		})
	})
}

func IA(iaStr string) uint32 {
	ia, _ := addr.IAFromString(iaStr)
	return ia.Uint32()
}
