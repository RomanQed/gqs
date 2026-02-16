package gqs

import (
	"context"
	"errors"
	"github.com/romanqed/gqs/job"
	"time"
)

var (
	// ErrBadStatus indicates that an invalid job status was supplied to Cleaner.
	//
	// Cleaner implementations are expected to restrict deletion to terminal
	// states (for example, Done or Dead). Supplying a non-terminal status
	// such as Pending or Processing should result in ErrBadStatus.
	ErrBadStatus = errors.New("bad job status")
)

// Cleaner provides a mechanism for permanently removing jobs from storage.
//
// Cleaner is intended for administrative and retention-management use.
// It does not participate in normal job processing and must not modify
// non-terminal jobs.
//
// Typical usage includes:
//
//   - removing completed jobs older than a certain time
//   - purging dead jobs after inspection
//
// Clean must only delete jobs in terminal states (such as Done or Dead).
// Implementations must reject attempts to delete Pending or Processing jobs.
type Cleaner interface {

	// Clean deletes jobs matching the given status and time condition.
	//
	// The status parameter specifies which job state to target.
	// If status is job.Unknown (zero value), implementations may interpret
	// this as a request to delete all terminal jobs (for example, Done and Dead).
	//
	// The before parameter restricts deletion to jobs whose UpdatedAt
	// timestamp is less than or equal to the provided time.
	// If before is nil, no time-based filtering is applied.
	//
	// Clean returns the number of deleted jobs.
	//
	// Clean must not delete jobs in non-terminal states. If status refers
	// to a non-terminal state, ErrBadStatus should be returned.
	//
	// Clean does not affect currently Processing jobs and does not interact
	// with visibility timeouts.
	Clean(ctx context.Context, status job.Status, before *time.Time) (int64, error)
}
