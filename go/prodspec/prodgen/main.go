// Copyright 2019 Anapaya Systems

// This program generates all the config files for a SCION deployment.
// CAUTION: The tool has no arguments and neither it should have any.
// This way, the build of the configs is fully deterministic and reproducible.

package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"

	"github.com/scionproto/scion/go/prodspec/schema"
)

type ISD struct {
	Name string
}

type ISDs map[int]*ISD

type AS struct {
	Sites []string
	ISDs  []int
	Core  bool
	MTU   int
	BR    map[string]*BR
	BS    map[string]*BS
	CS    map[string]*CS
	PS    map[string]*PS
	SIG   map[string]*SIG
}

type BR struct {
}

type BS struct {
}

type CS struct {
	Host string
}

type PS struct {
}

type SIG struct {
}

type Interface struct {
	SIMNumber    string
	IP4MaskInt   string
	Routes6      string
	IP4MaskExt   string
	IP6MaskInt   string
	Routes4      string
	OSName       string
	IP4Peer      string
	PublicKey    string
	PeerName     string
	SerialNumber string
	IMEI         string
	PhoneNumber  string
	PhysicalName string
	IP4PTP       string
}

type Host struct {
	MachineType  int
	SerialNumber string
	Interfaces   map[string]*Interface
}

type Site struct {
	Location string
	Hosts    map[string]*Host
}

type Organization struct {
	Name      string
	ShortName string
	Sites     map[string]*Site
	ASes      map[string]*AS
}

func main() {
	f, err := ioutil.ReadFile("/home/sustrik/layout/isd.yml")
	if err != nil {
		fail("Cannot open input file", err)
	}
	isds := ISDs{}
	err = yaml.Unmarshal(f, &isds)
	if err != nil {
		fail("Cannot parse the isd.yml file", err)
	}

	f, err = ioutil.ReadFile("/home/sustrik/layout/in7.yml")
	if err != nil {
		fail("Cannot open input file", err)
	}
	inp := Organization{}
	err = yaml.Unmarshal(f, &inp)
	if err != nil {
		fail("Cannot parse the input file", err)
	}

	// Build the prodspec.
	lt := schema.Layout{}

	for iname, isd := range isds {
		i := lt.NewISD(fmt.Sprintf("%d", iname))
		i.Name = isd.Name
	}

	org := lt.NewOrganization(inp.ShortName)
	for sname, site := range inp.Sites {
		st := lt.NewSite(sname + "." + inp.ShortName)
		org.AddSite(st)
		for hname, host := range site.Hosts {
			h := lt.NewHost(hname + "." + st.ID + ".fsnets.com")
			h.MachineType = host.MachineType
			h.SerialNumber = host.SerialNumber
			st.AddHost(h)
			for iname, iface := range host.Interfaces {
				i := lt.NewInterface(h.ID + "/" + iname)
				i.SIMNumber = iface.SIMNumber
				i.IP4MaskInt = iface.IP4MaskInt
				i.Routes6 = iface.Routes6
				i.IP4MaskExt = iface.IP4MaskExt
				i.IP6MaskInt = iface.IP6MaskInt
				i.Routes4 = iface.Routes4
				i.OSName = iface.OSName
				i.IP4Peer = iface.IP4Peer
				i.PublicKey = iface.PublicKey
				i.PeerName = iface.PeerName
				i.SerialNumber = iface.SerialNumber
				i.IMEI = iface.IMEI
				i.PhoneNumber = iface.PhoneNumber
				i.PhysicalName = iface.PhysicalName
				i.IP4PTP = iface.IP4PTP
				h.AddInterface(i)
			}
		}
	}
	for asname, as := range inp.ASes {
		a := lt.NewAS(asname)
		a.Core = as.Core
		a.MTU = as.MTU
		for _, iname := range as.ISDs {
			isd := lt.ISD[fmt.Sprintf("%d", iname)]
			if isd == nil {
				fail(fmt.Sprintf("Unknown ISD: %d", iname), nil)
			}
			a.AddISD(isd)
		}
		org.AddAS(a)

		for brname, _ := range as.BR {
			b := lt.NewBR(asname + "/" + brname)
			b.Name = fmt.Sprintf("br%d-%s-%s", as.ISDs[0], asname, brname)
			a.AddBR(b)
		}

		for bsname, _ := range as.BS {
			b := lt.NewBS(asname + "/" + bsname)
			b.Name = fmt.Sprintf("bs%d-%s-%s", as.ISDs[0], asname, bsname)
			a.AddBS(b)
		}

		for csname, _ := range as.CS {
			c := lt.NewCS(asname + "/" + csname)
			c.Name = fmt.Sprintf("cs%d-%s-%s", as.ISDs[0], asname, csname)
			a.AddCS(c)
		}

		for psname, _ := range as.PS {
			p := lt.NewPS(asname + "/" + psname)
			p.Name = fmt.Sprintf("cs%d-%s-%s", as.ISDs[0], asname, psname)
			a.AddPS(p)
		}

		for signame, _ := range as.SIG {
			s := lt.NewSIG(asname + "/" + signame)
			s.Name = fmt.Sprintf("sig%d-%s-%s", as.ISDs[0], asname, signame)
			a.AddSIG(s)
		}
	}

	err = lt.Save("prodspec.toml")
	if err != nil {
		fail("Cannot save layout file", err)
	}
}

func fail(msg string, err error) {
	if err == nil {
		fmt.Printf("%s\n", msg)
	} else {
		fmt.Printf("%s: %s\n", msg, err.Error())
	}
	os.Exit(1)
}
