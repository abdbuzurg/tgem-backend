package service

import (
	"backend-v2/internal/dto"
	"backend-v2/internal/repository"
	"backend-v2/model"
	"backend-v2/pkg/utils"
)

type invoiceOutputOutOfProjectService struct {
	invoiceOutputOutOfProjectRepo repository.IInvoiceOutputOutOfProjectRepository
	invoiceOutputRepo             repository.IInvoiceOutputRepository
	materialLocationRepo          repository.IMaterialLocationRepository
}

func InitInvoiceOutputOutOfProjectService(
	invoiceOutputOutOfProjectRepo repository.IInvoiceOutputOutOfProjectRepository,
	invoiceOutputRepo repository.IInvoiceOutputRepository,
	materialLocationRepo repository.IMaterialLocationRepository,
) IInvoiceOutputOutOfProjectService {
	return &invoiceOutputOutOfProjectService{
		invoiceOutputOutOfProjectRepo: invoiceOutputOutOfProjectRepo,
		invoiceOutputRepo:             invoiceOutputRepo,
		materialLocationRepo:          materialLocationRepo,
	}
}

type IInvoiceOutputOutOfProjectService interface {
	GetPaginated(page, limit int, data model.InvoiceOutputOutOfProject) ([]dto.InvoiceOutputOutOfProjectPaginated, error)
	Count(projectID uint) (int64, error)
}

func (service *invoiceOutputOutOfProjectService) GetPaginated(page, limit int, data model.InvoiceOutputOutOfProject) ([]dto.InvoiceOutputOutOfProjectPaginated, error) {
	return service.invoiceOutputOutOfProjectRepo.GetPaginated(page, limit, data)
}

func (service *invoiceOutputOutOfProjectService) Count(projectID uint) (int64, error) {
	return service.invoiceOutputOutOfProjectRepo.Count(projectID)
}

func (service *invoiceOutputOutOfProjectService) Create(data dto.InvoiceOutputOutOfProject) (model.InvoiceOutputOutOfProject, error) {
	countInProject, err := service.invoiceOutputRepo.Count(data.Details.ProjectID)
	if err != nil {
		return model.InvoiceOutputOutOfProject{}, err
	}

	countOutOfProject, err := service.invoiceOutputOutOfProjectRepo.Count(data.Details.ProjectID)
	if err != nil {
		return model.InvoiceOutputOutOfProject{}, err
	}

	data.Details.DeliveryCode = utils.UniqueCodeGeneration("О", countInProject+countOutOfProject+1, data.Details.ProjectID)

	invoiceMaterialForCreate := []model.InvoiceMaterials{}
	serialNumberMovements := []model.SerialNumberMovement{}
	for _, invoiceMaterial := range data.Items {
		if len(invoiceMaterial.SerialNumbers) == 0 {
		}

		if len(invoiceMaterial.SerialNumbers) != 0 {
		}

	}

	invoiceOutput, err := service.invoiceOutputOutOfProjectRepo.Create(dto.InvoiceOutputOutOfProjectCreateQueryData{
		Invoice:               data.Details,
		InvoiceMaterials:      invoiceMaterialForCreate,
		SerialNumberMovements: serialNumberMovements,
	})
	if err != nil {
		return model.InvoiceOutputOutOfProject{}, err
	}

	return invoiceOutput, nil
}
