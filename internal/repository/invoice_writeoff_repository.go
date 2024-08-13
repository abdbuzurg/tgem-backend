package repository

import (
	"backend-v2/internal/dto"
	"backend-v2/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type invoiceWriteOffRepository struct {
	db *gorm.DB
}

func InitInvoiceWriteOffRepository(db *gorm.DB) IInvoiceWriteOffRepository {
	return &invoiceWriteOffRepository{
		db: db,
	}
}

type IInvoiceWriteOffRepository interface {
	GetAll() ([]model.InvoiceWriteOff, error)
	GetPaginated(page, limit int, filter dto.InvoiceWriteOffSearchParameters) ([]dto.InvoiceWriteOffPaginated, error)
	GetByID(id uint) (model.InvoiceWriteOff, error)
	Create(data dto.InvoiceWriteOffMutationData) (model.InvoiceWriteOff, error)
	Update(data dto.InvoiceWriteOffMutationData) (model.InvoiceWriteOff, error)
	Delete(id uint) error
	Count(filter dto.InvoiceWriteOffSearchParameters) (int64, error)
	GetMaterialsForEdit(id uint) ([]dto.InvoiceWriteOffMaterialsForEdit, error)
	Confirmation(data dto.InvoiceWriteOffConfirmationData) error
}

func (repo *invoiceWriteOffRepository) GetAll() ([]model.InvoiceWriteOff, error) {
	data := []model.InvoiceWriteOff{}
	err := repo.db.Order("id desc").Find(&data).Error
	return data, err
}

func (repo *invoiceWriteOffRepository) GetPaginated(page, limit int, filter dto.InvoiceWriteOffSearchParameters) ([]dto.InvoiceWriteOffPaginated, error) {
	data := []dto.InvoiceWriteOffPaginated{}
	err := repo.db.
		Raw(`
      SELECT 
        invoice_write_offs.id as id,
        invoice_write_offs.write_off_type as write_off_type,
        invoice_write_offs.released_worker_id as released_worker_id,
        workers.name as released_worker_name,
        invoice_write_offs.delivery_code as delivery_code,
        invoice_write_offs.date_of_invoice as date_of_invoice,
        invoice_write_offs.confirmation as confirmation,
        invoice_write_offs.date_of_confirmation as date_of_confirmation
      FROM invoice_write_offs 
      INNER JOIN workers ON workers.id = invoice_write_offs.released_worker_id
      WHERE
      invoice_write_offs.project_id = ? AND
			(nullif(?, '') IS NULL OR invoice_write_offs.write_off_type = ?) 
      ORDER BY invoice_write_offs.id DESC 
      LIMIT ? 
      OFFSET ?`,
			filter.ProjectID,
			filter.WriteOffType, filter.WriteOffType,
			limit, (page-1)*limit,
		).
		Scan(&data).Error

	return data, err
}

func (repo *invoiceWriteOffRepository) GetByID(id uint) (model.InvoiceWriteOff, error) {
	data := model.InvoiceWriteOff{}
	err := repo.db.First(&data, "id = ?", id).Error
	return data, err
}

func (repo *invoiceWriteOffRepository) Create(data dto.InvoiceWriteOffMutationData) (model.InvoiceWriteOff, error) {
	result := data.InvoiceWriteOff
	err := repo.db.Transaction(func(tx *gorm.DB) error {

		if err := tx.Create(&result).Error; err != nil {
			return err
		}

		for index := range data.InvoiceMaterials {
			data.InvoiceMaterials[index].InvoiceID = result.ID
		}

		if err := tx.CreateInBatches(&data.InvoiceMaterials, 15).Error; err != nil {
			return err
		}

		err := tx.Exec(`
      UPDATE invoice_counts
      SET count = count + 1
      WHERE
        project_id = ? AND
        invoice_type = 'writeoff';
      `, result.ProjectID).Error
		if err != nil {
			return err
		}

		return nil
	})

	return result, err
}

func (repo *invoiceWriteOffRepository) Update(data dto.InvoiceWriteOffMutationData) (model.InvoiceWriteOff, error) {
	result := data.InvoiceWriteOff
	err := repo.db.Transaction(func(tx *gorm.DB) error {

		err := tx.Model(&result).Select("*").Where("id = ?", result.ID).Updates(&result).Error
		if err != nil {
			return err
		}

		err = tx.Exec(`
      DELETE FROM invoice_materials
      WHERE invoice_type = 'writeoff' AND invoice_id = ?
    `, result.ID).Error
		if err != nil {
			return err
		}

		for index := range data.InvoiceMaterials {
			data.InvoiceMaterials[index].InvoiceID = result.ID
		}

		if err := tx.CreateInBatches(&data.InvoiceMaterials, 15).Error; err != nil {
			return err
		}

		return nil
	})

	return result, err
}

func (repo *invoiceWriteOffRepository) Delete(id uint) error {
	return repo.db.Transaction(func(tx *gorm.DB) error {
		err := tx.Exec(`
      DELETE FROM invoice_materials
      WHERE invoice_type = 'writeoff' AND invoice_id = ?
    `, id).Error
		if err != nil {
			return err
		}

		if err := tx.Delete(model.InvoiceWriteOff{}, "id = ?", id).Error; err != nil {
			return err
		}

		return nil
	})
}

func (repo *invoiceWriteOffRepository) Count(filter dto.InvoiceWriteOffSearchParameters) (int64, error) {
	var count int64
	err := repo.db.Raw(`
      SELECT COUNT(*)
      FROM invoice_write_offs 
      WHERE
      invoice_write_offs.project_id = ? AND
			(nullif(?, '') IS NULL OR invoice_write_offs.write_off_type = ?) 
    `,
		filter.ProjectID,
		filter.WriteOffType, filter.WriteOffType,
	).Scan(&count).Error
	return count, err
}

func (repo *invoiceWriteOffRepository) GetMaterialsForEdit(id uint) ([]dto.InvoiceWriteOffMaterialsForEdit, error) {
	result := []dto.InvoiceWriteOffMaterialsForEdit{}
	err := repo.db.Raw(`
    SELECT 
      materials.id as material_id,
      materials.name as material_name,
      materials.unit as unit,
      invoice_materials.amount as amount,
      material_costs.id  as material_cost_id,
      material_costs.cost_m19 as material_cost,
      invoice_materials.notes as  notes,
      materials.has_serial_number as has_serial_number
    FROM invoice_materials
    INNER JOIN material_costs ON invoice_materials.material_cost_id = material_costs.id
    INNER JOIN materials ON material_costs.material_id = materials.id
    WHERE
      invoice_materials.invoice_type='writeoff' AND
      invoice_materials.invoice_id = ?;
    `, id).Scan(&result).Error

	return result, err
}

func (repo *invoiceWriteOffRepository) Confirmation(data dto.InvoiceWriteOffConfirmationData) error {
	return repo.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.InvoiceWriteOff{}).Select("*").Where("id = ?", data.InvoiceWriteOff.ID).Updates(&data.InvoiceWriteOff).Error; err != nil {
			return err
		}

		if err := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "id"}},
			DoUpdates: clause.AssignmentColumns([]string{"amount"}),
		}).Create(&data.MaterialsInLocation).Error; err != nil {
			return err
		}

		if err := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "id"}},
			DoUpdates: clause.AssignmentColumns([]string{"amount"}),
		}).Create(&data.MaterialsInWriteOff).Error; err != nil {
			return err
		}

		return nil
	})
}
