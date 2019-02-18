// Copyright 2019 Anapaya Systems.

// Package patroni contains a patroni connection pool implementation. A patroni cluster has at most
// one connection that is writable and multiple connections that can be used for reads. The pool
// will prefer to always return the writable connection so that clients get the most consistent
// data.
//
// Internally the lib uses consul to detect on which IPs patroni runs. Since consul might have an
// outdated view on the state of the cluster, the lib uses the patroni API to verify the information
// from consul.
//
// Users should request a fresh connection per request. The returned error of using the connection
// should always be reported.
// Example:
//  pool, err := NewConnPool(...)
//  // deal with error
//  conn := pool.WriteConn()
//  if conn == nil {
//    // no write connection
//  }
//  _, err := conn.DB().Exec(.....)
//  conn.ReportErr(err) // always do that.
package patroni

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	consulapi "github.com/hashicorp/consul/api"
	// pgx postgres driver
	_ "github.com/jackc/pgx/stdlib"
	"golang.org/x/net/context/ctxhttp"

	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/lib/log"
	"github.com/scionproto/scion/go/lib/periodic"
)

const (
	LeaderTag = "master"

	DefaultPatroniPort    = 8008
	DefaultUpdateInterval = 2 * time.Second
	DefaultUpdateTimeout  = 5 * time.Second
)

// Conf contains configuration for patroni.
type Conf struct {
	// PatroniPort the port on which the patroni API listens.
	PatroniPort int
	// ClusterKey the key under which the patroni services are in consul.
	ClusterKey string
	// ConnString in the form: postgresql://user:password@host:port
	// "host" must always be contained literally, the library will insert the correct IP.
	ConnString string
	// UpdateInterval is the interval at which the cluster is periodically refreshed.
	// In case of an error the cluster is immediately refreshed.
	UpdateInterval time.Duration
	// UpdateTimeout is the timeout for the update of the connection pool.
	// Should be at least size(cluster) * time.Second.
	UpdateTimeout time.Duration
}

func (cfg *Conf) InitDefaults() {
	if cfg.PatroniPort == 0 {
		cfg.PatroniPort = DefaultPatroniPort
	}
	if cfg.UpdateInterval == 0 {
		cfg.UpdateInterval = DefaultUpdateInterval
	}
	if cfg.UpdateTimeout == 0 {
		cfg.UpdateTimeout = DefaultUpdateTimeout
	}
}

// ConnRef is a reference to a connection.
// A ConnRef should only be used once (or for only for a short time)
// and any error should be reported.
type ConnRef struct {
	version uint64
	*clientConn
}

// DB returns the DB connection.
func (r *ConnRef) DB() *sql.DB {
	return r.clientConn.db
}

func (r *ConnRef) String() string {
	if r == nil {
		return "ConnRef<nil>"
	}
	return fmt.Sprintf("ConnRef{IP: %s, v: %d}", r.ip, r.version)
}

// ReportErr should be called after using this connection reference.
// The err should be the one that was returned when using the connection, it can also be nil.
func (r *ConnRef) ReportErr(err error) {
	if err == nil {
		return
	}
	// TODO(lukedirtwalker): check if the error is really a connection error and not a user error.
	// currently I'm not sure yet how to distinguish the cases.
	r.pool.reportErr(r, err)
}

// clientConn represents a connection to a specific database host.
type clientConn struct {
	ip   string
	pool *ConnPool
	db   *sql.DB
}

// ConnPool is a pool of patroni connections.
type ConnPool struct {
	consulC        *consulapi.Client
	cfg            Conf
	dataMtx        sync.RWMutex
	version        uint64
	leader         string
	healthyConns   map[string]*clientConn
	unhealthyConns map[string]*clientConn

	refreshRunner *periodic.Runner
	updating      bool
	lastIPs       *ips
}

// NewConnPool creates a new patroni connection pool.
// If the initial cluster detection fails this returns an error.
func NewConnPool(consulC *consulapi.Client, cfg Conf) (*ConnPool, error) {
	cfg.InitDefaults()
	p := &ConnPool{
		consulC: consulC,
		cfg:     cfg,
	}
	ctx, cancelF := context.WithTimeout(context.Background(), time.Second)
	defer cancelF()
	ips, err := ipsFromConsul(ctx, cfg.ClusterKey, consulC)
	if err != nil {
		return nil, common.NewBasicError("Failed to load patroni cluster from consul", err)
	}
	ips = ipsFromPatroni(ctx, ips, cfg)
	p.updateFromIPs(ips)
	p.refreshRunner = periodic.StartPeriodicTask(&connPoolUpdater{ConnPool: p},
		periodic.NewTicker(cfg.UpdateInterval), cfg.UpdateTimeout)
	return p, nil
}

// WriteConn returns a reference to a connection that can be used to write to the DB, or nil
// if there is none available.
func (c *ConnPool) WriteConn() *ConnRef {
	c.dataMtx.RLock()
	defer c.dataMtx.RUnlock()
	return c.writeConn()
}

// ReadConn returns a reference to a connection that can be used to read from the DB, or nil
// if there is none available.
func (c *ConnPool) ReadConn() *ConnRef {
	c.dataMtx.RLock()
	defer c.dataMtx.RUnlock()
	// Always prefer to return the "leader" connection.
	if wConn := c.writeConn(); wConn != nil {
		return wConn
	}
	// Return the first available healthy conn,
	for _, conn := range c.healthyConns {
		return &ConnRef{
			version:    c.version,
			clientConn: conn,
		}
	}
	// or nil if none is found.
	return nil
}

// Close closes the connection pool and all opened connections.
func (c *ConnPool) Close() {
	c.refreshRunner.Stop()
	// now close all connections.
	c.dataMtx.Lock()
	defer c.dataMtx.Unlock()
	closeConns := func(conns map[string]*clientConn) {
		for _, conn := range conns {
			if err := conn.db.Close(); err != nil {
				log.Warn("[patroni] Error while closing connection", "err", err)
			}
		}
	}
	closeConns(c.unhealthyConns)
	closeConns(c.healthyConns)
}

// returns a ref to the connection for writing or nil. Callers should acquire the read lock.
func (c *ConnPool) writeConn() *ConnRef {
	conn, ok := c.healthyConns[c.leader]
	if !ok {
		return nil
	}
	return &ConnRef{
		version:    c.version,
		clientConn: conn,
	}
}

func (c *ConnPool) reportErr(r *ConnRef, err error) {
	c.dataMtx.Lock()
	defer c.dataMtx.Unlock()
	// if the connection ref is old we don't care.
	if r.version < c.version {
		return
	}
	if conn, ok := c.healthyConns[r.ip]; ok {
		delete(c.healthyConns, r.ip)
		c.unhealthyConns[r.ip] = conn
	}
	// If we are already updating no need to trigger again.
	if c.updating {
		return
	}
	// make sure we only trigger once.
	c.updating = true
	c.refreshRunner.TriggerRun()
}

func (c *ConnPool) update(ctx context.Context) {
	c.dataMtx.Lock()
	if !c.updating {
		c.updating = true
	}
	c.dataMtx.Unlock()
	ips, err := ipsFromConsul(ctx, c.cfg.ClusterKey, c.consulC)
	if err != nil {
		log.Error("[patroni] Failed to fetch IPs from consul, using cached ones", "err", err)
		ips = c.lastIPs
	}
	ips = ipsFromPatroni(ctx, ips, c.cfg)
	c.updateFromIPs(ips)
}

func hasLeaderTag(tags []string) bool {
	for _, t := range tags {
		if t == LeaderTag {
			return true
		}
	}
	return false
}

// updateFromIPs updates the connections in the pool with the given ips.
// The ips should be verified beforehand.
func (c *ConnPool) updateFromIPs(onlineIPs *ips) {
	c.dataMtx.Lock()
	defer c.dataMtx.Unlock()

	healthyConns := make(map[string]*clientConn)
	unhealthyConns := make(map[string]*clientConn)

	// collect all existing connections, so that we can reuse them.
	existingConns := make(map[string]*clientConn, len(c.healthyConns)+len(c.unhealthyConns))
	for ip, conn := range c.healthyConns {
		existingConns[ip] = conn
	}
	for ip, conn := range c.unhealthyConns {
		existingConns[ip] = conn
	}
	// now mark ips we found again as healthy.
	for _, ip := range append(onlineIPs.replicaIPs, onlineIPs.leaderIP) {
		if ip == "" {
			continue
		}
		if conn, ok := existingConns[ip]; ok {
			healthyConns[ip] = conn
			continue
		}
		log.Info("[patroni] New IP, creating conn", "ip", ip)
		db, err := createSqlDB(ip, c.cfg)
		conn := &clientConn{
			ip:   ip,
			db:   db,
			pool: c,
		}
		if err != nil {
			log.Warn("[patroni] Connection unhealthy", "ip", ip, "err", err)
			unhealthyConns[ip] = conn
		} else {
			healthyConns[ip] = conn
		}
	}
	c.leader = onlineIPs.leaderIP
	c.healthyConns = healthyConns
	c.unhealthyConns = unhealthyConns
	// close no longer referenced conns
	for ip, conn := range existingConns {
		_, healthy := healthyConns[ip]
		_, unhealthy := unhealthyConns[ip]
		if !healthy && !unhealthy {
			log.Info("[patroni] Connection is gone", "ip", ip)
			conn.db.Close()
		}
	}
	c.version++
	c.lastIPs = onlineIPs
	c.updating = false
	log.Trace("[patroni] Updated connections",
		"healthyCnt", len(c.healthyConns), "unhealthyCnt", len(c.unhealthyConns))
}

func createSqlDB(ip string, cfg Conf) (*sql.DB, error) {
	if ip == "" {
		return nil, nil
	}
	cs := connString(ip, cfg)
	// TODO(lukedirtwalker): Remove password from output.
	log.Info("[patroni] Connecting to postgres DB", "db", cs)
	db, err := sql.Open("pgx", cs)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, common.NewBasicError("Initial DB ping failed, connection broken?", err)
	}
	return db, nil
}

func connString(ip string, cfg Conf) string {
	return strings.Replace(cfg.ConnString, "host", ip, -1)
}

var _ periodic.Task = (*connPoolUpdater)(nil)

// connPoolUpdater is just a wrapper for ConnPool so that we can use it in a periodic.Runner.
type connPoolUpdater struct {
	*ConnPool
}

func (c *connPoolUpdater) Run(ctx context.Context) {
	c.update(ctx)
}

// ips is a group of patroni IP addresses returned from consul.
type ips struct {
	leaderIP   string
	replicaIPs []string
}

// ipsFromConsul loads all known patroni IPs from consul.
// Note that this will contain the view of consul which might be outdated.
// Use ipsFromPatroni to get a more up to date view.
func ipsFromConsul(ctx context.Context,
	clusterKey string, consulC *consulapi.Client) (*ips, error) {

	qo := &consulapi.QueryOptions{}
	qo = qo.WithContext(ctx)
	svcList, _, err := consulC.Catalog().Service(clusterKey, "", qo)
	if err != nil {
		return nil, common.NewBasicError("Failed to list members", err)
	}
	if len(svcList) == 0 {
		return nil, common.NewBasicError("No members", nil)
	}
	conns := &ips{}
	for _, svc := range svcList {
		if hasLeaderTag(svc.ServiceTags) {
			conns.leaderIP = svc.ServiceAddress
		} else {
			conns.replicaIPs = append(conns.replicaIPs, svc.ServiceAddress)
		}
	}
	return conns, nil
}

// ipsFromPatroni verifies the ips returned from consul. To do this it contacts the patroni server
// on each ip and verifies that it actually has the status as claimed by consul.
// This is needed because consul might have a slightly out of date view on the cluster.
func ipsFromPatroni(ctx context.Context, consulIPs *ips, cfg Conf) *ips {
	verifiedIPs := &ips{}
	if consulIPs.leaderIP != "" {
		subCtx, cancelF := context.WithTimeout(ctx, 500*time.Millisecond)
		status := checkNodeStatus(subCtx, consulIPs.leaderIP, leaderNode, cfg, true)
		cancelF()
		switch status {
		case leader:
			verifiedIPs.leaderIP = consulIPs.leaderIP
		case replica:
			verifiedIPs.replicaIPs = append(verifiedIPs.replicaIPs, consulIPs.leaderIP)
		}
	}
	for _, ip := range consulIPs.replicaIPs {
		subCtx, cancelF := context.WithTimeout(ctx, 500*time.Millisecond)
		status := checkNodeStatus(subCtx, ip, replicaNode, cfg, true)
		cancelF()
		switch status {
		case leader:
			if verifiedIPs.leaderIP != "" {
				log.Error("[patroni] detected 2 leader nodes",
					"ip1", verifiedIPs.leaderIP, "ip2", ip)
				return nil
			}
			verifiedIPs.leaderIP = ip
		case replica:
			verifiedIPs.replicaIPs = append(verifiedIPs.replicaIPs, ip)
		}
	}
	return verifiedIPs
}

type nodeType string

const (
	leaderNode  nodeType = "master"
	replicaNode nodeType = "replica"
)

func (t nodeType) String() string {
	return string(t)
}

func (t nodeType) other() nodeType {
	if t == leaderNode {
		return replicaNode
	}
	return leaderNode
}

func (t nodeType) nodeStatusOk() nodeStatus {
	if t == leaderNode {
		return leader
	}
	return replica
}

type nodeStatus uint32

const (
	unhealthy nodeStatus = iota
	leader
	replica
)

func checkNodeStatus(ctx context.Context, ip string, expectedType nodeType,
	cfg Conf, checkOtherType bool) nodeStatus {

	url := fmt.Sprintf("http://%s:%d/%s", ip, cfg.PatroniPort, expectedType)
	resp, err := ctxhttp.Get(ctx, nil, url)
	if err != nil {
		log.Warn("[patroni] Failed to check ip",
			"ip", ip, "expectedType", expectedType, "err", err)
		return unhealthy
	}
	if resp.StatusCode == http.StatusOK {
		return expectedType.nodeStatusOk()
	}
	if resp.StatusCode == http.StatusServiceUnavailable {
		if checkOtherType {
			return checkNodeStatus(ctx, ip, expectedType.other(), cfg, false)
		}
		log.Warn("[patroni] Got ServiceUnavailable(503) for both master and replica endpoint",
			"ip", ip)
		// consider this one unhealthy.
		return unhealthy
	}
	log.Warn("[patroni] Unknown status code", "ip", ip, "statusCode", resp.StatusCode)
	return unhealthy
}
