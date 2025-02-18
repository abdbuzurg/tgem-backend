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
	GetPaginatedFiltered(page, limit int, filter dto.InvoiceInputSearchParameters) ([]dto.InvoiceInputPaginated, error)
	GetByID(id uint) (model.InvoiceInput, error)
	Create(data dto.InvoiceInputCreateQueryData) (model.InvoiceInput, error)
	Update(data dto.InvoiceInputCreateQueryData) (model.InvoiceInput, error)
	Delete(id uint) error
	Count(filter dto.InvoiceInputSearchParameters) (int64, error)
	UniqueCode(projectID uint) ([]dto.DataForSelect[string], error)
	UniqueWarehouseManager(projectID uint) ([]dto.DataForSelect[uint], error)
	UniqueReleased(projectID uint) ([]dto.DataForSelect[uint], error)
	ReportFilterData(filter dto.InvoiceInputReportFilterRequest) ([]dto.InvoiceInputReportData, error)
	Confirmation(data dto.InvoiceInputConfirmationQueryData) error
	GetMaterialsForEdit(id uint) ([]dto.InvoiceInputMaterialForEdit, error)
	GetSerialNumbersForEdit(invoiceID uint, materialCostID uint) ([]string, error)
	Import(data []dto.InvoiceInputImportData) error
	GetAllDeliveryCodes(projectID uint) ([]string, error)
	GetAllWarehouseManagers(projectID uint) ([]dto.DataForSelect[uint], error)
	GetAllReleasedWorkers(projectID uint) ([]dto.DataForSelect[uint], error)
	GetAllMaterialsThatAreInInvoiceInput(projectID uint) ([]dto.DataForSelect[uint], error)
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

func (repo *invoiceInputRespository) GetPaginatedFiltered(page, limit int, filter dto.InvoiceInputSearchParameters) ([]dto.InvoiceInputPaginated, error) {
	data := []dto.InvoiceInputPaginated{}
	var err error
	if len(filter.Materials) != 0 {
		err = repo.db.Raw(`
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
        invoice_inputs.id IN (
          SELECT invoice_materials.invoice_id
          FROM invoice_materials
          WHERE 
            invoice_materials.project_id = ? AND
            invoice_materials.invoice_type = 'input' AND
            invoice_materials.material_cost_id IN (
              SELECT material_costs.id
              FROM material_costs
              WHERE material_costs.material_id IN ?
            ) 
      )
      ORDER BY invoice_inputs.id DESC 
      LIMIT ? 
      OFFSET ?;
      `, filter.ProjectID,
			filter.ProjectID,
			filter.Materials,
			limit, (page-1)*limit,
		).Scan(&data).Error
	} else {
		dateFrom := filter.DateFrom.String()
		dateFrom = dateFrom[:len(dateFrom)-10]
		dateTo := filter.DateTo.String()
		dateTo = dateTo[:len(dateTo)-10]
		err = repo.db.
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
        (nullif(?, '') IS NULL OR delivery_code = ?) AND 
        (nullif(?, '0001-01-01 00:00:00') IS NULL OR ? <= invoice_inputs.date_of_invoice) AND 
        (nullif(?, '0001-01-01 00:00:00') IS NULL OR invoice_inputs.date_of_invoice <= ?)
      ORDER BY invoice_inputs.id DESC 
      LIMIT ? 
      OFFSET ?;
    `,
				filter.ProjectID,
				filter.WarehouseManagerWorkerID, filter.WarehouseManagerWorkerID,
				filter.ReleasedWorkerID, filter.ReleasedWorkerID,
				filter.DeliveryCode, filter.DeliveryCode,
				dateFrom, dateFrom,
				dateTo, dateTo,
				limit, (page-1)*limit,
			).
			Scan(&data).Error
	}

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

		err := tx.Exec(`
      UPDATE invoice_counts
      SET count = count + 1
      WHERE
        invoice_type = 'input' AND
        project_id = ?
      `, result.ProjectID).Error
		if err != nil {
			return err
		}

		return nil

	})
	return result, err
}

func (repo *invoiceInputRespository) Update(data dto.InvoiceInputCreateQueryData) (model.InvoiceInput, error) {
	result := data.InvoiceData
	err := repo.db.Transaction(func(tx *gorm.DB) error {

		err := tx.Model(&result).Select("*").Where("id = ?", result.ID).Updates(&result).Error
		if err != nil {
			return err
		}

		err = tx.Exec(`
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

		err = tx.Exec(`
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
	return repo.db.Transaction(func(tx *gorm.DB) error {

		if err := tx.Delete(&model.InvoiceInput{}, "id = ?", id).Error; err != nil {
			return err
		}

		if err := tx.Delete(&model.InvoiceMaterials{}, "invoice_type = 'input' AND invoice_id = ?", id).Error; err != nil {
			return err
		}

		return nil
	})
}

func (repo *invoiceInputRespository) Count(filter dto.InvoiceInputSearchParameters) (int64, error) {
	var count int64
	var err error
	if len(filter.Materials) > 0 {
		err = repo.db.Raw(`
      SELECT COUNT(*) 
      FROM invoice_inputs
      INNER JOIN workers AS warehouse_manager ON warehouse_manager.id = invoice_inputs.warehouse_manager_worker_id
      INNER JOIN workers AS released ON released.id = invoice_inputs.released_worker_id
      WHERE 
        invoice_inputs.project_id = ? AND
        invoice_inputs.id IN (
          SELECT invoice_materials.invoice_id
          FROM invoice_materials
          WHERE 
            invoice_materials.project_id = ? AND
            invoice_materials.invoice_type = 'input' AND
            invoice_materials.material_cost_id IN (
              SELECT material_costs.id
              FROM material_costs
              WHERE material_costs.material_id IN ?
            ) 
      )
      `, filter.ProjectID,
			filter.ProjectID,
			filter.Materials,
		).Scan(&count).Error
	} else {
		dateFrom := filter.DateFrom.String()
		dateFrom = dateFrom[:len(dateFrom)-10]
		dateTo := filter.DateTo.String()
		dateTo = dateTo[:len(dateTo)-10]
		err = repo.db.
			Raw(`
      SELECT COUNT(*) 
      FROM invoice_inputs
      INNER JOIN workers AS warehouse_manager ON warehouse_manager.id = invoice_inputs.warehouse_manager_worker_id
      INNER JOIN workers AS released ON released.id = invoice_inputs.released_worker_id
      WHERE 
        invoice_inputs.project_id = ? AND
        (nullif(?, 0) IS NULL OR warehouse_manager_worker_id = ?) AND
        (nullif(?, 0) IS NULL OR released_worker_id = ?) AND
        (nullif(?, '') IS NULL OR delivery_code = ?) AND 
        (nullif(?, '0001-01-01 00:00:00') IS NULL OR ? <= invoice_inputs.date_of_invoice) AND 
        (nullif(?, '0001-01-01 00:00:00') IS NULL OR invoice_inputs.date_of_invoice <= ?)
    `,
				filter.ProjectID,
				filter.WarehouseManagerWorkerID, filter.WarehouseManagerWorkerID,
				filter.ReleasedWorkerID, filter.ReleasedWorkerID,
				filter.DeliveryCode, filter.DeliveryCode,
				dateFrom, dateFrom,
				dateTo, dateTo,
			).
      Scan(&count).Error
	}
	return count, err
}

func (repo *invoiceInputRespository) UniqueCode(projectID uint) ([]dto.DataForSelect[string], error) {
	data := []dto.DataForSelect[string]{}
	err := repo.db.Raw(`
      SELECT 
        delivery_code as "label",
        delivery_code as "value"
      FROM invoice_inputs
      WHERE project_id = ?
      ORDER BY id DESC;
    `, projectID).Scan(&data).Error
	return data, err
}

func (repo *invoiceInputRespository) UniqueWarehouseManager(projectID uint) ([]dto.DataForSelect[uint], error) {
	data := []dto.DataForSelect[uint]{}
	err := repo.db.Raw(`
      SELECT 
        workers.id as "value",
        workers.name as "label"
      FROM workers
      WHERE workers.id IN (
        SELECT DISTINCT(invoice_inputs.warehouse_manager_worker_id)
        FROM invoice_inputs
        WHERE invoice_inputs.project_id = ?
      )
    `, projectID).Scan(&data).Error

	return data, err
}

func (repo *invoiceInputRespository) UniqueReleased(projectID uint) ([]dto.DataForSelect[uint], error) {
	data := []dto.DataForSelect[uint]{}
	err := repo.db.Raw(`
      SELECT 
        workers.id as "value",
        workers.name as "label"
      FROM workers
      WHERE workers.id IN (
        SELECT DISTINCT(invoice_inputs.released_worker_id)
        FROM invoice_inputs
        WHERE invoice_inputs.project_id = ?
      )
    `, projectID).Scan(&data).Error
	return data, err
}

func (repo *invoiceInputRespository) ReportFilterData(filter dto.InvoiceInputReportFilterRequest) ([]dto.InvoiceInputReportData, error) {
	data := []dto.InvoiceInputReportData{}
	dateFrom := filter.DateFrom.String()
	dateFrom = dateFrom[:len(dateFrom)-10]
	dateTo := filter.DateTo.String()
	dateTo = dateTo[:len(dateTo)-10]
	err := repo.
		db.
		Raw(`
      SELECT 
        invoice_inputs.id as id,
        warehouse_manager.name as warehouse_manager_name,
        released.name as released_name,
        invoice_inputs.delivery_code as delivery_code,
        invoice_inputs.notes as notes,
        invoice_inputs.date_of_invoice as date_of_invoice
      FROM invoice_inputs 
      INNER JOIN workers as warehouse_manager ON warehouse_manager.id = invoice_inputs.warehouse_manager_worker_id
      INNER JOIN workers as released ON released.id = invoice_inputs.released_worker_id
      WHERE 
        invoice_inputs.project_id = ? AND
        (nullif(?, '') IS NULL OR invoice_inputs.delivery_code = ?) AND
        (nullif(?, 0) IS NULL OR invoice_inputs.released_worker_id = ?) AND
        (nullif(?, 0) IS NULL OR invoice_inputs.warehouse_manager_worker_id = ?) AND
        (nullif(?, '0001-01-01 00:00:00') IS NULL OR ? <= invoice_inputs.date_of_invoice) AND 
        (nullif(?, '0001-01-01 00:00:00') IS NULL OR invoice_inputs.date_of_invoice <= ?) ORDER BY invoice_inputs.id DESC		`,
			filter.ProjectID,
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

func (repo *invoiceInputRespository) Import(data []dto.InvoiceInputImportData) error {
	return repo.db.Transaction(func(tx *gorm.DB) error {
		for index, invoice := range data {
			invoiceInput := invoice.Details
			if err := tx.Create(&invoiceInput).Error; err != nil {
				return err
			}

			for subIndex := range invoice.Items {
				data[index].Items[subIndex].InvoiceID = invoiceInput.ID
			}

			if err := tx.CreateInBatches(&data[index].Items, 15).Error; err != nil {
				return err
			}
		}

		err := tx.Exec(`
      UPDATE invoice_counts
      SET count = count + ?
      WHERE
        invoice_type = 'input' AND
        project_id = ?
      `, len(data), data[0].Details.ProjectID).Error
		if err != nil {
			return err
		}

		return nil
	})
}

func (repo *invoiceInputRespository) GetAllDeliveryCodes(projectID uint) ([]string, error) {
	result := []string{}
	err := repo.db.Raw(`SELECT delivery_code FROM invoice_inputs WHERE project_id = ?`, projectID).Scan(&result).Error
	return result, err
}

func (repo *invoiceInputRespository) GetAllWarehouseManagers(projectID uint) ([]dto.DataForSelect[uint], error) {
	result := []dto.DataForSelect[uint]{}
	err := repo.db.Raw(`
    SELECT DISTINCT 
      workers.id as "value", 
      workers.name as "label"
    FROM invoice_inputs
    INNER JOIN workers ON workers.id = invoice_inputs.warehouse_manager_worker_id
    WHERE invoice_inputs.project_id = ?;
    `, projectID).Scan(&result).Error

	return result, err
}

func (repo *invoiceInputRespository) GetAllReleasedWorkers(projectID uint) ([]dto.DataForSelect[uint], error) {
	result := []dto.DataForSelect[uint]{}
	err := repo.db.Raw(`
    SELECT DISTINCT 
      workers.id as "value", 
      workers.name as "label"
    FROM invoice_inputs
    INNER JOIN workers ON workers.id = invoice_inputs.released_worker_id
    WHERE invoice_inputs.project_id = ?;
    `, projectID).Scan(&result).Error
	return result, err
}

func (repo *invoiceInputRespository) GetAllMaterialsThatAreInInvoiceInput(projectID uint) ([]dto.DataForSelect[uint], error) {
	result := []dto.DataForSelect[uint]{}
	err := repo.db.Raw(`
    SELECT DISTINCT
      materials.id as "value",
      materials.name as "label"
    FROM invoice_materials
    INNER JOIN material_costs ON invoice_materials.material_cost_id = material_costs.id
    INNER JOIN materials ON materials.id = material_costs.material_id
    WHERE 
      invoice_materials.project_id = ? AND
      invoice_materials.invoice_type = 'input'
    `, projectID).Scan(&result).Error

	return result, err
}
