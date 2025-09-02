package event

const OrderCompleteFailed Type = "OrderCompleteFailed"

type OrderCompleteFailedPayload struct {
	OrderID string `json:"order_id"`
	Status  string `json:"status"`
}
