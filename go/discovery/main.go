// Copyright 2017 Anapaya Systems

package main

import (
	"flag"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/scionproto/scion/go/discovery/acl"
	"github.com/scionproto/scion/go/discovery/dynamic"
	"github.com/scionproto/scion/go/discovery/metrics"
	"github.com/scionproto/scion/go/discovery/static"
	"github.com/scionproto/scion/go/discovery/util"
	"github.com/scionproto/scion/go/lib/addr"
	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/lib/log"
)

var (
	id = flag.String("id", "", "Element ID, e.g. 'ds4-ff00:0:2f-9'")
	ia = flag.String("ia", "", "ISD-AS, e.g. '4-ff00:0:2f'")

	topofile = flag.String("static-topo", "", "Static topology file to serve (required)")
	usefmod  = flag.Bool("usefmod", true, "Use file modification time for static topo timestamp")
	aclfile  = flag.String("acl", "", "File with ACL entries for full topology (required)")

	zk = flag.String("zk", "",
		"Comma-separated list (host:port, ...) of Zookeper instances to talk to (required)")
	zkfreq    = flag.Duration("zkfreq", 5*time.Second, "How often to query Zookeeper")
	zktimeout = flag.Duration("zktimeout", 10*time.Second, "Timeout for connecting to Zookeeper")

	certfn = flag.String("cert", "", "Certificate file for TLS. If unset, serve plain HTTP")
	keyfn  = flag.String("key", "", "Key file to use for TLS. If unset, serve plain HTTP")

	laddress = flag.String("listen", "", "Address to serve on (host:port or ip:port or :port)")
	paddress = flag.String("prom", "", "Address to serve prom/pprof data on. Disabled if empty")

	// Global channel that the SIGHUP handler can listen on
	sighup chan os.Signal
)

func init() {
	// Add a SIGHUP handler as soon as possible on startup, to reduce the
	// chance that a premature SIGHUP will kill the process. This channel is
	// used by confSig below.
	sighup = make(chan os.Signal, 1)
	signal.Notify(sighup, syscall.SIGHUP)
}

func main() {
	os.Setenv("TZ", "UTC")
	log.AddLogFileFlags()
	log.AddLogConsFlags()
	isdas, err := verifyFlags()
	if err != nil {
		log.Crit(err.Error())
		log.Flush()
		os.Exit(1)
	}
	zklist := strings.Split(*zk, ",")
	if err := log.SetupFromFlags(*id); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s", err)
		os.Exit(1)
	}
	metrics.Init(*id)

	log.Info("Loading ACLs", "filename", *aclfile)
	if err = acl.Load(*aclfile); err != nil {
		log.Error(err.Error())
		log.Flush()
		os.Exit(1)
	}
	log.Info("Loading static topology", "filename", *topofile)
	if err = static.Load(*topofile, *usefmod); err != nil {
		log.Error("Could not load static topology file file", "filename", *topofile, "err", err)
		log.Flush()
		os.Exit(1)
	}
	dynamic.Setup(isdas, *topofile)
	setupSignals()

	go func() {
		log.Debug("Starting topo update loop", "zklist", zklist, "frequency", *zkfreq,
			"conntimeout", *zktimeout)
		for {
			dynamic.UpdateFromZK(zklist, *id, *zktimeout)
			time.Sleep(*zkfreq)
		}
	}()

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

	if *paddress != "" {
		go func() {
			log.Info("Starting private (prometheus, pprof) server", "addr", *paddress)
			http.DefaultServeMux.HandleFunc("/", metrics.MakeMainDebugPageHandler())
			err := runHTTPServer(*paddress, *certfn, *keyfn, http.DefaultServeMux)
			// If runHTTPServer returns, there will be a non-nil err
			log.Crit("Could not start private HTTP server", "err", err)
			log.Flush()
			os.Exit(1)
		}()
	}
	log.Info("Starting public server", "addr", *laddress)
	err = runHTTPServer(*laddress, *certfn, *keyfn, pubMux)
	log.Crit("Could not start public HTTP server", "err", err)
	log.Flush()
	os.Exit(1)
}

func verifyFlags() (addr.IA, error) {
	flag.Parse()
	if *id == "" {
		return addr.IA{}, common.NewBasicError("No element ID specified", nil)
	}
	if *ia == "" {
		return addr.IA{}, common.NewBasicError("No ISD-AS specified", nil)
	}
	isdas, err := addr.IAFromString(*ia)
	if err != nil {
		return addr.IA{}, common.NewBasicError("Could not parse ISD-AS", err, "isd-as", ia)
	}
	if *topofile == "" {
		return addr.IA{}, common.NewBasicError("No static topology file specified", nil)
	}
	if *aclfile == "" {
		return addr.IA{}, common.NewBasicError("No ACL file specified", nil)
	}
	if *laddress == "" {
		return addr.IA{}, common.NewBasicError("No address to listen on specified", nil)
	}
	if *zk == "" {
		return addr.IA{}, common.NewBasicError("No Zookeeper specified", nil)
	}
	return isdas, nil
}

func setupSignals() {
	exitsigs := make(chan os.Signal, 2)
	signal.Notify(exitsigs, os.Interrupt)
	signal.Notify(exitsigs, syscall.SIGTERM)
	go func() {
		// sighup is a global var set up in init()
		for range sighup {
			// Reload relevant static configuration files
			log.Info("Reloading ACL", "filename", *aclfile)
			err := acl.Load(*aclfile)
			if err != nil {
				log.Error("ACL file reload failed", "err", err)
				// If there was an error, we should not try to reload the topofile
				continue
			}
			log.Info("Reloading static topology file", "filename", *topofile)
			err = static.Load(*topofile, *usefmod)
			if err != nil {
				log.Error("Static topology file reload failed", "err", err)
			}
		}
	}() // End of HUP signal handler
	go func() {
		<-exitsigs
		log.Info("Exiting")
		log.Flush()
		os.Exit(1)
	}() // End of SIGINT/SIGTERM handler
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
