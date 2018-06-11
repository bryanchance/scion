// Copyright 2018 Anapaya Systems

package db

import (
	"github.com/scionproto/scion/go/lib/addr"
)

type Site struct {
	ID            uint
	Name          string `gorm:"unique;not null"`
	VHost         string
	MetricsPort   uint16
	Hosts         []Host         `gorm:"association_autoupdate:false;association_autocreate:false"`
	ASEntries     []ASEntry      `json:",omitempty"`
	PathSelectors []PathSelector `json:",omitempty"`
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

type ASEntry struct {
	ID       uint
	Name     string
	ISD      string    `gorm:"column:isd_id"`
	AS       string    `gorm:"column:as_id"`
	Policies string    `gorm:"column:policies"`
	SiteID   uint      `sql:"type:integer REFERENCES sites ON DELETE CASCADE ON UPDATE CASCADE, UNIQUE (site_id, name, isd_id, as_id)" json:"-"`
	SIGs     []SIG     `json:",omitempty"`
	Networks []Network `json:",omitempty"`
}

func (ASEntry) TableName() string {
	return "asentries"
}

func (as *ASEntry) ToAddrIA() (addr.IA, error) {
	return addr.IAFromString(as.ISD + "-" + as.AS)
}

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
