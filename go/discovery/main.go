// Copyright 2017 Anapaya Systems

package main

import (
	"flag"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/scionproto/scion/go/discovery/dsconfig"
	"github.com/scionproto/scion/go/discovery/dynamic"
	"github.com/scionproto/scion/go/discovery/metrics"
	"github.com/scionproto/scion/go/discovery/static"
	"github.com/scionproto/scion/go/discovery/util"
	"github.com/scionproto/scion/go/lib/addr"
	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/lib/env"
	"github.com/scionproto/scion/go/lib/log"
	"github.com/scionproto/scion/go/lib/periodic"
)

type Config struct {
	General env.General
	Logging env.Logging
	Metrics env.Metrics
	Infra   env.Infra
	DS      dsconfig.Config
}

var (
	ia          addr.IA
	dynUpdater  *periodic.Runner
	environment *env.Env
)

func init() {
	flag.Usage = env.Usage
}

func main() {
	os.Exit(realMain())
}

func realMain() int {
	defer log.Flush()
	env.AddFlags()
	flag.Parse()
	if v, ok := env.CheckFlags(dsconfig.Sample); !ok {
		return v
	}
	config, err := setup(env.ConfigFile())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		flag.Usage()
		return 1
	}
	defer log.LogPanicAndExit()

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
	} else {
		log.Info("Starting plain HTTP server")
		return http.ListenAndServe(address, smux)
	}
}
