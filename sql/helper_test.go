package sql_test

import (
	"context"
	"database/sql"
	"testing"

	gsql "github.com/romanqed/gqs/sql"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"

	_ "modernc.org/sqlite"
)

func newTestDB(t *testing.T) *bun.DB {
	t.Helper()
	sqlDB, err := sql.Open("sqlite", "file::memory:?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)")
	if err != nil {
		t.Fatal(err)
	}
	sqlDB.SetMaxOpenConns(1) // important for sqlite
	db := bun.NewDB(sqlDB, sqlitedialect.New())
	ctx := context.Background()
	if err := gsql.InitDB(ctx, db); err != nil {
		t.Fatal(err)
	}
	return db
}
