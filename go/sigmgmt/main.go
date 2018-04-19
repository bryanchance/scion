// Copyright 2018 Anapaya Systems

package main

import (
	"flag"
	"net/http"
	"os"

	log "github.com/inconshreveable/log15"

	liblog "github.com/scionproto/scion/go/lib/log"
	"github.com/scionproto/scion/go/sigmgmt/config"
	"github.com/scionproto/scion/go/sigmgmt/db"
	"github.com/scionproto/scion/go/sigmgmt/web"
)

var (
	id          = flag.String("id", "", "Element ID, e.g. 'sm4-21-9'")
	bindAddress = flag.String("bind", "", "Address to serve on (host:port or ip:port or :port)")
	cfgPath     = flag.String("config", "cfg.json", "Path to configuration file")
)

func main() {
	liblog.AddDefaultLogFlags()
	flag.Parse()
	liblog.Setup(*id)

	cfg, err := config.LoadConfig(*cfgPath)
	if err != nil {
		fatal("Unable to load config file", "err", err)
	}
	dbase, err := db.New(cfg.DBPath)
	if err != nil {
		fatal("Unable to connect to database", "err", err)
	}

	// FIXME(scrye): The application currently provides no authentication, so
	// it is vulnerable to CSRF (and just about any other attack under the
	// sun). It should only be deployed in test or isolated production
	// environments for now.
	app := web.NewApplication(cfg, dbase)
	http.HandleFunc("/", app.ServeHTTP)
	log.Info("Starting HTTP server", "address", *bindAddress)
	log.Info("HTTP server exit", "reason", http.ListenAndServe(*bindAddress, nil))
}

func fatal(msg string, desc ...interface{}) {
	log.Crit(msg, desc...)
	os.Exit(1)
}
