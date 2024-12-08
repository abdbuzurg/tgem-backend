package service

import (
	"backend-v2/internal/dto"
	"backend-v2/internal/repository"
	"backend-v2/model"
	"backend-v2/pkg/utils"
	"fmt"
	"os"
	// "os/exec"
	"path/filepath"
	// "runtime"
	"strconv"
	"time"

	"github.com/xuri/excelize/v2"
)

type invoiceOutputService struct {
	invoiceOutputRepo    repository.IInvoiceOutputRepository
	invoiceMaterialRepo  repository.IInvoiceMaterialsRepository
	workerRepo           repository.IWorkerRepository
	teamRepo             repository.ITeamRepository
	objectRepo           repository.IObjectRepository
	materialCostRepo     repository.IMaterialCostRepository
	materialLocationRepo repository.IMaterialLocationRepository
	materialRepo         repository.IMaterialRepository
	districtRepo         repository.IDistrictRepository
	serialNumberRepo     repository.ISerialNumberRepository
	invoiceCountRepo     repository.IInvoiceCountRepository
	projectRepo          repository.IProjectRepository
}

func InitInvoiceOutputService(
	invoiceOutputRepo repository.IInvoiceOutputRepository,
	invoiceMaterialRepo repository.IInvoiceMaterialsRepository,
	workerRepo repository.IWorkerRepository,
	teamRepo repository.ITeamRepository,
	objectRepo repository.IObjectRepository,
	materialCostRepo repository.IMaterialCostRepository,
	materialLocationRepo repository.IMaterialLocationRepository,
	materialRepo repository.IMaterialRepository,
	districtRepo repository.IDistrictRepository,
	serialNumberRepo repository.ISerialNumberRepository,
	invoiceCountRepo repository.IInvoiceCountRepository,
	projectRepo repository.IProjectRepository,
) IInvoiceOutputService {
	return &invoiceOutputService{
		invoiceOutputRepo:    invoiceOutputRepo,
		invoiceMaterialRepo:  invoiceMaterialRepo,
		workerRepo:           workerRepo,
		teamRepo:             teamRepo,
		objectRepo:           objectRepo,
		materialCostRepo:     materialCostRepo,
		materialLocationRepo: materialLocationRepo,
		materialRepo:         materialRepo,
		districtRepo:         districtRepo,
		serialNumberRepo:     serialNumberRepo,
		invoiceCountRepo:     invoiceCountRepo,
		projectRepo:          projectRepo,
	}
}

type IInvoiceOutputService interface {
	GetAll() ([]model.InvoiceOutput, error)
	GetPaginated(page, limit int, data model.InvoiceOutput) ([]dto.InvoiceOutputPaginated, error)
	GetByID(id uint) (model.InvoiceOutput, error)
	GetDocument(deliveryCode string) (string, error)
	GetInvoiceMaterialsWithoutSerialNumbers(id uint) ([]dto.InvoiceMaterialsWithoutSerialNumberView, error)
	GetInvoiceMaterialsWithSerialNumbers(id uint) ([]dto.InvoiceMaterialsWithSerialNumberView, error)
	Create(data dto.InvoiceOutput) (model.InvoiceOutput, error)
	Update(data dto.InvoiceOutput) (model.InvoiceOutput, error)
	Delete(id uint) error
	Count(projectID uint) (int64, error)
	Confirmation(id uint) error
	UniqueCode(projectID uint) ([]dto.DataForSelect[string], error)
	UniqueWarehouseManager(projectID uint) ([]dto.DataForSelect[uint], error)
	UniqueRecieved(projectID uint) ([]dto.DataForSelect[uint], error)
	UniqueDistrict(projectID uint) ([]dto.DataForSelect[uint], error)
	UniqueTeam(projectID uint) ([]dto.DataForSelect[uint], error)
	Report(filter dto.InvoiceOutputReportFilterRequest) (string, error)
	GetTotalMaterialAmount(projectID, materialID uint) (float64, error)
	GetSerialNumbersByMaterial(projectID, materialID uint) ([]string, error)
	GetAvailableMaterialsInWarehouse(projectID uint) ([]dto.AvailableMaterialsInWarehouse, error)
	GetMaterialsForEdit(id uint) ([]dto.InvoiceOutputMaterialsForEdit, error)
	Import(filePath string, projectID uint, workerID uint) error
}

func (service *invoiceOutputService) GetAll() ([]model.InvoiceOutput, error) {
	return service.invoiceOutputRepo.GetAll()
}

func (service *invoiceOutputService) GetByID(id uint) (model.InvoiceOutput, error) {
	return service.invoiceOutputRepo.GetByID(id)
}

func (service *invoiceOutputService) GetPaginated(page, limit int, data model.InvoiceOutput) ([]dto.InvoiceOutputPaginated, error) {
	return service.invoiceOutputRepo.GetPaginatedFiltered(page, limit, data)
}

func (service *invoiceOutputService) GetInvoiceMaterialsWithoutSerialNumbers(id uint) ([]dto.InvoiceMaterialsWithoutSerialNumberView, error) {
	return service.invoiceMaterialRepo.GetInvoiceMaterialsWithoutSerialNumbers(id, "output")
}

func (service *invoiceOutputService) GetInvoiceMaterialsWithSerialNumbers(id uint) ([]dto.InvoiceMaterialsWithSerialNumberView, error) {
	queryData, err := service.invoiceMaterialRepo.GetInvoiceMaterialsWithSerialNumbers(id, "output")
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

func (service *invoiceOutputService) Create(data dto.InvoiceOutput) (model.InvoiceOutput, error) {
	count, err := service.invoiceCountRepo.CountInvoice("output", data.Details.ProjectID)
	if err != nil {
		return model.InvoiceOutput{}, err
	}

	data.Details.DeliveryCode = utils.UniqueCodeGeneration("О", int64(count+1), data.Details.ProjectID)

	invoiceMaterialForCreate := []model.InvoiceMaterials{}
	serialNumberMovements := []model.SerialNumberMovement{}
	for _, invoiceMaterial := range data.Items {
		if len(invoiceMaterial.SerialNumbers) == 0 {
			materialInfoSorted, err := service.materialLocationRepo.GetMaterialAmountSortedByCostM19InLocation(data.Details.ProjectID, invoiceMaterial.MaterialID, "warehouse", 0)
			if err != nil {
				return model.InvoiceOutput{}, err
			}

			index := 0
			for invoiceMaterial.Amount > 0 {
				invoiceMaterialCreate := model.InvoiceMaterials{
					ProjectID:      data.Details.ProjectID,
					ID:             0,
					MaterialCostID: materialInfoSorted[index].MaterialCostID,
					InvoiceID:      0,
					InvoiceType:    "output",
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
			MC_IDs_AND_SN_IDs, err := service.serialNumberRepo.GetMaterialCostIDsByCodesInLocation(invoiceMaterial.MaterialID, invoiceMaterial.SerialNumbers, "warehouse", 0)
			if err != nil {
				return model.InvoiceOutput{}, err
			}

			var invoiceMaterialCreate model.InvoiceMaterials
			for index, oneEntry := range MC_IDs_AND_SN_IDs {

				serialNumberMovements = append(serialNumberMovements, model.SerialNumberMovement{
					ID:             0,
					SerialNumberID: oneEntry.SerialNumberID,
					ProjectID:      data.Details.ProjectID,
					InvoiceID:      0,
					InvoiceType:    "output",
					IsDefected:     false,
					Confirmation:   false,
				})

				if index == 0 {
					invoiceMaterialCreate = model.InvoiceMaterials{
						ProjectID:      data.Details.ProjectID,
						ID:             0,
						MaterialCostID: oneEntry.MaterialCostID,
						InvoiceID:      data.Details.ID,
						InvoiceType:    "output",
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
						InvoiceType:    "output",
						IsDefected:     false,
						Amount:         0,
						Notes:          invoiceMaterial.Notes,
					}

				}

			}

			invoiceMaterialForCreate = append(invoiceMaterialForCreate, invoiceMaterialCreate)
		}

	}

	if err := service.GenerateExcelFile(data); err != nil {
		return model.InvoiceOutput{}, err
	}

	invoiceOutput, err := service.invoiceOutputRepo.Create(dto.InvoiceOutputCreateQueryData{
		Invoice:               data.Details,
		InvoiceMaterials:      invoiceMaterialForCreate,
		SerialNumberMovements: serialNumberMovements,
	})
	if err != nil {
		return model.InvoiceOutput{}, err
	}

	return invoiceOutput, err
}

func (service *invoiceOutputService) Update(data dto.InvoiceOutput) (model.InvoiceOutput, error) {
	invoiceMaterialForCreate := []model.InvoiceMaterials{}
	serialNumberMovements := []model.SerialNumberMovement{}
	for _, invoiceMaterial := range data.Items {
		if len(invoiceMaterial.SerialNumbers) == 0 {
			materialInfoSorted, err := service.materialLocationRepo.GetMaterialAmountSortedByCostM19InLocation(data.Details.ProjectID, invoiceMaterial.MaterialID, "warehouse", 0)
			if err != nil {
				return model.InvoiceOutput{}, err
			}

			index := 0
			for invoiceMaterial.Amount > 0 {
				invoiceMaterialCreate := model.InvoiceMaterials{
					ProjectID:      data.Details.ProjectID,
					ID:             0,
					MaterialCostID: materialInfoSorted[index].MaterialCostID,
					InvoiceID:      0,
					InvoiceType:    "output",
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
			MC_IDs_AND_SN_IDs, err := service.serialNumberRepo.GetMaterialCostIDsByCodesInLocation(invoiceMaterial.MaterialID, invoiceMaterial.SerialNumbers, "warehouse", 0)
			if err != nil {
				return model.InvoiceOutput{}, err
			}

			var invoiceMaterialCreate model.InvoiceMaterials
			for index, oneEntry := range MC_IDs_AND_SN_IDs {

				serialNumberMovements = append(serialNumberMovements, model.SerialNumberMovement{
					ID:             0,
					SerialNumberID: oneEntry.SerialNumberID,
					ProjectID:      data.Details.ProjectID,
					InvoiceID:      0,
					InvoiceType:    "output",
					IsDefected:     false,
					Confirmation:   false,
				})

				if index == 0 {
					invoiceMaterialCreate = model.InvoiceMaterials{
						ProjectID:      data.Details.ProjectID,
						ID:             0,
						MaterialCostID: oneEntry.MaterialCostID,
						InvoiceID:      data.Details.ID,
						InvoiceType:    "output",
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
						InvoiceType:    "output",
						IsDefected:     false,
						Amount:         0,
						Notes:          invoiceMaterial.Notes,
					}

				}

			}

			invoiceMaterialForCreate = append(invoiceMaterialForCreate, invoiceMaterialCreate)
		}

	}

	excelFilePath := filepath.Join("./pkg/excels/output/", data.Details.DeliveryCode+".xlsx")
	if err := os.Remove(excelFilePath); err != nil {
		return model.InvoiceOutput{}, err
	}

	if err := service.GenerateExcelFile(data); err != nil {
		return model.InvoiceOutput{}, err
	}

	invoiceOutput, err := service.invoiceOutputRepo.Update(dto.InvoiceOutputCreateQueryData{
		Invoice:               data.Details,
		InvoiceMaterials:      invoiceMaterialForCreate,
		SerialNumberMovements: serialNumberMovements,
	})
	if err != nil {
		return model.InvoiceOutput{}, err
	}

	return invoiceOutput, nil
}

func (service *invoiceOutputService) Delete(id uint) error {

	invoiceOutput, err := service.invoiceOutputRepo.GetByID(id)
	if err != nil {
		return err
	}

	excelFilePath := filepath.Join("./pkg/excels/output/", invoiceOutput.DeliveryCode+".xlsx")
	if err := os.Remove(excelFilePath); err != nil {
		return err
	}

	return service.invoiceOutputRepo.Delete(id)
}

func (service *invoiceOutputService) Count(projectID uint) (int64, error) {
	return service.invoiceOutputRepo.Count(projectID)
}

func (service *invoiceOutputService) Confirmation(id uint) error {
	invoiceOutput, err := service.invoiceOutputRepo.GetByID(id)
	if err != nil {
		return err
	}
	invoiceOutput.Confirmation = true

	invoiceMaterials, err := service.invoiceMaterialRepo.GetByInvoice(invoiceOutput.ProjectID, invoiceOutput.ID, "output")
	if err != nil {
		return err
	}

	materialsInWarehouse, err := service.materialLocationRepo.GetMaterialsInLocationBasedOnInvoiceID(0, "warehouse", id, "output")
	if err != nil {
		return err
	}

	materialsInTeam, err := service.materialLocationRepo.GetMaterialsInLocationBasedOnInvoiceID(invoiceOutput.TeamID, "team", id, "output")
	if err != nil {
		return err
	}

	for _, invoiceMaterial := range invoiceMaterials {
		materialInWarehouseIndex := -1
		for index, materialInWarehouse := range materialsInWarehouse {
			if materialInWarehouse.MaterialCostID == invoiceMaterial.MaterialCostID {
				materialInWarehouseIndex = index
				break
			}
		}

		materialsInWarehouse[materialInWarehouseIndex].Amount -= invoiceMaterial.Amount

		materialInTeamIndex := -1
		for index, materialInTeam := range materialsInTeam {
			if materialInTeam.MaterialCostID == invoiceMaterial.MaterialCostID {
				materialInTeamIndex = index
				break
			}
		}

		if materialInTeamIndex != -1 {
			materialsInTeam[materialInTeamIndex].Amount += invoiceMaterial.Amount
		} else {
			materialsInTeam = append(materialsInTeam, model.MaterialLocation{
				ProjectID:      invoiceOutput.ProjectID,
				MaterialCostID: invoiceMaterial.MaterialCostID,
				LocationType:   "team",
				LocationID:     invoiceOutput.TeamID,
				Amount:         invoiceMaterial.Amount,
			})
		}
	}

	err = service.invoiceOutputRepo.Confirmation(dto.InvoiceOutputConfirmationQueryData{
		InvoiceData:        invoiceOutput,
		WarehouseMaterials: materialsInWarehouse,
		TeamMaterials:      materialsInTeam,
	})

	return err
}

func (service *invoiceOutputService) UniqueCode(projectID uint) ([]dto.DataForSelect[string], error) {
	return service.invoiceOutputRepo.UniqueCode(projectID)
}

func (service *invoiceOutputService) UniqueWarehouseManager(projectID uint) ([]dto.DataForSelect[uint], error) {
	return service.invoiceOutputRepo.UniqueWarehouseManager(projectID)
}

func (service *invoiceOutputService) UniqueRecieved(projectID uint) ([]dto.DataForSelect[uint], error) {
	return service.invoiceOutputRepo.UniqueRecieved(projectID)
}

func (service *invoiceOutputService) UniqueDistrict(projectID uint) ([]dto.DataForSelect[uint], error) {
	return service.invoiceOutputRepo.UniqueDistrict(projectID)
}

func (service *invoiceOutputService) UniqueTeam(projectID uint) ([]dto.DataForSelect[uint], error) {
	return service.invoiceOutputRepo.UniqueTeam(projectID)
}

func (service *invoiceOutputService) Report(filter dto.InvoiceOutputReportFilterRequest) (string, error) {
	invoices, err := service.invoiceOutputRepo.ReportFilterData(filter)
	if err != nil {
		return "", err
	}

	templateFilePath := filepath.Join("./pkg/excels/templates/", "Invoice Output Report.xlsx")
	f, err := excelize.OpenFile(templateFilePath)
	if err != nil {
		return "", err
	}
	sheetName := "Sheet1"

	rowCount := 2
	for _, invoice := range invoices {
		invoiceMaterialRepo, err := service.invoiceOutputRepo.GetMaterialDataForReport(invoice.ID)
		if err != nil {
			return "", err
		}

		for _, invoiceMaterial := range invoiceMaterialRepo {
			f.SetCellStr(sheetName, "A"+fmt.Sprint(rowCount), invoice.DeliveryCode)
			f.SetCellStr(sheetName, "B"+fmt.Sprint(rowCount), invoice.WarehouseManagerName)
			f.SetCellStr(sheetName, "C"+fmt.Sprint(rowCount), invoice.RecipientName)
			f.SetCellStr(sheetName, "D"+fmt.Sprint(rowCount), invoice.TeamNumber)
			f.SetCellStr(sheetName, "E"+fmt.Sprint(rowCount), invoice.TeamLeaderName)

			dateOfInvoice := invoice.DateOfInvoice.String()
			dateOfInvoice = dateOfInvoice[:len(dateOfInvoice)-10]
			f.SetCellValue(sheetName, "F"+fmt.Sprint(rowCount), dateOfInvoice)

			f.SetCellStr(sheetName, "G"+fmt.Sprint(rowCount), invoiceMaterial.MaterialName)
			f.SetCellStr(sheetName, "H"+fmt.Sprint(rowCount), invoiceMaterial.MaterialUnit)
			f.SetCellFloat(sheetName, "I"+fmt.Sprint(rowCount), invoiceMaterial.Amount, 2, 64)

			materialCostFloat, _ := invoiceMaterial.MaterialCostM19.Float64()
			f.SetCellFloat(sheetName, "J"+fmt.Sprint(rowCount), materialCostFloat, 2, 64)
			f.SetCellValue(sheetName, "K"+fmt.Sprint(rowCount), invoiceMaterial.Notes)
			rowCount++
		}
	}

	currentTime := time.Now()
	fileName := fmt.Sprintf(
		"Отсчет накладной отпуск - %s.xlsx",
		currentTime.Format("02-01-2006"),
	)

	tempFilePath := filepath.Join("./pkg/excels/temp/", fileName)

	f.SaveAs(tempFilePath)

	if err := f.Close(); err != nil {
		fmt.Println(err)
	}

	return fileName, nil
}

func (service *invoiceOutputService) GetTotalMaterialAmount(projectID, materialID uint) (float64, error) {
	return service.materialLocationRepo.GetTotalAmountInWarehouse(projectID, materialID)
}

func (service *invoiceOutputService) GetSerialNumbersByMaterial(projectID, materialID uint) ([]string, error) {
	return service.serialNumberRepo.GetCodesByMaterialID(projectID, materialID, "warehouse")
}

func (service *invoiceOutputService) GetAvailableMaterialsInWarehouse(projectID uint) ([]dto.AvailableMaterialsInWarehouse, error) {
	data, err := service.invoiceOutputRepo.GetAvailableMaterialsInWarehouse(projectID)
	if err != nil {
		return []dto.AvailableMaterialsInWarehouse{}, err
	}

	result := []dto.AvailableMaterialsInWarehouse{}
	currentMaterial := dto.AvailableMaterialsInWarehouse{}
	for index, oneEntry := range data {
		if currentMaterial.ID == oneEntry.ID {
			currentMaterial.Amount += oneEntry.Amount
		} else {
			if index != 0 {
				result = append(result, currentMaterial)
			}
			currentMaterial = oneEntry
		}
	}

	if len(data) != 0 {
		result = append(result, currentMaterial)
	}

	return result, err

}

func (service *invoiceOutputService) GetMaterialsForEdit(id uint) ([]dto.InvoiceOutputMaterialsForEdit, error) {
	data, err := service.invoiceOutputRepo.GetMaterialsForEdit(id)
	if err != nil {
		return []dto.InvoiceOutputMaterialsForEdit{}, nil
	}

	var result []dto.InvoiceOutputMaterialsForEdit
	for index, entry := range data {
		if index == 0 {
			result = append(result, entry)
			continue
		}

		lastItemIndex := len(result) - 1
		if result[lastItemIndex].MaterialID == entry.MaterialID {
			result[lastItemIndex].Amount += entry.Amount
			result[lastItemIndex].WarehouseAmount += entry.WarehouseAmount
		} else {
			result = append(result, entry)
		}
	}

	return result, nil
}

func (service *invoiceOutputService) Import(filePath string, projectID uint, workerID uint) error {
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

	count, err := service.invoiceCountRepo.CountInvoice("output", projectID)
	if err != nil {
		f.Close()
		os.Remove(filePath)
		return fmt.Errorf("Файл не имеет данных")
	}

	index := 1
	importData := []dto.InvoiceOutputImportData{}
	currentInvoiceOutput := model.InvoiceOutput{}
	currentInvoiceMaterials := []model.InvoiceMaterials{}
	for len(rows) > index {
		excelInvoiceOutput := model.InvoiceOutput{
			ID:               0,
			ProjectID:        projectID,
			ReleasedWorkerID: workerID,
			Confirmation:     false,
			Notes:            "",
		}

		districtName, err := f.GetCellValue(sheetName, "L"+fmt.Sprint(index+1))
		if err != nil {
			f.Close()
			os.Remove(filePath)
			return fmt.Errorf("Нету данных в ячейке L%v: %v", index+1, err)
		}

		district, err := service.districtRepo.GetByName(districtName)
		if err != nil {
			f.Close()
			os.Remove(filePath)
			return fmt.Errorf("Имя %v в ячейке L%v не найдено в базе: %v", districtName, index+1, err)
		}

		excelInvoiceOutput.DistrictID = district.ID

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

		excelInvoiceOutput.WarehouseManagerWorkerID = warehouseManager.ID

		recipientName, err := f.GetCellValue(sheetName, "C"+fmt.Sprint(index+1))
		if err != nil {
			f.Close()
			os.Remove(filePath)
			return fmt.Errorf("Нету данных в ячейке C%v: %v", index+1, err)
		}

		recipient, err := service.workerRepo.GetByName(recipientName)
		if err != nil {
			f.Close()
			os.Remove(filePath)
			return fmt.Errorf("Имя %v в ячейке C%v не найдено в базе: %v", recipientName, index+1, err)
		}

		excelInvoiceOutput.RecipientWorkerID = recipient.ID

		teamNumber, err := f.GetCellValue(sheetName, "D"+fmt.Sprint(index+1))
		if err != nil {
			f.Close()
			os.Remove(filePath)
			return fmt.Errorf("Нету данных в ячейке D%v: %v", index+1, err)
		}

		team, err := service.teamRepo.GetByNumber(teamNumber)
		if err != nil {
			f.Close()
			os.Remove(filePath)
			return fmt.Errorf("Имя %v в ячейке D%v не найдено в базе: %v", teamNumber, index+1, err)
		}

		excelInvoiceOutput.TeamID = team.ID

		dateOfInvoiceInExcel, err := f.GetCellValue(sheetName, "F"+fmt.Sprint(index+1))
		if err != nil {
			f.Close()
			os.Remove(filePath)
			return fmt.Errorf("Нету данных в ячейке F%v: %v", index+1, err)
		}

		dateLayout := "2006/01/02"
		dateOfInvoice, err := time.Parse(dateLayout, dateOfInvoiceInExcel)
		if err != nil {
			f.Close()
			os.Remove(filePath)
			return fmt.Errorf("Неправильные данные в ячейке F%v: %v", index+1, err)
		}

		excelInvoiceOutput.DateOfInvoice = dateOfInvoice
		if index == 1 {
			currentInvoiceOutput = excelInvoiceOutput
		}

		excelInvoiceMaterial := model.InvoiceMaterials{
			InvoiceType: "output",
			IsDefected:  false,
			ProjectID:   projectID,
		}

		materialName, err := f.GetCellValue(sheetName, "G"+fmt.Sprint(index+1))
		if err != nil {
			f.Close()
			os.Remove(filePath)
			return fmt.Errorf("Нету данных в ячейке G%v: %v", index+1, err)
		}

		material, err := service.materialRepo.GetByName(materialName)
		if err != nil {
			f.Close()
			os.Remove(filePath)
			return fmt.Errorf("Материал %v в ячейке G%v не найдено в базе: %v", materialName, index+1, err)
		}

		materialCost, err := service.materialCostRepo.GetByMaterialID(material.ID)
		if err != nil {
			f.Close()
			os.Remove(filePath)
			return fmt.Errorf("Цена Материала %v в ячейке G%v не найдено в базе: %v", materialName, index+1, err)
		}

		excelInvoiceMaterial.MaterialCostID = materialCost[0].ID

		amountExcel, err := f.GetCellValue(sheetName, "I"+fmt.Sprint(index+1))
		if err != nil {
			f.Close()
			os.Remove(filePath)
			return fmt.Errorf("Нету данных в ячейке I%v: %v", index+1, err)
		}

		amount, err := strconv.ParseFloat(amountExcel, 64)
		if err != nil {
			f.Close()
			os.Remove(filePath)
			return fmt.Errorf("Нету данных в ячейке I%v: %v", index+1, err)
		}

		excelInvoiceMaterial.Amount = amount

		if currentInvoiceOutput.DateOfInvoice.Equal(excelInvoiceOutput.DateOfInvoice) {
			currentInvoiceMaterials = append(currentInvoiceMaterials, excelInvoiceMaterial)
		} else {
			count++
			currentInvoiceOutput.DeliveryCode = utils.UniqueCodeGeneration("O", int64(count), projectID)
			importData = append(importData, dto.InvoiceOutputImportData{
				Details: currentInvoiceOutput,
				Items:   currentInvoiceMaterials,
			})

			currentInvoiceOutput = excelInvoiceOutput
			currentInvoiceMaterials = []model.InvoiceMaterials{excelInvoiceMaterial}
		}

		index++
	}

	currentInvoiceOutput.DeliveryCode = utils.UniqueCodeGeneration("O", int64(count), projectID)
	importData = append(importData, dto.InvoiceOutputImportData{
		Details: currentInvoiceOutput,
		Items:   currentInvoiceMaterials,
	})

	return service.invoiceOutputRepo.Import(importData)
}

func (service *invoiceOutputService) GenerateExcelFile(data dto.InvoiceOutput) error {

	templateFilePath := filepath.Join("./pkg/excels/templates/output.xlsx")
	f, err := excelize.OpenFile(templateFilePath)
	if err != nil {
		return err
	}

	sheetName := "Отпуск"
	startingRow := 5
	f.InsertRows(sheetName, startingRow, len(data.Items))

	defaultStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Size:      8,
			VertAlign: "center",
		},
		Border: []excelize.Border{
			{Type: "left", Color: "#000000", Style: 1},
			{Type: "top", Color: "#000000", Style: 1},
			{Type: "right", Color: "#000000", Style: 1},
			{Type: "bottom", Color: "#000000", Style: 1},
		},
		Alignment: &excelize.Alignment{
			Horizontal: "left",
			WrapText:   true,
			Vertical:   "center",
		},
	})

	materialNamingStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Size:      8,
			VertAlign: "center",
		},
		Border: []excelize.Border{
			{Type: "left", Color: "#000000", Style: 1},
			{Type: "top", Color: "#000000", Style: 1},
			{Type: "bottom", Color: "#000000", Style: 1},
		},
		Alignment: &excelize.Alignment{
			Horizontal: "left",
			Vertical:   "center",
			WrapText:   true,
		},
	})

	workerNamingStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Size:      9,
			VertAlign: "center",
		},
		Alignment: &excelize.Alignment{
			Horizontal: "left",
			WrapText:   true,
			Vertical:   "center",
		},
	})

	f.SetCellValue(sheetName, "C1", fmt.Sprintf(`НАКЛАДНАЯ 
№ %s
от %s года       
на отпуск материала 
`, data.Details.DeliveryCode, utils.DateConverter(data.Details.DateOfInvoice)))

	project, err := service.projectRepo.GetByID(data.Details.ProjectID)
	if err != nil {
		return err
	}

	f.SetCellValue(sheetName, "C3", fmt.Sprintf("Отпуск разрешил: %s", project.ProjectManager))

	district, err := service.districtRepo.GetByID(data.Details.DistrictID)
	if err != nil {
		return err
	}
	f.MergeCell(sheetName, "D1", "F1")
	f.SetCellStr(sheetName, "D1", fmt.Sprintf(`%s
Регион: %s `, project.Name, district.Name))

	for index, oneEntry := range data.Items {
		material, err := service.materialRepo.GetByID(oneEntry.MaterialID)
		if err != nil {
			return err
		}
		f.SetCellStyle(sheetName, "A"+fmt.Sprint(startingRow+index), "F"+fmt.Sprint(startingRow+index), defaultStyle)
		f.SetCellStyle(sheetName, "B"+fmt.Sprint(startingRow+index), "B"+fmt.Sprint(startingRow+index), materialNamingStyle)

		f.SetCellInt(sheetName, "A"+fmt.Sprint(startingRow+index), index+1)
		f.SetCellStr(sheetName, "B"+fmt.Sprint(startingRow+index), material.Code)
		f.SetCellStr(sheetName, "C"+fmt.Sprint(startingRow+index), material.Name)
		f.SetCellStr(sheetName, "D"+fmt.Sprint(startingRow+index), material.Unit)
		f.SetCellFloat(sheetName, "E"+fmt.Sprint(startingRow+index), oneEntry.Amount, 3, 64)
		f.SetCellStr(sheetName, "F"+fmt.Sprint(startingRow+index), oneEntry.Notes)
	}

	warehouseManager, err := service.workerRepo.GetByID(data.Details.WarehouseManagerWorkerID)
	if err != nil {
		return err
	}
	f.SetCellStyle(sheetName, "C"+fmt.Sprint(6+len(data.Items)), "C"+fmt.Sprint(6+len(data.Items)), workerNamingStyle)
	f.SetCellStr(sheetName, "C"+fmt.Sprint(6+len(data.Items)), warehouseManager.Name)

	released, err := service.workerRepo.GetByID(data.Details.ReleasedWorkerID)
	if err != nil {
		return err
	}
	f.SetCellStyle(sheetName, "C"+fmt.Sprint(8+len(data.Items)), "C"+fmt.Sprint(8+len(data.Items)), workerNamingStyle)
	f.SetCellStr(sheetName, "C"+fmt.Sprint(8+len(data.Items)), released.Name)

	teamData, err := service.teamRepo.GetTeamNumberAndTeamLeadersByID(data.Details.ProjectID, data.Details.TeamID)
	if err != nil {
		return err
	}
	f.SetCellStyle(sheetName, "C"+fmt.Sprint(10+len(data.Items)), "C"+fmt.Sprint(10+len(data.Items)), workerNamingStyle)
	f.SetCellStr(sheetName, "C"+fmt.Sprint(10+len(data.Items)), teamData[0].TeamLeaderName)

	recipient, err := service.workerRepo.GetByID(data.Details.RecipientWorkerID)
	if err != nil {
		return err
	}
	f.SetCellStyle(sheetName, "C"+fmt.Sprint(12+len(data.Items)), "C"+fmt.Sprint(12+len(data.Items)), workerNamingStyle)
	f.SetCellStr(sheetName, "C"+fmt.Sprint(12+len(data.Items)), recipient.Name)

	excelFilePath := filepath.Join("./pkg/excels/output/", data.Details.DeliveryCode+".xlsx")
	if err := f.SaveAs(excelFilePath); err != nil {
		return err
	}

	return nil
}

func (service *invoiceOutputService) GetDocument(deliveryCode string) (string, error) {
	invoiceOutput, err := service.invoiceOutputRepo.GetByDeliveryCode(deliveryCode)
	if err != nil {
		return "", err
	}

	if invoiceOutput.Confirmation {
		return ".pdf", nil
	} else {
		return ".xlsx", nil
	}
}
