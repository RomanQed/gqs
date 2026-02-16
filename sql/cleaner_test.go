package sql_test

import (
	"context"
	"testing"
	"time"

	"github.com/romanqed/gqs/job"
	"github.com/romanqed/gqs/message"
	gsql "github.com/romanqed/gqs/sql"
)

func TestCleaner(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	pusher := gsql.NewPusher(db)
	puller := gsql.NewPuller(db)
	cleaner := gsql.NewCleaner(db)

	msg := message.NewMessage()
	if err := pusher.Push(ctx, msg, 0); err != nil {
		t.Fatal(err)
	}

	jobs, _ := puller.Pull(ctx, 1, time.Second)
	j := jobs[0]
	_ = puller.Complete(ctx, j)

	count, err := cleaner.Clean(ctx, job.Done, nil)
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("expected 1 deleted job, got %d", count)
	}
}
