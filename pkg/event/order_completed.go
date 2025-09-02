package event

const OrderCompleted Type = "OrderCompleted"

type OrderCompletedPayload struct {
	OrderID string `json:"order_id"`
	Status  string `json:"status"`
}
