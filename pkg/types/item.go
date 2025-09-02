package types

type Item struct {
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
	Name      string `json:"name"`
	Price     int    `json:"price"`
}
