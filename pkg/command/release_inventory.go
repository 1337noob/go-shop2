package command

import "shop/pkg/types"

const ReleaseInventory Type = "ReleaseInventory"

type ReleaseInventoryPayload struct {
	OrderItems []types.Item `json:"order_items"`
}
