package model

import "time"

type OrderStatus string

const (
	OrderStatusCreated    OrderStatus = "created"
	OrderStatusProcessing OrderStatus = "processing"
	OrderStatusCompleted  OrderStatus = "completed"
	OrderStatusFailed     OrderStatus = "failed"
	OrderStatusCancelled  OrderStatus = "cancelled"
)

type Order struct {
	ID              string      `json:"id"`
	UserID          string      `json:"user_id"`
	PaymentMethodID string      `json:"payment_method_id"`
	Phone           string      `json:"phone"`
	Email           string      `json:"email"`
	Status          OrderStatus `json:"status"`
	Items           []OrderItem `json:"items"`
	CreatedAt       time.Time   `json:"created_at"`
	UpdatedAt       time.Time   `json:"updated_at"`
}
