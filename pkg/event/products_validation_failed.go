package event

const TypeProductsValidationFailed Type = "ProductsValidationFailed"

type ProductsValidationFailedPayload struct {
	Products []Product
	Error    string `json:"error"`
}
