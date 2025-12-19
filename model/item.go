package model

import "time"

type Item struct {
	ID                   string    `json:"id"`
	Name                 string    `json:"name"`
	Price                int       `json:"price"`
	Description          string    `json:"description"`
	UserID               string    `json:"user_id"`
	BuyerID              *string   `json:"buyer_id,omitempty"` // Nullable
	Status               string `json:"status"` // on_sale, sold
	ViewsCount           int       `json:"views_count"`
	AINegotiationEnabled bool      `json:"ai_negotiation_enabled"`
	MinPrice             *int   `json:"min_price"`
    ImageURL             string `json:"image_url"`
    InitialPrice         int    `json:"initial_price"`
	CreatedAt            time.Time `json:"created_at"`
}
