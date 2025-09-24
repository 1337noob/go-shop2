package event

import "shop/pkg/types"

const InventoryReserved Type = "InventoryReserved"

type InventoryReservedPayload struct {
	OrderID    string       `json:"order_id"`
	OrderItems []types.Item `json:"order_items"`
}
