package model

import (
	"shop/pkg/types"
	"time"
)

type OrderStatus string

const (
	StatusOrderCreated      OrderStatus = "order_created"
	StatusProductsValidated OrderStatus = "products_validated"
	StatusInventoryReserved OrderStatus = "inventory_reserved"
	StatusPaymentCompleted  OrderStatus = "payment_completed"
	StatusOrderCompleted    OrderStatus = "order_completed"
	StatusOrderCancelled    OrderStatus = "order_cancelled"
)

type Order struct {
	ID                string       `json:"id"`
	UserID            string       `json:"user_id"`
	OrderItems        []types.Item `json:"order_items"`
	PaymentID         string       `json:"payment_id"`
	PaymentMethodID   string       `json:"payment_method_id"`
	PaymentType       string       `json:"payment_type"`
	PaymentGateway    string       `json:"payment_gateway"`
	PaymentSum        int          `json:"payment_sum"`
	PaymentExternalID string       `json:"payment_external_id"`
	PaymentStatus     string       `json:"payment_status"`
	Status            OrderStatus  `json:"status"`
	CreatedAt         time.Time    `json:"created_at"`
	UpdatedAt         time.Time    `json:"updated_at"`
}
