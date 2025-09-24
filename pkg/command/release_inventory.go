package command

import "shop/pkg/types"

const ReleaseInventory Type = "ReleaseInventory"

type ReleaseInventoryPayload struct {
	OrderID    string       `json:"order_id"`
	OrderItems []types.Item `json:"order_items"`
}
