package gqs_test

import (
	"context"
	"log/slog"
	"sync/atomic"
	"testing"
	"time"

	"github.com/romanqed/gqs"
	"github.com/romanqed/gqs/job"
)

type mockCleaner struct {
	count atomic.Int64
}

func (m *mockCleaner) Clean(ctx context.Context, status job.Status, before *time.Time) (int64, error) {
	m.count.Add(1)
	return 1, nil
}

func TestCleanWorkerBasic(t *testing.T) {
	cleaner := &mockCleaner{}
	logger := slog.Default()

	cfg := &gqs.CleanConfig{
		Status:   job.Done,
		Interval: 50 * time.Millisecond,
		Before:   false,
	}

	w := gqs.NewCleanWorker(cleaner, cfg, logger)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := w.Start(ctx); err != nil {
		t.Fatal(err)
	}

	time.Sleep(150 * time.Millisecond)

	if err := w.Stop(time.Second); err != nil {
		t.Fatal(err)
	}

	if cleaner.count.Load() == 0 {
		t.Fatal("expected cleaner to run at least once")
	}
}

func TestCleanWorkerLifecycleErrors(t *testing.T) {
	cleaner := &mockCleaner{}
	logger := slog.Default()

	cfg := &gqs.CleanConfig{
		Status:   job.Done,
		Interval: time.Second,
	}

	w := gqs.NewCleanWorker(cleaner, cfg, logger)

	ctx := context.Background()

	if err := w.Start(ctx); err != nil {
		t.Fatal(err)
	}

	if err := w.Start(ctx); err == nil {
		t.Fatal("expected ErrDoubleStarted")
	}

	if err := w.Stop(time.Second); err != nil {
		t.Fatal(err)
	}

	if err := w.Stop(time.Second); err == nil {
		t.Fatal("expected ErrDoubleStopped")
	}
}
