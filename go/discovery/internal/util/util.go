// Copyright 2017 Anapaya Systems

package util

import (
	"encoding/json"
	"net"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/scionproto/scion/go/discovery/internal/acl"
	"github.com/scionproto/scion/go/discovery/internal/metrics"
	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/lib/topology"
)

type AtomicTopo struct {
	atomic.Value
}

func (a *AtomicTopo) Load() []byte {
	return a.Value.Load().([]byte)
}

func MarshalToJSON(rt *topology.RawTopo) ([]byte, error) {
	b, err := json.MarshalIndent(rt, "", "  ")
	if err != nil {
		return nil, common.NewBasicError("Could not re-marshal topology", err)
	}
	return b, nil
}

type counters struct {
	reqProcessTime prometheus.Counter
	totalReqs      prometheus.Counter
	totalBytes     prometheus.Counter
	totalErrors    prometheus.Counter
}

func countersFromLabels(labels prometheus.Labels) counters {
	return counters{
		reqProcessTime: metrics.RequestProcessTime.With(labels),
		totalReqs:      metrics.TotalRequests.With(labels),
		totalBytes:     metrics.TotalBytes.With(labels),
		totalErrors:    metrics.TotalServerErrors.With(labels),
	}
}

func MakeHandler(topo *AtomicTopo, promLabels prometheus.Labels) func(http.ResponseWriter,
	*http.Request) {

	c := countersFromLabels(promLabels)
	return func(w http.ResponseWriter, r *http.Request) {
		handle(time.Now(), w, topo.Load(), c)
	}
}

func MakeACLHandler(topo *AtomicTopo, promLabels prometheus.Labels) func(http.ResponseWriter,
	*http.Request) {

	c := countersFromLabels(promLabels)
	deniedReqs := metrics.TotalDenials.With(promLabels)

	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ip, _, _ := net.SplitHostPort(r.RemoteAddr)
		if !acl.IsAllowed(net.ParseIP(ip)) {
			deniedReqs.Inc()
			http.Error(w, "Access denied to full topology", http.StatusForbidden)
		} else {
			writeBody(w, topo.Load(), c.totalBytes, c.totalErrors)
		}
		c.reqProcessTime.Add(time.Since(start).Seconds())
		c.totalReqs.Inc()
	}
}

func MakeDefaultHandler(full *AtomicTopo, endhost *AtomicTopo,
	infraLabels, endhostLabels prometheus.Labels) func(http.ResponseWriter, *http.Request) {

	cInfra := countersFromLabels(infraLabels)
	cEndhost := countersFromLabels(endhostLabels)

	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ip, _, _ := net.SplitHostPort(r.RemoteAddr)
		if !acl.IsAllowed(net.ParseIP(ip)) {
			handle(start, w, endhost.Load(), cEndhost)
			return
		}
		handle(start, w, full.Load(), cInfra)
	}
}

func handle(start time.Time, w http.ResponseWriter, body []byte, c counters) {
	writeBody(w, body, c.totalBytes, c.totalErrors)
	c.reqProcessTime.Add(time.Since(start).Seconds())
	c.totalReqs.Inc()
}

func writeBody(w http.ResponseWriter, body []byte, totalBytes, totalErrors prometheus.Counter) {
	if len(body) <= 0 {
		http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)
		totalErrors.Inc()
		return
	}
	w.Write(body)
	totalBytes.Add(float64(len(body)))
}
