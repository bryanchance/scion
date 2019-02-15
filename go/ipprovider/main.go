// Copyright 2018 Anapaya Systems

// This tool provides IP ranges that are meant to be accessible from different ASes
// to the peer SIGs. The IP ranges can be configured via sigmgmt UI.
package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"

	"github.com/scionproto/scion/go/ipprovider/internal/config"
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
	env.LogAppStarted("IPProvider", cfg.IPProvider.LocalIA.String())
	defer env.LogAppStopped("IPProvider", cfg.IPProvider.LocalIA.String())

	// Open a database connection. The database is shared with sigmgmt.
	dbase, err := gorm.Open("sqlite3", cfg.IPProvider.DBPath)
	if err != nil {
		log.Crit("Unable to connect to database", "err", err)
		return 1
	}
	defer dbase.Close()
	db.MigrateDB(dbase)

	// Initialize SCION connection.
	conn, err := initSCIONConnection()
	if err != nil {
		log.Crit("Cannot open SCION connection", "err", err)
		return 1
	}

	// Main loop of the application. It receives scraping requests from the network,
	// queries the sigmgmt database and sends the results back to the requester.
	for {
		req, addr, err := receiveRequest(conn)
		if err != nil {
			log.Error("Cannot receive a request", "err", err)
			continue
		}
		log.Info("Request received", "AS", addr.IA, "ID", req.ID)
		allocs, err := getAllocationsFromDB(dbase, addr.IA)
		if err != nil {
			log.Error("Cannot read allocations from the database",
				"AS", addr.IA, "err", err)
			continue
		}
		err = sendReply(conn, addr, allocations.NewReply(req.ID, allocs))
		if err != nil {
			log.Error("Cannot send a reply", "AS", addr.IA, "err", err)
			continue
		}
	}
}

func initSCIONConnection() (*snet.SCIONConn, error) {
	if err := snet.Init(cfg.IPProvider.LocalIA, cfg.Sciond.Path,
		reliable.NewDispatcherService(reliable.DefaultDispPath)); err != nil {

		return nil, err
	}
	snetConn, err := snet.ListenSCION("udp4", &snet.Addr{
		IA: cfg.IPProvider.LocalIA,
		Host: &addr.AppAddr{
			L3: addr.HostFromIP(cfg.IPProvider.LocalAddr),
			L4: addr.NewL4UDPInfo(cfg.IPProvider.LocalPort),
		},
	})
	if err != nil {
		return nil, err
	}
	return snetConn.(*snet.SCIONConn), nil
}

func receiveRequest(conn *snet.SCIONConn) (*allocations.Request, *snet.Addr, error) {
	b := make([]byte, 2000, 2000)
	size, addr, err := conn.ReadFromSCION(b)
	if err != nil {
		return nil, nil, err
	}
	b = b[:size]
	req := &allocations.Request{}
	err = proto.ParseFromRaw(req, req.ProtoId(), b)
	if err != nil {
		return nil, nil, err
	}
	return req, addr, nil
}

func getAllocationsFromDB(dbase *gorm.DB, ia addr.IA) ([]*allocations.Allocation, error) {
	var sites []*db.Site
	if err := dbase.Find(&sites).Error; err != nil {
		return nil, err
	}
	allocs := []*allocations.Allocation{}
	for _, site := range sites {
		a := allocations.Allocation{IA: ia.String()}
		var networks []*db.SiteNetwork
		if err := dbase.Where("site_id = ?",
			strconv.FormatUint(uint64(site.ID), 10)).Find(&networks).Error; err != nil {

			return nil, err
		}
		for _, network := range networks {
			ok, err := checkACL(network.ACL, ia)
			if err != nil {
				return nil, err
			}
			if ok {
				a.Networks = append(a.Networks, network.CIDR)
			}
		}
		if len(a.Networks) > 0 {
			allocs = append(allocs, &a)
		}
	}
	return allocs, nil
}

func checkACL(acl string, ia addr.IA) (bool, error) {
	for _, elem := range strings.Split(acl, " ") {
		if elem == "*" {
			return true, nil
		}
		aclia, err := addr.IAFromString(elem)
		if err != nil {
			return false, common.NewBasicError("Invalid ACL", err,
				"acl", elem)
		}
		if ia.Equal(aclia) {
			return true, nil
		}
	}
	return false, nil
}

func sendReply(conn *snet.SCIONConn, addr *snet.Addr, reply *allocations.Reply) error {
	b, err := proto.PackRoot(reply)
	if err != nil {
		return err
	}
	_, err = conn.WriteTo(b, addr)
	if err != nil {
		return err
	}
	return nil
}
