package event

import "shop/pkg/types"

const InventoryReleased Type = "InventoryReleased"

type InventoryReleasedPayload struct {
	OrderID    string       `json:"order_id"`
	OrderItems []types.Item `json:"order_items"`
}
