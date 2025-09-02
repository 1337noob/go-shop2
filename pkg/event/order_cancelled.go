package event

const OrderCancelled Type = "OrderCancelled"

type OrderCancelledPayload struct {
	OrderID string `json:"order_id"`
}
