package event

import "encoding/json"

type Type string

type Event struct {
	ID      string          `json:"event_id"`
	Type    Type            `json:"event_type"`
	SagaID  string          `json:"saga_id"`
	Payload json.RawMessage `json:"payload"`
}
