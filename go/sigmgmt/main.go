// Copyright 2018 Anapaya Systems

package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
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
	dbase, err := db.New(cfg.DBPath)
	if err != nil {
		fatal("Unable to connect to database", "err", err)
	}

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

func configureRouter(cfg *config.Global, dbase *db.DB) http.Handler {
	r := mux.NewRouter()

	jwtAuth := jwt.NewJWTAuth(cfg)
	auth := ctrl.NewAuthController(cfg, jwtAuth)
	r.HandleFunc("/api/auth", auth.GetToken).Methods("POST")

	sc := ctrl.NewSiteController(dbase, cfg)
	ac := ctrl.NewASController(dbase, cfg)

	type methodMap map[string]func(http.ResponseWriter, *http.Request, http.HandlerFunc)
	for path, methods := range map[string]methodMap{
		"/api/sites": {
			"GET":  sc.GetSites,
			"POST": sc.PostSite,
		},
		"/api/sites/{site}": {
			"GET":    sc.GetSite,
			"PUT":    sc.PutSite,
			"DELETE": sc.DeleteSite,
		},
		"/api/sites/{site}/reload-config": {
			"GET": sc.ReloadConfig,
		},
		"/api/sites/{site}/paths": {
			"GET":  sc.GetPathPredicates,
			"POST": sc.AddPathPredicate,
		},
		"/api/sites/{site}/paths/{path}": {
			"PUT":    sc.PutPathPredicate,
			"DELETE": sc.DeletePathPredicate,
		},
		"/api/sites/{site}/ias": {
			"GET":  ac.GetASes,
			"POST": ac.PostAS,
		},
		"/api/sites/{site}/ias/{ia}": {
			"GET":    ac.GetAS,
			"PUT":    ac.UpdateAS,
			"DELETE": ac.DeleteAS,
		},
		"/api/sites/{site}/ias/{ia}/policies": {
			"PUT": ac.UpdatePolicy,
		},
		"/api/sites/{site}/ias/{ia}/networks": {
			"GET":  ac.GetNetworks,
			"POST": ac.PostNetwork,
		},
		"/api/sites/{site}/ias/{ia}/networks/{network}": {
			"DELETE": ac.DeleteNetwork,
		},
		"/api/sites/{site}/ias/{ia}/sigs": {
			"GET":  ac.GetSIGs,
			"POST": ac.PostSIG,
		},
		"/api/sites/{site}/ias/{ia}/sigs/{sig}": {
			"PUT":    ac.UpdateSIG,
			"DELETE": ac.DeleteSIG,
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
