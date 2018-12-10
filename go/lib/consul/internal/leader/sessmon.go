// Copyright 2018 Anapaya Systems

package leader

import (
	"context"
	"time"

	consulapi "github.com/hashicorp/consul/api"

	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/lib/consul/consulconfig"
	"github.com/scionproto/scion/go/lib/log"
)

// sessMon manages the consul session.
// It creates a new session if there isn't one or the current session is invalidated.
// Once it has a session it makes sure that this session is periodically refreshed.
type sessMon struct {
	// key is the kv key we use to lock on, this is constant.
	key string
	// c is the consul client, this is constant.
	c                *consulapi.Client
	sessRefreshClose chan struct{}
	ctx              context.Context
	cancelF          context.CancelFunc
	logger           log.Logger
	cfg              consulconfig.LeaderElectorConf
	ln               *notifier
}

func StartSessMon(key string, c *consulapi.Client, cfg consulconfig.LeaderElectorConf) *sessMon {
	cfg.InitDefaults()
	s := &sessMon{
		key:    key,
		c:      c,
		logger: log.New("key", key),
		cfg:    cfg,
		ln: &notifier{
			acquired: cfg.AcquiredLeader,
			lost:     cfg.LostLeader,
		},
	}
	s.ctx, s.cancelF = context.WithCancel(context.Background())
	go func() {
		defer log.LogPanicAndExit()
		s.run()
	}()
	return s
}

func (s *sessMon) Stop() {
	s.cancelF()
}

func (s *sessMon) Key() string {
	return s.key
}

func (s *sessMon) run() {
	for {
		sessId, rClose, err := s.createSession()
		if err != nil {
			if isContextErr(s.ctx, s.logger) {
				return
			}
			s.logger.Error("[LeaderElector] Error during create session", "err", err)
			sleep(s.ctx, time.Second)
			continue
		}
		lm := startLeaderMon(s.key, s.c, sessId, s.ln)
		if err := s.checkSessionValid(sessId); err != nil {
			if isContextErr(s.ctx, s.logger) {
				lm.Cancel()
				close(rClose)
				return
			}
			s.logger.Error("[LeaderElector] Error during session check", "err", err)
		}
		lm.Cancel()
		close(rClose)
	}
}

func (s *sessMon) createSession() (string, chan struct{}, error) {
	se := &consulapi.SessionEntry{
		Name:      s.key,
		TTL:       s.cfg.SessionTTL,
		LockDelay: s.cfg.LockDelay,
	}
	sessId, _, err := s.c.Session().Create(se, (&consulapi.WriteOptions{}).WithContext(s.ctx))
	if err != nil {
		// TODO(lukedirtwalker): We need a way to notify
		// the parent if the error happens too many times.
		// Is it possible to recover here (e.g. rpc error making call: Missing node registration)?
		return "", nil, common.NewBasicError("Failed to create session", err)
	}
	s.logger.Info("[LeaderElector] Consul session created", "sessId", sessId)
	sessRefreshClose := make(chan struct{})
	s.startPeriodicSessRefresh(sessId, s.cfg.SessionTTL, sessRefreshClose)
	return sessId, sessRefreshClose, nil
}

func (s *sessMon) startPeriodicSessRefresh(sessId, initialTTL string, closeC chan struct{}) {
	go func() {
		defer log.LogPanicAndExit()
		err := s.c.Session().RenewPeriodic(initialTTL, sessId, nil, closeC)
		if err == consulapi.ErrSessionExpired {
			// if the session expired we are done here.
			s.logger.Debug("[LeaderElector] Session expired")
			return
		}
		if err != nil {
			s.logger.Error("[LeaderElector] Session renewal failed", "err", err)
		}
	}()
}

func (s *sessMon) checkSessionValid(sessId string) error {
	var ok bool
	var modIdx uint64
	var err error
	for {
		ok, modIdx, err = s.checkSessionBlock(sessId, modIdx)
		if err != nil {
			return common.NewBasicError("Failed to check session", err)
		}
		if !ok {
			return nil
		}
	}
}

func (s *sessMon) checkSessionBlock(sessId string, modIdx uint64) (bool, uint64, error) {
	s.logger.Trace("[LeaderElector] Checking session")
	qo := &consulapi.QueryOptions{
		WaitIndex: modIdx,
	}
	qo = qo.WithContext(s.ctx)
	sess, qm, err := s.c.Session().Info(sessId, qo)
	if err != nil {
		return false, 0, common.NewBasicError("Failed to check session state", err)
	}
	if sess == nil {
		return false, 0, nil
	}
	return true, qm.LastIndex, nil
}
