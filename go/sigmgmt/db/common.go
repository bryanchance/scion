// Copyright 2018 Anapaya Systems

package db

import (
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

func (pp *PathPolicyFile) GetExtPolicies() ([]*pathpol.ExtPolicy, error) {
	var extPolicies []*pathpol.ExtPolicy
	for _, polMap := range pp.Code {
		for name, origPol := range polMap {
			policy := pathpol.ExtPolicy{}
			if origPol == nil || origPol.Policy == nil {
				policy.Policy = &pathpol.Policy{}
			} else {
				policy.Policy = origPol.Policy
			}
			if origPol != nil {
				policy.Extends = origPol.Extends
			}
			policy.Policy.Name = name
			extPolicies = append(extPolicies, &policy)
		}
	}
	return extPolicies, nil
}
