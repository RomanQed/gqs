package job

import (
	"github.com/romanqed/gqs/message"
	"time"
)

// Job represents a message managed by the queue storage.
//
// It embeds message.Message and augments it with delivery state and
// scheduling information.
//
// CreatedAt records when the job was initially enqueued.
// UpdatedAt records the last state transition or modification.
//
// Status represents the current state in the job lifecycle.
// Attempts counts how many times the job has been pulled for execution.
// LockedUntil defines the visibility timeout; while set and in the future,
// the job is considered owned by a worker.
// NextRunAt specifies the earliest time the job may be pulled.
//
// Job instances should be treated as snapshots of storage state.
// Mutating fields directly does not change the underlying queue state;
// transitions must be performed through the Puller interface.
type Job struct {
	message.Message

	CreatedAt time.Time
	UpdatedAt time.Time

	Status      Status
	Attempts    uint32
	LockedUntil *time.Time
	NextRunAt   time.Time
}
