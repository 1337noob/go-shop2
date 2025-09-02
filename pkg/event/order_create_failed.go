package event

const OrderCreateFailed Type = "OrderCreateFailed"

type OrderCreateFailedPayload struct {
	OrderID string `json:"order_id"`
	Error   string `json:"error"`
}
