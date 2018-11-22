// Copyright 2018 Anapaya Systems

// Package quagga provides an interface to announce and retract routes to the Quagga Routing Suite.
package quagga

import (
	"net"
	"sync"

	"github.com/osrg/gobgp/zebra"

	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/lib/log"
	"github.com/scionproto/scion/go/sig/base"
	"github.com/scionproto/scion/go/sig/internal/sigconfig"
	"github.com/scionproto/scion/go/sig/sigcmn"
)

const (
	sigRouteMetric  = 99
	sigRouteType    = 15
	zebraAPIVersion = 3
)

var (
	cli  *zebra.Client
	lock sync.Mutex
)

func Init(cfg sigconfig.Conf) error {
	if !cfg.ExportRoutes {
		return nil
	}
	if cli != nil {
		return common.NewBasicError("Quagga exporter can only be initialized once", nil)
	}
	var err error
	cli, err = zebra.NewClient("unix", cfg.ZServApi, sigRouteType, zebraAPIVersion)
	if err != nil {
		return err
	}
	// Register callbacks
	cbs := base.EventCallbacks{
		NetworkChanged:      networkChanged,
		RemoteHealthChanged: remoteHealthChanged,
	}
	base.AddEventListener("quagga-exporter", cbs)
	go func() {
		defer log.LogPanicAndExit()
		drain()
	}()
	log.Info("Quagga exporter initialized", "socket", cfg.ZServApi)
	return nil
}

func networkChanged(params base.NetworkChangedParams) {
	lock.Lock()
	defer lock.Unlock()
	updateRoute(params.IpNet, !params.Added)
}

func remoteHealthChanged(params base.RemoteHealthChangedParams) {
	lock.Lock()
	defer lock.Unlock()
	for _, ipnet := range params.Nets {
		updateRoute(*ipnet, !params.Healthy)
	}
}

func updateRoute(n net.IPNet, withdraw bool) {
	prefixLen, _ := n.Mask.Size()
	routeBody := &zebra.IPRouteBody{
		Type:         sigRouteType,
		SAFI:         zebra.SAFI_UNICAST,
		Prefix:       n.IP,
		PrefixLength: uint8(prefixLen),
	}
	if !withdraw {
		routeBody.Message = zebra.MESSAGE_METRIC | zebra.MESSAGE_NEXTHOP
		routeBody.Nexthops = []net.IP{sigcmn.Host.IP()}
		routeBody.Metric = sigRouteMetric
	}
	cli.SendIPRoute(0, routeBody, withdraw)
	log.Debug("Route updated", "prefix", n, "withdraw", withdraw)
}

func drain() {
	defer log.LogPanicAndExit()
	for {
		m, ok := <-cli.Receive()
		if !ok {
			// TODO(shitz): We should do something more here than just logging this and give up.
			log.Error("Connection to Quagga lost")
			return
		}
		log.Debug("Received message from Quagga", "header", m.Header, "body", m.Body)
	}
}
