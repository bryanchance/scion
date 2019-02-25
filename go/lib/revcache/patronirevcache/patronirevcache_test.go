// Copyright 2019 Anapaya Systems
// +build patronirunning

package patronirevcache

import (
	"context"
	"database/sql"
	"io/ioutil"
	"os"
	"testing"

	consulapi "github.com/hashicorp/consul/api"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/scionproto/scion/go/lib/ctrl/path_mgmt"
	"github.com/scionproto/scion/go/lib/infra/modules/patroni"
	"github.com/scionproto/scion/go/lib/revcache"
	"github.com/scionproto/scion/go/lib/revcache/revcachetest"
	"github.com/scionproto/scion/go/lib/xtest"
)

var (
	sqlSchema string
)

func loadSchema(t *testing.T) {
	sql, err := ioutil.ReadFile("../../../path_srv/postgres/schema.sql")
	xtest.FailOnErr(t, err)
	sqlSchema = string(sql)
}

func consulAddr() string {
	addr, ok := os.LookupEnv("CONSUL_ADDRESS")
	if ok {
		return addr
	}
	return "127.0.0.1:8500"
}

func setupDB(t *testing.T) *Backend {
	consulCfg := consulapi.DefaultConfig()
	consulCfg.Address = consulAddr()
	c, err := consulapi.NewClient(consulCfg)
	xtest.FailOnErr(t, err)
	cfg := patroni.Conf{
		ClusterKey: "ptest",
		ConnString: "postgresql://psdb:password@host/postgres?sslmode=disable",
	}
	db, err := New(c, cfg)
	xtest.FailOnErr(t, err)
	return db
}

func (b *Backend) dropSchema(ctx context.Context, t *testing.T) {
	conn := b.retry.Pool.WriteConn()
	if conn == nil {
		t.Fatalf("No write connection to drop schema")
	}
	_, err := conn.DB().ExecContext(ctx, "DROP SCHEMA IF EXISTS psdb CASCADE;")
	xtest.FailOnErr(t, err)
}

func (b *Backend) initSchema(ctx context.Context, t *testing.T) {
	conn := b.retry.Pool.WriteConn()
	if conn == nil {
		t.Fatalf("No write connection to drop schema")
	}
	b.dropSchema(ctx, t)
	_, err := conn.DB().ExecContext(ctx, "CREATE SCHEMA psdb;")
	xtest.FailOnErr(t, err)
	_, err = conn.DB().ExecContext(ctx, sqlSchema)
	xtest.FailOnErr(t, err)
}

var _ (revcachetest.TestableRevCache) = (*testRevCache)(nil)

type testRevCache struct {
	*Backend
}

func (c *testRevCache) InsertExpired(t *testing.T, ctx context.Context,
	rev *path_mgmt.SignedRevInfo) {

	rErr := c.Backend.retry.DoWrite(ctx, func(ctx context.Context, db *sql.DB) error {
		newInfo, err := rev.RevInfo()
		xtest.FailOnErr(t, err)
		k := revcache.NewKey(newInfo.IA(), newInfo.IfID)
		packedRev, err := rev.Pack()
		xtest.FailOnErr(t, err)
		tx, err := db.BeginTx(ctx, nil)
		xtest.FailOnErr(t, err)
		query := `
				INSERT INTO Revocations
					(IsdID, AsID, IfID, LinkType, RawTimeStamp, RawTTL, RawSignedRev, Expiration)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
		_, err = tx.ExecContext(ctx, query, k.IA.I, k.IA.A, k.IfId, newInfo.LinkType,
			newInfo.RawTimestamp, newInfo.RawTTL, packedRev, newInfo.Expiration())
		xtest.FailOnErr(t, err)
		err = tx.Commit()
		xtest.FailOnErr(t, err)
		return nil
	})
	xtest.FailOnErr(t, rErr)
}

func TestRevCacheSuite(t *testing.T) {
	loadSchema(t)
	db := setupDB(t)
	rc := &testRevCache{Backend: db}
	Convey("RevcacheSuite", t, func() {
		revcachetest.TestRevCache(t,
			func() revcachetest.TestableRevCache {
				rc.initSchema(context.Background(), t)
				return rc
			},
			func() {
				rc.dropSchema(context.Background(), t)
			})
	})
}
