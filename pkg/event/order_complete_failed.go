package event

const TypeOrderCompleteFailed Type = "OrderCompleteFailed"

type OrderCompleteFailedPayload struct {
	OrderID string `json:"order_id"`
	Status  string `json:"status"`
}
