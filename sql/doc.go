// Package sql provides a bun-based SQL storage implementation for gqs.
//
// This package implements gqs interfaces (Pusher, Puller, Observer,
// Cleaner) using a relational database via github.com/uptrace/bun.
//
// # Overview
//
// The SQL backend provides:
//
//   - durable persistence of jobs
//   - atomic state transitions
//   - visibility timeout (lease) semantics
//   - retry-safe Pull using UPDATE ... RETURNING
//
// It is compatible with SQLite, PostgreSQL and other bun-supported
// dialects, subject to their transactional guarantees.
//
// # Concurrency Model
//
// Pull operations are implemented using a single atomic UPDATE statement
// with a subquery to avoid race conditions between selection and
// state transition.
//
// Correct behavior under high concurrency depends on:
//
//   - proper indexing
//   - database isolation guarantees
//   - write contention characteristics of the chosen backend
//
// SQLite users are strongly encouraged to enable WAL mode and
// configure an appropriate busy_timeout.
//
// # Schema
//
// The backend expects a "jobs" table corresponding to jobModel.
// InitDB (or MustInitDB) creates:
//
//   - the jobs table (if not exists)
//   - index (status, next_run_at)
//   - index (status, locked_until)
//   - index (status, updated_at)
//
// These indexes are required for efficient Pull and Clean operations.
//
// InitDB is idempotent and runs inside a transaction.
// It does not perform destructive migrations.
// Schema evolution must be handled externally.
//
// # Database Lifecycle
//
// This package does not manage connection pooling, migrations,
// or database lifecycle.
//
// The caller is responsible for:
//
//   - creating and configuring *bun.DB
//   - connection limits
//   - WAL/busy_timeout configuration (for SQLite)
//   - running InitDB before use
//
// # Limitations
//
// The SQL backend uses status + timestamp fields to implement
// lease semantics. It does not use lease tokens or optimistic
// locking versions.
//
// Exactly-once processing is not guaranteed.
// Delivery semantics remain at-least-once.
//
// # Summary
//
// Package sql provides a pragmatic, storage-backed implementation
// of gqs suitable for embedded (SQLite) and server-grade
// (PostgreSQL) deployments, while keeping queue logic
// storage-agnostic.
package sql
