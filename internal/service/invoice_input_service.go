package service

import (
	"backend-v2/internal/dto"
	"backend-v2/internal/repository"
	"backend-v2/model"
	"backend-v2/pkg/utils"
	"fmt"
	"strconv"

	"github.com/xuri/excelize/v2"
)

type invoiceInputService struct {
	invoiceInputRepo     repository.IInovoiceInputRepository
	invoiceMaterialRepo  repository.IInvoiceMaterialsRepository
	materailLocationRepo repository.IMaterialLocationRepository
	workerRepo           repository.IWorkerRepository
	materialCostRepo     repository.IMaterialCostRepository
	materialRepo         repository.IMaterialRepository
	serialNumberRepo     repository.ISerialNumberRepository
}

func InitInvoiceInputService(
	invoiceInputRepo repository.IInovoiceInputRepository,
	invoiceMaterialRepo repository.IInvoiceMaterialsRepository,
	materialLocationRepo repository.IMaterialLocationRepository,
	workerRepo repository.IWorkerRepository,
	materialCostRepo repository.IMaterialCostRepository,
	materialRepo repository.IMaterialRepository,
	serialNumberRepo repository.ISerialNumberRepository,
) IInvoiceInputService {
	return &invoiceInputService{
		invoiceInputRepo:     invoiceInputRepo,
		invoiceMaterialRepo:  invoiceMaterialRepo,
		materailLocationRepo: materialLocationRepo,
		workerRepo:           workerRepo,
		materialCostRepo:     materialCostRepo,
		materialRepo:         materialRepo,
		serialNumberRepo:     serialNumberRepo,
	}
}

type IInvoiceInputService interface {
	GetAll() ([]model.InvoiceInput, error)
	GetPaginated(page, limit int, data model.InvoiceInput) ([]dto.InvoiceInputPaginated, error)
	GetByID(id uint) (model.InvoiceInput, error)
	Create(data dto.InvoiceInput) (dto.InvoiceInput, error)
	// Update(data dto.InvoiceInput) (dto.InvoiceInput, error)
	Delete(id uint) error
	Count(projectID uint) (int64, error)
	Confirmation(id, projectID uint) error
	UniqueCode(projectID uint) ([]string, error)
	UniqueWarehouseManager(projectID uint) ([]string, error)
	UniqueReleased(projectID uint) ([]string, error)
	Report(filter dto.InvoiceInputReportFilterRequest, projectID uint) (string, error)
	NewMaterialCost(data model.MaterialCost) error
	NewMaterialAndItsCost(data dto.NewMaterialDataFromInvoiceInput) error
}

func (service *invoiceInputService) GetAll() ([]model.InvoiceInput, error) {
	return service.invoiceInputRepo.GetAll()
}

func (service *invoiceInputService) GetPaginated(page, limit int, data model.InvoiceInput) ([]dto.InvoiceInputPaginated, error) {
	return service.invoiceInputRepo.GetPaginatedFiltered(page, limit, data)
}

func (service *invoiceInputService) GetByID(id uint) (model.InvoiceInput, error) {
	return service.invoiceInputRepo.GetByID(id)
}

func (service *invoiceInputService) Create(data dto.InvoiceInput) (dto.InvoiceInput, error) {

	count, err := service.invoiceInputRepo.Count(data.Details.ProjectID)
	if err != nil {
		return dto.InvoiceInput{}, err
	}

	code := utils.UniqueCodeGeneration("П", count+1, data.Details.ProjectID)
	data.Details.DeliveryCode = code

	invoiceInput, err := service.invoiceInputRepo.Create(data.Details)
	if err != nil {
		return dto.InvoiceInput{}, err
	}
	data.Details = invoiceInput

	for _, item := range data.Items {
		invoiceMaterial, err := service.invoiceMaterialRepo.Create(model.InvoiceMaterials{
			ProjectID:      data.Details.ProjectID,
			MaterialCostID: item.MaterialData.MaterialCostID,
			InvoiceID:      invoiceInput.ID,
			IsDefected:     item.MaterialData.IsDefected,
			InvoiceType:    item.MaterialData.InvoiceType,
			Amount:         item.MaterialData.Amount,
			Notes:          item.MaterialData.Notes,
		})
		if err != nil {
			return dto.InvoiceInput{}, err
		}

		if len(item.SerialNumbers) == 0 {
			continue
		}

		var serialNumbers []model.SerialNumber
		for _, serialNumber := range item.SerialNumbers {
			serialNumbers = append(serialNumbers, model.SerialNumber{
				Code:           serialNumber,
				Status:         "pending",
				StatusID:       invoiceMaterial.ID,
				MaterialCostID: invoiceMaterial.MaterialCostID,
			})
		}
		_, err = service.serialNumberRepo.CreateInBatches(serialNumbers)
		if err != nil {
			return dto.InvoiceInput{}, err
		}
	}

	return data, nil
}

// func (service *invoiceInputService) Update(data dto.InvoiceInput) (dto.InvoiceInput, error) {
// 	data.Details.DateOfEdit = time.Now()
// 	_, err := service.invoiceInputRepo.Update(data.Details)
// 	if err != nil {
// 		return dto.InvoiceInput{}, err
// 	}
//
// 	previousInvoiceMaterials, err := service.invoiceMaterialRepo.GetByInvoice(data.Details.ID, "input")
// 	if err != nil {
// 		return dto.InvoiceInput{}, err
// 	}
//
// 	indexesOfOldInvoiceMaterials, commonInvoiceMaterials, toBeAddedInvoiceMaterials, toBeDeletedInvoiceMaterials := utils.InvoiceMaterailsDevider(data.Items, previousInvoiceMaterials)
// 	for index, invoiceMaterial := range commonInvoiceMaterials {
// 		invoiceMaterial.ID = uint(indexesOfOldInvoiceMaterials[index])
// 		invoiceMaterial, err = service.invoiceMaterialRepo.Update(invoiceMaterial)
// 		if err != nil {
// 			return dto.InvoiceInput{}, err
// 		}
// 	}
//
// 	for _, invoiceMaterial := range toBeAddedInvoiceMaterials {
// 		_, err := service.invoiceMaterialRepo.Create(invoiceMaterial)
// 		if err != nil {
// 			return dto.InvoiceInput{}, err
// 		}
// 	}
//
// 	for _, invoiceMaterial := range toBeDeletedInvoiceMaterials {
// 		err := service.invoiceMaterialRepo.Delete(invoiceMaterial.ID)
// 		if err != nil {
// 			return dto.InvoiceInput{}, err
// 		}
// 	}
//
// 	return data, nil
// }

func (service *invoiceInputService) Delete(id uint) error {
	return service.invoiceInputRepo.Delete(id)
}

func (service *invoiceInputService) Count(projectID uint) (int64, error) {
	return service.invoiceInputRepo.Count(projectID)
}

func (service *invoiceInputService) Confirmation(id, projectID uint) error {
	invoiceInput, err := service.invoiceInputRepo.GetByID(id)
	if err != nil {
		return err
	}
	invoiceInput.Confirmed = true

	invoiceInput, err = service.invoiceInputRepo.Update(invoiceInput)
	if err != nil {
		return err
	}

	invoiceMaterials, err := service.invoiceMaterialRepo.GetByInvoice(invoiceInput.ProjectID, invoiceInput.ID, "input")
	if err != nil {
		return err
	}

	fmt.Println(invoiceMaterials)

	for _, invoiceMaterial := range invoiceMaterials {

		fmt.Println(invoiceMaterial)
		materialLocation, err := service.materailLocationRepo.GetByMaterialCostIDOrCreate(invoiceInput.ProjectID, invoiceMaterial.MaterialCostID, "warehouse", 0)
		if err != nil {
			return err
		}

		materialLocation.Amount += invoiceMaterial.Amount
		_, err = service.materailLocationRepo.Update(materialLocation)
		if err != nil {
			return err
		}

		serialNumbers, err := service.serialNumberRepo.GetByStatus("pending", invoiceMaterial.ID)
		if err != nil {
			return err
		}

		for _, serialNumber := range serialNumbers {
			serialNumber.Status = "warehouse"
			serialNumber.StatusID = materialLocation.ID
			_, err := service.serialNumberRepo.Update(serialNumber)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (service *invoiceInputService) UniqueCode(projectID uint) ([]string, error) {
	return service.invoiceInputRepo.UniqueCode(projectID)
}

func (service *invoiceInputService) UniqueWarehouseManager(projectID uint) ([]string, error) {
	ids, err := service.invoiceInputRepo.UniqueWarehouseManager(projectID)
	if err != nil {
		return []string{}, err
	}

	result := []string{}
	for _, id := range ids {
		idconv, _ := strconv.ParseUint(id, 10, 32)
		warehouseManager, err := service.workerRepo.GetByID(uint(idconv))
		if err != nil {
			return []string{}, err
		}

		result = append(result, warehouseManager.Name)
	}

	return result, nil
}

func (service *invoiceInputService) UniqueReleased(projectID uint) ([]string, error) {
	ids, err := service.invoiceInputRepo.UniqueReleased(projectID)
	if err != nil {
		return []string{}, err
	}

	result := []string{}
	for _, id := range ids {
		idconv, _ := strconv.ParseUint(id, 10, 32)
		released, err := service.workerRepo.GetByID(uint(idconv))
		if err != nil {
			return []string{}, err
		}

		result = append(result, released.Name)
	}

	return result, nil
}

func (service *invoiceInputService) Report(filter dto.InvoiceInputReportFilterRequest, projectID uint) (string, error) {
	newFilter := dto.InvoiceInputReportFilter{
		Code:     filter.Code,
		DateFrom: filter.DateFrom,
		DateTo:   filter.DateTo,
	}

	var err error
	if filter.WarehouseManager != "" {
		warehouseManager, err := service.workerRepo.GetByName(filter.WarehouseManager)
		if err != nil {
			return "", err
		}

		newFilter.WarehouseManagerID = warehouseManager.ID
	} else {
		newFilter.WarehouseManagerID = 0
	}

	if filter.Released != "" {
		released, err := service.workerRepo.GetByName(filter.Released)
		if err != nil {
			return "", err
		}

		newFilter.ReleasedID = released.ID
	} else {
		newFilter.ReleasedID = 0
	}

	invoices, err := service.invoiceInputRepo.ReportFilterData(newFilter, projectID)
	if err != nil {
		return "", err
	}

	f, err := excelize.OpenFile("./pkg/excels/report/Invoice Input Report.xlsx")
	if err != nil {
		return "", err
	}
	sheetName := "Sheet1"

	rowCount := 2
	for _, invoice := range invoices {
		invoiceMaterials, err := service.invoiceMaterialRepo.GetByInvoice(filter.ProjectID, invoice.ID, "input")
		if err != nil {
			return "", err
		}

		fmt.Println(invoiceMaterials)
		for index, invoiceMaterial := range invoiceMaterials {
			materialCost, err := service.materialCostRepo.GetByID(invoiceMaterial.MaterialCostID)
			if err != nil {
				return "", nil
			}

			material, err := service.materialRepo.GetByID(materialCost.MaterialID)
			if err != nil {
				return "", nil
			}

			fmt.Println(materialCost, material)
			if index == 0 {
				f.SetCellValue(sheetName, "A"+fmt.Sprint(rowCount), invoice.DeliveryCode)

				warehouseManager, err := service.workerRepo.GetByID(invoice.WarehouseManagerWorkerID)
				if err != nil {
					return "", err
				}
				f.SetCellValue(sheetName, "B"+fmt.Sprint(rowCount), warehouseManager.Name)

				released, err := service.workerRepo.GetByID(invoice.ReleasedWorkerID)
				if err != nil {
					return "", err
				}

				f.SetCellValue(sheetName, "C"+fmt.Sprint(rowCount), released.Name)
				dateOfInvoice := invoice.DateOfInvoice.String()
				dateOfInvoice = dateOfInvoice[:len(dateOfInvoice)-10]
				f.SetCellValue(sheetName, "D"+fmt.Sprint(rowCount), dateOfInvoice)
			} else {
				f.SetCellValue(sheetName, "A"+fmt.Sprint(rowCount), "-")
				f.SetCellValue(sheetName, "B"+fmt.Sprint(rowCount), "-")
				f.SetCellValue(sheetName, "C"+fmt.Sprint(rowCount), "-")
				f.SetCellValue(sheetName, "D"+fmt.Sprint(rowCount), "-")
			}

			f.SetCellValue(sheetName, "E"+fmt.Sprint(rowCount), material.Name)
			f.SetCellValue(sheetName, "F"+fmt.Sprint(rowCount), material.Unit)
			f.SetCellValue(sheetName, "G"+fmt.Sprint(rowCount), invoiceMaterial.Amount)
			f.SetCellValue(sheetName, "H"+fmt.Sprint(rowCount), materialCost.CostM19)
			f.SetCellValue(sheetName, "I"+fmt.Sprint(rowCount), invoiceMaterial.Notes)
			rowCount++

		}
	}

	fileName := "Invoice Input Report " + fmt.Sprint(rowCount) + ".xlsx"
	f.SaveAs("./pkg/excels/report/" + fileName)
	if err := f.Close(); err != nil {
		fmt.Println(err)
	}

	return fileName, nil
}

func (service *invoiceInputService) NewMaterialCost(data model.MaterialCost) error {
	_, err := service.materialCostRepo.Create(data)
	return err
}

func (service *invoiceInputService) NewMaterialAndItsCost(data dto.NewMaterialDataFromInvoiceInput) error {

	material, err := service.materialRepo.Create(model.Material{
		Category:        data.Category,
		Code:            data.Code,
		Name:            data.Name,
		Unit:            data.Unit,
		ProjectID:       data.ProjectID,
		Notes:           data.Notes,
		HasSerialNumber: data.HasSerialNumber,
	})
	if err != nil {
		return err
	}
	_, err = service.materialCostRepo.Create(model.MaterialCost{
		MaterialID:       material.ID,
		CostPrime:        data.CostPrime,
		CostM19:          data.CostM19,
		CostWithCustomer: data.CostWithCustomer,
	})

	return err
}
