// Copyright 2019 Anapaya Systems

// Package patronirevcache implements the revocation cache interface for a patroni cluster.
package patronirevcache

import (
	"context"
	"database/sql"

	consulapi "github.com/hashicorp/consul/api"

	"github.com/scionproto/scion/go/lib/ctrl/path_mgmt"
	"github.com/scionproto/scion/go/lib/infra/modules/patroni"
	"github.com/scionproto/scion/go/lib/revcache"
	"github.com/scionproto/scion/go/lib/revcache/pgrevcache"
)

var _ revcache.RevCache = (*Backend)(nil)

// Backend implements the revocation cache interface for a patroni cluster.
type Backend struct {
	retry *patroni.RetryHelper
}

// New creates a new revocation cache backend that connects to a patroni cluster.
func New(c *consulapi.Client, cfg patroni.Conf) (*Backend, error) {
	pool, err := patroni.NewConnPool(c, cfg)
	if err != nil {
		return nil, err
	}
	return &Backend{
		retry: &patroni.RetryHelper{Pool: pool},
	}, nil
}

func (b *Backend) Get(ctx context.Context,
	k *revcache.Key) (*path_mgmt.SignedRevInfo, bool, error) {

	var rev *path_mgmt.SignedRevInfo
	var ok bool
	rErr := b.retry.DoRead(ctx, func(ctx context.Context, db *sql.DB) error {
		var err error
		rev, ok, err = pgrevcache.NewFromDB(db).Get(ctx, k)
		return err
	})
	return rev, ok, rErr
}

func (b *Backend) GetAll(ctx context.Context,
	keys map[revcache.Key]struct{}) ([]*path_mgmt.SignedRevInfo, error) {

	var revs []*path_mgmt.SignedRevInfo
	rErr := b.retry.DoRead(ctx, func(ctx context.Context, db *sql.DB) error {
		var err error
		revs, err = pgrevcache.NewFromDB(db).GetAll(ctx, keys)
		return err
	})
	return revs, rErr
}

func (b *Backend) Insert(ctx context.Context,
	rev *path_mgmt.SignedRevInfo) (bool, error) {

	var ok bool
	rErr := b.retry.DoWrite(ctx, func(ctx context.Context, db *sql.DB) error {
		var err error
		ok, err = pgrevcache.NewFromDB(db).Insert(ctx, rev)
		return err
	})
	return ok, rErr
}

func (b *Backend) DeleteExpired(ctx context.Context) (int64, error) {
	var ret int64
	rErr := b.retry.DoWrite(ctx, func(ctx context.Context, db *sql.DB) error {
		var err error
		ret, err = pgrevcache.NewFromDB(db).DeleteExpired(ctx)
		return err
	})
	return ret, rErr
}
