package event

const TypeOrderCancelFailed Type = "OrderCancelFailed"

type OrderCancelFailedPayload struct {
	ID    string `json:"id"`
	Error string `json:"error"`
}
