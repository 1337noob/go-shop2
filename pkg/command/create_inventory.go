package command

const CreateInventory Type = "CreateInventory"

type CreateInventoryPayload struct {
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
	//ReservedQuantity int    `json:"reserved_quantity"`
}
