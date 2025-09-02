package event

const TypePaymentCompleted = "PaymentCompleted"

type PaymentCompletedPayload struct {
	PaymentID  string `json:"payment_id"`
	OrderID    string `json:"order_id"`
	UserID     string `json:"user_id"`
	Amount     int    `json:"amount"`
	MethodID   string `json:"method_id"`
	ExternalID string `json:"external_id"`
	Status     string `json:"status"`
}
