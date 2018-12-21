// Copyright 2018 Anapaya Systems
// +build infrarunning

package trustdbpg

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/scionproto/scion/go/lib/infra/modules/trust/trustdb"
	"github.com/scionproto/scion/go/lib/infra/modules/trust/trustdb/trustdbtest"
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
	connection = fmt.Sprintf("host=%s port=5433 user=csdb password=password sslmode=disable",
		dbHost)
}

func (db *trustDB) dropSchema(ctx context.Context) error {
	_, err := db.db.ExecContext(ctx, "DROP SCHEMA IF EXISTS csdb CASCADE;")
	return err
}

func (db *trustDB) initSchema(ctx context.Context) error {
	if err := db.dropSchema(ctx); err != nil {
		return err
	}
	if _, err := db.db.ExecContext(ctx, "CREATE SCHEMA csdb;"); err != nil {
		return err
	}
	sql, err := ioutil.ReadFile("../../../../../../cert_srv/postgres/schema.sql")
	if err != nil {
		return err
	}
	_, err = db.db.ExecContext(ctx, string(sql))
	return err
}

func TestTrustDBSuite(t *testing.T) {
	Convey("TrustDBSuite", t, func() {
		db, err := New(connection)
		xtest.FailOnErr(t, err)
		err = db.initSchema(context.Background())
		xtest.FailOnErr(t, err)

		setup := func() trustdb.TrustDB {
			db.initSchema(context.Background())
			return db
		}
		cleanup := func(_ trustdb.TrustDB) {
			db.dropSchema(context.Background())
		}
		trustdbtest.TestTrustDB(t, setup, cleanup)

		db.dropSchema(context.Background())
	})
}
