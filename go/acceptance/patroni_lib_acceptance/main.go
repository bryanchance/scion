// Copyright 2019 Anapaya Systems

package main

import (
	"context"
	"database/sql"
	"flag"
	"os"
	"time"

	consulapi "github.com/hashicorp/consul/api"

	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/lib/env"
	"github.com/scionproto/scion/go/lib/infra/modules/patroni"
	"github.com/scionproto/scion/go/lib/log"
)

var (
	key = flag.String("key", "", "Key where the patroni cluster is stored")
	// ctrlFile is used to indicate that the user expects the cluster to be in a failing state, i.e.
	// If the ctrlFile doesn't exist and no write connection is returned, the test fails immediately
	// if the ctrlFile exists and no write connection is returned,
	// it is expected that a read connection is returned.
	ctrlFile = flag.String("ctrlFile", "", "Control file path")
	agent    = flag.String("agent", "", "Consul agent")
)

func main() {
	os.Exit(realMain())
}

type testConn struct {
	db *sql.DB
}

func realMain() int {
	log.AddLogConsFlags()
	flag.Parse()
	log.SetupFromFlags("patroni_lib_acceptance")
	environ := env.SetupEnv(func() {})
	cfg := consulapi.DefaultConfig()
	cfg.Address = *agent
	c, err := consulapi.NewClient(cfg)
	if err != nil {
		log.Crit("Error during setup", "err", err)
		return 1
	}

	pool, err := patroni.NewConnPool(c, patroni.Conf{
		ClusterKey: "ptest",
		DBUser:     "postgres",
		DBPass:     "password",
	})
	if err != nil {
		log.Crit("Failed to setup conn pool", "err", err)
		return 1
	}
	stopChan := make(chan struct{})
	errChan := make(chan error)
	go testPool(pool, stopChan, errChan)
	select {
	case <-environ.AppShutdownSignal:
		close(stopChan)
		<-errChan
		return 0
	case err := <-errChan:
		log.Crit("Failed", "err", err)
		return 1
	}
}

func testPool(pool *patroni.ConnPool, stopChan chan struct{}, errChan chan error) {
	t := test{pool: pool}
	for {
		select {
		case <-stopChan:
			errChan <- nil
			return
		case <-time.After(500 * time.Millisecond):
			if err := t.checkPool(); err != nil {
				errChan <- err
				return
			}
		}
	}
}

type test struct {
	pool    *patroni.ConnPool
	readErr int
}

func (t *test) checkPool() error {
	if err := t.checkWriteConn(); err != nil {
		return err
	}

	return t.checkReadConn()
}

func (t *test) checkWriteConn() error {
	wConn := t.pool.WriteConn()
	if wConn == nil {
		if crtlFileMissing() {
			return common.NewBasicError("No write connection found", nil)
		}
		log.Trace("No write conn returned.")
		// ok the tester expected to be no write connection
		return nil
	}
	log.Trace("Test write", "conn", wConn)
	ctx, cancelF := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancelF()
	if _, err := wConn.DB().ExecContext(ctx, "DROP SCHEMA IF EXISTS test CASCADE;"); err != nil {
		if crtlFileMissing() {
			return common.NewBasicError("Write Failed to drop schema", err)
		}
		wConn.ReportErr(err)
		log.Trace("Reported failed write")
		return nil
	}
	if _, err := wConn.DB().ExecContext(ctx, "CREATE SCHEMA test;"); err != nil {
		if crtlFileMissing() {
			return common.NewBasicError("Write Failed to create a schema", err)
		}
		wConn.ReportErr(err)
		log.Trace("Reported failed write")
		return nil
	}
	log.Trace("OK: Write conn is ok.")
	return nil
}

func (t *test) checkReadConn() error {
	rConn := t.pool.ReadConn()
	if rConn == nil {
		// We might not get read connections for a short time
		// if we have only one patroni node and that switches from follower to leader.
		if t.readErr > 2 {
			return common.NewBasicError("No read connection found", nil)
		}
		t.readErr++
		log.Trace("Temp no read connection", "readErrors", t.readErr)
		return nil
	}
	log.Trace("Test read", "conn", rConn)
	ctx, cancelF := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancelF()
	if err := rConn.DB().PingContext(ctx); err != nil {
		// We might not get read failures for a short time
		// if we have only one patroni node and that switches from follower to leader.
		if t.readErr > 2 {
			return common.NewBasicError("Read Failed", err)
		}
		log.Trace("Read failure", "prevReadErrors", t.readErr)
		t.readErr++
	}
	t.readErr = 0
	log.Trace("OK: Read conn is ok.")
	return nil
}

func crtlFileMissing() bool {
	_, err := os.Stat(*ctrlFile)
	return os.IsNotExist(err)
}
