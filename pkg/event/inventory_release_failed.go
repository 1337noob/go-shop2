package event

import "shop/pkg/types"

const InventoryReleaseFailed Type = "InventoryReleaseFailed"

type InventoryReleaseFailedPayload struct {
	OrderID    string       `json:"order_id"`
	OrderItems []types.Item `json:"order_items"`
	Error      string       `json:"error"`
}
