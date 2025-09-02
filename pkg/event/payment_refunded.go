package event

const TypePaymentRefunded = "PaymentRefunded"

type PaymentRefundedPayload struct {
	PaymentID string `json:"payment_id"`
	Status    string `json:"status"`
}
