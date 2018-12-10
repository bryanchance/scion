// Copyright 2018 Anapaya Systems

package leader

import (
	"sync"

	"github.com/scionproto/scion/go/lib/log"
)

// notifier dispatches leader changes.
type notifier struct {
	leader   bool
	mutex    sync.Mutex
	acquired func()
	lost     func()
	logger   log.Logger
}

func (n *notifier) setLeader(leader bool) {
	n.mutex.Lock()
	defer n.mutex.Unlock()
	n.leader = leader
	n.logger.Info("[LeaderElector] Leader change", "leader", leader)
	if n.leader {
		n.acquired()
	} else {
		n.lost()
	}
}
