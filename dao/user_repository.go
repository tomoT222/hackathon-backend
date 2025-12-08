package dao

import (
	"database/sql"
	"hackathon-backend/model"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Insert(user *model.User) error {
	query := `INSERT INTO users (id, name, email) VALUES (?, ?, ?)`
	_, err := r.db.Exec(query, user.ID, user.Name, user.Email)
	return err
}
