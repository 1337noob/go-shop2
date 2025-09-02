package command

import "shop/pkg/types"

const ValidateProducts Type = "ValidateProducts"

type ValidateProductsPayload struct {
	OrderItems []types.Item `json:"order_items"`
}
