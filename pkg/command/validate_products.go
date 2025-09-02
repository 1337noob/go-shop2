package command

import "shop/pkg/types"

const ValidateProducts Type = "ValidateProducts"

type ValidateProductsPayload struct {
	OrderItems []types.SagaOrderItem `json:"order_items"`
}
