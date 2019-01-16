// Copyright 2018 Anapaya Systems

package ctrl

import (
	"github.com/scionproto/scion/go/sigmgmt/db"
	"github.com/scionproto/scion/go/sigmgmt/util"
)

func parseHosts(hosts []db.Host) error {
	for _, host := range hosts {
		if err := util.ValidateIdentifier(host.Name); err != nil {
			return err
		}
		if err := util.ValidateUser(host.User); host.User != "" && err != nil {
			return err
		}
		if err := util.ValidateKey(host.Key); host.Key != "" && err != nil {
			return err
		}
	}
	return nil
}

func validateSite(site *db.Site) (string, error) {
	if site.VHost != "" {
		if err := util.ValidateIdentifier(site.VHost); err != nil {
			return "Bad VHost", err
		}
	}
	if err := parseHosts(site.Hosts); err != nil {
		return "Bad hosts", err
	}
	return "", nil
}
