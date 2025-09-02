package command

const CreateCategory Type = "CreateCategory"

type CreateCategoryPayload struct {
	Name string `json:"name"`
}
