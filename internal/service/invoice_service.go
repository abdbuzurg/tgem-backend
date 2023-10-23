package service

import (
	"backend-v2/internal/dto"
	"backend-v2/internal/repository"
	"backend-v2/model"
	"backend-v2/pkg/utils"
	"fmt"
)

type invoiceService struct {
	projectRepo          repository.IProjectRepository
	invoiceRepo          repository.IInvoiceRepository
	invoiceMaterialsRepo repository.IInvoiceMaterialsRepository
	objectRepo           repository.IObjectRepository
	materialCostRepo     repository.IMaterialCostRepository
	materialLocationRepo repository.IMaterialLocationRepository
	materialRepo         repository.IMaterialRepository
	workerRepo           repository.IWorkerRepository
	teamRepo             repository.ITeamRepository
}

func InitInvoiceService(
	projectRepo repository.IProjectRepository,
	invoiceRepo repository.IInvoiceRepository,
	invoiceMaterialsRepo repository.IInvoiceMaterialsRepository,
	objectRepo repository.IObjectRepository,
	materialCostRepo repository.IMaterialCostRepository,
	materialLocationRepo repository.IMaterialLocationRepository,
	materialRepo repository.IMaterialRepository,
	workerRepo repository.IWorkerRepository,
	teamRepo repository.ITeamRepository,

) IInvoiceService {
	return &invoiceService{
		projectRepo:          projectRepo,
		invoiceRepo:          invoiceRepo,
		invoiceMaterialsRepo: invoiceMaterialsRepo,
		materialCostRepo:     materialCostRepo,
		materialLocationRepo: materialLocationRepo,
		materialRepo:         materialRepo,
		objectRepo:           objectRepo,
		workerRepo:           workerRepo,
		teamRepo:             teamRepo,
	}
}

type IInvoiceService interface {
	GetAll(invoiceType string) ([]model.Invoice, error)
	GetPaginated(invoiceType string, page, limit int, data model.Invoice) ([]dto.InvoicePaginatedData, error)
	GetByID(id uint) (dto.InvoiceFullData, error)
	Create(data dto.InvoiceDataUpdateOrCreate) (bool, error)
	Update(data dto.InvoiceDataUpdateOrCreate) (bool, error)
	Delete(id uint) error
	Count(invoiceType string) (int64, error)
}

func (service *invoiceService) GetAll(invoiceType string) ([]model.Invoice, error) {
	return service.invoiceRepo.GetAll(invoiceType)
}

func (service *invoiceService) GetPaginated(invoiceType string, page, limit int, filter model.Invoice) ([]dto.InvoicePaginatedData, error) {
	var paginatedInvoiceData []model.Invoice
	var err error
	if !(utils.IsEmptyFields(filter)) {
		paginatedInvoiceData, err = service.invoiceRepo.GetPaginatedFiltered(invoiceType, page, limit, filter)
	} else {
		paginatedInvoiceData, err = service.invoiceRepo.GetPaginated(invoiceType, page, limit)
	}

	if err != nil {
		return []dto.InvoicePaginatedData{}, err
	}

	data := []dto.InvoicePaginatedData{}
	for _, invoice := range paginatedInvoiceData {
		warehouseManager, err := service.workerRepo.GetByID(invoice.WarehouseManagerWorkerID)
		if err != nil {
			return []dto.InvoicePaginatedData{}, err
		}

		released, err := service.workerRepo.GetByID(invoice.ReleasedWorkerID)
		if err != nil {
			return []dto.InvoicePaginatedData{}, err
		}

		object, err := service.objectRepo.GetByID(invoice.ObjectID)
		if err != nil {
			return []dto.InvoicePaginatedData{}, err
		}

		data = append(data, dto.InvoicePaginatedData{
			ID:                   invoice.ID,
			WarehouseManagerName: warehouseManager.Name,
			ReleasedName:         released.Name,
			ObjectName:           object.Name,
			DateOfInvoice:        invoice.DateOfInvoice.Format("2006-01-02"),
		})
	}

	return data, nil
}

func (service *invoiceService) GetByID(id uint) (dto.InvoiceFullData, error) {
	invoiceData, err := service.invoiceRepo.GetByID(id)
	if err != nil {
		return dto.InvoiceFullData{}, err
	}

	project, err := service.projectRepo.GetByID(invoiceData.ProjectID)
	if err != nil {
		return dto.InvoiceFullData{}, err
	}

	team, err := service.teamRepo.GetByID(invoiceData.TeamID)
	if err != nil {
		return dto.InvoiceFullData{}, err
	}

	warehouseManager, err := service.workerRepo.GetByID(invoiceData.WarehouseManagerWorkerID)
	if err != nil {
		return dto.InvoiceFullData{}, err
	}

	released, err := service.workerRepo.GetByID(invoiceData.ReleasedWorkerID)
	if err != nil {
		return dto.InvoiceFullData{}, err
	}

	driver := model.Worker{}
	if invoiceData.DriverWorkerID != 0 {
		driver, err = service.workerRepo.GetByID(invoiceData.DriverWorkerID)
		if err != nil {
			return dto.InvoiceFullData{}, err
		}
	}

	recipient := model.Worker{}
	if invoiceData.RecipientWorkerID != 0 {
		recipient, err = service.workerRepo.GetByID(invoiceData.RecipientWorkerID)
		if err != nil {
			return dto.InvoiceFullData{}, err
		}
	}

	operatorAdd, err := service.workerRepo.GetByID(invoiceData.OperatorAddWorkerID)
	if err != nil {
		return dto.InvoiceFullData{}, err
	}

	operatorEdit := model.Worker{}
	if invoiceData.OperatorEditWorkerID != 0 {
		operatorEdit, err = service.workerRepo.GetByID(invoiceData.OperatorEditWorkerID)
		if err != nil {
			return dto.InvoiceFullData{}, err
		}
	}

	object, err := service.objectRepo.GetByID(invoiceData.ObjectID)
	if err != nil {
		return dto.InvoiceFullData{}, err
	}

	invoiceMaterials, err := service.invoiceMaterialsRepo.GetByInvoiceID(invoiceData.ID)
	if err != nil {
		return dto.InvoiceFullData{}, err
	}

	var invoiceItems []dto.InvoiceMaterial
	for _, invoiceMaterial := range invoiceMaterials {
		materialCostOfNewMaterial, err := service.materialCostRepo.GetByID(invoiceMaterial.MaterialCostID)
		if err != nil {
			return dto.InvoiceFullData{}, err
		}

		newMaterialInInvoice, err := service.materialRepo.GetByID(materialCostOfNewMaterial.MaterialID)
		if err != nil {
			return dto.InvoiceFullData{}, err
		}

		invoiceItems = append(invoiceItems, dto.InvoiceMaterial{
			Name:   newMaterialInInvoice.Name,
			Code:   newMaterialInInvoice.Code,
			Unit:   newMaterialInInvoice.Unit,
			Amount: invoiceMaterial.Amount,
			Notes:  invoiceMaterial.Notes,
		})
	}

	return dto.InvoiceFullData{
		Invoice: dto.InvoiceDetails{
			ID:               invoiceData.ID,
			ProjectName:      project.Name,
			TeamNumber:       team.Number,
			WarehouseManager: warehouseManager.Name,
			Released:         released.Name,
			Driver:           driver.Name,
			Recipient:        recipient.Name,
			OperatorAdd:      operatorAdd.Name,
			OperatorEdit:     operatorEdit.Name,
			ObjectName:       object.Name,
			DeliveryCode:     invoiceData.DeliveryCode,
			District:         invoiceData.District,
			CarNumber:        invoiceData.CarNumber,
			Notes:            invoiceData.Notes,
			DateOfInvoice:    invoiceData.DateOfInvoice.Format("2006-01-02"),
			DateOfAddition:   invoiceData.DateOfAddition.Format("2006-01-02"),
			DateOfEdit:       invoiceData.DateOfEdit.Format("2006-01-02"),
		},
		Materials: invoiceItems,
	}, nil
}

func (service *invoiceService) Create(data dto.InvoiceDataUpdateOrCreate) (bool, error) {
	updatedInvoice, err := service.invoiceRepo.Create(data.Invoice)
	if err != nil {
		return false, err
	}

	for _, newMaterialInInvoice := range data.Materials {
		materialCostOfNewMaterial, err := service.materialCostRepo.GetByMaterialID(newMaterialInInvoice.ID)
		if err != nil {
			return false, err
		}

		switch utils.MaterialDirection(updatedInvoice.InvoiceType) {
		case "warehouse":
			materialLocation, err := service.materialLocationRepo.GetByMaterialCostIDOrCreate(materialCostOfNewMaterial.ID, "warehouse", 0)
			if err != nil {
				return false, err
			}

			materialLocation.Amount += newMaterialInInvoice.Amount
			_, err = service.materialLocationRepo.Update(materialLocation)
			if err != nil {
				return false, err
			}

		case "teams":

			materialLocationWarehouse, err := service.materialLocationRepo.GetByMaterialCostIDOrCreate(materialCostOfNewMaterial.ID, "warehouse", 0)
			if err != nil {
				return false, err
			}

			if materialLocationWarehouse.Amount < newMaterialInInvoice.Amount {
				return false, fmt.Errorf("the amount of materail %v is more than the warehouse has", newMaterialInInvoice.Name)
			}

			materialLocationTeam, err := service.materialLocationRepo.GetByMaterialCostIDOrCreate(materialCostOfNewMaterial.ID, "teams", updatedInvoice.TeamID)
			if err != nil {
				return false, err
			}

			materialLocationTeam.Amount += newMaterialInInvoice.Amount
			materialLocationWarehouse.Amount -= newMaterialInInvoice.Amount

			_, err = service.materialLocationRepo.Update(materialLocationTeam)
			if err != nil {
				return false, err
			}

			_, err = service.materialLocationRepo.Update(materialLocationWarehouse)
			if err != nil {
				return false, err
			}

		}

		_, err = service.invoiceMaterialsRepo.Create(model.InvoiceMaterials{
			MaterialCostID: materialCostOfNewMaterial.ID,
			InvoiceID:      updatedInvoice.ID,
			Amount:         newMaterialInInvoice.Amount,
			Notes:          newMaterialInInvoice.InvoiceNotes,
		})
		if err != nil {
			return false, err
		}

	}

	return true, nil
}

func (service *invoiceService) Update(data dto.InvoiceDataUpdateOrCreate) (bool, error) {
	updatedInvoice, err := service.invoiceRepo.Update(data.Invoice)
	if err != nil {
		return false, err
	}

	oldInvoiceMaterials, err := service.invoiceMaterialsRepo.GetByInvoiceID(updatedInvoice.ID)
	if err != nil {
		return false, err
	}

	existsInUpdatedInvoice := []int{}
	for _, newMaterialInInvoice := range data.Materials {
		materialCostOfNewMaterial, err := service.materialCostRepo.GetByMaterialID(newMaterialInInvoice.ID)
		if err != nil {
			return false, err
		}

		existsInInvoice := false
		existsInInvoiceIndex := -1
		for index, oldMaterialInInvoice := range oldInvoiceMaterials {
			if materialCostOfNewMaterial.ID == oldMaterialInInvoice.ID {
				existsInInvoice = true
				existsInInvoiceIndex = index
				break
			}
		}

		if existsInInvoice && existsInInvoiceIndex != -1 {
			oldInvoiceMaterials[existsInInvoiceIndex].Amount = newMaterialInInvoice.Amount
			oldInvoiceMaterials[existsInInvoiceIndex].Notes = newMaterialInInvoice.Notes
			_, err := service.invoiceMaterialsRepo.Update(oldInvoiceMaterials[existsInInvoiceIndex])
			if err != nil {
				return false, err
			}

			existsInUpdatedInvoice = append(existsInUpdatedInvoice, existsInInvoiceIndex)

			switch utils.MaterialDirection(updatedInvoice.InvoiceType) {
			case "warehouse":
				materialLocation, err := service.materialLocationRepo.GetByMaterialCostIDOrCreate(materialCostOfNewMaterial.ID, "warehouse", 0)
				if err != nil {
					return false, err
				}

				materialLocation.Amount = newMaterialInInvoice.Amount
				_, err = service.materialLocationRepo.Update(materialLocation)
				if err != nil {
					return false, err
				}

			case "team":
				materialLocationWarehouse, err := service.materialLocationRepo.GetByMaterialCostIDOrCreate(materialCostOfNewMaterial.ID, "warehouse", 0)
				if err != nil {
					return false, err
				}

				if materialLocationWarehouse.Amount < newMaterialInInvoice.Amount {
					return false, fmt.Errorf("the amount of materail %v is more than the warehouse has", newMaterialInInvoice.Name)
				}

				materialLocationTeam, err := service.materialLocationRepo.GetByMaterialCostIDOrCreate(materialCostOfNewMaterial.ID, "teams", updatedInvoice.TeamID)
				if err != nil {
					return false, err
				}

				diffInAmount := newMaterialInInvoice.Amount - oldInvoiceMaterials[existsInInvoiceIndex].Amount
				materialLocationWarehouse.Amount -= diffInAmount
				materialLocationTeam.Amount += diffInAmount

				_, err = service.materialLocationRepo.Update(materialLocationTeam)
				if err != nil {
					return false, err
				}

				_, err = service.materialLocationRepo.Update(materialLocationWarehouse)
				if err != nil {
					return false, err
				}
			}
		} else {
			switch utils.MaterialDirection(updatedInvoice.InvoiceType) {
			case "warehouse":
				materialLocation, err := service.materialLocationRepo.GetByMaterialCostIDOrCreate(materialCostOfNewMaterial.ID, "warehouse", 0)
				if err != nil {
					return false, err
				}

				materialLocation.Amount += newMaterialInInvoice.Amount
				_, err = service.materialLocationRepo.Update(materialLocation)
				if err != nil {
					return false, err
				}

			case "teams":

				materialLocationWarehouse, err := service.materialLocationRepo.GetByMaterialCostIDOrCreate(materialCostOfNewMaterial.ID, "warehouse", 0)
				if err != nil {
					return false, err
				}

				if materialLocationWarehouse.Amount < newMaterialInInvoice.Amount {
					return false, fmt.Errorf("the amount of materail %v is more than the warehouse has", newMaterialInInvoice.Name)
				}

				materialLocationTeam, err := service.materialLocationRepo.GetByMaterialCostIDOrCreate(materialCostOfNewMaterial.ID, "teams", updatedInvoice.TeamID)
				if err != nil {
					return false, err
				}

				materialLocationTeam.Amount += newMaterialInInvoice.Amount
				materialLocationWarehouse.Amount -= newMaterialInInvoice.Amount

				_, err = service.materialLocationRepo.Update(materialLocationTeam)
				if err != nil {
					return false, err
				}

				_, err = service.materialLocationRepo.Update(materialLocationWarehouse)
				if err != nil {
					return false, err
				}

			}

			_, err := service.invoiceMaterialsRepo.Create(model.InvoiceMaterials{
				MaterialCostID: materialCostOfNewMaterial.ID,
				InvoiceID:      updatedInvoice.ID,
				Amount:         newMaterialInInvoice.Amount,
				Notes:          newMaterialInInvoice.InvoiceNotes,
			})
			if err != nil {
				return false, err
			}
		}
	}

	for index, oldMaterialInInvoice := range oldInvoiceMaterials {
		exists := false
		for _, existingIndex := range existsInUpdatedInvoice {
			if existingIndex == index {
				exists = true
			}
		}

		if !exists {
			switch utils.MaterialDirection(updatedInvoice.InvoiceType) {
			case "warehouse":
				materialLocation, err := service.materialLocationRepo.GetByMaterialCostIDOrCreate(oldMaterialInInvoice.MaterialCostID, "warehouse", 0)
				if err != nil {
					return false, err
				}
				materialLocation.Amount -= oldMaterialInInvoice.Amount
				_, err = service.materialLocationRepo.Update(materialLocation)
				if err != nil {
					return false, err
				}
			case "team":
				materialLocationWarehouse, err := service.materialLocationRepo.GetByMaterialCostIDOrCreate(oldMaterialInInvoice.MaterialCostID, "warehouse", 0)
				if err != nil {
					return false, err
				}

				materialLocationTeam, err := service.materialLocationRepo.GetByMaterialCostIDOrCreate(oldMaterialInInvoice.MaterialCostID, "teams", updatedInvoice.TeamID)
				if err != nil {
					return false, err
				}

				materialLocationTeam.Amount -= oldMaterialInInvoice.Amount
				materialLocationWarehouse.Amount += oldMaterialInInvoice.Amount

				_, err = service.materialLocationRepo.Update(materialLocationTeam)
				if err != nil {
					return false, err
				}

				_, err = service.materialLocationRepo.Update(materialLocationWarehouse)
				if err != nil {
					return false, err
				}

			}
			err := service.invoiceMaterialsRepo.Delete(oldMaterialInInvoice.ID)
			if err != nil {
				return false, err
			}
		}
	}

	return true, nil
}

func (service *invoiceService) Delete(id uint) error {
	return service.invoiceRepo.Delete(id)
}

func (service *invoiceService) Count(invoiceType string) (int64, error) {
	return service.invoiceRepo.Count(invoiceType)
}
