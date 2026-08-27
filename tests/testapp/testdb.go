package testapp

import (
	"database/sql"
	"testing"

	"github.com/ashep/go-app/testpostgres"
	"github.com/jackc/pgx/v5/stdlib"
)

type TestDB struct {
	DSN string

	t *testing.T
	d *sql.DB
}

func newDB(t *testing.T) *TestDB {
	t.Helper()

	tp := testpostgres.New(t)

	return &TestDB{
		DSN: tp.DSN(),

		t: t,
		d: stdlib.OpenDBFromPool(tp.DB()),
	}
}
