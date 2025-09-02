package model

type PaymentStatus string

const (
	PaymentStatusPending   PaymentStatus = "PENDING"
	PaymentStatusCompleted PaymentStatus = "COMPLETED"
	PaymentStatusFailed    PaymentStatus = "FAILED"
	PaymentStatusRefunded  PaymentStatus = "REFUNDED"
)

type Payment struct {
	ID         string        `json:"id"`
	OrderID    string        `json:"order_id"`
	UserID     string        `json:"user_id"`
	Amount     int           `json:"amount"`
	ExternalID string        `json:"external_id"`
	Status     PaymentStatus `json:"status"`
	MethodID   string        `json:"method_id"`
}
