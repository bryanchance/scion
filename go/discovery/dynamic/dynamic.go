// Copyright 2017 Anapaya Systems

package dynamic

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/gavv/monotime"
	log "github.com/inconshreveable/log15"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/samuel/go-zookeeper/zk"
	"zombiezen.com/go/capnproto2"

	"github.com/netsec-ethz/scion/go/discovery/metrics"
	"github.com/netsec-ethz/scion/go/discovery/static"
	"github.com/netsec-ethz/scion/go/discovery/util"
	"github.com/netsec-ethz/scion/go/lib/addr"
	"github.com/netsec-ethz/scion/go/lib/common"
	"github.com/netsec-ethz/scion/go/lib/topology"
	"github.com/netsec-ethz/scion/go/proto"
)

var baseTopo *topology.RawTopo
var TopoFull *util.AtomicTopo
var TopoLimited *util.AtomicTopo
var isd_as *addr.ISD_AS

type wrappedLogger struct{}

func (wrappedLogger) Printf(s string, r ...interface{}) {
	// Only used for debugging the zookeeper lib
	//log.Debug("go-zookeeper", "msg", fmt.Sprintf(s, r...))
}

func init() {
	TopoFull = &util.AtomicTopo{}
	TopoFull.Store([]byte{})
	TopoLimited = &util.AtomicTopo{}
	TopoLimited.Store([]byte{})
}

func Setup(IA *addr.ISD_AS, basetopofn string) *common.Error {
	var cerr *common.Error
	isd_as = IA
	baseTopo, cerr = topology.LoadRawFromFile(basetopofn)
	return cerr
}

func UpdateFromZK(zks []string, id string, sessionTimeout time.Duration) {
	// We declare a bunch of variables early so we can use goto freely.
	var c *zk.Conn
	var cerr *common.Error
	var wl wrappedLogger

	// These should never be used with their default values
	labels := prometheus.Labels{"result": "unknown"}
	rt := &topology.RawTopo{}

	start := monotime.Now()

	// Make ZK connection and set logger
	c, _, err := zk.Connect(zks, sessionTimeout)
	if err != nil {
		log.Error("Error connecting to Zookeeper", "err", err)
		labels["result"] = "connect-error"
		goto Out
	}
	defer func() { closeZkConn(c) }()
	c.SetLogger(wl)

	// Get the base topo from the static part of DS so the two version agree as
	// much as is useful.
	if err := json.Unmarshal(static.DiskTopo, rt); err != nil {
		log.Error("Could not re-parse topology.", "err", err)
		labels["result"] = "basetopo-parse-error"
		goto Out
	}
	// Check each service and overwrite it with ZK contents iff ZK is non-empty
	if !updateServices(rt, c) {
		// we don't jump to Out here since we still want this degraded topo to
		// be used to make a new served topology. Also, while it would be nice
		// to have per-service-type error counts, it's probably not worth the
		// extra code here.
		labels["result"] = "service-update-error"
	}
	metrics.TotalZkUpdateTime.Add(float64(monotime.Since(start).Seconds()))

	// Make sure Bind info is removed from the Topo and update timestamp
	topology.StripBind(rt)
	updateTimestamps(rt)

	if cerr = marshallAndUpdate(rt, TopoFull); cerr != nil {
		log.Error("Could not marshal full topo", "err", cerr)
		labels["result"] = "marshall-full-error"
		goto Out
	}

	// Trim non-public services
	topology.StripServices(rt)

	if cerr = marshallAndUpdate(rt, TopoLimited); cerr != nil {
		log.Error("Could not marshal reduced topo", "err", cerr)
		labels["result"] = "marshall-reduced-error"
		goto Out
	}

	// We only want to update these if the update was successful, so the Out
	// label has to be below
	metrics.ZKLastUpdate.Set(float64(time.Now().Unix()))
	metrics.TotalDynamicUpdateTime.Add(float64(monotime.Since(start).Seconds()))

Out:
	// Update Prometheus metrics
	metrics.TotalZkUpdates.With(labels).Inc()
}

func closeZkConn(c *zk.Conn) {
	s := c.State()
	if s == zk.StateConnected || s == zk.StateHasSession || s == zk.StateConnecting {
		c.Close()
	}
}

func updateServices(rt *topology.RawTopo, c *zk.Conn) bool {
	var ok string

	rt.BeaconService, ok = fillService(c, isd_as, common.BS, rt.BeaconService)
	metrics.TotalServiceUpdates.With(prometheus.Labels{"result": ok, "service": common.BS}).Inc()
	success := ok == "success"

	rt.CertificateService, ok = fillService(c, isd_as, common.CS, rt.CertificateService)
	metrics.TotalServiceUpdates.With(prometheus.Labels{"result": ok, "service": common.CS}).Inc()
	success = success && ok == "success"

	rt.PathService, ok = fillService(c, isd_as, common.PS, rt.PathService)
	metrics.TotalServiceUpdates.With(prometheus.Labels{"result": ok, "service": common.PS}).Inc()
	success = success && ok == "success"

	rt.SibraService, ok = fillService(c, isd_as, common.SB, rt.SibraService)
	metrics.TotalServiceUpdates.With(prometheus.Labels{"result": ok, "service": common.SB}).Inc()
	return success && ok == "success"
	// There currently is no RAINS service anywhere, so this would always fail
	//rt.RainsService, ok = fillService(c, isd_as, common.RS, rt.RainsService)
	//metrics.TotalServiceUpdates.With(prometheusLabels{"status": ok, "service": "bs"}).Inc()
	//success = success && ok
}

func updateTimestamps(rt *topology.RawTopo) {
	ts := time.Now()
	rt.Timestamp = ts.Unix()
	rt.TimestampHuman = ts.Format(time.RFC3339)
}

func marshallAndUpdate(rt *topology.RawTopo, topo *util.AtomicTopo) *common.Error {
	b, cerr := util.MarshalToJSON(rt)
	if cerr != nil {
		return cerr
	}
	topo.Store(b)
	return nil
}

func fillService(c *zk.Conn, isd_as *addr.ISD_AS, servicetype string,
	fallback map[string]topology.RawAddrInfo) (map[string]topology.RawAddrInfo, string) {
	service, err := getZkService(c, isd_as, servicetype)
	if err != nil {
		log.Warn("Could not fetch service entries from ZK, using fallback",
			"servicetype", servicetype, "err", err)
		return fallback, "fetch-failure"
	}
	if len(service) == 0 {
		log.Warn("Service listing is empty on ZK, using fallback", "servicetype", servicetype)
		return fallback, "service-empty"
	}
	return service, "success"
}

func getZkService(connection *zk.Conn, isdas *addr.ISD_AS,
	servertype string) (map[string]topology.RawAddrInfo, error) {
	services := make(map[string]topology.RawAddrInfo)
	partybase := fmt.Sprintf("/%s/%s/party", isdas, servertype)
	children, _, err := connection.Children(partybase)
	if err != nil {
		return nil, err
	}
	if len(children) == 0 {
		return nil, nil
	}
	for _, child := range children {
		partypath := fmt.Sprintf("%s/%s", partybase, child)
		data, _, err := connection.Get(partypath)
		if err != nil {
			return nil, err
		}
		sinfo, cerr := decodePartydata(data)
		if cerr != nil {
			log.Error(cerr.Desc, cerr.Ctx...)
			return nil, err
		}
		sid, err := sinfo.Id()
		if err != nil {
			log.Error("Could not decode service Id", "sinfo", sinfo, "err", err)
			return nil, err
		}
		addrs, err := sinfo.Addrs()
		if err != nil {
			log.Error("Could not decode service Addrs", "sinfo", sinfo, "err", err)
			return nil, err
		}
		for i := 0; i < addrs.Len(); i++ {
			var saddr string
			addr, err := addrs.At(i).Addr()
			if err != nil {
				log.Error("Could not fetch service address", "addrs", addrs, "err", err)
				return nil, err
			}
			ip := net.IP(addr)
			if ip == nil {
				return nil, errors.New(fmt.Sprintf("Could not parse IP '%v'", ip))
			}
			saddr = ip.String()
			// Make a RemoteAddrInfo we can put into the topology for later serving
			services[sid] = raiFromAddrPortOverlay(saddr, int(addrs.At(i).Port()), 0)
		}
	}
	return services, nil
}

func decodePartydata(b []byte) (*proto.ZkId, *common.Error) {
	// FIXME(klausman): switch to use Cerealizable
	decoded := make([]byte, len(b))
	length, err := base64.StdEncoding.Decode(decoded, b)
	if err != nil {
		return nil, common.NewError("Could not base64-decode party data", "err", err)
	}
	msg, err := capnp.NewPackedDecoder(bytes.NewBuffer(decoded[:length])).Decode()
	if err != nil {
		return nil, common.NewError("Could not decode party data", "err", err)
	}
	zkid, err := proto.ReadRootZkId(msg)
	if err != nil {
		return nil, common.NewError("Could not read root ZkId from party data", "err", err)
	}
	return &zkid, nil
}

func raiFromAddrPortOverlay(addr string, l4port int, overlayport int) topology.RawAddrInfo {
	return topology.RawAddrInfo{Public: []topology.RawAddrPortOverlay{
		{
			RawAddrPort: topology.RawAddrPort{
				Addr:   addr,
				L4Port: l4port,
			},
			OverlayPort: overlayport,
		},
	},
	}
}
