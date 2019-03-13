// Copyright 2018 Anapaya Systems.

// Package postgres contains an implementation of the path DB interface for postgres databases.
package postgres

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"sort"
	"strings"
	"time"

	// pgx postgres driver
	_ "github.com/jackc/pgx/stdlib"

	"github.com/scionproto/scion/go/lib/addr"
	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/lib/ctrl/seg"
	"github.com/scionproto/scion/go/lib/infra/modules/db"
	"github.com/scionproto/scion/go/lib/pathdb"
	"github.com/scionproto/scion/go/lib/pathdb/query"
	"github.com/scionproto/scion/go/proto"
)

var _ pathdb.PathDB = (*Backend)(nil)

// Backend implements that path DB interface for postgres connections.
type Backend struct {
	*executor
	db *sql.DB
}

// New creates a new postgres path DB backend using the given connection string.
// The connection string can be anything that is supported by pgx.
// (https://godoc.org/github.com/jackc/pgx/stdlib)
func New(connection string) (*Backend, error) {
	db, err := sql.Open("pgx", connection)
	if err != nil {
		return nil, err
	}
	return NewFromDB(db), nil
}

// NewFromDB creates a new postgres path DB backend from the given db handle.
// The db handle must be a connection to a postgres database.
func NewFromDB(db *sql.DB) *Backend {
	return &Backend{
		executor: &executor{
			db: db,
		},
		db: db,
	}
}

func (b *Backend) BeginTransaction(ctx context.Context,
	opts *sql.TxOptions) (pathdb.Transaction, error) {

	tx, err := b.db.BeginTx(ctx, opts)
	if err != nil {
		return nil, common.NewBasicError("Failed to create transaction", err)
	}
	return &transaction{
		executor: &executor{
			db: tx,
		},
		tx: tx,
	}, nil
}

var _ (pathdb.Transaction) = (*transaction)(nil)

type transaction struct {
	*executor
	tx *sql.Tx
}

func (tx *transaction) Commit() error {
	return tx.tx.Commit()
}

func (tx *transaction) Rollback() error {
	return tx.tx.Rollback()
}

var _ (pathdb.ReadWrite) = (*executor)(nil)

type executor struct {
	db db.Sqler
}

func (e *executor) Insert(ctx context.Context, segMeta *seg.Meta) (int, error) {
	return e.InsertWithHPCfgIDs(ctx, segMeta, []*query.HPCfgID{&query.NullHpCfgID})
}

func (e *executor) InsertWithHPCfgIDs(ctx context.Context, segMeta *seg.Meta,
	hpCfgIDs []*query.HPCfgID) (int, error) {

	var inserted int
	err := db.DoInTx(ctx, e.db, func(ctx context.Context, tx *sql.Tx) error {
		var err error
		inserted, err = insertInternal(ctx, tx, segMeta, hpCfgIDs)
		return err
	})
	return inserted, err
}

func insertInternal(ctx context.Context, tx *sql.Tx, segMeta *seg.Meta,
	hpCfgIDs []*query.HPCfgID) (int, error) {

	ps := segMeta.Segment
	segID, err := ps.ID()
	if err != nil {
		return 0, err
	}
	fullId, err := ps.FullId()
	if err != nil {
		return 0, err
	}
	packedSeg, err := ps.Pack()
	if err != nil {
		return 0, err
	}
	exp := ps.MaxExpiry()
	info, _ := ps.InfoF()
	// TODO(lukedirtwalker): If possible we should do this also in the query below!
	existingFullId, err := getFullId(ctx, tx, segID)
	if err != nil {
		return 0, err
	}
	query := `INSERT INTO Segments (SegID, FullID, LastUpdated, InfoTs, Segment, MaxExpiry,
			StartIsdID, StartAsID, EndIsdID, EndAsID)
			SELECT $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
			WHERE NOT EXISTS
			(SELECT * FROM Segments AS existing WHERE existing.SegID = $1 AND existing.InfoTs > $4)
			ON CONFLICT (SegID)
			DO UPDATE SET FullID = $2, LastUpdated = $3, InfoTs = $4, Segment = $5, MaxExpiry = $6
			RETURNING RowID, xmax = 0`
	var segRowId int64
	var inserted bool
	err = tx.QueryRowContext(ctx, query, segID, fullId, time.Now(), info.Timestamp(), packedSeg,
		exp, ps.FirstIA().I, ps.FirstIA().A, ps.LastIA().I, ps.LastIA().A).
		Scan(&segRowId, &inserted)
	// If nothing was modified it means there is already a newer version of this segment in the DB.
	if err == sql.ErrNoRows {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	// Insert segType information.
	if err := insertType(ctx, tx, segRowId, segMeta.Type); err != nil {
		return 0, err
	}
	if inserted {
		// Insert all interfaces.
		if err = insertInterfaces(ctx, tx, ps.ASEntries, segRowId); err != nil {
			return 0, err
		}
	} else if !bytes.Equal(fullId, existingFullId) { // updated only
		// Delete all old interfaces and then insert the new ones.
		// Calculating the actual diffset would be better, but this is way easier to implement.
		_, err := tx.ExecContext(ctx, `DELETE FROM IntfToSeg WHERE SegRowID = $1`, segRowId)
		if err != nil {
			return 0, err
		}
		if err := insertInterfaces(ctx, tx, ps.ASEntries, segRowId); err != nil {
			return 0, err
		}
	}
	// Insert hpCfgId information.
	for _, hpCfgId := range hpCfgIDs {
		if err = insertHPCfgID(ctx, tx, segRowId, hpCfgId); err != nil {
			return 0, err
		}
	}
	return 1, nil
}

func (e *executor) Delete(ctx context.Context, params *query.Params) (int, error) {
	return e.deleteInTx(ctx, func(tx *sql.Tx) (sql.Result, error) {
		q, args := buildQuery(params)
		query := fmt.Sprintf(
			`DELETE FROM Segments WHERE RowId IN(SELECT toDel.RowID FROM (%s) AS toDel)`, q)
		return tx.ExecContext(ctx, query, args...)
	})
}

func (e *executor) DeleteExpired(ctx context.Context, now time.Time) (int, error) {
	return e.deleteInTx(ctx, func(tx *sql.Tx) (sql.Result, error) {
		delStmt := `DELETE FROM Segments WHERE MaxExpiry < $1`
		return tx.ExecContext(ctx, delStmt, now)
	})
}

func (e *executor) Get(ctx context.Context, params *query.Params) (query.Results, error) {
	stmt, args := buildQuery(params)
	rows, err := e.db.QueryContext(ctx, stmt, args...)
	if err != nil {
		return nil, common.NewBasicError("Error looking up path segment", err, "q", stmt)
	}
	defer rows.Close()
	var res query.Results
	prevID := -1
	var curRes *query.Result
	for rows.Next() {
		var segRowID int
		var rawSeg sql.RawBytes
		var lastUpdated time.Time
		hpCfgID := &query.HPCfgID{IA: addr.IA{}}
		err = rows.Scan(&segRowID, &rawSeg, &lastUpdated, &hpCfgID.IA.I, &hpCfgID.IA.A, &hpCfgID.ID)
		if err != nil {
			return nil, common.NewBasicError("Error reading DB response", err)
		}
		// Check if we have a new segment.
		if segRowID != prevID {
			if curRes != nil {
				res = append(res, curRes)
			}
			curRes = &query.Result{
				LastUpdate: lastUpdated,
			}
			var err error
			curRes.Seg, err = seg.NewSegFromRaw(common.RawBytes(rawSeg))
			if err != nil {
				return nil, common.NewBasicError("Error unmarshalling segment", err)
			}
		}
		// Append hpCfgID to result
		curRes.HpCfgIDs = append(curRes.HpCfgIDs, hpCfgID)
		prevID = segRowID
	}
	if curRes != nil {
		res = append(res, curRes)
	}
	sort.Sort(query.ByLastUpdate(res))
	return res, nil
}

func (e *executor) GetAll(ctx context.Context) (<-chan query.ResultOrErr, error) {
	stmt, args := buildQuery(nil)
	rows, err := e.db.QueryContext(ctx, stmt, args...)
	if err != nil {
		return nil, common.NewBasicError("Error looking up path segment", err, "q", stmt)
	}
	resCh := make(chan query.ResultOrErr)
	go func() {
		defer close(resCh)
		defer rows.Close()
		prevID := -1
		var curRes *query.Result
		for rows.Next() {
			var segRowID int
			var rawSeg sql.RawBytes
			var lastUpdated time.Time
			hpCfgID := &query.HPCfgID{IA: addr.IA{}}
			err = rows.Scan(&segRowID, &rawSeg, &lastUpdated,
				&hpCfgID.IA.I, &hpCfgID.IA.A, &hpCfgID.ID)
			if err != nil {
				resCh <- query.ResultOrErr{
					Err: common.NewBasicError("Error reading DB response", err)}
				return
			}
			// Check if we have a new segment.
			if segRowID != prevID {
				if curRes != nil {
					resCh <- query.ResultOrErr{Result: curRes}
				}
				curRes = &query.Result{
					LastUpdate: lastUpdated,
				}
				var err error
				curRes.Seg, err = seg.NewSegFromRaw(common.RawBytes(rawSeg))
				if err != nil {
					resCh <- query.ResultOrErr{
						Err: common.NewBasicError("Error unmarshalling segment", err)}
					return
				}
			}
			// Append hpCfgID to result
			curRes.HpCfgIDs = append(curRes.HpCfgIDs, hpCfgID)
			prevID = segRowID
		}
		if curRes != nil {
			resCh <- query.ResultOrErr{Result: curRes}
		}
	}()
	return resCh, nil
}

func (e *executor) InsertNextQuery(ctx context.Context, dst addr.IA,
	nextQuery time.Time) (bool, error) {

	query := `INSERT INTO NextQuery (IsdID, AsID, NextQuery)
		SELECT $1, $2, $3
		WHERE NOT EXISTS
			(SELECT * FROM NextQuery AS existing
			WHERE existing.IsdID = $1 AND existing.AsID = $2 AND existing.NextQuery > $3)
		ON CONFLICT(IsdID, AsID) DO UPDATE SET NextQuery = $3`
	var r sql.Result
	err := db.DoInTx(ctx, e.db, func(ctx context.Context, tx *sql.Tx) error {
		var err error
		r, err = tx.ExecContext(ctx, query, dst.I, dst.A, nextQuery)
		return err
	})
	if err != nil {
		return false, common.NewBasicError("Failed to execute insert NextQuery stmt", err)
	}
	n, err := r.RowsAffected()
	return n > 0, err
}

func (e *executor) GetNextQuery(ctx context.Context, dst addr.IA) (*time.Time, error) {

	query := `SELECT NextQuery from NextQuery WHERE IsdID = $1 AND AsID = $2`
	rows, err := e.db.QueryContext(ctx, query, dst.I, dst.A)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if !rows.Next() {
		return nil, nil
	}
	var next time.Time
	rows.Scan(&next)
	return &next, nil
}

func (e *executor) deleteInTx(ctx context.Context,
	delFunc func(tx *sql.Tx) (sql.Result, error)) (int, error) {

	var res sql.Result
	err := db.DoInTx(ctx, e.db, func(ctx context.Context, tx *sql.Tx) error {
		var err error
		res, err = delFunc(tx)
		return err
	})
	if err != nil {
		return 0, err
	}
	deleted, _ := res.RowsAffected()
	return int(deleted), nil
}

func buildQuery(params *query.Params) (string, []interface{}) {
	var args []interface{}
	query := []string{
		"SELECT DISTINCT s.RowID, s.Segment, s.LastUpdated," +
			" h.IsdID, h.AsID, h.CfgID FROM Segments s",
		"JOIN HpCfgIds h ON h.SegRowID=s.RowID",
	}
	if params == nil {
		query = append(query, " ORDER BY s.RowID")
		return strings.Join(query, "\n"), args
	}
	eq := func(name string, val interface{}) string {
		q := fmt.Sprintf("%s=$%d", name, len(args)+1)
		args = append(args, val)
		return q
	}
	joins := []string{}
	where := []string{}
	if len(params.SegIDs) > 0 {
		subQ := make([]string, 0, len(params.SegIDs))
		for _, segID := range params.SegIDs {
			subQ = append(subQ, eq("s.SegID", segID))
		}
		where = append(where, fmt.Sprintf("(%s)", strings.Join(subQ, " OR ")))
	}
	if len(params.SegTypes) > 0 {
		joins = append(joins, "JOIN SegTypes t ON t.SegRowID=s.RowID")
		subQ := []string{}
		for _, segType := range params.SegTypes {
			subQ = append(subQ, eq("t.Type", segType))
		}
		where = append(where, fmt.Sprintf("(%s)", strings.Join(subQ, " OR ")))
	}
	if len(params.HpCfgIDs) > 0 {
		subQ := []string{}
		for _, hpCfgID := range params.HpCfgIDs {
			subQ = append(subQ, fmt.Sprintf("(%s AND %s AND %s)",
				eq("h.IsdID", hpCfgID.IA.I), eq("h.AsID", hpCfgID.IA.A), eq("h.CfgID", hpCfgID.ID)))
		}
		where = append(where, fmt.Sprintf("(%s)", strings.Join(subQ, " OR ")))
	}
	if len(params.Intfs) > 0 {
		joins = append(joins, "JOIN IntfToSeg i ON i.SegRowID=s.RowID")
		subQ := []string{}
		for _, spec := range params.Intfs {
			subQ = append(subQ, fmt.Sprintf("(%s AND %s AND %s)",
				eq("i.IsdID", spec.IA.I), eq("i.AsID", spec.IA.A), eq("i.IntfID", spec.IfID)))
		}
		where = append(where, fmt.Sprintf("(%s)", strings.Join(subQ, " OR ")))
	}
	if len(params.StartsAt) > 0 {
		subQ := []string{}
		for _, as := range params.StartsAt {
			if as.A == 0 {
				subQ = append(subQ, fmt.Sprintf("(%s)", eq("s.StartIsdID", as.I)))
			} else {
				subQ = append(subQ, fmt.Sprintf("(%s AND %s)",
					eq("s.StartIsdID", as.I), eq("s.StartAsID", as.A)))
			}
		}
		where = append(where, fmt.Sprintf("(%s)", strings.Join(subQ, " OR ")))
	}
	if len(params.EndsAt) > 0 {
		subQ := []string{}
		for _, as := range params.EndsAt {
			if as.A == 0 {
				subQ = append(subQ, fmt.Sprintf("(%s)", eq("s.EndIsdID", as.I)))
			} else {
				subQ = append(subQ, fmt.Sprintf("(%s AND %s)",
					eq("s.EndIsdID", as.I), eq("s.EndAsID", as.A)))
			}
		}
		where = append(where, fmt.Sprintf("(%s)", strings.Join(subQ, " OR ")))
	}
	if params.MinLastUpdate != nil {
		where = append(where, fmt.Sprintf("(s.LastUpdated>$%d)", len(args)+1))
		args = append(args, params.MinLastUpdate)
	}
	// Assemble the query.
	if len(joins) > 0 {
		query = append(query, strings.Join(joins, "\n"))
	}
	if len(where) > 0 {
		query = append(query, fmt.Sprintf("WHERE %s", strings.Join(where, " AND\n")))
	}
	query = append(query, " ORDER BY s.RowID")
	return strings.Join(query, "\n"), args
}

func insertInterfaces(ctx context.Context, tx *sql.Tx,
	ases []*seg.ASEntry, segRowID int64) error {

	stmtStr := `INSERT INTO IntfToSeg
		(IsdID, ASID, IntfID, SegRowID) VALUES ($1, $2, $3, $4) ON CONFLICT DO NOTHING`
	stmt, err := tx.PrepareContext(ctx, stmtStr)
	if err != nil {
		return common.NewBasicError("Failed to prepare insert into IntfToSeg", err)
	}
	defer stmt.Close()
	for _, as := range ases {
		ia := as.IA()
		for idx, hop := range as.HopEntries {
			hof, err := hop.HopField()
			if err != nil {
				return common.NewBasicError("Failed to extract hop field", err)
			}
			if hof.ConsIngress != 0 {
				_, err = stmt.ExecContext(ctx, ia.I, ia.A, hof.ConsIngress, segRowID)
				if err != nil {
					return common.NewBasicError("Failed to insert Ingress into IntfToSeg", err)
				}
			}
			// Only insert the Egress interface for the first hop entry in an AS entry.
			if idx == 0 && hof.ConsEgress != 0 {
				_, err := stmt.ExecContext(ctx, ia.I, ia.A, hof.ConsEgress, segRowID)
				if err != nil {
					return common.NewBasicError("Failed to insert Egress into IntfToSeg", err)
				}
			}
		}
	}
	return nil
}

func insertType(ctx context.Context, tx *sql.Tx, segRowId int64,
	segType proto.PathSegType) error {

	query := `INSERT INTO SegTypes (SegRowID, Type) VALUES ($1, $2)
		ON CONFLICT DO NOTHING`
	_, err := tx.ExecContext(ctx, query, segRowId, segType)
	if err != nil {
		return common.NewBasicError("Failed to insert type", err)
	}
	return nil
}

func insertHPCfgID(ctx context.Context, tx *sql.Tx, segRowID int64,
	hpCfgID *query.HPCfgID) error {

	query := `INSERT INTO HpCfgIds (SegRowID, IsdID, AsID, CfgID) VALUES ($1, $2, $3, $4)
		ON CONFLICT DO NOTHING`
	_, err := tx.ExecContext(ctx, query, segRowID, hpCfgID.IA.I, hpCfgID.IA.A, hpCfgID.ID)
	if err != nil {
		return common.NewBasicError("Failed to insert hpCfgID", err)
	}
	return nil
}

func getFullId(ctx context.Context, tx *sql.Tx, segId common.RawBytes) (common.RawBytes, error) {
	query := `SELECT FullID FROM Segments WHERE SegID = $1`
	rows, err := tx.QueryContext(ctx, query, segId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if !rows.Next() {
		return nil, nil
	}
	var fullId common.RawBytes
	rows.Scan(&fullId)
	return fullId, nil
}
