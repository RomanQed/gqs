package sql

import (
	"context"
	"github.com/romanqed/gqs"
	"github.com/romanqed/gqs/job"
	"github.com/uptrace/bun"
	"time"
)

// Cleaner implements gqs.Cleaner using a SQL backend.
//
// Cleaner permanently removes terminal jobs from storage.
// It is intended for retention management and administrative cleanup.
//
// This implementation deletes rows directly from the jobs table
// and does not participate in visibility timeout or processing logic.
type Cleaner struct {
	db *bun.DB
}

// NewCleaner creates a new SQL-backed Cleaner.
//
// The provided *bun.DB must be properly configured and connected.
// Schema initialization must be completed before using Cleaner.
func NewCleaner(db *bun.DB) *Cleaner {
	return &Cleaner{
		db: db,
	}
}

// Clean deletes jobs matching the provided status and time filter.
//
// Only terminal states are allowed:
//
//   - job.Done
//   - job.Dead
//
// If status is job.Unknown (zero value), both Done and Dead jobs
// are eligible for deletion.
//
// If status refers to a non-terminal state (such as Pending or Processing),
// ErrBadStatus is returned.
//
// If before is non-nil, only jobs with updated_at <= *before
// are deleted. If before is nil, no time-based filtering is applied.
//
// Clean returns the number of deleted rows.
//
// Clean does not attempt to lock or coordinate with running workers.
// Deleting Processing jobs is explicitly disallowed by status checks.
func (c *Cleaner) Clean(ctx context.Context, status job.Status, before *time.Time) (int64, error) {
	if status != 0 && status != job.Dead && status != job.Done {
		return 0, gqs.ErrBadStatus
	}
	query := c.db.NewDelete().Model((*jobModel)(nil))
	if status != 0 {
		query.Where("status = ?", status)
	} else {
		query.Where("status IN (?, ?)", job.Done, job.Dead)
	}
	if before != nil {
		query.Where("updated_at <= ?", before)
	}
	res, err := query.Exec(ctx)
	if err != nil {
		return 0, err
	}
	return getAffected(res), nil
}
