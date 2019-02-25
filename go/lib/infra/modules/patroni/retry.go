// Copyright 2019 Anapaya Systems.

package patroni

import (
	"context"
	"database/sql"
	"time"

	"github.com/scionproto/scion/go/lib/common"
)

const (
	// ErrBadConnection is the error string in case a repeated read/write action fails because
	// either there is no connection or the connection failed. The error contains an errHist context
	// that contains errors from the tries, if there were any.
	ErrBadConnection = "Bad connection"
	// ErrNoConnection is the error string in case no connection is returned.
	ErrNoConnection = "No connection"
)

var (
	// WaitBetweenRetry is the duration reads/writes wait for before retrying, if there is no
	// connection.
	WaitBetweenRetry = 500 * time.Millisecond
	// MaxWriteRetries is the maximum amount of retries when doing a write.
	MaxWriteRetries = 3
	// MaxReadRetries is the maximum amount of retries when doing a read.
	MaxReadRetries = 3
)

// DBAction describes an action on the database. It receives the context and a db connection.
type DBAction func(context.Context, *sql.DB) error

// RetryHelper is a helper to retry read/write actions on the DB, if the connection pool doesn't
// immediately return a connection, or if a connection is suddenly broken.
type RetryHelper struct {
	Pool *ConnPool
}

// DoWrite tries to get a write connection and execute the write action. In case of an error or if
// there is no write connection the function sleeps and retries. The sleep time and retries can be
// configured using the package variables.
func (r *RetryHelper) DoWrite(ctx context.Context, write DBAction) error {
	return r.doInternally(ctx, write, (*ConnPool).WriteConn)
}

// DoRead tries to get a read connection and execute the read action. In case of an error or if
// there is no read connection the function sleeps and retries. The sleep time and retries can be
// configured using the package variables.
func (r *RetryHelper) DoRead(ctx context.Context, read DBAction) error {
	return r.doInternally(ctx, read, (*ConnPool).ReadConn)
}

func (r *RetryHelper) doInternally(ctx context.Context, action DBAction,
	connProvider func(*ConnPool) *ConnRef) error {

	var errHist []error
	for i := 0; i < MaxReadRetries; i++ {
		conn := connProvider(r.Pool)
		if conn != nil {
			err := action(ctx, conn.DB())
			if !conn.ReportErr(err) {
				// No need to retry if it is an error that doesn't impact the pool.
				return err
			}
			errHist = append(errHist, err)
		} else {
			errHist = append(errHist, common.NewBasicError(ErrNoConnection, nil))
		}
		if i >= MaxReadRetries {
			break // break early to prevent a useless sleep.
		}
		if err := SleepOrTimeOut(ctx, WaitBetweenRetry); err != nil {
			// context timeout so return
			return common.NewBasicError("Err during sleep", err, "errHist", errHist, "retries", i)
		}
	}
	return common.NewBasicError(ErrBadConnection, nil, "errHist", errHist)
}

// SleepOrTimeOut sleeps for the given duration respecting the context timeout.
func SleepOrTimeOut(ctx context.Context, dur time.Duration) error {
	select {
	case <-time.After(dur):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
