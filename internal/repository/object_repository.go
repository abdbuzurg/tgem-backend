package repository

import (
	"backend-v2/internal/dto"
	"backend-v2/model"

	"gorm.io/gorm"
)

type objectRepository struct {
	db *gorm.DB
}

func InitObjectRepository(db *gorm.DB) IObjectRepository {
	return &objectRepository{
		db: db,
	}
}

type IObjectRepository interface {
	GetAll() ([]model.Object, error)
	GetPaginated(page, limit int) ([]model.Object, error)
	GetPaginatedFiltered(page, limit int, filter model.Object) ([]dto.ObjectPaginatedQuery, error)
	GetByID(id uint) (model.Object, error)
	GetByRangeOfIDs(ids []uint) ([]model.Object, error)
	Create(data model.Object) (model.Object, error)
	Update(data model.Object) (model.Object, error)
	Delete(id uint) error
	Count() (int64, error)
	GetByName(name string) (model.Object, error)
}

func (repo *objectRepository) GetAll() ([]model.Object, error) {
	data := []model.Object{}
	err := repo.db.Order("id desc").Find(&data).Error
	return data, err
}

func (repo *objectRepository) GetPaginated(page, limit int) ([]model.Object, error) {
	data := []model.Object{}
	err := repo.db.Order("id desc").Offset((page - 1) * limit).Limit(limit).Find(&data).Error
	return data, err
}

func (repo *objectRepository) GetPaginatedFiltered(page, limit int, filter model.Object) ([]dto.ObjectPaginatedQuery, error) {
	var data []dto.ObjectPaginatedQuery
	err := repo.db.
		Raw(`SELECT
          objects.id AS id,
          objects.type AS object_type, 
          objects.name AS object_name, 
          objects.status AS object_status, 
          workers.name AS  supervisor_name
        FROM supervisor_objects
        INNER JOIN objects ON objects.id = supervisor_objects.object_id 
        INNER JOIN workers ON workers.id = supervisor_objects.supervisor_worker_id
        WHERE
          (nullif(?, 0) IS NULL OR objects.object_detailed_id = ?) AND
          (nullif(?, '') IS NULL OR objects.type = ?) AND
          (nullif(?, '') IS NULL OR objects.name = ?) AND
          (nullif(?, '') IS NULL OR objects.status = ?) ORDER BY objects.id DESC LIMIT ? OFFSET ?`,
			filter.ObjectDetailedID, filter.ObjectDetailedID,
			filter.Type, filter.Type,
			filter.Name, filter.Name,
			filter.Status, filter.Status, limit, (page-1)*limit,
		).
		Scan(&data).Error

	return data, err
}

func (repo *objectRepository) GetByID(id uint) (model.Object, error) {
	data := model.Object{}
	err := repo.db.Find(&data, "id = ?", id).Error
	return data, err
}

func (repo *objectRepository) Create(data model.Object) (model.Object, error) {
	err := repo.db.Create(&data).Error
	return data, err
}

func (repo *objectRepository) Update(data model.Object) (model.Object, error) {
	err := repo.db.Model(&model.Object{}).Select("*").Updates(&data).Error
	return data, err
}

func (repo *objectRepository) Delete(id uint) error {
	return repo.db.Delete(&model.Object{}, "id = ?", id).Error
}

func (repo *objectRepository) Count() (int64, error) {
	var count int64
	err := repo.db.Model(&model.Object{}).Count(&count).Error
	return count, err
}

func (repo *objectRepository) GetByName(name string) (model.Object, error) {
	var data model.Object
	err := repo.db.
		Raw("SELECT * FROM objects WHERE name = ?").
		Error
	return data, err
}

func (repo *objectRepository) GetByRangeOfIDs(ids []uint) ([]model.Object, error) {
	var data []model.Object
	err := repo.db.Model(model.Object{}).Select("*").Where("id IN ?", ids).Scan(&data).Error
	return data, err
}
