package event

import "shop/pkg/types"

const InventoryReleaseFailed Type = "InventoryReleaseFailed"

type InventoryReleaseFailedPayload struct {
	OrderItems []types.Item `json:"order_items"`
	Error      string       `json:"error"`
}
