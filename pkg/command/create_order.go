package command

const CreateOrder Type = "CreateOrder"

type OrderItem struct {
	ProductId string `json:"product_id"`
	Quantity  int    `json:"quantity"`
}

type CreateOrderPayload struct {
	UserID          string      `json:"user_id"`
	PaymentMethodID string      `json:"payment_method_id"`
	Phone           string      `json:"phone"`
	Email           string      `json:"email"`
	OrderItems      []OrderItem `json:"order_items"`
}
