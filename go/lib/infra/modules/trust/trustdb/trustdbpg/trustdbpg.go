// Copyright 2018 Anapaya Systems

package trustdbpg

import (
	"context"
	"database/sql"

	_ "github.com/lib/pq"

	"github.com/scionproto/scion/go/lib/addr"
	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/lib/infra/modules/trust/trustdb"
	"github.com/scionproto/scion/go/lib/scrypto"
	"github.com/scionproto/scion/go/lib/scrypto/cert"
	"github.com/scionproto/scion/go/lib/scrypto/trc"
)

type sqler interface {
	ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
	QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...interface{}) *sql.Row
}

var _ trustdb.TrustDB = (*trustDB)(nil)

type trustDB struct {
	db *sql.DB
	*executor
}

func New(connection string) (*trustDB, error) {
	db, err := sql.Open("postgres", connection)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, common.NewBasicError("Initial DB ping failed, connection broken?", err)
	}
	return &trustDB{
		db: db,
		executor: &executor{
			db: db,
		},
	}, nil
}

func (db *trustDB) Close() error {
	return db.db.Close()
}

func (db *trustDB) BeginTransaction(ctx context.Context,
	opts *sql.TxOptions) (trustdb.Transaction, error) {

	tx, err := db.db.BeginTx(ctx, opts)
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

type executor struct {
	db sqler
}

func (db *executor) GetIssCertVersion(ctx context.Context, ia addr.IA,
	version uint64) (*cert.Certificate, error) {

	if version == scrypto.LatestVer {
		return db.GetIssCertMaxVersion(ctx, ia)
	}
	var raw common.RawBytes
	query := `SELECT Data FROM IssuerCerts WHERE IsdID=$1 AND AsID=$2 AND Version=$3`
	err := db.db.QueryRowContext(ctx, query, ia.I, ia.A, version).Scan(&raw)
	return certFromQueryRow(raw, ia, version, err)
}

func (db *executor) GetIssCertMaxVersion(ctx context.Context,
	ia addr.IA) (*cert.Certificate, error) {

	var raw common.RawBytes
	query := `SELECT Data FROM IssuerCerts WHERE IsdID=$1 AND AsID=$2 ORDER BY Version desc LIMIT 1`
	err := db.db.QueryRowContext(ctx, query, ia.I, ia.A).Scan(&raw)
	return certFromQueryRow(raw, ia, scrypto.LatestVer, err)
}

func (db *executor) InsertIssCert(ctx context.Context, crt *cert.Certificate) (int64, error) {
	return db.inTx(ctx, func(ctx context.Context, tx *sql.Tx) (int64, error) {
		ra, _, err := insertIssCert(ctx, tx, crt)
		return ra, err
	})
}

func (db *executor) GetLeafCertVersion(ctx context.Context, ia addr.IA,
	version uint64) (*cert.Certificate, error) {

	if version == scrypto.LatestVer {
		return db.GetLeafCertMaxVersion(ctx, ia)
	}
	var raw common.RawBytes
	query := `SELECT Data FROM LeafCerts WHERE IsdID=$1 AND AsID=$2 AND Version=$3`
	err := db.db.QueryRowContext(ctx, query, ia.I, ia.A, version).Scan(&raw)
	return certFromQueryRow(raw, ia, version, err)
}

func (db *executor) GetLeafCertMaxVersion(ctx context.Context,
	ia addr.IA) (*cert.Certificate, error) {

	var raw common.RawBytes
	query := `SELECT Data FROM LeafCerts WHERE IsdID=$1 AND AsID=$2 ORDER BY Version desc LIMIT 1`
	err := db.db.QueryRowContext(ctx, query, ia.I, ia.A).Scan(&raw)
	return certFromQueryRow(raw, ia, scrypto.LatestVer, err)
}

func (db *executor) InsertLeafCert(ctx context.Context, crt *cert.Certificate) (int64, error) {
	return db.inTx(ctx, func(ctx context.Context, tx *sql.Tx) (int64, error) {
		return insertLeafCert(ctx, tx, crt)
	})
}

func (db *executor) GetChainVersion(ctx context.Context, ia addr.IA,
	version uint64) (*cert.Chain, error) {

	if version == scrypto.LatestVer {
		return db.GetChainMaxVersion(ctx, ia)
	}
	query := `
	SELECT Data, 0 AS idx FROM LeafCerts WHERE IsdID=$1 AND AsID=$2 AND Version=$3
	UNION ALL
	SELECT ic.Data, ch.OrderKey AS idx FROM IssuerCerts ic, Chains ch
	WHERE ic.RowID IN (
		SELECT IssCertsRowID FROM Chains WHERE IsdID=$1 AND AsID=$2 AND Version=$3
	)
	ORDER BY idx`
	rows, err := db.db.QueryContext(ctx, query, ia.I, ia.A, version)
	return parseChain(rows, err)
}

func (db *executor) GetChainMaxVersion(ctx context.Context, ia addr.IA) (*cert.Chain, error) {
	query := `
	WITH m AS (
		SELECT MAX(Version) AS v FROM CHAINS WHERE IsdID=$1 AND AsID=$2
	)
	SELECT Data, 0 AS idx FROM LeafCerts lc, m WHERE lc.IsdID=$1 AND lc.AsID=$2 AND Version=m.v
	UNION ALL
	SELECT ic.Data, ch.OrderKey AS idx FROM IssuerCerts ic, Chains ch, m
	WHERE ic.RowID IN (
		SELECT IssCertsRowID FROM Chains c, m WHERE c.IsdID=$1 AND c.AsID=$2 AND Version=m.v
	)
	AND ch.IsdID=$1 AND ch.AsID=$2 AND ch.Version=m.v
	ORDER BY idx`
	rows, err := db.db.QueryContext(ctx, query, ia.I, ia.A)
	return parseChain(rows, err)
}

func (db *executor) GetAllChains(ctx context.Context) ([]*cert.Chain, error) {
	query := `
	SELECT cic.OrderKey, cic.Data AS IData, lc.Data AS LData FROM (
		SELECT ch.*, ic.Data FROM Chains ch
		LEFT JOIN IssuerCerts ic ON ch.IssCertsRowID = ic.RowID) as cic
	INNER JOIN LeafCerts lc USING (IsdID, AsID, Version)
	ORDER BY cic.IsdID, cic.AsID, cic.Version, cic.OrderKey`
	rows, err := db.db.QueryContext(ctx, query)
	if err != nil {
		return nil, common.NewBasicError("Chain Database access error", err)
	}
	defer rows.Close()
	var chains []*cert.Chain
	var leafRaw common.RawBytes
	var issCertRaw common.RawBytes
	var orderKey int64
	var lastOrderKey int64 = 1
	currentCerts := make([]*cert.Certificate, 0, 2)
	for rows.Next() {
		err = rows.Scan(&orderKey, &issCertRaw, &leafRaw)
		if err != nil {
			return nil, err
		}
		// Wrap around means we start processing a new chain entry.
		if orderKey <= lastOrderKey {
			if len(currentCerts) > 0 {
				chain, err := cert.ChainFromSlice(currentCerts)
				if err != nil {
					return nil, err
				}
				chains = append(chains, chain)
				currentCerts = currentCerts[:0]
			}
			// While the leaf entry is in every result row,
			// it has to be the first entry in the chain we are building.
			crt, err := cert.CertificateFromRaw(leafRaw)
			if err != nil {
				return nil, err
			}
			currentCerts = append(currentCerts, crt)
		}
		crt, err := cert.CertificateFromRaw(issCertRaw)
		if err != nil {
			return nil, err
		}
		currentCerts = append(currentCerts, crt)
		lastOrderKey = orderKey
	}
	if len(currentCerts) > 0 {
		chain, err := cert.ChainFromSlice(currentCerts)
		if err != nil {
			return nil, err
		}
		chains = append(chains, chain)
	}
	return chains, nil
}

func (db *executor) InsertChain(ctx context.Context, chain *cert.Chain) (int64, error) {
	return db.inTx(ctx, func(ctx context.Context, tx *sql.Tx) (int64, error) {
		return insertChain(ctx, tx, chain)
	})
}

func insertChain(ctx context.Context, tx *sql.Tx, chain *cert.Chain) (int64, error) {
	var err error
	if _, err = insertLeafCert(ctx, tx, chain.Leaf); err != nil {
		return 0, err
	}
	var issRowId int64
	if _, issRowId, err = insertIssCert(ctx, tx, chain.Issuer); err != nil {
		return 0, err
	}
	query := `
	INSERT INTO Chains (IsdID, AsID, Version, OrderKey, IssCertsRowID) 
	VALUES ($1, $2, $3, $4, $5)
	ON CONFLICT (IsdID, AsID, Version, OrderKey) DO NOTHING`
	ia, ver := chain.IAVer()
	res, err := tx.ExecContext(ctx, query, ia.I, ia.A, ver, 1, issRowId)
	if err != nil {
		return 0, common.NewBasicError("Failed to insert chain", err)
	}
	return res.RowsAffected()
}

func (db *executor) GetTRCVersion(ctx context.Context,
	isd addr.ISD, version uint64) (*trc.TRC, error) {

	if version == scrypto.LatestVer {
		return db.GetTRCMaxVersion(ctx, isd)
	}
	query := `SELECT Data FROM TRCs WHERE IsdID = $1 AND Version = $2`
	var raw common.RawBytes
	err := db.db.QueryRowContext(ctx, query, isd, version).Scan(&raw)
	return trcFromQueryRow(err, raw, isd, version)
}

func (db *executor) GetTRCMaxVersion(ctx context.Context, isd addr.ISD) (*trc.TRC, error) {
	query := `SELECT Data FROM TRCs WHERE IsdID=$1 ORDER BY Version desc LIMIT 1`
	var raw common.RawBytes
	err := db.db.QueryRowContext(ctx, query, isd).Scan(&raw)
	return trcFromQueryRow(err, raw, isd, scrypto.LatestVer)
}

func (db *executor) InsertTRC(ctx context.Context, trcobj *trc.TRC) (int64, error) {
	return db.inTx(ctx, func(ctx context.Context, tx *sql.Tx) (int64, error) {
		return insertTRC(ctx, tx, trcobj)
	})
}

func (db *executor) GetAllTRCs(ctx context.Context) ([]*trc.TRC, error) {
	rows, err := db.db.QueryContext(ctx, `SELECT Data FROM TRCs`)
	if err != nil {
		return nil, common.NewBasicError("TRC Database access error", err)
	}
	defer rows.Close()
	var trcs []*trc.TRC
	var raw common.RawBytes
	for rows.Next() {
		err = rows.Scan(&raw)
		if err != nil {
			return nil, common.NewBasicError("Failed to scan rows", err)
		}
		trcobj, err := trc.TRCFromRaw(raw, false)
		if err != nil {
			return nil, common.NewBasicError("TRC parse error", err)
		}
		trcs = append(trcs, trcobj)
	}
	return trcs, nil
}

func (db *executor) GetCustKey(ctx context.Context, ia addr.IA) (common.RawBytes, uint64, error) {
	var key common.RawBytes
	var version uint64
	query := `SELECT Key, Version FROM CustKeys WHERE IsdID = $1 AND AsID = $2`
	err := db.db.QueryRowContext(ctx, query, ia.I, ia.A).Scan(&key, &version)
	if err == sql.ErrNoRows {
		return nil, 0, nil
	}
	if err != nil {
		return nil, 0, common.NewBasicError("Failed to look up cust key", err)
	}
	return key, version, nil
}

func (db *executor) InsertCustKey(ctx context.Context, ia addr.IA,
	version uint64, key common.RawBytes, oldVersion uint64) error {

	if version == oldVersion {
		return common.NewBasicError("Same version as oldVersion not allowed",
			nil, "version", version)
	}
	_, err := db.inTx(ctx, func(ctx context.Context, tx *sql.Tx) (int64, error) {
		return insertCustKey(ctx, tx, ia, version, key, oldVersion)
	})
	return err
}

func (db *executor) inTx(ctx context.Context,
	insert func(ctx context.Context, tx *sql.Tx) (int64, error)) (int64, error) {

	tx, ok := db.db.(*sql.Tx)
	if !ok {
		var err error
		if tx, err = db.db.(*sql.DB).BeginTx(ctx, nil); err != nil {
			return 0, common.NewBasicError("Failed to create tx", err)
		}
	}
	affected, err := insert(ctx, tx)
	if err != nil {
		if !ok {
			tx.Rollback()
		}
		return 0, common.NewBasicError("Failed to insert in tx", err)
	}
	if !ok {
		if err := tx.Commit(); err != nil {
			return 0, common.NewBasicError("Failed to commit", err)
		}
	}
	return affected, nil
}

type transaction struct {
	tx *sql.Tx
	*executor
}

func (db *transaction) Commit() error {
	if db.tx == nil {
		return common.NewBasicError("Transaction already done", nil)
	}
	err := db.tx.Commit()
	if err != nil {
		return common.NewBasicError("Failed to commit transaction", err)
	}
	db.tx = nil
	return nil
}

func (db *transaction) Rollback() error {
	if db.tx == nil {
		return common.NewBasicError("Transaction already done", nil)
	}
	err := db.tx.Rollback()
	db.tx = nil
	if err != nil {
		return common.NewBasicError("Failed to rollback transaction", err)
	}
	return nil
}

func certFromQueryRow(raw common.RawBytes, ia addr.IA,
	v uint64, err error) (*cert.Certificate, error) {

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, common.NewBasicError("Cert Database access error", err)
	}
	crt, err := cert.CertificateFromRaw(raw)
	if err != nil {
		if v == scrypto.LatestVer {
			return nil, common.NewBasicError("Cert parse error", err, "ia", ia, "version", "max")
		}
		return nil, common.NewBasicError("Cert parse error", err, "ia", ia, "version", v)
	}
	return crt, nil
}

func insertIssCert(ctx context.Context, tx *sql.Tx, crt *cert.Certificate) (int64, int64, error) {
	raw, err := crt.JSON(false)
	if err != nil {
		return 0, 0, common.NewBasicError("Unable to convert issuer cert to JSON", err)
	}
	query := `
	WITH ins AS (
		INSERT INTO IssuerCerts (IsdID, AsID, Version, Data) VALUES ($1, $2, $3, $4)
		ON CONFLICT (IsdID, AsID, Version) DO NOTHING
		RETURNING RowID
	)
	SELECT ins.RowId, 'i' FROM ins WHERE EXISTS(SELECT ins.RowID)
	UNION ALL
	SELECT RowId, 'e' FROM IssuerCerts WHERE IsdID = $1 AND AsId=$2 AND Version=$3 AND 
		NOT EXISTS(SELECT * FROM ins)`
	var rowId int64
	var mode string
	err = tx.QueryRowContext(ctx, query, crt.Subject.I, crt.Subject.A, crt.Version, raw).
		Scan(&rowId, &mode)
	if err != nil {
		return 0, 0, common.NewBasicError("Failed to insert issuer cert", err)
	}
	var rowsAffected int64
	if mode == "i" {
		rowsAffected = 1
	}
	return rowsAffected, rowId, nil
}

func insertLeafCert(ctx context.Context, tx *sql.Tx, crt *cert.Certificate) (int64, error) {
	query := `
	INSERT INTO LeafCerts (IsdID, AsID, Version, Data) VALUES ($1, $2, $3, $4)
	ON CONFLICT (IsdID, AsID, Version) DO NOTHING`
	raw, err := crt.JSON(false)
	if err != nil {
		return 0, common.NewBasicError("Unable to convert leaf cert to JSON", err)
	}
	res, err := tx.ExecContext(ctx, query, crt.Subject.I, crt.Subject.A, crt.Version, raw)
	if err != nil {
		return 0, common.NewBasicError("Failed to insert leaf cert", err)
	}
	return res.RowsAffected()
}

func parseChain(rows *sql.Rows, err error) (*cert.Chain, error) {
	if err != nil {
		return nil, common.NewBasicError("Chain Database access error", err)
	}
	defer rows.Close()
	certs := make([]*cert.Certificate, 0, 2)
	var raw common.RawBytes
	var pos int64
	for rows.Next() {
		if err = rows.Scan(&raw, &pos); err != nil {
			return nil, err
		}
		crt, err := cert.CertificateFromRaw(raw)
		if err != nil {
			return nil, err
		}
		certs = append(certs, crt)
	}
	if len(certs) == 0 {
		return nil, nil
	}
	return cert.ChainFromSlice(certs)
}

func trcFromQueryRow(err error, raw common.RawBytes,
	isd addr.ISD, v uint64) (*trc.TRC, error) {

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, common.NewBasicError("TRC Database access error", err)
	}
	trcobj, err := trc.TRCFromRaw(raw, false)
	if err != nil {
		if v == scrypto.LatestVer {
			return nil, common.NewBasicError("TRC parse error", err, "isd", isd, "version", "max")
		}
		return nil, common.NewBasicError("TRC parse error", err, "isd", isd, "version", v)
	}
	return trcobj, nil
}

func insertTRC(ctx context.Context, tx *sql.Tx, trcobj *trc.TRC) (int64, error) {
	raw, err := trcobj.JSON(false)
	if err != nil {
		return 0, common.NewBasicError("Unable to convert TRC to JSON", err)
	}
	query := `
	INSERT INTO TRCs (IsdID, Version, Data) VALUES($1, $2, $3)
	ON CONFLICT (IsdID, Version) DO NOTHING`
	res, err := tx.ExecContext(ctx, query, trcobj.ISD, trcobj.Version, raw)
	if err != nil {
		return 0, common.NewBasicError("Failed to insert TRC", err,
			"isd", trcobj.ISD, "version", trcobj.Version)
	}
	return res.RowsAffected()
}

func insertCustKey(ctx context.Context, tx *sql.Tx, ia addr.IA,
	version uint64, key common.RawBytes, oldVersion uint64) (int64, error) {

	query := `
	INSERT INTO CustKeys (IsdID, AsID, Version, Key) VALUES ($1, $2, $3, $4)
	ON CONFLICT (IsdID, AsID) DO UPDATE SET Version = $3, Key = $4 
	WHERE CustKeys.Version = $5`
	r, err := tx.ExecContext(ctx, query, ia.I, ia.A, version, key, oldVersion)
	if err != nil {
		return 0, common.NewBasicError("Failed to insert cust key", err)
	}
	n, err := r.RowsAffected()
	if err != nil {
		return 0, common.NewBasicError("Unable to determine affected rows", err)
	}
	if n == 0 {
		return 0, common.NewBasicError("Cust keys has been modified", nil, "ia", ia,
			"newVersion", version, "oldVersion", oldVersion)
	}
	query = `INSERT INTO CustKeysLog (IsdID, AsID, Version, Key) VALUES ($1, $2, $3, $4)`
	r, err = tx.ExecContext(ctx, query, ia.I, ia.A, version, key)
	if err != nil {
		return 0, common.NewBasicError("Failed to insert CustKeysLog", err)
	}
	return r.RowsAffected()
}
