// Copyright 2018 Anapaya Systems

package dynamic

import (
	"context"
	"time"

	"github.com/scionproto/scion/go/lib/addr"
	"github.com/scionproto/scion/go/lib/periodic"
)

var _ periodic.Task = (*ZkTask)(nil)

type ZkTask struct {
	IA        addr.IA
	Instances []string
	Timeout   time.Duration
}

func (t *ZkTask) Run(_ context.Context) {
	UpdateFromZK(t.Instances, t.IA, t.Timeout)
}
