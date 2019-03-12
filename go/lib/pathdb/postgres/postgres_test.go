// Copyright 2018 Anapaya Systems.

package postgres

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/scionproto/scion/go/lib/pathdb/pathdbtest"
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

var _ pathdbtest.TestablePathDB = (*TestPathDB)(nil)

type TestPathDB struct {
	*Backend
}

func (b *TestPathDB) Prepare(t *testing.T, ctx context.Context) {
	_, err := b.db.ExecContext(ctx, "DROP SCHEMA IF EXISTS psdb CASCADE;")
	xtest.FailOnErr(t, err)
	_, err = b.db.ExecContext(ctx, "CREATE SCHEMA psdb;")
	xtest.FailOnErr(t, err)
	_, err = b.db.ExecContext(ctx, sqlSchema)
	xtest.FailOnErr(t, err)
}

func TestPathDBSuite(t *testing.T) {
	checkPostgres(t)
	loadSchema(t)
	db, err := New(connection)
	xtest.FailOnErr(t, err)
	tdb := &TestPathDB{Backend: db}
	Convey("PathDBSuite", t, func() {
		pathdbtest.TestPathDB(t, tdb)
	})
}
