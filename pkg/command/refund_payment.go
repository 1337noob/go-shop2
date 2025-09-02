package command

const RefundPayment Type = "RefundPayment"

type RefundPaymentPayload struct {
	PaymentID string `json:"payment_id"`
}
