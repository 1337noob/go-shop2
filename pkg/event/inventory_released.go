package event

import "shop/pkg/types"

const InventoryReleased Type = "InventoryReleased"

type InventoryReleasedPayload struct {
	OrderItems []types.Item `json:"order_items"`
}
