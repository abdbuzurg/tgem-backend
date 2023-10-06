package service

import (
	"backend-v2/internal/repository"
	"backend-v2/model"
	"backend-v2/pkg/utils"
)

type invoiceInvoiceMaterialsService struct {
	invoiceMaterialsRepo repository.IInvoiceMaterialsRepository
}

func InitInvoiceMaterialsService(invoiceMaterialsRepo repository.IInvoiceMaterialsRepository) IInvoiceMaterialsService {
	return &invoiceInvoiceMaterialsService{
		invoiceMaterialsRepo: invoiceMaterialsRepo,
	}
}

type IInvoiceMaterialsService interface {
	GetAll() ([]model.InvoiceMaterials, error)
	GetPaginated(page, limit int, data model.InvoiceMaterials) ([]model.InvoiceMaterials, error)
	GetByID(id uint) (model.InvoiceMaterials, error)
	Create(data model.InvoiceMaterials) (model.InvoiceMaterials, error)
	Update(data model.InvoiceMaterials) (model.InvoiceMaterials, error)
	Delete(id uint) error
	Count() (int64, error)
}

func (service *invoiceInvoiceMaterialsService) GetAll() ([]model.InvoiceMaterials, error) {
	return service.invoiceMaterialsRepo.GetAll()
}

func (service *invoiceInvoiceMaterialsService) GetPaginated(page, limit int, data model.InvoiceMaterials) ([]model.InvoiceMaterials, error) {
	if !(utils.IsEmptyFields(data)) {
		return service.invoiceMaterialsRepo.GetPaginatedFiltered(page, limit, data)
	}

	return service.invoiceMaterialsRepo.GetPaginated(page, limit)
}

func (service *invoiceInvoiceMaterialsService) GetByID(id uint) (model.InvoiceMaterials, error) {
	return service.invoiceMaterialsRepo.GetByID(id)
}

func (service *invoiceInvoiceMaterialsService) Create(data model.InvoiceMaterials) (model.InvoiceMaterials, error) {
	return service.invoiceMaterialsRepo.Create(data)
}

func (service *invoiceInvoiceMaterialsService) Update(data model.InvoiceMaterials) (model.InvoiceMaterials, error) {
	return service.invoiceMaterialsRepo.Update(data)
}

func (service *invoiceInvoiceMaterialsService) Delete(id uint) error {
	return service.invoiceMaterialsRepo.Delete(id)
}

func (service *invoiceInvoiceMaterialsService) Count() (int64, error) {
	return service.invoiceMaterialsRepo.Count()
}
