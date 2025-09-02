package event

import (
	"shop/pkg/types"
	"time"
)

const OrderCreated Type = "OrderCreated"

type OrderCreatedPayload struct {
	OrderID         string       `json:"order_id"`
	UserID          string       `json:"user_id"`
	PaymentMethodID string       `json:"payment_method_id"`
	Phone           string       `json:"phone"`
	Email           string       `json:"email"`
	Status          string       `json:"status"`
	OrderItems      []types.Item `json:"order_items"`
	CreatedAt       time.Time    `json:"created_at"`
	UpdatedAt       time.Time    `json:"updated_at"`
}
