package command

import "shop/pkg/types"

const ReserveInventory Type = "ReserveInventory"

type ReserveInventoryPayload struct {
	OrderID    string       `json:"order_id"`
	OrderItems []types.Item `json:"order_items"`
}
