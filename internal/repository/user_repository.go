package repository

import (
	"backend-v2/model"

	"gorm.io/gorm"
)

type userRepository struct {
	db *gorm.DB
}

func InitUserRepository(db *gorm.DB) IUserRepository {
	return &userRepository{
		db: db,
	}
}

type IUserRepository interface {
	GetAll() ([]model.User, error)
	GetPaginated(page, limit int) ([]model.User, error)
	GetPaginatedFiltered(page, limit int, filter model.User) ([]model.User, error)
	GetByID(id uint) (model.User, error)
	Create(data model.User) (model.User, error)
	Update(data model.User) (model.User, error)
	Delete(id uint) error
	Count() (int64, error)
	GetByUsername(username string) (model.User, error)
}

func (repo *userRepository) GetAll() ([]model.User, error) {
	data := []model.User{}
	err := repo.db.Order("id desc").Find(&data).Error
	return data, err
}

func (repo *userRepository) GetPaginated(page, limit int) ([]model.User, error) {
	data := []model.User{}
	err := repo.db.Order("id desc").Offset((page - 1) * limit).Limit(limit).Find(&data).Error
	return data, err
}

func (repo *userRepository) GetPaginatedFiltered(page, limit int, filter model.User) ([]model.User, error) {
	data := []model.User{}
	err := repo.db.
		Raw(`SELECT * FROM users WHERE
			(nullif(?, '') IS NULL OR worker_id = ?) AND
			(nullif(?, '') IS NULL OR username = ?) ORDER BY id DESC LIMIT ? OFFSET ?`,
			filter.WorkerID, filter.WorkerID, 
      filter.Username, filter.Username, 
      limit, (page-1)*limit,
		).
		Scan(&data).Error

	return data, err
}

func (repo *userRepository) GetByID(id uint) (model.User, error) {
	data := model.User{}
	err := repo.db.Find(&data, "id = ?", id).Error
	return data, err
}

func (repo *userRepository) Create(data model.User) (model.User, error) {
	err := repo.db.Create(&data).Error
	return data, err
}

func (repo *userRepository) Update(data model.User) (model.User, error) {
	err := repo.db.Model(&model.User{}).Select("*").Updates(&data).Error
	return data, err
}

func (repo *userRepository) Delete(id uint) error {
	return repo.db.Delete(&model.User{}, "id = ?", id).Error
}

func (repo *userRepository) Count() (int64, error) {
	var count int64
	err := repo.db.Model(&model.User{}).Count(&count).Error
	return count, err
}

func (repo *userRepository) GetByUsername(username string) (model.User, error) {
	var data model.User
	err := repo.db.First(&data, "username = ?", username).Error
	return data, err
}
