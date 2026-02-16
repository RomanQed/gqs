package gqs

import (
	"context"
	"github.com/google/uuid"
	"github.com/romanqed/gqs/job"
)

// Observer provides read-only access to jobs stored in the queue.
//
// Observer does not modify job state and does not participate in
// visibility timeout or lifecycle transitions. It is intended for
// diagnostic, monitoring, and administrative use cases.
//
// Methods of Observer return authoritative snapshots of storage state
// at the time of the call. Returned Job values must be treated as
// immutable views; mutating them does not affect the underlying queue.
type Observer interface {

	// Get returns the job identified by id.
	//
	// If no job with the given id exists, Get returns (nil, nil).
	//
	// The returned Job represents the current storage snapshot,
	// including its Status, Attempts, and scheduling metadata.
	//
	// Get must not change job state.
	Get(ctx context.Context, id uuid.UUID) (*job.Job, error)

	// List returns up to limit jobs matching the provided status.
	//
	// If status is job.Unknown (zero value), implementations may interpret
	// it as "no status filter" and return jobs in any state.
	//
	// If limit is zero or negative, implementations may return all matching
	// jobs, subject to storage-specific constraints.
	//
	// The returned slice contains independent snapshots of job state.
	// Modifying the returned Job values does not affect the queue.
	//
	// List is intended for inspection and administrative tools and should
	// not be used as part of the normal consumption workflow.
	List(ctx context.Context, status job.Status, limit int) ([]*job.Job, error)
}
