package service

import (
	"backend-v2/internal/dto"
	"backend-v2/internal/repository"
)

type invoiceCorrectionService struct {
	invoiceCorrection    repository.IInvoiceCorrectionRepository
	invoiceObjectRepo    repository.IInvoiceObjectRepository
	invoiceMaterialsRepo repository.IInvoiceMaterialsRepository
	materialLocationRepo repository.IMaterialLocationRepository
}

func InitInvoiceCorrectionService(
	invoiceCorrection repository.IInvoiceCorrectionRepository,
	invoiceObjectRepo repository.IInvoiceObjectRepository,
	invoiceMaterialsRepo repository.IInvoiceMaterialsRepository,
	materialLocationRepo repository.IMaterialLocationRepository,
) IInvoiceCorrectionService {
	return &invoiceCorrectionService{
		invoiceCorrection:    invoiceCorrection,
		invoiceObjectRepo:    invoiceObjectRepo,
		invoiceMaterialsRepo: invoiceMaterialsRepo,
		materialLocationRepo: materialLocationRepo,
	}
}

type IInvoiceCorrectionService interface {
	GetAll(projectID uint) ([]dto.InvoiceObjectPaginated, error)
	GetTotalAmounInLocationByTeamName(projectID, materialID uint, teamNumber string) (float64, error)
	GetInvoiceMaterialsByInvoiceObjectID(id uint) ([]dto.InvoiceCorrectionMaterialsData, error)
  GetSerialNumberOfMaterialInTeam(projectID uint, materialID uint, teamNumber string) ([]string, error)
}

func (service *invoiceCorrectionService) GetAll(projectID uint) ([]dto.InvoiceObjectPaginated, error) {
	return service.invoiceObjectRepo.GetForCorrection(projectID)
}

func (service *invoiceCorrectionService) GetTotalAmounInLocationByTeamName(projectID, materialID uint, teamNumber string) (float64, error) {
	return service.materialLocationRepo.GetTotalAmountInTeamsByTeamNumber(projectID, materialID, teamNumber)
}

func (service *invoiceCorrectionService) GetInvoiceMaterialsByInvoiceObjectID(id uint) ([]dto.InvoiceCorrectionMaterialsData, error) {
	return service.invoiceCorrection.GetInvoiceMaterialsDataByInvoiceObjectID(id)
}

func (service *invoiceCorrectionService) GetSerialNumberOfMaterialInTeam(projectID uint, materialID uint, teamNumber string) ([]string, error) {
  return service.invoiceCorrection.GetSerialNumberOfMaterialInTeam(projectID, materialID, teamNumber)
}
