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

func MakeHandler(topo *AtomicTopo, promLabels prometheus.Labels) func(http.ResponseWriter,
	*http.Request) {

	reqProcessTime := metrics.RequestProcessTime.With(promLabels)
	totalReqs := metrics.TotalRequests.With(promLabels)
	totalBytes := metrics.TotalBytes.With(promLabels)
	totalErrors := metrics.TotalServerErrors.With(promLabels)

	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		if body := topo.Load(); len(body) > 0 {
			w.Write(body)
			totalBytes.Add(float64(len(body)))
		} else {
			http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)
			totalErrors.Inc()
		}
		reqProcessTime.Add(time.Since(start).Seconds())
		totalReqs.Inc()
	}
}

func MakeACLHandler(topo *AtomicTopo, promLabels prometheus.Labels) func(http.ResponseWriter,
	*http.Request) {

	reqProcessTime := metrics.RequestProcessTime.With(promLabels)
	totalReqs := metrics.TotalRequests.With(promLabels)
	deniedReqs := metrics.TotalDenials.With(promLabels)
	totalBytes := metrics.TotalBytes.With(promLabels)
	totalErrors := metrics.TotalServerErrors.With(promLabels)

	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ip, _, _ := net.SplitHostPort(r.RemoteAddr)
		if !acl.IsAllowed(net.ParseIP(ip)) {
			deniedReqs.Inc()
			http.Error(w, "Access denied to full topology", http.StatusForbidden)
		} else {
			if body := topo.Load(); len(body) > 0 {
				w.Write(body)
				totalBytes.Add(float64(len(body)))
			} else {
				http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)
				totalErrors.Inc()
			}
		}
		reqProcessTime.Add(time.Since(start).Seconds())
		totalReqs.Inc()
	}
}
