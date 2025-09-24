package event

const PaymentRefunded = "PaymentRefunded"

type PaymentRefundedPayload struct {
	OrderID   string `json:"order_id"`
	PaymentID string `json:"payment_id"`
}
