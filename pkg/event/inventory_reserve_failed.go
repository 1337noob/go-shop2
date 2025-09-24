package event

import "shop/pkg/types"

const InventoryReserveFailed Type = "InventoryReserveFailed"

type InventoryReserveFailedPayload struct {
	OrderID    string       `json:"order_id"`
	OrderItems []types.Item `json:"order_items"`
	Error      string       `json:"error"`
}
