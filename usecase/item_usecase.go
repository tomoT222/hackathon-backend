package usecase

import (
	"hackathon-backend/dao"
	"hackathon-backend/model"
)

type ItemUsecase struct {
	repo *dao.ItemRepository
}

func NewItemUsecase(repo *dao.ItemRepository) *ItemUsecase {
	return &ItemUsecase{repo: repo}
}

func (u *ItemUsecase) GetAllItems() ([]model.Item, error) {
	return u.repo.GetAll()
}

func (u *ItemUsecase) GetItemByID(id string) (*model.Item, error) {
	return u.repo.GetByID(id)
}
