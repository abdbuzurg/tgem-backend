package service

import (
	"backend-v2/internal/dto"
	"backend-v2/internal/repository"
	"backend-v2/model"
)

type invoiceObjectSubmissionService struct {
	invoiceObjectSubmissionRepo repository.IInvoiceObjectSubmissionRepository
	objectRepo                  repository.IObjectRepository
	workerRepo                  repository.IWorkerRepository
}

func InitInvoiceObjectSubmissionService(
	invoiceObjectSubmissionRepo repository.IInvoiceObjectSubmissionRepository,
	objectRepo repository.IObjectRepository,
	workerRepo repository.IWorkerRepository,
) IInvoiceObjectSubmissionService {
	return &invoiceObjectSubmissionService{
		invoiceObjectSubmissionRepo: invoiceObjectSubmissionRepo,
	}
}

type IInvoiceObjectSubmissionService interface {
	GetPaginated(limit, page int, projectID uint) ([]dto.InvoiceObjectSubmissionPaginated, error)
	Create(data model.InvoiceObjectSubmission) (model.InvoiceObjectSubmission, error)
	Delete(id uint) error
}

func (service *invoiceObjectSubmissionService) GetPaginated(limit, page int, projectID uint) ([]dto.InvoiceObjectSubmissionPaginated, error) {
	data, err := service.invoiceObjectSubmissionRepo.GetPaginated(limit, page, projectID)
	if err != nil {
		return []dto.InvoiceObjectSubmissionPaginated{}, err
	}

	result := []dto.InvoiceObjectSubmissionPaginated{}
	for _, invoice := range data {
		object, err := service.objectRepo.GetByID(invoice.ObjectID)
		if err != nil {
			return []dto.InvoiceObjectSubmissionPaginated{}, err
		}

		worker, err := service.workerRepo.GetByID(invoice.SuperwisorWorkerID)
		if err != nil {
			return []dto.InvoiceObjectSubmissionPaginated{}, err
		}

		result = append(result, dto.InvoiceObjectSubmissionPaginated{
			ID:             invoice.ID,
			Supervisor:     worker.Name,
			ObjectName:     object.Name,
			ApprovalStatus: invoice.ApprovalStatus,
		})
	}

	return result, nil
}

func (service *invoiceObjectSubmissionService) Create(data model.InvoiceObjectSubmission) (model.InvoiceObjectSubmission, error) {
	return service.invoiceObjectSubmissionRepo.Create(data)
}

func (service *invoiceObjectSubmissionService) Delete(id uint) error {
	return service.invoiceObjectSubmissionRepo.Delete(id)
}
