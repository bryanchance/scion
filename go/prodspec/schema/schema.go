// This file was generated from schema.yml by schema/generator

package schema

import "github.com/scionproto/scion/go/lib/common"

type Layout struct {
    Generator string
    GeneratorVersion string
    GeneratorBuildChain string
    SIG map[string]*SIG
    Organization map[string]*Organization
    ISD map[string]*ISD
    Site map[string]*Site
    CS map[string]*CS
    BS map[string]*BS
    Host map[string]*Host
    PS map[string]*PS
    AS map[string]*AS
    Interface map[string]*Interface
    BR map[string]*BR
}

// SIG object.

type SIG struct {
    ID string `toml:"-"`
    Layout *Layout `toml:"-"`
    Name string `toml:"Name,omitempty"`
    AS *AS `toml:"-"`
    REFAS string `toml:"AS,omitempty"`
}

func (self *Layout) NewSIG(id string) *SIG {
    if self.SIG == nil {
        self.SIG = make(map[string]*SIG)
    }
    if _, ok := self.SIG[id]; ok {
        panic("SIG already exists")
    }
    n := &SIG{ID: id, Layout: self}
    self.SIG[id] = n
    return n
}


func (self *SIG) SetAS(ref *AS) error {
    if self.AS != nil {
        return common.NewBasicError("Reference is already set", nil)
    }
    self._AS(ref)
    ref._SIG(self)
    return nil
}

func (self *SIG) _AS(ref *AS) {
    self.AS = ref
}

// Organization object.

type Organization struct {
    ID string `toml:"-"`
    Layout *Layout `toml:"-"`
    Site []*Site `toml:"-"`
    REFSite []string `toml:"Site,omitempty"`
    AS []*AS `toml:"-"`
    REFAS []string `toml:"AS,omitempty"`
}

func (self *Layout) NewOrganization(id string) *Organization {
    if self.Organization == nil {
        self.Organization = make(map[string]*Organization)
    }
    if _, ok := self.Organization[id]; ok {
        panic("Organization already exists")
    }
    n := &Organization{ID: id, Layout: self}
    self.Organization[id] = n
    return n
}


func (self *Organization) AddSite(ref *Site) error {
    for _, r := range self.Site {
        if r == ref {
            return common.NewBasicError("Reference is already present", nil)
        }
    }
    self._Site(ref)
    ref._Organization(self)
    return nil
}

func (self *Organization) _Site(ref *Site) {
    if self.Site == nil {
        self.Site = make([]*Site, 0)
    }
    self.Site = append(self.Site, ref)
}

func (self *Organization) AddAS(ref *AS) error {
    for _, r := range self.AS {
        if r == ref {
            return common.NewBasicError("Reference is already present", nil)
        }
    }
    self._AS(ref)
    ref._Organization(self)
    return nil
}

func (self *Organization) _AS(ref *AS) {
    if self.AS == nil {
        self.AS = make([]*AS, 0)
    }
    self.AS = append(self.AS, ref)
}

// ISD object.

type ISD struct {
    ID string `toml:"-"`
    Layout *Layout `toml:"-"`
    Name string `toml:"Name,omitempty"`
    AS []*AS `toml:"-"`
    REFAS []string `toml:"AS,omitempty"`
}

func (self *Layout) NewISD(id string) *ISD {
    if self.ISD == nil {
        self.ISD = make(map[string]*ISD)
    }
    if _, ok := self.ISD[id]; ok {
        panic("ISD already exists")
    }
    n := &ISD{ID: id, Layout: self}
    self.ISD[id] = n
    return n
}


func (self *ISD) AddAS(ref *AS) error {
    for _, r := range self.AS {
        if r == ref {
            return common.NewBasicError("Reference is already present", nil)
        }
    }
    self._AS(ref)
    ref._ISD(self)
    return nil
}

func (self *ISD) _AS(ref *AS) {
    if self.AS == nil {
        self.AS = make([]*AS, 0)
    }
    self.AS = append(self.AS, ref)
}

// Site object.

type Site struct {
    ID string `toml:"-"`
    Layout *Layout `toml:"-"`
    Host []*Host `toml:"-"`
    REFHost []string `toml:"Host,omitempty"`
    Organization *Organization `toml:"-"`
    REFOrganization string `toml:"Organization,omitempty"`
}

func (self *Layout) NewSite(id string) *Site {
    if self.Site == nil {
        self.Site = make(map[string]*Site)
    }
    if _, ok := self.Site[id]; ok {
        panic("Site already exists")
    }
    n := &Site{ID: id, Layout: self}
    self.Site[id] = n
    return n
}


func (self *Site) AddHost(ref *Host) error {
    for _, r := range self.Host {
        if r == ref {
            return common.NewBasicError("Reference is already present", nil)
        }
    }
    self._Host(ref)
    ref._Site(self)
    return nil
}

func (self *Site) _Host(ref *Host) {
    if self.Host == nil {
        self.Host = make([]*Host, 0)
    }
    self.Host = append(self.Host, ref)
}

func (self *Site) SetOrganization(ref *Organization) error {
    if self.Organization != nil {
        return common.NewBasicError("Reference is already set", nil)
    }
    self._Organization(ref)
    ref._Site(self)
    return nil
}

func (self *Site) _Organization(ref *Organization) {
    self.Organization = ref
}

// CS object.

type CS struct {
    ID string `toml:"-"`
    Layout *Layout `toml:"-"`
    Name string `toml:"Name,omitempty"`
    AS *AS `toml:"-"`
    REFAS string `toml:"AS,omitempty"`
}

func (self *Layout) NewCS(id string) *CS {
    if self.CS == nil {
        self.CS = make(map[string]*CS)
    }
    if _, ok := self.CS[id]; ok {
        panic("CS already exists")
    }
    n := &CS{ID: id, Layout: self}
    self.CS[id] = n
    return n
}


func (self *CS) SetAS(ref *AS) error {
    if self.AS != nil {
        return common.NewBasicError("Reference is already set", nil)
    }
    self._AS(ref)
    ref._CS(self)
    return nil
}

func (self *CS) _AS(ref *AS) {
    self.AS = ref
}

// BS object.

type BS struct {
    ID string `toml:"-"`
    Layout *Layout `toml:"-"`
    Name string `toml:"Name,omitempty"`
    AS *AS `toml:"-"`
    REFAS string `toml:"AS,omitempty"`
}

func (self *Layout) NewBS(id string) *BS {
    if self.BS == nil {
        self.BS = make(map[string]*BS)
    }
    if _, ok := self.BS[id]; ok {
        panic("BS already exists")
    }
    n := &BS{ID: id, Layout: self}
    self.BS[id] = n
    return n
}


func (self *BS) SetAS(ref *AS) error {
    if self.AS != nil {
        return common.NewBasicError("Reference is already set", nil)
    }
    self._AS(ref)
    ref._BS(self)
    return nil
}

func (self *BS) _AS(ref *AS) {
    self.AS = ref
}

// Host object.

type Host struct {
    ID string `toml:"-"`
    Layout *Layout `toml:"-"`
    MachineType int `toml:"MachineType,omitempty"`
    SerialNumber string `toml:"SerialNumber,omitempty"`
    Site *Site `toml:"-"`
    REFSite string `toml:"Site,omitempty"`
    Interface []*Interface `toml:"-"`
    REFInterface []string `toml:"Interface,omitempty"`
    Location string `toml:"Location,omitempty"`
}

func (self *Layout) NewHost(id string) *Host {
    if self.Host == nil {
        self.Host = make(map[string]*Host)
    }
    if _, ok := self.Host[id]; ok {
        panic("Host already exists")
    }
    n := &Host{ID: id, Layout: self}
    self.Host[id] = n
    return n
}


func (self *Host) SetSite(ref *Site) error {
    if self.Site != nil {
        return common.NewBasicError("Reference is already set", nil)
    }
    self._Site(ref)
    ref._Host(self)
    return nil
}

func (self *Host) _Site(ref *Site) {
    self.Site = ref
}

func (self *Host) AddInterface(ref *Interface) error {
    for _, r := range self.Interface {
        if r == ref {
            return common.NewBasicError("Reference is already present", nil)
        }
    }
    self._Interface(ref)
    ref._Host(self)
    return nil
}

func (self *Host) _Interface(ref *Interface) {
    if self.Interface == nil {
        self.Interface = make([]*Interface, 0)
    }
    self.Interface = append(self.Interface, ref)
}

// PS object.

type PS struct {
    ID string `toml:"-"`
    Layout *Layout `toml:"-"`
    Name string `toml:"Name,omitempty"`
    AS *AS `toml:"-"`
    REFAS string `toml:"AS,omitempty"`
}

func (self *Layout) NewPS(id string) *PS {
    if self.PS == nil {
        self.PS = make(map[string]*PS)
    }
    if _, ok := self.PS[id]; ok {
        panic("PS already exists")
    }
    n := &PS{ID: id, Layout: self}
    self.PS[id] = n
    return n
}


func (self *PS) SetAS(ref *AS) error {
    if self.AS != nil {
        return common.NewBasicError("Reference is already set", nil)
    }
    self._AS(ref)
    ref._PS(self)
    return nil
}

func (self *PS) _AS(ref *AS) {
    self.AS = ref
}

// AS object.

type AS struct {
    ID string `toml:"-"`
    Layout *Layout `toml:"-"`
    SIG []*SIG `toml:"-"`
    REFSIG []string `toml:"SIG,omitempty"`
    Organization *Organization `toml:"-"`
    REFOrganization string `toml:"Organization,omitempty"`
    ISD []*ISD `toml:"-"`
    REFISD []string `toml:"ISD,omitempty"`
    CS []*CS `toml:"-"`
    REFCS []string `toml:"CS,omitempty"`
    BS []*BS `toml:"-"`
    REFBS []string `toml:"BS,omitempty"`
    Core bool `toml:"Core,omitempty"`
    MTU int `toml:"MTU,omitempty"`
    PS []*PS `toml:"-"`
    REFPS []string `toml:"PS,omitempty"`
    BR []*BR `toml:"-"`
    REFBR []string `toml:"BR,omitempty"`
}

func (self *Layout) NewAS(id string) *AS {
    if self.AS == nil {
        self.AS = make(map[string]*AS)
    }
    if _, ok := self.AS[id]; ok {
        panic("AS already exists")
    }
    n := &AS{ID: id, Layout: self}
    self.AS[id] = n
    return n
}


func (self *AS) AddSIG(ref *SIG) error {
    for _, r := range self.SIG {
        if r == ref {
            return common.NewBasicError("Reference is already present", nil)
        }
    }
    self._SIG(ref)
    ref._AS(self)
    return nil
}

func (self *AS) _SIG(ref *SIG) {
    if self.SIG == nil {
        self.SIG = make([]*SIG, 0)
    }
    self.SIG = append(self.SIG, ref)
}

func (self *AS) SetOrganization(ref *Organization) error {
    if self.Organization != nil {
        return common.NewBasicError("Reference is already set", nil)
    }
    self._Organization(ref)
    ref._AS(self)
    return nil
}

func (self *AS) _Organization(ref *Organization) {
    self.Organization = ref
}

func (self *AS) AddISD(ref *ISD) error {
    for _, r := range self.ISD {
        if r == ref {
            return common.NewBasicError("Reference is already present", nil)
        }
    }
    self._ISD(ref)
    ref._AS(self)
    return nil
}

func (self *AS) _ISD(ref *ISD) {
    if self.ISD == nil {
        self.ISD = make([]*ISD, 0)
    }
    self.ISD = append(self.ISD, ref)
}

func (self *AS) AddCS(ref *CS) error {
    for _, r := range self.CS {
        if r == ref {
            return common.NewBasicError("Reference is already present", nil)
        }
    }
    self._CS(ref)
    ref._AS(self)
    return nil
}

func (self *AS) _CS(ref *CS) {
    if self.CS == nil {
        self.CS = make([]*CS, 0)
    }
    self.CS = append(self.CS, ref)
}

func (self *AS) AddBS(ref *BS) error {
    for _, r := range self.BS {
        if r == ref {
            return common.NewBasicError("Reference is already present", nil)
        }
    }
    self._BS(ref)
    ref._AS(self)
    return nil
}

func (self *AS) _BS(ref *BS) {
    if self.BS == nil {
        self.BS = make([]*BS, 0)
    }
    self.BS = append(self.BS, ref)
}

func (self *AS) AddPS(ref *PS) error {
    for _, r := range self.PS {
        if r == ref {
            return common.NewBasicError("Reference is already present", nil)
        }
    }
    self._PS(ref)
    ref._AS(self)
    return nil
}

func (self *AS) _PS(ref *PS) {
    if self.PS == nil {
        self.PS = make([]*PS, 0)
    }
    self.PS = append(self.PS, ref)
}

func (self *AS) AddBR(ref *BR) error {
    for _, r := range self.BR {
        if r == ref {
            return common.NewBasicError("Reference is already present", nil)
        }
    }
    self._BR(ref)
    ref._AS(self)
    return nil
}

func (self *AS) _BR(ref *BR) {
    if self.BR == nil {
        self.BR = make([]*BR, 0)
    }
    self.BR = append(self.BR, ref)
}

// Interface object.

type Interface struct {
    ID string `toml:"-"`
    Layout *Layout `toml:"-"`
    SerialNumber string `toml:"SerialNumber,omitempty"`
    IP4MaskInt string `toml:"IP4MaskInt,omitempty"`
    IP6MaskInt string `toml:"IP6MaskInt,omitempty"`
    SIMNumber string `toml:"SIMNumber,omitempty"`
    PublicKey string `toml:"PublicKey,omitempty"`
    Host *Host `toml:"-"`
    REFHost string `toml:"Host,omitempty"`
    PhoneNumber string `toml:"PhoneNumber,omitempty"`
    PhysicalName string `toml:"PhysicalName,omitempty"`
    IP4PTP string `toml:"IP4PTP,omitempty"`
    Routes6 string `toml:"Routes6,omitempty"`
    PeerName string `toml:"PeerName,omitempty"`
    Routes4 string `toml:"Routes4,omitempty"`
    IP4MaskExt string `toml:"IP4MaskExt,omitempty"`
    IP4Peer string `toml:"IP4Peer,omitempty"`
    OSName string `toml:"OSName,omitempty"`
    IMEI string `toml:"IMEI,omitempty"`
}

func (self *Layout) NewInterface(id string) *Interface {
    if self.Interface == nil {
        self.Interface = make(map[string]*Interface)
    }
    if _, ok := self.Interface[id]; ok {
        panic("Interface already exists")
    }
    n := &Interface{ID: id, Layout: self}
    self.Interface[id] = n
    return n
}


func (self *Interface) SetHost(ref *Host) error {
    if self.Host != nil {
        return common.NewBasicError("Reference is already set", nil)
    }
    self._Host(ref)
    ref._Interface(self)
    return nil
}

func (self *Interface) _Host(ref *Host) {
    self.Host = ref
}

// BR object.

type BR struct {
    ID string `toml:"-"`
    Layout *Layout `toml:"-"`
    Name string `toml:"Name,omitempty"`
    AS *AS `toml:"-"`
    REFAS string `toml:"AS,omitempty"`
}

func (self *Layout) NewBR(id string) *BR {
    if self.BR == nil {
        self.BR = make(map[string]*BR)
    }
    if _, ok := self.BR[id]; ok {
        panic("BR already exists")
    }
    n := &BR{ID: id, Layout: self}
    self.BR[id] = n
    return n
}


func (self *BR) SetAS(ref *AS) error {
    if self.AS != nil {
        return common.NewBasicError("Reference is already set", nil)
    }
    self._AS(ref)
    ref._BR(self)
    return nil
}

func (self *BR) _AS(ref *AS) {
    self.AS = ref
}

func (self *Layout) pointersToStrings() {
    for _, v := range self.SIG {
        _ = v
        v.REFAS = v.AS.ID
    }
    for _, v := range self.Organization {
        _ = v
        v.REFSite = []string{}
        for _, p := range v.Site {
            v.REFSite = append(v.REFSite, p.ID)
        }
        v.REFAS = []string{}
        for _, p := range v.AS {
            v.REFAS = append(v.REFAS, p.ID)
        }
    }
    for _, v := range self.ISD {
        _ = v
        v.REFAS = []string{}
        for _, p := range v.AS {
            v.REFAS = append(v.REFAS, p.ID)
        }
    }
    for _, v := range self.Site {
        _ = v
        v.REFHost = []string{}
        for _, p := range v.Host {
            v.REFHost = append(v.REFHost, p.ID)
        }
        v.REFOrganization = v.Organization.ID
    }
    for _, v := range self.CS {
        _ = v
        v.REFAS = v.AS.ID
    }
    for _, v := range self.BS {
        _ = v
        v.REFAS = v.AS.ID
    }
    for _, v := range self.Host {
        _ = v
        v.REFSite = v.Site.ID
        v.REFInterface = []string{}
        for _, p := range v.Interface {
            v.REFInterface = append(v.REFInterface, p.ID)
        }
    }
    for _, v := range self.PS {
        _ = v
        v.REFAS = v.AS.ID
    }
    for _, v := range self.AS {
        _ = v
        v.REFSIG = []string{}
        for _, p := range v.SIG {
            v.REFSIG = append(v.REFSIG, p.ID)
        }
        v.REFOrganization = v.Organization.ID
        v.REFISD = []string{}
        for _, p := range v.ISD {
            v.REFISD = append(v.REFISD, p.ID)
        }
        v.REFCS = []string{}
        for _, p := range v.CS {
            v.REFCS = append(v.REFCS, p.ID)
        }
        v.REFBS = []string{}
        for _, p := range v.BS {
            v.REFBS = append(v.REFBS, p.ID)
        }
        v.REFPS = []string{}
        for _, p := range v.PS {
            v.REFPS = append(v.REFPS, p.ID)
        }
        v.REFBR = []string{}
        for _, p := range v.BR {
            v.REFBR = append(v.REFBR, p.ID)
        }
    }
    for _, v := range self.Interface {
        _ = v
        v.REFHost = v.Host.ID
    }
    for _, v := range self.BR {
        _ = v
        v.REFAS = v.AS.ID
    }
}

func (self *Layout) stringsToPointers() {
    for k, v := range self.SIG {
        v.Layout = self
        v.ID = k
        v.AS = self.AS[v.REFAS]
    }
    for k, v := range self.Organization {
        v.Layout = self
        v.ID = k
        v.Site = []*Site{}
        for _, p := range v.REFSite {
            v.Site = append(v.Site, self.Site[p])
        }
        v.AS = []*AS{}
        for _, p := range v.REFAS {
            v.AS = append(v.AS, self.AS[p])
        }
    }
    for k, v := range self.ISD {
        v.Layout = self
        v.ID = k
        v.AS = []*AS{}
        for _, p := range v.REFAS {
            v.AS = append(v.AS, self.AS[p])
        }
    }
    for k, v := range self.Site {
        v.Layout = self
        v.ID = k
        v.Host = []*Host{}
        for _, p := range v.REFHost {
            v.Host = append(v.Host, self.Host[p])
        }
        v.Organization = self.Organization[v.REFOrganization]
    }
    for k, v := range self.CS {
        v.Layout = self
        v.ID = k
        v.AS = self.AS[v.REFAS]
    }
    for k, v := range self.BS {
        v.Layout = self
        v.ID = k
        v.AS = self.AS[v.REFAS]
    }
    for k, v := range self.Host {
        v.Layout = self
        v.ID = k
        v.Site = self.Site[v.REFSite]
        v.Interface = []*Interface{}
        for _, p := range v.REFInterface {
            v.Interface = append(v.Interface, self.Interface[p])
        }
    }
    for k, v := range self.PS {
        v.Layout = self
        v.ID = k
        v.AS = self.AS[v.REFAS]
    }
    for k, v := range self.AS {
        v.Layout = self
        v.ID = k
        v.SIG = []*SIG{}
        for _, p := range v.REFSIG {
            v.SIG = append(v.SIG, self.SIG[p])
        }
        v.Organization = self.Organization[v.REFOrganization]
        v.ISD = []*ISD{}
        for _, p := range v.REFISD {
            v.ISD = append(v.ISD, self.ISD[p])
        }
        v.CS = []*CS{}
        for _, p := range v.REFCS {
            v.CS = append(v.CS, self.CS[p])
        }
        v.BS = []*BS{}
        for _, p := range v.REFBS {
            v.BS = append(v.BS, self.BS[p])
        }
        v.PS = []*PS{}
        for _, p := range v.REFPS {
            v.PS = append(v.PS, self.PS[p])
        }
        v.BR = []*BR{}
        for _, p := range v.REFBR {
            v.BR = append(v.BR, self.BR[p])
        }
    }
    for k, v := range self.Interface {
        v.Layout = self
        v.ID = k
        v.Host = self.Host[v.REFHost]
    }
    for k, v := range self.BR {
        v.Layout = self
        v.ID = k
        v.AS = self.AS[v.REFAS]
    }
}