package repository

import (
	"backend-v2/model"

	"gorm.io/gorm"
)

type operationMaterialRepository struct {
	db *gorm.DB
}

func InitOperationMaterialRepository(db *gorm.DB) IOperationMaterialRepository {
	return &operationMaterialRepository{
		db: db,
	}
}

type IOperationMaterialRepository interface {
	GetByMaterialCostID(materialCostID uint) (model.OperationMaterial, error)
}

func (repo *operationMaterialRepository) GetByMaterialCostID(materialCostID uint) (model.OperationMaterial, error) {
	result := model.OperationMaterial{}
	err := repo.db.Raw(`
    SELECT *
    FROM operation_materials
    WHERE operation_materials.material_id IN (
      SELECT materials.id
      FROM materials
      WHERE materials.id IN (
        SELECT DISTINCT(material_costs.material_id)
        FROM material_costs
        WHERE material_costs.id = ?
      )
    )
    `, materialCostID).Scan(&result).Error

	return result, err
}
