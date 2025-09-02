package event

const PaymentFailed = "PaymentFailed"

type PaymentFailedPayload struct {
	PaymentID       string `json:"payment_id"`
	OrderID         string `json:"order_id"`
	UserID          string `json:"user_id"`
	PaymentSum      int    `json:"payment_sum"`
	PaymentMethodID string `json:"payment_method_id"`
	Error           string `json:"error"`
}
