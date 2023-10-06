package service

import (
	"backend-v2/internal/repository"
	"backend-v2/model"
	"backend-v2/pkg/utils"
)

type invoiceService struct {
	invoiceRepo repository.IInvoiceRepository
}

func InitInvoiceService(invoiceRepo repository.IInvoiceRepository) IInvoiceService {
	return &invoiceService{
		invoiceRepo: invoiceRepo,
	}
}

type IInvoiceService interface {
	GetAll() ([]model.Invoice, error)
	GetPaginated(page, limit int, data model.Invoice) ([]model.Invoice, error)
	GetByID(id uint) (model.Invoice, error)
	Create(data model.Invoice) (model.Invoice, error)
	Update(data model.Invoice) (model.Invoice, error)
	Delete(id uint) error
	Count() (int64, error)
}

func (service *invoiceService) GetAll() ([]model.Invoice, error) {
	return service.invoiceRepo.GetAll()
}

func (service *invoiceService) GetPaginated(page, limit int, data model.Invoice) ([]model.Invoice, error) {
	if !(utils.IsEmptyFields(data)) {
		return service.invoiceRepo.GetPaginatedFiltered(page, limit, data)
	}

	return service.invoiceRepo.GetPaginated(page, limit)
}

func (service *invoiceService) GetByID(id uint) (model.Invoice, error) {
	return service.invoiceRepo.GetByID(id)
}

func (service *invoiceService) Create(data model.Invoice) (model.Invoice, error) {
	return service.invoiceRepo.Create(data)
}

func (service *invoiceService) Update(data model.Invoice) (model.Invoice, error) {
	return service.invoiceRepo.Update(data)
}

func (service *invoiceService) Delete(id uint) error {
	return service.invoiceRepo.Delete(id)
}

func (service *invoiceService) Count() (int64, error) {
	return service.invoiceRepo.Count()
}
