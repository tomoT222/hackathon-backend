package model

import "time"

type Item struct {
	ID                   string    `json:"id"`
	Name                 string    `json:"name"`
	Price                int       `json:"price"`
	Description          string    `json:"description"`
	UserID               string    `json:"user_id"`
	BuyerID              *string   `json:"buyer_id,omitempty"` // Nullable
	Status               string    `json:"status"`
	ViewsCount           int       `json:"views_count"`
	AINegotiationEnabled bool      `json:"ai_negotiation_enabled"`
	MinPrice             *int      `json:"min_price,omitempty"` // Nullable
	CreatedAt            time.Time `json:"created_at"`
}
