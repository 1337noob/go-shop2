package event

const PaymentCompleted = "PaymentCompleted"

type PaymentCompletedPayload struct {
	PaymentID         string `json:"payment_id"`
	OrderID           string `json:"order_id"`
	UserID            string `json:"user_id"`
	PaymentSum        int    `json:"payment_sum"`
	PaymentMethodID   string `json:"payment_method_id"`
	PaymentExternalID string `json:"payment_external_id"`
	PaymentType       string `json:"payment_type"`
	PaymentGateway    string `json:"payment_gateway"`
	PaymentStatus     string `json:"payment_status"`
}
