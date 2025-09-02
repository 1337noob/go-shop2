package event

import "time"

const TypeInventoryCreated Type = "InventoryCreated"

type InventoryCreatedPayload struct {
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
	//ReservedQuantity int       `json:"reserved_quantity"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
