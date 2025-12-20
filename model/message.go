package model

import "time"

type Message struct {
	ID             string    `json:"id"`
	ItemID         string    `json:"item_id"`
	SenderID       string    `json:"sender_id"`
	SenderName     string    `json:"sender_name"`
	Content        string    `json:"content"`
	IsAIResponse   bool      `json:"is_ai_response"`
	IsApproved     bool      `json:"is_approved"`
	AIReasoning    string    `json:"ai_reasoning,omitempty"` // Derived from logs for sellers
	SuggestedPrice *int      `json:"suggested_price,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
}

type NegotiationLog struct {
	ID            string    `json:"id"`
	ItemID        string    `json:"item_id"`
	UserID        string    `json:"user_id"`
	ProposedPrice int       `json:"proposed_price"`
	AIDecision    string    `json:"ai_decision"` // ACCEPT, REJECT, COUNTER, ANSWER
	CounterPrice  int       `json:"counter_price"`
	AIReasoning   string    `json:"ai_reasoning"`
	LogTime       time.Time `json:"log_time"`
}
