package sql

import (
	"context"
	"database/sql"

	. "github.com/0xor1/tlbx/pkg/core"
	"github.com/0xor1/tlbx/pkg/isql"
	"github.com/0xor1/tlbx/pkg/web/app"
)

type tlbxKey struct {
	name string
}

func Mware(name string, sql isql.ReplicaSet) func(app.Tlbx) {
	return func(tlbx app.Tlbx) {
		tlbx.Set(tlbxKey{name}, &client{tlbx: tlbx, name: name, sql: sql})
	}
}

func Get(tlbx app.Tlbx, name string) Client {
	return tlbx.Get(tlbxKey{name}).(Client)
}

type Client interface {
	Base() isql.ReplicaSet
	ReadTx(isoLevel ...sql.IsolationLevel) Tx
	WriteTx(isoLevel ...sql.IsolationLevel) Tx
	ClientCore
}

type ClientCore interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	Query(rowsFn func(isql.Rows), query string, args ...interface{}) error
	QueryRow(query string, args ...interface{}) isql.Row
}

type Tx interface {
	ClientCore
	Rollback()
	Commit()
}

type DoTx interface {
	Rollback()
	Do()
	Commit()
}

type DoTxAdder interface {
	Add(tx Tx, do func(Tx))
}

type DoTxs interface {
	DoTx
	DoTxAdder
}

func newDoTx(tx Tx, do func(Tx)) DoTx {
	return &doTx{
		tx: tx,
		do: do,
	}
}

func NewDoTxs() DoTxs {
	return &doTxs{
		dos: make([]DoTx, 0, 10),
	}
}

type doTxs struct {
	dos []DoTx
}

func (t *doTxs) Add(tx Tx, do func(Tx)) {
	t.dos = append(t.dos, newDoTx(tx, do))
}

func (t *doTxs) Rollback() {
	for _, do := range t.dos {
		do.Rollback()
	}
}

func (t *doTxs) Do() {
	for _, do := range t.dos {
		do.Do()
	}
}

func (t *doTxs) Commit() {
	for _, do := range t.dos {
		do.Commit()
	}
}

type doTx struct {
	tx Tx
	do func(Tx)
}

func (t *doTx) Rollback() {
	t.tx.Rollback()
}
func (t *doTx) Do() {
	t.do(t.tx)
}
func (t *doTx) Commit() {
	t.tx.Commit()
}

type NoopDoTx struct{}

func (_ *NoopDoTx) Rollback() {}
func (_ *NoopDoTx) Do()       {}
func (_ *NoopDoTx) Commit()   {}

type tx struct {
	tx        isql.Tx
	tlbx      app.Tlbx
	sqlClient *client
	scrub     func()
	done      bool
}

func (t *tx) Exec(query string, args ...interface{}) (res sql.Result, err error) {
	t.sqlClient.do(func(q string) { res, err = t.tx.ExecContext(t.tlbx.Ctx(), q, args...) }, query)
	return
}

func (t *tx) Query(rowsFn func(isql.Rows), query string, args ...interface{}) (err error) {
	t.sqlClient.do(func(q string) {
		var rows isql.Rows
		rows, err = t.tx.QueryContext(t.tlbx.Ctx(), q, args...)
		if rows != nil {
			defer rows.Close()
			rowsFn(rows)
		}
	}, query)
	return
}

func (t *tx) QueryRow(query string, args ...interface{}) (row isql.Row) {
	t.sqlClient.do(func(q string) { row = t.tx.QueryRowContext(t.tlbx.Ctx(), q, args...) }, query)
	return
}

func (t *tx) Rollback() {
	if !t.done {
		t.sqlClient.do(func(q string) {
			err := t.tx.Rollback()
			if err != nil && err != sql.ErrTxDone {
				PanicOn(err)
			}
			t.done = true
			t.scrub()
		}, "ROLLBACK")
	}
}

func (t *tx) Commit() {
	t.sqlClient.do(func(q string) {
		PanicOn(t.tx.Commit())
		t.done = true
		t.scrub()
	}, "COMMIT")
}

type client struct {
	tlbx    app.Tlbx
	name    string
	sql     isql.ReplicaSet
	readTx  *tx
	writeTx *tx
}

func (c *client) Base() isql.ReplicaSet {
	return c.sql
}

func (c *client) ReadTx(isoLevel ...sql.IsolationLevel) Tx {
	if c.readTx == nil {
		var t isql.Tx
		var err error
		c.do(func(s string) {
			il := sql.LevelDefault
			if len(isoLevel) > 0 {
				il = isoLevel[0]
			}
			t, err = c.sql.RandSlave().BeginTx(
				context.Background(),
				&sql.TxOptions{
					Isolation: il,
					ReadOnly:  true,
				})
		}, "START READ TRANSACTION")
		PanicOn(err)
		c.readTx = &tx{
			tx:        t,
			tlbx:      c.tlbx,
			sqlClient: c,
			scrub: func() {
				c.readTx = nil
			},
		}
	}
	return c.readTx
}

func (c *client) WriteTx(isoLevel ...sql.IsolationLevel) Tx {
	if c.writeTx == nil {
		var t isql.Tx
		var err error
		c.do(func(s string) {
			il := sql.LevelDefault
			if len(isoLevel) > 0 {
				il = isoLevel[0]
			}
			t, err = c.sql.Primary().BeginTx(
				context.Background(),
				&sql.TxOptions{
					Isolation: il,
					ReadOnly:  false,
				})
		}, "START WRITE TRANSACTION")
		PanicOn(err)
		c.writeTx = &tx{
			tx:        t,
			tlbx:      c.tlbx,
			sqlClient: c,
			scrub: func() {
				c.writeTx = nil
			},
		}
	}
	return c.writeTx
}

func (c *client) Exec(query string, args ...interface{}) (res sql.Result, err error) {
	c.do(func(q string) { res, err = c.sql.Primary().ExecContext(c.tlbx.Ctx(), q, args...) }, query)
	return
}

func (c *client) Query(rowsFn func(isql.Rows), query string, args ...interface{}) (err error) {
	c.do(func(q string) {
		var rows isql.Rows
		rows, err = c.sql.RandSlave().QueryContext(c.tlbx.Ctx(), q, args...)
		if rows != nil {
			defer rows.Close()
			rowsFn(rows)
		}
	}, query)
	return
}

func (c *client) QueryRow(query string, args ...interface{}) (row isql.Row) {
	c.do(func(q string) { row = c.sql.RandSlave().QueryRowContext(c.tlbx.Ctx(), q, args...) }, query)
	return
}

func (c *client) do(do func(string), query string) {
	// no query should ever even come close to 1 second in execution time
	start := NowUnixMilli()
	do(`SET STATEMENT max_statement_time=1 FOR ` + query)
	c.tlbx.LogActionStats(&app.ActionStats{
		Milli:  NowUnixMilli() - start,
		Type:   "SQL",
		Name:   c.name,
		Action: query,
	})
}
