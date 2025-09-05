package model

import (
	"shop/pkg/types"
	"time"
)

type Status string

const (
	StatusInit         Status = "init"
	StatusRunning      Status = "running"
	StatusCompensating Status = "compensating"
	StatusCompleted    Status = "completed"
	StatusCompensated  Status = "compensated"
)

type Saga struct {
	ID           string            `json:"id"`
	CurrentStep  int               `json:"current_step"`
	Status       Status            `json:"status"`
	Steps        []Step            `json:"steps"`
	Payload      types.SagaPayload `json:"payload"`
	Compensating bool              `json:"compensating"`
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
}

//type SagaPayload struct {
//	UserID              string      `json:"user_id"`
//	OrderID             string      `json:"order_id"`
//	OrderItems          []SagaOrderItem `json:"order_items"`
//	PaymentID           string      `json:"payment_id"`
//	PaymentSum          int         `json:"payment_sum"`
//	PaymentMethodID     string      `json:"payment_method_id"`
//	NotificationID      string      `json:"notification_id"`
//	NotificationType    string      `json:"notification_type"`
//	NotificationContent string      `json:"notification_content"`
//}

//type SagaOrderItem struct {
//	ProductID string `json:"product_id"`
//	Name      string `json:"name"`
//	Quantity  int    `json:"quantity"`
//	Price     int    `json:"price"`
//}
