package command

import "encoding/json"

type Type string

type Command struct {
	ID      string          `json:"command_id"`
	Type    Type            `json:"command_type"`
	SagaID  string          `json:"saga_id,omitempty"`
	Payload json.RawMessage `json:"payload"`
}
