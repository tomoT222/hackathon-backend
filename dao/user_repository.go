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

func (r *UserRepository) GetByEmail(email string) (*model.User, error) {
	var user model.User
	query := `SELECT id, name, email FROM users WHERE email = ?`
	err := r.db.QueryRow(query, email).Scan(&user.ID, &user.Name, &user.Email)
	if err != nil {
		return nil, err
	}
	return &user, nil
}
