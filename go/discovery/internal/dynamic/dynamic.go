// Copyright 2017 Anapaya Systems

package dynamic

import (
	"context"
	"encoding/json"
	"net"
	"time"

	consulapi "github.com/hashicorp/consul/api"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/scionproto/scion/go/discovery/internal/metrics"
	"github.com/scionproto/scion/go/discovery/internal/static"
	"github.com/scionproto/scion/go/discovery/internal/util"
	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/lib/consul"
	"github.com/scionproto/scion/go/lib/log"
	"github.com/scionproto/scion/go/lib/periodic"
	"github.com/scionproto/scion/go/lib/topology"
)

var TopoFull *util.AtomicTopo
var TopoLimited *util.AtomicTopo

const (
	ErrConn              = "connect-error"
	ErrParseBaseTopo     = "basetopo-parse-error"
	ErrServiceUpdate     = "service-update-error"
	ErrMarshalFull       = "marshall-full-error"
	ErrMarshalReduced    = "marshall-reduced-error"
	ErrFetch             = "fetch-failure"
	ErrServiceEmpty      = "service-empty"
	ErrOwnServiceMissing = "own-service-missing"
	PartialUpdate        = "partial-update"
	Success              = "success"
)

func init() {
	TopoFull = &util.AtomicTopo{}
	TopoFull.Store([]byte{})
	TopoLimited = &util.AtomicTopo{}
	TopoLimited.Store([]byte{})
}

var _ periodic.Task = (*Updater)(nil)

type Updater struct {
	ID        string
	SvcPrefix string
	Client    *consulapi.Client
	TTL       time.Duration
}

// Run fetches all available services from consul and updates the dynamic topology.
func (u *Updater) Run(ctx context.Context) {
	defer updateTimeCounter(time.Now(), metrics.TotalDynamicUpdateTime)
	labels := prometheus.Labels{"result": Success}
	defer incCounterVec(labels, metrics.TotalConsulUpdates)

	if err := u.update(ctx); err != nil {
		log.Error("Error while updating dynamic topology from consul", "err", err)
		labels["result"] = metrics.GetErrorLabel(err)
	}
}

// update fetches all available services from consul and updates the dynamic topology.
func (u *Updater) update(ctx context.Context) error {
	rt := &topology.RawTopo{}
	// Get the base topo from the static part of DS so the two versions
	// agree as much as is useful.
	if err := json.Unmarshal(static.DiskTopo.Load(), rt); err != nil {
		return metrics.NewError(ErrParseBaseTopo,
			common.NewBasicError("Unable to re-parse topology", err))
	}
	// Update topology from consul service.
	ok := u.updateServices(ctx, rt)
	// Finalize topology and store it. We want to store it even if not
	// all services were set. They will simply fallback to static value.
	if err := u.store(rt); err != nil {
		return err
	}
	// Do not count update where not all services were fetched as successful.
	if !ok {
		return metrics.NewError(PartialUpdate,
			common.NewBasicError("Only partial update possible", nil))
	}
	log.Trace("Successfully updated dynamic topology from consul")
	metrics.ConsulLastUpdate.SetToCurrentTime()
	return nil
}

// store updates the timestamps and stores the topology as the full and reduced version.
func (u *Updater) store(rt *topology.RawTopo) error {
	u.updateTimestamps(rt)
	// Set the full topology.
	if err := marshallAndUpdate(rt, TopoFull); err != nil {
		return metrics.NewError(ErrMarshalFull,
			common.NewBasicError("Unable to marshal full topology", err))
	}
	// Bind address and and non-public services are stripped.
	topology.StripBind(rt)
	topology.StripServices(rt)
	// Set the reduced topology.
	if err := marshallAndUpdate(rt, TopoLimited); err != nil {
		return metrics.NewError(ErrMarshalFull,
			common.NewBasicError("Unable to marshal reduced topology", err))
	}
	return nil
}

// updateTimestamps updates the timestamps of the raw topo.
func (u *Updater) updateTimestamps(rt *topology.RawTopo) {
	ts := time.Now()
	rt.Timestamp = ts.Unix()
	rt.TimestampHuman = ts.Format(time.RFC3339)
	rt.TTL = uint32(u.TTL / time.Second)
}

// updateServices updates the services in the raw topology. If a service cannot
// be fetched from consul, the original service map is kept.
func (u *Updater) updateServices(ctx context.Context, rt *topology.RawTopo) bool {
	defer updateTimeCounter(time.Now(), metrics.TotalConsulUpdateTime)

	ok, allOk := true, true
	rt.BeaconService, ok = u.fillService(ctx, consul.BS, rt.BeaconService)
	allOk = allOk && ok
	rt.CertificateService, ok = u.fillService(ctx, consul.CS, rt.CertificateService)
	allOk = allOk && ok
	rt.PathService, ok = u.fillService(ctx, consul.PS, rt.PathService)
	allOk = allOk && ok
	rt.SIG, ok = u.fillService(ctx, consul.SIG, rt.SIG)
	allOk = allOk && ok
	rt.DiscoveryService, ok = u.fillDiscoveryService(ctx, rt.DiscoveryService)
	return allOk && ok
}

func (u *Updater) fillDiscoveryService(ctx context.Context,
	fallback map[string]*topology.RawSrvInfo) (map[string]*topology.RawSrvInfo, bool) {

	labels := prometheus.Labels{"result": Success, "service": string(consul.DS)}
	defer incCounterVec(labels, metrics.TotalConsulServiceUpdates)

	service, err := u.getService(ctx, consul.DS)
	if err != nil {
		log.Warn("Could not fetch service entries from Consul, using fallback",
			"servicetype", consul.DS, "err", err)
		labels["result"] = ErrFetch
		return fallback, false
	}
	succ := true
	if _, ok := service[u.ID]; !ok {
		log.Warn("Own service not passing on Consul, using fallback")
		labels["result"] = ErrOwnServiceMissing
		succ = false
	}
	// Ensure that the local entry is consistent with the static topology.
	service[u.ID] = fallback[u.ID]
	return service, succ
}

// fillService fetches the service info from consul and returns it. If consul is not
// reachable, or the response is empty, fallback is returned.
func (u *Updater) fillService(ctx context.Context, svcType consul.SvcType,
	fallback map[string]*topology.RawSrvInfo) (map[string]*topology.RawSrvInfo, bool) {

	labels := prometheus.Labels{"result": Success, "service": string(svcType)}
	defer incCounterVec(labels, metrics.TotalConsulServiceUpdates)

	service, err := u.getService(ctx, svcType)
	if err != nil {
		log.Warn("Could not fetch service entries from Consul, using fallback",
			"servicetype", svcType, "err", err)
		labels["result"] = ErrFetch
		return fallback, false
	}
	if len(service) == 0 {
		log.Warn("Service listing is empty on Consul, using fallback", "servicetype", svcType)
		labels["result"] = ErrServiceEmpty
		return fallback, false
	}
	return service, true
}

// getService fetches the service info from consul and parses it.
func (u *Updater) getService(ctx context.Context,
	svcType consul.SvcType) (map[string]*topology.RawSrvInfo, error) {

	q := (&consulapi.QueryOptions{RequireConsistent: true}).WithContext(ctx)
	log.Info("Requesting service", "svc", u.SvcPrefix+string(svcType))
	svcs, _, err := u.Client.Health().Service(u.SvcPrefix+string(svcType), "", true, q)
	if err != nil {
		return nil, err
	}
	services := make(map[string]*topology.RawSrvInfo, len(svcs))
	for _, svc := range svcs {
		id, rsi, err := parseConsulService(svc)
		if err != nil {
			log.Warn("Unable to parse service, skipping", "err", err, "id", id)
			continue
		}
		services[id] = rsi
	}
	return services, nil
}

// parseConsulService parses a single ServiceEntry.
func parseConsulService(svc *consulapi.ServiceEntry) (string, *topology.RawSrvInfo, error) {
	if svc.Service.ID == "" {
		return "Unset", nil, common.NewBasicError("ServiceID not set", nil)
	}
	if svc.Service.Port == 0 {
		return svc.Service.ID, nil, common.NewBasicError("ServicePort not set", nil)
	}
	ip := net.ParseIP(svc.Service.Address)
	if ip == nil {
		return svc.Service.ID, nil, common.NewBasicError("Unable to parse service address", nil,
			"addr", svc.Service.Address)
	}
	return svc.Service.ID, rsiFromAddrPortOverlay(ip, svc.Service.Port, 0), nil
}

// rsiFromAddrPortOverlay creates the raw server info.
func rsiFromAddrPortOverlay(ip net.IP, l4port int, overlayport int) *topology.RawSrvInfo {
	addrs := make(topology.RawAddrMap)
	rpbo := &topology.RawPubBindOverlay{
		Public: topology.RawAddrPortOverlay{
			RawAddrPort: topology.RawAddrPort{
				Addr:   ip.String(),
				L4Port: l4port,
			},
			OverlayPort: overlayport,
		},
	}
	if ip.To4() != nil {
		addrs["IPv4"] = rpbo
	} else {
		addrs["IPv6"] = rpbo
	}
	return &topology.RawSrvInfo{
		Addrs: addrs,
	}
}

// marshallAndUpdate marshalls the raw topology and stores it atomically.
func marshallAndUpdate(rt *topology.RawTopo, topo *util.AtomicTopo) error {
	b, err := util.MarshalToJSON(rt)
	if err != nil {
		return err
	}
	topo.Store(b)
	return nil
}

// updateTimeCounter updates the counter with the time difference
// between start and the current time.
func updateTimeCounter(start time.Time, c prometheus.Counter) {
	c.Add(float64(time.Since(start).Seconds()))
}

// incCounterVec increments the counter vector with the provided labels.
func incCounterVec(labels prometheus.Labels, c *prometheus.CounterVec) {
	c.With(labels).Inc()
}
