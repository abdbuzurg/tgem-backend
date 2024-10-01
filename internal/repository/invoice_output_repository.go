package repository

import (
	"backend-v2/internal/dto"
	"backend-v2/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type invoiceOutputRepository struct {
	db *gorm.DB
}

func InitInvoiceOutputRepository(db *gorm.DB) IInvoiceOutputRepository {
	return &invoiceOutputRepository{
		db: db,
	}
}

type IInvoiceOutputRepository interface {
	GetAll() ([]model.InvoiceOutput, error)
	GetPaginated(page, limit int) ([]model.InvoiceOutput, error)
	GetPaginatedFiltered(page, limit int, filter model.InvoiceOutput) ([]dto.InvoiceOutputPaginated, error)
	GetByID(id uint) (model.InvoiceOutput, error)
	GetUnconfirmedByObjectInvoices() ([]model.InvoiceOutput, error)
	GetAvailableMaterialsInWarehouse(projectID uint) ([]dto.AvailableMaterialsInWarehouse, error)
	GetDataForExcel(id uint) (dto.InvoiceOutputDataForExcelQueryResult, error)
	Create(data dto.InvoiceOutputCreateQueryData) (model.InvoiceOutput, error)
	Update(data dto.InvoiceOutputCreateQueryData) (model.InvoiceOutput, error)
	Delete(id uint) error
	Count(projectID uint) (int64, error)
	UniqueCode(projectID uint) ([]dto.DataForSelect[string], error)
	UniqueWarehouseManager(projectID uint) ([]dto.DataForSelect[uint], error)
	UniqueRecieved(projectID uint) ([]dto.DataForSelect[uint], error)
	UniqueDistrict(projectID uint) ([]dto.DataForSelect[uint], error)
	UniqueTeam(projectID uint) ([]dto.DataForSelect[uint], error)
	ReportFilterData(filter dto.InvoiceOutputReportFilterRequest) ([]dto.InvoiceOutputDataForReport, error)
	Confirmation(data dto.InvoiceOutputConfirmationQueryData) error
	GetMaterialsForEdit(id uint) ([]dto.InvoiceOutputMaterialsForEdit, error)
	GetMaterialDataForReport(invoiceID uint) ([]dto.InvoiceOutputMaterialDataForReport, error)
	Import(data []dto.InvoiceOutputImportData) error
}

func (repo *invoiceOutputRepository) GetAll() ([]model.InvoiceOutput, error) {
	data := []model.InvoiceOutput{}
	err := repo.db.Order("id desc").Find(&data).Error
	return data, err
}

func (repo *invoiceOutputRepository) GetPaginated(page, limit int) ([]model.InvoiceOutput, error) {
	data := []model.InvoiceOutput{}
	err := repo.db.Order("id desc").Offset((page - 1) * limit).Limit(limit).Find(&data).Error
	return data, err
}

func (repo *invoiceOutputRepository) GetPaginatedFiltered(page, limit int, filter model.InvoiceOutput) ([]dto.InvoiceOutputPaginated, error) {
	data := []dto.InvoiceOutputPaginated{}
	err := repo.db.
		Raw(`
      SELECT 
        invoice_outputs.id as id,
        invoice_outputs.delivery_code as delivery_code,
        districts.name as district_name,
        districts.id as district_id,
        teams.number as team_name,
        teams.id as team_id,
        warehouse_manager.id as warehouse_manager_id,
        warehouse_manager.name as warehouse_manager_name,
        released.name as released_name,
        recipient.id as recipient_id,
        recipient.name as recipient_name,
        invoice_outputs.date_of_invoice as date_of_invoice,
        invoice_outputs.confirmation as confirmation,
        invoice_outputs.notes as notes
      FROM invoice_outputs
        INNER JOIN districts ON districts.id = invoice_outputs.district_id
        INNER JOIN teams ON teams.id = invoice_outputs.team_id
        INNER JOIN workers AS warehouse_manager ON warehouse_manager.id = invoice_outputs.warehouse_manager_worker_id
        INNER JOIN workers AS released ON released.id = invoice_outputs.released_worker_id
        INNER JOIN workers AS recipient ON recipient.id = invoice_outputs.recipient_worker_id
      WHERE
        invoice_outputs.project_id = ? AND
        (nullif(?, 0) IS NULL OR invoice_outputs.district_id = ?) AND
        (nullif(?, 0) IS NULL OR invoice_outputs.warehouse_manager_worker_id = ?) AND
        (nullif(?, 0) IS NULL OR invoice_outputs.released_worker_id = ?) AND
        (nullif(?, 0) IS NULL OR invoice_outputs.recipient_worker_id = ?) AND
        (nullif(?, 0) IS NULL OR invoice_outputs.team_id = ?) AND
        (nullif(?, '') IS NULL OR invoice_outputs.delivery_code = ?) ORDER BY invoice_outputs.id DESC LIMIT ? OFFSET ?;

      `,
			filter.ProjectID,
			filter.DistrictID, filter.DistrictID,
			filter.WarehouseManagerWorkerID, filter.WarehouseManagerWorkerID,
			filter.ReleasedWorkerID, filter.ReleasedWorkerID,
			filter.RecipientWorkerID, filter.RecipientWorkerID,
			filter.TeamID, filter.TeamID,
			filter.DeliveryCode, filter.DeliveryCode,
			limit, (page-1)*limit,
		).
		Scan(&data).Error

	return data, err
}

func (repo *invoiceOutputRepository) GetByID(id uint) (model.InvoiceOutput, error) {
	data := model.InvoiceOutput{}
	err := repo.db.Find(&data, "id = ?", id).Error
	return data, err
}

func (repo *invoiceOutputRepository) Create(data dto.InvoiceOutputCreateQueryData) (model.InvoiceOutput, error) {
	result := data.Invoice
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

		for index := range data.SerialNumberMovements {
			data.SerialNumberMovements[index].InvoiceID = result.ID
		}

		if err := tx.CreateInBatches(&data.SerialNumberMovements, 15).Error; err != nil {
			return err
		}

		err := tx.Exec(`
      UPDATE invoice_counts
      SET count = count + 1
      WHERE
        invoice_type = 'output' AND
        project_id = ?
      `, result.ProjectID).Error
		if err != nil {
			return err
		}

		return nil

	})
	return result, err
}

func (repo *invoiceOutputRepository) Update(data dto.InvoiceOutputCreateQueryData) (model.InvoiceOutput, error) {
	result := data.Invoice
	err := repo.db.Transaction(func(tx *gorm.DB) error {
		err := tx.Model(&model.InvoiceOutput{}).Select("*").Where("id = ?", result.ID).Updates(&result).Error
		if err != nil {
			return err
		}

		if err = tx.Delete(model.InvoiceMaterials{}, "invoice_id = ? AND invoice_type='output'", result.ID).Error; err != nil {
			return nil
		}

		for index := range data.InvoiceMaterials {
			data.InvoiceMaterials[index].InvoiceID = result.ID
		}

		if err := tx.CreateInBatches(&data.InvoiceMaterials, 15).Error; err != nil {
			return err
		}

		if err = tx.Delete(model.SerialNumberMovement{}, "invoice_id = ? AND invoice_type='output'", result.ID).Error; err != nil {
			return err
		}

		for index := range data.SerialNumberMovements {
			data.SerialNumberMovements[index].InvoiceID = result.ID
		}

		if err := tx.CreateInBatches(&data.SerialNumberMovements, 15).Error; err != nil {
			return err
		}

		return nil
	})

	return result, err
}

func (repo *invoiceOutputRepository) Delete(id uint) error {
	return repo.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Delete(&model.InvoiceOutput{}, "id = ?", id).Error; err != nil {
			return err
		}

		if err := tx.Delete(&model.InvoiceMaterials{}, "invoice_type = 'output' AND invoice_id = ?", id).Error; err != nil {
			return err
		}

		return nil
	})
}

func (repo *invoiceOutputRepository) Count(projectID uint) (int64, error) {
	var count int64
	err := repo.db.Raw("SELECT COUNT(*) FROM invoice_outputs WHERE project_id = ?", projectID).Scan(&count).Error
	return count, err
}

func (repo *invoiceOutputRepository) GetUnconfirmedByObjectInvoices() ([]model.InvoiceOutput, error) {
	data := []model.InvoiceOutput{}
	err := repo.db.Find(&data, "confirmation = TRUE AND object_confirmation = FALSE ORDER BY id DESC").Error
	return data, err
}

func (repo *invoiceOutputRepository) UniqueCode(projectID uint) ([]dto.DataForSelect[string], error) {
	data := []dto.DataForSelect[string]{}
	err := repo.db.Raw(`
      SELECT 
        delivery_code as "label",
        delivery_code as "value"
      FROM invoice_outputs
      WHERE project_id = ?
      ORDER BY id DESC;
    `, projectID).Scan(&data).Error

	return data, err
}

func (repo *invoiceOutputRepository) UniqueWarehouseManager(projectID uint) ([]dto.DataForSelect[uint], error) {
	data := []dto.DataForSelect[uint]{}
	err := repo.db.Raw(`
      SELECT 
        workers.id as "value",
        workers.name as "label"
      FROM workers
      WHERE workers.id IN (
        SELECT DISTINCT(invoice_outputs.warehouse_manager_worker_id)
        FROM invoice_outputs
        WHERE invoice_outputs.project_id = ?
      )
    `, projectID).Scan(&data).Error
	return data, err
}

func (repo *invoiceOutputRepository) UniqueRecieved(projectID uint) ([]dto.DataForSelect[uint], error) {
	data := []dto.DataForSelect[uint]{}
	err := repo.db.Raw(`
      SELECT 
        workers.id as "value",
        workers.name as "label"
      FROM workers
      WHERE workers.id IN (
        SELECT DISTINCT(invoice_outputs.recipient_worker_id)
        FROM invoice_outputs
        WHERE invoice_outputs.project_id = ?
      )
    `, projectID).Scan(&data).Error
	return data, err
}

func (repo *invoiceOutputRepository) UniqueDistrict(projectID uint) ([]dto.DataForSelect[uint], error) {
	data := []dto.DataForSelect[uint]{}
	err := repo.db.Raw(`
      SELECT 
        districts.id as "value",
        districts.name as "label"
      FROM districts
      WHERE districts.id IN (
        SELECT DISTINCT(invoice_outputs.district_id)
        FROM invoice_outputs
        WHERE invoice_outputs.project_id = ?
      )
    `, projectID).Scan(&data).Error
	return data, err
}

func (repo *invoiceOutputRepository) UniqueTeam(projectID uint) ([]dto.DataForSelect[uint], error) {
	data := []dto.DataForSelect[uint]{}
	err := repo.db.Raw(`
      SELECT 
        teams.id as "value",
        CONCAT(teams.number, ' (', workers.name, ')') as "label"
      FROM teams
      INNER JOIN team_leaders ON team_leaders.team_id = teams.id
      INNER JOIN workers ON workers.id = team_leaders.leader_worker_id
      WHERE teams.id IN (
        SELECT DISTINCT(invoice_outputs.team_id)
        FROM invoice_outputs
        WHERE invoice_outputs.project_id = ?
      )
    `, projectID).Scan(&data).Error
	return data, err
}

func (repo *invoiceOutputRepository) ReportFilterData(filter dto.InvoiceOutputReportFilterRequest) ([]dto.InvoiceOutputDataForReport, error) {
	data := []dto.InvoiceOutputDataForReport{}
	dateFrom := filter.DateFrom.String()
	dateFrom = dateFrom[:len(dateFrom)-10]
	dateTo := filter.DateTo.String()
	dateTo = dateTo[:len(dateTo)-10]
	err := repo.
		db.
		Raw(`
      SELECT 
        invoice_outputs.id as id,
        invoice_outputs.delivery_code as delivery_code,
        warehouse_manager.name as warehouse_manager_name,
        recipient_worker.name as recipient_name,
        teams.number as team_number,
        leader_worker.name as team_leader_name,
        invoice_outputs.date_of_invoice as date_of_invoice
      FROM invoice_outputs
      INNER JOIN workers as warehouse_manager ON warehouse_manager.id = invoice_outputs.warehouse_manager_worker_id
      INNER JOIN workers as recipient_worker ON recipient_worker.id = invoice_outputs.recipient_worker_id
      INNER JOIN teams ON teams.id = invoice_outputs.team_id 
      INNER JOIN team_leaders ON teams.id = team_leaders.team_id
      INNER JOIN workers as leader_worker ON leader_worker.id = team_leaders.leader_worker_id
      WHERE
        invoice_outputs.project_id = ? AND
        invoice_outputs.confirmation = true AND
        (nullif(?, '') IS NULL OR invoice_outputs.delivery_code = ?) AND
        (nullif(?, 0) IS NULL OR invoice_outputs.recipient_worker_id = ?) AND
        (nullif(?, 0) IS NULL OR invoice_outputs.warehouse_manager_worker_id = ?) AND
        (nullif(?, 0) IS NULL OR invoice_outputs.team_id = ?) AND
        (nullif(?, '0001-01-01 00:00:00') IS NULL OR ? <= invoice_outputs.date_of_invoice) AND 
        (nullif(?, '0001-01-01 00:00:00') IS NULL OR invoice_outputs.date_of_invoice <= ?) ORDER BY invoice_outputs.id DESC
		`,
			filter.ProjectID,
			filter.Code, filter.Code,
			filter.ReceivedID, filter.ReceivedID,
			filter.WarehouseManagerID, filter.WarehouseManagerID,
			filter.TeamID, filter.TeamID,
			dateFrom, dateFrom,
			dateTo, dateTo).
		Scan(&data).Error

	return data, err
}

func (repo *invoiceOutputRepository) AmountOfMaterialInALocation(materialID uint) (float64, error) {
	return 0, nil
}

func (repo *invoiceOutputRepository) GetAvailableMaterialsInWarehouse(projectID uint) ([]dto.AvailableMaterialsInWarehouse, error) {
	data := []dto.AvailableMaterialsInWarehouse{}
	err := repo.db.Raw(`
    SELECT 
      materials.id AS id,
      materials.name AS name,
      materials.unit AS unit,
      materials.has_serial_number as has_serial_number,
      material_locations.amount as amount
    FROM material_locations
    INNER JOIN material_costs ON material_costs.id = material_locations.material_cost_id
    INNER JOIN materials ON materials.id = material_costs.material_id
    WHERE
      material_locations.project_id = ? AND
      material_locations.location_type = 'warehouse' AND
      material_locations.amount > 0
    ORDER BY materials.name
    `, projectID).Scan(&data).Error

	return data, err
}

func (repo *invoiceOutputRepository) Confirmation(data dto.InvoiceOutputConfirmationQueryData) error {
	return repo.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.InvoiceOutput{}).Select("*").Where("id = ?", data.InvoiceData.ID).Updates(&data.InvoiceData).Error; err != nil {
			return err
		}

		if err := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "id"}},
			DoUpdates: clause.AssignmentColumns([]string{"amount"}),
		}).Create(&data.WarehouseMaterials).Error; err != nil {
			return err
		}

		if err := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "id"}},
			DoUpdates: clause.AssignmentColumns([]string{"amount"}),
		}).Create(&data.TeamMaterials).Error; err != nil {
			return err
		}

		if err := tx.Exec(`
        UPDATE serial_number_movements
        SET confirmation = true
        WHERE 
          serial_number_movements.invoice_type = 'output' AND
          serial_number_movements.invoice_id = ?
      `, data.InvoiceData.ID).Error; err != nil {
			return err
		}

		if err := tx.Exec(`
        UPDATE serial_number_locations
        SET 
          location_type = 'team',
          location_id = ?
        WHERE serial_number_locations.serial_number_id IN (
          SELECT serial_number_movements.serial_number_id
          FROM serial_number_movements
          WHERE
            serial_number_movements.invoice_type = 'output' AND
            serial_number_movements.invoice_id = ?
        )
      `, data.InvoiceData.TeamID, data.InvoiceData.ID).Error; err != nil {
			return err
		}

		return nil
	})
}

func (repo *invoiceOutputRepository) GetDataForExcel(id uint) (dto.InvoiceOutputDataForExcelQueryResult, error) {
	data := dto.InvoiceOutputDataForExcelQueryResult{}
	err := repo.db.Raw(`
      SELECT 
        invoice_outputs.id as id,
        projects.name as project_name,
        invoice_outputs.delivery_code as delivery_code,
        districts.name as district_name,
        team_leader.name as team_leader_name,
        warehouse_manager.name as warehouse_manager_name,
        released.name as released_name,
        recipient.name as recipient_name,
        invoice_outputs.date_of_invoice as date_of_invoice
      FROM invoice_outputs
      INNER JOIN projects ON projects.id = invoice_outputs.project_id
      INNER JOIN districts ON districts.id = invoice_outputs.district_id
      INNER JOIN teams ON teams.id = invoice_outputs.team_id
      INNER JOIN team_leaders ON team_leaders.team_id = teams.id
      INNER JOIN workers AS team_leader ON team_leader.id = team_leaders.leader_worker_id
      INNER JOIN workers AS warehouse_manager ON warehouse_manager.id = invoice_outputs.warehouse_manager_worker_id
      INNER JOIN workers AS released ON released.id = invoice_outputs.released_worker_id
      INNER JOIN workers AS recipient ON recipient.id = invoice_outputs.recipient_worker_id
      WHERE
        invoice_outputs.id = ?
      ORDER BY team_leaders.id DESC LIMIT 1;
    `, id).Scan(&data).Error
	return data, err
}

func (repo *invoiceOutputRepository) GetMaterialsForEdit(id uint) ([]dto.InvoiceOutputMaterialsForEdit, error) {
	result := []dto.InvoiceOutputMaterialsForEdit{}
	err := repo.db.Raw(`
    SELECT 
      materials.id as material_id,
      materials.name as material_name,
      materials.unit as material_unit,
      material_locations.amount as warehouse_amount,
      invoice_materials.amount as amount,
      invoice_materials.notes as notes,
      materials.has_serial_number as has_serial_number
    FROM invoice_materials
    INNER JOIN material_costs ON material_costs.id = invoice_materials.material_cost_id
    INNER JOIN materials ON materials.id = material_costs.material_id
    INNER JOIN material_locations ON material_locations.material_cost_id = invoice_materials.material_cost_id
    WHERE
      material_locations.location_type = 'warehouse' AND
      invoice_materials.invoice_type = 'output' AND
      invoice_materials.invoice_id = ?
    `, id).Scan(&result).Error

	return result, err
}

func (repo *invoiceOutputRepository) GetMaterialDataForReport(invoiceID uint) ([]dto.InvoiceOutputMaterialDataForReport, error) {
	result := []dto.InvoiceOutputMaterialDataForReport{}
	err := repo.db.Raw(`
    SELECT 
      materials.name as material_name,
      materials.unit as material_unit,
      material_costs.cost_m19 as material_cost_m19,
      invoice_materials.notes as notes,
      invoice_materials.amount as amount
    FROM invoice_materials
    INNER JOIN material_costs ON material_costs.id = invoice_materials.material_cost_id
    INNER JOIN materials ON materials.id = material_costs.material_id
    WHERE 
      invoice_materials.invoice_type = 'output' AND
      invoice_materials.invoice_id = ?;
    `, invoiceID).Scan(&result).Error

	return result, err
}

func (repo *invoiceOutputRepository) Import(data []dto.InvoiceOutputImportData) error {
	return repo.db.Transaction(func(tx *gorm.DB) error {

		for index, invoice := range data {
			invoiceOutput := invoice.Details
			if err := tx.Create(&invoiceOutput).Error; err != nil {
				return err
			}

			for subIndex := range invoice.Items {
				data[index].Items[subIndex].InvoiceID = invoiceOutput.ID
			}

			if err := tx.CreateInBatches(&data[index].Items, 15).Error; err != nil {
				return err
			}
		}

		err := tx.Exec(`
      UPDATE invoice_counts
      SET count = count + ?
      WHERE
        invoice_type = 'output' AND
        project_id = ?
      `, len(data), data[0].Details.ProjectID).Error
		if err != nil {
			return err
		}

		return nil
	})
}
