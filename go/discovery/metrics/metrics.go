// Copyright 2017 Anapaya Systems

package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/scionproto/scion/go/lib/prom"
)

var (
	TotalRequests          *prometheus.CounterVec
	TotalBytes             *prometheus.CounterVec
	TotalDenials           *prometheus.CounterVec
	RequestProcessTime     *prometheus.CounterVec
	TotalACLLoads          *prometheus.CounterVec
	TotalACLChecks         *prometheus.CounterVec
	TotalACLCheckTime      *prometheus.CounterVec
	TotalTopoLoads         *prometheus.CounterVec
	TotalZkUpdateTime      prometheus.Counter
	ZKLastUpdate           prometheus.Counter
	TotalZkUpdates         *prometheus.CounterVec
	TotalServiceUpdates    *prometheus.CounterVec
	TotalDynamicUpdateTime prometheus.Counter
)

func Init(elem string) {
	namespace := "discovery"
	constLabels := prometheus.Labels{"elem": elem}
	reqLabels := []string{"src", "scope"}
	loadLabels := []string{"result"}

	newC := func(name, help string) prometheus.Counter {
		v := prom.NewCounter(namespace, "", name, help, constLabels)
		prometheus.MustRegister(v)
		return v
	}
	newCVec := func(name, help string, lNames []string) *prometheus.CounterVec {
		v := prom.NewCounterVec(namespace, "", name, help, constLabels, lNames)
		prometheus.MustRegister(v)
		return v
	}
	// Global counters used by both the static and the dynamic part
	TotalRequests = newCVec("total_requests", "Number of requests served", reqLabels)
	TotalBytes = newCVec("total_bytes_sent", "Number of bytes served", reqLabels)
	TotalDenials = newCVec("total_denied_topo_requests",
		"Number of denials for full topology requests", reqLabels)
	RequestProcessTime = newCVec("request_process_seconds",
		"Time spent processing requests", reqLabels)

	// ACL file
	TotalACLLoads = newCVec("total_acl_loads",
		"Number of times the ACL file has been (re)loaded", loadLabels)
	TotalACLChecks = newCVec("total_acl_checks", "Number of ACL checks performed", loadLabels)
	TotalACLCheckTime = newCVec("total_acl_check_time",
		"Total time spent checking ACLs", loadLabels)

	// Metrics used only by the static part
	TotalTopoLoads = newCVec("total_topofile_loads",
		"Number of times the topology file has been (re)loaded", loadLabels)

	// Metrics used only by the dynamic part
	TotalZkUpdateTime = newC("total_zookeeper_update_time",
		"Total amount of time spent on getting updates from Zookeeper")
	ZKLastUpdate = newC("zookeeper_last_update", "Time of last successful update from Zookeeper")
	TotalZkUpdates = newCVec("total_zookeeper_updates",
		"Total number of (attempted) updates from zookeeper", loadLabels)
	TotalServiceUpdates = newCVec("total_service_updates",
		"Total number of service updates from Zookeper", []string{"service", "result"})
	TotalDynamicUpdateTime = newC("total_dynamic_update_time",
		"Total amount of time spent on updating the dynamic topologies")
}

func MakeMainDebugPageHandler() func(http.ResponseWriter, *http.Request) {
	// Just a very simple page on / with links to the relevant debug things
	return func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
			<head><title>Discovery Service Debug Page</title></head>
			<body>
			<h1>Discovery Service Debug Page</h1>
			<p><a href=/metrics>Metrics</a></p>
			<p><a href=/debug/pprof>PProf</a></p>
			</body>
			</html>`))
	}
}
