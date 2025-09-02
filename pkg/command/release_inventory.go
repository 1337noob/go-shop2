package command

const ReleaseInventory Type = "ReleaseInventory"

type ReleaseProductsPayload struct {
	Items []InventoryItem `json:"items"`
}
