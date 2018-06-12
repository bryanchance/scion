// Copyright 2018 Anapaya Systems

package ctrl

import (
	"strconv"
	"strings"

	"github.com/scionproto/scion/go/sigmgmt/db"
	"github.com/scionproto/scion/go/sigmgmt/util"
)

func stringToIntSlice(s string) ([]int, error) {
	sArr := strings.Split(s, ",")
	iArr := []int{}
	for _, str := range sArr {
		i, err := strconv.Atoi(str)
		if err != nil {
			return nil, err
		}
		iArr = append(iArr, i)
	}
	return iArr, nil
}

func intSliceToString(iArr []int) string {
	sArr := []string{}
	for _, i := range iArr {
		str := strconv.Itoa(i)
		sArr = append(sArr, str)
	}
	return strings.Join(sArr, ",")
}

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
