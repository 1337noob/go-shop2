package outbox

import (
	"context"
	"time"
)

const (
	StatusInit    MessageStatus = "init"
	StatusPending MessageStatus = "pending"
	StatusSent    MessageStatus = "sent"
	StatusError   MessageStatus = "error"
)

type MessageStatus string

type Message struct {
	ID        string        `json:"id"`
	Topic     string        `json:"topic"`
	Key       string        `json:"key"`
	Payload   any           `json:"payload"`
	Status    MessageStatus `json:"status"`
	CreatedAt time.Time     `json:"created_at"`
}

type Outbox interface {
	Publish(ctx context.Context, message Message) error
	GetNotSent(ctx context.Context, limit int) ([]Message, error)
	BatchMarkAsPending(ctx context.Context, ids []string) error
	BatchMarkAsSent(ctx context.Context, ids []string) error
	BatchMarkAsError(ctx context.Context, ids []string) error
}
