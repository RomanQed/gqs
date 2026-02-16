package sql_test

import (
	"context"
	"testing"
	"time"

	"github.com/romanqed/gqs/job"
	"github.com/romanqed/gqs/message"
	gsql "github.com/romanqed/gqs/sql"
)

func TestPullAndComplete(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	pusher := gsql.NewPusher(db)
	puller := gsql.NewPuller(db)

	msg := message.NewMessage()

	if err := pusher.Push(ctx, msg, 0); err != nil {
		t.Fatal(err)
	}

	jobs, err := puller.Pull(ctx, 1, time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if len(jobs) != 1 {
		t.Fatalf("expected 1 job, got %d", len(jobs))
	}

	j := jobs[0]
	if j.Status != job.Processing {
		t.Fatalf("expected Processing, got %v", j.Status)
	}

	if err := puller.Complete(ctx, j); err != nil {
		t.Fatal(err)
	}
	if j.Status != job.Done {
		t.Fatalf("expected Done, got %v", j.Status)
	}
}

func TestPullAndReturn(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	pusher := gsql.NewPusher(db)
	puller := gsql.NewPuller(db)

	msg := message.NewMessage()
	if err := pusher.Push(ctx, msg, 0); err != nil {
		t.Fatal(err)
	}

	jobs, err := puller.Pull(ctx, 1, time.Second)
	if err != nil {
		t.Fatal(err)
	}

	j := jobs[0]

	if err := puller.Return(ctx, j, time.Second); err != nil {
		t.Fatal(err)
	}

	if j.Status != job.Pending {
		t.Fatalf("expected Pending, got %v", j.Status)
	}
}

func TestPullAndKill(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	pusher := gsql.NewPusher(db)
	puller := gsql.NewPuller(db)

	msg := message.NewMessage()
	if err := pusher.Push(ctx, msg, 0); err != nil {
		t.Fatal(err)
	}

	jobs, err := puller.Pull(ctx, 1, time.Second)
	if err != nil {
		t.Fatal(err)
	}

	j := jobs[0]

	if err := puller.Kill(ctx, j); err != nil {
		t.Fatal(err)
	}

	if j.Status != job.Dead {
		t.Fatalf("expected Dead, got %v", j.Status)
	}
}

func TestExtendLock(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	pusher := gsql.NewPusher(db)
	puller := gsql.NewPuller(db)

	msg := message.NewMessage()
	if err := pusher.Push(ctx, msg, 0); err != nil {
		t.Fatal(err)
	}

	jobs, err := puller.Pull(ctx, 1, time.Second)
	if err != nil {
		t.Fatal(err)
	}
	j := jobs[0]

	old := j.LockedUntil
	if err := puller.ExtendLock(ctx, j, time.Second*2); err != nil {
		t.Fatal(err)
	}

	if !j.LockedUntil.After(*old) {
		t.Fatal("lock was not extended")
	}
}

func TestLeaseExpiration(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	pusher := gsql.NewPusher(db)
	puller := gsql.NewPuller(db)

	msg := message.NewMessage()
	_ = pusher.Push(ctx, msg, 0)

	_, _ = puller.Pull(ctx, 1, time.Millisecond*50)

	time.Sleep(time.Millisecond * 80)

	jobs, err := puller.Pull(ctx, 1, time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if len(jobs) != 1 {
		t.Fatal("expected job to be re-acquired after lease expiration")
	}
}
