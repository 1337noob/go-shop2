package command

const ReserveInventory Type = "ReserveInventory"

type InventoryItem struct {
	ProductId string `json:"product_id"`
	Quantity  int    `json:"quantity"`
}

type ReserveProductsPayload struct {
	Items []InventoryItem `json:"items"`
}
