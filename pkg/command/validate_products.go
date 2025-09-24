package command

import "shop/pkg/types"

const ValidateProducts Type = "ValidateProducts"

type ValidateProductsPayload struct {
	OrderID    string       `json:"order_id"`
	OrderItems []types.Item `json:"order_items"`
}
