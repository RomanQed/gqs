package gqs

import (
	"context"
	"errors"
	"github.com/romanqed/gqs/job"
	"github.com/romanqed/gqs/message"
	"log/slog"
	"time"

	"github.com/romanqed/gqs/internal"
)

// MessageHandler defines the user-provided function that processes
// a message pulled from the queue.
//
// The provided context is canceled when:
//
//   - the worker is shutting down
//   - the job lease is lost
//
// The handler must be idempotent. gqs provides at-least-once delivery
// semantics, and a message may be executed more than once if a worker
// crashes or fails to complete it before the visibility timeout expires.
//
// If the handler returns nil, the job is marked as Done.
// If the handler returns a non-nil error, the job is either retried
// according to BackoffConfig or transitioned to Dead.
type MessageHandler func(ctx context.Context, msg *message.Message) error

type errChan chan error

// WorkerConfig defines runtime behavior of a Worker.
//
// Concurrency specifies the number of concurrent message handlers.
//
// Queue specifies the internal buffering capacity between pulling
// jobs from storage and dispatching them to handlers.
//
// BatchSize defines the maximum number of jobs fetched in a single Pull.
//
// PullInterval defines how often the worker polls storage for new jobs.
//
// LockTimeout defines the visibility timeout (lease duration) assigned
// to each pulled job.
//
// Backoff defines the retry policy applied when a handler returns an error.
type WorkerConfig struct {
	Concurrency  int
	Queue        int
	BatchSize    int
	PullInterval time.Duration
	LockTimeout  time.Duration
	Backoff      BackoffConfig
}

// Worker coordinates pulling, dispatching, retrying and completing jobs.
//
// Worker implements an at-least-once processing model:
//
//  1. Periodically Pull jobs from storage.
//  2. Transition them to Processing with a visibility timeout.
//  3. Dispatch them to the user-provided MessageHandler.
//  4. Extend the visibility timeout while the handler runs.
//  5. On success, mark the job as Done.
//  6. On failure, reschedule or permanently fail the job
//     according to BackoffConfig.
//
// Worker does not guarantee exactly-once delivery.
// Handlers must be idempotent.
//
// Worker has a strict lifecycle:
//   - Start may only be called once.
//   - Stop gracefully shuts down pull and worker goroutines.
//   - Stop waits until all in-flight handlers finish or the timeout expires.
type Worker struct {
	lcBase
	puller    Puller
	pullTask  internal.TimerTask
	pool      *internal.WorkerPool[*job.Job]
	log       *slog.Logger
	handler   MessageHandler
	batchSize int
	interval  time.Duration
	lock      time.Duration
	halfLock  time.Duration
	backoff   backoffCounter
}

// NewWorker creates a new Worker instance.
//
// The worker is not started automatically. Call Start to begin processing.
//
// The provided Puller implementation defines storage semantics.
// The provided MessageHandler defines user processing logic.
func NewWorker(puller Puller, handler MessageHandler, config *WorkerConfig, log *slog.Logger) *Worker {
	return &Worker{
		puller:    puller,
		pool:      internal.NewWorkerPool[*job.Job](config.Concurrency, config.Queue, log),
		log:       log,
		handler:   handler,
		batchSize: config.BatchSize,
		interval:  config.PullInterval,
		lock:      config.LockTimeout,
		halfLock:  config.LockTimeout / 2,
		backoff:   backoffCounter{config.Backoff},
	}
}

func (w *Worker) pull(ctx context.Context) {
	jobs, err := w.puller.Pull(ctx, w.batchSize, w.lock)
	if err != nil {
		w.log.Error("pull failed", "err", err)
		return
	}
	for _, entry := range jobs {
		if !w.pool.Push(entry) {
			w.log.Debug("job push interrupted via shutdown", "id", entry.Id)
			return // pool closed, stop handle any jobs, LockUntil fix possible pull-hold
		}
	}
}

func do(handler MessageHandler, ctx context.Context, msg *message.Message) errChan {
	ret := make(errChan, 1)
	go func() {
		ret <- handler(ctx, msg)
	}()
	return ret
}

func (w *Worker) handleOrExtend(ctx context.Context, jb *job.Job) error {
	wrapped, cancel := context.WithCancel(ctx)
	defer cancel()
	errCh := do(w.handler, wrapped, &jb.Message)
	timer := time.NewTimer(w.halfLock)
	defer timer.Stop()
	for {
		select {
		case <-timer.C:
			if err := w.puller.ExtendLock(ctx, jb, w.lock); err != nil {
				cancel()
				return err
			}
			timer.Reset(w.halfLock)
		case err := <-errCh:
			return err
		}
	}
}

func (w *Worker) handle(ctx context.Context, jb *job.Job) {
	err := w.handleOrExtend(ctx, jb)
	if err == nil {
		if err := w.puller.Complete(ctx, jb); err != nil {
			w.log.Error("cannot complete job", "id", jb.Id, "err", err)
		}
		return
	}
	if errors.Is(err, ErrLockLost) {
		w.log.Warn("job lock lost", "id", jb.Id, "err", err)
		return
	}
	backoff, ok := w.backoff.next(jb.Attempts)
	if !ok {
		if err := w.puller.Kill(ctx, jb); err != nil {
			w.log.Error("cannot kill job", "id", jb.Id, "err", err)
		}
		return
	}
	if err := w.puller.Return(ctx, jb, backoff); err != nil {
		w.log.Error("cannot return job", "id", jb.Id, "err", err)
	}
}

// Start begins background pulling and processing of jobs.
//
// Start returns ErrDoubleStarted if the worker has already been started.
//
// The provided context controls cancellation of the worker. When ctx
// is canceled, pulling stops and in-flight handlers receive a canceled
// context.
func (w *Worker) Start(ctx context.Context) error {
	if err := w.tryStart(); err != nil {
		return err
	}
	w.pool.Start(ctx, w.handle)
	w.pullTask.Start(ctx, w.pull, w.interval)
	return nil
}

func (w *Worker) doStop() internal.DoneChan {
	first := w.pullTask.Stop()
	second := w.pool.Stop()
	return internal.Combine(first, second)
}

// Stop initiates graceful shutdown of the worker.
//
// Stop performs the following steps:
//
//  1. Stops periodic pulling of new jobs.
//  2. Cancels the internal worker pool.
//  3. Waits for all in-flight handlers to complete.
//
// If shutdown does not complete within the specified timeout,
// ErrStopTimeout is returned. In this case, background goroutines
// may still be terminating.
//
// Stop returns ErrDoubleStopped if the worker is not running.
func (w *Worker) Stop(timeout time.Duration) error {
	return w.tryStop(timeout, w.doStop)
}
