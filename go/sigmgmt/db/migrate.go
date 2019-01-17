// Copyright 2018 Anapaya Systems

package db

import "github.com/jinzhu/gorm"

// MigrateDB migrates the sqlite DB to reflect the current models
func MigrateDB(dbase *gorm.DB) {
	dbase.Exec("PRAGMA foreign_keys = ON;")
	dbase.AutoMigrate(&Site{})
	dbase.AutoMigrate(&Host{})
	dbase.AutoMigrate(&SiteNetwork{})
	dbase.AutoMigrate(&PathSelector{})
	dbase.AutoMigrate(&ASEntry{})
	dbase.AutoMigrate(&SIG{})
	dbase.AutoMigrate(&Network{})
	dbase.AutoMigrate(&TrafficPolicy{})
	dbase.AutoMigrate(&TrafficClass{})
	dbase.AutoMigrate(&PathPolicyFile{})
}
