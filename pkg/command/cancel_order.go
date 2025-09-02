package command

const CancelOrder Type = "CancelOrder"

type CancelOrderPayload struct {
	OrderID string `json:"order_id"`
}
