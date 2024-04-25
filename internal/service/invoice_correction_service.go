package service

import (
	"backend-v2/internal/dto"
	"backend-v2/internal/repository"
)

type invoiceCorrectionService struct {
	invoiceObjectRepo    repository.IInvoiceObjectRepository
	invoiceMaterialsRepo repository.IInvoiceMaterialsRepository
  materialLocationRepo repository.IMaterialLocationRepository
}

func InitInvoiceCorrectionService(
	invoiceObjectRepo repository.IInvoiceObjectRepository,
	invoiceMaterialsRepo repository.IInvoiceMaterialsRepository,
  materialLocationRepo repository.IMaterialLocationRepository,
) IInvoiceCorrectionService {
	return &invoiceCorrectionService{
		invoiceObjectRepo:    invoiceObjectRepo,
		invoiceMaterialsRepo: invoiceMaterialsRepo,
    materialLocationRepo: materialLocationRepo,
	}
}

type IInvoiceCorrectionService interface {
	GetAll(projectID uint) ([]dto.InvoiceObjectPaginated, error)
	GetMaterialsFromInvoiceObjectForCorrection(projectID, invoiceID uint) ([]dto.InvoiceMaterialsView, error)
  GetTotalAmounInLocationByTeamName(projectID, materialID uint, teamNumber string) (float64, error)
}

func (service *invoiceCorrectionService) GetAll(projectID uint) ([]dto.InvoiceObjectPaginated, error) {
	return service.invoiceObjectRepo.GetForCorrection(projectID)
}

func (service *invoiceCorrectionService) GetMaterialsFromInvoiceObjectForCorrection(
	projectID, invoiceID uint,
) ([]dto.InvoiceMaterialsView, error) {
	return service.invoiceMaterialsRepo.GetByInvoiceData(projectID, invoiceID, "object")
}

func(service *invoiceCorrectionService) GetTotalAmounInLocationByTeamName(projectID, materialID uint, teamNumber string) (float64, error) {
  return service.materialLocationRepo.GetTotalAmountInTeamsByTeamNumber(projectID, materialID, teamNumber)
}
