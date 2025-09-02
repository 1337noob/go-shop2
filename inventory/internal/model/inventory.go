package model

import "time"

type Inventory struct {
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
	//ReservedQuantity int       `json:"reserved_quantity"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Item struct {
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
}
