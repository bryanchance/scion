// Copyright 2018 Anapaya Systems

package db

import (
	"database/sql"
	"fmt"
	"net"
	"strconv"
	"strings"

	_ "github.com/mattn/go-sqlite3"

	"github.com/scionproto/scion/go/lib/addr"
	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/lib/pathmgr"
	"github.com/scionproto/scion/go/lib/pktcls"
	"github.com/scionproto/scion/go/lib/sqlite"
	"github.com/scionproto/scion/go/sig/config"
	"github.com/scionproto/scion/go/sig/mgmt"
	"github.com/scionproto/scion/go/sig/siginfo"
)

const (
	SchemaVersion = 1
	Schema        = `
	CREATE TABLE Sites (
		Name TEXT PRIMARY KEY NOT NULL,
		MetricsPort INTEGER NOT NULL,
		VHost TEXT NOT NULL
	);

	CREATE TABLE Hosts (
		Name TEXT NOT NULL,
		Site TEXT REFERENCES Sites ON DELETE CASCADE ON UPDATE CASCADE,
		PRIMARY KEY (Site, Name)
	);

	CREATE TABLE ASEntries (
		IsdID INTEGER NOT NULL,
		AsID INTEGER NOT NULL,
		Site TEXT REFERENCES Sites ON DELETE CASCADE ON UPDATE CASCADE,
		Policy TEXT NOT NULL,
		PRIMARY KEY (Site, IsdID, AsID)
	);

	CREATE TABLE SIGs (
		Name TEXT NOT NULL,
		Address TEXT NOT NULL,
		CtrlPort INTEGER NOT NULL,
		EncapPort INTEGER NOT NULL,
		Site TEXT REFERENCES Sites,
		IsdID INTEGER NOT NULL,
		AsID INTEGER NOT NULL,
		FOREIGN KEY (Site, IsdID, AsID) REFERENCES ASEntries ON DELETE CASCADE ON UPDATE CASCADE,
		PRIMARY KEY (Site, IsdID, AsID, Name)
	);

	CREATE TABLE Networks (
		CIDR TEXT NOT NULL,
		Site TEXT REFERENCES Sites ON DELETE CASCADE ON UPDATE CASCADE,
		IsdID INTEGER NOT NULL,
		AsID INTEGER NOT NULL,
		FOREIGN KEY (Site, IsdID, AsID) REFERENCES ASEntries ON DELETE CASCADE ON UPDATE CASCADE,
		PRIMARY KEY (Site, IsdID, AsID, CIDR)
	);

	CREATE TABLE Filters (
		Name TEXT NOT NULL,
		Filter TEXT NOT NULL,
		Site TEXT REFERENCES Sites ON DELETE CASCADE ON UPDATE CASCADE,
		PRIMARY KEY (Site, Name)
	);

	CREATE TABLE Sessions (
		Name INTEGER NOT NULL,
		Site TEXT REFERENCES Sites ON DELETE CASCADE ON UPDATE CASCADE,
		IsdID INTEGER NOT NULL,
		AsID INTEGER NOT NULL,
		FilterName TEXT NOT NULL,
		FOREIGN KEY (Site, FilterName) REFERENCES Filters ON DELETE CASCADE ON UPDATE CASCADE,
		FOREIGN KEY (Site, IsdID, AsID) REFERENCES ASEntries ON DELETE CASCADE ON UPDATE CASCADE,
		PRIMARY KEY (Site, IsdID, AsID, Name)
	);

	CREATE TABLE SessionAliases (
		Name TEXT NOT NULL,
		Sessions TEXT NOT NULL,
		Site TEXT REFERENCES Sites ON DELETE CASCADE ON UPDATE CASCADE,
		IsdID INTEGER NOT NULL,
		AsID INTEGER NOT NULL,
		FOREIGN KEY (Site, IsdID, AsID) REFERENCES ASEntries ON DELETE CASCADE ON UPDATE CASCADE,
		PRIMARY KEY (Site, IsdID, AsID, Name)
	);
	`
)

type DB struct {
	*sql.DB
}

func New(path string) (*DB, error) {
	if db, err := sqlite.New(path, Schema, SchemaVersion); err != nil {
		return nil, err
	} else {
		return &DB{DB: db}, nil
	}
}

func (db *DB) InsertSite(site *Site) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	_, err = tx.Exec(`INSERT INTO Sites (Name, VHost, MetricsPort) VALUES (?, ?, ?)`,
		site.Name, site.VHost, site.MetricsPort)
	if err != nil {
		rerr := tx.Rollback()
		return common.NewBasicError("database transaction error", err, "rollback_error", rerr)
	}
	for _, host := range site.Hosts {
		_, err := tx.Exec(`INSERT INTO Hosts (Name, Site) VALUES (?, ?)`, host, site.Name)
		if err != nil {
			rerr := tx.Rollback()
			return common.NewBasicError("database transaction error", err, "rollback_error", rerr)
		}
	}
	return tx.Commit()
}

func (db *DB) DeleteSite(name string) error {
	_, err := db.Exec(`DELETE FROM Sites WHERE Name = ?`, name)
	return err
}

func (db *DB) GetSite(name string) (*Site, error) {
	rows, err := db.Query(`
		SELECT Sites.Name, Sites.VHost, Sites.MetricsPort, Hosts.Name
		FROM Sites
		INNER JOIN Hosts
		WHERE Hosts.Site = Sites.Name AND Sites.Name = ?`, name)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	site := Site{}
	for rows.Next() {
		var host string
		if err := rows.Scan(&site.Name, &site.VHost, &site.MetricsPort, &host); err != nil {
			return nil, err
		}
		site.Hosts = append(site.Hosts, host)
	}
	return &site, rows.Err()
}

func (db *DB) QuerySites() (map[string]*Site, error) {
	rows, err := db.Query(`
		SELECT Sites.Name, Sites.VHost, Sites.MetricsPort, Hosts.Name
		FROM Sites
		INNER JOIN Hosts
		WHERE Hosts.Site = Sites.Name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	sites := make(map[string]*Site)
	for rows.Next() {
		var siteStr, vHostStr, hostStr string
		var metricsPort uint16
		if err := rows.Scan(&siteStr, &vHostStr, &metricsPort, &hostStr); err != nil {
			return nil, err
		}
		if _, ok := sites[siteStr]; !ok {
			sites[siteStr] = &Site{Name: siteStr, VHost: vHostStr, MetricsPort: metricsPort}
		}
		sites[siteStr].Hosts = append(sites[siteStr].Hosts, hostStr)
	}
	return sites, rows.Err()
}

func (db *DB) InsertAS(site string, ia addr.ISD_AS) error {
	// Insert an empty string in the policy field to not have the hassle of
	// dealing with NULLs
	_, err := db.Exec(
		`INSERT INTO ASEntries (IsdID, AsID, Site, Policy) VALUES(?, ?, ?, ?)`,
		ia.I, ia.A, site, "")
	return err
}

func (db *DB) DeleteAS(site string, ia addr.ISD_AS) error {
	_, err := db.Exec(
		`DELETE FROM ASEntries WHERE IsdID = ? AND AsID = ? AND Site = ?`,
		ia.I, ia.A, site)
	return err
}

func (db *DB) QueryASes(site string) ([]addr.ISD_AS, error) {
	rows, err := db.Query(`SELECT IsdID, AsID FROM ASEntries WHERE Site = ?`, site)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ases []addr.ISD_AS
	for rows.Next() {
		ia := addr.ISD_AS{}
		err := rows.Scan(&ia.I, &ia.A)
		if err != nil {
			return nil, err
		}
		ases = append(ases, ia)
	}
	return ases, rows.Err()
}

func (db *DB) SetPolicy(site string, ia addr.ISD_AS, policy string) error {
	_, err := db.Exec(
		`UPDATE ASEntries SET Policy = ?  WHERE Site = ? AND IsdID = ? AND AsID = ?`,
		policy, site, ia.I, ia.A)
	return err
}

func (db *DB) GetPolicy(site string, ia addr.ISD_AS) (string, error) {
	var policy string
	err := db.QueryRow(
		`SELECT Policy FROM ASEntries WHERE Site = ? AND IsdID = ? AND AsID = ?`,
		site, ia.I, ia.A).Scan(&policy)
	if err != nil {
		return "", err
	}
	return policy, nil
}

func (db *DB) GetPolicies(site string) (map[addr.ISD_AS]string, error) {
	rows, err := db.Query(`SELECT IsdID, AsID, Policy FROM ASEntries WHERE Site = ?`, site)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	policies := make(map[addr.ISD_AS]string)
	for rows.Next() {
		var isd, as int
		var policy string
		if err := rows.Scan(&isd, &as, &policy); err != nil {
			return nil, err
		}
		policies[addr.ISD_AS{I: isd, A: as}] = policy
	}
	return policies, rows.Err()
}

func (db *DB) InsertSIG(site string, ia addr.ISD_AS, sig *config.SIG) error {
	_, err := db.Exec(
		`INSERT INTO SIGs (Name, Address, CtrlPort, EncapPort, IsdID, AsID, Site)
			VALUES(?, ?, ?, ?, ?, ?, ?)`,
		sig.Id, sig.Addr.String(), sig.CtrlPort, sig.EncapPort, ia.I, ia.A, site)
	return err
}

func (db *DB) DeleteSIG(site string, ia addr.ISD_AS, sigName string) error {
	_, err := db.Exec(
		`DELETE FROM SIGs WHERE Name = ? AND IsdID = ? AND AsID = ? AND Site = ?`,
		sigName, ia.I, ia.A, site)
	return err
}

func (db *DB) QuerySIGs(site string, ia addr.ISD_AS) (config.SIGSet, error) {
	rows, err := db.Query(
		`SELECT Name, Address, CtrlPort, EncapPort FROM SIGs
			WHERE Site = ? AND IsdID = ? AND AsID = ?`,
		site, ia.I, ia.A)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	sigs := make(config.SIGSet)
	for rows.Next() {
		var name, address string
		var ctrlPort, encapPort int
		rows.Scan(&name, &address, &ctrlPort, &encapPort)
		sig := &config.SIG{
			Id:        siginfo.SigIdType(name),
			Addr:      net.ParseIP(address),
			CtrlPort:  uint16(ctrlPort),
			EncapPort: uint16(encapPort),
		}
		if sig.Addr == nil {
			// Bug in pre-insertion validation
			panic(fmt.Sprintf("bad IP address in db: %s", address))
		}
		sigs[sig.Id] = sig
	}
	return sigs, rows.Err()
}

func (db *DB) InsertNetwork(site string, ia addr.ISD_AS, network *net.IPNet) error {
	_, err := db.Exec(
		`INSERT INTO Networks (CIDR, Site, IsdID, AsID) VALUES (?, ?, ?, ?)`,
		network.String(), site, ia.I, ia.A)
	return err
}

func (db *DB) DeleteNetwork(site string, ia addr.ISD_AS, network string) error {
	_, err := db.Exec(
		`DELETE FROM Networks WHERE CIDR = ? AND IsdID = ? AND AsID = ? AND Site = ?`,
		network, ia.I, ia.A, site)
	return err
}

func (db *DB) QueryNetworks(site string, ia addr.ISD_AS) ([]*config.IPNet, error) {
	rows, err := db.Query(
		`SELECT CIDR FROM Networks WHERE site = ? AND IsdID = ? AND AsID = ?`,
		site, ia.I, ia.A)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var networks []*config.IPNet
	for rows.Next() {
		var cidr string
		if err := rows.Scan(&cidr); err != nil {
			return nil, err
		}
		_, network, err := net.ParseCIDR(cidr)
		if err != nil {
			// Bug in pre-insertion validation
			panic(fmt.Sprintf("bad IP network address in db: %s", cidr))
		}
		networks = append(networks, (*config.IPNet)(network))
	}
	return networks, rows.Err()
}

func (db *DB) InsertFilter(site string, name string, pp *pathmgr.PathPredicate) error {
	_, err := db.Exec(
		`INSERT INTO Filters (Name, Filter, Site) VALUES (?, ?, ?)`,
		name, pp.String(), site)
	return err
}

func (db *DB) DeleteFilter(site string, name string) error {
	_, err := db.Exec(`DELETE FROM Filters WHERE Name = ? AND Site = ?`, name, site)
	return err
}

func (db *DB) QueryFilters(site string) (pktcls.ActionMap, error) {
	rows, err := db.Query(`SELECT Name, Filter FROM Filters WHERE Site = ?`, site)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	actionMap := make(pktcls.ActionMap)
	for rows.Next() {
		var name, ppString string
		if err := rows.Scan(&name, &ppString); err != nil {
			return nil, err
		}
		pp, err := pathmgr.NewPathPredicate(ppString)
		if err != nil {
			// Bug in pre-insertion validation
			panic(fmt.Sprintf("bad filter string in db: %s", ppString))
		}
		actionMap[name] = pktcls.NewActionFilterPaths(name, pp)
	}
	return actionMap, rows.Err()
}

func (db *DB) InsertSession(site string, ia addr.ISD_AS, name uint8, filter string) error {
	_, err := db.Exec(
		`INSERT INTO Sessions (Name, FilterName, Site, IsdID, AsID) VALUES (?, ?, ?, ?, ?)`,
		name, filter, site, ia.I, ia.A)
	return err
}

func (db *DB) DeleteSession(site string, ia addr.ISD_AS, name uint8) error {
	_, err := db.Exec(
		`DELETE FROM Sessions WHERE Name = ? AND Site = ? AND IsdID = ? AND AsID = ?`,
		name, site, ia.I, ia.A)
	return err
}

func (db *DB) QuerySessions(site string, ia addr.ISD_AS) (config.SessionMap, error) {
	rows, err := db.Query(
		`SELECT Name, FilterName FROM Sessions WHERE Site = ? AND IsdID = ? AND AsID = ?`,
		site, ia.I, ia.A)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	sessionMap := make(config.SessionMap)
	for rows.Next() {
		var name int
		var filter string
		if err := rows.Scan(&name, &filter); err != nil {
			return nil, err
		}
		sessionMap[mgmt.SessionType(name)] = filter
	}
	return sessionMap, rows.Err()
}

func (db *DB) InsertSessionAlias(site string, ia addr.ISD_AS, name string,
	sessions []uint8) error {

	_, err := db.Exec(
		`INSERT INTO SessionAliases (Name, Sessions, Site, IsdID, AsID) VALUES (?, ?, ?, ?, ?)`,
		name, strings.Join(applyConvertUInt8(sessions), ","), site, ia.I, ia.A)
	return err
}

func (db *DB) DeleteSessionAlias(site string, ia addr.ISD_AS, name string) error {
	_, err := db.Exec(
		`DELETE FROM SessionAliases WHERE Name = ? AND Site = ? AND IsdID = ? AND AsID = ?`,
		name, site, ia.I, ia.A)
	return err
}

func (db *DB) GetSessionAliases(site string) (map[addr.ISD_AS]SessionAliasMap, error) {
	rows, err := db.Query(
		`SELECT Name, Sessions, IsdID, AsID FROM SessionAliases WHERE Site = ?`, site)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	sessionAliases := make(map[addr.ISD_AS]SessionAliasMap)
	for rows.Next() {
		var aliasName, sessionList string
		var isd, as int
		if err := rows.Scan(&aliasName, &sessionList, &isd, &as); err != nil {
			return nil, err
		}
		ia := addr.ISD_AS{I: isd, A: as}
		_, ok := sessionAliases[ia]
		if !ok {
			sessionAliases[ia] = make(SessionAliasMap)
		}
		sessionAliases[ia][aliasName] = sessionList
	}
	return sessionAliases, nil
}

func (db *DB) GetSiteConfig(site string) (*config.Cfg, error) {
	cfg := &config.Cfg{}
	ases, err := db.QueryASes(site)
	if err != nil {
		return nil, err
	}
	cfg.ASes = make(map[addr.ISD_AS]*config.ASEntry)
	for _, ia := range ases {
		asEntry, err := db.GetASDetails(site, ia)
		if err != nil {
			return nil, err
		}
		cfg.ASes[ia] = asEntry
	}
	actions, err := db.QueryFilters(site)
	if err != nil {
		return nil, err
	}
	cfg.Actions = actions
	return cfg, nil
}

func (db *DB) GetASDetails(site string, ia addr.ISD_AS) (*config.ASEntry, error) {
	networks, err := db.QueryNetworks(site, ia)
	if err != nil {
		return nil, err
	}
	sigs, err := db.QuerySIGs(site, ia)
	if err != nil {
		return nil, err
	}
	sessions, err := db.QuerySessions(site, ia)
	if err != nil {
		return nil, err
	}
	return &config.ASEntry{
		Nets:     networks,
		Sigs:     sigs,
		Sessions: sessions,
	}, nil
}

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
	Hosts       []string
	MetricsPort uint16
}

func (s *Site) PrintHosts() string {
	return strings.Join(s.Hosts, ", ")
}
