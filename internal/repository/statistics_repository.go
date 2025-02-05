package repository

import "gorm.io/gorm"

type statisticsRepository struct {
	db *gorm.DB
}

type IStatisticsRepository interface {
	CountInvoiceInputs(projectID uint) (int, error)
	CountInvoiceOutputs(projectID uint) (int, error)
	CountInvoiceReturns(projectID uint) (int, error)
	CountInvoiceWriteOffs(projectID uint) (int, error)
	CountInvoiceInputUniqueCreators(projectID uint) ([]int, error)
	CountInvoiceInputCreatorInvoices(projectID, workerID uint) (int, error)
	CountInvoiceOutputUniqueCreators(projectID uint) ([]int, error)
	CountInvoiceOutputCreatorInvoices(projectID, workerID uint) (int, error)
}

func NewStatisticsRepository(db *gorm.DB) IStatisticsRepository {
	return &statisticsRepository{
		db: db,
	}
}

func (repo *statisticsRepository) CountInvoiceInputs(projectID uint) (int, error) {
	count := 0
	err := repo.db.Raw(`SELECT COUNT(*) FROM invoice_inputs WHERE project_id = ?`, projectID).Scan(&count).Error
	return count, err
}

func (repo *statisticsRepository) CountInvoiceOutputs(projectID uint) (int, error) {
	count := 0
	err := repo.db.Raw(`SELECT COUNT(*) FROM invoice_outputs WHERE project_id = ?`, projectID).Scan(&count).Error
	return count, err
}

func (repo *statisticsRepository) CountInvoiceReturns(projectID uint) (int, error) {
	count := 0
	err := repo.db.Raw(`SELECT COUNT(*) FROM invoice_returns WHERE project_id = ?`, projectID).Scan(&count).Error
	return count, err
}

func (repo *statisticsRepository) CountInvoiceWriteOffs(projectID uint) (int, error) {
	count := 0
	err := repo.db.Raw(`SELECT COUNT(*) FROM invoice_write_offs WHERE project_id = ?`, projectID).Scan(&count).Error
	return count, err
}

func (repo *statisticsRepository) CountInvoiceInputUniqueCreators(projectID uint) ([]int, error) {
	uniqueCreators := []int{}
	err := repo.db.Raw(`SELECT DISTINCT released_worker_id FROM invoice_inputs WHERE project_id = ?`, projectID).Scan(&uniqueCreators).Error
	return uniqueCreators, err
}

func (repo *statisticsRepository) CountInvoiceInputCreatorInvoices(projectID, workerID uint) (int, error) {
	count := 0
	err := repo.db.Raw(`SELECT COUNT(*) FROM invoice_inputs WHERE  project_id = ? AND released_worker_id = ?`, projectID, workerID).Scan(&count).Error
	return count, err
}

func (repo *statisticsRepository) CountInvoiceOutputUniqueCreators(projectID uint) ([]int, error) {
	uniqueCreators := []int{}
	err := repo.db.Raw(`SELECT DISTINCT released_worker_id FROM invoice_outputs WHERE project_id = ?`, projectID).Scan(&uniqueCreators).Error
	return uniqueCreators, err
}

func (repo *statisticsRepository) CountInvoiceOutputCreatorInvoices(projectID, workerID uint) (int, error) {
	count := 0
	err := repo.db.Raw(`SELECT COUNT(*) FROM invoice_outputs WHERE  project_id = ? AND released_worker_id = ?`, projectID, workerID).Scan(&count).Error
	return count, err
}
