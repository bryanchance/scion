// Copyright 2019 Anapaya Systems

package trustdbpatroni

import (
	"context"
	"database/sql"

	consulapi "github.com/hashicorp/consul/api"

	"github.com/scionproto/scion/go/lib/addr"
	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/lib/infra/modules/patroni"
	"github.com/scionproto/scion/go/lib/infra/modules/trust/trustdb"
	"github.com/scionproto/scion/go/lib/infra/modules/trust/trustdb/trustdbpg"
	"github.com/scionproto/scion/go/lib/scrypto/cert"
	"github.com/scionproto/scion/go/lib/scrypto/trc"
)

type Backend struct {
	conns *patroni.ConnPool
	retry *patroni.RetryHelper
}

func New(c *consulapi.Client, cfg patroni.Conf) (*Backend, error) {
	pool, err := patroni.NewConnPool(c, cfg)
	if err != nil {
		return nil, err
	}
	return &Backend{
		conns: pool,
		retry: &patroni.RetryHelper{Pool: pool},
	}, nil
}

func (b *Backend) GetIssCertVersion(ctx context.Context, ia addr.IA,
	version uint64) (*cert.Certificate, error) {

	var ret *cert.Certificate
	rErr := b.retry.DoRead(ctx, func(ctx context.Context, db *sql.DB) error {
		var err error
		ret, err = trustdbpg.NewFromDB(db).GetIssCertVersion(ctx, ia, version)
		return err
	})
	return ret, rErr
}

func (b *Backend) GetIssCertMaxVersion(ctx context.Context, ia addr.IA) (*cert.Certificate, error) {
	var ret *cert.Certificate
	rErr := b.retry.DoRead(ctx, func(ctx context.Context, db *sql.DB) error {
		var err error
		ret, err = trustdbpg.NewFromDB(db).GetIssCertMaxVersion(ctx, ia)
		return err
	})
	return ret, rErr
}

func (b *Backend) GetLeafCertVersion(ctx context.Context,
	ia addr.IA, version uint64) (*cert.Certificate, error) {

	var ret *cert.Certificate
	rErr := b.retry.DoRead(ctx, func(ctx context.Context, db *sql.DB) error {
		var err error
		ret, err = trustdbpg.NewFromDB(db).GetLeafCertVersion(ctx, ia, version)
		return err
	})
	return ret, rErr
}

func (b *Backend) GetLeafCertMaxVersion(ctx context.Context,
	ia addr.IA) (*cert.Certificate, error) {

	var ret *cert.Certificate
	rErr := b.retry.DoRead(ctx, func(ctx context.Context, db *sql.DB) error {
		var err error
		ret, err = trustdbpg.NewFromDB(db).GetLeafCertMaxVersion(ctx, ia)
		return err
	})
	return ret, rErr
}

func (b *Backend) GetChainVersion(ctx context.Context,
	ia addr.IA, version uint64) (*cert.Chain, error) {

	var ret *cert.Chain
	rErr := b.retry.DoRead(ctx, func(ctx context.Context, db *sql.DB) error {
		var err error
		ret, err = trustdbpg.NewFromDB(db).GetChainVersion(ctx, ia, version)
		return err
	})
	return ret, rErr
}

func (b *Backend) GetChainMaxVersion(ctx context.Context, ia addr.IA) (*cert.Chain, error) {
	var ret *cert.Chain
	rErr := b.retry.DoRead(ctx, func(ctx context.Context, db *sql.DB) error {
		var err error
		ret, err = trustdbpg.NewFromDB(db).GetChainMaxVersion(ctx, ia)
		return err
	})
	return ret, rErr
}

func (b *Backend) GetAllChains(ctx context.Context) ([]*cert.Chain, error) {
	var ret []*cert.Chain
	rErr := b.retry.DoRead(ctx, func(ctx context.Context, db *sql.DB) error {
		var err error
		ret, err = trustdbpg.NewFromDB(db).GetAllChains(ctx)
		return err
	})
	return ret, rErr
}

func (b *Backend) GetTRCVersion(ctx context.Context,
	isd addr.ISD, version uint64) (*trc.TRC, error) {

	var ret *trc.TRC
	rErr := b.retry.DoRead(ctx, func(ctx context.Context, db *sql.DB) error {
		var err error
		ret, err = trustdbpg.NewFromDB(db).GetTRCVersion(ctx, isd, version)
		return err
	})
	return ret, rErr
}

func (b *Backend) GetTRCMaxVersion(ctx context.Context, isd addr.ISD) (*trc.TRC, error) {
	var ret *trc.TRC
	rErr := b.retry.DoRead(ctx, func(ctx context.Context, db *sql.DB) error {
		var err error
		ret, err = trustdbpg.NewFromDB(db).GetTRCMaxVersion(ctx, isd)
		return err
	})
	return ret, rErr
}

func (b *Backend) GetAllTRCs(ctx context.Context) ([]*trc.TRC, error) {
	var ret []*trc.TRC
	rErr := b.retry.DoRead(ctx, func(ctx context.Context, db *sql.DB) error {
		var err error
		ret, err = trustdbpg.NewFromDB(db).GetAllTRCs(ctx)
		return err
	})
	return ret, rErr
}

func (b *Backend) GetCustKey(ctx context.Context, ia addr.IA) (common.RawBytes, uint64, error) {
	var ret common.RawBytes
	var v uint64
	rErr := b.retry.DoRead(ctx, func(ctx context.Context, db *sql.DB) error {
		var err error
		ret, v, err = trustdbpg.NewFromDB(db).GetCustKey(ctx, ia)
		return err
	})
	return ret, v, rErr
}

func (b *Backend) InsertIssCert(ctx context.Context, crt *cert.Certificate) (int64, error) {
	var ret int64
	rErr := b.retry.DoWrite(ctx, func(ctx context.Context, db *sql.DB) error {
		var err error
		ret, err = trustdbpg.NewFromDB(db).InsertIssCert(ctx, crt)
		return err
	})
	return ret, rErr
}

func (b *Backend) InsertLeafCert(ctx context.Context, crt *cert.Certificate) (int64, error) {
	var ret int64
	rErr := b.retry.DoWrite(ctx, func(ctx context.Context, db *sql.DB) error {
		var err error
		ret, err = trustdbpg.NewFromDB(db).InsertLeafCert(ctx, crt)
		return err
	})
	return ret, rErr
}

func (b *Backend) InsertChain(ctx context.Context, chain *cert.Chain) (int64, error) {
	var ret int64
	rErr := b.retry.DoWrite(ctx, func(ctx context.Context, db *sql.DB) error {
		var err error
		ret, err = trustdbpg.NewFromDB(db).InsertChain(ctx, chain)
		return err
	})
	return ret, rErr
}

func (b *Backend) InsertTRC(ctx context.Context, trcobj *trc.TRC) (int64, error) {
	var ret int64
	rErr := b.retry.DoWrite(ctx, func(ctx context.Context, db *sql.DB) error {
		var err error
		ret, err = trustdbpg.NewFromDB(db).InsertTRC(ctx, trcobj)
		return err
	})
	return ret, rErr
}

func (b *Backend) InsertCustKey(ctx context.Context, ia addr.IA,
	version uint64, key common.RawBytes, oldVersion uint64) error {

	return b.retry.DoWrite(ctx, func(ctx context.Context, db *sql.DB) error {
		err := trustdbpg.NewFromDB(db).InsertCustKey(ctx, ia, version, key, oldVersion)
		return err
	})
}

func (b *Backend) Close() error {
	b.conns.Close()
	return nil
}

func (b *Backend) BeginTransaction(ctx context.Context,
	opts *sql.TxOptions) (trustdb.Transaction, error) {

	var errHist []error
	for i := 0; i < patroni.MaxWriteRetries; i++ {
		wconn := b.conns.WriteConn()
		if wconn != nil {
			tx, err := trustdbpg.NewFromDB(wconn.DB()).BeginTransaction(ctx, opts)
			if err == nil {
				return &transaction{
					conn: wconn,
					tx:   tx,
				}, nil
			}
			if !wconn.ReportErr(err) {
				// No need to retry if it is an error that doesn't impact the pool.
				return nil, err
			}
			errHist = append(errHist, err)
		} else {
			errHist = append(errHist, common.NewBasicError(patroni.ErrNoConnection, nil))
		}
		if err := patroni.SleepOrTimeOut(ctx, patroni.WaitBetweenRetry); err != nil {
			return nil, common.NewBasicError("Err during sleep", err, "errHist", errHist)
		}
	}
	return nil, common.NewBasicError(patroni.ErrBadConnection, nil, "errHist", errHist)
}

type transaction struct {
	conn *patroni.ConnRef
	tx   trustdb.Transaction
}

func (tx *transaction) GetIssCertVersion(ctx context.Context,
	ia addr.IA, version uint64) (*cert.Certificate, error) {

	ret, err := tx.tx.GetIssCertVersion(ctx, ia, version)
	tx.conn.ReportErr(err)
	return ret, err
}

func (tx *transaction) GetIssCertMaxVersion(ctx context.Context,
	ia addr.IA) (*cert.Certificate, error) {

	ret, err := tx.tx.GetIssCertMaxVersion(ctx, ia)
	tx.conn.ReportErr(err)
	return ret, err
}

func (tx *transaction) GetLeafCertVersion(ctx context.Context,
	ia addr.IA, version uint64) (*cert.Certificate, error) {

	ret, err := tx.tx.GetLeafCertVersion(ctx, ia, version)
	tx.conn.ReportErr(err)
	return ret, err
}

func (tx *transaction) GetLeafCertMaxVersion(ctx context.Context,
	ia addr.IA) (*cert.Certificate, error) {

	ret, err := tx.tx.GetLeafCertMaxVersion(ctx, ia)
	tx.conn.ReportErr(err)
	return ret, err
}

func (tx *transaction) GetChainVersion(ctx context.Context,
	ia addr.IA, version uint64) (*cert.Chain, error) {

	ret, err := tx.tx.GetChainVersion(ctx, ia, version)
	tx.conn.ReportErr(err)
	return ret, err
}

func (tx *transaction) GetChainMaxVersion(ctx context.Context, ia addr.IA) (*cert.Chain, error) {
	ret, err := tx.tx.GetChainMaxVersion(ctx, ia)
	tx.conn.ReportErr(err)
	return ret, err
}

func (tx *transaction) GetAllChains(ctx context.Context) ([]*cert.Chain, error) {
	ret, err := tx.tx.GetAllChains(ctx)
	tx.conn.ReportErr(err)
	return ret, err
}

func (tx *transaction) GetTRCVersion(ctx context.Context,
	isd addr.ISD, version uint64) (*trc.TRC, error) {

	ret, err := tx.tx.GetTRCVersion(ctx, isd, version)
	tx.conn.ReportErr(err)
	return ret, err
}

func (tx *transaction) GetTRCMaxVersion(ctx context.Context, isd addr.ISD) (*trc.TRC, error) {
	ret, err := tx.tx.GetTRCMaxVersion(ctx, isd)
	tx.conn.ReportErr(err)
	return ret, err
}

func (tx *transaction) GetAllTRCs(ctx context.Context) ([]*trc.TRC, error) {
	ret, err := tx.tx.GetAllTRCs(ctx)
	tx.conn.ReportErr(err)
	return ret, err
}

func (tx *transaction) GetCustKey(ctx context.Context,
	ia addr.IA) (common.RawBytes, uint64, error) {

	ret, v, err := tx.tx.GetCustKey(ctx, ia)
	tx.conn.ReportErr(err)
	return ret, v, err
}

func (tx *transaction) InsertIssCert(ctx context.Context, crt *cert.Certificate) (int64, error) {
	ret, err := tx.tx.InsertIssCert(ctx, crt)
	tx.conn.ReportErr(err)
	return ret, err
}

func (tx *transaction) InsertLeafCert(ctx context.Context, crt *cert.Certificate) (int64, error) {
	ret, err := tx.tx.InsertLeafCert(ctx, crt)
	tx.conn.ReportErr(err)
	return ret, err
}

func (tx *transaction) InsertChain(ctx context.Context, chain *cert.Chain) (int64, error) {
	ret, err := tx.tx.InsertChain(ctx, chain)
	tx.conn.ReportErr(err)
	return ret, err
}

func (tx *transaction) InsertTRC(ctx context.Context, trcobj *trc.TRC) (int64, error) {
	ret, err := tx.tx.InsertTRC(ctx, trcobj)
	tx.conn.ReportErr(err)
	return ret, err
}

func (tx *transaction) InsertCustKey(ctx context.Context, ia addr.IA,
	version uint64, key common.RawBytes, oldVersion uint64) error {

	err := tx.tx.InsertCustKey(ctx, ia, version, key, oldVersion)
	tx.conn.ReportErr(err)
	return err
}

func (tx *transaction) Commit() error {
	err := tx.tx.Commit()
	tx.conn.ReportErr(err)
	return err
}

func (tx *transaction) Rollback() error {
	err := tx.tx.Rollback()
	tx.conn.ReportErr(err)
	return err
}
