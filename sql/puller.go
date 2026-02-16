package sql

import (
	"context"
	"github.com/romanqed/gqs"
	"github.com/romanqed/gqs/job"
	"github.com/uptrace/bun"
	"time"
)

// Puller implements gqs.Puller using a SQL backend.
//
// Puller performs atomic state transitions using UPDATE ... RETURNING
// semantics to ensure safe concurrent access across multiple workers.
//
// The implementation assumes:
//
//   - durable writes
//   - transactional guarantees provided by the underlying database
//   - correct indexing of status and scheduling columns
//
// Puller enforces visibility timeout semantics using the locked_until
// column.
type Puller struct {
	db *bun.DB
}

// NewPuller creates a new SQL-backed Puller.
//
// The provided *bun.DB must be properly configured and connected.
// Schema initialization must be completed before using Puller.
func NewPuller(db *bun.DB) *Puller {
	return &Puller{
		db: db,
	}
}

// Pull selects up to batch eligible jobs and transitions them
// to Processing state atomically.
//
// A job is eligible if:
//
//   - next_run_at <= now
//   - status = Pending
//     OR
//   - status = Processing AND locked_until < now
//
// Eligible jobs are transitioned to Processing,
// attempts are incremented,
// locked_until is set to now + lock,
// updated_at is refreshed.
//
// Pull returns the updated job snapshots.
//
// Pull relies on a single UPDATE ... WHERE id IN (subquery)
// statement with RETURNING to avoid race conditions between
// selection and state transition.
func (p *Puller) Pull(ctx context.Context, batch int, lock time.Duration) ([]*job.Job, error) {
	now := time.Now()
	lockUntil := now.Add(lock)
	subQuery := p.db.NewSelect().
		Model((*jobModel)(nil)).
		Column("id").
		Where("next_run_at <= ?", now).
		WhereGroup("AND", func(sq *bun.SelectQuery) *bun.SelectQuery {
			return sq.
				Where("status = ?", job.Pending).
				WhereOr("status = ? AND locked_until < ?", job.Processing, now)
		}).
		Order("next_run_at ASC").
		Limit(batch)
	var jobs []*job.Job
	err := p.db.NewUpdate().
		Model((*jobModel)(nil)).
		Set("status = ?", job.Processing).
		Set("attempts = attempts + 1").
		Set("locked_until = ?", lockUntil).
		Set("updated_at = ?", now).
		Where("id IN (?)", subQuery).
		Returning("*").
		Scan(ctx, &jobs)
	if err != nil {
		return nil, err
	}
	return jobs, nil
}

// ExtendLock extends the visibility timeout of a Processing job.
//
// The job must currently be in Processing state.
// If no rows are affected, ErrLockLost is returned.
//
// ExtendLock updates locked_until and updated_at.
//
// This method does not guarantee exclusive ownership;
// it only ensures the row was still Processing at update time.
func (p *Puller) ExtendLock(ctx context.Context, jb *job.Job, lock time.Duration) error {
	now := time.Now()
	newLock := now.Add(lock)
	res, err := p.db.NewUpdate().
		Model((*jobModel)(nil)).
		Set("locked_until = ?", newLock).
		Set("updated_at = ?", now).
		Where("id = ?", jb.Id).
		Where("status = ?", job.Processing).
		Exec(ctx)
	if err != nil {
		return err
	}
	if !isAffected(res) {
		return gqs.ErrLockLost
	}
	jb.UpdatedAt = now
	jb.LockedUntil = &newLock
	jb.Status = job.Processing
	return nil
}

// Complete transitions a Processing job to Done state.
//
// The job must currently be in Processing state.
// If the update affects no rows, ErrCompleteFailed is returned.
//
// Complete clears locked_until and updates updated_at.
func (p *Puller) Complete(ctx context.Context, jb *job.Job) error {
	now := time.Now()
	res, err := p.db.NewUpdate().
		Model((*jobModel)(nil)).
		Set("status = ?", job.Done).
		Set("locked_until = NULL").
		Set("updated_at = ?", now).
		Where("id = ?", jb.Id).
		Where("status = ?", job.Processing).
		Exec(ctx)
	if err != nil {
		return err
	}
	if !isAffected(res) {
		return gqs.ErrCompleteFailed
	}
	jb.Status = job.Done
	jb.LockedUntil = nil
	jb.UpdatedAt = now
	return nil
}

// Return reschedules a Processing job back to Pending state.
//
// next_run_at is set to now + backoff.
// locked_until is cleared.
// updated_at is refreshed.
//
// If the update affects no rows, ErrJobLost is returned.
//
// Return is typically used after handler failure when
// retry attempts to remain.
func (p *Puller) Return(ctx context.Context, jb *job.Job, backoff time.Duration) error {
	now := time.Now()
	nextRun := now.Add(backoff)
	res, err := p.db.NewUpdate().
		Model((*jobModel)(nil)).
		Set("status = ?", job.Pending).
		Set("next_run_at = ?", nextRun).
		Set("locked_until = NULL").
		Set("updated_at = ?", now).
		Where("id = ?", jb.Id).
		Where("status = ?", job.Processing).
		Exec(ctx)
	if err != nil {
		return err
	}
	if !isAffected(res) {
		return gqs.ErrJobLost
	}
	jb.Status = job.Pending
	jb.NextRunAt = nextRun
	jb.LockedUntil = nil
	jb.UpdatedAt = now
	return nil
}

// Kill transitions a job to Dead state.
//
// The job must be in Pending or Processing state.
// locked_until is cleared.
// updated_at is refreshed.
//
// If the update affects no rows, ErrJobLost is returned.
//
// Kill is typically used when retry limits are exceeded.
func (p *Puller) Kill(ctx context.Context, jb *job.Job) error {
	now := time.Now()
	res, err := p.db.NewUpdate().
		Model((*jobModel)(nil)).
		Set("status = ?", job.Dead).
		Set("locked_until = NULL").
		Set("updated_at = ?", now).
		Where("id = ?", jb.Id).
		Where("status IN (?, ?)", job.Pending, job.Processing).
		Exec(ctx)
	if err != nil {
		return err
	}
	if !isAffected(res) {
		return gqs.ErrJobLost
	}
	jb.Status = job.Dead
	jb.LockedUntil = nil
	jb.UpdatedAt = now
	return nil
}
