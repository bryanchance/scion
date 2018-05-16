// Copyright 2018 Anapaya Systems

package db

import (
	"fmt"
	"strconv"

	"github.com/scionproto/scion/go/lib/addr"
)

func applyConvertUInt8(input []uint8) []string {
	var output []string
	for _, object := range input {
		output = append(output, strconv.Itoa(int(object)))
	}
	return output
}

type SessionAliasMap map[string]string

type Site struct {
	Name        string
	VHost       string
	Hosts       []Host
	MetricsPort uint16
}

type Host struct {
	Name string
	User string
	Key  string
}

type Filter struct {
	Name string
	PP   string
}

type ASConfig struct {
	Name  string
	Value string
}

type AS struct {
	Name   string
	ISD    string
	AS     string
	Policy string
}

func ASFromAddrIA(ia addr.IA) *AS {
	return &AS{ISD: fmt.Sprint(ia.I), AS: ia.A.String()}
}

func (as *AS) ToAddrIA() (addr.IA, error) {
	iaStr := fmt.Sprintf("%s-%s", as.ISD, as.AS)
	return addr.IAFromString(iaStr)
}
