package model

import "time"

type Product struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	Price      int       `json:"price"`
	CategoryID string    `json:"category_id"`
	CreatedAt  time.Time `json:"created_at"`
}
