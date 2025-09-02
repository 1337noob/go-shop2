package command

import "shop/pkg/types"

const CreateOrder Type = "CreateOrder"

type CreateOrderPayload struct {
	UserID          string       `json:"user_id"`
	PaymentMethodID string       `json:"payment_method_id"`
	Phone           string       `json:"phone"`
	Email           string       `json:"email"`
	OrderItems      []types.Item `json:"order_items"`
}
