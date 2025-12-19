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
	// 修正: buyer_idとstatus, image_urlも取得するように変更
	// schema.sql: id, name, price, description, user_id, buyer_id, status, image_url, initial_price
	query := `
		SELECT id, name, price, description, user_id, buyer_id, status, image_url, initial_price
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
		var imageURL sql.NullString
		
		if err := rows.Scan(&item.ID, &item.Name, &item.Price, &item.Description, &item.UserID, &buyerID, &item.Status, &imageURL, &item.InitialPrice); err != nil {
			return nil, err
		}
		
		if buyerID.Valid {
			item.BuyerID = &buyerID.String
		}
		if imageURL.Valid {
			item.ImageURL = imageURL.String
		}
		
		items = append(items, item)
	}
	
	if err := rows.Err(); err != nil {
		return nil, err
	}
	
	return items, nil
}

func (r *ItemRepository) IncrementViewCount(id string) error {
	_, err := r.db.Exec("UPDATE items SET views_count = views_count + 1 WHERE id = ?", id)
	return err
}

func (r *ItemRepository) GetByID(id string) (*model.Item, error) {
	// Select with new columns
	query := `
		SELECT id, name, price, description, user_id, buyer_id, status, views_count, ai_negotiation_enabled, min_price, created_at, image_url, initial_price 
		FROM items 
		WHERE id = ?
	`
	row := r.db.QueryRow(query, id)
	
	var item model.Item
	var buyerID sql.NullString
	var minPrice sql.NullInt64
	var imageURL sql.NullString
	
	if err := row.Scan(&item.ID, &item.Name, &item.Price, &item.Description, &item.UserID, &buyerID, &item.Status, &item.ViewsCount, &item.AINegotiationEnabled, &minPrice, &item.CreatedAt, &imageURL, &item.InitialPrice); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Not found
		}
		return nil, err
	}
	
	if buyerID.Valid {
		item.BuyerID = &buyerID.String
	}
	if minPrice.Valid {
		val := int(minPrice.Int64)
		item.MinPrice = &val
	}
	if imageURL.Valid {
		item.ImageURL = imageURL.String
	}
	
	return &item, nil
}

func (r *ItemRepository) Insert(item *model.Item) error {
	query := `INSERT INTO items (id, name, price, description, user_id, status, ai_negotiation_enabled, min_price, image_url, initial_price) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.Exec(query, item.ID, item.Name, item.Price, item.Description, item.UserID, item.Status, item.AINegotiationEnabled, item.MinPrice, item.ImageURL, item.InitialPrice)
	return err
}

func (r *ItemRepository) Update(item *model.Item) error {
	query := `UPDATE items SET name=?, price=?, description=?, user_id=?, buyer_id=?, status=?, ai_negotiation_enabled=?, min_price=?, image_url=?, initial_price=? WHERE id=?`
	_, err := r.db.Exec(query, item.Name, item.Price, item.Description, item.UserID, item.BuyerID, item.Status, item.AINegotiationEnabled, item.MinPrice, item.ImageURL, item.InitialPrice, item.ID)
	return err
}
