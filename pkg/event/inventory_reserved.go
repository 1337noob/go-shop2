package event

import "shop/pkg/types"

const InventoryReserved Type = "InventoryReserved"

type InventoryReservedPayload struct {
	OrderItems []types.Item `json:"order_items"`
}
