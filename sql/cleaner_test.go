package sql_test

import (
	"context"
	"testing"
	"time"

	gsql "github.com/romanqed/gqs/sql/v2"
	"github.com/romanqed/gqs/v2"
	"github.com/romanqed/gqs/v2/job"
	"github.com/romanqed/gqs/v2/message"
)

func TestCleaner(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	var pusher gqs.Pusher = gsql.NewPusher(db)
	var puller gqs.Puller = gsql.NewPuller(db)
	var cleaner gqs.Cleaner = gsql.NewCleaner(db)

	msg := message.New()
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
