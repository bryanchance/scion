// Copyright 2018 Anapaya Systems.

package pgrevcache

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/scionproto/scion/go/lib/ctrl/path_mgmt"
	"github.com/scionproto/scion/go/lib/revcache"
	"github.com/scionproto/scion/go/lib/revcache/revcachetest"
	"github.com/scionproto/scion/go/lib/xtest"
)

var (
	connection string

	sqlSchema string
)

func checkPostgres(t *testing.T) {
	_, ok := os.LookupEnv("POSTGRESRUNNING")
	if !ok {
		t.Skip("Postgres is not running")
	}
}

func init() {
	// sslmode=disable is because dockerized postgres doesn't have SSL enabled.
	connection = fmt.Sprintf("postgresql://psdb:password@%s/postgres?sslmode=disable",
		xtest.PostgresHost())
}

func loadSchema(t *testing.T) {
	sql, err := ioutil.ReadFile("../../../path_srv/postgres/schema.sql")
	xtest.FailOnErr(t, err)
	sqlSchema = string(sql)
}

var _ (revcachetest.TestableRevCache) = (*testRevCache)(nil)

type testRevCache struct {
	*PgRevCache
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

func (c *testRevCache) Prepare(t *testing.T, ctx context.Context) {
	schemaResetSql := "DROP SCHEMA IF EXISTS psdb CASCADE;\nCREATE SCHEMA psdb;"
	_, err := c.db.ExecContext(ctx, schemaResetSql)
	xtest.FailOnErr(t, err)
	_, err = c.db.ExecContext(ctx, sqlSchema)
	xtest.FailOnErr(t, err)
}

func TestRevcacheSuite(t *testing.T) {
	checkPostgres(t)
	loadSchema(t)
	db, err := New(connection)
	xtest.FailOnErr(t, err)
	c := &testRevCache{
		PgRevCache: db,
	}
	Convey("RevCache Suite", t, func() {
		revcachetest.TestRevCache(t, c)
	})
}
