// Copyright 2018 Anapaya Systems

package db

const (
	SchemaVersion = 1
	Schema        = `
	CREATE TABLE Sites (
		Name TEXT PRIMARY KEY NOT NULL,
		MetricsPort INTEGER NOT NULL,
		VHost TEXT NOT NULL
	);

	CREATE TABLE Hosts (
		Name TEXT NOT NULL,
		User TEXT NOT NULL,
		Key TEXT NOT NULL,
		Site TEXT REFERENCES Sites ON DELETE CASCADE ON UPDATE CASCADE,
		PRIMARY KEY (Site, Name, User, Key)
	);

	CREATE TABLE ASEntries (
		Name TEXT NOT NULL,
		IsdID INTEGER NOT NULL,
		AsID INTEGER NOT NULL,
		Site TEXT REFERENCES Sites ON DELETE CASCADE ON UPDATE CASCADE,
		Policy TEXT NOT NULL,
		PRIMARY KEY (Site, IsdID, AsID)
	);

	CREATE TABLE ASConfig (
		Name TEXT PRIMAY KEY NOT NULL,
		Value TEXT NOT NULL,
		IsdID INTEGER NOT NULL,
		AsID INTEGER NOT NULL,
		Site TEXT REFERENCES Sites ON DELETE CASCADE ON UPDATE CASCADE,
		FOREIGN KEY (Site, IsdID, AsID) REFERENCES ASEntries ON DELETE CASCADE ON UPDATE CASCADE,
		PRIMARY KEY (Site, IsdID, AsID, Name)
	);

	CREATE TABLE SIGs (
		Name TEXT NOT NULL,
		Address TEXT NOT NULL,
		CtrlPort INTEGER NOT NULL,
		EncapPort INTEGER NOT NULL,
		Site TEXT REFERENCES Sites,
		IsdID INTEGER NOT NULL,
		AsID INTEGER NOT NULL,
		FOREIGN KEY (Site, IsdID, AsID) REFERENCES ASEntries ON DELETE CASCADE ON UPDATE CASCADE,
		PRIMARY KEY (Site, IsdID, AsID, Name)
	);

	CREATE TABLE Networks (
		ID INTEGER PRIMARY KEY AUTOINCREMENT,
		CIDR TEXT NOT NULL,
		Site TEXT REFERENCES Sites ON DELETE CASCADE ON UPDATE CASCADE,
		IsdID INTEGER NOT NULL,
		AsID INTEGER NOT NULL,
		FOREIGN KEY (Site, IsdID, AsID) REFERENCES ASEntries ON DELETE CASCADE ON UPDATE CASCADE,
		UNIQUE (ID, Site, IsdID, AsID, CIDR)
	);

	CREATE TABLE Filters (
		Name TEXT NOT NULL,
		Filter TEXT NOT NULL,
		Site TEXT REFERENCES Sites ON DELETE CASCADE ON UPDATE CASCADE,
		PRIMARY KEY (Site, Name)
	);

	CREATE TABLE Sessions (
		Name INTEGER NOT NULL,
		Site TEXT REFERENCES Sites ON DELETE CASCADE ON UPDATE CASCADE,
		IsdID INTEGER NOT NULL,
		AsID INTEGER NOT NULL,
		FilterName TEXT NOT NULL,
		FOREIGN KEY (Site, FilterName) REFERENCES Filters ON DELETE CASCADE ON UPDATE CASCADE,
		FOREIGN KEY (Site, IsdID, AsID) REFERENCES ASEntries ON DELETE CASCADE ON UPDATE CASCADE,
		PRIMARY KEY (Site, IsdID, AsID, Name)
	);

	CREATE TABLE SessionAliases (
		Name TEXT NOT NULL,
		Sessions TEXT NOT NULL,
		Site TEXT REFERENCES Sites ON DELETE CASCADE ON UPDATE CASCADE,
		IsdID INTEGER NOT NULL,
		AsID INTEGER NOT NULL,
		FOREIGN KEY (Site, IsdID, AsID) REFERENCES ASEntries ON DELETE CASCADE ON UPDATE CASCADE,
		PRIMARY KEY (Site, IsdID, AsID, Name)
	);
	`
)
