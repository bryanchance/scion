// Copyright 2018 Anapaya Systems

package consul

import (
	"context"
	"sync"
	"time"

	consulapi "github.com/hashicorp/consul/api"

	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/lib/log"
	"github.com/scionproto/scion/go/lib/periodic"
)

const (
	DefaultLEInterval = time.Second
	DefaultLETimeout  = 5 * time.Minute
)

type LeaderElectorConf struct {
	// Interval configures how much time between run loops should pass.
	// If not set the default DefaultLEInterval is used.
	Interval time.Duration
	// Timeout indicates how long the leader elector should blockingly wait for leader changes.
	// If not set the default DefaultLETimeout is used.
	Timeout time.Duration
	// LockDelay is the lock delay passed to the consul API.
	// If set to zero the default of the consul API is used.
	// See also: https://www.consul.io/api/session.html#lockdelay
	LockDelay time.Duration
	// SessionTTL is the TTL of the session. By default this should be 10s.
	// See https://www.consul.io/api/session.html#ttl
	SessionTTL string
}

func (c *LeaderElectorConf) initDefaults() {
	if c.Interval == 0 {
		c.Interval = DefaultLEInterval
	}
	if c.Timeout == 0 {
		c.Timeout = DefaultLETimeout
	}
	if c.SessionTTL == "" {
		c.SessionTTL = "10s" // minimum (see https://www.consul.io/api/session.html#ttl)
	}
}

// LeaderElector handles leader election for a specific key.
type LeaderElector interface {
	// IsLeader returns if we are currently the leader.
	IsLeader() bool
	// Release releases the leadership lock at consul explicitly.
	Release()
	// Stops releases the leader and stops the process.
	Stop()
	// Key returns the key for which this LeaderElector is running.
	Key() string
}

// StartLeaderElector starts a new LeaderElector for the given key.
func StartLeaderElector(c *consulapi.Client, key string, conf LeaderElectorConf) LeaderElector {
	conf.initDefaults()
	le := &leaderElector{
		key:    key,
		c:      c,
		logger: log.New("key", key),
		cfg:    conf,
	}
	return &leaderElectorRunner{
		leaderElector: le,
		runner: periodic.StartPeriodicTask(le,
			periodic.NewTicker(conf.Interval), conf.Timeout),
	}
}

var _ (periodic.Task) = (*leaderElector)(nil)

// leaderElector handles leader election for a specific key.
// leaderElector implements the periodic.Task interface.
// The recommendation is to use a short ticker interval, e.g. 1s, but a high timeout, e.g. 5m.
// The timeout is used to blockingly wait for a leader change.
//
// Implementation details:
// The implementation follows: https://www.consul.io/docs/guides/leader-election.html
// We use consul to acquire a lock on the KV store for the key.
// The logic is as follows:
// 1. Create a session or check the existing one is valid.
// 2. Query the KV store if a session holds the lock.
// 2.1. If the lock is not held: try to acquire it; goto 1.
// 2.2. If the lock is held: blockingly wait until the KV entry changes; goto 1.
type leaderElector struct {
	// key is the kv key we use to lock on, this is constant.
	key string
	// c is the consul client, this is constant.
	c *consulapi.Client
	// logger is the logger, this is constant.
	logger log.Logger
	// sessM is a mutex used to protect sessId and leader variables.
	// Locking should only happen in public methods.
	sessM            sync.RWMutex
	sessId           string
	leader           bool
	sessRefreshClose chan struct{}
	// modIdx is for waiting on modifications
	modIdx uint64
	cfg    LeaderElectorConf
}

func (le *leaderElector) Key() string {
	return le.key
}

func (le *leaderElector) IsLeader() bool {
	le.sessM.RLock()
	defer le.sessM.RUnlock()
	return le.leader
}

func (le *leaderElector) Release() {
	le.sessM.Lock()
	defer le.sessM.Unlock()
	le.release()
}

// Run implements the periodic.Task interface.
func (le *leaderElector) Run(ctx context.Context) {
	if _, ok := ctx.Deadline(); !ok {
		le.logger.Error("[LeaderElector] ctx needs deadline.")
		return
	}
	le.sessM.Lock()
	if err := le.checkSession(ctx); err != nil {
		le.clearCurrentSession()
		le.sessM.Unlock()
		// TODO(lukedirtwalker): We should create metrics for errors to alert on.
		le.logger.Error("[LeaderElector] Failed to check session", "err", err)
		return
	}
	le.sessM.Unlock()
	var sessId string
	var err error
	sessId, le.modIdx, err = le.lockHolderSess(ctx)
	if err != nil {
		le.sessM.Lock()
		le.loseLeader()
		le.sessM.Unlock()
		le.logger.Error("[LeaderElector] Failed to get lockholder session", "err", err)
		// retry
		return
	}
	le.sessM.Lock()
	defer le.sessM.Unlock()
	if le.leader && le.sessId != sessId {
		le.loseLeader()
		// do not immediately retry, we need to check our session first,
		// most likely it is no longer valid.
		return
	}
	if sessId == "" {
		le.leader = le.tryAcquire()
		// make sure we don't wait next time we check the lock holder session.
		le.modIdx = 0
	}
}

// checkSession checks if the current session is valid, and creates a new one if not.
func (le *leaderElector) checkSession(ctx context.Context) error {
	var err error
	var sess *consulapi.SessionEntry
	if le.sessId != "" {
		le.logger.Trace("[LeaderElector] Checking session")
		qo := (&consulapi.QueryOptions{}).WithContext(ctx)
		sess, _, err = le.c.Session().Info(le.sessId, qo)
		if err != nil {
			return common.NewBasicError("Failed to check session state", err)
		}
	}
	if sess == nil {
		le.clearCurrentSession()
		return le.createSession(ctx)
	}
	return nil
}

func (le *leaderElector) clearCurrentSession() {
	le.loseLeader()
	if le.sessId != "" {
		close(le.sessRefreshClose)
		le.modIdx = 0
		le.sessId = ""
	}
}

func (le *leaderElector) createSession(ctx context.Context) error {
	var err error
	se := &consulapi.SessionEntry{
		Name:      le.key,
		TTL:       le.cfg.SessionTTL,
		LockDelay: le.cfg.LockDelay,
	}
	le.sessId, _, err = le.c.Session().Create(se, (&consulapi.WriteOptions{}).WithContext(ctx))
	if err != nil {
		// TODO(lukedirtwalker): We need a way to notify
		// the parent if the error happens too many times.
		return common.NewBasicError("Failed to create session", err)
	}
	le.logger.Info("[LeaderElector] Consul session created", "sessId", le.sessId)
	le.startPeriodicSessRefresh(le.cfg.SessionTTL)
	return nil
}

// startPeriodicSessRefresh starts a go routine to periodically refresh the session.
func (le *leaderElector) startPeriodicSessRefresh(initialTTL string) {
	le.sessRefreshClose = make(chan struct{})
	go func() {
		defer log.LogPanicAndExit()
		err := le.c.Session().RenewPeriodic(initialTTL, le.sessId, nil, le.sessRefreshClose)
		if err == consulapi.ErrSessionExpired {
			// if the session expired we are done here.
			le.logger.Debug("[LeaderElector] Session expired")
			return
		}
		if err != nil {
			le.logger.Error("[LeaderElector] Session renewal failed", "err", err)
		}
	}()
}

// lockHolderSess returns the session Id and modIdx of the current lock holder.
// This function blocks as long as le.modIdx is equal to the current modified index in consul.
// The context is used to set a timeout on the blocking, therefore it must have a deadline.
func (le *leaderElector) lockHolderSess(ctx context.Context) (string, uint64, error) {
	var waitTime time.Duration
	deadline, ok := ctx.Deadline()
	now := time.Now()
	if ok && deadline.After(now) {
		waitTime = deadline.Sub(now) - 100*time.Millisecond
		if waitTime < 0 {
			// Use a non zero value to prevent the default.
			waitTime = time.Nanosecond
		}
	}
	q := &consulapi.QueryOptions{
		WaitTime:  waitTime,
		WaitIndex: le.modIdx,
	}
	// do not set the context on q, the wait time should roughly enforce it.
	kvp, _, err := le.c.KV().Get(le.key, q)
	if err != nil {
		return "", 0, common.NewBasicError("Failed to get KV", err)
	}
	if kvp == nil {
		return "", 0, nil
	}
	return kvp.Session, kvp.ModifyIndex, nil
}

// tryAcquire tries to acquire the leadership lock.
func (le *leaderElector) tryAcquire() bool {
	le.logger.Trace("[LeaderElector] Try to be leader")
	kvPair := &consulapi.KVPair{
		Key:     le.key,
		Session: le.sessId,
	}
	acquired, _, err := le.c.KV().Acquire(kvPair, nil)
	if err != nil {
		le.logger.Error("[LeaderElector] Error while acquiring leader lock", "err", err)
		return false
	}
	if acquired {
		le.logger.Info("[LeaderElector] Became leader")
	} else {
		le.logger.Trace("[LeaderElector] Not leader")
	}
	return acquired
}

func (le *leaderElector) loseLeader() {
	if le.leader {
		le.logger.Info("[LeaderElector] Lost leader")
		le.leader = false
	}
}

func (le *leaderElector) release() {
	if !le.leader {
		return
	}
	le.leader = false
	kvp := &consulapi.KVPair{
		Key:     le.key,
		Session: le.sessId,
	}
	ok, _, err := le.c.KV().Release(kvp, nil)
	le.logger.Info("[LeaderElector] Released leadership", "ok", ok, "err", err)
}

type leaderElectorRunner struct {
	*leaderElector
	runner *periodic.Runner
}

// Stop stops leader election. It releases leader and stops periodically refreshing the session.
func (rle *leaderElectorRunner) Stop() {
	rle.runner.Kill()
	// now that we know that Run is no longer called we don't need locking:
	rle.release()
	rle.clearCurrentSession()
}
