// Copyright 2018 Anapaya Systems

package leader

import (
	"context"
	"strings"
	"time"

	"github.com/scionproto/scion/go/lib/common"
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

func isNoClusterLeader(err error, logger log.Logger) bool {
	errString := common.FmtError(err)
	if strings.Contains(errString, "500 (No cluster leader)") {
		logger.Trace("[LeaderElector] No cluster leader", "err", err)
		return true
	}
	return false
}
