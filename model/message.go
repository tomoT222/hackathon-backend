package model

import "time"

type Message struct {
	ID           string    `json:"id"`
	ItemID       string    `json:"item_id"`
	SenderID     string    `json:"sender_id"`
	Content      string    `json:"content"`
	IsAIResponse bool      `json:"is_ai_response"`
    IsApproved   bool      `json:"is_approved"`
	CreatedAt    time.Time `json:"created_at"`
    // Optional: Include Reasoning if joined? 
    AIReasoning  string    `json:"ai_reasoning,omitempty"` 
}

type NegotiationLog struct {
	ID            string    `json:"id"`
	ItemID        string    `json:"item_id"`
	UserID        string    `json:"user_id"`
	ProposedPrice int       `json:"proposed_price"`
	AIDecision    string    `json:"ai_decision"` // ACCEPT, REJECT, COUNTER
	AIReasoning   string    `json:"ai_reasoning"`
	LogTime       time.Time `json:"log_time"`
}
