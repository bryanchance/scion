// Copyright 2017 Anapaya Systems

package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/scionproto/scion/go/lib/prom"
)

var (
	TotalRequests             *prometheus.CounterVec
	TotalBytes                *prometheus.CounterVec
	TotalDenials              *prometheus.CounterVec
	RequestProcessTime        *prometheus.CounterVec
	TotalACLLoads             *prometheus.CounterVec
	TotalACLChecks            *prometheus.CounterVec
	TotalACLCheckTime         *prometheus.CounterVec
	TotalTopoLoads            *prometheus.CounterVec
	TotalConsulUpdateTime     prometheus.Counter
	ConsulLastUpdate          prometheus.Gauge
	TotalConsulUpdates        *prometheus.CounterVec
	TotalConsulServiceUpdates *prometheus.CounterVec
	TotalDynamicUpdateTime    prometheus.Counter
)

func Init(elem string) {
	namespace := "discovery"
	reqLabels := []string{"src", "scope"}
	loadLabels := []string{"result"}

	newC := func(name, help string) prometheus.Counter {
		return prom.NewCounter(namespace, "", name, help)
	}
	newCVec := func(name, help string, lNames []string) *prometheus.CounterVec {
		return prom.NewCounterVec(namespace, "", name, help, lNames)
	}
	newG := func(name, help string) prometheus.Gauge {
		return prom.NewGauge(namespace, "", name, help)
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
	TotalDynamicUpdateTime = newC("total_dynamic_update_time",
		"Total amount of time spent on updating the dynamic topologies")
	TotalConsulUpdateTime = newC("total_consul_update_time",
		"Total amount of time spent on getting updates from Consul")
	ConsulLastUpdate = newG("consul_last_update", "Time of last successful update from Consul")
	TotalConsulUpdates = newCVec("total_consul_updates",
		"Total number of (attempted) updates from Consul", loadLabels)
	TotalConsulServiceUpdates = newCVec("total_consul_service_updates",
		"Total number of service updates from Consul", []string{"service", "result"})

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
