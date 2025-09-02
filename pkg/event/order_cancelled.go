package event

const TypeOrderCancelled Type = "OrderCancelled"

type OrderCancelledPayload struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}
