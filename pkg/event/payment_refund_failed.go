package event

const PaymentRefundFailed = "PaymentRefundFailed"

type PaymentRefundFailedPayload struct {
	PaymentID string `json:"payment_id"`
	Error     string `json:"error"`
}
