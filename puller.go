package gqs

import (
	"context"
	"errors"
	"github.com/romanqed/gqs/job"
	"time"
)

var (
	// ErrJobLost indicates that the referenced job no longer exists in storage
	// or cannot be found in its expected state.
	//
	// This error may occur if the job was concurrently removed or transitioned
	// by another actor.
	ErrJobLost = errors.New("job lost")

	// ErrLockLost indicates that the caller no longer owns the job lock.
	//
	// This typically happens when the visibility timeout expires and the job
	// is pulled by another worker before the current worker completes or
	// extends the lock.
	ErrLockLost = errors.New("lock lost")

	// ErrCompleteFailed indicates that a job could not be completed due to
	// a state mismatch or concurrent modification.
	//
	// Implementations may return this error when Complete is called on a job
	// that is not currently in the Processing state.
	ErrCompleteFailed = errors.New("complete failed")
)

// Puller defines the read-write contract for consuming and managing jobs
// in the queue lifecycle.
//
// Puller provides visibility timeout semantics similar to systems such
// as Amazon SQS:
//
//   - Pull transitions jobs from Pending to Processing.
//   - While Processing, a job is temporarily invisible to other consumers.
//   - LockedUntil defines the visibility timeout (lease).
//   - If a worker crashes or fails to complete the job before the timeout,
//     the job becomes eligible for pulling again.
//
// The queue provides at-least-once delivery semantics. Handlers must be
// idempotent, as a job may be processed more than once.
type Puller interface {

	// Pull selects up to batch jobs that are eligible for execution and
	// transitions them into the Processing state.
	//
	// The lock parameter defines the visibility timeout (lease duration).
	// Implementations must ensure that:
	//
	//   - returned jobs are atomically transitioned to Processing
	//   - Attempts is incremented for each pulled job
	//   - LockedUntil is set to now + lock
	//
	// Only jobs whose NextRunAt is in the past and whose lock (if any)
	// has expired are eligible.
	//
	// The returned jobs represent authoritative storage state.
	//
	// If ctx is canceled, Pull should abort and return an error.
	Pull(ctx context.Context, batch int, lock time.Duration) ([]*job.Job, error)

	// ExtendLock extends the visibility timeout of a job currently in
	// the Processing state.
	//
	// The lock parameter defines the new lease duration starting from
	// the time of the call.
	//
	// If the job is no longer in Processing state or the caller no longer
	// owns the lease, ErrLockLost should be returned.
	//
	// ExtendLock must not succeed if the job is already transitioned
	// to a terminal state.
	ExtendLock(ctx context.Context, job *job.Job, lock time.Duration) error

	// Complete transitions a job from Processing to Done.
	//
	// Complete must only succeed if the job is currently in Processing
	// state and the caller owns the lease.
	//
	// On success, the job becomes terminal and will not be retried.
	//
	// If the job is missing or no longer in Processing state,
	// an implementation may return ErrLockLost or ErrCompleteFailed.
	Complete(ctx context.Context, job *job.Job) error

	// Return transitions a job from Processing back to Pending and
	// schedules it for future execution.
	//
	// The backoff parameter specifies the delay before the job becomes
	// eligible for pulling again.
	//
	// Implementations must:
	//
	//   - set Status to Pending
	//   - clear LockedUntil
	//   - set NextRunAt to now + backoff
	//
	// Return must only succeed if the job is currently in Processing state.
	// If the lease is lost or the job no longer exists, ErrJobLost or
	// ErrLockLost should be returned.
	Return(ctx context.Context, job *job.Job, backoff time.Duration) error

	// Kill transitions a job to the Dead state.
	//
	// A Dead job is considered permanently failed and will not be retried.
	//
	// Implementations may allow Kill to be called on Pending or Processing
	// jobs. If the job does not exist, ErrJobLost should be returned.
	Kill(ctx context.Context, job *job.Job) error
}
