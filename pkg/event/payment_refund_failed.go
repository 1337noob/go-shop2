package event

const PaymentRefundFailed = "PaymentRefundFailed"

type PaymentRefundFailedPayload struct {
	OrderID   string `json:"order_id"`
	PaymentID string `json:"payment_id"`
	Error     string `json:"error"`
}
