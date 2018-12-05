// Copyright 2018 Anapaya Systems.
// +build infrarunning

package postgres

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/scionproto/scion/go/lib/pathdb"
	"github.com/scionproto/scion/go/lib/pathdb/pathdbtest"
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
	connection = fmt.Sprintf("host=%s port=5432 user=psdb password=password sslmode=disable",
		dbHost)
}

func (b *Backend) dropSchema(ctx context.Context) error {
	_, err := b.db.ExecContext(ctx, "DROP SCHEMA IF EXISTS psdb CASCADE;")
	return err
}

func (b *Backend) initSchema(ctx context.Context) error {
	if err := b.dropSchema(ctx); err != nil {
		return err
	}
	if _, err := b.db.ExecContext(ctx, "CREATE SCHEMA psdb;"); err != nil {
		return err
	}
	sql, err := ioutil.ReadFile("../../../path_srv/postgres/schema.sql")
	if err != nil {
		return err
	}
	_, err = b.db.ExecContext(ctx, string(sql))
	return err
}

func TestPathDBSuite(t *testing.T) {
	Convey("PathDBSuite", t, func() {
		db, err := New(connection)
		xtest.FailOnErr(t, err)
		err = db.initSchema(context.Background())
		xtest.FailOnErr(t, err)

		pathdbtest.TestPathDB(t,
			func() pathdb.PathDB {
				_, err = db.db.Exec("DELETE FROM Segments")
				xtest.FailOnErr(t, err)
				_, err = db.db.Exec("DELETE FROM NextQuery")
				xtest.FailOnErr(t, err)
				return db
			},
			func() {
				_, err = db.db.Exec("DELETE FROM Segments")
				xtest.FailOnErr(t, err)
			},
		)
		// cleanup
		db.dropSchema(context.Background())
	})
}
