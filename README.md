# gqs [![gqs](https://img.shields.io/github/v/release/RomanQed/gqs?style=for-the-badge&label=gqs&color=blue)](https://github.com/RomanQed/gqs/releases)
[![Go Reference](https://pkg.go.dev/badge/github.com/romanqed/gqs.svg)](https://pkg.go.dev/github.com/romanqed/gqs)
[![Go Report Card](https://goreportcard.com/badge/github.com/romanqed/gqs)](https://goreportcard.com/report/github.com/romanqed/gqs)
[![License](https://img.shields.io/github/license/RomanQed/gqs?style=for-the-badge)](LICENSE)

Lightweight Go queue service with at-least-once delivery semantics and pluggable storage backends.
Designed to be minimal, explicit, and backend-agnostic — suitable for embedded systems (SQLite) and server-grade deployments.

## Getting Started

To use this library, you will need:

* Go 1.24 or higher
* A supported storage backend (e.g., SQL module)

## Features

* Storage-agnostic queue core (`gqs`)
* At-least-once delivery model
* Visibility timeout (lease) semantics
* Explicit state machine (`Pending → Processing → Done/Dead`)
* Retry with configurable exponential backoff
* Immediate kill shortcut (`ErrKill`)
* Graceful shutdown with timeout
* Worker pool with bounded internal queue
* Pluggable backends (e.g., `gqs/sql`)
* Fully race-safe (`go test -race` clean)

## Installing

### Core module

```bash
go get github.com/romanqed/gqs@v1.0.0
```

### SQL backend

```bash
go get github.com/romanqed/gqs/sql@v1.0.0
```

## Usage Examples

### Basic SQL setup (SQLite)

```go
package main

import (
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/romanqed/gqs"
	gsql "github.com/romanqed/gqs/sql"
	"github.com/romanqed/gqs/message"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	_ "modernc.org/sqlite"
)

func main() {
	ctx := context.Background()

	sqlDB, err := sql.Open("sqlite", "file:test.db?_pragma=journal_mode(WAL)")
	if err != nil {
		log.Fatal(err)
	}
	sqlDB.SetMaxOpenConns(1)

	db := bun.NewDB(sqlDB, sqlitedialect.New())
	defer db.Close()

	if err := gsql.InitDB(ctx, db); err != nil {
		log.Fatal(err)
	}

	pusher := gsql.NewPusher(db)
	puller := gsql.NewPuller(db)

	handler := func(ctx context.Context, msg *message.Message) error {
		log.Printf("processing: %v\n", msg.Metadata)
		return nil
	}

	worker := gqs.NewWorker(puller, handler, &gqs.WorkerConfig{
		Concurrency:  2,
		Queue:        32,
		BatchSize:    8,
		PullInterval: 100 * time.Millisecond,
		LockTimeout:  5 * time.Second,
	}, nil)

	if err := worker.Start(ctx); err != nil {
		log.Fatal(err)
	}

	msg := message.NewMessage()
	msg.Set("hello", "world")

	if err := pusher.Push(ctx, msg, 0); err != nil {
		log.Fatal(err)
	}

	time.Sleep(2 * time.Second)

	_ = worker.Stop(time.Second)
}
```

## Processing Model

### Delivery semantics

gqs provides **at-least-once** delivery.

A job may be executed more than once if:

* A worker crashes before completion
* The visibility timeout expires
* A lease is lost

Handlers must be idempotent.

### State machine

```
Pending → Processing → Done
Pending → Processing → Dead
Processing → Pending (retry)
```

Terminal states:

* `Done`
* `Dead`

### Retry and Backoff

Retry behavior is controlled via `BackoffConfig`.

If a handler returns:

* `nil` → job becomes `Done`
* `ErrKill` → job becomes `Dead` immediately
* other error → job is retried with backoff
* retry limit exceeded → job becomes `Dead`

### Immediate termination

```go
return gqs.ErrKill
```

Use `ErrKill` to permanently mark a job as `Dead` without retry.

## Graceful Shutdown

`Worker.Stop(timeout)`:

* Stops pulling new jobs
* Cancels internal workers
* Waits for in-flight handlers
* Returns `ErrStopTimeout` if not completed in time

## SQLite Notes

For SQLite, it is strongly recommended to:

* Enable WAL mode
* Configure busy_timeout
* Limit open connections (`SetMaxOpenConns(1)`)

## Built With

* [Go](https://go.dev)
* [bun](https://bun.uptrace.dev/) — SQL builder
* [modernc SQLite](https://pkg.go.dev/modernc.org/sqlite) (optional)

## Authors

* **[RomanQed](https://github.com/RomanQed)** — *Main work*

See also the list of [contributors](https://github.com/RomanQed/gqs/contributors)
who participated in this project.

## License

This project is licensed under the Apache License 2.0 — see the [LICENSE](LICENSE) file for details.
