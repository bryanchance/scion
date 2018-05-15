// Copyright 2018 Anapaya Systems

package db

import (
	"database/sql"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"

	_ "github.com/mattn/go-sqlite3"

	"github.com/scionproto/scion/go/lib/addr"
	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/lib/pktcls"
	"github.com/scionproto/scion/go/lib/spath/spathmeta"
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
		User TEXT NOT NULL,
		Key TEXT NOT NULL,
		Site TEXT REFERENCES Sites ON DELETE CASCADE ON UPDATE CASCADE,
		PRIMARY KEY (Site, Name, User, Key)
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
		ID INTEGER PRIMARY KEY AUTOINCREMENT,
		CIDR TEXT NOT NULL,
		Site TEXT REFERENCES Sites ON DELETE CASCADE ON UPDATE CASCADE,
		IsdID INTEGER NOT NULL,
		AsID INTEGER NOT NULL,
		FOREIGN KEY (Site, IsdID, AsID) REFERENCES ASEntries ON DELETE CASCADE ON UPDATE CASCADE,
		UNIQUE (ID, Site, IsdID, AsID, CIDR)
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
		_, err := tx.Exec(`INSERT INTO Hosts (Name, User, Key, Site) VALUES (?, ?, ?, ?)`,
			host.Name, host.User, host.Key, site.Name)
		if err != nil {
			rerr := tx.Rollback()
			return common.NewBasicError("database transaction error", err, "rollback_error", rerr)
		}
	}
	return tx.Commit()
}

func (db *DB) UpdateSite(site *Site) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	_, err = tx.Exec(`UPDATE Sites SET VHost = ?, MetricsPort = ? WHERE Name = ?`,
		site.VHost, site.MetricsPort, site.Name)
	if err != nil {
		rerr := tx.Rollback()
		return common.NewBasicError("database transaction error", err, "rollback_error", rerr)
	}
	// Remove all hosts and add them again
	_, err = tx.Exec(`DELETE FROM Hosts WHERE Site = ?`, site.Name)
	if err != nil {
		rerr := tx.Rollback()
		return common.NewBasicError("database transaction error", err, "rollback_error", rerr)
	}
	for _, host := range site.Hosts {
		_, err := tx.Exec(`INSERT INTO Hosts (Name, User, Key, Site) VALUES (?, ?, ?, ?)`,
			host.Name, host.User, host.Key, site.Name)
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

func (db *DB) GetSiteWithHosts(name string) (*Site, error) {
	rows, err := db.Query(`
		SELECT Sites.Name, Sites.VHost, Sites.MetricsPort, Hosts.Name, Hosts.User, Hosts.Key
		FROM Sites
		INNER JOIN Hosts
		WHERE Hosts.Site = Sites.Name AND Sites.Name = ?`, name)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	site := Site{}
	for rows.Next() {
		host := Host{}
		if err := rows.Scan(&site.Name, &site.VHost, &site.MetricsPort,
			&host.Name, &host.User, &host.Key); err != nil {
			return nil, err
		}
		site.Hosts = append(site.Hosts, host)
	}
	// Check if result is present
	if site.Name == "" {
		return nil, errors.New("Site not found")
	}
	return &site, rows.Err()
}

func (db *DB) GetSite(name string) (*Site, error) {
	rows, err := db.Query(`
		SELECT Sites.Name, Sites.VHost, Sites.MetricsPort
		FROM Sites
		WHERE Sites.Name = ?`, name)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	site := Site{Hosts: []Host{}}
	for rows.Next() {
		if err := rows.Scan(&site.Name, &site.VHost, &site.MetricsPort); err != nil {
			return nil, err
		}
	}
	// Check if result is present
	if site.Name == "" {
		return nil, errors.New("Site not found")
	}
	return &site, rows.Err()
}

func (db *DB) QuerySites() ([]*Site, error) {
	rows, err := db.Query(`
		SELECT Sites.Name, Sites.VHost, Sites.MetricsPort
		FROM Sites`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	sites := []*Site{}
	for rows.Next() {
		var siteStr, vHostStr string
		var metricsPort uint16
		if err := rows.Scan(&siteStr, &vHostStr, &metricsPort); err != nil {
			return nil, err
		}
		sites = append(sites, &Site{Name: siteStr, VHost: vHostStr, MetricsPort: metricsPort})
	}
	return sites, rows.Err()
}

func (db *DB) InsertAS(site string, ia addr.IA) error {
	// Insert an empty string in the policy field to not have the hassle of
	// dealing with NULLs
	_, err := db.Exec(
		`INSERT INTO ASEntries (IsdID, AsID, Site, Policy) VALUES(?, ?, ?, ?)`,
		ia.I, ia.A, site, "")
	return err
}

func (db *DB) DeleteAS(site string, ia addr.IA) error {
	_, err := db.Exec(
		`DELETE FROM ASEntries WHERE IsdID = ? AND AsID = ? AND Site = ?`,
		ia.I, ia.A, site)
	return err
}

func (db *DB) QueryASes(site string) ([]addr.IA, error) {
	rows, err := db.Query(`SELECT IsdID, AsID FROM ASEntries WHERE Site = ?`, site)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ases []addr.IA
	for rows.Next() {
		ia := addr.IA{}
		err := rows.Scan(&ia.I, &ia.A)
		if err != nil {
			return nil, err
		}
		ases = append(ases, ia)
	}
	return ases, rows.Err()
}

func (db *DB) SetPolicy(site string, ia addr.IA, policy string) error {
	_, err := db.Exec(
		`UPDATE ASEntries SET Policy = ?  WHERE Site = ? AND IsdID = ? AND AsID = ?`,
		policy, site, ia.I, ia.A)
	return err
}

func (db *DB) GetPolicy(site string, ia addr.IA) (string, error) {
	var policy string
	err := db.QueryRow(
		`SELECT Policy FROM ASEntries WHERE Site = ? AND IsdID = ? AND AsID = ?`,
		site, ia.I, ia.A).Scan(&policy)
	if err != nil {
		return "", err
	}
	return policy, nil
}

func (db *DB) GetPolicies(site string) (map[addr.IA]string, error) {
	rows, err := db.Query(`SELECT IsdID, AsID, Policy FROM ASEntries WHERE Site = ?`, site)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	policies := make(map[addr.IA]string)
	for rows.Next() {
		var ia addr.IA
		var policy string
		if err := rows.Scan(&ia.I, &ia.A, &policy); err != nil {
			return nil, err
		}
		policies[ia] = policy
	}
	return policies, rows.Err()
}

func (db *DB) InsertSIG(site string, ia addr.IA, sig *config.SIG) error {
	_, err := db.Exec(
		`INSERT INTO SIGs (Name, Address, CtrlPort, EncapPort, IsdID, AsID, Site)
			VALUES(?, ?, ?, ?, ?, ?, ?)`,
		sig.Id, sig.Addr.String(), sig.CtrlPort, sig.EncapPort, ia.I, ia.A, site)
	return err
}

func (db *DB) UpdateSIG(site string, ia addr.IA, sig *config.SIG) error {
	_, err := db.Exec(
		`UPDATE SIGs SET Address = ?, CtrlPort = ?, EncapPort = ?
		WHERE Name = ? AND IsdID = ? and AsID = ? and Site = ?`,
		sig.Addr.String(), sig.CtrlPort, sig.EncapPort, sig.Id, ia.I, ia.A, site)
	return err
}

func (db *DB) DeleteSIG(site string, ia addr.IA, sigName string) error {
	_, err := db.Exec(
		`DELETE FROM SIGs WHERE Name = ? AND IsdID = ? AND AsID = ? AND Site = ?`,
		sigName, ia.I, ia.A, site)
	return err
}

func (db *DB) QuerySIGs(site string, ia addr.IA) (config.SIGSet, error) {
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

func (db *DB) InsertNetwork(site string, ia addr.IA, network *net.IPNet) error {
	_, err := db.Exec(
		`INSERT INTO Networks (CIDR, Site, IsdID, AsID) VALUES (?, ?, ?, ?)`,
		network.String(), site, ia.I, ia.A)
	return err
}

func (db *DB) DeleteNetwork(site string, ia addr.IA, network int) error {
	_, err := db.Exec(
		`DELETE FROM Networks WHERE ID = ? AND IsdID = ? AND AsID = ? AND Site = ?`,
		network, ia.I, ia.A, site)
	return err
}

func (db *DB) QueryNetworks(site string, ia addr.IA) (map[int]*config.IPNet, error) {
	rows, err := db.Query(
		`SELECT ID, CIDR FROM Networks WHERE site = ? AND IsdID = ? AND AsID = ?`,
		site, ia.I, ia.A)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	networks := map[int]*config.IPNet{}
	for rows.Next() {
		var id int
		var cidr string
		if err := rows.Scan(&id, &cidr); err != nil {
			return nil, err
		}
		_, network, err := net.ParseCIDR(cidr)
		if err != nil {
			// Bug in pre-insertion validation
			panic(fmt.Sprintf("bad IP network address in db: %s", cidr))
		}
		networks[id] = (*config.IPNet)(network)
	}
	return networks, rows.Err()
}

func (db *DB) InsertFilter(site string, name string, pp *spathmeta.PathPredicate) error {
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
		pp, err := spathmeta.NewPathPredicate(ppString)
		if err != nil {
			// Bug in pre-insertion validation
			panic(fmt.Sprintf("bad filter string in db: %s", ppString))
		}
		actionMap[name] = pktcls.NewActionFilterPaths(name, pktcls.NewCondPathPredicate(pp))
	}
	return actionMap, rows.Err()
}

func (db *DB) QueryRawFilters(site string) ([]Filter, error) {
	rows, err := db.Query(`SELECT Name, Filter FROM Filters WHERE Site = ?`, site)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	filters := []Filter{}
	for rows.Next() {
		var filter Filter
		if err := rows.Scan(&filter.Name, &filter.PP); err != nil {
			return nil, err
		}
		filters = append(filters, filter)
	}
	return filters, rows.Err()
}

func (db *DB) InsertSession(site string, ia addr.IA, name uint8, filter string) error {
	_, err := db.Exec(
		`INSERT INTO Sessions (Name, FilterName, Site, IsdID, AsID) VALUES (?, ?, ?, ?, ?)`,
		name, filter, site, ia.I, ia.A)
	return err
}

func (db *DB) DeleteSession(site string, ia addr.IA, name uint8) error {
	_, err := db.Exec(
		`DELETE FROM Sessions WHERE Name = ? AND Site = ? AND IsdID = ? AND AsID = ?`,
		name, site, ia.I, ia.A)
	return err
}

func (db *DB) QuerySessions(site string, ia addr.IA) (config.SessionMap, error) {
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

func (db *DB) InsertSessionAlias(site string, ia addr.IA, name string,
	sessions []uint8) error {

	_, err := db.Exec(
		`INSERT INTO SessionAliases (Name, Sessions, Site, IsdID, AsID) VALUES (?, ?, ?, ?, ?)`,
		name, strings.Join(applyConvertUInt8(sessions), ","), site, ia.I, ia.A)
	return err
}

func (db *DB) DeleteSessionAlias(site string, ia addr.IA, name string) error {
	_, err := db.Exec(
		`DELETE FROM SessionAliases WHERE Name = ? AND Site = ? AND IsdID = ? AND AsID = ?`,
		name, site, ia.I, ia.A)
	return err
}

func (db *DB) GetSessionAliases(site string) (map[addr.IA]SessionAliasMap, error) {
	rows, err := db.Query(
		`SELECT Name, Sessions, IsdID, AsID FROM SessionAliases WHERE Site = ?`, site)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	sessionAliases := make(map[addr.IA]SessionAliasMap)
	for rows.Next() {
		var aliasName, sessionList string
		var ia addr.IA
		if err := rows.Scan(&aliasName, &sessionList, &ia.I, &ia.A); err != nil {
			return nil, err
		}
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
	cfg.ASes = make(map[addr.IA]*config.ASEntry)
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

func (db *DB) GetASDetails(site string, ia addr.IA) (*config.ASEntry, error) {
	networkMap, err := db.QueryNetworks(site, ia)
	if err != nil {
		return nil, err
	}
	networks := []*config.IPNet{}
	for _, network := range networkMap {
		networks = append(networks, network)
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
