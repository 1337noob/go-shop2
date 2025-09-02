package event

const TypeProductCreated Type = "ProductCreated"

type ProductCreatedPayload struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Price      int    `json:"price"`
	CategoryID string `json:"category_id"`
}
