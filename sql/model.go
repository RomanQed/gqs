package sql

import (
	"github.com/romanqed/gqs/job"
	"github.com/romanqed/gqs/message"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type jobModel struct {
	bun.BaseModel `bun:"table:jobs"`
	Id            uuid.UUID `bun:"id,pk,type:uuid"`

	CreatedAt time.Time `bun:"created_at,nullzero,notnull,default:current_timestamp"`
	UpdatedAt time.Time `bun:"updated_at,nullzero,notnull,default:current_timestamp"`

	Status      job.Status `bun:"status,notnull,default:0"`
	Attempts    uint32     `bun:"attempts,notnull,default:0"`
	LockedUntil *time.Time `bun:"locked_until,nullzero,default:null"`
	NextRunAt   time.Time  `bun:"next_run_at,notnull"`

	Metadata map[string]any `bun:"metadata,type:jsonb"`
	Payload  []byte         `bun:"payload,type:blob"`
}

func (jm *jobModel) toJob() *job.Job {
	return &job.Job{
		Message: message.Message{
			Id:       jm.Id,
			Metadata: jm.Metadata,
			Payload:  jm.Payload,
		},
		CreatedAt:   jm.CreatedAt,
		UpdatedAt:   jm.UpdatedAt,
		Status:      jm.Status,
		Attempts:    jm.Attempts,
		LockedUntil: jm.LockedUntil,
		NextRunAt:   jm.NextRunAt,
	}
}

func fromMessage(msg *message.Message, delay time.Duration) *jobModel {
	now := time.Now()
	return &jobModel{
		Id:          msg.Id,
		Metadata:    msg.Metadata,
		Payload:     msg.Payload,
		CreatedAt:   now,
		UpdatedAt:   now,
		Status:      job.Pending,
		LockedUntil: nil,
		NextRunAt:   now.Add(delay),
	}
}
