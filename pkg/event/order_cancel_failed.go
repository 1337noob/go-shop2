package event

const OrderCancelFailed Type = "OrderCancelFailed"

type OrderCancelFailedPayload struct {
	OrderID string `json:"order_id"`
	Error   string `json:"error"`
}
