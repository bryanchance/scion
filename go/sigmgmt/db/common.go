// Copyright 2018 Anapaya Systems

package db

import (
	"encoding/json"
	"net"

	"github.com/scionproto/scion/go/lib/pathpol"
	"github.com/scionproto/scion/go/sig/anaconfig"
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

func (pp *PathPolicyFile) GetExtPolicies() (map[string]*pathpol.ExtPolicy, error) {
	var policies []string
	err := json.Unmarshal([]byte(pp.CodeStr), &policies)
	if err != nil {
		return nil, err
	}
	extPolicies := make(map[string]*pathpol.ExtPolicy, len(policies))
	for _, policy := range policies {
		var extPolicy pathpol.ExtPolicy
		err := json.Unmarshal([]byte(policy), &extPolicy)
		if err != nil {
			return nil, err
		}
		extPolicies[extPolicy.Name] = &extPolicy
	}
	return extPolicies, nil
}
