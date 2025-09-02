package event

import "shop/pkg/types"

const InventoryReserveFailed Type = "InventoryReserveFailed"

type InventoryReserveFailedPayload struct {
	OrderItems []types.Item `json:"order_items"`
	Error      string       `json:"error"`
}
