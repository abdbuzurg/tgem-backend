package repository

import (
	"backend-v2/internal/dto"
	"backend-v2/model"

	"gorm.io/gorm"
)

type invoiceMaterialsRepository struct {
	db *gorm.DB
}

func InitInvoiceMaterialsRepository(db *gorm.DB) IInvoiceMaterialsRepository {
	return &invoiceMaterialsRepository{
		db: db,
	}
}

type IInvoiceMaterialsRepository interface {
	GetAll() ([]model.InvoiceMaterials, error)
	GetPaginated(page, limit int) ([]model.InvoiceMaterials, error)
	GetPaginatedFiltered(page, limit int, filter model.InvoiceMaterials) ([]model.InvoiceMaterials, error)
	GetByID(id uint) (model.InvoiceMaterials, error)
	Create(data model.InvoiceMaterials) (model.InvoiceMaterials, error)
	Update(data model.InvoiceMaterials) (model.InvoiceMaterials, error)
	Delete(id uint) error
	Count() (int64, error)
	GetByInvoice(projectID, invoiceID uint, invoceType string) ([]model.InvoiceMaterials, error)
  GetByMaterialCostID(materialCostID uint, invoiceType string, invoiceID uint) (model.InvoiceMaterials, error)
  GetByInvoiceData(projectID, invoiceID uint, invoiceType string) ([]dto.InvoiceMaterialsView, error)
}

func (repo *invoiceMaterialsRepository) GetAll() ([]model.InvoiceMaterials, error) {
	data := []model.InvoiceMaterials{}
	err := repo.db.Order("id desc").Find(&data).Error
	return data, err
}

func (repo *invoiceMaterialsRepository) GetPaginated(page, limit int) ([]model.InvoiceMaterials, error) {
	data := []model.InvoiceMaterials{}
	err := repo.db.Order("id desc").Offset((page - 1) * limit).Limit(limit).Find(&data).Error
	return data, err
}

func (repo *invoiceMaterialsRepository) GetPaginatedFiltered(page, limit int, filter model.InvoiceMaterials) ([]model.InvoiceMaterials, error) {
	data := []model.InvoiceMaterials{}
	err := repo.db.
		Raw(`SELECT * FROM invoice_materials WHERE
			(nullif(?, '') IS NULL OR material_cost_id = ?) AND
			(nullif(?, '') IS NULL OR invoice_id = ?) AND
			(nullif(?, '') IS NULL OR amount = ?)ORDER BY id DESC LIMIT ? OFFSET ?`,
			filter.MaterialCostID, filter.MaterialCostID,
			filter.InvoiceID, filter.InvoiceID,
			filter.Amount, filter.Amount,
			limit, (page-1)*limit,
		).
		Scan(&data).Error

	return data, err
}

func (repo *invoiceMaterialsRepository) GetByID(id uint) (model.InvoiceMaterials, error) {
	data := model.InvoiceMaterials{}
	err := repo.db.Find(&data, "id = ?", id).Error
	return data, err
}

func (repo *invoiceMaterialsRepository) Create(data model.InvoiceMaterials) (model.InvoiceMaterials, error) {
	err := repo.db.Create(&data).Error
	return data, err
}

func (repo *invoiceMaterialsRepository) Update(data model.InvoiceMaterials) (model.InvoiceMaterials, error) {
	err := repo.db.Model(&model.InvoiceMaterials{}).Select("*").Updates(&data).Error
	return data, err
}

func (repo *invoiceMaterialsRepository) Delete(id uint) error {
	return repo.db.Delete(&model.InvoiceMaterials{}, "id = ?", id).Error
}

func (repo *invoiceMaterialsRepository) Count() (int64, error) {
	var count int64
	err := repo.db.Model(&model.InvoiceMaterials{}).Count(&count).Error
	return count, err
}

func (repo *invoiceMaterialsRepository) GetByInvoice(projectID, invoiceID uint, invoiceType string) ([]model.InvoiceMaterials, error) {
	data := []model.InvoiceMaterials{}
	err := repo.db.Find(&data, "invoice_id = ? AND invoice_type = ? AND project_id = ?", invoiceID, invoiceType, projectID).Error
	return data, err
}

func(repo *invoiceMaterialsRepository) GetByMaterialCostID(
  materialCostID uint, 
  invoiceType string, 
  invoiceID uint,
) (model.InvoiceMaterials, error) {
  var data model.InvoiceMaterials
  err := repo.db.Raw(`
    SELECT * FROM invoice_materials WHERE material_cost_id = ? AND invoice_type = ? AND invoice_id = ?
    `, materialCostID, invoiceType, invoiceID).Scan(&data).Error
  return data, err
}

func(repo *invoiceMaterialsRepository) GetByInvoiceData(
  projectID, invoiceID uint, 
  invoiceType string,
) ([]dto.InvoiceMaterialsView, error) {
  data := []dto.InvoiceMaterialsView{}
  err := repo.db.Raw(`
    SELECT 
      invoice_materials.id as id,
      materials.name as material_name,
      material_costs.cost_m19 as cost_m19,
      invoice_materials.amount as amount,
      invoice_materials.notes as notes
    FROM invoice_materials
      INNER JOIN material_costs ON material_costs.id = invoice_materials.material_cost_id
      INNER JOIN materials ON materials.id = material_costs.material_id
    WHERE
      invoice_materials.project_id = ? AND
      invoice_materials.invoice_type = ? AND
      invoice_materials.invoice_id = ?;
    `, projectID, invoiceType, invoiceID).Scan(&data).Error

  return data, err
}
