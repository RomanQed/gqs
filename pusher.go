package gqs

import (
	"context"
	"github.com/romanqed/gqs/message"
	"time"
)

// Pusher defines the write-side entry point of a queue.
type Pusher interface {

	// Push enqueues a new message for future processing.
	//
	// The provided context controls cancellation of the enqueue operation
	// itself. It does not affect the lifetime of the enqueued message.
	//
	// The delay parameter specifies the minimum duration that must elapse
	// before the message becomes eligible for pulling. A zero delay makes
	// the message immediately available. A positive delay schedules the
	// message for deferred execution.
	//
	// Implementations are expected to:
	//
	//   - persist the message durably before returning nil
	//   - initialize internal scheduling metadata (for example, NextRunAt)
	//   - assign creation timestamps if applicable
	//
	// Push must not mutate msg after returning.
	//
	// If Push returns a non-nil error, the message must not be considered
	// enqueued.
	//
	// Implementations may return context-related errors if ctx is canceled
	// or times out.
	Push(ctx context.Context, msg *message.Message, delay time.Duration) error
}
