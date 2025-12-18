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

func (r *UserRepository) GetByID(id string) (*model.User, error) {
	var user model.User
	query := `SELECT id, name, email FROM users WHERE id = ?`
	err := r.db.QueryRow(query, id).Scan(&user.ID, &user.Name, &user.Email)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) Update(user *model.User) error {
	query := `UPDATE users SET name = ?, email = ? WHERE id = ?`
	_, err := r.db.Exec(query, user.Name, user.Email, user.ID)
	return err
}
