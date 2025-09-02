package command

import "shop/pkg/types"

const ReserveInventory Type = "ReserveInventory"

type ReserveProductsPayload struct {
	OrderItems []types.Item `json:"order_items"`
}
