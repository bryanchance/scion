// Copyright 2017 Anapaya Systems

package static

import (
	"io/ioutil"
	"os"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/scionproto/scion/go/discovery/metrics"
	"github.com/scionproto/scion/go/discovery/util"
	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/lib/log"
	"github.com/scionproto/scion/go/lib/topology"
)

var TopoFull *util.AtomicTopo
var TopoLimited *util.AtomicTopo

// The on-disk topo as it was the last time we loaded it
var DiskTopo []byte

const (
	ERRFILEREAD       = "file-read-error"
	ERRFILESTAT       = "stat-error"
	ERRMARSHALFULL    = "marshal-full-error"
	ERRMARSHALREDUCED = "marshal-reduced-error"
	SUCCESS           = "success"
)

func init() {
	TopoFull = &util.AtomicTopo{}
	TopoFull.Store([]byte(nil))
	TopoLimited = &util.AtomicTopo{}
	TopoLimited.Store([]byte(nil))
}

func Load(filename string, usefmod bool) error {
	l := prometheus.Labels{"result": ""}
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		l["result"] = ERRFILEREAD
		metrics.TotalTopoLoads.With(l).Inc()
		return common.NewBasicError("Could not load topology.", err, "filename", filename)
	}
	DiskTopo = b
	rt, err := topology.LoadRaw(b)
	if err != nil {
		return err
	}
	if usefmod {
		log.Debug("Resetting topology timestamp to file modification time")
		fi, err := os.Stat(filename)
		if err != nil {
			l["result"] = ERRFILESTAT
			metrics.TotalTopoLoads.With(l).Inc()
			return common.NewBasicError("Could not stat topo file", err, "filename", filename)
		}
		rt.Timestamp = fi.ModTime().Unix()
		rt.TimestampHuman = fi.ModTime().Format(common.TimeFmt)
	}

	topology.StripBind(rt)
	b, err = util.MarshalToJSON(rt)
	if err != nil {
		l["result"] = ERRMARSHALFULL
		metrics.TotalTopoLoads.With(l).Inc()
		return err
	}
	TopoFull.Store(b)

	// We can edit the topo since we have a "copy" of it in TopoFull now
	topology.StripServices(rt)
	b, err = util.MarshalToJSON(rt)
	if err != nil {
		l["result"] = "marshal-limited-error"
		metrics.TotalTopoLoads.With(l).Inc()
		return err
	}
	TopoLimited.Store(b)
	l["result"] = SUCCESS
	metrics.TotalTopoLoads.With(l).Inc()
	return nil
}
