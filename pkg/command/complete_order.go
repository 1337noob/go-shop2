package command

const CompleteOrder Type = "CompleteOrder"

type CompleteOrderPayload struct {
	OrderID string `json:"order_id"`
}
