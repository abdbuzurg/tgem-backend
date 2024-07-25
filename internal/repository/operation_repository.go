package repository

import (
	"backend-v2/internal/dto"
	"backend-v2/model"
	"errors"

	"gorm.io/gorm"
)

type operationRepository struct {
	db *gorm.DB
}

func InitOperationRepository(db *gorm.DB) IOperationRepository {
	return &operationRepository{
		db: db,
	}
}

type IOperationRepository interface {
	GetPaginated(page, limit int, filter dto.OperationSearchParameters) ([]dto.OperationPaginated, error)
	GetByID(id uint) (model.Operation, error)
	GetByName(name string, projectID uint) (model.Operation, error)
	GetAll(projectID uint) ([]model.Operation, error)
	Create(data dto.Operation) (model.Operation, error)
	Update(data dto.Operation) (model.Operation, error)
	Delete(id uint) error
	Count(filter dto.OperationSearchParameters) (int64, error)
}

func (repo *operationRepository) GetPaginated(page, limit int, filter dto.OperationSearchParameters) ([]dto.OperationPaginated, error) {
	data := []dto.OperationPaginated{}
	err := repo.db.
		Raw(`
      SELECT 
        operations.id as id,
        operations.name as name,
        operations.code as code,
        operations.cost_prime as cost_prime,
        operations.cost_with_customer as cost_with_customer,
        materials.id as material_id,
        materials.name as material_name
      FROM operations 
      FULL JOIN operation_materials ON operation_materials.operation_id = operations.id
      FULL JOIN materials ON operation_materials.material_id = materials.id
      WHERE
        operations.project_id = ? AND
        (nullif(?, '') IS NULL OR operations.name = ?) AND
        (nullif(?, '') IS NULL OR operations.code = ?) AND
        (nullif(?, 0) IS NULL OR materials.id = ?) 
			ORDER BY operations.id DESC 
      LIMIT ? 
      OFFSET ?`,
			filter.ProjectID,
			filter.Name, filter.Name,
			filter.Code, filter.Code,
			filter.MaterialID, filter.MaterialID,
			limit, (page-1)*limit,
		).
		Scan(&data).Error

	return data, err
}

func (repo *operationRepository) GetByID(id uint) (model.Operation, error) {
	data := model.Operation{}
	err := repo.db.Find(&data, "id = ?", id).Error
	return data, err
}

func (repo *operationRepository) Create(data dto.Operation) (model.Operation, error) {
	result := model.Operation{
		Code:             data.Code,
		Name:             data.Name,
		ProjectID:        data.ProjectID,
		CostPrime:        data.CostPrime,
		CostWithCustomer: data.CostWithCustomer,
	}

	err := repo.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&result).Error; err != nil {
			return err
		}

		if data.MaterialID != 0 {
			if err := tx.Create(&model.OperationMaterial{OperationID: result.ID, MaterialID: data.MaterialID}).Error; err != nil {
				return err
			}
		}

		return nil
	})

	return result, err
}

func (repo *operationRepository) Update(data dto.Operation) (model.Operation, error) {
	result := model.Operation{
		ID:               data.ID,
		Name:             data.Name,
		Code:             data.Code,
		ProjectID:        data.ProjectID,
		CostPrime:        data.CostPrime,
		CostWithCustomer: data.CostWithCustomer,
	}

	err := repo.db.Transaction(func(tx *gorm.DB) error {

		if err := tx.Model(&model.Operation{}).Select("*").Where("id = ?", result.ID).Updates(&result).Error; err != nil {
			return err
		}

		if err := tx.Delete(&model.OperationMaterial{}, "operation_id = ?", result.ID).Error; err != nil {
			return err
		}

		if data.MaterialID != 0 {
			if err := tx.Create(&model.OperationMaterial{OperationID: result.ID, MaterialID: data.MaterialID}).Error; err != nil {
				return err
			}
		}

		return nil
	})

	return result, err
}

func (repo *operationRepository) Delete(id uint) error {
	return repo.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Delete(&model.OperationMaterial{}, "operation_id = ?", id).Error; err != nil {
			return err
		}

		if err := tx.Delete(&model.Operation{}, "id = ?", id).Error; err != nil {
			return err
		}

		return nil
	})
}

func (repo *operationRepository) Count(filter dto.OperationSearchParameters) (int64, error) {
	count := int64(0)
	err := repo.db.Raw(`
    SELECT COUNT(*)
    FROM operations
    FULL JOIN operation_materials ON operation_materials.operation_id = operations.id
    FULL JOIN materials ON  operation_materials.material_id = materials.id
    WHERE
      operations.project_id = ? AND
      (nullif(?, '') IS NULL OR operations.name = ?) AND
      (nullif(?, '') IS NULL OR operations.code = ?) AND
      (nullif(?, 0) IS NULL OR materials.id = ?)
    `,
		filter.ProjectID,
		filter.Name, filter.Name,
		filter.Code, filter.Code,
		filter.MaterialID, filter.MaterialID,
	).Scan(&count).Error

	return count, err
}

func (repo *operationRepository) GetByName(name string, projectID uint) (model.Operation, error) {
	result := model.Operation{}
	err := repo.db.First(&result, "name = ? AND project_id = ?", name, projectID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = nil
	}
	return result, err
}

func (repo *operationRepository) GetAll(projectID uint) ([]model.Operation, error) {
	result := []model.Operation{}
	err := repo.db.Find(&result, "project_id = ?", projectID).Error
	return result, err
}
