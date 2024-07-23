package service

import (
	"backend-v2/internal/dto"
	"backend-v2/internal/repository"
	"backend-v2/model"
	"backend-v2/pkg/utils"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/xuri/excelize/v2"
)

type invoiceInputService struct {
	invoiceInputRepo         repository.IInovoiceInputRepository
	invoiceMaterialRepo      repository.IInvoiceMaterialsRepository
	materialLocationRepo     repository.IMaterialLocationRepository
	workerRepo               repository.IWorkerRepository
	materialCostRepo         repository.IMaterialCostRepository
	materialRepo             repository.IMaterialRepository
	serialNumberRepo         repository.ISerialNumberRepository
	serialNumberMovementRepo repository.ISerialNumberMovementRepository
	invoiceCountRepo         repository.IInvoiceCountRepository
}

func InitInvoiceInputService(
	invoiceInputRepo repository.IInovoiceInputRepository,
	invoiceMaterialRepo repository.IInvoiceMaterialsRepository,
	materialLocationRepo repository.IMaterialLocationRepository,
	workerRepo repository.IWorkerRepository,
	materialCostRepo repository.IMaterialCostRepository,
	materialRepo repository.IMaterialRepository,
	serialNumberRepo repository.ISerialNumberRepository,
	serialNumberMovementRepo repository.ISerialNumberMovementRepository,
	invoiceCountRepo repository.IInvoiceCountRepository,
) IInvoiceInputService {
	return &invoiceInputService{
		invoiceInputRepo:         invoiceInputRepo,
		invoiceMaterialRepo:      invoiceMaterialRepo,
		materialLocationRepo:     materialLocationRepo,
		workerRepo:               workerRepo,
		materialCostRepo:         materialCostRepo,
		materialRepo:             materialRepo,
		serialNumberRepo:         serialNumberRepo,
		serialNumberMovementRepo: serialNumberMovementRepo,
		invoiceCountRepo:         invoiceCountRepo,
	}
}

type IInvoiceInputService interface {
	GetAll() ([]model.InvoiceInput, error)
	GetPaginated(page, limit int, data model.InvoiceInput) ([]dto.InvoiceInputPaginated, error)
	GetByID(id uint) (model.InvoiceInput, error)
	GetInvoiceMaterialsWithoutSerialNumbers(id uint) ([]dto.InvoiceMaterialsWithoutSerialNumberView, error)
	GetInvoiceMaterialsWithSerialNumbers(id uint) ([]dto.InvoiceMaterialsWithSerialNumberView, error)
	Create(data dto.InvoiceInput) (model.InvoiceInput, error)
	Update(data dto.InvoiceInput) (model.InvoiceInput, error)
	Delete(id uint) error
	Count(projectID uint) (int64, error)
	Confirmation(id, projectID uint) error
	UniqueCode(projectID uint) ([]dto.DataForSelect[string], error)
	UniqueWarehouseManager(projectID uint) ([]dto.DataForSelect[uint], error)
	UniqueReleased(projectID uint) ([]dto.DataForSelect[uint], error)
	Report(filter dto.InvoiceInputReportFilterRequest) (string, error)
	NewMaterialCost(data model.MaterialCost) error
	NewMaterialAndItsCost(data dto.NewMaterialDataFromInvoiceInput) error
	GetMaterialsForEdit(id uint) ([]dto.InvoiceInputMaterialForEdit, error)
	Import(filePath string, projectID uint, workerID uint) error
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

func (service *invoiceInputService) GetInvoiceMaterialsWithoutSerialNumbers(id uint) ([]dto.InvoiceMaterialsWithoutSerialNumberView, error) {
	return service.invoiceMaterialRepo.GetInvoiceMaterialsWithoutSerialNumbers(id, "input")
}

func (service *invoiceInputService) GetInvoiceMaterialsWithSerialNumbers(id uint) ([]dto.InvoiceMaterialsWithSerialNumberView, error) {
	queryData, err := service.invoiceMaterialRepo.GetInvoiceMaterialsWithSerialNumbers(id, "input")
	if err != nil {
		return []dto.InvoiceMaterialsWithSerialNumberView{}, err
	}

	result := []dto.InvoiceMaterialsWithSerialNumberView{}
	current := dto.InvoiceMaterialsWithSerialNumberView{}
	for index, materialInfo := range queryData {
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
			result = append(result, current)
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

	if len(queryData) != 0 {
		result = append(result, current)
	}

	return result, nil
}

func (service *invoiceInputService) Create(data dto.InvoiceInput) (model.InvoiceInput, error) {

	count, err := service.invoiceInputRepo.Count(data.Details.ProjectID)
	if err != nil {
		return model.InvoiceInput{}, err
	}

	code := utils.UniqueCodeGeneration("П", count+1, data.Details.ProjectID)
	data.Details.DeliveryCode = code

	var invoiceMaterials []model.InvoiceMaterials
	var serialNumbers []model.SerialNumber
	var serialNumberMovements []model.SerialNumberMovement
	for _, item := range data.Items {
		invoiceMaterials = append(invoiceMaterials, model.InvoiceMaterials{
			ProjectID:      data.Details.ProjectID,
			MaterialCostID: item.MaterialData.MaterialCostID,
			IsDefected:     item.MaterialData.IsDefected,
			InvoiceType:    "input",
			Amount:         item.MaterialData.Amount,
			Notes:          item.MaterialData.Notes,
		})

		if len(item.SerialNumbers) == 0 {
			continue
		}

		for _, serialNumber := range item.SerialNumbers {
			serialNumbers = append(serialNumbers, model.SerialNumber{
				Code:           serialNumber,
				ProjectID:      data.Details.ProjectID,
				MaterialCostID: item.MaterialData.MaterialCostID,
			})

			serialNumberMovements = append(serialNumberMovements, model.SerialNumberMovement{
				ProjectID:    data.Details.ProjectID,
				InvoiceType:  "input",
				IsDefected:   false,
				Confirmation: false,
			})
		}

	}

	invoiceInput, err := service.invoiceInputRepo.Create(dto.InvoiceInputCreateQueryData{
		InvoiceData:          data.Details,
		InvoiceMaterials:     invoiceMaterials,
		SerialNumbers:        serialNumbers,
		SerialNumberMovement: serialNumberMovements,
	})
	if err != nil {
		return model.InvoiceInput{}, err
	}

	return invoiceInput, nil
}

func (service *invoiceInputService) Update(data dto.InvoiceInput) (model.InvoiceInput, error) {
	var invoiceMaterials []model.InvoiceMaterials
	var serialNumbers []model.SerialNumber
	var serialNumberMovements []model.SerialNumberMovement
	for _, item := range data.Items {
		invoiceMaterials = append(invoiceMaterials, model.InvoiceMaterials{
			ProjectID:      data.Details.ProjectID,
			MaterialCostID: item.MaterialData.MaterialCostID,
			IsDefected:     item.MaterialData.IsDefected,
			InvoiceType:    "input",
			Amount:         item.MaterialData.Amount,
			Notes:          item.MaterialData.Notes,
		})

		if len(item.SerialNumbers) == 0 {
			continue
		}

		for _, serialNumber := range item.SerialNumbers {
			serialNumbers = append(serialNumbers, model.SerialNumber{
				Code:           serialNumber,
				ProjectID:      data.Details.ProjectID,
				MaterialCostID: item.MaterialData.MaterialCostID,
			})

			serialNumberMovements = append(serialNumberMovements, model.SerialNumberMovement{
				ProjectID:    data.Details.ProjectID,
				InvoiceType:  "input",
				IsDefected:   false,
				Confirmation: false,
			})
		}

	}

	invoiceInput, err := service.invoiceInputRepo.Update(dto.InvoiceInputCreateQueryData{
		InvoiceData:          data.Details,
		InvoiceMaterials:     invoiceMaterials,
		SerialNumbers:        serialNumbers,
		SerialNumberMovement: serialNumberMovements,
	})
	if err != nil {
		return model.InvoiceInput{}, err
	}

	return invoiceInput, nil
}

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

	invoiceMaterials, err := service.invoiceMaterialRepo.GetByInvoice(invoiceInput.ProjectID, invoiceInput.ID, "input")
	if err != nil {
		return err
	}

	materialsInWarehouse, err := service.materialLocationRepo.GetMaterialsInLocationBasedOnInvoiceID(0, "warehouse", id, "input")
	if err != nil {
		return err
	}

	toBeUpdated := []model.MaterialLocation{}
	toBeCreated := []model.MaterialLocation{}
	for _, invoiceMaterial := range invoiceMaterials {
		indexOfExistingMaterial := -1
		for index, oneMaterialInWarehouse := range materialsInWarehouse {
			if oneMaterialInWarehouse.MaterialCostID == invoiceMaterial.MaterialCostID {
				indexOfExistingMaterial = index
				break
			}
		}

		if indexOfExistingMaterial == -1 {
			toBeCreated = append(toBeCreated, model.MaterialLocation{
				ID:             0,
				MaterialCostID: invoiceMaterial.MaterialCostID,
				ProjectID:      invoiceInput.ProjectID,
				LocationID:     0,
				LocationType:   "warehouse",
				Amount:         invoiceMaterial.Amount,
			})
		} else {
			materialsInWarehouse[indexOfExistingMaterial].Amount += invoiceMaterial.Amount
			toBeUpdated = append(toBeUpdated, materialsInWarehouse[indexOfExistingMaterial])
		}
	}

	serialNumberMovements, err := service.serialNumberMovementRepo.GetByInvoice(id, "input")
	if err != nil {
		return err
	}

	serialNumberLocations := []model.SerialNumberLocation{}
	for _, serialNumberMovement := range serialNumberMovements {
		serialNumberLocations = append(serialNumberLocations, model.SerialNumberLocation{
			SerialNumberID: serialNumberMovement.SerialNumberID,
			ProjectID:      projectID,
			LocationID:     0,
			LocationType:   "warehouse",
		})
	}

	err = service.invoiceInputRepo.Confirmation(dto.InvoiceInputConfirmationQueryData{
		InvoiceData:          invoiceInput,
		ToBeUpdatedMaterials: toBeUpdated,
		ToBeCreatedMaterials: toBeCreated,
		SerialNumbers:        serialNumberLocations,
	})

	return err
}

func (service *invoiceInputService) UniqueCode(projectID uint) ([]dto.DataForSelect[string], error) {
	return service.invoiceInputRepo.UniqueCode(projectID)
}

func (service *invoiceInputService) UniqueWarehouseManager(projectID uint) ([]dto.DataForSelect[uint], error) {
	return service.invoiceInputRepo.UniqueWarehouseManager(projectID)
}

func (service *invoiceInputService) UniqueReleased(projectID uint) ([]dto.DataForSelect[uint], error) {
	return service.invoiceInputRepo.UniqueReleased(projectID)
}

func (service *invoiceInputService) Report(filter dto.InvoiceInputReportFilterRequest) (string, error) {
	invoices, err := service.invoiceInputRepo.ReportFilterData(filter)
	if err != nil {
		return "", err
	}

	templateFilePath := filepath.Join("./pkg/excels/templates/", "Invoice Input Report.xlsx")
	f, err := excelize.OpenFile(templateFilePath)
	if err != nil {
		return "", err
	}
	sheetName := "Sheet1"

	rowCount := 2
	for _, invoice := range invoices {
		invoiceMaterials, err := service.invoiceMaterialRepo.GetDataForReport(invoice.ID, "input")
		if err != nil {
			return "", err
		}

		for _, invoiceMaterial := range invoiceMaterials {
			f.SetCellStr(sheetName, "A"+fmt.Sprint(rowCount), invoice.DeliveryCode)
			f.SetCellStr(sheetName, "B"+fmt.Sprint(rowCount), invoice.WarehouseManagerName)
			f.SetCellStr(sheetName, "C"+fmt.Sprint(rowCount), invoice.ReleasedName)

			dateOfInvoice := invoice.DateOfInvoice.String()
			dateOfInvoice = dateOfInvoice[:len(dateOfInvoice)-10]
			f.SetCellStr(sheetName, "D"+fmt.Sprint(rowCount), dateOfInvoice)

			f.SetCellValue(sheetName, "E"+fmt.Sprint(rowCount), invoiceMaterial.MaterialName)
			f.SetCellValue(sheetName, "F"+fmt.Sprint(rowCount), invoiceMaterial.MaterialUnit)
			f.SetCellFloat(sheetName, "G"+fmt.Sprint(rowCount), invoiceMaterial.InvoiceMaterialAmount, 2, 64)

			costM19, _ := invoiceMaterial.MaterialCostM19.Float64()
			f.SetCellFloat(sheetName, "H"+fmt.Sprint(rowCount), costM19, 2, 64)
			f.SetCellValue(sheetName, "I"+fmt.Sprint(rowCount), invoiceMaterial.InvoiceMaterialNotes)
			rowCount++
		}
	}

	currentTime := time.Now()
	fileName := fmt.Sprintf(
		"Отсчет накладной приход - %s.xlsx",
		currentTime.Format("02-01-2006"),
	)

	tempFilePath := filepath.Join("./pkg/excels/temp/", fileName)

	f.SaveAs(tempFilePath)

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

func (service *invoiceInputService) GetMaterialsForEdit(id uint) ([]dto.InvoiceInputMaterialForEdit, error) {
	return service.invoiceInputRepo.GetMaterialsForEdit(id)
}

func (service *invoiceInputService) Import(filePath string, projectID uint, workerID uint) error {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		f.Close()
		os.Remove(filePath)
		return fmt.Errorf("Не смог открыть файл: %v", err)
	}

	sheetName := "Sheet1"
	rows, err := f.GetRows(sheetName)
	if err != nil {
		f.Close()
		os.Remove(filePath)
		return fmt.Errorf("Не смог найти таблицу 'Импорт': %v", err)
	}

	if len(rows) == 1 {
		f.Close()
		os.Remove(filePath)
		return fmt.Errorf("Файл не имеет данных")
	}

	count, err := service.invoiceCountRepo.CountInvoice("input", projectID)
	if err != nil {
		f.Close()
		os.Remove(filePath)
		return fmt.Errorf("Файл не имеет данных")
	}

	index := 1
	importData := []dto.InvoiceInputImportData{}
	currentInvoiceInput := model.InvoiceInput{}
	currentInvoiceMaterials := []model.InvoiceMaterials{}
	for len(rows) > index {
		excelInvoiceInput := model.InvoiceInput{
			ID:               0,
			ProjectID:        projectID,
			ReleasedWorkerID: workerID,
			Confirmed:        false,
			Notes:            "",
		}

		warehouseManagerName, err := f.GetCellValue(sheetName, "B"+fmt.Sprint(index+1))
		if err != nil {
			f.Close()
			os.Remove(filePath)
			return fmt.Errorf("Нету данных в ячейке B%v: %v", index+1, err)
		}

		warehouseManager, err := service.workerRepo.GetByName(warehouseManagerName)
		if err != nil {
			f.Close()
			os.Remove(filePath)
			return fmt.Errorf("Имя %v в ячейке B%v не найдено в базе: %v", warehouseManagerName, index+1, err)
		}

		excelInvoiceInput.WarehouseManagerWorkerID = warehouseManager.ID

		dateOfInvoiceInExcel, err := f.GetCellValue(sheetName, "D"+fmt.Sprint(index+1))
		if err != nil {
			f.Close()
			os.Remove(filePath)
			return fmt.Errorf("Нету данных в ячейке D%v: %v", index+1, err)
		}

		dateLayout := "2006/01/02"
		dateOfInvoice, err := time.Parse(dateLayout, dateOfInvoiceInExcel)
		if err != nil {
			f.Close()
			os.Remove(filePath)
			return fmt.Errorf("Неправильные данные в ячейке D%v: %v", index+1, err)
		}

		excelInvoiceInput.DateOfInvoice = dateOfInvoice
		if index == 1 {
			currentInvoiceInput = excelInvoiceInput
		}

		excelInvoiceMaterial := model.InvoiceMaterials{
			InvoiceType: "input",
			IsDefected:  false,
			ProjectID:   projectID,
		}

		materialName, err := f.GetCellValue(sheetName, "E"+fmt.Sprint(index+1))
		if err != nil {
			f.Close()
			os.Remove(filePath)
			return fmt.Errorf("Нету данных в ячейке E%v: %v", index+1, err)
		}

		material, err := service.materialRepo.GetByName(materialName)
		if err != nil {
			f.Close()
			os.Remove(filePath)
			return fmt.Errorf("Материал %v в ячейке E%v не найдено в базе: %v", warehouseManagerName, index+1, err)
		}

		materialCost, err := service.materialCostRepo.GetByMaterialID(material.ID)
		if err != nil {
			f.Close()
			os.Remove(filePath)
			return fmt.Errorf("Цена Материала %v в ячейке E%v не найдено в базе: %v", warehouseManagerName, index+1, err)
		}

		excelInvoiceMaterial.MaterialCostID = materialCost[0].ID

		amountExcel, err := f.GetCellValue(sheetName, "G"+fmt.Sprint(index+1))
		if err != nil {
			f.Close()
			os.Remove(filePath)
			return fmt.Errorf("Нету данных в ячейке G%v: %v", index+1, err)
		}

		amount, err := strconv.ParseFloat(amountExcel, 64)
		if err != nil {
			f.Close()
			os.Remove(filePath)
			return fmt.Errorf("Нету данных в ячейке G%v: %v", index+1, err)
		}

		excelInvoiceMaterial.Amount = amount

		if currentInvoiceInput.DateOfInvoice.Equal(excelInvoiceInput.DateOfInvoice) {
			currentInvoiceMaterials = append(currentInvoiceMaterials, excelInvoiceMaterial)
		} else {
			currentInvoiceInput.DeliveryCode = utils.UniqueCodeGeneration("П", int64(count), projectID)
			count++
			importData = append(importData, dto.InvoiceInputImportData{
				Details: currentInvoiceInput,
				Items:   currentInvoiceMaterials,
			})

			currentInvoiceInput = excelInvoiceInput
			currentInvoiceMaterials = []model.InvoiceMaterials{excelInvoiceMaterial}
		}

		index++
	}

			currentInvoiceInput.DeliveryCode = utils.UniqueCodeGeneration("П", int64(count), projectID)
	importData = append(importData, dto.InvoiceInputImportData{
		Details: currentInvoiceInput,
		Items:   currentInvoiceMaterials,
	})

	return service.invoiceInputRepo.Import(importData)
}
