// Copyright 2018 Anapaya Systems

// IPScraper asks peer ASes (as configured via sigmgmt) about the IP ranges they want to export
// to the local AS. After receiving the replies it writes the results into sigmgmt database.
// Sigmgmt then pushes the new IP allocations to the production.
package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"

	"github.com/scionproto/scion/go/ipscraper/internal/config"
	"github.com/scionproto/scion/go/lib/addr"
	"github.com/scionproto/scion/go/lib/allocations"
	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/lib/env"
	"github.com/scionproto/scion/go/lib/log"
	"github.com/scionproto/scion/go/lib/snet"
	"github.com/scionproto/scion/go/lib/sock/reliable"
	"github.com/scionproto/scion/go/proto"
	"github.com/scionproto/scion/go/sigmgmt/db"
)

var (
	cfg config.Config
)

type scrapeEntry struct {
	IA          addr.IA
	IAID        uint
	IP          net.IP
	Port        uint16
	Allocations []*allocations.Allocation
}

type scrapeEntries map[string]*scrapeEntry

func main() {
	os.Exit(realMain())
}

func realMain() int {
	env.AddFlags()
	flag.Parse()
	if returnCode, ok := env.CheckFlags(config.Sample); !ok {
		return returnCode
	}
	if _, err := toml.DecodeFile(env.ConfigFile(), &cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Cannot decode TOML config: %s\n", err)
		return 1
	}

	// Initialize logging.
	if err := env.InitLogging(&cfg.Logging); err != nil {
		fmt.Fprintf(os.Stderr, "Cannot initialize logging: %s\n", err)
		return 1
	}
	defer log.Flush()
	defer log.LogPanicAndExit()
	env.LogAppStarted("IPScraper", cfg.IPScraper.LocalIA.String())
	defer env.LogAppStopped("IPScraper", cfg.IPScraper.LocalIA.String())

	// Open a database connection. The database is shared with sigmgmt.
	dbase, err := gorm.Open("sqlite3", cfg.IPScraper.DBPath)
	if err != nil {
		log.Crit("Unable to connect to database", "err", err)
		return 1
	}
	defer dbase.Close()
	db.MigrateDB(dbase)

	// Initialize SCION connection.
	scionConn, err := initSCIONConnection()
	if err != nil {
		log.Crit("Cannot open SCION connection", "err", err)
		return 1
	}

	// Get the remote IPProvider addresses from the sigmgmt database.
	var asEntries []*db.ASEntry
	if err := dbase.Find(&asEntries).Error; err != nil {
		log.Crit("Can't read AS entries from the DB", "err", err)
		return 1
	}

	// Send scraping requests to remote ASes.
	entries, err := prepareScrapingRequests(asEntries)
	if err != nil {
		log.Crit("Cannot prepare scraping requests", "err", err)
		return 1
	}
	for id, entry := range entries {
		sendScrapingRequest(context.Background(), scionConn, id, entry)
	}

	// Wait for replies.
	entries, err = receiveScrapingReplies(scionConn, entries)
	if err != nil {
		log.Crit("Cannot receive replies", "err", err)
		return 1
	}

	// Write the scraping results into the sigmgmt database.
	err = writeRepliesToDB(dbase, entries)
	if err != nil {
		log.Crit("Cannot write results to database", "err", err)
		return 1
	}

	return 0
}

func initSCIONConnection() (*snet.SCIONConn, error) {
	if err := snet.Init(cfg.IPScraper.LocalIA, cfg.Sciond.Path,
		reliable.NewDispatcherService(reliable.DefaultDispPath)); err != nil {

		return nil, err
	}
	snetConn, err := snet.ListenSCION("udp4", &snet.Addr{
		IA: cfg.IPScraper.LocalIA,
		Host: &addr.AppAddr{
			L3: addr.HostFromIP(cfg.IPScraper.LocalAddr),
		},
	})
	if err != nil {
		return nil, err
	}
	return snetConn.(*snet.SCIONConn), nil
}

func prepareScrapingRequests(asEntries []*db.ASEntry) (scrapeEntries, error) {
	// Scraping requests are sent in batches. To ensure that request IDs are unique let's
	// start with the current timestamp and increment the ID by 1 for each subsequent request.
	id := uint64(time.Now().UnixNano())
	entries := scrapeEntries{}
	for _, asEntry := range asEntries {
		if len(asEntry.IPAllocationProvider) == 0 {
			continue
		}
		ia, err := asEntry.ToAddrIA()
		if err != nil {
			return nil, err
		}
		host, port, err := parseAddrPort(asEntry.IPAllocationProvider)
		if err != nil {
			return nil, err
		}
		entries[fmt.Sprintf("0x%016x", id)] = &scrapeEntry{
			IA:   ia,
			IAID: asEntry.ID,
			IP:   *host,
			Port: uint16(port),
		}
		id++
	}
	return entries, nil
}

func sendScrapingRequest(ctx context.Context, conn *snet.SCIONConn, id string, entry *scrapeEntry) {
	log.Info("Sending scraping request.", "ID", id,
		"AS", entry.IA, "IP", entry.IP, "port", entry.Port)

	b, err := proto.PackRoot(allocations.NewRequest(id))
	if err != nil {
		log.Error("Cannot create scraping request", "err", err)
		return
	}
	sAddr := &snet.Addr{
		IA: entry.IA,
		Host: &addr.AppAddr{
			L3: addr.HostFromIP(entry.IP),
			L4: addr.NewL4UDPInfo(entry.Port),
		},
	}
	_, err = conn.WriteTo(b, sAddr)
	if err != nil {
		log.Error("Cannot send packet", "err", err)
		return
	}
}

func receiveScrapingReplies(conn *snet.SCIONConn, entries scrapeEntries) (scrapeEntries, error) {
	err := conn.SetReadDeadline(time.Now().Add(time.Duration(
		cfg.IPScraper.Timeout) * time.Second))
	if err != nil {
		return nil, err
	}
	done := scrapeEntries{}
	for len(entries) > 0 {
		rep := receiveScrapingReply(conn)
		if rep == nil {
			// Timeout.
			for _, entry := range entries {
				log.Warn("Scraping request timed out", "AS", entry.IA)
			}
			break
		}
		entry, ok := entries[rep.ID]
		if !ok {
			log.Warn("Unsolicited or duplicate reply received", "ID", rep.ID)
			continue
		}
		entry.Allocations = rep.Allocations
		done[rep.ID] = entry
		delete(entries, rep.ID)
		log.Info("Scraping request succeeded", "AS", entry.IA)
	}
	return done, nil
}

func receiveScrapingReply(conn *snet.SCIONConn) *allocations.Reply {
	for {
		b := make([]byte, 2000, 2000)
		size, _, err := conn.ReadFromSCION(b)
		if err != nil {
			if common.IsTimeoutErr(err) {
				return nil
			}
			log.Error("Cannot receive a scraping reply", "err", err)
			continue
		}
		b = b[:size]
		rep := &allocations.Reply{}
		err = proto.ParseFromRaw(rep, rep.ProtoId(), b)
		if err != nil {
			log.Error("Cannot parse the reply", "err", err)
			continue
		}
		return rep
	}
}

func writeRepliesToDB(dbase *gorm.DB, entries scrapeEntries) error {
	tx := dbase.Begin()
	for _, entry := range entries {
		for _, alloc := range entry.Allocations {
			// Delete the old scraped networks for the AS.
			tx.Where(fmt.Sprintf("as_entry_id = \"%d\" AND scraped = TRUE",
				entry.IAID)).Delete(db.Network{})
			// Insert the new scraped networks for the AS.
			for _, network := range alloc.Networks {
				n := db.Network{
					CIDR:      network,
					ASEntryID: entry.IAID,
					Scraped:   true,
				}
				tx.Create(&n)
			}
		}
	}
	tx.Commit()
	if tx.Error != nil {
		return tx.Error
	}
	return nil
}

func parseAddrPort(s string) (*net.IP, uint16, error) {
	hoststr, portstr, err := net.SplitHostPort(s)
	if err != nil {
		return nil, 0, err
	}
	host := net.ParseIP(hoststr)
	if host == nil {
		return nil, 0, err
	}
	port, err := strconv.ParseInt(portstr, 10, 16)
	if err != nil {
		return nil, 0, err
	}
	return &host, uint16(port), nil
}
