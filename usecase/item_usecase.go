package usecase

import (
	"hackathon-backend/dao"
	"hackathon-backend/model"
	"errors"
	"math/rand"
	"time"

	"github.com/oklog/ulid/v2"
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

func (u *ItemUsecase) CreateItem(name string, price int, description string, userID string) (*model.Item, error) {
	// Generate ULID
	entropy := ulid.Monotonic(rand.New(rand.NewSource(time.Now().UnixNano())), 0)
	id := ulid.MustNew(ulid.Now(), entropy).String()

	item := &model.Item{
		ID:          id,
		Name:        name,
		Price:       price,
		Description: description,
		UserID:      userID,
		Status:      "on_sale",
	}

	if err := u.repo.Insert(item); err != nil {
		return nil, err
	}

	return item, nil
}

func (u *ItemUsecase) PurchaseItem(itemID string, buyerID string) (*model.Item, error) {
	item, err := u.repo.GetByID(itemID)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, errors.New("item not found")
	}

	if item.Status == "sold" {
		return nil, errors.New("item already sold")
	}

	item.BuyerID = &buyerID
	item.Status = "sold"

	if err := u.repo.Update(item); err != nil {
		return nil, err
	}

	return item, nil
}
