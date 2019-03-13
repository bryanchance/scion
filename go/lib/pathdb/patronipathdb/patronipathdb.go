// Copyright 2019 Anapaya Systems.

// Package patronipathdb implements the path DB interface for a patroni cluster.
package patronipathdb

import (
	"context"
	"database/sql"
	"time"

	consulapi "github.com/hashicorp/consul/api"

	"github.com/scionproto/scion/go/lib/addr"
	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/lib/ctrl/seg"
	"github.com/scionproto/scion/go/lib/infra/modules/patroni"
	"github.com/scionproto/scion/go/lib/pathdb"
	"github.com/scionproto/scion/go/lib/pathdb/postgres"
	"github.com/scionproto/scion/go/lib/pathdb/query"
)

var _ pathdb.PathDB = (*Backend)(nil)

// Backend implements the path DB interface for a patroni cluster.
type Backend struct {
	retry *patroni.RetryHelper
}

// New creates a new path DB backend that connects to a patroni cluster.
func New(c *consulapi.Client, cfg patroni.Conf) (*Backend, error) {
	pool, err := patroni.NewConnPool(c, cfg)
	if err != nil {
		return nil, err
	}
	return &Backend{
		retry: &patroni.RetryHelper{Pool: pool},
	}, nil
}

func (b *Backend) Insert(ctx context.Context, sm *seg.Meta) (int, error) {
	var ret int
	rErr := b.retry.DoWrite(ctx, func(ctx context.Context, db *sql.DB) error {
		var err error
		ret, err = postgres.NewFromDB(db).Insert(ctx, sm)
		return err
	})
	return ret, rErr
}

func (b *Backend) InsertWithHPCfgIDs(ctx context.Context,
	sm *seg.Meta, hpCfgIDs []*query.HPCfgID) (int, error) {

	var ret int
	rErr := b.retry.DoWrite(ctx, func(ctx context.Context, db *sql.DB) error {
		var err error
		ret, err = postgres.NewFromDB(db).InsertWithHPCfgIDs(ctx, sm, hpCfgIDs)
		return err
	})
	return ret, rErr
}

func (b *Backend) Delete(ctx context.Context, params *query.Params) (int, error) {
	var ret int
	rErr := b.retry.DoWrite(ctx, func(ctx context.Context, db *sql.DB) error {
		var err error
		ret, err = postgres.NewFromDB(db).Delete(ctx, params)
		return err
	})
	return ret, rErr
}

func (b *Backend) DeleteExpired(ctx context.Context, now time.Time) (int, error) {
	var ret int
	rErr := b.retry.DoWrite(ctx, func(ctx context.Context, db *sql.DB) error {
		var err error
		ret, err = postgres.NewFromDB(db).DeleteExpired(ctx, now)
		return err
	})
	return ret, rErr
}

func (b *Backend) Get(ctx context.Context, params *query.Params) (query.Results, error) {
	var ret query.Results
	rErr := b.retry.DoRead(ctx, func(ctx context.Context, db *sql.DB) error {
		var err error
		ret, err = postgres.NewFromDB(db).Get(ctx, params)
		return err
	})
	return ret, rErr
}

func (b *Backend) GetAll(ctx context.Context) (<-chan query.ResultOrErr, error) {
	var ret <-chan query.ResultOrErr
	rErr := b.retry.DoRead(ctx, func(ctx context.Context, db *sql.DB) error {
		var err error
		ret, err = postgres.NewFromDB(db).GetAll(ctx)
		return err
	})
	return ret, rErr
}

func (b *Backend) InsertNextQuery(ctx context.Context,
	dst addr.IA, nextQuery time.Time) (bool, error) {

	var ret bool
	rErr := b.retry.DoWrite(ctx, func(ctx context.Context, db *sql.DB) error {
		var err error
		ret, err = postgres.NewFromDB(db).InsertNextQuery(ctx, dst, nextQuery)
		return err
	})
	return ret, rErr
}

func (b *Backend) GetNextQuery(ctx context.Context, dst addr.IA) (*time.Time, error) {
	var ret *time.Time
	rErr := b.retry.DoRead(ctx, func(ctx context.Context, db *sql.DB) error {
		var err error
		ret, err = postgres.NewFromDB(db).GetNextQuery(ctx, dst)
		return err
	})
	return ret, rErr
}

func (b *Backend) BeginTransaction(ctx context.Context,
	opts *sql.TxOptions) (pathdb.Transaction, error) {

	var errHist []error
	for i := 0; i < patroni.MaxWriteRetries; i++ {
		wconn := b.retry.Pool.WriteConn()
		if wconn != nil {
			tx, err := postgres.NewFromDB(wconn.DB()).BeginTransaction(ctx, opts)
			if err == nil {
				return &transaction{
					conn: wconn,
					tx:   tx,
				}, nil
			}
			if !wconn.ReportErr(err) {
				// No need to retry if it is an error that doesn't impact the pool.
				return nil, err
			}
			errHist = append(errHist, err)
		} else {
			errHist = append(errHist, common.NewBasicError(patroni.ErrNoConnection, nil))
		}
		if err := patroni.SleepOrTimeOut(ctx, patroni.WaitBetweenRetry); err != nil {
			return nil, common.NewBasicError("Err during sleep", err, "errHist", errHist)
		}
	}
	return nil, common.NewBasicError(patroni.ErrBadConnection, nil, "errHist", errHist)
}

var _ (pathdb.Transaction) = (*transaction)(nil)

type transaction struct {
	conn *patroni.ConnRef
	tx   pathdb.Transaction
}

func (tx *transaction) Get(ctx context.Context, params *query.Params) (query.Results, error) {
	ret, err := tx.tx.Get(ctx, params)
	tx.conn.ReportErr(err)
	return ret, err
}

func (tx *transaction) GetAll(ctx context.Context) (<-chan query.ResultOrErr, error) {
	ret, err := tx.tx.GetAll(ctx)
	tx.conn.ReportErr(err)
	return ret, err
}

func (tx *transaction) GetNextQuery(ctx context.Context, dst addr.IA) (*time.Time, error) {
	ret, err := tx.tx.GetNextQuery(ctx, dst)
	tx.conn.ReportErr(err)
	return ret, err
}

func (tx *transaction) Insert(ctx context.Context, segMeta *seg.Meta) (int, error) {
	ret, err := tx.tx.Insert(ctx, segMeta)
	tx.conn.ReportErr(err)
	return ret, err
}

func (tx *transaction) InsertWithHPCfgIDs(ctx context.Context, segMeta *seg.Meta,
	hpCfgIDs []*query.HPCfgID) (int, error) {

	ret, err := tx.tx.InsertWithHPCfgIDs(ctx, segMeta, hpCfgIDs)
	tx.conn.ReportErr(err)
	return ret, err
}

func (tx *transaction) Delete(ctx context.Context, params *query.Params) (int, error) {
	ret, err := tx.tx.Delete(ctx, params)
	tx.conn.ReportErr(err)
	return ret, err
}

func (tx *transaction) DeleteExpired(ctx context.Context, now time.Time) (int, error) {
	ret, err := tx.tx.DeleteExpired(ctx, now)
	tx.conn.ReportErr(err)
	return ret, err
}

func (tx *transaction) InsertNextQuery(ctx context.Context,
	dst addr.IA, nextQuery time.Time) (bool, error) {

	ret, err := tx.tx.InsertNextQuery(ctx, dst, nextQuery)
	tx.conn.ReportErr(err)
	return ret, err
}

func (tx *transaction) Commit() error {
	err := tx.tx.Commit()
	tx.conn.ReportErr(err)
	return err
}

func (tx *transaction) Rollback() error {
	err := tx.tx.Rollback()
	tx.conn.ReportErr(err)
	return err
}
