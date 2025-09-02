package event

const TypeInventoryReleased Type = "InventoryReleased"

type InventoryReleasedPayload struct {
	Items []InventoryItem `json:"items"`
}
