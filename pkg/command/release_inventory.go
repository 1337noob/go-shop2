package command

import "shop/pkg/types"

const ReleaseInventory Type = "ReleaseInventory"

type ReleaseProductsPayload struct {
	OrderItems []types.Item `json:"order_items"`
}
