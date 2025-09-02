package command

const CreateProduct Type = "CreateProduct"

type CreateProductPayload struct {
	Name       string `json:"name"`
	Price      int    `json:"price"`
	CategoryID string `json:"category_id"`
}
