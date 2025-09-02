package command

const ProcessPayment Type = "ProcessPayment"

type ProcessPaymentPayload struct {
	OrderID         string `json:"order_id"`
	UserID          string `json:"user_id"`
	PaymentSum      int    `json:"payment_sum"`
	PaymentMethodID string `json:"payment_method_id"`
}
