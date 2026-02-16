package gqs_test

import (
	"context"
	"database/sql"
	"errors"

	"log/slog"
	"sync/atomic"
	"testing"
	"time"

	"github.com/romanqed/gqs"
	"github.com/romanqed/gqs/job"
	"github.com/romanqed/gqs/message"
	gsql "github.com/romanqed/gqs/sql"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"

	_ "modernc.org/sqlite"
)

func newTestDB(t *testing.T) *bun.DB {
	t.Helper()
	sqlDB, err := sql.Open("sqlite", "file::memory:?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)")
	if err != nil {
		t.Fatal(err)
	}
	sqlDB.SetMaxOpenConns(1) // important for sqlite
	db := bun.NewDB(sqlDB, sqlitedialect.New())
	ctx := context.Background()
	if err := gsql.InitDB(ctx, db); err != nil {
		t.Fatal(err)
	}
	return db
}

func TestWorkerProcessesJob(t *testing.T) {
	db := newTestDB(t)

	pusher := gsql.NewPusher(db)
	puller := gsql.NewPuller(db)
	observer := gsql.NewObserver(db)

	logger := slog.Default()

	handlerCalled := make(chan struct{}, 1)

	handler := func(ctx context.Context, msg *message.Message) error {
		handlerCalled <- struct{}{}
		return nil
	}

	cfg := &gqs.WorkerConfig{
		Concurrency:  1,
		Queue:        10,
		BatchSize:    1,
		PullInterval: 20 * time.Millisecond,
		LockTimeout:  200 * time.Millisecond,
	}

	worker := gqs.NewWorker(puller, handler, cfg, logger)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := worker.Start(ctx); err != nil {
		t.Fatal(err)
	}

	msg := message.NewMessage()
	if err := pusher.Push(ctx, msg, 0); err != nil {
		t.Fatal(err)
	}

	select {
	case <-handlerCalled:
	case <-time.After(time.Second):
		t.Fatal("handler not called")
	}

	time.Sleep(100 * time.Millisecond)

	j, err := observer.Get(ctx, msg.Id)
	if err != nil {
		t.Fatal(err)
	}
	if j.Status != job.Done {
		t.Fatalf("expected Done, got %v", j.Status)
	}

	if err := worker.Stop(time.Second); err != nil {
		t.Fatal(err)
	}
}

func TestWorkerRetry(t *testing.T) {
	db := newTestDB(t)

	pusher := gsql.NewPusher(db)
	puller := gsql.NewPuller(db)
	observer := gsql.NewObserver(db)

	logger := slog.Default()

	var calls atomic.Int32

	handler := func(ctx context.Context, msg *message.Message) error {
		if calls.Add(1) < 2 {
			return errors.New("fail once")
		}
		return nil
	}

	cfg := &gqs.WorkerConfig{
		Concurrency:  1,
		Queue:        10,
		BatchSize:    1,
		PullInterval: 20 * time.Millisecond,
		LockTimeout:  200 * time.Millisecond,
		Backoff: gqs.BackoffConfig{
			MaxRetries:      3,
			InitialInterval: 10 * time.Millisecond,
			MaxInterval:     100 * time.Millisecond,
			Multiplier:      1,
		},
	}

	worker := gqs.NewWorker(puller, handler, cfg, logger)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_ = worker.Start(ctx)

	msg := message.NewMessage()
	_ = pusher.Push(ctx, msg, 0)

	time.Sleep(300 * time.Millisecond)

	j, _ := observer.Get(ctx, msg.Id)
	if j.Status != job.Done {
		t.Fatalf("expected Done after retry, got %v", j.Status)
	}

	_ = worker.Stop(time.Second)
}

func TestWorkerKillShortcut(t *testing.T) {
	db := newTestDB(t)

	pusher := gsql.NewPusher(db)
	puller := gsql.NewPuller(db)
	observer := gsql.NewObserver(db)

	logger := slog.Default()

	handler := func(ctx context.Context, msg *message.Message) error {
		return gqs.ErrKill
	}

	cfg := &gqs.WorkerConfig{
		Concurrency:  1,
		Queue:        10,
		BatchSize:    1,
		PullInterval: 20 * time.Millisecond,
		LockTimeout:  200 * time.Millisecond,
	}

	worker := gqs.NewWorker(puller, handler, cfg, logger)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_ = worker.Start(ctx)

	msg := message.NewMessage()
	_ = pusher.Push(ctx, msg, 0)

	time.Sleep(200 * time.Millisecond)

	j, _ := observer.Get(ctx, msg.Id)
	if j.Status != job.Dead {
		t.Fatalf("expected Dead, got %v", j.Status)
	}

	_ = worker.Stop(time.Second)
}
