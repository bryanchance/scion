// Copyright 2018 Anapaya Systems.
// +build infrarunning

package pgrevcache

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/scionproto/scion/go/lib/revcache"
	"github.com/scionproto/scion/go/lib/revcache/revcachetest"
	"github.com/scionproto/scion/go/lib/xtest"
)

var (
	connection string
)

func init() {
	var dbHost string
	if xtest.RunsInDocker() {
		var ok bool
		dbHost, ok = os.LookupEnv("DOCKER0")
		if !ok {
			panic("Expected DOCKER0 env variable")
		}
	} else {
		dbHost = "localhost"
	}
	// sslmode=disable is because dockerized postgres doesn't have SSL enabled.
	connection = fmt.Sprintf("host=%s port=5432 user=pathdb password=password sslmode=disable",
		dbHost)
}

func (c *pgRevCache) dropSchema(ctx context.Context) error {
	_, err := c.db.ExecContext(ctx, "DROP SCHEMA IF EXISTS pathdb CASCADE;")
	return err
}

func (c *pgRevCache) initSchema(ctx context.Context) error {
	if err := c.dropSchema(ctx); err != nil {
		return err
	}
	if _, err := c.db.ExecContext(ctx, "CREATE SCHEMA pathdb;"); err != nil {
		return err
	}
	sql, err := ioutil.ReadFile("../../../path_srv/postgres/schema.sql")
	if err != nil {
		return err
	}
	_, err = c.db.ExecContext(ctx, string(sql))
	return err
}

func TestRevcacheSuite(t *testing.T) {
	Convey("RevCache Suite", t, func() {
		c, err := New(connection)
		xtest.FailOnErr(t, err)
		err = c.initSchema(context.Background())
		xtest.FailOnErr(t, err)
		revcachetest.TestRevCache(t,
			func() revcache.RevCache {
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
