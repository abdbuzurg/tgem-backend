package repository

import (
	"backend-v2/internal/dto"
	"backend-v2/model"

	"gorm.io/gorm"
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
	GetPaginatedFiltered(page, limit int, filter model.InvoiceOutput) ([]model.InvoiceOutput, error)
	GetByID(id uint) (model.InvoiceOutput, error)
	GetUnconfirmedByObjectInvoices() ([]model.InvoiceOutput, error)
	Create(data model.InvoiceOutput) (model.InvoiceOutput, error)
	Update(data model.InvoiceOutput) (model.InvoiceOutput, error)
	Delete(id uint) error
	Count(projectID uint) (int64, error)
	UniqueCode(projectID uint) ([]string, error)
	UniqueWarehouseManager(projectID uint) ([]string, error)
	UniqueRecieved(projectID uint) ([]string, error)
	UniqueDistrict(projectID uint) ([]string, error)
	UniqueObject(projectID uint) ([]string, error)
	UniqueTeam(projectID uint) ([]string, error)
	ReportFilterData(filter dto.InvoiceOutputReportFilter, projectID uint) ([]model.InvoiceOutput, error)
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

func (repo *invoiceOutputRepository) GetPaginatedFiltered(page, limit int, filter model.InvoiceOutput) ([]model.InvoiceOutput, error) {
	data := []model.InvoiceOutput{}
	err := repo.db.
		Raw(`SELECT * FROM invoice_outputs WHERE
			(nullif(?, 0) IS NULL OR project_id = ?) AND
			(nullif(?, 0) IS NULL OR district_id = ?) AND
			(nullif(?, 0) IS NULL OR warehouse_manager_worker_id = ?) AND
			(nullif(?, 0) IS NULL OR released_worker_id = ?) AND
			(nullif(?, 0) IS NULL OR recipient_worker_id = ?) AND
			(nullif(?, 0) IS NULL OR team_id = ?) AND
			(nullif(?, 0) IS NULL OR object_id = ?) AND
			(nullif(?, '') IS NULL OR delivery_code = ?) AND
			(nullif(?, 0) IS NULL OR operator_add_worker_id = ?) AND
			(nullif(?, 0) IS NULL OR operator_edit_worker_id = ?) ORDER BY id DESC LIMIT ? OFFSET ?`,
			filter.ProjectID, filter.ProjectID,
			filter.DistrictID, filter.DistrictID,
			filter.WarehouseManagerWorkerID, filter.WarehouseManagerWorkerID,
			filter.ReleasedWorkerID, filter.ReleasedWorkerID,
			filter.RecipientWorkerID, filter.RecipientWorkerID,
			filter.TeamID, filter.TeamID,
			filter.ObjectID, filter.ObjectID,
			filter.DeliveryCode, filter.DeliveryCode,
			filter.OperatorAddWorkerID, filter.OperatorAddWorkerID,
			filter.OperatorEditWorkerID, filter.OperatorEditWorkerID,
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

func (repo *invoiceOutputRepository) Create(data model.InvoiceOutput) (model.InvoiceOutput, error) {
	err := repo.db.Create(&data).Error
	return data, err
}

func (repo *invoiceOutputRepository) Update(data model.InvoiceOutput) (model.InvoiceOutput, error) {
	err := repo.db.Model(&model.InvoiceOutput{}).Select("*").Where("id = ?", data.ID).Updates(&data).Error
	return data, err
}

func (repo *invoiceOutputRepository) Delete(id uint) error {
	return repo.db.Delete(&model.InvoiceOutput{}, "id = ?", id).Error
}

func (repo *invoiceOutputRepository) Count(projectID uint) (int64, error) {
	var count int64
	err := repo.db.Raw("SELECT COUNT(*) FROM invoice_inputs WHERE project_id = ?", projectID).Scan(&count).Error
	return count, err
}

func (repo *invoiceOutputRepository) GetUnconfirmedByObjectInvoices() ([]model.InvoiceOutput, error) {
	data := []model.InvoiceOutput{}
	err := repo.db.Find(&data, "confirmation = TRUE AND object_confirmation = FALSE ORDER BY id DESC").Error
	return data, err
}

func (repo *invoiceOutputRepository) UniqueCode(projectID uint) ([]string, error) {
	data := []string{}
	err := repo.db.Raw("SELECT DISTINCT delivery_code FROM invoice_outputs WHERE project_id = ?", projectID).Scan(&data).Error
	return data, err
}

func (repo *invoiceOutputRepository) UniqueWarehouseManager(projectID uint) ([]string, error) {
	data := []string{}
	err := repo.db.Raw("SELECT DISTINCT warehouse_manager_worker_id FROM invoice_outputs WHERE project_id = ?", projectID).Scan(&data).Error
	return data, err
}

func (repo *invoiceOutputRepository) UniqueRecieved(projectID uint) ([]string, error) {
	data := []string{}
	err := repo.db.Raw("SELECT DISTINCT recieved_worker_id FROM invoice_outputs WHERE project_id = ?", projectID).Scan(&data).Error
	return data, err
}

func (repo *invoiceOutputRepository) UniqueDistrict(projectID uint) ([]string, error) {
	data := []string{}
	err := repo.db.Raw("SELECT DISTINCT district_id FROM invoice_outputs WHERE project_id = ?", projectID).Scan(&data).Error
	return data, err
}

func (repo *invoiceOutputRepository) UniqueObject(projectID uint) ([]string, error) {
	data := []string{}
	err := repo.db.Raw("SELECT DISTINCT object_id FROM invoice_outputs WHERE project_id = ?", projectID).Scan(&data).Error
	return data, err
}

func (repo *invoiceOutputRepository) UniqueTeam(projectID uint) ([]string, error) {
	data := []string{}
	err := repo.db.Raw("SELECT DISTINCT team_id FROM invoice_outputs WHERE project_id = ?", projectID).Scan(&data).Error
	return data, err
}

func (repo *invoiceOutputRepository) ReportFilterData(filter dto.InvoiceOutputReportFilter, projectID uint) ([]model.InvoiceOutput, error) {
	data := []model.InvoiceOutput{}
	dateFrom := filter.DateFrom.String()
	dateFrom = dateFrom[:len(dateFrom)-10]
	dateTo := filter.DateTo.String()
	dateTo = dateTo[:len(dateTo)-10]
	err := repo.
		db.
		Raw(`SELECT * FROM invoice_outputs WHERE
			(nullif(?, 0) IS NULL OR project_id = ?) AND
			(nullif(?, '') IS NULL OR delivery_code = ?) AND
			(nullif(?, 0) IS NULL OR recipient_worker_id = ?) AND
			(nullif(?, 0) IS NULL OR warehouse_manager_worker_id = ?) AND
			(nullif(?, 0) IS NULL OR district_id = ?) AND
			(nullif(?, 0) IS NULL OR object_id = ?) AND
			(nullif(?, 0) IS NULL OR team_id = ?) AND
			(nullif(?, '0001-01-01 00:00:00') IS NULL OR ? <= date_of_invoice) AND 
			(nullif(?, '0001-01-01 00:00:00') IS NULL OR date_of_invoice <= ?) ORDER BY id DESC
		`,
			projectID, projectID,
			filter.Code, filter.Code,
			filter.ReceivedID, filter.ReceivedID,
			filter.WarehouseManagerID, filter.WarehouseManagerID,
			filter.DistrictID, filter.DistrictID,
			filter.ObjectID, filter.ObjectID,
			filter.TeamID, filter.TeamID,
			dateFrom, dateFrom,
			dateTo, dateTo).
		Scan(&data).Error

	return data, err
}
