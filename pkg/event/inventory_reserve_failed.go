package event

const TypeInventoryReserveFailed Type = "InventoryReserveFailed"

type InventoryReserveFailedPayload struct {
	Items []InventoryItem `json:"items"`
	Error string          `json:"error"`
}
