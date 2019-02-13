// Copyright 2017 Anapaya Systems

package static

import (
	"io/ioutil"
	"os"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/scionproto/scion/go/discovery/internal/metrics"
	"github.com/scionproto/scion/go/discovery/internal/util"
	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/lib/log"
	"github.com/scionproto/scion/go/lib/topology"
)

var TopoFull *util.AtomicTopo
var TopoLimited *util.AtomicTopo

// The on-disk topo as it was the last time we loaded it
var DiskTopo *util.AtomicTopo

const (
	ErrFileRead       = "file-read-error"
	ErrFileStat       = "stat-error"
	ErrMarshalFull    = "marshal-full-error"
	ErrMarshalEndhost = "marshal-endhost-error"
	Success           = "success"
)

func init() {
	TopoFull = &util.AtomicTopo{}
	TopoFull.Store([]byte(nil))
	TopoLimited = &util.AtomicTopo{}
	TopoLimited.Store([]byte(nil))
	DiskTopo = &util.AtomicTopo{}
	DiskTopo.Store([]byte(nil))
}

func Load(filename string, usefmod bool) error {
	l := prometheus.Labels{"result": ""}
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		l["result"] = ErrFileRead
		metrics.TotalTopoLoads.With(l).Inc()
		return common.NewBasicError("Could not load topology.", err, "filename", filename)
	}
	DiskTopo.Store(b)
	rt, err := topology.LoadRaw(b)
	if err != nil {
		return err
	}
	if usefmod {
		log.Debug("Resetting topology timestamp to file modification time")
		fi, err := os.Stat(filename)
		if err != nil {
			l["result"] = ErrFileStat
			metrics.TotalTopoLoads.With(l).Inc()
			return common.NewBasicError("Could not stat topo file", err, "filename", filename)
		}
		rt.Timestamp = fi.ModTime().Unix()
		rt.TimestampHuman = fi.ModTime().Format(common.TimeFmt)
	}

	b, err = util.MarshalToJSON(rt)
	if err != nil {
		l["result"] = ErrMarshalFull
		metrics.TotalTopoLoads.With(l).Inc()
		return err
	}
	TopoFull.Store(b)

	// We can edit the topo since we have a "copy" of it in TopoFull now
	topology.StripBind(rt)
	topology.StripServices(rt)
	b, err = util.MarshalToJSON(rt)
	if err != nil {
		l["result"] = ErrMarshalEndhost
		metrics.TotalTopoLoads.With(l).Inc()
		return err
	}
	TopoLimited.Store(b)
	l["result"] = Success
	metrics.TotalTopoLoads.With(l).Inc()
	return nil
}
