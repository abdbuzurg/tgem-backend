package service

import (
	"backend-v2/internal/dto"
	"backend-v2/internal/repository"
	"backend-v2/model"
	"backend-v2/pkg/utils"
)

type invoiceObjectService struct {
	invoiceObjectRepo    repository.IInvoiceObjectRepository
	objectRepo           repository.IObjectRepository
	workerRepo           repository.IWorkerRepository
	teamRepo             repository.ITeamRepository
	materialLocationRepo repository.IMaterialLocationRepository
	serialNumberRepo     repository.ISerialNumberRepository
	materialCostRepo     repository.IMaterialCostRepository
	invoiceMaterialRepo  repository.IInvoiceMaterialsRepository
}

func InitInvoiceObjectService(
	invoiceObjectRepo repository.IInvoiceObjectRepository,
	objectRepo repository.IObjectRepository,
	workerRepo repository.IWorkerRepository,
	teamRepo repository.ITeamRepository,
	materialLocation repository.IMaterialLocationRepository,
	serialNumberRepo repository.ISerialNumberRepository,
	materialCostRepo repository.IMaterialCostRepository,
	invoiceMaterialRepo repository.IInvoiceMaterialsRepository,
) IInvoiceObjectService {
	return &invoiceObjectService{
		invoiceObjectRepo:    invoiceObjectRepo,
		objectRepo:           objectRepo,
		workerRepo:           workerRepo,
		teamRepo:             teamRepo,
		materialLocationRepo: materialLocation,
		serialNumberRepo:     serialNumberRepo,
		materialCostRepo:     materialCostRepo,
		invoiceMaterialRepo:  invoiceMaterialRepo,
	}
}

type IInvoiceObjectService interface {
	GetPaginated(limit, page int, projectID uint) ([]dto.InvoiceObjectPaginated, error)
	Create(data dto.InvoiceObjectCreate) (model.InvoiceObject, error)
	Delete(id uint) error
	GetObjects(projectID, userID, roleID uint) ([]model.Object, error)
	GetTeamsMaterials(projectID, teamID uint) ([]model.Material, error)
	GetSerialNumberOfMaterial(projectID, materialID uint) ([]string, error)
	GetAvailableMaterialAmount(projectID, materialID, teamID uint) (float64, error)
  Count(projectID uint) (int64, error)
  GetInvoiceObjectFullData(projectID, id uint) (dto.InvoiceObjectFullData, error)
}

func (service *invoiceObjectService) GetPaginated(limit, page int, projectID uint) ([]dto.InvoiceObjectPaginated, error) {
  return service.invoiceObjectRepo.GetPaginated(page, limit, projectID)
}

func (service *invoiceObjectService) Create(data dto.InvoiceObjectCreate) (model.InvoiceObject, error) {

	count, err := service.invoiceObjectRepo.Count(data.Details.ProjectID)
	if err != nil {
		return model.InvoiceObject{}, err
	}

	code := utils.UniqueCodeGeneration("ПО", count+1, data.Details.ProjectID)
	data.Details.DeliveryCode = code

	invoiceObject, err := service.invoiceObjectRepo.Create(data.Details)
	if err != nil {
		return model.InvoiceObject{}, err
	}
	data.Details = invoiceObject

	for _, invoiceMaterial := range data.Items {
		materialCosts, err := service.materialCostRepo.GetByMaterialIDSorted(invoiceMaterial.MaterialID)
		if err != nil {
			return model.InvoiceObject{}, err
		}

		materialLocations := []model.MaterialLocation{}

		for _, materialCost := range materialCosts {
			materialLocation, err := service.materialLocationRepo.GetByMaterialCostIDOrCreate(invoiceObject.ProjectID, materialCost.ID, "teams", invoiceObject.TeamID)
			if err != nil {
				return model.InvoiceObject{}, err
			}

			materialLocations = append(materialLocations, materialLocation)
		}

		index := 0
		for invoiceMaterial.Amount > 0 {
			invoiceMaterialCreate := model.InvoiceMaterials{
				ProjectID:      invoiceObject.ProjectID,
				ID:             0,
				MaterialCostID: materialCosts[index].ID,
				InvoiceID:      invoiceObject.ID,
				InvoiceType:    "object",
				IsDefected:     false,
				Amount:         0,
				Notes:          "",
			}
			if materialLocations[index].Amount <= invoiceMaterial.Amount {
				invoiceMaterialCreate.Amount = materialLocations[index].Amount
				invoiceMaterial.Amount -= materialLocations[index].Amount
				materialLocations[index].Amount = 0
			} else {
				materialLocations[index].Amount -= invoiceMaterial.Amount
				invoiceMaterialCreate.Amount = invoiceMaterial.Amount
				invoiceMaterial.Amount = 0
			}

			invoiceMaterialCreate, err = service.invoiceMaterialRepo.Create(invoiceMaterialCreate)
			if err != nil {
				return model.InvoiceObject{}, err
			}

			index++
		}

	}

	return data.Details, nil
}

func (service *invoiceObjectService) Delete(id uint) error {
	return service.invoiceObjectRepo.Delete(id)
}

func (service *invoiceObjectService) GetObjects(projectID, userID, roleID uint) ([]model.Object, error) {

	result := []model.Object{}
	return result, nil
}

func (service *invoiceObjectService) GetTeamsMaterials(projectID, teamID uint) ([]model.Material, error) {
	return service.materialLocationRepo.GetUniqueMaterialsFromLocation(projectID, teamID, "teams")
}

func (service *invoiceObjectService) GetSerialNumberOfMaterial(projectID, materialID uint) ([]string, error) {
	return service.serialNumberRepo.GetCodesByMaterialIDAndStatus(projectID, materialID, "teams")
}

func (service *invoiceObjectService) GetAvailableMaterialAmount(projectID, materialID, teamID uint) (float64, error) {
	return service.materialLocationRepo.GetTotalAmountInLocation(projectID, materialID, teamID, "teams")
}

func(service *invoiceObjectService) Count(projectID uint) (int64, error) {
  return service.invoiceObjectRepo.Count(projectID)
} 

func (service *invoiceObjectService) GetInvoiceObjectFullData(projectID, id uint) (dto.InvoiceObjectFullData, error) {
  invoiceObject, err := service.invoiceObjectRepo.GetByID(id)
  if err != nil {
    return dto.InvoiceObjectFullData{}, err
  }

  invoiceObjectMaterials, err := service.invoiceMaterialRepo.GetByInvoiceData(projectID, invoiceObject.ID, "object")
  if err != nil {
    return dto.InvoiceObjectFullData{}, err
  }

  result := dto.InvoiceObjectFullData{
    Details: invoiceObject,
    Items: []dto.InvoiceObjectFullDataItem{},
  }

  for _, invoiceMaterial := range invoiceObjectMaterials {
    result.Items = append(result.Items, dto.InvoiceObjectFullDataItem{
      ID: invoiceMaterial.ID,
      MaterialName: invoiceObject.ObjectName,
      Amount: invoiceMaterial.Amount,
      Notes: invoiceMaterial.Notes,
    })
  }

  return result, nil
}
