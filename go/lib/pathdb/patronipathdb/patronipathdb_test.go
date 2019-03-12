// Copyright 2019 Anapaya Systems.

package patronipathdb

import (
	"context"
	"io/ioutil"
	"os"
	"testing"

	consulapi "github.com/hashicorp/consul/api"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/scionproto/scion/go/lib/infra/modules/patroni"
	"github.com/scionproto/scion/go/lib/pathdb/pathdbtest"
	"github.com/scionproto/scion/go/lib/xtest"
)

var (
	sqlSchema string
)

func checkPatroni(t *testing.T) {
	_, ok := os.LookupEnv("PATRONIRUNNING")
	if !ok {
		t.Skip("Patroni is not running")
	}
}

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
		DBUser:     "psdb",
		DBPass:     "password",
	}
	db, err := New(c, cfg)
	xtest.FailOnErr(t, err)
	return db
}

var _ pathdbtest.TestablePathDB = (*TestPathDB)(nil)

type TestPathDB struct {
	*Backend
}

func (b *TestPathDB) Prepare(t *testing.T, ctx context.Context) {
	conn := b.retry.Pool.WriteConn()
	if conn == nil {
		t.Fatalf("No write connection to drop schema")
	}
	_, err := conn.DB().ExecContext(ctx, "DROP SCHEMA IF EXISTS psdb CASCADE;")
	xtest.FailOnErr(t, err)
	_, err = conn.DB().ExecContext(ctx, "CREATE SCHEMA psdb;")
	xtest.FailOnErr(t, err)
	_, err = conn.DB().ExecContext(ctx, sqlSchema)
	xtest.FailOnErr(t, err)
}

func TestPathDBSuite(t *testing.T) {
	checkPatroni(t)
	loadSchema(t)
	db := setupDB(t)
	tdb := &TestPathDB{Backend: db}
	Convey("PathDBSuite", t, func() {
		pathdbtest.TestPathDB(t, tdb)
	})
}
