// Copyright 2018 Anapaya Systems

package leader

import (
	"context"
	"time"

	consulapi "github.com/hashicorp/consul/api"

	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/lib/log"
)

// leaderMon tries to acquire a lock on the key at consul.
// leaderMon informs leaderNotifier about leader changes.
type leaderMon struct {
	key     string
	c       *consulapi.Client
	sessId  string
	ctx     context.Context
	cancelF context.CancelFunc
	logger  log.Logger
	ln      *notifier
}

func startLeaderMon(key string, c *consulapi.Client, sessId string, ln *notifier) *leaderMon {
	ln.logger = log.New("key", key, "session", sessId)
	lm := &leaderMon{
		key:    key,
		c:      c,
		sessId: sessId,
		logger: log.New("key", key, "session", sessId),
		ln:     ln,
	}
	lm.ctx, lm.cancelF = context.WithCancel(context.Background())
	go func() {
		defer log.LogPanicAndExit()
		lm.run()
	}()
	return lm
}

func (l *leaderMon) Cancel() {
	l.cancelF()
}

func (l *leaderMon) run() {
	for l.ctx.Err() == nil {
		leader, err := l.acquire()
		if err != nil {
			if isContextErr(l.ctx, l.logger) {
				return
			}
			l.logger.Error("[LeaderElector] Failed to acquire leader", "err", err)
			continue
		}
		if leader {
			l.ln.setLeader(true)
		}
		sessId, err := l.waitForChanges()
		if err != nil {
			if leader {
				l.ln.setLeader(false)
			}
			if isContextErr(l.ctx, l.logger) {
				return
			}
			l.logger.Error("[LeaderElector] Failed to wait for changes", "err", err)
			continue
		}
		l.logger.Debug("[LeaderElector] Session changed", "newSessId", sessId)
		if leader {
			l.ln.setLeader(false)
		}
	}
}

func (l *leaderMon) sessionName(sessId string) string {
	qo := &consulapi.QueryOptions{}
	qo = qo.WithContext(l.ctx)
	session, _, err := l.c.Session().Info(sessId, qo)
	if err != nil {
		l.logger.Warn("Failed to list session", "err", err)
		return "Could not determine"
	}
	return session.Name
}

func (l *leaderMon) acquire() (bool, error) {
	var sessId string
	var err error
	for ; sessId == ""; sleep(l.ctx, time.Second) {
		if acquired, err := l.tryAcquire(); err != nil || acquired {
			return acquired, err
		}
		sessId, _, err = l.lockSessIdBlock(0)
		if err != nil {
			return false, err
		}
	}
	return l.sessId == sessId, nil
}

func (l *leaderMon) tryAcquire() (bool, error) {
	l.logger.Trace("[LeaderElector] Try to be leader")
	kvPair := &consulapi.KVPair{
		Key:     l.key,
		Session: l.sessId,
	}
	qo := &consulapi.WriteOptions{}
	qo = qo.WithContext(l.ctx)
	acquired, _, err := l.c.KV().Acquire(kvPair, qo)
	if err != nil {
		return false, common.NewBasicError("Error while acquiring leader lock", err)
	}
	return acquired, nil
}

func (l *leaderMon) waitForChanges() (string, error) {
	sessId, modIdx, err := l.lockSessIdBlock(0)
	if err != nil {
		return "", common.NewBasicError("Error while waiting on changes", err)
	}
	if sessId != l.sessId {
		l.logger.Info("[LeaderElector] Current leader",
			"leaderSession", sessId, "leaderName", l.sessionName(sessId))
	}
	for {
		var newSessId string
		newSessId, modIdx, err = l.lockSessIdBlock(modIdx)
		if err != nil {
			return "", common.NewBasicError("Error while waiting on changes", err)
		}
		if newSessId != sessId {
			return sessId, nil
		}
		sessId = newSessId
	}
}

func (l *leaderMon) lockSessIdBlock(modIdx uint64) (string, uint64, error) {
	qo := &consulapi.QueryOptions{
		WaitIndex: modIdx,
	}
	// make it cancellable
	qo = qo.WithContext(l.ctx)
	kvp, _, err := l.c.KV().Get(l.key, qo)
	if err != nil {
		return "", 0, common.NewBasicError("Failed to get KV", err)
	}
	if kvp == nil {
		return "", 0, nil
	}
	return kvp.Session, kvp.ModifyIndex, nil
}
