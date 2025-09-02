package command

const ProcessPayment Type = "ProcessPayment"

type ProcessPaymentPayload struct {
	OrderID  string `json:"order_id"`
	UserID   string `json:"user_id"`
	Amount   int    `json:"amount"`
	MethodID string `json:"method_id"`
}
