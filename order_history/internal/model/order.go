package model

import "time"

const (
	StatusOrderCreated OrderStatus = "order_created"
)

type OrderStatus string

type Order struct {
	ID           string       `json:"id"`
	UserID       string       `json:"user_id"`
	Items        []Item       `json:"items"`
	Payment      Payment      `json:"payment"`
	Notification Notification `json:"notification"`
	Status       OrderStatus  `json:"status"`
	CreatedAt    time.Time    `json:"created_at"`
	UpdatedAt    time.Time    `json:"updated_at"`
}

type Item struct {
	ProductId string `json:"product_id"`
	Name      string `json:"name"`
	Qty       int64  `json:"qty"`
	Price     int64  `json:"price"`
}

type Payment struct {
	Id         string `json:"id"`
	PaymentSum int64  `json:"payment_sum"`
	ExternalId string `json:"external_id"`
	Status     string `json:"status"`
}

type Notification struct {
	Id     string `json:"id"`
	Type   string `json:"type"`
	Status string `json:"status"`
}
