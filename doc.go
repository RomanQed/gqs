// Package gqs provides a storage-agnostic queue with at-least-once
// delivery semantics and visibility timeout behavior.
//
// # Overview
//
// gqs models a durable message queue with explicit state transitions.
// It separates transport data (message.Message) from delivery state
// (job.Job) and defines a set of interfaces for pushing, pulling,
// observing and cleaning jobs.
//
// The package does not mandate any particular storage backend.
// Implementations may use SQLite, PostgreSQL, or any other durable store.
//
// # Delivery Semantics
//
// gqs provides at-least-once processing guarantees.
//
// A job may be delivered more than once if:
//
//   - a worker crashes before completing it
//   - the visibility timeout expires
//   - the lease is lost due to concurrent processing
//
// Handlers must therefore be idempotent.
//
// Visibility Timeout (Lease Model)
//
// When a job is pulled, it transitions from Pending to Processing and
// receives a visibility timeout (LockedUntil). While the lease is valid,
// the job is not eligible for pulling by other workers.
//
// If the lease expires before completion, the job becomes eligible again.
//
// The Worker automatically extends the lease while a handler is running.
//
// # State Machine
//
// Jobs follow this lifecycle:
//
//	Pending    -> Processing
//	Processing -> Done
//	Processing -> Pending   (via Return)
//	Processing -> Dead
//
// Terminal states (Done, Dead) are not retried unless explicitly requeued.
//
// # Retry Policy
//
// Retry behavior is controlled by BackoffConfig.
//
// When a handler returns an error:
//
//   - If the maximum retry limit is not exceeded,
//     the job is rescheduled with a computed backoff delay.
//   - Otherwise, the job transitions to Dead.
//
// Attempts are incremented each time a job is successfully pulled.
//
// Worker
//
//	coordinates pulling, dispatching, retrying and completing jobs.
//
// It:
//
//   - periodically polls storage for eligible jobs
//   - dispatches them to a configurable worker pool
//   - extends job leases while handlers execute
//   - applies retry/backoff logic on failure
//   - supports graceful shutdown with timeout
//
// Worker does not guarantee exactly-once delivery.
//
// # Interfaces
//
// gqs defines the following primary interfaces:
//
//	Pusher   — enqueue messages
//	Puller   — manage job lifecycle transitions
//	Observer — inspect job state
//	Cleaner  — remove terminal jobs
//
// These interfaces allow storage implementations to be plugged in
// without coupling the queue logic to a specific database.
//
// # Concurrency Model
//
// Worker uses a bounded internal queue and a fixed-size worker pool.
// Pulling and processing are decoupled to smooth load.
//
// Shutdown is graceful: in-flight handlers are allowed to finish,
// subject to a configurable timeout.
//
// # Storage Expectations
//
// Implementations of Puller must ensure atomic state transitions,
// durable persistence and correct visibility timeout handling.
//
// gqs assumes that storage provides reliable write semantics.
// Behavior under concurrent writers depends on the chosen backend.
//
// # Summary
//
// gqs provides a minimal yet structured foundation for building
// durable background processing systems with explicit lifecycle control,
// retry semantics and pluggable storage backends.
package gqs
