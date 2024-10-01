package repository

import (
	"backend-v2/internal/dto"
	"backend-v2/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type invoiceCorrectionRepository struct {
	db *gorm.DB
}

func InitInvoiceCorrectionRepository(db *gorm.DB) IInvoiceCorrectionRepository {
	return &invoiceCorrectionRepository{
		db: db,
	}
}

type IInvoiceCorrectionRepository interface {
	GetPaginated(page, limit int, projectID uint) ([]dto.InvoiceCorrectionPaginated, error)
	GetInvoiceMaterialsDataByInvoiceObjectID(id uint) ([]dto.InvoiceCorrectionMaterialsData, error)
	GetSerialNumberOfMaterialInTeam(projectID uint, materialID uint, teamID uint) ([]string, error)
	Create(data dto.InvoiceCorrectionCreateQuery) (model.InvoiceObject, error)
	UniqueTeam(projectID uint) ([]dto.DataForSelect[uint], error)
	UniqueObject(projectID uint) ([]dto.ObjectDataForSelect, error)
	ReportFilterData(filter dto.InvoiceCorrectionReportFilter) ([]dto.InvoiceCorrectionReportData, error)
	Count(projectID uint) (int64, error)
  GetOperationsByInvoiceObjectID(id uint) ([]dto.InvoiceCorrectionOperationsData, error)
}

func (repo *invoiceCorrectionRepository) GetPaginated(page, limit int, projectID uint) ([]dto.InvoiceCorrectionPaginated, error) {
	data := []dto.InvoiceCorrectionPaginated{}
	err := repo.db.Raw(`
    SELECT 
      invoice_objects.id as id,
      workers.name as supervisor_name,
      objects.name as object_name,
      objects.type as object_type,
      teams.id as team_id,
      teams.number as team_number,
      invoice_objects.date_of_invoice as date_of_invoice,
      invoice_objects.delivery_code as delivery_code,
      invoice_objects.confirmed_by_operator as confirmed_by_operator  

    FROM invoice_objects
      INNER JOIN workers ON workers.id = invoice_objects.supervisor_worker_id
      INNER JOIN objects ON objects.id = invoice_objects.object_id
      INNER JOIN teams ON teams.id = invoice_objects.team_id
    WHERE
      invoice_objects.project_id = ? AND
      invoice_objects.confirmed_by_operator = false
    ORDER BY invoice_objects.id DESC LIMIT ? OFFSET ?;
    `, projectID, limit, (page-1)*limit).Scan(&data).Error

	return data, err

}

func (repo *invoiceCorrectionRepository) GetInvoiceMaterialsDataByInvoiceObjectID(id uint) ([]dto.InvoiceCorrectionMaterialsData, error) {
	data := []dto.InvoiceCorrectionMaterialsData{}
	err := repo.db.Raw(`
    SELECT 
      invoice_materials.id as invoice_material_id,
      materials.name as material_name,
      materials.id as material_id,
      invoice_materials.notes as notes,
      invoice_materials.amount as material_amount
    FROM invoice_materials
    INNER JOIN material_costs ON material_costs.id = invoice_materials.material_cost_id
    INNER JOIN materials ON materials.id = material_costs.material_id
    WHERE 
      invoice_materials.invoice_type = 'object' AND
      invoice_materials.invoice_id = ?
    `, id).Scan(&data).Error

	return data, err
}

func (repo *invoiceCorrectionRepository) GetSerialNumberOfMaterialInTeam(projectID uint, materialID uint, teamID uint) ([]string, error) {
	data := []string{}
	err := repo.db.Raw(`
    SELECT serial_numbers.code
    FROM material_locations
    INNER JOIN teams ON team.id = material_locations.location_id
    INNER JOIN serial_numbers ON serial_numbers.material_cost_id = material_locations.material_cost_id
    INNER JOIN serial_number_locations ON serial_number_locations.serial_number_id = serial_numbers.id
    INNER JOIN material_costs ON material_locations.material_cost_id = material_costs.id
    INNER JOIN materials ON materials.id = material_costs.material_id
    WHERE
      materials.project_id = ? AND
      materials.id = ? AND
      teams.id = ? AND
      material_locations.location_type = serial_number_locations.location_type AND
      material_locations.location_id = serial_number_locations.location_id
    `, projectID, materialID, teamID).Scan(&data).Error

	return data, err
}

func (repo *invoiceCorrectionRepository) Create(data dto.InvoiceCorrectionCreateQuery) (model.InvoiceObject, error) {
	result := data.Details
	err := repo.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.InvoiceObject{}).Select("*").Where("id = ?", result.ID).Updates(&result).Error; err != nil {
			return err
		}

		if err := tx.Create(&data.OperatorDetails).Error; err != nil {
			return err
		}

		for index := range data.Items {
			data.Items[index].InvoiceID = result.ID
		}

		if err := tx.CreateInBatches(&data.Items, 15).Error; err != nil {
			return err
		}

    if err := tx.Delete(&model.ObjectOperation{}, "invoice_object_id = ?", result.ID).Error; err != nil {
      return err
    }

    if err := tx.CreateInBatches(&data.ObjectOperations, 15).Error; err != nil {
      return err
    }

		if err := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "id"}},
			DoUpdates: clause.AssignmentColumns([]string{"amount"}),
		}).Create(&data.TeamLocation).Error; err != nil {
			return err
		}

		if err := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "id"}},
			DoUpdates: clause.AssignmentColumns([]string{"amount"}),
		}).Create(&data.ObjectLocation).Error; err != nil {
			return err
		}

		return nil
	})

	return result, err
}

func (repo *invoiceCorrectionRepository) UniqueObject(projectID uint) ([]dto.ObjectDataForSelect, error) {
	result := []dto.ObjectDataForSelect{}
	err := repo.db.Raw(`
    SELECT 
      objects.id as id,
      objects.name as object_name,
      objects.type as object_type
    FROM objects
    WHERE objects.id IN (
      SELECT DISTINCT(invoice_objects.object_id)
      FROM invoice_objects
      WHERE invoice_objects.project_id = ?
    )
  `, projectID).Scan(&result).Error

	return result, err
}

func (repo *invoiceCorrectionRepository) UniqueTeam(projectID uint) ([]dto.DataForSelect[uint], error) {
	result := []dto.DataForSelect[uint]{}
	err := repo.db.Raw(`
      SELECT 
        teams.id as "value",
        CONCAT(teams.number, ' (', workers.name, ')') as "label"
      FROM teams
      INNER JOIN team_leaders ON team_leaders.team_id = teams.id
      INNER JOIN workers ON workers.id = team_leaders.leader_worker_id
      WHERE teams.id IN (
        SELECT DISTINCT(invoice_objects.team_id)
        FROM invoice_objects
        WHERE invoice_objects.project_id = ?
      )
    `, projectID).Scan(&result).Error

	return result, err
}

func (repo *invoiceCorrectionRepository) ReportFilterData(filter dto.InvoiceCorrectionReportFilter) ([]dto.InvoiceCorrectionReportData, error) {
	data := []dto.InvoiceCorrectionReportData{}
	dateFrom := filter.DateFrom.String()
	dateFrom = dateFrom[:len(dateFrom)-10]
	dateTo := filter.DateTo.String()
	dateTo = dateTo[:len(dateTo)-10]
	err := repo.
		db.
		Raw(`
      SELECT 
        invoice_objects.id as id,
        invoice_objects.delivery_code as delivery_code,
        objects.name as object_name,
        objects.type as object_type,
        teams.number as team_number,
        team_leader.name as team_leader_name,
        invoice_objects.date_of_invoice as date_of_invoice,
        operator.name as operator_name,
        invoice_objects.date_of_correction as date_of_correction
      FROM invoice_objects
      INNER JOIN objects ON objects.id = invoice_objects.object_id
      INNER JOIN teams ON teams.id = invoice_objects.team_id
      INNER JOIN team_leaders ON team_leaders.team_id = teams.id
      INNER JOIN workers AS team_leader ON team_leader.id = team_leaders.leader_worker_id
      INNER JOIN invoice_object_operators ON invoice_object_operators.invoice_object_id = invoice_objects.id
      INNER JOIN workers AS operator ON operator.id = invoice_object_operators.operator_worker_id 
      WHERE 
        invoice_objects.project_id = ? AND
        invoice_objects.confirmed_by_operator = true AND
        (nullif(?, 0) IS NULL OR invoice_objects.object_id = ?) AND
        (nullif(?, 0) IS NULL OR invoice_objects.team_id = ?) AND
        (nullif(?, '0001-01-01 00:00:00') IS NULL OR ? <= invoice_objects.date_of_invoice) AND 
        (nullif(?, '0001-01-01 00:00:00') IS NULL OR invoice_objects.date_of_invoice <= ?) ORDER BY invoice_objects.id DESC
    `,
			filter.ProjectID,
			filter.ObjectID, filter.ObjectID,
			filter.TeamID, filter.TeamID,
			dateFrom, dateFrom,
			dateTo, dateTo,
		).
		Scan(&data).Error

	return data, err
}

func (repo *invoiceCorrectionRepository) Count(projectID uint) (int64, error) {
	var count int64
	err := repo.db.Raw(`
    SELECT COUNT(*) 
    FROM invoice_objects 
    WHERE 
      confirmed_by_operator = false AND
      project_id = ?`, projectID).Scan(&count).Error
	return count, err
}

func (repo *invoiceCorrectionRepository) GetOperationsByInvoiceObjectID(id uint) ([]dto.InvoiceCorrectionOperationsData, error) {
  result := []dto.InvoiceCorrectionOperationsData{}
  err := repo.db.Raw(`
    SELECT 
      operations.id as operation_id,
      operations.name as operation_name,
      object_operations.amount as amount,
      materials.name as material_name
    FROM invoice_objects
    INNER JOIN object_operations ON object_operations.invoice_object_id = invoice_objects.id
    INNER JOIN operations ON operations.id = object_operations.operation_id
    INNER JOIN operation_materials ON operation_materials.operation_id = operations.id
    INNER JOIN materials ON operation_materials.material_id = materials.id
    WHERE
      invoice_objects.id = ?;
    `, id).Scan(&result).Error

  return result, err
}
