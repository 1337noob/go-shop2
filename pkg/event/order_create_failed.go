package event

const TypeOrderCreateFailed Type = "OrderCreateFailed"

type OrderCreateFailedPayload struct {
	ID    string `json:"id"`
	Error string `json:"error"`
}
