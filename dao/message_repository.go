package dao

import (
	"database/sql"
	"hackathon-backend/model"
)

type MessageRepository struct {
	db *sql.DB
}

func NewMessageRepository(db *sql.DB) *MessageRepository {
	return &MessageRepository{db: db}
}

func (r *MessageRepository) CreateMessage(msg *model.Message) error {
	query := `INSERT INTO messages (id, item_id, sender_id, content, is_ai_response, is_approved, suggested_price, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.Exec(query, msg.ID, msg.ItemID, msg.SenderID, msg.Content, msg.IsAIResponse, msg.IsApproved, msg.SuggestedPrice, msg.CreatedAt)
	return err
}

func (r *MessageRepository) GetMessagesByItemID(itemID string) ([]model.Message, error) {
	query := `SELECT m.id, m.item_id, m.sender_id, m.content, m.is_ai_response, m.is_approved, m.suggested_price, m.created_at, l.ai_reasoning 
              FROM messages m 
              LEFT JOIN negotiation_logs l ON m.item_id = l.item_id AND ABS(TIMESTAMPDIFF(SECOND, m.created_at, l.log_time)) < 2 AND m.is_ai_response = TRUE
              WHERE m.item_id = ? 
              ORDER BY m.created_at ASC`
              
	rows, err := r.db.Query(query, itemID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var msgs []model.Message
	for rows.Next() {
		var msg model.Message
        var reasoning sql.NullString
        var suggestedPrice sql.NullInt64
		if err := rows.Scan(&msg.ID, &msg.ItemID, &msg.SenderID, &msg.Content, &msg.IsAIResponse, &msg.IsApproved, &suggestedPrice, &msg.CreatedAt, &reasoning); err != nil {
			return nil, err
		}
        if reasoning.Valid {
            msg.AIReasoning = reasoning.String
        }
        if suggestedPrice.Valid {
            val := int(suggestedPrice.Int64)
            msg.SuggestedPrice = &val
        }
		msgs = append(msgs, msg)
	}
	return msgs, nil
}

func (r *MessageRepository) GetMessageByID(id string) (*model.Message, error) {
    query := `SELECT id, item_id, sender_id, content, is_ai_response, is_approved, suggested_price, created_at FROM messages WHERE id = ?`
    var m model.Message
    var suggestedPrice sql.NullInt64
    err := r.db.QueryRow(query, id).Scan(&m.ID, &m.ItemID, &m.SenderID, &m.Content, &m.IsAIResponse, &m.IsApproved, &suggestedPrice, &m.CreatedAt)
    if err != nil {
        return nil, err
    }
    if suggestedPrice.Valid {
        val := int(suggestedPrice.Int64)
        m.SuggestedPrice = &val
    }
    return &m, nil
}

func (r *MessageRepository) ApproveMessage(messageID string) error {
    query := `UPDATE messages SET is_approved = TRUE WHERE id = ?`
    _, err := r.db.Exec(query, messageID)
    return err
}

func (r *MessageRepository) DeleteMessage(id string) error {
	_, err := r.db.Exec("DELETE FROM messages WHERE id = ?", id)
	return err
}

func (r *MessageRepository) CreateNegotiationLog(log *model.NegotiationLog) error {
	query := `INSERT INTO negotiation_logs (id, item_id, user_id, proposed_price, ai_decision, counter_price, ai_reasoning, log_time) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.Exec(query, log.ID, log.ItemID, log.UserID, log.ProposedPrice, log.AIDecision, log.CounterPrice, log.AIReasoning, log.LogTime)
	return err
}
