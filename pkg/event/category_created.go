package event

const TypeCategoryCreated Type = "CategoryCreated"

type CategoryCreatedPayload struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
