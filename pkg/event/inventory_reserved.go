package event

const TypeInventoryReserved Type = "InventoryReserved"

type InventoryItem struct {
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
}

type InventoryReservedPayload struct {
	Items []InventoryItem `json:"items"`
}
