// Copyright 2018 Anapaya Systems

package leader

import (
	"context"
	"time"

	"github.com/scionproto/scion/go/lib/log"
)

func sleep(ctx context.Context, d time.Duration) bool {
	select {
	case <-ctx.Done():
		return false
	case <-time.After(d):
		return true
	}
}

func isContextErr(ctx context.Context, logger log.Logger) bool {
	if ctx.Err() != nil {
		logger.Trace("[LeaderElector] Context canceled", "err", ctx.Err())
		return true
	}
	return false
}
