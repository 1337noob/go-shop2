package event

const TypeInventoryReleaseFailed Type = "InventoryReleaseFailed"

type InventoryReleaseFailedPayload struct {
	Items []InventoryItem `json:"items"`
	Error string          `json:"error"`
}
