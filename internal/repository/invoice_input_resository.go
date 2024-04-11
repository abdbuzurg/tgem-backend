package repository

import (
	"backend-v2/internal/dto"
	"backend-v2/model"

	"gorm.io/gorm"
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
	GetPaginatedFiltered(page, limit int, filter model.InvoiceInput) ([]model.InvoiceInput, error)
	GetByID(id uint) (model.InvoiceInput, error)
	Create(data model.InvoiceInput) (model.InvoiceInput, error)
	Update(data model.InvoiceInput) (model.InvoiceInput, error)
	Delete(id uint) error
	Count(projectID uint) (int64, error)
	UniqueCode(projectID uint) ([]string, error)
	UniqueWarehouseManager(projectID uint) ([]string, error)
	UniqueReleased(projectID uint) ([]string, error)
	ReportFilterData(filter dto.InvoiceInputReportFilter, projectID uint) ([]model.InvoiceInput, error)
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

func (repo *invoiceInputRespository) GetPaginatedFiltered(page, limit int, filter model.InvoiceInput) ([]model.InvoiceInput, error) {
	data := []model.InvoiceInput{}
	err := repo.db.
		Raw(`SELECT * FROM invoice_inputs WHERE
			(nullif(?, 0) IS NULL OR project_id = ?) AND
			(nullif(?, 0) IS NULL OR warehouse_manager_worker_id = ?) AND
			(nullif(?, 0) IS NULL OR released_worker_id = ?) AND
			(nullif(?, '') IS NULL OR delivery_code = ?) ORDER BY id DESC LIMIT ? OFFSET ?`,
			filter.ProjectID, filter.ProjectID,
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

func (repo *invoiceInputRespository) Create(data model.InvoiceInput) (model.InvoiceInput, error) {
	err := repo.db.Create(&data).Error
	return data, err
}

func (repo *invoiceInputRespository) Update(data model.InvoiceInput) (model.InvoiceInput, error) {
	err := repo.db.Model(&model.InvoiceInput{}).Select("*").Where("id = ?", data.ID).Updates(&data).Error
	return data, err
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
