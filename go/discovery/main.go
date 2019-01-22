// Copyright 2017 Anapaya Systems

package main

import (
	"flag"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"

	consulapi "github.com/hashicorp/consul/api"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/scionproto/scion/go/discovery/internal/config"
	"github.com/scionproto/scion/go/discovery/internal/dynamic"
	"github.com/scionproto/scion/go/discovery/internal/metrics"
	"github.com/scionproto/scion/go/discovery/internal/static"
	"github.com/scionproto/scion/go/discovery/internal/util"
	"github.com/scionproto/scion/go/lib/addr"
	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/lib/env"
	"github.com/scionproto/scion/go/lib/fatal"
	"github.com/scionproto/scion/go/lib/log"
	"github.com/scionproto/scion/go/lib/periodic"
)

var (
	ia           addr.IA
	dynUpdater   *periodic.Runner
	environment  *env.Env
	consulClient *consulapi.Client
)

func init() {
	flag.Usage = env.Usage
}

func main() {
	os.Exit(realMain())
}

func realMain() int {
	fatal.Init()
	env.AddFlags()
	flag.Parse()
	if v, ok := env.CheckFlags(config.Sample); !ok {
		return v
	}
	config, err := setupBasic()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	defer log.Flush()
	defer env.LogAppStopped(common.DS, config.General.ID)
	defer log.LogPanicAndExit()
	if err := setup(config); err != nil {
		log.Crit("Setup failed", "err", err)
		return 1
	}
	http.DefaultServeMux.Handle("/metrics", promhttp.Handler())

	pubMux := http.NewServeMux()

	l := prometheus.Labels{"src": "static", "scope": "endhost"}
	f := util.MakeHandler(static.TopoLimited, l)
	pubMux.HandleFunc("/discovery/v1/static/reduced.json", f)

	l = prometheus.Labels{"src": "static", "scope": "infrastructure"}
	f = util.MakeACLHandler(static.TopoFull, l)
	pubMux.HandleFunc("/discovery/v1/static/full.json", f)

	l = prometheus.Labels{"src": "dynamic", "scope": "endhost"}
	f = util.MakeHandler(dynamic.TopoLimited, l)
	pubMux.HandleFunc("/discovery/v1/dynamic/reduced.json", f)

	l = prometheus.Labels{"src": "dynamic", "scope": "infrastructure"}
	f = util.MakeACLHandler(dynamic.TopoFull, l)
	pubMux.HandleFunc("/discovery/v1/dynamic/full.json", f)

	fatalC := make(chan error, 2)
	if config.Metrics.Prometheus != "" {
		go func() {
			log.Info("Starting private (prometheus, pprof) server", "addr",
				config.Metrics.Prometheus)
			http.DefaultServeMux.HandleFunc("/", metrics.MakeMainDebugPageHandler())
			err := runHTTPServer(config.Metrics.Prometheus, config.DS.Cert, config.DS.Key,
				http.DefaultServeMux)
			// If runHTTPServer returns, there will be a non-nil err
			fatalC <- common.NewBasicError("Could not start prometheus HTTP server", err)
		}()
	}

	go func() {
		log.Info("Starting public server", "addr", config.DS.ListenAddr)
		err := runHTTPServer(config.DS.ListenAddr, config.DS.Cert, config.DS.Key, pubMux)
		// If runHTTPServer returns, there will be a non-nil err
		fatalC <- common.NewBasicError("Could not start discovery HTTP server", err)
	}()
	select {
	case <-environment.AppShutdownSignal:
		// Whenever we receive a SIGINT or SIGTERM we exit without an error.
		return 0
	case err := <-fatalC:
		log.Crit("Unable to listen and serve", "err", err)
		return 1
	}
}

func runHTTPServer(address string, certfile, keyfile string, smux *http.ServeMux) error {
	if certfile != "" && keyfile != "" {
		log.Info("Starting TLS server", "certfile", certfile, "keyfile", keyfile)
		return http.ListenAndServeTLS(address, certfile, keyfile, smux)
	}
	log.Info("Starting plain HTTP server")
	return http.ListenAndServe(address, smux)
}
