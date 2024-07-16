package repository

import (
	"backend-v2/internal/dto"
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
	GetPaginated(page, limit int, projectID uint) ([]dto.MaterialCostView, error)
	GetPaginatedFiltered(page, limit int, filter dto.MaterialCostSearchFilter) ([]dto.MaterialCostView, error)
	GetByID(id uint) (model.MaterialCost, error)
	GetByMaterialID(materialID uint) ([]model.MaterialCost, error)
	GetByMaterialIDSorted(materialID uint) ([]model.MaterialCost, error)
	Create(data model.MaterialCost) (model.MaterialCost, error)
	Update(data model.MaterialCost) (model.MaterialCost, error)
	Delete(id uint) error
	Count(filter dto.MaterialCostSearchFilter) (int64, error)
	CreateInBatch(data []model.MaterialCost) ([]model.MaterialCost, error)
}

func (repo *materialCostRepository) GetAll() ([]model.MaterialCost, error) {
	data := []model.MaterialCost{}
	err := repo.db.Order("id desc").Find(&data).Error
	return data, err
}

func (repo *materialCostRepository) GetPaginated(page, limit int, projectID uint) ([]dto.MaterialCostView, error) {
	data := []dto.MaterialCostView{}
	err := repo.db.Raw(`
    SELECT 
        material_costs.id as id,
        material_costs.cost_prime as cost_prime,
        material_costs.cost_m19 as cost_m19,
        material_costs.cost_with_customer as cost_with_customer,
        materials.name as material_name
    FROM material_costs
    INNER JOIN materials ON materials.id = material_costs.material_id
    WHERE materials.project_id = ?
    ORDER BY material_costs.id DESC LIMIT ? OFFSET ?
    `, projectID, limit, (page-1)*limit).Scan(&data).Error
	return data, err
}

func (repo *materialCostRepository) GetPaginatedFiltered(page, limit int, filter dto.MaterialCostSearchFilter) ([]dto.MaterialCostView, error) {
	data := []dto.MaterialCostView{}
	err := repo.db.
		Raw(`
      SELECT 
        material_costs.id as id,
        material_costs.cost_prime as cost_prime,
        material_costs.cost_m19 as cost_m19,
        material_costs.cost_with_customer as cost_with_customer,
        materials.name as material_name
      FROM material_costs 
      INNER JOIN materials ON materials.id = material_costs.material_id
      WHERE 
        materials.project_id = ? AND
        (nullif(?, '') IS NULL OR materials.name = ?) AND
      ORDER BY id DESC LIMIT ? OFFSET ?`,
			filter.ProjectID,
			filter.MaterialName, filter.MaterialName,
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

func (repo *materialCostRepository) GetByMaterialIDSorted(materialID uint) ([]model.MaterialCost, error) {
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
	err := repo.db.Model(&model.MaterialCost{}).Select("*").Where("id = ?", data.ID).Updates(&data).Error
	return data, err
}

func (repo *materialCostRepository) Delete(id uint) error {
	return repo.db.Delete(&model.MaterialCost{}, "id = ?", id).Error
}

func (repo *materialCostRepository) Count(filter dto.MaterialCostSearchFilter) (int64, error) {
	var count int64
	err := repo.db.Raw(`
    SELECT COUNT(*)
    FROM material_costs
    INNER JOIN materials ON materials.id = material_costs.material_id
    WHERE
      materials.project_id = ? AND
      (nullif(?, '') IS NULL OR materials.name = ?)
    `,
		filter.ProjectID,
		filter.MaterialName, filter.MaterialName,
	).Scan(&count).Error
	return count, err
}

func (repo *materialCostRepository) CreateInBatch(data []model.MaterialCost) ([]model.MaterialCost, error) {
	err := repo.db.CreateInBatches(&data, 15).Error
	return data, err
}
