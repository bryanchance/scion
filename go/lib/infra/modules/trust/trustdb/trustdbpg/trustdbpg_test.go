// Copyright 2018 Anapaya Systems
// +build infrarunning

package trustdbpg

import (
	"context"
	"fmt"
	"io/ioutil"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/scionproto/scion/go/lib/infra/modules/trust/trustdb"
	"github.com/scionproto/scion/go/lib/infra/modules/trust/trustdb/trustdbtest"
	"github.com/scionproto/scion/go/lib/scrypto/cert"
	"github.com/scionproto/scion/go/lib/xtest"
)

var (
	connection string
	sqlSchema  string
)

func init() {
	// sslmode=disable is because dockerized postgres doesn't have SSL enabled.
	connection = fmt.Sprintf("host=%s port=5433 user=csdb password=password sslmode=disable",
		xtest.PostgresHost())
}

func loadSchema(t *testing.T) {
	sql, err := ioutil.ReadFile("../../../../../../cert_srv/postgres/schema.sql")
	xtest.FailOnErr(t, err)
	sqlSchema = string(sql)
}

func (db *TrustDB) dropSchema(ctx context.Context) error {
	_, err := db.db.ExecContext(ctx, "DROP SCHEMA IF EXISTS csdb CASCADE;")
	return err
}

func (db *TrustDB) initSchema(ctx context.Context) error {
	if err := db.dropSchema(ctx); err != nil {
		return err
	}
	if _, err := db.db.ExecContext(ctx, "CREATE SCHEMA csdb;"); err != nil {
		return err
	}
	_, err := db.db.ExecContext(ctx, sqlSchema)
	return err
}

func TestTrustDBSuite(t *testing.T) {
	loadSchema(t)
	db := setupDB(t)
	Convey("TrustDBSuite", t, func() {
		setup := func() trustdb.TrustDB {
			xtest.FailOnErr(t, db.initSchema(context.Background()))
			return db
		}
		cleanup := func(_ trustdb.TrustDB) {
			xtest.FailOnErr(t, db.dropSchema(context.Background()))
		}
		trustdbtest.TestTrustDB(t, setup, cleanup)
	})
}

func TestConcurrentInsertion(t *testing.T) {
	db := setupDB(t)
	Convey("Concurrent Insert", t, func() {
		ctx, cancelF := context.WithTimeout(context.Background(), trustdbtest.Timeout)
		defer cancelF()
		defer db.dropSchema(context.Background())
		chain, err := cert.ChainFromFile("../trustdbtest/testdata/ISD1-ASff00_0_311-V1.crt", false)
		xtest.FailOnErr(t, err)
		var ins1 int64
		var ins2 int64
		chains, err := db.GetAllChains(ctx)
		SoMsg("No chains expected", chains, ShouldBeEmpty)
		xtest.FailOnErr(t, err)
		tx1, err := db.BeginTransaction(ctx, nil)
		xtest.FailOnErr(t, err)
		ins1, err = tx1.InsertIssCert(ctx, chain.Issuer)
		SoMsg("No err expected", err, ShouldBeNil)
		tx2, err := db.BeginTransaction(ctx, nil)
		xtest.FailOnErr(t, err)
		go func() {
			time.Sleep(100 * time.Millisecond)
			err := tx1.Commit()
			xtest.FailOnErr(t, err)
		}()
		ins2, err = tx2.InsertIssCert(ctx, chain.Issuer)
		SoMsg("No err expected", err, ShouldBeNil)
		err = tx2.Commit()
		xtest.FailOnErr(t, err)
		SoMsg("One insertion expected", ins1+ins2, ShouldEqual, int64(1))
	})
}

func setupDB(t *testing.T) *TrustDB {
	db, err := New(connection)
	xtest.FailOnErr(t, err)
	err = db.initSchema(context.Background())
	xtest.FailOnErr(t, err)
	return db
}
