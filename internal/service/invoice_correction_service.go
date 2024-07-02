package service

import (
	"backend-v2/internal/dto"
	"backend-v2/internal/repository"
	"backend-v2/model"
)

type invoiceCorrectionService struct {
	invoiceCorrectionRepo repository.IInvoiceCorrectionRepository
	invoiceObjectRepo     repository.IInvoiceObjectRepository
	invoiceMaterialsRepo  repository.IInvoiceMaterialsRepository
	materialLocationRepo  repository.IMaterialLocationRepository
}

func InitInvoiceCorrectionService(
	invoiceCorrection repository.IInvoiceCorrectionRepository,
	invoiceObjectRepo repository.IInvoiceObjectRepository,
	invoiceMaterialsRepo repository.IInvoiceMaterialsRepository,
	materialLocationRepo repository.IMaterialLocationRepository,
) IInvoiceCorrectionService {
	return &invoiceCorrectionService{
		invoiceCorrectionRepo: invoiceCorrection,
		invoiceObjectRepo:     invoiceObjectRepo,
		invoiceMaterialsRepo:  invoiceMaterialsRepo,
		materialLocationRepo:  materialLocationRepo,
	}
}

type IInvoiceCorrectionService interface {
	GetAll(projectID uint) ([]dto.InvoiceCorrectionPaginated, error)
	GetTotalAmounInLocationByTeamName(projectID, materialID uint, teamNumber string) (float64, error)
	GetInvoiceMaterialsByInvoiceObjectID(id uint) ([]dto.InvoiceCorrectionMaterialsData, error)
	GetSerialNumberOfMaterialInTeam(projectID uint, materialID uint, teamID uint) ([]string, error)
	Create(data dto.InvoiceCorrectionCreate) (model.InvoiceObject, error)
}

func (service *invoiceCorrectionService) GetAll(projectID uint) ([]dto.InvoiceCorrectionPaginated, error) {
	return service.invoiceObjectRepo.GetForCorrection(projectID)
}

func (service *invoiceCorrectionService) GetTotalAmounInLocationByTeamName(projectID, materialID uint, teamNumber string) (float64, error) {
	return service.materialLocationRepo.GetTotalAmountInTeamsByTeamNumber(projectID, materialID, teamNumber)
}

func (service *invoiceCorrectionService) GetInvoiceMaterialsByInvoiceObjectID(id uint) ([]dto.InvoiceCorrectionMaterialsData, error) {
	return service.invoiceCorrectionRepo.GetInvoiceMaterialsDataByInvoiceObjectID(id)
}

func (service *invoiceCorrectionService) GetSerialNumberOfMaterialInTeam(projectID uint, materialID uint, teamID uint) ([]string, error) {
	return service.invoiceCorrectionRepo.GetSerialNumberOfMaterialInTeam(projectID, materialID, teamID)
}

func (service *invoiceCorrectionService) Create(data dto.InvoiceCorrectionCreate) (model.InvoiceObject, error) {

	invoiceObject, err := service.invoiceObjectRepo.GetByID(data.Details.ID)
	if err != nil {
		return model.InvoiceObject{}, err
	}

	invoiceObject.ConfirmedByOperator = true
	invoiceObject.DateOfCorrection = data.Details.DateOfCorrection

	invoiceMaterialForCreate := []model.InvoiceMaterials{}

	for _, invoiceMaterial := range data.Items {
		materialInfoSorted, err := service.materialLocationRepo.GetMaterialAmountSortedByCostM19InLocation(invoiceObject.ProjectID, invoiceMaterial.MaterialID, "team", invoiceObject.TeamID)
		if err != nil {
			return model.InvoiceObject{}, err
		}

		index := 0
		for invoiceMaterial.MaterialAmount > 0 {
			invoiceMaterialCreate := model.InvoiceMaterials{
				ProjectID:      invoiceObject.ProjectID,
				ID:             0,
				MaterialCostID: materialInfoSorted[index].MaterialCostID,
				InvoiceID:      0,
				InvoiceType:    "correction",
				IsDefected:     false,
				Amount:         0,
				Notes:          invoiceMaterial.Notes,
			}

			if materialInfoSorted[index].MaterialAmount <= invoiceMaterial.MaterialAmount {
				invoiceMaterialCreate.Amount = materialInfoSorted[index].MaterialAmount
				invoiceMaterial.MaterialAmount -= materialInfoSorted[index].MaterialAmount
			} else {
				invoiceMaterialCreate.Amount = invoiceMaterial.MaterialAmount
				invoiceMaterial.MaterialAmount = 0
			}

			invoiceMaterialForCreate = append(invoiceMaterialForCreate, invoiceMaterialCreate)
			index++
		}

	}

	toBeUpdatedTeamLocations := []model.MaterialLocation{}
	toBeUpdatedObjectLocations := []model.MaterialLocation{}
	for _, invoiceMaterial := range invoiceMaterialForCreate {
		materialInTeamLocation, err := service.materialLocationRepo.GetByMaterialCostIDOrCreate(
			invoiceObject.ProjectID,
			invoiceMaterial.MaterialCostID,
			"team",
			invoiceObject.TeamID,
		)
		if err != nil {
			return model.InvoiceObject{}, err
		}

		materialInObjectLocation, err := service.materialLocationRepo.GetByMaterialCostIDOrCreate(
			invoiceObject.ProjectID,
			invoiceMaterial.MaterialCostID,
			"object",
			invoiceObject.ObjectID,
		)
    if err != nil {
      return model.InvoiceObject{}, err
    }

    materialInTeamLocation.Amount -= invoiceMaterial.Amount
	  materialInObjectLocation.Amount += invoiceMaterial.Amount

    toBeUpdatedTeamLocations = append(toBeUpdatedTeamLocations, materialInTeamLocation)
    toBeUpdatedObjectLocations = append(toBeUpdatedObjectLocations, materialInObjectLocation)
  }

	result, err := service.invoiceCorrectionRepo.Create(dto.InvoiceCorrectionCreateQuery{
		Details: invoiceObject,
		Items:   invoiceMaterialForCreate,
    TeamLocation: toBeUpdatedTeamLocations,
    ObjectLocation: toBeUpdatedObjectLocations,
	})

	return result, nil
}
