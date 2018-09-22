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

package sigcmn

import (
	"flag"
	"fmt"
	"net"
	"time"

	"github.com/scionproto/scion/go/lib/addr"
	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/lib/log"
	"github.com/scionproto/scion/go/lib/pathmgr"
	"github.com/scionproto/scion/go/lib/sciond"
	"github.com/scionproto/scion/go/lib/snet"
	"github.com/scionproto/scion/go/sig/mgmt"
)

const (
	DefaultCtrlPort  = 10081
	DefaultEncapPort = 10080
	MaxPort          = (1 << 16) - 1
	SIGHdrSize       = 8
)

var (
	CtrlPort       = flag.Int("ctrlport", DefaultCtrlPort, "control data port (e.g., keepalives)")
	EncapPort      = flag.Int("encapport", DefaultEncapPort, "encapsulation data port")
	sciondPath     = flag.String("sciond", sciond.GetDefaultSCIONDPath(nil), "SCIOND socket path")
	dispatcherPath = flag.String("dispatcher", "", "SCION Dispatcher path")
	SigTun         = flag.String("tun", "sig", "Name of TUN device to create")
)

var (
	DefV4Net = &net.IPNet{IP: net.IPv4zero, Mask: net.CIDRMask(0, net.IPv4len*8)}
	DefV6Net = &net.IPNet{IP: net.IPv6zero, Mask: net.CIDRMask(0, net.IPv6len*8)}
)

var (
	IA       addr.IA
	Host     addr.HostAddr
	PathMgr  *pathmgr.PR
	CtrlConn snet.Conn
	MgmtAddr *mgmt.Addr
)

const (
	initAttempts = 100
	initInterval = time.Second
)

func Init(ia addr.IA, ip net.IP) error {
	var err error
	IA = ia
	Host = addr.HostFromIP(ip)
	if err = ValidatePort("local ctrl", *CtrlPort); err != nil {
		return err
	}
	if err = ValidatePort("local encap", *EncapPort); err != nil {
		return err
	}
	MgmtAddr = mgmt.NewAddr(Host, uint16(*CtrlPort), uint16(*EncapPort))

	// Initialize custom network context.
	timers := &pathmgr.Timers{
		NormalRefire: 10 * time.Second,
	}
	sd := sciond.NewService(*sciondPath)
	if PathMgr, err = pathmgr.New(sd, timers, log.Root()); err != nil {
		return common.NewBasicError("Error creating path manager", err)
	}
	network := snet.NewNetworkWithPR(ia, *dispatcherPath, PathMgr)
	// Initialize SCION local networking module
	err = initSNET(network, initAttempts, initInterval)
	if err != nil {
		return common.NewBasicError("Error initializing SCION Network module", err)
	}
	l4 := addr.NewL4UDPInfo(uint16(*CtrlPort))
	CtrlConn, err = snet.ListenSCION(
		"udp4", &snet.Addr{IA: IA, Host: &addr.AppAddr{L3: Host, L4: l4}})
	if err != nil {
		return common.NewBasicError("Error creating ctrl socket", err)
	}
	return nil
}

func CtrlSnetAddr() *snet.Addr {
	l4 := addr.NewL4UDPInfo(uint16(*CtrlPort))
	return &snet.Addr{IA: IA, Host: &addr.AppAddr{L3: Host, L4: l4}}
}

func EncapSnetAddr() *snet.Addr {
	l4 := addr.NewL4UDPInfo(uint16(*EncapPort))
	return &snet.Addr{IA: IA, Host: &addr.AppAddr{L3: Host, L4: l4}}
}

func ValidatePort(desc string, port int) error {
	if port < 1 || port > MaxPort {
		return common.NewBasicError(fmt.Sprintf("Invalid %s port", desc), nil,
			"min", 1, "max", MaxPort, "actual", port)
	}
	return nil
}

// initSNET initializes snet. The number of attempts is specified, as well as the sleep duration.
// This allows the service to wait for a limited time for sciond to become available
func initSNET(network *snet.SCIONNetwork, attempts int, sleep time.Duration) (err error) {
	// Initialize SCION local networking module
	for i := 0; i < attempts; i++ {
		if err = snet.InitWithNetwork(network); err == nil {
			break
		}
		log.Error("Unable to initialize snet", "Retry interval", sleep, "err", err)
		time.Sleep(sleep)
	}
	return err
}
