// Copyright 2019 Anapaya Systems

package main

import (
	"fmt"
	"os"

	"github.com/scionproto/scion/go/prodspec/schema"
)

func main() {
	lt, hash, err := schema.LoadUnvalidated("prodspec.toml")
	if err != nil {
		fail("Cannot load ProdSpec file", err)
	}

	validateLayout(lt)
	for _, v := range lt.Organization {
		validateOrganization(v)
	}
	for _, v := range lt.Host {
		validateHost(v)
	}
	for _, v := range lt.Interface {
		validateInterface(v)
	}
	for _, v := range lt.ISD {
		validateISD(v)
	}
	for _, v := range lt.AS {
		validateAS(v)
	}
	for _, v := range lt.BR {
		validateBR(v)
	}
	for _, v := range lt.BS {
		validateBS(v)
	}
	for _, v := range lt.CS {
		validateCS(v)
	}
	for _, v := range lt.PS {
		validatePS(v)
	}
	for _, v := range lt.SIG {
		validateSIG(v)
	}

	vf, err := os.Create("validated")
	if err != nil {
		fail("Cannot create 'validated' file", err)
	}
	_, err = vf.WriteString(hash)
	if err != nil {
		fail("Cannot write to 'validated' file", err)
	}
}

func fail(msg string, err error) {
	fmt.Printf("%s: %s\n", msg, err.Error())
	os.Exit(1)
}
