package event

import "time"

const TypeOrderCreated Type = "OrderCreated"

type OrderCreatedItem struct {
	ProductId string `json:"product_id"`
	Quantity  int    `json:"quantity"`
}

type OrderCreatedPayload struct {
	ID              string             `json:"id"`
	UserID          string             `json:"user_id"`
	PaymentMethodID string             `json:"payment_method_id"`
	Phone           string             `json:"phone"`
	Email           string             `json:"email"`
	Status          string             `json:"status"`
	Items           []OrderCreatedItem `json:"items"`
	CreatedAt       time.Time          `json:"created_at"`
	UpdatedAt       time.Time          `json:"updated_at"`
}
