package sql

import (
	"context"
	"database/sql"
	"errors"
	"github.com/google/uuid"
	"github.com/romanqed/gqs/job"
	"github.com/uptrace/bun"
)

// Observer implements gqs.Observer using a SQL backend.
//
// Observer provides read-only access to job state stored in the database.
// It does not participate in visibility timeout handling or state
// transitions and must not modify job records.
//
// Returned Job values represent authoritative snapshots of storage state
// at the time of the query.
type Observer struct {
	db *bun.DB
}

// NewObserver creates a new SQL-backed Observer.
//
// The provided *bun.DB must be properly configured and connected.
// Schema initialization must be completed before using Observer.
func NewObserver(db *bun.DB) *Observer {
	return &Observer{
		db: db,
	}
}

// Get retrieves a job by its identifier.
//
// If no job with the given id exists, Get returns (nil, nil).
//
// The returned Job is a snapshot of the current database state.
// Modifying the returned value does not affect storage.
//
// Get performs a simple SELECT query and does not apply
// any locking or transactional semantics beyond what the
// underlying database provides.
func (o *Observer) Get(ctx context.Context, id uuid.UUID) (*job.Job, error) {
	var ret jobModel
	err := o.db.NewSelect().
		Model(&ret).
		Where("id = ?", id).
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return ret.toJob(), nil
}

// List returns up to limit jobs filtered by status.
//
// If status is job.Unknown (zero value), no status filter is applied.
//
// If limit is zero or negative, no LIMIT clause is added
// and all matching rows may be returned.
//
// The returned slice contains independent Job snapshots.
// Mutating them does not affect the underlying storage.
//
// List is intended for administrative or diagnostic use and
// should not be used as part of normal job consumption logic.
func (o *Observer) List(ctx context.Context, status job.Status, limit int) ([]*job.Job, error) {
	var ret []*job.Job
	query := o.db.NewSelect().Model((*jobModel)(nil))
	if status != 0 {
		query.Where("status = ?", status)
	}
	if limit > 0 {
		query.Limit(limit)
	}
	if err := query.Scan(ctx, &ret); err != nil {
		return nil, err
	}
	return ret, nil
}
