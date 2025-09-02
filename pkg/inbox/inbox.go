package inbox

import (
	"context"
	"encoding/json"
	"time"
)

const (
	StatusInit      MessageStatus = "init"
	StatusPending   MessageStatus = "pending"
	StatusCompleted MessageStatus = "completed"
	StatusError     MessageStatus = "error"
)

type MessageStatus string

type Message struct {
	//ID          string          `json:"id"`
	MessageID   string          `json:"message_id"`
	MessageType string          `json:"message_type"`
	Topic       string          `json:"topic"`
	Key         string          `json:"key"`
	Payload     json.RawMessage `json:"payload"`
	Status      MessageStatus   `json:"status"`
	CreatedAt   time.Time       `json:"created_at"`
}

type Inbox interface {
	Store(ctx context.Context, message Message) error
	Exists(ctx context.Context, messageID string) (bool, error)
	MarkAsPending(ctx context.Context, messageID string) error
	MarkAsCompleted(ctx context.Context, messageID string) error
	MarkAsError(ctx context.Context, messageID string) error
}
