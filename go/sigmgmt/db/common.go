// Copyright 2018 Anapaya Systems

package db

import (
	"net"

	"github.com/scionproto/scion/go/lib/pktcls"
	"github.com/scionproto/scion/go/sig/anaconfig"
	"github.com/scionproto/scion/go/sigmgmt/parser"
)

// IPNetsFromNetworks converts a slice of networks to a slice of IPNet
func IPNetsFromNetworks(networks []Network) ([]*config.IPNet, error) {
	ipNets := []*config.IPNet{}
	for _, network := range networks {
		_, ipNet, err := net.ParseCIDR(network.CIDR)
		if err != nil {
			return nil, err
		}
		ipNets = append(ipNets, (*config.IPNet)(ipNet))
	}
	return ipNets, nil
}

// ActionMapFromSelectors converts a slice of path selectors to an ActionMap
func ActionMapFromSelectors(selectors []PathSelector) (pktcls.ActionMap, error) {
	actionMap := make(pktcls.ActionMap)
	for _, selector := range selectors {
		condTree, err := parser.BuildPredicateTree(selector.Filter)
		if err != nil {
			return nil, err
		}
		actionMap[selector.Name] = pktcls.NewActionFilterPaths(selector.Name, condTree)
	}
	return actionMap, nil
}
