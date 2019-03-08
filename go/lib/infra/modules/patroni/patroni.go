// Copyright 2019 Anapaya Systems.

// Package patroni contains a patroni connection pool implementation. A patroni cluster has at most
// one connection that is writable and multiple connections that can be used for reads. The pool
// will prefer to always return the writable connection so that clients get the most consistent
// data.
//
// Internally the lib uses consul to detect on which URLs patroni runs. Since consul might have an
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
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
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
	LeaderTag  = "master"
	ReplicaTag = "replica"

	DefaultUpdateInterval = 2 * time.Second
	DefaultUpdateTimeout  = 5 * time.Second
)

var (
	// DBErrors is the list of errors that should be considered as an issue with the DB.
	// If one of these errors is reported the connection is considered as unhealthy.
	DBErrors = []error{io.EOF}
)

// Conf contains configuration for patroni.
type Conf struct {
	// DBUser the username for the database.
	DBUser string
	// DBPass the password for the database.
	DBPass string
	// ClusterKey the key under which the patroni services are in consul.
	ClusterKey string
	// UpdateInterval is the interval at which the cluster is periodically refreshed.
	// In case of an error the cluster is immediately refreshed.
	UpdateInterval time.Duration
	// UpdateTimeout is the timeout for the update of the connection pool.
	// Should be at least size(cluster) * time.Second.
	UpdateTimeout time.Duration
}

func (cfg *Conf) InitDefaults() {
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
	return fmt.Sprintf("ConnRef{Node: %s, v: %d}", r.node, r.version)
}

// ReportErr should be called after using this connection reference.
// The err should be the one that was returned when using the connection, it can also be nil.
// The return value indicates whether the error was recognized as a connection error, in which case
// it would make sense for the caller to get a new connection and retry.
func (r *ConnRef) ReportErr(err error) bool {
	connErr := isConnErr(err)
	if connErr {
		r.pool.reportErr(r, err)
	}
	return connErr
}

func isConnErr(err error) bool {
	if err == nil {
		return false
	}
	for _, dbErr := range DBErrors {
		if common.ContainsErr(err, dbErr) {
			return true
		}
	}
	return false
}

// clientConn represents a connection to a specific database host.
type clientConn struct {
	node string
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
	lastNodes     map[string]patroniNode
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
	nodes, err := nodesFromConsul(ctx, cfg.ClusterKey, consulC)
	if err != nil {
		return nil, common.NewBasicError("Failed to load patroni cluster from consul", err)
	}
	nodes = selectReachableNodes(ctx, nodes)
	p.updateFromNodes(nodes)
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
	log.Debug("[patroni] Reporting connection", "node", r.node, "err", err)
	if conn, ok := c.healthyConns[r.node]; ok {
		delete(c.healthyConns, r.node)
		c.unhealthyConns[r.node] = conn
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
	nodes, err := nodesFromConsul(ctx, c.cfg.ClusterKey, c.consulC)
	if err != nil {
		log.Error("[patroni] Failed to fetch nodes from consul, using cached ones", "err", err)
		nodes = c.lastNodes
	}
	nodes = selectReachableNodes(ctx, nodes)
	c.updateFromNodes(nodes)
}

// updateFromNodes updates the connections in the pool with the given nodes.
// The API exposed by each node should be tested for reachability prior to calling this method.
func (c *ConnPool) updateFromNodes(onlineNodes map[string]patroniNode) {
	c.dataMtx.Lock()
	defer c.dataMtx.Unlock()

	healthyConns := make(map[string]*clientConn)
	unhealthyConns := make(map[string]*clientConn)
	// collect all existing connections, so that we can reuse them.
	existingConns := make(map[string]*clientConn, len(c.healthyConns)+len(c.unhealthyConns))
	copyInto(existingConns, c.healthyConns)
	copyInto(existingConns, c.unhealthyConns)
	var leader string
	for node, info := range onlineNodes {
		if info.Role == LeaderTag {
			leader = node
		}
		if conn, ok := existingConns[node]; ok {
			if c.sqlUrlUnchanged(node, info) {
				healthyConns[node] = conn
				continue
			} else {
				log.Info("[patroni] ConnUrl changed", "node", node, "url", info.ConnUrl)
				conn.db.Close()
			}
		}
		log.Info("[patroni] New node, creating conn", "node", node, "url", info.ConnUrl)
		db, err := c.createSqlDB(info)
		if err != nil {
			log.Warn("[patroni] Couldn't establish connection",
				"node", node, "url", info.ConnUrl, "err", err)
			continue
		}
		conn := &clientConn{
			node: node,
			db:   db,
			pool: c,
		}
		if err = db.Ping(); err != nil {
			log.Warn("[patroni] Connection unhealthy", "node", node, "err", err)
			unhealthyConns[node] = conn
		} else {
			healthyConns[node] = conn
		}
	}
	c.leader = leader
	c.healthyConns = healthyConns
	c.unhealthyConns = unhealthyConns
	closeUnreferencedConns(existingConns, healthyConns, unhealthyConns)
	c.version++
	c.lastNodes = onlineNodes
	c.updating = false
	log.Trace("[patroni] Updated connections", "leader", c.leader,
		"healthyCnt", len(c.healthyConns), "unhealthyCnt", len(c.unhealthyConns))
}

func (c *ConnPool) createSqlDB(node patroniNode) (*sql.DB, error) {
	u, err := url.Parse(node.ConnUrl)
	if err != nil {
		return nil, common.NewBasicError("Failed to parse url", err, "url", node.ConnUrl)
	}
	u.User = url.UserPassword(c.cfg.DBUser, c.cfg.DBPass)
	db, err := sql.Open("pgx", u.String())
	if err != nil {
		return nil, err
	}
	return db, nil
}

func (c *ConnPool) sqlUrlUnchanged(node string, info patroniNode) bool {
	oldInfo := c.lastNodes[node]
	return oldInfo.ConnUrl == info.ConnUrl
}

var _ periodic.Task = (*connPoolUpdater)(nil)

// connPoolUpdater is just a wrapper for ConnPool so that we can use it in a periodic.Runner.
type connPoolUpdater struct {
	*ConnPool
}

func (c *connPoolUpdater) Run(ctx context.Context) {
	c.update(ctx)
}

type patroniNode struct {
	// ApiUrl is the patroni API url.
	ApiUrl string `json:"api_url"`
	// ConnUrl is the postgres connection url.
	ConnUrl string `json:"conn_url"`
	// Role is the role of the node.
	Role string
}

func nodesFromConsul(ctx context.Context, clusterKey string,
	consulC *consulapi.Client) (map[string]patroniNode, error) {

	key := fmt.Sprintf("service/%s/members/", clusterKey)
	qo := &consulapi.QueryOptions{}
	qo = qo.WithContext(ctx)
	kvPairs, _, err := consulC.KV().List(key, qo)
	if err != nil {
		return nil, common.NewBasicError("Failed to list members", err)
	}
	infos := make(map[string]patroniNode, len(kvPairs))
	for _, kv := range kvPairs {
		var info patroniNode
		if err := json.Unmarshal(kv.Value, &info); err != nil {
			return nil, common.NewBasicError("Failed to parse", err)
		}
		infos[strings.TrimPrefix(kv.Key, key)] = info
	}
	return infos, nil
}

// selectReachableNodes checks the given nodes from consul against the patroni API. The result only
// contains reachable nodes (as seen from the patroni API).
func selectReachableNodes(ctx context.Context,
	consulNodes map[string]patroniNode) map[string]patroniNode {

	okNodes := make(map[string]patroniNode)
	for node, info := range consulNodes {
		subCtx, cancelF := context.WithTimeout(ctx, 500*time.Millisecond)
		status := checkNode(subCtx, node, info)
		cancelF()
		switch status {
		case leader:
			info.Role = LeaderTag
			okNodes[node] = info
		case replica:
			info.Role = ReplicaTag
			okNodes[node] = info
		}
	}
	return okNodes
}

type nodeType string

const (
	leaderNode  nodeType = LeaderTag
	replicaNode nodeType = ReplicaTag
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

func checkNode(ctx context.Context, nodeName string, info patroniNode) nodeStatus {
	if info.Role == LeaderTag || info.Role == ReplicaTag {
		return checkNodeStatus(ctx, nodeName, nodeType(info.Role), info.ApiUrl, true)
	}
	log.Debug("[patroni] Node role not handled, marking unhealthy",
		"role", info.Role, "node", nodeName)
	return unhealthy
}

func checkNodeStatus(ctx context.Context, nodeName string, expectedType nodeType,
	apiUrl string, checkOtherType bool) nodeStatus {

	url := strings.Replace(apiUrl, "patroni", expectedType.String(), -1)
	resp, err := ctxhttp.Get(ctx, nil, url)
	if err != nil {
		log.Warn("[patroni] Failed to check node", "url", url, "err", err)
		return unhealthy
	}
	if resp.StatusCode == http.StatusOK {
		return expectedType.nodeStatusOk()
	}
	if resp.StatusCode == http.StatusServiceUnavailable {
		if checkOtherType {
			return checkNodeStatus(ctx, nodeName, expectedType.other(), apiUrl, false)
		}
		log.Debug("[patroni] Got ServiceUnavailable(503) for both master and replica endpoint",
			"node", nodeName)
		// consider this one unhealthy.
		return unhealthy
	}
	log.Warn("[patroni] Unknown status code", "node", nodeName, "statusCode", resp.StatusCode)
	return unhealthy
}

// closeUnreferencedConns closes DB connections that are in allConns but neither in healthy nor in
// unhealthy conns.
func closeUnreferencedConns(allCons, healthyConns, unhealthyConns map[string]*clientConn) {
	for key, conn := range allCons {
		_, healthy := healthyConns[key]
		_, unhealthy := unhealthyConns[key]
		if !healthy && !unhealthy {
			log.Info("[patroni] Connection is gone", "key", key)
			conn.db.Close()
		}
	}
}

// copyInto copies the values from src into dst.
func copyInto(dst, src map[string]*clientConn) {
	for key, conn := range src {
		dst[key] = conn
	}
}
