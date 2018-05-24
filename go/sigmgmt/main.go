// Copyright 2018 Anapaya Systems

package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/urfave/negroni"

	"github.com/scionproto/scion/go/lib/log"
	jwt "github.com/scionproto/scion/go/sigmgmt/auth"
	"github.com/scionproto/scion/go/sigmgmt/config"
	"github.com/scionproto/scion/go/sigmgmt/ctrl"
	"github.com/scionproto/scion/go/sigmgmt/db"
)

var (
	id          = flag.String("id", "", "Element ID, e.g. 'sm4-21-9'")
	bindAddress = flag.String("bind", "", "Address to serve on (host:port or ip:port or :port)")
	cfgPath     = flag.String("config", "cfg.json", "Path to configuration file")
)

func main() {
	os.Setenv("TZ", "UTC")
	log.AddLogFileFlags()
	flag.Parse()
	if err := log.SetupFromFlags(*id); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s", err)
		os.Exit(1)
	}
	cfg, err := config.LoadConfig(*cfgPath)
	if err != nil {
		fatal("Unable to load config file", "err", err)
	}
	dbase, err := gorm.Open("sqlite3", cfg.DBPath)
	if err != nil {
		fatal("Unable to connect to database", "err", err)
	}
	defer dbase.Close()
	db.MigrateDB(dbase)

	fmt.Printf("WebUI started, go to: https://%s\n", *bindAddress)
	log.Debug("Started WebUI on", "address", *bindAddress)
	router := configureRouter(cfg, dbase)
	exit := http.ListenAndServeTLS(*bindAddress, cfg.TLSCertificate, cfg.TLSKey, router)
	fatal("HTTP server exit", "reason", exit)
}

func fatal(msg string, desc ...interface{}) {
	log.Crit(msg, desc...)
	os.Exit(1)
}

func configureRouter(cfg *config.Global, dbase *gorm.DB) http.Handler {
	r := mux.NewRouter()

	jwtAuth := jwt.NewJWTAuth(cfg)
	auth := ctrl.NewAuthController(cfg, jwtAuth)
	r.HandleFunc("/api/auth", auth.GetToken).Methods("POST")

	c := ctrl.NewController(dbase, cfg)

	type methodMap map[string]func(http.ResponseWriter, *http.Request, http.HandlerFunc)
	for path, methods := range map[string]methodMap{
		"/api/sites": {
			"GET":  c.GetSites,
			"POST": c.PostSite,
		},
		"/api/sites/{site}": {
			"GET":    c.GetSite,
			"PUT":    c.PutSite,
			"DELETE": c.DeleteSite,
		},
		"/api/sites/{site}/reload-config": {
			"GET": c.ReloadConfig,
		},
		"/api/sites/{site}/paths": {
			"GET":  c.GetPathSelectors,
			"POST": c.PostPathSelector,
		},
		"/api/paths/{selector}": {
			"PUT":    c.PutPathSelector,
			"DELETE": c.DeletePathSelector,
		},
		"/api/sites/{site}/ases": {
			"GET":  c.GetASes,
			"POST": c.PostAS,
		},
		"/api/ases/{as}": {
			"GET":    c.GetAS,
			"PUT":    c.UpdateAS,
			"DELETE": c.DeleteAS,
		},
		"/api/ases/{as}/policies": {
			"PUT": c.UpdatePolicy,
		},
		"/api/ases/{as}/networks": {
			"GET":  c.GetNetworks,
			"POST": c.PostNetwork,
		},
		"/api/networks/{network}": {
			"DELETE": c.DeleteNetwork,
		},
		"/api/ases/{as}/sigs": {
			"GET":  c.GetSIGs,
			"POST": c.PostSIG,
		},
		"/api/sigs/default": {
			"GET": c.GetDefaultSIG,
		},
		"/api/sigs/{sig}": {
			"PUT":    c.UpdateSIG,
			"DELETE": c.DeleteSIG,
		},
		"/api/token/refresh": {
			"POST": auth.RefreshToken,
		},
	} {
		for method, handler := range methods {
			r.Handle(
				path,
				negroni.New(
					negroni.HandlerFunc(jwtAuth.RequiredAuthenticated),
					negroni.HandlerFunc(handler),
				),
			).Methods(method)
		}
	}

	r.PathPrefix("/doc/").Handler(http.FileServer(http.Dir(cfg.WebAssetRoot)))
	// TODO(worxli): check if files exist, otherwise serve index.html - or use nginx instead!
	r.PathPrefix("/").Handler(http.FileServer(http.Dir(cfg.WebAssetRoot + "webui")))

	handler := handlers.CORS(
		handlers.AllowedMethods([]string{"POST", "PUT", "OPTIONS", "DELETE", "GET", "HEAD"}),
		handlers.AllowedHeaders([]string{"Content-Type", "Authorization"}),
		handlers.AllowedOrigins([]string{"*"}),
	)(r)

	return handler
}
