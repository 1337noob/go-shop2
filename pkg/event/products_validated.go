package event

const TypeProductsValidated Type = "ProductsValidated"

type Product struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Price int    `json:"price"`
}

type ProductsValidatedPayload struct {
	Products []Product
}
