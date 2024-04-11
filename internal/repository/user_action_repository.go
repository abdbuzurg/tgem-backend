package repository

import (
	"backend-v2/model"

	"gorm.io/gorm"
)

type userActionRepository struct {
	db *gorm.DB
}

func InitUserActionRepository(db *gorm.DB) IUserActionRepository {
	return &userActionRepository{
		db: db,
	}
}

type IUserActionRepository interface {
  GetAllByUserID(userID uint) ([]model.UserAction, error)
  Create(data model.UserAction) (model.UserAction, error)
}

func (repo *userActionRepository) GetAllByUserID(userID uint) ([]model.UserAction, error) {
  var data []model.UserAction
  err := repo.db.Order("id DESC").Find(&data, "user_id = ?", userID).Error
  return data, err
}

func (repo *userActionRepository) Create(data model.UserAction) (model.UserAction, error) {
	err := repo.db.Create(&data).Error
	return data, err
}
