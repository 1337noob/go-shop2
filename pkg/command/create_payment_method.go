package command

const CreatePaymentMethod Type = "CreatePaymentMethod"

type CreatePaymentMethodPayload struct {
	UserID      string `json:"user_id"`
	Gateway     string `json:"gateway"`
	PaymentType string `json:"payment_type"`
	Token       string `json:"token"`
}
