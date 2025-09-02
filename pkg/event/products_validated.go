package event

import "shop/pkg/types"

const ProductsValidated Type = "ProductsValidated"

type ProductsValidatedPayload struct {
	OrderItems []types.Item `json:"order_items"`
}
