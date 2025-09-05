package command

import "shop/pkg/types"

const ReserveInventory Type = "ReserveInventory"

type ReserveInventoryPayload struct {
	OrderItems []types.Item `json:"order_items"`
}
