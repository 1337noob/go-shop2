package command

import "shop/pkg/types"

const SagaCreateOrder Type = "SagaCreateOrder"

type SagaCreateOrderPayload struct {
	UserID          string       `json:"user_id"`
	PaymentMethodID string       `json:"payment_method_id"`
	OrderItems      []types.Item `json:"order_items"`
}
