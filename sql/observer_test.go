package sql_test

import (
	"context"
	"testing"

	"github.com/romanqed/gqs/job"
	"github.com/romanqed/gqs/message"
	gsql "github.com/romanqed/gqs/sql"
)

func TestPusherAndObserver(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	pusher := gsql.NewPusher(db)
	observer := gsql.NewObserver(db)

	msg := &message.Message{
		Metadata: map[string]any{"a": 1},
		Payload:  []byte("data"),
	}

	if err := pusher.Push(ctx, msg, 0); err != nil {
		t.Fatal(err)
	}

	j, err := observer.Get(ctx, msg.Id)
	if err != nil {
		t.Fatal(err)
	}
	if j == nil {
		t.Fatal("job not found")
	}
	if j.Status != job.Pending {
		t.Fatalf("expected Pending, got %v", j.Status)
	}
}
