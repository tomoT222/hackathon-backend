package dao

import (
	"database/sql"
	"hackathon-backend/model"
)

type ItemRepository struct {
	db *sql.DB
}

func NewItemRepository(db *sql.DB) *ItemRepository {
	return &ItemRepository{db: db}
}

func (r *ItemRepository) GetAll() ([]model.Item, error) {
	// 修正: buyer_idとstatusも取得するように変更
	// schema.sql: id, name, price, description, user_id, buyer_id, status
	query := `
		SELECT id, name, price, description, user_id, buyer_id, status 
		FROM items 
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []model.Item
	for rows.Next() {
		var item model.Item
		var buyerID sql.NullString
		
		if err := rows.Scan(&item.ID, &item.Name, &item.Price, &item.Description, &item.UserID, &buyerID, &item.Status); err != nil {
			return nil, err
		}
		
		if buyerID.Valid {
			item.BuyerID = &buyerID.String
		}
		
		items = append(items, item)
	}
	
	if err := rows.Err(); err != nil {
		return nil, err
	}
	
	return items, nil
}

func (r *ItemRepository) GetByID(id string) (*model.Item, error) {
	query := `
		SELECT id, name, price, description, user_id, buyer_id, status 
		FROM items 
		WHERE id = ?
	`
	row := r.db.QueryRow(query, id)
	
	var item model.Item
	var buyerID sql.NullString
	
	if err := row.Scan(&item.ID, &item.Name, &item.Price, &item.Description, &item.UserID, &buyerID, &item.Status); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Not found
		}
		return nil, err
	}
	
	if buyerID.Valid {
		item.BuyerID = &buyerID.String
	}
	
	return &item, nil
}

func (r *ItemRepository) Insert(item *model.Item) error {
	query := `INSERT INTO items (id, name, price, description, user_id, status) VALUES (?, ?, ?, ?, ?, ?)`
	_, err := r.db.Exec(query, item.ID, item.Name, item.Price, item.Description, item.UserID, item.Status)
	return err
}

func (r *ItemRepository) Update(item *model.Item) error {
	query := `UPDATE items SET name=?, price=?, description=?, user_id=?, buyer_id=?, status=? WHERE id=?`
	_, err := r.db.Exec(query, item.Name, item.Price, item.Description, item.UserID, item.BuyerID, item.Status, item.ID)
	return err
}
