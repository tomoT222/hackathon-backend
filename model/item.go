package model

type Item struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Price       int     `json:"price"`
	Description string  `json:"description"`
	UserID      string  `json:"user_id"`
	BuyerID     *string `json:"buyer_id,omitempty"` // Nullable
	Status      string  `json:"status"`
}
