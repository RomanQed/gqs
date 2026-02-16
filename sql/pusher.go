package sql

import (
	"context"
	"github.com/romanqed/gqs/message"
	"github.com/uptrace/bun"
	"time"
)

// Pusher implements gqs.Pusher using a SQL backend.
//
// Pusher inserts new jobs into storage in the Pending state.
// It does not perform any deduplication or idempotency checks.
// The caller is responsible for ensuring that message identifiers
// are unique if required.
type Pusher struct {
	db *bun.DB
}

// NewPusher creates a new SQL-backed Pusher.
//
// The provided *bun.DB must be properly configured and connected.
// Schema initialization must be completed before pushing jobs.
func NewPusher(db *bun.DB) *Pusher {
	return &Pusher{
		db: db,
	}
}

// Push inserts a new message into storage.
//
// The message is scheduled for execution after the specified delay.
// Internally, delay determines the initial NextRunAt timestamp.
//
// Push does not modify the provided message after insertion.
// If insertion fails, no job is created.
//
// Push respects the provided context for cancellation.
func (p *Pusher) Push(ctx context.Context, msg *message.Message, delay time.Duration) error {
	model := fromMessage(msg, delay)
	_, err := p.db.NewInsert().
		Model(model).
		Exec(ctx)
	return err
}
