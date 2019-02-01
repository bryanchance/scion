// Copyright 2019 Anapaya Systems

// This program creates an empty sigmgmt database with all the tables
// present but with no data inside them.
//
// usage: sigmgmt_createdb <database-file>
package main

import (
	"fmt"
	"os"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"

	"github.com/scionproto/scion/go/lib/integration"
	"github.com/scionproto/scion/go/lib/log"
	"github.com/scionproto/scion/go/sigmgmt/db"
)

var (
	name = "sigmgmt_createdb"
)

func main() {
	os.Exit(realMain())
}

func realMain() int {
	if err := integration.Init(name); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to init: %s\n", err)
		return 1
	}
	defer log.LogPanicAndExit()
	defer log.Flush()
	dbase, err := gorm.Open("sqlite3", os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %s\n", err)
		return 1
	}
	defer dbase.Close()
	db.MigrateDB(dbase)
	return 0
}
