// Copyright 2018 Anapaya Systems.

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
	"github.com/scionproto/scion/go/lib/revcache"
)

var _ revcache.RevCache = (*pgRevCache)(nil)

type pgRevCache struct {
	db *sql.DB
}

func New(connection string) (*pgRevCache, error) {
	db, err := sql.Open("pgx", connection)
	if err != nil {
		return nil, err
	}
	return &pgRevCache{
		db: db,
	}, nil
}

func (c *pgRevCache) Get(ctx context.Context,
	k *revcache.Key) (*path_mgmt.SignedRevInfo, bool, error) {

	query := `
		SELECT RawSignedRev FROM Revocations 
		WHERE IsdID=$1 AND AsID=$2 AND IfID=$3
		AND Expiration > NOW()`
	rows, err := c.db.QueryContext(ctx, query, k.IA.I, k.IA.A, k.IfId)
	if err != nil {
		return nil, false, err
	}
	defer rows.Close()
	if !rows.Next() {
		return nil, false, nil
	}
	var raw common.RawBytes
	rows.Scan(&raw)
	sr, err := path_mgmt.NewSignedRevInfoFromRaw(raw)
	return sr, true, err
}

func (c *pgRevCache) GetAll(ctx context.Context,
	keys map[revcache.Key]struct{}) ([]*path_mgmt.SignedRevInfo, error) {

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
	var revs []*path_mgmt.SignedRevInfo
	for rows.Next() {
		var raw common.RawBytes
		rows.Scan(&raw)
		sr, err := path_mgmt.NewSignedRevInfoFromRaw(raw)
		if err != nil {
			return nil, err
		}
		revs = append(revs, sr)
	}
	return revs, nil
}

func (c *pgRevCache) Insert(ctx context.Context, rev *path_mgmt.SignedRevInfo) (bool, error) {
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

func (c *pgRevCache) DeleteExpired(ctx context.Context) (int64, error) {
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
