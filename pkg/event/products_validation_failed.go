package event

import "shop/pkg/types"

const ProductsValidationFailed Type = "ProductsValidationFailed"

type ProductsValidationFailedPayload struct {
	OrderItems []types.Item `json:"order_items"`
	Error      string       `json:"error"`
}
