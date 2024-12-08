package repository

import (
	"backend-v2/internal/dto"
	"backend-v2/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type invoiceReturnRepository struct {
	db *gorm.DB
}

func InitInvoiceReturnRepository(db *gorm.DB) IInvoiceReturnRepository {
	return &invoiceReturnRepository{
		db: db,
	}
}

type IInvoiceReturnRepository interface {
	GetAll() ([]model.InvoiceReturn, error)
	GetByID(id uint) (model.InvoiceReturn, error)
	GetPaginatedTeam(page, limit int, projectID uint) ([]dto.InvoiceReturnTeamPaginatedQueryData, error)
	GetPaginatedObject(page, limit int, projectID uint) ([]dto.InvoiceReturnObjectPaginatedQueryData, error)
  GetByDeliveryCode(deliveryCode string) (model.InvoiceReturn, error)
	Create(data dto.InvoiceReturnCreateQueryData) (model.InvoiceReturn, error)
	Update(data dto.InvoiceReturnCreateQueryData) (model.InvoiceReturn, error)
	Delete(id uint) error
	CountBasedOnType(projectID uint, invoiceType string) (int64, error)
	Count(projectID uint) (int64, error)
	UniqueCode(projectID uint) ([]string, error)
	UniqueTeam(projectID uint) ([]uint, error)
	UniqueObject(projectID uint) ([]uint, error)
	ReportFilterData(filter dto.InvoiceReturnReportFilter, projectID uint) ([]model.InvoiceReturn, error)
	GetInvoiceReturnMaterialsForExcel(id uint) ([]dto.InvoiceReturnMaterialsForExcel, error)
	GetInvoiceReturnTeamDataForExcel(id uint) (dto.InvoiceReturnTeamDataForExcel, error)
	GetInvoiceReturnObjectDataForExcel(id uint) (dto.InvoiceReturnObjectDataForExcel, error)
	Confirmation(data dto.InvoiceReturnConfirmDataQuery) error
	GetMaterialsForEdit(id uint, locationType string, locationID uint) ([]dto.InvoiceReturnMaterialForEdit, error)
}

func (repo *invoiceReturnRepository) GetAll() ([]model.InvoiceReturn, error) {
	data := []model.InvoiceReturn{}
	err := repo.db.Order("id desc").Find(&data).Error
	return data, err
}

func (repo *invoiceReturnRepository) GetPaginatedTeam(page, limit int, projectID uint) ([]dto.InvoiceReturnTeamPaginatedQueryData, error) {
	data := []dto.InvoiceReturnTeamPaginatedQueryData{}
	err := repo.db.Raw(`
    SELECT 
      invoice_returns.id as id,
      invoice_returns.delivery_code,
      districts.name as district_name,
      teams.number as team_number,
      workers.name as team_leader_name,
      acceptor_worker.name as acceptor_name,
      invoice_returns.date_of_invoice as date_of_invoice,
      invoice_returns.confirmation as confirmation
    FROM invoice_returns
    INNER JOIN districts ON districts.id = invoice_returns.district_id
    INNER JOIN teams ON teams.id = invoice_returns.returner_id
    INNER JOIN team_leaders ON team_leaders.team_id = teams.id
    INNER JOIN workers ON workers.id = team_leaders.leader_worker_id
    INNER JOIN workers AS acceptor_worker ON acceptor_worker.id = invoice_returns.accepted_by_worker_id
    WHERE
      invoice_returns.project_id = ? AND
      invoice_returns.returner_type = 'team'
    ORDER by invoice_returns.id DESC
    LIMIT ? OFFSET ?;
    `, projectID, limit, (page-1)*limit).Scan(&data).Error

	return data, err
}

func (repo *invoiceReturnRepository) GetPaginatedObject(page, limit int, projectID uint) ([]dto.InvoiceReturnObjectPaginatedQueryData, error) {
	data := []dto.InvoiceReturnObjectPaginatedQueryData{}
	err := repo.db.Raw(`
    SELECT 
      invoice_returns.id as id,
      invoice_returns.delivery_code,
      districts.name as district_name,
      objects.name as object_name,
      objects.type as object_type,
      workers.name as object_supervisor_name,
      teams.number as team_number,
      leader.name as team_leader_name,
      acceptor_worker.name as acceptor_name,
      invoice_returns.date_of_invoice as date_of_invoice,
      invoice_returns.confirmation as confirmation
    FROM invoice_returns
    INNER JOIN districts ON districts.id = invoice_returns.district_id
    INNER JOIN objects ON objects.id = invoice_returns.returner_id
    INNER JOIN object_supervisors ON objects.id = object_supervisors.object_id
    INNER JOIN workers ON workers.id = object_supervisors.supervisor_worker_id
    INNER JOIN teams ON teams.id = invoice_returns.acceptor_id 
    INNER JOIN team_leaders ON team_leaders.team_id = teams.id
    INNER JOIN workers AS leader ON leader.id = team_leaders.leader_worker_id
    INNER JOIN workers AS acceptor_worker ON acceptor_worker.id = invoice_returns.accepted_by_worker_id
    WHERE
      invoice_returns.project_id = ? AND
      invoice_returns.returner_type = 'object'
    ORDER by invoice_returns.id DESC
    LIMIT ? OFFSET ?;
    `, projectID, limit, (page-1)*limit).Scan(&data).Error

	return data, err
}

func (repo *invoiceReturnRepository) GetByID(id uint) (model.InvoiceReturn, error) {
	data := model.InvoiceReturn{}
	err := repo.db.Find(&data, "id = ?", id).Error
	return data, err
}

func (repo *invoiceReturnRepository) Create(data dto.InvoiceReturnCreateQueryData) (model.InvoiceReturn, error) {
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
        invoice_type = 'return' AND
        project_id = ?
      `, result.ProjectID).Error
		if err != nil {
			return err
		}

		return nil
	})

	return result, err
}

func (repo *invoiceReturnRepository) Update(data dto.InvoiceReturnCreateQueryData) (model.InvoiceReturn, error) {
	result := data.Invoice
	err := repo.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.InvoiceReturn{}).Select("*").Where("id = ?", result.ID).Updates(&result).Error; err != nil {
			return err
		}

		if err := tx.Delete(model.InvoiceMaterials{}, "invoice_id = ? AND invoice_type='return'", result.ID).Error; err != nil {
			return nil
		}

		for index := range data.InvoiceMaterials {
			data.InvoiceMaterials[index].InvoiceID = result.ID
		}

		if err := tx.CreateInBatches(&data.InvoiceMaterials, 15).Error; err != nil {
			return err
		}

		if err := tx.Delete(model.SerialNumberMovement{}, "invoice_id = ? AND invoice_type='output'", result.ID).Error; err != nil {
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

func (repo *invoiceReturnRepository) Delete(id uint) error {
	return repo.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Delete(&model.InvoiceReturn{}, "id = ?", id).Error; err != nil {
			return err
		}

		if err := tx.Delete(&model.InvoiceMaterials{}, "invoice_type = 'return' AND invoice_id = ?", id).Error; err != nil {
			return err
		}

		return nil

	})
}

func (repo *invoiceReturnRepository) CountBasedOnType(projectID uint, invoiceType string) (int64, error) {
	var count int64
	err := repo.db.Raw("SELECT COUNT(*) FROM invoice_returns WHERE project_id = ? AND returner_type = ?", projectID, invoiceType).Scan(&count).Error
	return count, err
}

func (repo *invoiceReturnRepository) Count(projectID uint) (int64, error) {
	var count int64
	err := repo.db.Raw("SELECT COUNT(*) FROM invoice_returns WHERE project_id = ?", projectID).Scan(&count).Error
	return count, err
}

func (repo *invoiceReturnRepository) UniqueCode(projectID uint) ([]string, error) {
	var data []string
	err := repo.db.Raw("SELECT DISTINCT delivery_code FROM invoice_returns WHERE project_id = ?", projectID).Scan(&data).Error
	return data, err
}

func (repo *invoiceReturnRepository) UniqueTeam(projectID uint) ([]uint, error) {
	var data []uint
	err := repo.db.Raw("SELECT DISTINCT returner_id FROM invoice_returns WHERE returner_type='teams' AND project_id = ?", projectID).Scan(&data).Error
	return data, err
}

func (repo *invoiceReturnRepository) UniqueObject(projectID uint) ([]uint, error) {
	var data []uint
	err := repo.db.Raw("SELECT DISTINCT returner_id FROM invoice_returns WHERE returner_type='objects' AND project_id = ?", projectID).Scan(&data).Error
	return data, err
}

func (repo *invoiceReturnRepository) ReportFilterData(filter dto.InvoiceReturnReportFilter, projectID uint) ([]model.InvoiceReturn, error) {
	data := []model.InvoiceReturn{}
	dateFrom := filter.DateFrom.String()
	dateFrom = dateFrom[:len(dateFrom)-10]
	dateTo := filter.DateTo.String()
	dateTo = dateTo[:len(dateTo)-10]
	err := repo.
		db.
		Raw(`SELECT * FROM invoice_returns WHERE
			project_id = ? AND
			(nullif(?, '') IS NULL OR delivery_code = ?) AND
			(nullif(?, '') IS NULL OR returner_type = ?) AND
			(nullif(?, 0) IS NULL OR returner_id = ?) AND
			(nullif(?, '0001-01-01 00:00:00') IS NULL OR ? <= date_of_invoice) AND 
			(nullif(?, '0001-01-01 00:00:00') IS NULL OR date_of_invoice <= ?) ORDER BY id DESC
		`,
			projectID,
			filter.Code, filter.Code,
			filter.ReturnerType, filter.ReturnerType,
			filter.ReturnerID, filter.ReturnerID,
			dateFrom, dateFrom,
			dateTo, dateTo).
		Scan(&data).Error

	return data, err
}

func (repo *invoiceReturnRepository) GetInvoiceReturnMaterialsForExcel(id uint) ([]dto.InvoiceReturnMaterialsForExcel, error) {
	data := []dto.InvoiceReturnMaterialsForExcel{}
	err := repo.db.Raw(`
    SELECT 
      materials.code as material_code,
      materials.name as material_name,
      materials.unit as material_unit,
      invoice_materials.is_defected as material_defected,
      invoice_materials.amount as material_amount,
      invoice_materials.notes as material_notes	
    FROM invoice_materials
    INNER JOIN material_costs ON material_costs.id = invoice_materials.material_cost_id
    INNER JOIN materials ON materials.id = material_costs.material_id
    WHERE
      invoice_materials.invoice_type = 'return' AND
      invoice_materials.invoice_id = 1;
    `).Scan(&data).Error

	return data, err
}

func (repo *invoiceReturnRepository) GetInvoiceReturnTeamDataForExcel(id uint) (dto.InvoiceReturnTeamDataForExcel, error) {
	data := dto.InvoiceReturnTeamDataForExcel{}
	err := repo.db.Raw(`
    SELECT 
      projects.name as project_name,
      invoice_returns.date_of_invoice as date_of_invoice,
      districts.name as district_name,
      invoice_returns.delivery_code as delivery_code,
      teams.number as team_number,
      team_leader_worker.name as team_leader_name,
      acceptor_worker.name as acceptor_name
    FROM invoice_returns
    INNER JOIN projects ON projects.id = invoice_returns.project_id
    INNER JOIN districts ON districts.id = invoice_returns.district_id
    INNER JOIN teams ON teams.id = invoice_returns.returner_id
    INNER JOIN team_leaders ON team_leaders.team_id = teams.id
    INNER JOIN workers AS team_leader_worker ON team_leader_worker.id = team_leaders.leader_worker_id
    INNER JOIN workers AS acceptor_worker ON acceptor_worker.id = invoice_returns.accepted_by_worker_id
    WHERE
      invoice_returns.id = ?
    LIMIT 1
    `, id).Scan(&data).Error

	return data, err
}

func (repo *invoiceReturnRepository) GetInvoiceReturnObjectDataForExcel(id uint) (dto.InvoiceReturnObjectDataForExcel, error) {
	data := dto.InvoiceReturnObjectDataForExcel{}
	err := repo.db.Raw(`
    SELECT 
      invoice_returns.delivery_code AS delivery_code,
      invoice_returns.date_of_invoice AS date_of_invoice,
      projects.name as project_name,
      districts.name as district_name,
      objects.type as object_type,
      objects.name as object_name,
      supervisor_worker.name as supervisor_name,
      team_leader_worker.name as team_leader_name
    FROM invoice_returns
    INNER JOIN projects ON projects.id = invoice_returns.project_id
    INNER JOIN districts ON districts.id = invoice_returns.district_id
    INNER JOIN objects ON objects.id = invoice_returns.returner_id
    INNER JOIN object_supervisors ON object_supervisors.object_id = objects.id
    INNER JOIN workers AS supervisor_worker ON supervisor_worker.id = object_supervisors.supervisor_worker_id
    INNER JOIN workers AS team_leader_worker ON team_leader_worker.id = invoice_returns.accepted_by_worker_id
    WHERE 
      invoice_returns.id = ?
    LIMIT 1;	
    `, id).Scan(&data).Error

	return data, err
}

func (repo *invoiceReturnRepository) Confirmation(data dto.InvoiceReturnConfirmDataQuery) error {
	return repo.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.InvoiceReturn{}).Select("*").Where("id = ?", data.Invoice.ID).Updates(&data.Invoice).Error; err != nil {
			return err
		}

		if err := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "id"}},
			DoUpdates: clause.AssignmentColumns([]string{"amount"}),
		}).Create(&data.MaterialsInReturnerLocation).Error; err != nil {
			return err
		}

		if err := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "id"}},
			DoUpdates: clause.AssignmentColumns([]string{"amount"}),
		}).Create(&data.MaterialsInAcceptorLocation).Error; err != nil {
			return err
		}

		if len(data.MaterialsDefected) != 0 {
			if err := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "id"}},
				DoUpdates: clause.AssignmentColumns([]string{"amount", "material_location_id"}),
			}).Create(&data.MaterialsDefected).Error; err != nil {
				return err
			}
		}

		if err := tx.CreateInBatches(&data.NewMaterialsInAcceptorLocationWithNewDefect, 15).Error; err != nil {
			return err
		}

		for index := range data.NewMaterialsDefected {
			data.NewMaterialsDefected[index].MaterialLocationID = data.NewMaterialsInAcceptorLocationWithNewDefect[index].ID
		}

		if err := tx.CreateInBatches(&data.NewMaterialsDefected, 15).Error; err != nil {
			return err
		}

		if err := tx.Exec(`
        UPDATE serial_number_movements
        SET confirmation = true
        WHERE 
          serial_number_movements.invoice_type = 'return' AND
          serial_number_movements.invoice_id = ?
      `, data.Invoice.ID).Error; err != nil {
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
            serial_number_movements.invoice_type = 'return' AND
            serial_number_movements.invoice_id = ?
        )
      `, data.Invoice.AcceptorID, data.Invoice.ID).Error; err != nil {
			return err
		}

		return nil
	})
}

func (repo *invoiceReturnRepository) GetMaterialsForEdit(id uint, locationType string, locationID uint) ([]dto.InvoiceReturnMaterialForEdit, error) {
	result := []dto.InvoiceReturnMaterialForEdit{}
	err := repo.db.Raw(`
    SELECT 
      materials.id as material_id,
      materials.name as material_name,
      materials.unit as unit,
      invoice_materials.amount as amount,
      invoice_materials.notes as  notes,
      materials.has_serial_number as has_serial_number,
      invoice_materials.is_defected as is_defective, 
      material_locations.amount as holder_amount
    FROM invoice_materials
    INNER JOIN material_costs ON invoice_materials.material_cost_id = material_costs.id
    INNER JOIN materials ON material_costs.material_id = materials.id
    INNER JOIN material_locations ON material_locations.material_cost_id = invoice_materials.material_cost_id
    WHERE
      invoice_materials.invoice_type='return' AND
      invoice_materials.invoice_id = ? AND
      material_locations.location_type = ? AND
      material_locations.location_id = ?
    ORDER BY materials.id
    `, id, locationType, locationID).Scan(&result).Error

	return result, err
}

func(repo *invoiceReturnRepository) GetByDeliveryCode(deliveryCode string) (model.InvoiceReturn, error) {
  result := model.InvoiceReturn{}
  err := repo.db.Raw(`SELECT * FROM invoice WHERE delivery_code = ?`, deliveryCode).Scan(&result).Error
  return result, err
}
