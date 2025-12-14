package usecase

import (
	"hackathon-backend/dao"
	"hackathon-backend/model"
	"math/rand"
	"time"

	"github.com/oklog/ulid/v2"
)

type UserUsecase struct {
	repo *dao.UserRepository
}

func NewUserUsecase(repo *dao.UserRepository) *UserUsecase {
	return &UserUsecase{repo: repo}
}

func (u *UserUsecase) RegisterUser(name, email string) (*model.User, error) {
    // 1. Check if user exists (Login)
    existingUser, err := u.repo.GetByEmail(email)
    if err == nil && existingUser != nil {
        return existingUser, nil
    }

	// 2. Register New User
	// Generate ULID
	entropy := ulid.Monotonic(rand.New(rand.NewSource(time.Now().UnixNano())), 0)
	id := ulid.MustNew(ulid.Now(), entropy).String()

	user := &model.User{
		ID:    id,
		Name:  name,
		Email: email,
	}

	if err := u.repo.Insert(user); err != nil {
		return nil, err
	}

	return user, nil
}
