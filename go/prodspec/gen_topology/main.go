// Copyright 2019 Anapaya Systems

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/scionproto/scion/go/lib/topology"
	"github.com/scionproto/scion/go/prodspec/schema"
)

func main() {
	lt, err := schema.Load("prodspec.toml")
	if err != nil {
		fail("Cannot load ProdSpec file", err)
	}

	for _, as := range lt.AS {
		// TODO(sustrik): Here we are assuming that each AS belongs to exactly one ISD.
		path := fmt.Sprintf("ISD%s/AS%s", as.ISD[0].ID, strings.Replace(as.ID, ":", "_", -1))
		err = os.MkdirAll(path, 0777)
		if err != nil {
			fail("Cannot create AS directory", err)
		}
		topo := topology.RawTopo{}

		topo.Core = as.Core
		topo.Overlay = "UDP/IPv4"
		topo.ISD_AS = fmt.Sprintf("%s-%s", as.ISD[0].ID, as.ID)
		topo.MTU = as.MTU

		topo.BorderRouters = map[string]*topology.RawBRInfo{}
		for _, br := range as.BR {
			topo.BorderRouters[br.Name] = &topology.RawBRInfo{}
		}

		// Save to topology.json file.
		bytes, err := json.MarshalIndent(topo, "", "  ")
		if err != nil {
			fail("Cannot marshal the topology", err)
		}
		f, err := os.Create(path + "/topology.json")
		if err != nil {
			fail("Cannot create topology file", err)
		}
		defer f.Close()
		_, err = f.Write(bytes)
		if err != nil {
			fail("Cannot write to topology file", err)
		}
	}
}

func fail(msg string, err error) {
	fmt.Printf("%s: %s\n", msg, err.Error())
	os.Exit(1)
}
