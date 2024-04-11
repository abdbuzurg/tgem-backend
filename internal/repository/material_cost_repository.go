package repository

import (
	"backend-v2/model"

	"gorm.io/gorm"
)

type materialCostRepository struct {
	db *gorm.DB
}

func InitMaterialCostRepository(db *gorm.DB) IMaterialCostRepository {
	return &materialCostRepository{
		db: db,
	}
}

type IMaterialCostRepository interface {
	GetAll() ([]model.MaterialCost, error)
	GetPaginated(page, limit int) ([]model.MaterialCost, error)
	GetPaginatedFiltered(page, limit int, filter model.MaterialCost) ([]model.MaterialCost, error)
	GetByID(id uint) (model.MaterialCost, error)
	GetByMaterialID(materialID uint) ([]model.MaterialCost, error)
  GetByMaterialIDSorted(materialID uint) ([]model.MaterialCost, error)
	Create(data model.MaterialCost) (model.MaterialCost, error)
	Update(data model.MaterialCost) (model.MaterialCost, error)
	Delete(id uint) error
	Count() (int64, error)
}

func (repo *materialCostRepository) GetAll() ([]model.MaterialCost, error) {
	data := []model.MaterialCost{}
	err := repo.db.Order("id desc").Find(&data).Error
	return data, err
}

func (repo *materialCostRepository) GetPaginated(page, limit int) ([]model.MaterialCost, error) {
	data := []model.MaterialCost{}
	err := repo.db.Order("id desc").Offset((page - 1) * limit).Limit(limit).Find(&data).Error
	return data, err
}

func (repo *materialCostRepository) GetPaginatedFiltered(page, limit int, filter model.MaterialCost) ([]model.MaterialCost, error) {
	data := []model.MaterialCost{}
	err := repo.db.
		Raw(`SELECT * FROM materials WHERE
			(nullif(?, '') IS NULL OR material_id = ?) AND
			(nullif(?, '') IS NULL OR cost_prime = ?) AND
			(nullif(?, '') IS NULL OR cost_m19 = ?) AND
			(nullif(?, '') IS NULL OR cost_with_customer = ?) ORDER BY id DESC LIMIT ? OFFSET ?`,
			filter.MaterialID, filter.MaterialID,
			filter.CostPrime, filter.CostPrime,
			filter.CostM19, filter.CostM19,
			filter.CostWithCustomer, filter.CostWithCustomer,
			limit, (page-1)*limit,
		).
		Scan(&data).Error

	return data, err
}

func (repo *materialCostRepository) GetByID(id uint) (model.MaterialCost, error) {
	data := model.MaterialCost{}
	err := repo.db.Find(&data, "id = ?", id).Error
	return data, err
}

func (repo *materialCostRepository) GetByMaterialID(materialID uint) ([]model.MaterialCost, error) {
	data := []model.MaterialCost{}
	err := repo.db.Find(&data, "material_id = ?", materialID).Error
	return data, err
}

func(repo *materialCostRepository)   GetByMaterialIDSorted(materialID uint) ([]model.MaterialCost, error) {
  data := []model.MaterialCost{}
  err := repo.db.Raw(`
    SELECT * FROM material_costs WHERE material_id = ? ORDER BY cost_m19 DESC
  `, materialID).Scan(&data).Error
  return data, err
}

func (repo *materialCostRepository) Create(data model.MaterialCost) (model.MaterialCost, error) {
	err := repo.db.Create(&data).Error
	return data, err
}

func (repo *materialCostRepository) Update(data model.MaterialCost) (model.MaterialCost, error) {
	err := repo.db.Model(&model.MaterialCost{}).Select("*").Updates(&data).Error
	return data, err
}

func (repo *materialCostRepository) Delete(id uint) error {
	return repo.db.Delete(&model.MaterialCost{}, "id = ?", id).Error
}

func (repo *materialCostRepository) Count() (int64, error) {
	var count int64
	err := repo.db.Model(&model.MaterialCost{}).Count(&count).Error
	return count, err
}
