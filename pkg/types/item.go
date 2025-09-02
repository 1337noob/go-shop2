package types

type Item struct {
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
	Name      string `json:"name,omitempty"`
	Price     int    `json:"price,omitempty"`
}
