package gqs

import (
	"context"
	"github.com/romanqed/gqs/internal"
	"github.com/romanqed/gqs/job"
	"log/slog"
	"time"
)

// CleanConfig defines the scheduling and filtering parameters
// for a CleanWorker.
//
// Status specifies which job state should be targeted for deletion.
// Only terminal states (such as job.Done or job.Dead) are valid.
//
// Interval defines how often the cleaner runs.
//
// If Before is true, deletion is restricted to jobs whose UpdatedAt
// timestamp is older than now - Delta.
//
// Delta defines the age threshold applied when Before is enabled.
type CleanConfig struct {
	Status   job.Status
	Interval time.Duration
	Before   bool
	Delta    time.Duration
}

// CleanWorker periodically invokes a Cleaner implementation
// according to the provided configuration.
//
// CleanWorker is intended for background retention management,
// such as removing completed or dead jobs after a configurable
// period of time.
//
// CleanWorker does not participate in job processing and does not
// affect visibility timeouts.
//
// CleanWorker has a strict lifecycle:
//   - Start may only be called once.
//   - Stop must be called to terminate the worker.
//   - Stop waits for the internal task to finish or until the timeout
//     expires.
type CleanWorker struct {
	lcBase
	cleaner  Cleaner
	task     internal.TimerTask
	log      *slog.Logger
	status   job.Status
	interval time.Duration
	before   bool
	delta    time.Duration
}

// NewCleanWorker creates a new CleanWorker using the provided
// Cleaner implementation and configuration.
//
// The worker is not started automatically. Call Start to begin
// periodic cleaning.
func NewCleanWorker(cleaner Cleaner, config *CleanConfig, log *slog.Logger) *CleanWorker {
	return &CleanWorker{
		cleaner:  cleaner,
		log:      log,
		status:   config.Status,
		interval: config.Interval,
		before:   config.Before,
		delta:    config.Delta,
	}
}

func (cw *CleanWorker) beforeStamp() *time.Time {
	if !cw.before {
		return nil
	}
	ret := time.Now()
	if cw.delta != 0 {
		ret = ret.Add(-cw.delta)
	}
	return &ret
}

func (cw *CleanWorker) clean(ctx context.Context) {
	before := cw.beforeStamp()
	count, err := cw.cleaner.Clean(ctx, cw.status, before)
	if err != nil {
		cw.log.Error("error while cleaning", "error", err)
	}
	cw.log.Info("cleaned jobs", "count", count)
}

// Start begins periodic execution of the cleaning task.
//
// Start returns ErrDoubleStarted if the worker has already been started.
//
// The provided context controls cancellation of the background task.
func (cw *CleanWorker) Start(ctx context.Context) error {
	if err := cw.tryStart(); err != nil {
		return err
	}
	cw.task.Start(ctx, cw.clean, cw.interval)
	return nil
}

// Stop terminates the background cleaning task.
//
// Stop waits until the task finishes or the specified timeout expires.
// If shutdown does not complete within the timeout, ErrStopTimeout
// is returned.
//
// Stop returns ErrDoubleStopped if the worker is not running.
func (cw *CleanWorker) Stop(timeout time.Duration) error {
	return cw.tryStop(timeout, cw.task.Stop)
}
