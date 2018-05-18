// Copyright 2018 Anapaya Systems

package db

import (
	"database/sql"
	"errors"
	"fmt"
	"net"

	_ "github.com/mattn/go-sqlite3"

	"github.com/scionproto/scion/go/lib/addr"
	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/lib/pktcls"
	"github.com/scionproto/scion/go/lib/spath/spathmeta"
	"github.com/scionproto/scion/go/lib/sqlite"
	"github.com/scionproto/scion/go/sig/config"
	"github.com/scionproto/scion/go/sig/siginfo"
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

func (db *DB) GetConfigValue(site string, ia addr.IA, name string) (*ASConfig, error) {
	rows, err := db.Query(
		`SELECT Name, Value FROM ASConfig WHERE Name = ? AND IsdID = ? AND AsID = ? AND Site = ?`,
		name, ia.I, ia.A, site)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	config := ASConfig{}
	for rows.Next() {
		if err := rows.Scan(&config.Name, &config.Value); err != nil {
			return nil, err
		}
	}
	return &config, rows.Err()
}

func (db *DB) SetConfigValue(site string, ia addr.IA, name string, value string) error {
	_, err := db.Exec(
		`REPLACE INTO ASConfig (Name, Value, IsdID, AsID, Site) VALUES(?, ?, ?, ?, ?)`,
		name, value, ia.I, ia.A, site)
	return err
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

func (db *DB) InsertAS(site string, name string, ia addr.IA) error {
	// Insert an empty string in the policy field to not have the hassle of
	// dealing with NULLs
	_, err := db.Exec(
		`INSERT INTO ASEntries (Name, IsdID, AsID, Site, Policy) VALUES(?, ?, ?, ?, ?)`,
		name, ia.I, ia.A, site, "")
	return err
}

func (db *DB) UpdateAS(site string, name string, ia addr.IA) error {
	_, err := db.Exec(
		`UPDATE ASEntries SET Name = ?  WHERE Site = ? AND IsdID = ? AND AsID = ?`,
		name, site, ia.I, ia.A)
	return err
}

func (db *DB) DeleteAS(site string, ia addr.IA) error {
	_, err := db.Exec(
		`DELETE FROM ASEntries WHERE IsdID = ? AND AsID = ? AND Site = ?`,
		ia.I, ia.A, site)
	return err
}

func (db *DB) QueryASes(site string) ([]AS, error) {
	rows, err := db.Query(`SELECT Name, IsdID, AsID FROM ASEntries WHERE Site = ?`, site)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ases := []AS{}
	for rows.Next() {
		var name string
		ia := addr.IA{}
		err := rows.Scan(&name, &ia.I, &ia.A)
		if err != nil {
			return nil, err
		}
		as := *ASFromAddrIA(ia)
		as.Name = name
		ases = append(ases, as)
	}
	return ases, rows.Err()
}

func (db *DB) QueryIAs(site string) ([]addr.IA, error) {
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

func (db *DB) GetAS(site string, ia addr.IA) (*AS, error) {
	as := ASFromAddrIA(ia)
	err := db.QueryRow(
		`SELECT Name, Policy FROM ASEntries WHERE Site = ? AND IsdID = ? AND AsID = ?`,
		site, ia.I, ia.A).Scan(&as.Name, &as.Policy)
	if err != nil {
		return nil, err
	}
	return as, nil
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

func (db *DB) UpdateFilter(site string, name string, pp *spathmeta.PathPredicate) error {
	_, err := db.Exec(
		`UPDATE Filters SET Filter = ? WHERE Name = ? AND Site = ?`,
		pp.String(), name, site)
	return err
}

func (db *DB) DeleteFilter(site string, name string) error {
	if name == "any" {
		return errors.New("Cannot delete default path selector!")
	}
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

func (db *DB) GetSiteConfig(site string) (*config.Cfg, error) {
	cfg := &config.Cfg{}
	ias, err := db.QueryIAs(site)
	if err != nil {
		return nil, err
	}
	cfg.ASes = make(map[addr.IA]*config.ASEntry)
	for _, ia := range ias {
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
	return &config.ASEntry{
		Nets: networks,
		Sigs: sigs,
	}, nil
}
