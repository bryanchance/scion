// Copyright 2017 Anapaya Systems

package static

import (
	"io/ioutil"
	"os"

	log "github.com/inconshreveable/log15"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/netsec-ethz/scion/go/discovery/metrics"
	"github.com/netsec-ethz/scion/go/discovery/util"
	"github.com/netsec-ethz/scion/go/lib/common"
	"github.com/netsec-ethz/scion/go/lib/topology"
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

func Load(filename string, usefmod bool) *common.Error {
	l := prometheus.Labels{"result": ""}
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		l["result"] = ERRFILEREAD
		metrics.TotalTopoLoads.With(l).Inc()
		return common.NewError("Could not load topology.", "filename", filename, "err", err)
	}
	DiskTopo = b
	rt, cerr := topology.LoadRaw(b)
	if cerr != nil {
		return cerr
	}
	if usefmod {
		log.Debug("Resetting topology timestamp to file modification time")
		fi, err := os.Stat(filename)
		if err != nil {
			l["result"] = ERRFILESTAT
			metrics.TotalTopoLoads.With(l).Inc()
			return common.NewError("Could not stat topo file", "filename", filename, "err", err)
		}
		rt.Timestamp = fi.ModTime().Unix()
		rt.TimestampHuman = fi.ModTime().Format(common.TimeFmt)
	}

	topology.StripBind(rt)
	b, cerr = util.MarshalToJSON(rt)
	if cerr != nil {
		l["result"] = ERRMARSHALFULL
		metrics.TotalTopoLoads.With(l).Inc()
		return cerr
	}
	TopoFull.Store(b)

	// We can edit the topo since we have a "copy" of it in TopoFull now
	topology.StripServices(rt)
	b, cerr = util.MarshalToJSON(rt)
	if cerr != nil {
		l["result"] = "marshal-limited-error"
		metrics.TotalTopoLoads.With(l).Inc()
		return cerr
	}
	TopoLimited.Store(b)
	l["result"] = SUCCESS
	metrics.TotalTopoLoads.With(l).Inc()
	return nil
}
