// Copyright 2017 Anapaya Systems

package main

import (
	"flag"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"

	consulapi "github.com/hashicorp/consul/api"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/scionproto/scion/go/discovery/internal/config"
	"github.com/scionproto/scion/go/discovery/internal/dynamic"
	"github.com/scionproto/scion/go/discovery/internal/metrics"
	"github.com/scionproto/scion/go/discovery/internal/static"
	"github.com/scionproto/scion/go/discovery/internal/util"
	"github.com/scionproto/scion/go/lib/addr"
	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/lib/discovery"
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
	c, err := setup(config)
	if err != nil {
		log.Crit("Setup failed", "err", err)
		return 1
	}
	defer c.Close()
	http.DefaultServeMux.Handle("/metrics", promhttp.Handler())

	pubMux := http.NewServeMux()

	l := metrics.CreateReqLabels(metrics.Static, metrics.Endhost, metrics.Endhost)
	f := util.MakeHandler(static.TopoLimited, l)
	pubMux.HandleFunc(fmt.Sprintf("/%s", discovery.Path(discovery.Static, discovery.Endhost)), f)

	l = metrics.CreateReqLabels(metrics.Static, metrics.Full, metrics.Full)
	f = util.MakeACLHandler(static.TopoFull, l)
	pubMux.HandleFunc(fmt.Sprintf("/%s", discovery.Path(discovery.Static, discovery.Full)), f)

	l = metrics.CreateReqLabels(metrics.Dynamic, metrics.Endhost, metrics.Endhost)
	f = util.MakeHandler(dynamic.TopoLimited, l)
	pubMux.HandleFunc(fmt.Sprintf("/%s", discovery.Path(discovery.Dynamic, discovery.Endhost)), f)

	l = metrics.CreateReqLabels(metrics.Dynamic, metrics.Full, metrics.Full)
	f = util.MakeACLHandler(dynamic.TopoFull, l)
	pubMux.HandleFunc(fmt.Sprintf("/%s", discovery.Path(discovery.Dynamic, discovery.Full)), f)

	lInfra := metrics.CreateReqLabels(metrics.Static, metrics.Default, metrics.Full)
	lEndhost := metrics.CreateReqLabels(metrics.Static, metrics.Default, metrics.Endhost)
	f = util.MakeDefaultHandler(static.TopoFull, static.TopoLimited, lInfra, lEndhost)
	pubMux.HandleFunc(fmt.Sprintf("/%s", discovery.Path(discovery.Static, discovery.Default)), f)

	lInfra = metrics.CreateReqLabels(metrics.Dynamic, metrics.Default, metrics.Full)
	lEndhost = metrics.CreateReqLabels(metrics.Dynamic, metrics.Default, metrics.Endhost)
	f = util.MakeDefaultHandler(dynamic.TopoFull, dynamic.TopoLimited, lInfra, lEndhost)
	pubMux.HandleFunc(fmt.Sprintf("/%s", discovery.Path(discovery.Dynamic, discovery.Default)), f)

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
