package repository

import (
	"backend-v2/internal/dto"
	"backend-v2/model"

	"gorm.io/gorm"
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
	GetPaginated(page, limit int) ([]model.InvoiceReturn, error)
	GetPaginatedFiltered(page, limit int, filter model.InvoiceReturn) ([]model.InvoiceReturn, error)
	GetByID(id uint) (model.InvoiceReturn, error)
	Create(data model.InvoiceReturn) (model.InvoiceReturn, error)
	Update(data model.InvoiceReturn) (model.InvoiceReturn, error)
	Delete(id uint) error
	Count(projectID uint) (int64, error)
	UniqueCode(projectID uint) ([]string, error)
	UniqueTeam(projectID uint) ([]uint, error)
	UniqueObject(projectID uint) ([]uint, error)
	ReportFilterData(filter dto.InvoiceReturnReportFilter, projectID uint) ([]model.InvoiceReturn, error)
}

func (repo *invoiceReturnRepository) GetAll() ([]model.InvoiceReturn, error) {
	data := []model.InvoiceReturn{}
	err := repo.db.Order("id desc").Find(&data).Error
	return data, err
}

func (repo *invoiceReturnRepository) GetPaginated(page, limit int) ([]model.InvoiceReturn, error) {
	data := []model.InvoiceReturn{}
	err := repo.db.Order("id desc").Offset((page - 1) * limit).Limit(limit).Find(&data).Error
	return data, err
}

func (repo *invoiceReturnRepository) GetPaginatedFiltered(page, limit int, filter model.InvoiceReturn) ([]model.InvoiceReturn, error) {
	data := []model.InvoiceReturn{}
	err := repo.db.
		Raw(`SELECT * FROM invoice_returns WHERE
			project_id = ? AND
			(nullif(?, '') IS NULL OR delivery_code = ?) AND
			(nullif(?, '') IS NULL OR returner_type = ?) AND
			(nullif(?, 0) IS NULL OR returner_id = ?) AND
			(nullif(?, 0) IS NULL OR operator_add_worker_id = ?) AND
			(nullif(?, 0) IS NULL OR operator_edit_worker_id = ?) ORDER BY id DESC LIMIT ? OFFSET ?`,
			filter.ProjectID,
			filter.DeliveryCode, filter.DeliveryCode,
			filter.ReturnerType, filter.ReturnerType,
			filter.ReturnerID, filter.ReturnerID,
			filter.OperatorAddWorkerID, filter.OperatorAddWorkerID,
			filter.OperatorEditWorkerID, filter.OperatorEditWorkerID,
			limit, (page-1)*limit,
		).
		Scan(&data).Error

	return data, err
}

func (repo *invoiceReturnRepository) GetByID(id uint) (model.InvoiceReturn, error) {
	data := model.InvoiceReturn{}
	err := repo.db.Find(&data, "id = ?", id).Error
	return data, err
}

func (repo *invoiceReturnRepository) Create(data model.InvoiceReturn) (model.InvoiceReturn, error) {
	err := repo.db.Create(&data).Error
	return data, err
}

func (repo *invoiceReturnRepository) Update(data model.InvoiceReturn) (model.InvoiceReturn, error) {
	err := repo.db.Model(&model.InvoiceReturn{}).Select("*").Where("id = ?", data.ID).Updates(&data).Error
	return data, err
}

func (repo *invoiceReturnRepository) Delete(id uint) error {
	return repo.db.Delete(&model.InvoiceReturn{}, "id = ?", id).Error
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
