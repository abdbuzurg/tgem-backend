package repository

import (
	"backend-v2/internal/dto"
	"backend-v2/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type invoiceInputRespository struct {
	db *gorm.DB
}

func InitInvoiceInputRepository(db *gorm.DB) IInovoiceInputRepository {
	return &invoiceInputRespository{
		db: db,
	}
}

type IInovoiceInputRepository interface {
	GetAll() ([]model.InvoiceInput, error)
	GetPaginated(page, limit int) ([]model.InvoiceInput, error)
	GetPaginatedFiltered(page, limit int, filter model.InvoiceInput) ([]dto.InvoiceInputPaginated, error)
	GetByID(id uint) (model.InvoiceInput, error)
	Create(data dto.InvoiceInputCreateQueryData) (model.InvoiceInput, error)
	Update(data dto.InvoiceInputCreateQueryData) (model.InvoiceInput, error)
	Delete(id uint) error
	Count(projectID uint) (int64, error)
	UniqueCode(projectID uint) ([]string, error)
	UniqueWarehouseManager(projectID uint) ([]string, error)
	UniqueReleased(projectID uint) ([]string, error)
	ReportFilterData(filter dto.InvoiceInputReportFilter, projectID uint) ([]model.InvoiceInput, error)
	Confirmation(data dto.InvoiceInputConfirmationQueryData) error
	GetMaterialsForEdit(id uint) ([]dto.InvoiceInputMaterialForEdit, error)
	GetSerialNumbersForEdit(invoiceID uint, materialCostID uint) ([]string, error)
}

func (repo *invoiceInputRespository) GetAll() ([]model.InvoiceInput, error) {
	data := []model.InvoiceInput{}
	err := repo.db.Order("id desc").Find(&data).Error
	return data, err
}

func (repo *invoiceInputRespository) GetPaginated(page, limit int) ([]model.InvoiceInput, error) {
	data := []model.InvoiceInput{}
	err := repo.db.Order("id desc").Offset((page - 1) * limit).Limit(limit).Find(&data).Error
	return data, err
}

func (repo *invoiceInputRespository) GetPaginatedFiltered(page, limit int, filter model.InvoiceInput) ([]dto.InvoiceInputPaginated, error) {
	data := []dto.InvoiceInputPaginated{}
	err := repo.db.
		Raw(`
      SELECT 
        invoice_inputs.id as id,
        invoice_inputs.confirmed as confirmation,
        invoice_inputs.delivery_code as delivery_code,
        warehouse_manager.name as warehouse_manager_name,
        released.name as released_name,
        invoice_inputs.date_of_invoice as date_of_invoice
      FROM invoice_inputs
        INNER JOIN workers AS warehouse_manager ON warehouse_manager.id = invoice_inputs.warehouse_manager_worker_id
        INNER JOIN workers AS released ON released.id = invoice_inputs.released_worker_id
      WHERE 
        invoice_inputs.project_id = ? AND
        (nullif(?, 0) IS NULL OR warehouse_manager_worker_id = ?) AND
        (nullif(?, 0) IS NULL OR released_worker_id = ?) AND
        (nullif(?, '') IS NULL OR delivery_code = ?) ORDER BY invoice_inputs.id DESC LIMIT ? OFFSET ?;
    `,
			filter.ProjectID,
			filter.WarehouseManagerWorkerID, filter.WarehouseManagerWorkerID,
			filter.ReleasedWorkerID, filter.ReleasedWorkerID,
			filter.DeliveryCode, filter.DeliveryCode,
			limit, (page-1)*limit,
		).
		Scan(&data).Error

	return data, err
}

func (repo *invoiceInputRespository) GetByID(id uint) (model.InvoiceInput, error) {
	data := model.InvoiceInput{}
	err := repo.db.Find(&data, "id = ?", id).Error
	return data, err
}

func (repo *invoiceInputRespository) Create(data dto.InvoiceInputCreateQueryData) (model.InvoiceInput, error) {
	result := data.InvoiceData
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

		serialNumbers := data.SerialNumbers
		if err := tx.CreateInBatches(&serialNumbers, 15).Error; err != nil {
			return err
		}

		for index := range data.SerialNumberMovement {
			data.SerialNumberMovement[index].SerialNumberID = serialNumbers[index].ID
			data.SerialNumberMovement[index].InvoiceID = result.ID
		}

		if err := tx.CreateInBatches(&data.SerialNumberMovement, 15).Error; err != nil {
			return err
		}

		return nil

	})
	return result, err
}

func (repo *invoiceInputRespository) Update(data dto.InvoiceInputCreateQueryData) (model.InvoiceInput, error) {
	result := data.InvoiceData
	err := repo.db.Transaction(func(tx *gorm.DB) error {

		err := repo.db.Model(&result).Select("*").Where("id = ?", result.ID).Updates(&result).Error
		if err != nil {
			return err
		}

		err = repo.db.Exec(`
      DELETE FROM invoice_materials
      WHERE invoice_type = 'input' AND invoice_id = ?
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

		err = repo.db.Exec(`
      DELETE FROM serial_number_movements
      WHERE invoice_type = 'input' AND invoice_id = ?
      `, result.ID).Error
		if err != nil {
			return err
		}

		serialNumbers := data.SerialNumbers
		if err := tx.CreateInBatches(&serialNumbers, 15).Error; err != nil {
			return err
		}

		for index := range data.SerialNumberMovement {
			data.SerialNumberMovement[index].SerialNumberID = serialNumbers[index].ID
			data.SerialNumberMovement[index].InvoiceID = result.ID
		}

		if err := tx.CreateInBatches(&data.SerialNumberMovement, 15).Error; err != nil {
			return err
		}

		return nil
	})

	return result, err
}

func (repo *invoiceInputRespository) Delete(id uint) error {
	return repo.db.Delete(&model.InvoiceInput{}, "id = ?", id).Error
}

func (repo *invoiceInputRespository) Count(projectID uint) (int64, error) {
	var count int64
	err := repo.db.Raw("SELECT COUNT(*) FROM invoice_inputs WHERE project_id = ?", projectID).Scan(&count).Error
	return count, err
}

func (repo *invoiceInputRespository) UniqueCode(projectID uint) ([]string, error) {
	data := []string{}
	err := repo.db.Raw("SELECT DISTINCT delivery_code FROM invoice_inputs WHERE project_id = ?", projectID).Scan(&data).Error
	return data, err
}

func (repo *invoiceInputRespository) UniqueWarehouseManager(projectID uint) ([]string, error) {
	data := []string{}
	err := repo.db.Raw("SELECT DISTINCT warehouse_manager_worker_id FROM invoice_inputs WHERE project_id = ?", projectID).Scan(&data).Error
	return data, err
}

func (repo *invoiceInputRespository) UniqueReleased(projectID uint) ([]string, error) {
	data := []string{}
	err := repo.db.Raw("SELECT DISTINCT released_worker_id FROM invoice_inputs WHERE project_id = ?", projectID).Scan(&data).Error
	return data, err
}

func (repo *invoiceInputRespository) ReportFilterData(filter dto.InvoiceInputReportFilter, projectID uint) ([]model.InvoiceInput, error) {
	data := []model.InvoiceInput{}
	dateFrom := filter.DateFrom.String()
	dateFrom = dateFrom[:len(dateFrom)-10]
	dateTo := filter.DateTo.String()
	dateTo = dateTo[:len(dateTo)-10]
	err := repo.
		db.
		Raw(`SELECT * FROM invoice_inputs WHERE project_id = ? AND
			(nullif(?, '') IS NULL OR delivery_code = ?) AND
			(nullif(?, 0) IS NULL OR released_worker_id = ?) AND
			(nullif(?, 0) IS NULL OR warehouse_manager_worker_id = ?) AND
			(nullif(?, '0001-01-01 00:00:00') IS NULL OR ? <= date_of_invoice) AND 
			(nullif(?, '0001-01-01 00:00:00') IS NULL OR date_of_invoice <= ?) ORDER BY id DESC
		`,
			projectID,
			filter.Code, filter.Code,
			filter.ReleasedID, filter.ReleasedID,
			filter.WarehouseManagerID, filter.WarehouseManagerID,
			dateFrom, dateFrom,
			dateTo, dateTo).
		Scan(&data).Error

	return data, err
}

func (repo *invoiceInputRespository) Confirmation(data dto.InvoiceInputConfirmationQueryData) error {
	return repo.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.InvoiceInput{}).Select("*").Where("id = ?", data.InvoiceData.ID).Updates(&data.InvoiceData).Error; err != nil {
			return err
		}

		if len(data.ToBeUpdatedMaterials) != 0 {
			if err := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "id"}},
				DoUpdates: clause.AssignmentColumns([]string{"amount"}),
			}).Create(&data.ToBeUpdatedMaterials).Error; err != nil {
				return err
			}
		}

		if len(data.ToBeCreatedMaterials) != 0 {
			if err := tx.Create(&data.ToBeCreatedMaterials).Error; err != nil {
				return err
			}
		}

		if err := tx.Exec(`
      UPDATE 
        serial_number_movements 
      SET confirmation = true 
      WHERE 
        invoice_id = ? AND 
        invoice_type = 'input'
    `, data.InvoiceData.ID).Error; err != nil {
			return err
		}

		if err := tx.CreateInBatches(&data.SerialNumbers, 15).Error; err != nil {
			return err
		}

		return nil
	})
}

func (repo *invoiceInputRespository) GetMaterialsForEdit(id uint) ([]dto.InvoiceInputMaterialForEdit, error) {
	result := []dto.InvoiceInputMaterialForEdit{}
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
      invoice_materials.invoice_type='input' AND
      invoice_materials.invoice_id = ?;
    `, id).Scan(&result).Error

	return result, err
}

func (repo *invoiceInputRespository) GetSerialNumbersForEdit(invoiceID uint, materialCostID uint) ([]string, error) {
	result := []string{}
	err := repo.db.Raw(`
    SELECT 
      serial_numbers.code
    FROM serial_number_movements
    INNER JOIN serial_numbers ON serial_number_movements.serial_number_id = serial_numbers.id
    WHERE
      serial_number_movements.invoice_type = 'input' AND
      serial_number_movements.invoice_id = ? AND
      material_costs.id = ?;
    `, invoiceID, materialCostID).Scan(&result).Error

	return result, err
}
