// Copyright 2018 Anapaya Systems.
// +build infrarunning

package pgrevcache

import (
	"context"
	"fmt"
	"io/ioutil"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/scionproto/scion/go/lib/ctrl/path_mgmt"
	"github.com/scionproto/scion/go/lib/revcache"
	"github.com/scionproto/scion/go/lib/revcache/revcachetest"
	"github.com/scionproto/scion/go/lib/xtest"
)

var (
	connection string
)

func init() {
	// sslmode=disable is because dockerized postgres doesn't have SSL enabled.
	connection = fmt.Sprintf("host=%s port=5432 user=psdb password=password sslmode=disable",
		xtest.PostgresHost())
}

var _ (revcachetest.TestableRevCache) = (*testRevCache)(nil)

type testRevCache struct {
	*pgRevCache
}

func (c *testRevCache) dropSchema(ctx context.Context) error {
	_, err := c.db.ExecContext(ctx, "DROP SCHEMA IF EXISTS psdb CASCADE;")
	return err
}

func (c *testRevCache) initSchema(ctx context.Context) error {
	if err := c.dropSchema(ctx); err != nil {
		return err
	}
	if _, err := c.db.ExecContext(ctx, "CREATE SCHEMA psdb;"); err != nil {
		return err
	}
	sql, err := ioutil.ReadFile("../../../path_srv/postgres/schema.sql")
	if err != nil {
		return err
	}
	_, err = c.db.ExecContext(ctx, string(sql))
	return err
}

func (c *testRevCache) InsertExpired(t *testing.T, ctx context.Context,
	rev *path_mgmt.SignedRevInfo) {

	newInfo, err := rev.RevInfo()
	xtest.FailOnErr(t, err)
	k := revcache.NewKey(newInfo.IA(), newInfo.IfID)
	packedRev, err := rev.Pack()
	xtest.FailOnErr(t, err)
	tx, err := c.db.BeginTx(ctx, nil)
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
}

func TestRevcacheSuite(t *testing.T) {
	Convey("RevCache Suite", t, func() {
		db, err := New(connection)
		xtest.FailOnErr(t, err)
		c := &testRevCache{
			pgRevCache: db,
		}
		err = c.initSchema(context.Background())
		xtest.FailOnErr(t, err)
		revcachetest.TestRevCache(t,
			func() revcachetest.TestableRevCache {
				_, err = c.db.Exec("DELETE FROM Revocations")
				xtest.FailOnErr(t, err)
				return c
			},
			func() {
				_, err = c.db.Exec("DELETE FROM Revocations")
				xtest.FailOnErr(t, err)
			},
		)
	})
}
