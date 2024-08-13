package service

import (
	"backend-v2/internal/dto"
	"backend-v2/internal/repository"
	"backend-v2/model"
	"backend-v2/pkg/utils"
	"fmt"
	// "github.com/xuri/excelize/v2"
)

type invoiceWriteOffService struct {
	invoiceWriteOffRepo  repository.IInvoiceWriteOffRepository
	workerRepo           repository.IWorkerRepository
	objectRepo           repository.IObjectRepository
	teamRepo             repository.ITeamRepository
	materialLocationRepo repository.IMaterialLocationRepository
	invoiceMaterialsRepo repository.IInvoiceMaterialsRepository
	materialRepo         repository.IMaterialRepository
	materialCostRepo     repository.IMaterialCostRepository
	invoiceCountRepo     repository.IInvoiceCountRepository
}

func InitInvoiceWriteOffService(
	invoiceWriteOffRepo repository.IInvoiceWriteOffRepository,
	workerRepo repository.IWorkerRepository,
	objectRepo repository.IObjectRepository,
	teamRepo repository.ITeamRepository,
	materialLocationRepo repository.IMaterialLocationRepository,
	invoiceMaterialsRepo repository.IInvoiceMaterialsRepository,
	materialRepo repository.IMaterialRepository,
	materialCostRepo repository.IMaterialCostRepository,
	invoiceCountRepo repository.IInvoiceCountRepository,
) IInvoiceWriteOffService {
	return &invoiceWriteOffService{
		invoiceWriteOffRepo:  invoiceWriteOffRepo,
		workerRepo:           workerRepo,
		objectRepo:           objectRepo,
		teamRepo:             teamRepo,
		materialLocationRepo: materialLocationRepo,
		invoiceMaterialsRepo: invoiceMaterialsRepo,
		materialRepo:         materialRepo,
		materialCostRepo:     materialCostRepo,
		invoiceCountRepo:     invoiceCountRepo,
	}
}

type IInvoiceWriteOffService interface {
	GetAll() ([]model.InvoiceWriteOff, error)
	GetPaginated(page, limit int, data dto.InvoiceWriteOffSearchParameters) ([]dto.InvoiceWriteOffPaginated, error)
	GetInvoiceMaterialsWithoutSerialNumbers(id uint) ([]dto.InvoiceMaterialsWithoutSerialNumberView, error)
	GetByID(id uint) (model.InvoiceWriteOff, error)
	Create(data dto.InvoiceWriteOff) (model.InvoiceWriteOff, error)
	Update(data dto.InvoiceWriteOff) (model.InvoiceWriteOff, error)
	Delete(id uint) error
	Count(filter dto.InvoiceWriteOffSearchParameters) (int64, error)
	GetMaterialsForEdit(id uint) ([]dto.InvoiceWriteOffMaterialsForEdit, error)
	Confirmation(id, projectID uint) error
}

func (service *invoiceWriteOffService) GetAll() ([]model.InvoiceWriteOff, error) {
	return service.invoiceWriteOffRepo.GetAll()
}

func (service *invoiceWriteOffService) GetByID(id uint) (model.InvoiceWriteOff, error) {
	return service.invoiceWriteOffRepo.GetByID(id)
}

func (service *invoiceWriteOffService) GetPaginated(page, limit int, data dto.InvoiceWriteOffSearchParameters) ([]dto.InvoiceWriteOffPaginated, error) {
	invoiceWriteOffs, err := service.invoiceWriteOffRepo.GetPaginated(page, limit, data)
	if err != nil {
		return []dto.InvoiceWriteOffPaginated{}, err
	}

	for index, invoiceWriteOff := range invoiceWriteOffs {
		switch invoiceWriteOff.WriteOffType {
		case "writeoff-warehouse":
			break
		case "loss-warehouse":
			break
		case "loss-team":
			team, err := service.teamRepo.GetTeamNumberAndTeamLeadersByID(data.ProjectID, invoiceWriteOff.WriteOffLocationID)
			if err != nil {
				return []dto.InvoiceWriteOffPaginated{}, err
			}

			invoiceWriteOffs[index].WriteOffLocationName = team[0].TeamNumber + " (" + team[0].TeamLeaderName + ")"
			break
		case "loss-object":
			object, err := service.objectRepo.GetByID(invoiceWriteOff.WriteOffLocationID)
			if err != nil {
				return []dto.InvoiceWriteOffPaginated{}, err
			}

			invoiceWriteOffs[index].WriteOffLocationName = object.Name
			break
		default:
			return []dto.InvoiceWriteOffPaginated{}, fmt.Errorf("Обноружен неправильный тип списание %v", invoiceWriteOff.WriteOffType)
		}

	}

	return invoiceWriteOffs, nil
}

func (service *invoiceWriteOffService) Create(data dto.InvoiceWriteOff) (model.InvoiceWriteOff, error) {

	count, err := service.invoiceCountRepo.CountInvoice("writeoff", data.Details.ProjectID)
	if err != nil {
		return model.InvoiceWriteOff{}, err
	}

	data.Details.DeliveryCode = utils.UniqueCodeGeneration("С", int64(count+1), data.Details.ProjectID)

	invoiceMaterials := []model.InvoiceMaterials{}
	for _, invoiceMaterial := range data.Items {
		invoiceMaterialForCreate := model.InvoiceMaterials{
			ID:             0,
			ProjectID:      data.Details.ProjectID,
			MaterialCostID: invoiceMaterial.MaterialCostID,
			InvoiceID:      0,
			InvoiceType:    "writeoff",
			IsDefected:     false,
			Amount:         invoiceMaterial.Amount,
			Notes:          invoiceMaterial.Notes,
		}

		invoiceMaterials = append(invoiceMaterials, invoiceMaterialForCreate)
	}

	invoiceWriteOff, err := service.invoiceWriteOffRepo.Create(dto.InvoiceWriteOffMutationData{
		InvoiceWriteOff:  data.Details,
		InvoiceMaterials: invoiceMaterials,
	})
	if err != nil {
		return model.InvoiceWriteOff{}, err
	}

	return invoiceWriteOff, nil
}

func (service *invoiceWriteOffService) Update(data dto.InvoiceWriteOff) (model.InvoiceWriteOff, error) {
	invoiceMaterials := []model.InvoiceMaterials{}
	for _, invoiceMaterial := range data.Items {
		invoiceMaterialForCreate := model.InvoiceMaterials{
			ID:             0,
			ProjectID:      data.Details.ProjectID,
			MaterialCostID: invoiceMaterial.MaterialCostID,
			InvoiceID:      0,
			InvoiceType:    "writeoff",
			IsDefected:     false,
			Amount:         invoiceMaterial.Amount,
			Notes:          invoiceMaterial.Notes,
		}

		invoiceMaterials = append(invoiceMaterials, invoiceMaterialForCreate)
	}

	invoiceWriteOff, err := service.invoiceWriteOffRepo.Update(dto.InvoiceWriteOffMutationData{
		InvoiceWriteOff:  data.Details,
		InvoiceMaterials: invoiceMaterials,
	})
	if err != nil {
		return model.InvoiceWriteOff{}, err
	}

	return invoiceWriteOff, nil
}

func (service *invoiceWriteOffService) Delete(id uint) error {
	return service.invoiceWriteOffRepo.Delete(id)
}

func (service *invoiceWriteOffService) Count(filter dto.InvoiceWriteOffSearchParameters) (int64, error) {
	return service.invoiceWriteOffRepo.Count(filter)
}

func (service *invoiceWriteOffService) GetInvoiceMaterialsWithoutSerialNumbers(id uint) ([]dto.InvoiceMaterialsWithoutSerialNumberView, error) {
	return service.invoiceMaterialsRepo.GetInvoiceMaterialsWithoutSerialNumbers(id, "writeoff")
}

func (service *invoiceWriteOffService) GetMaterialsForEdit(id uint) ([]dto.InvoiceWriteOffMaterialsForEdit, error) {
	return service.invoiceWriteOffRepo.GetMaterialsForEdit(id)
}

func (service *invoiceWriteOffService) Confirmation(id, projectID uint) error {
	invoiceWriteOff, err := service.invoiceWriteOffRepo.GetByID(id)
	if err != nil {
		return err
	}
	invoiceWriteOff.Confirmation = true

	invoiceMaterials, err := service.invoiceMaterialsRepo.GetByInvoice(invoiceWriteOff.ProjectID, invoiceWriteOff.ID, "writeoff")
	if err != nil {
		return err
	}

	materialsInTheLocation := []model.MaterialLocation{}

	switch invoiceWriteOff.WriteOffType {
	case "writeoff-warehouse":
		materialsInTheLocation, err = service.materialLocationRepo.GetMaterialsInLocationBasedOnInvoiceID(0, "warehouse", id, "writeoff")
		if err != nil {
			return err
		}

		break
	case "loss-warehouse":
		materialsInTheLocation, err = service.materialLocationRepo.GetMaterialsInLocationBasedOnInvoiceID(0, "warehouse", id, "writeoff")
		if err != nil {
			return err
		}

		break
	case "loss-team":
		materialsInTheLocation, err = service.materialLocationRepo.GetMaterialsInLocationBasedOnInvoiceID(0, "team", id, "writeoff")
		if err != nil {
			return err
		}

		break
	case "loss-object":
		materialsInTheLocation, err = service.materialLocationRepo.GetMaterialsInLocationBasedOnInvoiceID(0, "object", id, "writeoff")
		if err != nil {
			return err
		}

		break
	}

	materialsInWriteOffLocation, err := service.materialLocationRepo.GetByLocationType(invoiceWriteOff.WriteOffType)
	if err != nil {
		return err
	}

	for _, invoiceMaterial := range invoiceMaterials {
		indexOfExistingMaterialInLocation := -1
		for index, materialInTheLocation := range materialsInTheLocation {
			if materialInTheLocation.MaterialCostID == invoiceMaterial.MaterialCostID {
				indexOfExistingMaterialInLocation = index
				break
			}
		}

		if indexOfExistingMaterialInLocation == -1 {
			return fmt.Errorf("Ошибка, несанкционированный материал")
		}
		materialsInTheLocation[indexOfExistingMaterialInLocation].Amount -= invoiceMaterial.Amount

		indexOfExistingMaterialInWriteOffLocation := -1
		for index, materialInWriteOffLocation := range materialsInWriteOffLocation {
			if materialInWriteOffLocation.MaterialCostID == invoiceMaterial.MaterialCostID {
				indexOfExistingMaterialInWriteOffLocation = index
				break
			}
		}

		if indexOfExistingMaterialInWriteOffLocation != -1 {
			materialsInWriteOffLocation[indexOfExistingMaterialInWriteOffLocation].Amount += invoiceMaterial.Amount
		} else {
      materialsInWriteOffLocation = append(materialsInWriteOffLocation, model.MaterialLocation{
        ID: 0,
        MaterialCostID: invoiceMaterial.MaterialCostID,
        ProjectID: invoiceWriteOff.ProjectID,
        LocationID: 0,
        LocationType: invoiceWriteOff.WriteOffType,
        Amount: invoiceMaterial.Amount,
      })
    }

    fmt.Println(materialsInTheLocation)
    fmt.Println(materialsInWriteOffLocation)
	}

	err = service.invoiceWriteOffRepo.Confirmation(dto.InvoiceWriteOffConfirmationData{
		InvoiceWriteOff:     invoiceWriteOff,
		MaterialsInLocation: materialsInTheLocation,
		MaterialsInWriteOff: materialsInWriteOffLocation,
	})

	return err
}
