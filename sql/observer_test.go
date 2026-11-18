package sql_test

import (
	"context"
	"testing"

	gsql "github.com/romanqed/gqs/sql/v2"
	"github.com/romanqed/gqs/v2"
	"github.com/romanqed/gqs/v2/job"
	"github.com/romanqed/gqs/v2/message"
)

func TestPusherAndObserver(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	var pusher gqs.Pusher = gsql.NewPusher(db)
	var observer gqs.Observer = gsql.NewObserver(db)

	msg := &message.Message{
		Data:    map[string]any{"a": 1},
		Payload: []byte("data"),
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
