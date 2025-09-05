package types

type SagaPayload struct {
	UserID              string `json:"user_id"`
	OrderID             string `json:"order_id"`
	OrderItems          []Item `json:"order_items"`
	PaymentID           string `json:"payment_id"`
	PaymentSum          int    `json:"payment_sum"`
	PaymentMethodID     string `json:"payment_method_id"`
	PaymentExternalID   string `json:"payment_external_id"`
	NotificationID      string `json:"notification_id"`
	NotificationType    string `json:"notification_type"`
	NotificationContent string `json:"notification_content"`
}
