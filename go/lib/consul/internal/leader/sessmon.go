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
	c       *consulapi.Client
	ctx     context.Context
	cancelF context.CancelFunc
	logger  log.Logger
	cfg     consulconfig.LeaderElectorConf
	ln      *notifier
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
		// First create the session and in case of an error retry after a sleep,
		// a context error means that we have been cancelled and should quit.
		sessId, rClose, refreshErrC, err := s.createSession()
		if err != nil {
			if isContextErr(s.ctx, s.logger) {
				return
			}
			if s.cfg.LeaderIfNoClusterLeader && isNoClusterLeader(err, s.logger) {
				s.runWithoutClusterLeader()
				continue
			}
			s.logger.Error("[LeaderElector] Error during create session", "err", err)
			sleep(s.ctx, time.Second)
			continue
		}
		// Now that we have a session start the leader monitor for this session.
		lm := startLeaderMon(s.key, s.c, sessId, s.ln)
		// After starting the leader monitor we have to make sure our session is still valid.
		// Once this function returns we should stop being leader
		// since the session is no longer valid.
		if err := s.checkSessionValid(sessId, refreshErrC); err != nil {
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

func (s *sessMon) createSession() (string, chan struct{}, chan error, error) {
	se := &consulapi.SessionEntry{
		Name:      s.cfg.Name,
		TTL:       s.cfg.SessionTTL,
		LockDelay: s.cfg.LockDelay,
	}
	sessId, _, err := s.c.Session().Create(se, (&consulapi.WriteOptions{}).WithContext(s.ctx))
	if err != nil {
		// TODO(lukedirtwalker): We need a way to notify
		// the parent if the error happens too many times.
		// Is it possible to recover here (e.g. rpc error making call: Missing node registration)?
		return "", nil, nil, common.NewBasicError("Failed to create session", err)
	}
	s.logger.Info("[LeaderElector] Consul session created", "sessId", sessId)
	sessRefreshClose := make(chan struct{})
	errChan := make(chan error)
	s.startPeriodicSessRefresh(sessId, s.cfg.SessionTTL, sessRefreshClose, errChan)
	return sessId, sessRefreshClose, errChan, nil
}

func (s *sessMon) startPeriodicSessRefresh(sessId, initialTTL string, closeC chan struct{},
	errorC chan error) {

	go func() {
		defer log.LogPanicAndExit()
		err := s.c.Session().RenewPeriodic(initialTTL, sessId, nil, closeC)
		if err == consulapi.ErrSessionExpired {
			// if the session expired we are done here.
			s.logger.Debug("[LeaderElector] Session expired")
			return
		}
		if err != nil {
			errorC <- err
		}
	}()
}

// runWithoutClusterLeader runs the leadership acquired callback and waits until consul has
// a cluster leader again. Once consul has a cluster leader again the local leadership is droppped.
func (s *sessMon) runWithoutClusterLeader() {
	s.ln.acquired()
	defer s.ln.lost()
	for {
		// This API doesn't return an error even if there is no cluster leader,
		// in this case it will just return an empty string.
		consulLeader, err := s.c.Status().Leader()
		if err != nil {
			log.Error("[LeaderElector] Error during status/leader lookup", "err", err)
			return
		}
		if consulLeader != "" {
			return
		}
		time.Sleep(time.Second)
	}
}

// checkSessionValid blockingly checks that the session with the given sessId
// is valid and still exists in consul.
// In case it no longer exists or in case of an error (either refresh, or from fetching info)
// this method returns.
func (s *sessMon) checkSessionValid(sessId string, refreshErrC chan error) error {
	var modIdx uint64
	for {
		select {
		case err := <-refreshErrC:
			return common.NewBasicError("Failed to refresh session", err)
		case si := <-s.querySessionInfo(sessId, modIdx):
			if si.err != nil {
				return common.NewBasicError("Failed to check session", si.err)
			}
			if !si.sessionOk {
				return nil
			}
			modIdx = si.modIdx
		}
	}
}

type sessionInfo struct {
	sessionOk bool
	modIdx    uint64
	err       error
}

// querySessionInfo starts a go routine that blockingly waits for session info changes
// using the given sessId and modify index.
// Once the blocking call returns the result is put in the channel.
func (s *sessMon) querySessionInfo(sessId string, modIdx uint64) chan sessionInfo {
	s.logger.Trace("[LeaderElector] Checking session")
	qo := &consulapi.QueryOptions{
		WaitIndex: modIdx,
	}
	qo = qo.WithContext(s.ctx)
	retChan := make(chan sessionInfo)
	go func() {
		sess, qm, err := s.c.Session().Info(sessId, qo)
		if err != nil {
			retChan <- sessionInfo{err: common.NewBasicError("Failed to check session state", err)}
		} else if sess == nil {
			retChan <- sessionInfo{}
		} else {
			retChan <- sessionInfo{sessionOk: true, modIdx: qm.LastIndex}
		}
	}()
	return retChan
}
