package event

import "shop/pkg/types"

const ProductsValidated Type = "ProductsValidated"

type ProductsValidatedPayload struct {
	OrderID    string       `json:"order_id"`
	OrderItems []types.Item `json:"order_items"`
}
