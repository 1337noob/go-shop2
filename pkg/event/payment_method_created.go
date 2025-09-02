package event

const TypePaymentMethodCreated Type = "PaymentMethodCreated"

type PaymentMethodCreatedPayload struct {
	ID          string `json:"id"`
	UserID      string `json:"user_id"`
	Gateway     string `json:"gateway"`
	PaymentType string `json:"payment_type"`
	Token       string `json:"token"`
}
