package event

const PaymentRefunded = "PaymentRefunded"

type PaymentRefundedPayload struct {
	PaymentID string `json:"payment_id"`
}
