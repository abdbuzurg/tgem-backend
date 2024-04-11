package service

import (
	"backend-v2/internal/dto"
	"backend-v2/internal/repository"
	"backend-v2/model"
)

type invoiceObjectService struct {
	invoiceObjectRepo repository.IInvoiceObjectRepository
	objectRepo        repository.IObjectRepository
	workerRepo        repository.IWorkerRepository
	teamRepo          repository.ITeamRepository
}

func InitInvoiceObjectService(
	invoiceObjectRepo repository.IInvoiceObjectRepository,
	objectRepo repository.IObjectRepository,
	workerRepo repository.IWorkerRepository,
	teamRepo repository.ITeamRepository,
) IInvoiceObjectService {
	return &invoiceObjectService{
		invoiceObjectRepo: invoiceObjectRepo,
		objectRepo:        objectRepo,
		workerRepo:        workerRepo,
		teamRepo:          teamRepo,
	}
}

type IInvoiceObjectService interface {
	GetPaginated(limit, page int, projectID, roleID, workerID uint) ([]dto.InvoiceObjectPaginated, error)
	Create(data model.InvoiceObject) (model.InvoiceObject, error)
	Delete(id uint) error
}

func (service *invoiceObjectService) GetPaginated(limit, page int, projectID, roleID, workerID uint) ([]dto.InvoiceObjectPaginated, error) {

	if roleID == 1 {
		workerID = 0
	}
	data, err := service.invoiceObjectRepo.GetPaginated(limit, page, projectID, workerID)
	if err != nil {
		return []dto.InvoiceObjectPaginated{}, err
	}

	result := []dto.InvoiceObjectPaginated{}
	for _, invoice := range data {
		object, err := service.objectRepo.GetByID(invoice.ObjectID)
		if err != nil {
			return []dto.InvoiceObjectPaginated{}, err
		}

		worker, err := service.workerRepo.GetByID(invoice.SupervisorWorkerID)
		if err != nil {
			return []dto.InvoiceObjectPaginated{}, err
		}

		team, err := service.teamRepo.GetByID(invoice.TeamID)
		if err != nil {
			return []dto.InvoiceObjectPaginated{}, err
		}

		result = append(result, dto.InvoiceObjectPaginated{
			ID:           invoice.ID,
			Supervisor:   worker.Name,
			ObjectName:   object.Name,
			TeamName:     team.Number,
			DeliveryCode: invoice.DeliveryCode,
      DateOfInvoice: invoice.DateOfInvoice,
		})
	}

	return result, nil
}

func (service *invoiceObjectService) Create(data model.InvoiceObject) (model.InvoiceObject, error) {
	return service.invoiceObjectRepo.Create(data)
}

func (service *invoiceObjectService) Delete(id uint) error {
	return service.invoiceObjectRepo.Delete(id)
}
