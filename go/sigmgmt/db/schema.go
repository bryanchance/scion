// Copyright 2018 Anapaya Systems

package db

import (
	"github.com/scionproto/scion/go/lib/addr"
	"github.com/scionproto/scion/go/lib/pathpol"
	"github.com/scionproto/scion/go/lib/pktcls"
)

// Site is an AS owned and managed by this instance of sigmgmt.
type Site struct {
	ID             uint
	Name           string `gorm:"unique;not null"`
	VHost          string
	MetricsPort    uint16
	Hosts          []Host           `gorm:"association_autoupdate:false;association_autocreate:false"`
	ASEntries      []ASEntry        `json:",omitempty"`
	PathSelectors  []PathSelector   `json:",omitempty"`
	TrafficClasses []TrafficClass   `json:",omitempty"`
	PathPolicies   []PathPolicyFile `json:",omitempty"`
}

type Host struct {
	ID     uint
	Name   string
	User   string
	Key    string `gorm:"size:400"`
	SiteID uint   `sql:"type:integer REFERENCES sites ON DELETE CASCADE ON UPDATE CASCADE, UNIQUE (site_id, name)" json:"-"`
}

type PathSelector struct {
	ID     uint
	Name   string
	Filter string
	SiteID uint `sql:"type:integer REFERENCES sites ON DELETE CASCADE ON UPDATE CASCADE, UNIQUE (site_id, name), UNIQUE (site_id, filter)" json:"-"`
}

// ASEntry is a reference to a remote AS.
// Remote AS may or may not be managed by this sigmgmt instance.
type ASEntry struct {
	ID                   uint
	Name                 string
	ISD                  string          `gorm:"column:isd_id"`
	AS                   string          `gorm:"column:as_id"`
	Policies             string          `gorm:"column:policies"`
	IPAllocationProvider string          `json:",omitempty"`
	SiteID               uint            `sql:"type:integer REFERENCES sites ON DELETE CASCADE ON UPDATE CASCADE, UNIQUE (site_id, name, isd_id, as_id)" json:"-"`
	SIGs                 []SIG           `json:",omitempty"`
	Networks             []Network       `json:",omitempty"`
	Policy               []TrafficPolicy `json:",omitempty"`
}

func (ASEntry) TableName() string {
	return "asentries"
}

func (as *ASEntry) ToAddrIA() (addr.IA, error) {
	return addr.IAFromString(as.ISD + "-" + as.AS)
}

// SIGs field is obsolete since SIGs are now discovered automatically.
// Keeping it in place not to mess with the existing database deployments.
type SIG struct {
	ID        uint
	Name      string
	Address   string
	CtrlPort  uint16
	EncapPort uint16
	ASEntryID uint `sql:"type:integer REFERENCES asentries ON DELETE CASCADE ON UPDATE CASCADE, UNIQUE (name, as_entry_id)" json:"-"`
}

type Network struct {
	ID        uint
	CIDR      string `gorm:"column:cidr"`
	ASEntryID uint   `sql:"type:integer REFERENCES asentries ON DELETE CASCADE ON UPDATE CASCADE, UNIQUE (cidr, as_entry_id)" json:"-"`
}

type PathPolicyFile struct {
	ID      uint
	Name    string
	CodeStr string                          `json:"-"`
	Code    []map[string]*pathpol.ExtPolicy `gorm:"-"`
	Type    ppFileType
	SiteID  *uint `sql:"type:integer REFERENCES sites ON DELETE CASCADE ON UPDATE CASCADE, UNIQUE (site_id)"`
}

func (PathPolicyFile) TableName() string {
	return "path_policies"
}

type ppFileType string

const (
	// GlobalPolicy is a Path Policy File for all sites
	GlobalPolicy ppFileType = "global"
	// SitePolicy is a Path Policy File for the referenced site
	SitePolicy ppFileType = "site"
)

type TrafficPolicy struct {
	ID              uint
	Name            string
	TrafficClass    uint     `sql:"type:integer REFERENCES traffic_classes ON DELETE CASCADE ON UPDATE CASCADE"`
	Selectors       string   `json:"-"`
	SelectorIDs     []int    `gorm:"-" json:"Selectors"`
	ASEntryID       uint     `sql:"type:integer REFERENCES asentries ON DELETE CASCADE ON UPDATE CASCADE" json:"-"`
	PathPolicies    string   `json:"-"`
	PathPolicyNames []string `gorm:"-" json:"PathPolicies"`
}

func (TrafficPolicy) TableName() string {
	return "policies"
}

type TrafficClass struct {
	ID      uint
	Name    string
	CondStr string                 `gorm:"column:cond" json:",omitempty"`
	Cond    map[string]pktcls.Cond `gorm:"-"`
	SiteID  uint                   `sql:"type:integer REFERENCES sites ON DELETE CASCADE ON UPDATE CASCADE, UNIQUE (site_id, name), UNIQUE (site_id, cond)" json:"-"`
}
