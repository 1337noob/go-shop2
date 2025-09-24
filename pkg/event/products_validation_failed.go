package event

import "shop/pkg/types"

const ProductsValidationFailed Type = "ProductsValidationFailed"

type ProductsValidationFailedPayload struct {
	OrderID    string       `json:"order_id"`
	OrderItems []types.Item `json:"order_items"`
	Error      string       `json:"error"`
}
