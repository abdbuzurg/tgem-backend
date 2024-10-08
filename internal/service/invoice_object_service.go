package service

import (
	"backend-v2/internal/dto"
	"backend-v2/internal/repository"
	"backend-v2/model"
	"backend-v2/pkg/utils"

)

type invoiceObjectService struct {
	invoiceObjectRepo     repository.IInvoiceObjectRepository
	objectRepo            repository.IObjectRepository
	workerRepo            repository.IWorkerRepository
	teamRepo              repository.ITeamRepository
	materialLocationRepo  repository.IMaterialLocationRepository
	serialNumberRepo      repository.ISerialNumberRepository
	materialCostRepo      repository.IMaterialCostRepository
	invoiceMaterialRepo   repository.IInvoiceMaterialsRepository
	objectTeamsRepo       repository.IObjectTeamsRepository
	operationMaterialRepo repository.IOperationMaterialRepository
  operationRepo repository.IOperationRepository
  materialRepo repository.IMaterialRepository
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
	objectTeamsRepo repository.IObjectTeamsRepository,
	operationMaterialRepo repository.IOperationMaterialRepository,
  operationRepo repository.IOperationRepository,
  materialRepo repository.IMaterialRepository,
) IInvoiceObjectService {
	return &invoiceObjectService{
		invoiceObjectRepo:     invoiceObjectRepo,
		objectRepo:            objectRepo,
		workerRepo:            workerRepo,
		teamRepo:              teamRepo,
		materialLocationRepo:  materialLocation,
		serialNumberRepo:      serialNumberRepo,
		materialCostRepo:      materialCostRepo,
		invoiceMaterialRepo:   invoiceMaterialRepo,
		objectTeamsRepo:       objectTeamsRepo,
		operationMaterialRepo: operationMaterialRepo,
    operationRepo: operationRepo,
    materialRepo: materialRepo,
	}
}

type IInvoiceObjectService interface {
	GetPaginated(limit, page int, projectID uint) ([]dto.InvoiceObjectPaginated, error)
	Create(data dto.InvoiceObjectCreate) (model.InvoiceObject, error)
	Delete(id uint) error
	GetObjects(projectID, userID, roleID uint) ([]model.Object, error)
	GetInvoiceObjectDescriptiveDataByID(id uint) (dto.InvoiceObjectWithMaterialsDescriptive, error)
	GetTeamsMaterials(projectID, teamID uint) ([]dto.InvoiceObjectTeamMaterials, error)
	GetSerialNumberOfMaterial(projectID, materialID uint, locationID uint) ([]string, error)
	GetAvailableMaterialAmount(projectID, materialID, teamID uint) (float64, error)
	Count(projectID uint) (int64, error)
	GetTeamsFromObjectID(objectID uint) ([]dto.TeamDataForSelect, error)
  GetOperationsBasedOnMaterialsInTeamID(projectID, teamID uint) ([]dto.InvoiceObjectOperationsBasedOnTeam, error)
}

func (service *invoiceObjectService) GetInvoiceObjectDescriptiveDataByID(id uint) (dto.InvoiceObjectWithMaterialsDescriptive, error) {
	invoiceData, err := service.invoiceObjectRepo.GetInvoiceObjectDescriptiveDataByID(id)
	if err != nil {
		return dto.InvoiceObjectWithMaterialsDescriptive{}, err
	}

	invoiceMaterialsWithSerialNumberQueryResult, err := service.invoiceMaterialRepo.GetInvoiceMaterialsWithSerialNumbers(id, "object")
	if err != nil {
		return dto.InvoiceObjectWithMaterialsDescriptive{}, err
	}

	invoiceMaterialsWithoutSerailNumber, err := service.invoiceMaterialRepo.GetInvoiceMaterialsWithoutSerialNumbers(id, "object")
	if err != nil {
		return dto.InvoiceObjectWithMaterialsDescriptive{}, err
	}

	invoiceMaterialsWithSerialNumber := []dto.InvoiceMaterialsWithSerialNumberView{}
	current := dto.InvoiceMaterialsWithSerialNumberView{}
	for index, materialInfo := range invoiceMaterialsWithSerialNumberQueryResult {
		if index == 0 {
			current = dto.InvoiceMaterialsWithSerialNumberView{
				ID:            materialInfo.ID,
				MaterialName:  materialInfo.MaterialName,
				MaterialUnit:  materialInfo.MaterialUnit,
				SerialNumbers: []string{},
				Amount:        materialInfo.Amount,
				CostM19:       materialInfo.CostM19,
				Notes:         materialInfo.Notes,
			}
		}

		if current.MaterialName == materialInfo.MaterialName && current.CostM19.Equal(materialInfo.CostM19) {
			if len(current.SerialNumbers) == 0 {
				current.SerialNumbers = append(current.SerialNumbers, materialInfo.SerialNumber)
				continue
			}

			if current.SerialNumbers[len(current.SerialNumbers)-1] != materialInfo.SerialNumber {
				current.SerialNumbers = append(current.SerialNumbers, materialInfo.SerialNumber)
			}

		} else {
			invoiceMaterialsWithSerialNumber = append(invoiceMaterialsWithSerialNumber, current)
			current = dto.InvoiceMaterialsWithSerialNumberView{
				ID:            materialInfo.ID,
				MaterialName:  materialInfo.MaterialName,
				MaterialUnit:  materialInfo.MaterialUnit,
				SerialNumbers: []string{materialInfo.SerialNumber},
				Amount:        materialInfo.Amount,
				CostM19:       materialInfo.CostM19,
				Notes:         materialInfo.Notes,
			}
		}
	}

	if len(invoiceMaterialsWithSerialNumberQueryResult) != 0 {
		invoiceMaterialsWithSerialNumber = append(invoiceMaterialsWithSerialNumber, current)
	}

	return dto.InvoiceObjectWithMaterialsDescriptive{
		InvoiceData:                  invoiceData,
		MaterialsWithSerialNumber:    invoiceMaterialsWithSerialNumber,
		MaterialsWithoutSerialNumber: invoiceMaterialsWithoutSerailNumber,
	}, nil
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

	invoiceMaterialForCreate := []model.InvoiceMaterials{}
	serialNumberMovements := []model.SerialNumberMovement{}
	for _, invoiceMaterial := range data.Items {
		if len(invoiceMaterial.SerialNumbers) == 0 {
			materialInfoSorted, err := service.materialLocationRepo.GetMaterialAmountSortedByCostM19InLocation(data.Details.ProjectID, invoiceMaterial.MaterialID, "team", data.Details.TeamID)
			if err != nil {
				return model.InvoiceObject{}, err
			}

			index := 0
			for invoiceMaterial.Amount > 0 {
				invoiceMaterialCreate := model.InvoiceMaterials{
					ProjectID:      data.Details.ProjectID,
					ID:             0,
					MaterialCostID: materialInfoSorted[index].MaterialCostID,
					InvoiceID:      0,
					InvoiceType:    "object",
					IsDefected:     false,
					Amount:         0,
					Notes:          invoiceMaterial.Notes,
				}

				if materialInfoSorted[index].MaterialAmount <= invoiceMaterial.Amount {
					invoiceMaterialCreate.Amount = materialInfoSorted[index].MaterialAmount
					invoiceMaterial.Amount -= materialInfoSorted[index].MaterialAmount
				} else {
					invoiceMaterialCreate.Amount = invoiceMaterial.Amount
					invoiceMaterial.Amount = 0
				}

				invoiceMaterialForCreate = append(invoiceMaterialForCreate, invoiceMaterialCreate)
				index++
			}
		}

		if len(invoiceMaterial.SerialNumbers) != 0 {
			MC_IDs_AND_SN_IDs, err := service.serialNumberRepo.GetMaterialCostIDsByCodesInLocation(invoiceMaterial.MaterialID, invoiceMaterial.SerialNumbers, "teams", data.Details.TeamID)
			if err != nil {
				return model.InvoiceObject{}, err
			}

			var invoiceMaterialCreate model.InvoiceMaterials
			for index, oneEntry := range MC_IDs_AND_SN_IDs {

				serialNumberMovements = append(serialNumberMovements, model.SerialNumberMovement{
					ID:             0,
					SerialNumberID: oneEntry.SerialNumberID,
					ProjectID:      data.Details.ProjectID,
					InvoiceID:      0,
					InvoiceType:    "object",
					Confirmation:   false,
				})

				if index == 0 {
					invoiceMaterialCreate = model.InvoiceMaterials{
						ProjectID:      data.Details.ProjectID,
						ID:             0,
						MaterialCostID: oneEntry.MaterialCostID,
						InvoiceID:      data.Details.ID,
						InvoiceType:    "object",
						IsDefected:     false,
						Amount:         0,
						Notes:          invoiceMaterial.Notes,
					}
				}

				if oneEntry.MaterialCostID == invoiceMaterialCreate.MaterialCostID {
					invoiceMaterialCreate.Amount++
				} else {
					invoiceMaterialForCreate = append(invoiceMaterialForCreate, invoiceMaterialCreate)
					invoiceMaterialCreate = model.InvoiceMaterials{
						ProjectID:      data.Details.ProjectID,
						ID:             0,
						MaterialCostID: oneEntry.MaterialCostID,
						InvoiceID:      data.Details.ID,
						InvoiceType:    "object",
						IsDefected:     false,
						Amount:         0,
						Notes:          invoiceMaterial.Notes,
					}
				}
			}

			invoiceMaterialForCreate = append(invoiceMaterialForCreate, invoiceMaterialCreate)
		}
	}

  invoiceOperationsForCreate := []model.InvoiceOperations{}
  for _, invoiceOperation := range data.Operations{
    invoiceOperationsForCreate = append(invoiceOperationsForCreate, model.InvoiceOperations{
      ID: 0,
      ProjectID: data.Details.ProjectID,
      OperationID: invoiceOperation.OperationID,
      InvoiceID: 0,
      InvoiceType: "object",
      Amount: invoiceOperation.Amount,
      Notes: invoiceOperation.Notes,
    })
  }

	invoiceObject, err := service.invoiceObjectRepo.Create(dto.InvoiceObjectCreateQueryData{
		Invoice:               data.Details,
		InvoiceMaterials:      invoiceMaterialForCreate,
		InvoiceOperations:      invoiceOperationsForCreate,
		SerialNumberMovements: serialNumberMovements,
	})

	return invoiceObject, err
}

func (service *invoiceObjectService) Delete(id uint) error {
	return service.invoiceObjectRepo.Delete(id)
}

func (service *invoiceObjectService) GetObjects(projectID, userID, roleID uint) ([]model.Object, error) {

	result := []model.Object{}
	return result, nil
}

func (service *invoiceObjectService) GetTeamsMaterials(projectID, teamID uint) ([]dto.InvoiceObjectTeamMaterials, error) {
  materialsInTeam, err := service.materialLocationRepo.GetUniqueMaterialsFromLocation(projectID, teamID, "team")
  if err != nil {
    return []dto.InvoiceObjectTeamMaterials{}, err
  }

  result := []dto.InvoiceObjectTeamMaterials{}
  for _, entry := range materialsInTeam {
    amount, err := service.materialLocationRepo.GetTotalAmountInLocation(projectID, entry.ID, teamID, "team")
    if err != nil {
      return []dto.InvoiceObjectTeamMaterials{}, err
    }

    result = append(result, dto.InvoiceObjectTeamMaterials{
      MaterialID: entry.ID,
      MaterialName: entry.Name,
      MaterialUnit: entry.Unit,
      HasSerialNumber: entry.HasSerialNumber,
      Amount: amount,
    })
  }
  
	return result, nil 
}

func (service *invoiceObjectService) GetSerialNumberOfMaterial(projectID, materialID uint, locationID uint) ([]string, error) {
	return service.serialNumberRepo.GetCodesByMaterialIDAndLocation(projectID, materialID, "teams", locationID)
}

func (service *invoiceObjectService) GetAvailableMaterialAmount(projectID, materialID, teamID uint) (float64, error) {
	return service.materialLocationRepo.GetTotalAmountInLocation(projectID, materialID, teamID, "team")
}

func (service *invoiceObjectService) Count(projectID uint) (int64, error) {
	return service.invoiceObjectRepo.Count(projectID)
}

func (service *invoiceObjectService) GetTeamsFromObjectID(objectID uint) ([]dto.TeamDataForSelect, error) {
	return service.objectTeamsRepo.GetTeamsByObjectID(objectID)
}

func (service *invoiceObjectService) GetOperationsBasedOnMaterialsInTeamID(projectID, teamID uint) ([]dto.InvoiceObjectOperationsBasedOnTeam, error) {
  result := []dto.InvoiceObjectOperationsBasedOnTeam{}

  operationsAvailableForTeam, err := service.invoiceObjectRepo.GetOperationsBasedOnMaterialsInTeam(teamID)
  if err != nil {
    return result, err
  }

  operationWithoutMaterials, err := service.operationRepo.GetWithoutMaterialOperations(projectID)
  if err != nil {
    return result, err
  }

  for _, operation := range operationsAvailableForTeam {
    operationMaterial, err := service.operationMaterialRepo.GetByOperationID(operation.ID)
    if err != nil {
      return []dto.InvoiceObjectOperationsBasedOnTeam{}, err
    }

    material, err := service.materialRepo.GetByID(operationMaterial.MaterialID)
    if err != nil {
      return []dto.InvoiceObjectOperationsBasedOnTeam{}, err
    }

    result = append(result, dto.InvoiceObjectOperationsBasedOnTeam{
      OperationID: operation.ID,
      OperationName: operation.Name,
      MaterialID: operationMaterial.MaterialID,
      MaterialName: material.Name,
    })
  }

  for _, operation := range operationWithoutMaterials {
    result = append(result, dto.InvoiceObjectOperationsBasedOnTeam{
      OperationID: operation.ID,
      OperationName: operation.Name,
      MaterialID: 0,
      MaterialName: "",
    })
  }

  return result, nil
}
