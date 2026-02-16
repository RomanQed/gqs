package sql

import (
	"context"
	"errors"
	"github.com/uptrace/bun"
)

func createTable(ctx context.Context, db bun.IDB) error {
	_, err := db.NewCreateTable().
		Model((*jobModel)(nil)).
		IfNotExists().
		Exec(ctx)
	return err
}

func createRunIndex(ctx context.Context, db bun.IDB) error {
	_, err := db.NewCreateIndex().
		Model((*jobModel)(nil)).
		Index("idx_jobs_status_next").
		Column("status", "next_run_at").
		IfNotExists().
		Exec(ctx)
	return err
}

func createStatusIndex(ctx context.Context, db bun.IDB) error {
	_, err := db.NewCreateIndex().
		Model((*jobModel)(nil)).
		Index("idx_jobs_status_lock").
		Column("status", "locked_until").
		IfNotExists().
		Exec(ctx)
	return err
}

func createUpdatedIndex(ctx context.Context, db bun.IDB) error {
	_, err := db.NewCreateIndex().
		Model((*jobModel)(nil)).
		Index("idx_jobs_status_updated").
		Column("status", "updated_at").
		IfNotExists().
		Exec(ctx)
	return err
}

func initDB(ctx context.Context, db *bun.DB) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	if err := createTable(ctx, tx); err != nil {
		return errors.Join(err, tx.Rollback())
	}
	if err := createRunIndex(ctx, tx); err != nil {
		return errors.Join(err, tx.Rollback())
	}
	if err := createStatusIndex(ctx, tx); err != nil {
		return errors.Join(err, tx.Rollback())
	}
	if err := createUpdatedIndex(ctx, tx); err != nil {
		return errors.Join(err, tx.Rollback())
	}
	return tx.Commit()
}

// InitDB initializes the database schema required by the SQL backend.
//
// It creates the jobs table and required indexes inside a single
// transaction. If any step fails, the transaction is rolled back.
//
// InitDB is idempotent and may be safely called multiple times.
// It does not drop or modify existing tables beyond creating
// missing objects.
//
// The caller is responsible for providing a properly configured *bun.DB.
func InitDB(ctx context.Context, db *bun.DB) error {
	return initDB(ctx, db)
}

// MustInitDB behaves like InitDB but panics if initialization fails.
//
// This helper is intended for application bootstrap code where
// failure to initialize schema is considered unrecoverable.
func MustInitDB(ctx context.Context, db *bun.DB) {
	if err := initDB(ctx, db); err != nil {
		panic(err)
	}
}
