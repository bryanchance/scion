// Copyright 2018 Anapaya Systems.

// Package pgrevcache contains a revcache implementation fro postgres databases.
package pgrevcache

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	// pgx postgres driver
	_ "github.com/jackc/pgx/stdlib"

	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/lib/ctrl/path_mgmt"
	"github.com/scionproto/scion/go/lib/log"
	"github.com/scionproto/scion/go/lib/revcache"
)

var _ revcache.RevCache = (*PgRevCache)(nil)

// PgRevCache implements the revocation cache interface.
type PgRevCache struct {
	db *sql.DB
}

// New creates a new postgres revocation cache backend using the given connection string.
// The connection string can be anything that is supported by pgx
// (https://godoc.org/github.com/jackc/pgx/stdlib)
func New(connection string) (*PgRevCache, error) {
	db, err := sql.Open("pgx", connection)
	if err != nil {
		return nil, err
	}
	return &PgRevCache{
		db: db,
	}, nil
}

// NewFromDB creates a new postgres revocation backend using the given db handle.
// The db handle must be a connection to a postgres database.
func NewFromDB(db *sql.DB) *PgRevCache {
	return &PgRevCache{
		db: db,
	}
}

func (c *PgRevCache) Get(ctx context.Context, keys revcache.KeySet) (revcache.Revocations, error) {
	if len(keys) == 0 {
		return nil, nil
	}
	var ingroups []string
	var args []interface{}
	i := 1
	for k := range keys {
		placeholders := fmt.Sprintf("($%d, $%d, $%d)", i, i+1, i+2)
		ingroups = append(ingroups, placeholders)
		args = append(args, k.IA.I, k.IA.A, k.IfId)
		i += 3
	}
	query := fmt.Sprintf(`
		SELECT RawSignedRev FROM Revocations 
		WHERE (IsdID, AsID, IfID) IN (%s)
		AND Expiration > NOW()`,
		strings.Join(ingroups, ","))
	rows, err := c.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	revs := make(revcache.Revocations)
	for rows.Next() {
		var raw common.RawBytes
		rows.Scan(&raw)
		sr, err := path_mgmt.NewSignedRevInfoFromRaw(raw)
		if err != nil {
			return nil, err
		}
		info, err := sr.RevInfo()
		if err != nil {
			return nil, err
		}
		revs[*revcache.NewKey(info.IA(), info.IfID)] = sr
	}
	return revs, nil
}

func (c *PgRevCache) GetAll(ctx context.Context) (revcache.ResultChan, error) {
	resCh := make(chan revcache.RevOrErr, 50)
	query := `SELECT RawSignedRev FROM Revocations
			  WHERE Expiration > NOW()`
	rows, err := c.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	go func() {
		defer log.LogPanicAndExit()
		defer rows.Close()
		for rows.Next() {
			var raw common.RawBytes
			rows.Scan(&raw)
			sr, err := path_mgmt.NewSignedRevInfoFromRaw(raw)
			resCh <- revcache.RevOrErr{
				Rev: sr,
				Err: err,
			}
		}
		close(resCh)
	}()
	return resCh, nil
}

func (c *PgRevCache) Insert(ctx context.Context, rev *path_mgmt.SignedRevInfo) (bool, error) {
	newInfo, err := rev.RevInfo()
	if err != nil {
		panic(err)
	}
	ttl := newInfo.Expiration().Sub(time.Now())
	if ttl <= 0 {
		return false, nil
	}
	k := revcache.NewKey(newInfo.IA(), newInfo.IfID)
	packedRev, err := rev.Pack()
	if err != nil {
		return false, err
	}
	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}
	query := `
		INSERT INTO Revocations 
			(IsdID, AsID, IfID, LinkType, RawTimeStamp, RawTTL, RawSignedRev, Expiration)
		SELECT $1, $2, $3, $4, $5, $6, $7, $8
		WHERE NOT EXISTS
			(SELECT * FROM Revocations AS existing
			 WHERE existing.IsdID = $1 AND existing.AsID = $2 AND existing.IfID = $3 
			 AND existing.RawTimeStamp >= $5)
		ON CONFLICT (IsdID, AsID, IfID) DO UPDATE 
			SET RawTimeStamp = $5, RawTTL = $6, RawSignedRev = $7, Expiration = $8
		RETURNING xmax = 0`
	var inserted bool
	err = tx.QueryRowContext(ctx, query, k.IA.I, k.IA.A, k.IfId, newInfo.LinkType,
		newInfo.RawTimestamp, newInfo.RawTTL, packedRev, newInfo.Expiration()).Scan(&inserted)
	// If nothing was modified it means there is already a newer version of this revocation
	// in the DB.
	if err == sql.ErrNoRows {
		return false, tx.Commit()
	}
	if err != nil {
		return false, err
	}
	err = tx.Commit()
	if err != nil {
		return false, err
	}
	return true, nil
}

func (c *PgRevCache) DeleteExpired(ctx context.Context) (int64, error) {
	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	query := `DELETE FROM Revocations WHERE Expiration < NOW()`
	r, err := tx.ExecContext(ctx, query)
	if err != nil {
		tx.Rollback()
		return 0, err
	}
	if err = tx.Commit(); err != nil {
		return 0, err
	}
	return r.RowsAffected()
}
