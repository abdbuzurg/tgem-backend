package service

import (
	"backend-v2/internal/dto"
	"backend-v2/internal/repository"
	"fmt"
	"path/filepath"
	"time"

	"github.com/shopspring/decimal"
	"github.com/xuri/excelize/v2"
)

type mainReportService struct {
	mainReportRepository repository.IMainReportRepository
}

type IMainReportService interface {
	ProjectProgress(projectID uint) (string, error)
	ProjectProgressByGivenDay(projectID uint, date time.Time) (string, error)
	RemainingMaterialAnalysis(projectID uint) (string, error)
}

func InitMainReportService(mainReportRepository repository.IMainReportRepository) IMainReportService {
	return &mainReportService{
		mainReportRepository: mainReportRepository,
	}
}

func (service *mainReportService) ProjectProgress(projectID uint) (string, error) {
	materialData, err := service.mainReportRepository.MaterialDataForProgressReportInProject(projectID)
	if err != nil {
		return "", err
	}

	invoiceMaterialData, err := service.mainReportRepository.InvoiceMaterialDataForProgressReport(projectID)
	if err != nil {
		return "", err
	}

	dataForMaterials := []dto.ProgressReportData{}
	dataIndex := -1
	for _, material := range materialData {
		oneEntry := dto.ProgressReportData{
			MaterialID:                      material.ID,
			MaterialCode:                    material.Code,
			MaterialName:                    material.Name,
			MaterialUnit:                    material.Unit,
			MaterialAmountPlannedForProject: material.PlannedAmountForProject,
		}

		switch material.LocationType {
		case "warehouse":
			oneEntry.MaterialAmountInWarehouse = material.LocationAmount
			break
		case "team":
			oneEntry.MaterialAmountInTeams = material.LocationAmount
			break
		case "object":
			oneEntry.MaterialAmountInObjects = material.LocationAmount
			break
		case "loss-warehouse", "writeoff-warehouse", "loss-team", "loss-object", "writeoff-object":
			oneEntry.MaterialAmountInAllWriteOffs = material.LocationAmount
			break
		}

		if dataIndex == -1 {
			dataForMaterials = append(dataForMaterials, oneEntry)
			dataIndex++
			continue
		}

		if dataForMaterials[dataIndex].MaterialID == oneEntry.MaterialID {
			dataForMaterials[dataIndex].MaterialAmountInWarehouse += oneEntry.MaterialAmountInWarehouse
			dataForMaterials[dataIndex].MaterialAmountInTeams += oneEntry.MaterialAmountInTeams
			dataForMaterials[dataIndex].MaterialAmountInObjects += oneEntry.MaterialAmountInObjects
			dataForMaterials[dataIndex].MaterialAmountInAllWriteOffs += oneEntry.MaterialAmountInAllWriteOffs
		} else {
			dataIndex++
			dataForMaterials = append(dataForMaterials, oneEntry)
		}
	}

	for _, invoiceMaterial := range invoiceMaterialData {
		for index, oneEntry := range dataForMaterials {
			if oneEntry.MaterialID == invoiceMaterial.MaterialID {
				switch invoiceMaterial.InvoiceType {
				case "input":
					dataForMaterials[index].MaterialAmountRecieved += invoiceMaterial.Amount
					dataForMaterials[index].BudgetOfRecievedMaterials = decimal.Sum(dataForMaterials[index].BudgetOfRecievedMaterials, invoiceMaterial.CostWithCustomer.Mul(decimal.NewFromFloat(invoiceMaterial.Amount)))
					break
				case "object-correction":
					dataForMaterials[index].MaterialAmountInstalled += invoiceMaterial.Amount
					dataForMaterials[index].BudgetOfInstalledMaterials = decimal.Sum(dataForMaterials[index].BudgetOfInstalledMaterials, invoiceMaterial.CostWithCustomer.Mul(decimal.NewFromFloat(invoiceMaterial.Amount)))
					break
				}
				break
			}
		}

	}

	for index := range dataForMaterials {
		dataForMaterials[index].MaterialAmountWaitingToBeRecieved = dataForMaterials[index].MaterialAmountPlannedForProject - dataForMaterials[index].MaterialAmountRecieved
		dataForMaterials[index].MaterialAmountWaitingToBeInstalled = dataForMaterials[index].MaterialAmountRecieved - dataForMaterials[index].MaterialAmountInstalled
		dataForMaterials[index].BudgetOfMaterialsWaitingToBeInstalled = dataForMaterials[index].BudgetOfRecievedMaterials.Sub(dataForMaterials[index].BudgetOfInstalledMaterials)
	}

	invoiceOperationData, err := service.mainReportRepository.InvoiceOperationDataForProgressReport(projectID)
	if err != nil {
		return "", err
	}

	dataForOperation := []dto.InvoiceOperationDataForProgressReportQueryResult{}
	dataForOperationIndex := 0
	for index, invoiceOperation := range invoiceOperationData {
		if index == 0 {
			dataForOperation = append(dataForOperation, invoiceOperation)
			continue
		}

		if dataForOperation[dataForOperationIndex].ID == invoiceOperation.ID {
			dataForOperation[dataForOperationIndex].AmountInInvoice += invoiceOperation.AmountInInvoice
		} else {
			dataForOperation = append(dataForOperation, invoiceOperation)
			dataForOperationIndex++
		}
	}

	progressReportFilePath := filepath.Join("./pkg/excels/templates", "Прогресс Проекта.xlsx")
	f, err := excelize.OpenFile(progressReportFilePath)
	if err != nil {
		f.Close()
		return "", fmt.Errorf("Не смог открыть файл: %v", err)
	}

	sheetNameForMaterials := "Материалы"
	startingRow := 2

	for index, entry := range dataForMaterials {
		if err := f.SetCellStr(sheetNameForMaterials, "A"+fmt.Sprint(startingRow+index), entry.MaterialCode); err != nil {
			return "", fmt.Errorf("Ошибка при добавление данных в ячейцку %s%d: %v", "A", startingRow+index, err)
		}
		if err := f.SetCellStr(sheetNameForMaterials, "B"+fmt.Sprint(startingRow+index), entry.MaterialName); err != nil {
			return "", fmt.Errorf("Ошибка при добавление данных в ячейцку %s%d: %v", "B", startingRow+index, err)
		}
		if err := f.SetCellStr(sheetNameForMaterials, "C"+fmt.Sprint(startingRow+index), entry.MaterialUnit); err != nil {
			return "", fmt.Errorf("Ошибка при добавление данных в ячейцку %s%d: %v", "C", startingRow+index, err)
		}
		if err := f.SetCellFloat(sheetNameForMaterials, "D"+fmt.Sprint(startingRow+index), entry.MaterialAmountPlannedForProject, 2, 64); err != nil {
			return "", fmt.Errorf("Ошибка при добавление данных в ячейцку %s%d: %v", "D", startingRow+index, err)
		}
		if err := f.SetCellFloat(sheetNameForMaterials, "E"+fmt.Sprint(startingRow+index), entry.MaterialAmountRecieved, 2, 64); err != nil {
			return "", fmt.Errorf("Ошибка при добавление данных в ячейцку %s%d: %v", "E", startingRow+index, err)
		}
		if err := f.SetCellFloat(sheetNameForMaterials, "F"+fmt.Sprint(startingRow+index), entry.MaterialAmountWaitingToBeRecieved, 2, 64); err != nil {
			return "", fmt.Errorf("Ошибка при добавление данных в ячейцку %s%d: %v", "F", startingRow+index, err)
		}
		if err := f.SetCellFloat(sheetNameForMaterials, "G"+fmt.Sprint(startingRow+index), entry.MaterialAmountInstalled, 2, 64); err != nil {
			return "", fmt.Errorf("Ошибка при добавление данных в ячейцку %s%d: %v", "G", startingRow+index, err)
		}
		if err := f.SetCellFloat(sheetNameForMaterials, "H"+fmt.Sprint(startingRow+index), entry.MaterialAmountWaitingToBeInstalled, 2, 64); err != nil {
			return "", fmt.Errorf("Ошибка при добавление данных в ячейцку %s%d: %v", "H", startingRow+index, err)
		}
		if err := f.SetCellFloat(sheetNameForMaterials, "I"+fmt.Sprint(startingRow+index), entry.MaterialAmountInWarehouse, 2, 64); err != nil {
			return "", fmt.Errorf("Ошибка при добавление данных в ячейцку %s%d: %v", "I", startingRow+index, err)
		}
		if err := f.SetCellFloat(sheetNameForMaterials, "J"+fmt.Sprint(startingRow+index), entry.MaterialAmountInTeams, 2, 64); err != nil {
			return "", fmt.Errorf("Ошибка при добавление данных в ячейцку %s%d: %v", "J", startingRow+index, err)
		}
		if err := f.SetCellFloat(sheetNameForMaterials, "K"+fmt.Sprint(startingRow+index), entry.MaterialAmountInObjects, 2, 64); err != nil {
			return "", fmt.Errorf("Ошибка при добавление данных в ячейцку %s%d: %v", "K", startingRow+index, err)
		}
		if err := f.SetCellFloat(sheetNameForMaterials, "L"+fmt.Sprint(startingRow+index), entry.MaterialAmountInAllWriteOffs, 2, 64); err != nil {
			return "", fmt.Errorf("Ошибка при добавление данных в ячейцку %s%d: %v", "L", startingRow+index, err)
		}

		budgetOfRecievedMaterials, _ := entry.BudgetOfRecievedMaterials.Float64()
		if err := f.SetCellFloat(sheetNameForMaterials, "M"+fmt.Sprint(startingRow+index), budgetOfRecievedMaterials, 2, 64); err != nil {
			return "", fmt.Errorf("Ошибка при добавление данных в ячейцку %s%d: %v", "M", startingRow+index, err)
		}

		budgetOfInstalledMaterials, _ := entry.BudgetOfInstalledMaterials.Float64()
		if err := f.SetCellFloat(sheetNameForMaterials, "N"+fmt.Sprint(startingRow+index), budgetOfInstalledMaterials, 2, 64); err != nil {
			return "", fmt.Errorf("Ошибка при добавление данных в ячейцку %s%d: %v", "N", startingRow+index, err)
		}

		budgetOfMaterialsWaitingToBeInstalled, _ := entry.BudgetOfMaterialsWaitingToBeInstalled.Float64()
		if err := f.SetCellFloat(sheetNameForMaterials, "O"+fmt.Sprint(startingRow+index), budgetOfMaterialsWaitingToBeInstalled, 2, 64); err != nil {
			return "", fmt.Errorf("Ошибка при добавление данных в ячейцку %s%d: %v", "O", startingRow+index, err)
		}
	}

	sheetNameForOperations := "Услуги"

	for index, entry := range dataForOperation {
		if err := f.SetCellStr(sheetNameForOperations, "A"+fmt.Sprint(startingRow+index), entry.Code); err != nil {
			return "", fmt.Errorf("Ошибка при добавление данных в ячейцку %s%d: %v", "A", startingRow+index, err)
		}
		if err := f.SetCellStr(sheetNameForOperations, "B"+fmt.Sprint(startingRow+index), entry.Name); err != nil {
			return "", fmt.Errorf("Ошибка при добавление данных в ячейцку %s%d: %v", "B", startingRow+index, err)
		}
		if err := f.SetCellFloat(sheetNameForOperations, "C"+fmt.Sprint(startingRow+index), entry.PlannedAmountForProject, 2, 64); err != nil {
			return "", fmt.Errorf("Ошибка при добавление данных в ячейцку %s%d: %v", "C", startingRow+index, err)
		}
		if err := f.SetCellFloat(sheetNameForOperations, "D"+fmt.Sprint(startingRow+index), entry.AmountInInvoice, 2, 64); err != nil {
			return "", fmt.Errorf("Ошибка при добавление данных в ячейцку %s%d: %v", "D", startingRow+index, err)
		}
		if err := f.SetCellFloat(sheetNameForOperations, "E"+fmt.Sprint(startingRow+index), entry.PlannedAmountForProject-entry.AmountInInvoice, 2, 64); err != nil {
			return "", fmt.Errorf("Ошибка при добавление данных в ячейцку %s%d: %v", "E", startingRow+index, err)
		}

		budgetOfCompletedOperations, _ := entry.CostWithCustomer.Mul(decimal.NewFromFloat(entry.AmountInInvoice)).Float64()
		if err := f.SetCellFloat(sheetNameForOperations, "F"+fmt.Sprint(startingRow+index), budgetOfCompletedOperations, 2, 64); err != nil {
			return "", fmt.Errorf("Ошибка при добавление данных в ячейцку %s%d: %v", "F", startingRow+index, err)
		}

		budgetOfPlannedOperationsForProject, _ := entry.CostWithCustomer.Mul(decimal.NewFromFloat(entry.PlannedAmountForProject - entry.AmountInInvoice)).Float64()
		if err := f.SetCellFloat(sheetNameForOperations, "G"+fmt.Sprint(startingRow+index), budgetOfPlannedOperationsForProject, 2, 64); err != nil {
			return "", fmt.Errorf("Ошибка при добавление данных в ячейцку %s%d: %v", "G", startingRow+index, err)
		}
	}

	currentTime := time.Now()
	progressReportTmpFileName := fmt.Sprintf(
		"Прогресс Проекта - %s.xlsx",
		currentTime.Format("02-01-2006 15-04-05"),
	)

	progressReportTmpFilePath := filepath.Join("./pkg/excels/temp/", progressReportTmpFileName)
	if err := f.SaveAs(progressReportTmpFilePath); err != nil {
		return "", err
	}

	return progressReportTmpFilePath, nil
}

func (service *mainReportService) ProjectProgressByGivenDay(projectID uint, date time.Time) (string, error) {
	materialData, err := service.mainReportRepository.MaterialDataForProgressReportInProjectInGivenDate(projectID, date)
	if err != nil {
		return "", err
	}
  fmt.Println(materialData[0])

	dataForMaterials := []dto.ProgressReportData{
		{
			MaterialID:                         materialData[0].ID,
			MaterialCode:                       materialData[0].Code,
			MaterialName:                       materialData[0].Name,
			MaterialUnit:                       materialData[0].Unit,
			MaterialAmountPlannedForProject:    materialData[0].AmountPlannedForProject,
			MaterialAmountRecieved:             materialData[0].AmountReceived,
			MaterialAmountWaitingToBeRecieved:  materialData[0].AmountPlannedForProject - materialData[0].AmountReceived,
			MaterialAmountInstalled:            materialData[0].AmountInstalled,
			MaterialAmountWaitingToBeInstalled: materialData[0].AmountReceived - materialData[0].AmountInstalled,
			MaterialAmountInWarehouse:          materialData[0].AmountInWarehouse,
			MaterialAmountInTeams:              materialData[0].AmountInTeams,
			MaterialAmountInObjects:            materialData[0].AmountInObjects,
			MaterialAmountInAllWriteOffs:       materialData[0].AmountWriteOff,
			BudgetOfRecievedMaterials:          materialData[0].CostWithCustomer.Mul(decimal.NewFromFloat(materialData[0].AmountReceived)),
			BudgetOfInstalledMaterials:         materialData[0].CostWithCustomer.Mul(decimal.NewFromFloat(materialData[0].AmountInstalled)),
		},
	}
	dataForMaterialsIndex := 0
	for _, material := range materialData[1:] {
		if dataForMaterials[dataForMaterialsIndex].MaterialID == material.ID {
			dataForMaterials[dataForMaterialsIndex].MaterialAmountRecieved += material.AmountReceived
			dataForMaterials[dataForMaterialsIndex].MaterialAmountWaitingToBeRecieved -= material.AmountReceived
			dataForMaterials[dataForMaterialsIndex].MaterialAmountInstalled += material.AmountInstalled
			dataForMaterials[dataForMaterialsIndex].MaterialAmountWaitingToBeInstalled += material.AmountReceived - material.AmountInstalled
			dataForMaterials[dataForMaterialsIndex].MaterialAmountInWarehouse += material.AmountInWarehouse
			dataForMaterials[dataForMaterialsIndex].MaterialAmountInTeams += material.AmountInTeams
			dataForMaterials[dataForMaterialsIndex].MaterialAmountInObjects += material.AmountInObjects
			dataForMaterials[dataForMaterialsIndex].MaterialAmountInAllWriteOffs += material.AmountWriteOff
			dataForMaterials[dataForMaterialsIndex].BudgetOfRecievedMaterials = decimal.Sum(dataForMaterials[dataForMaterialsIndex].BudgetOfRecievedMaterials, material.CostWithCustomer.Mul(decimal.NewFromFloat(material.AmountReceived)))
			dataForMaterials[dataForMaterialsIndex].BudgetOfInstalledMaterials = decimal.Sum(dataForMaterials[dataForMaterialsIndex].BudgetOfInstalledMaterials, material.CostWithCustomer.Mul(decimal.NewFromFloat(material.AmountInstalled)))
		} else {
			dataForMaterials = append(dataForMaterials, dto.ProgressReportData{
				MaterialID:                         material.ID,
				MaterialCode:                       material.Code,
				MaterialName:                       material.Name,
				MaterialUnit:                       material.Unit,
				MaterialAmountPlannedForProject:    material.AmountPlannedForProject,
				MaterialAmountRecieved:             material.AmountReceived,
				MaterialAmountWaitingToBeRecieved:  material.AmountPlannedForProject - material.AmountReceived,
				MaterialAmountInstalled:            material.AmountInstalled,
				MaterialAmountWaitingToBeInstalled: material.AmountReceived - material.AmountInstalled,
				MaterialAmountInWarehouse:          material.AmountInWarehouse,
				MaterialAmountInTeams:              material.AmountInTeams,
				MaterialAmountInObjects:            material.AmountInObjects,
				MaterialAmountInAllWriteOffs:       material.AmountWriteOff,
				BudgetOfRecievedMaterials:          material.CostWithCustomer.Mul(decimal.NewFromFloat(material.AmountReceived)),
				BudgetOfInstalledMaterials:         material.CostWithCustomer.Mul(decimal.NewFromFloat(material.AmountInstalled)),
			})
			dataForMaterialsIndex++
		}
	}

	for index := range dataForMaterials {
		dataForMaterials[index].BudgetOfMaterialsWaitingToBeInstalled = dataForMaterials[index].BudgetOfRecievedMaterials.Sub(dataForMaterials[index].BudgetOfInstalledMaterials)
	}

	dataForOperation, err := service.mainReportRepository.InvoiceOperationDataForProgressReportInGivenDate(projectID, date)
	if err != nil {
		return "", err
	}

	progressReportFilePath := filepath.Join("./pkg/excels/templates", "Прогресс Проекта.xlsx")
	f, err := excelize.OpenFile(progressReportFilePath)
	if err != nil {
		f.Close()
		return "", fmt.Errorf("Не смог открыть файл: %v", err)
	}

	sheetNameForMaterials := "Материалы"
	startingRow := 2

	for index, entry := range dataForMaterials {
		if err := f.SetCellStr(sheetNameForMaterials, "A"+fmt.Sprint(startingRow+index), entry.MaterialCode); err != nil {
			return "", fmt.Errorf("Ошибка при добавление данных в ячейцку %s%d: %v", "A", startingRow+index, err)
		}
		if err := f.SetCellStr(sheetNameForMaterials, "B"+fmt.Sprint(startingRow+index), entry.MaterialName); err != nil {
			return "", fmt.Errorf("Ошибка при добавление данных в ячейцку %s%d: %v", "B", startingRow+index, err)
		}
		if err := f.SetCellStr(sheetNameForMaterials, "C"+fmt.Sprint(startingRow+index), entry.MaterialUnit); err != nil {
			return "", fmt.Errorf("Ошибка при добавление данных в ячейцку %s%d: %v", "C", startingRow+index, err)
		}
		if err := f.SetCellFloat(sheetNameForMaterials, "D"+fmt.Sprint(startingRow+index), entry.MaterialAmountPlannedForProject, 2, 64); err != nil {
			return "", fmt.Errorf("Ошибка при добавление данных в ячейцку %s%d: %v", "D", startingRow+index, err)
		}
		if err := f.SetCellFloat(sheetNameForMaterials, "E"+fmt.Sprint(startingRow+index), entry.MaterialAmountRecieved, 2, 64); err != nil {
			return "", fmt.Errorf("Ошибка при добавление данных в ячейцку %s%d: %v", "E", startingRow+index, err)
		}
		if err := f.SetCellFloat(sheetNameForMaterials, "F"+fmt.Sprint(startingRow+index), entry.MaterialAmountWaitingToBeRecieved, 2, 64); err != nil {
			return "", fmt.Errorf("Ошибка при добавление данных в ячейцку %s%d: %v", "F", startingRow+index, err)
		}
		if err := f.SetCellFloat(sheetNameForMaterials, "G"+fmt.Sprint(startingRow+index), entry.MaterialAmountInstalled, 2, 64); err != nil {
			return "", fmt.Errorf("Ошибка при добавление данных в ячейцку %s%d: %v", "G", startingRow+index, err)
		}
		if err := f.SetCellFloat(sheetNameForMaterials, "H"+fmt.Sprint(startingRow+index), entry.MaterialAmountWaitingToBeInstalled, 2, 64); err != nil {
			return "", fmt.Errorf("Ошибка при добавление данных в ячейцку %s%d: %v", "H", startingRow+index, err)
		}
		if err := f.SetCellFloat(sheetNameForMaterials, "I"+fmt.Sprint(startingRow+index), entry.MaterialAmountInWarehouse, 2, 64); err != nil {
			return "", fmt.Errorf("Ошибка при добавление данных в ячейцку %s%d: %v", "I", startingRow+index, err)
		}
		if err := f.SetCellFloat(sheetNameForMaterials, "J"+fmt.Sprint(startingRow+index), entry.MaterialAmountInTeams, 2, 64); err != nil {
			return "", fmt.Errorf("Ошибка при добавление данных в ячейцку %s%d: %v", "J", startingRow+index, err)
		}
		if err := f.SetCellFloat(sheetNameForMaterials, "K"+fmt.Sprint(startingRow+index), entry.MaterialAmountInObjects, 2, 64); err != nil {
			return "", fmt.Errorf("Ошибка при добавление данных в ячейцку %s%d: %v", "K", startingRow+index, err)
		}
		if err := f.SetCellFloat(sheetNameForMaterials, "L"+fmt.Sprint(startingRow+index), entry.MaterialAmountInAllWriteOffs, 2, 64); err != nil {
			return "", fmt.Errorf("Ошибка при добавление данных в ячейцку %s%d: %v", "L", startingRow+index, err)
		}

		budgetOfRecievedMaterials, _ := entry.BudgetOfRecievedMaterials.Float64()
		if err := f.SetCellFloat(sheetNameForMaterials, "M"+fmt.Sprint(startingRow+index), budgetOfRecievedMaterials, 2, 64); err != nil {
			return "", fmt.Errorf("Ошибка при добавление данных в ячейцку %s%d: %v", "M", startingRow+index, err)
		}

		budgetOfInstalledMaterials, _ := entry.BudgetOfInstalledMaterials.Float64()
		if err := f.SetCellFloat(sheetNameForMaterials, "N"+fmt.Sprint(startingRow+index), budgetOfInstalledMaterials, 2, 64); err != nil {
			return "", fmt.Errorf("Ошибка при добавление данных в ячейцку %s%d: %v", "N", startingRow+index, err)
		}

		budgetOfMaterialsWaitingToBeInstalled, _ := entry.BudgetOfMaterialsWaitingToBeInstalled.Float64()
		if err := f.SetCellFloat(sheetNameForMaterials, "O"+fmt.Sprint(startingRow+index), budgetOfMaterialsWaitingToBeInstalled, 2, 64); err != nil {
			return "", fmt.Errorf("Ошибка при добавление данных в ячейцку %s%d: %v", "O", startingRow+index, err)
		}
	}

	sheetNameForOperations := "Услуги"

	for index, entry := range dataForOperation {
		if err := f.SetCellStr(sheetNameForOperations, "A"+fmt.Sprint(startingRow+index), entry.Code); err != nil {
			return "", fmt.Errorf("Ошибка при добавление данных в ячейцку %s%d: %v", "A", startingRow+index, err)
		}
		if err := f.SetCellStr(sheetNameForOperations, "B"+fmt.Sprint(startingRow+index), entry.Name); err != nil {
			return "", fmt.Errorf("Ошибка при добавление данных в ячейцку %s%d: %v", "B", startingRow+index, err)
		}
		if err := f.SetCellFloat(sheetNameForOperations, "C"+fmt.Sprint(startingRow+index), entry.AmountPlannedForProject, 2, 64); err != nil {
			return "", fmt.Errorf("Ошибка при добавление данных в ячейцку %s%d: %v", "C", startingRow+index, err)
		}
		if err := f.SetCellFloat(sheetNameForOperations, "D"+fmt.Sprint(startingRow+index), entry.AmountInstalled, 2, 64); err != nil {
			return "", fmt.Errorf("Ошибка при добавление данных в ячейцку %s%d: %v", "D", startingRow+index, err)
		}
		if err := f.SetCellFloat(sheetNameForOperations, "E"+fmt.Sprint(startingRow+index), entry.AmountPlannedForProject-entry.AmountInstalled, 2, 64); err != nil {
			return "", fmt.Errorf("Ошибка при добавление данных в ячейцку %s%d: %v", "E", startingRow+index, err)
		}

		budgetOfCompletedOperations, _ := entry.CostWithCustomer.Mul(decimal.NewFromFloat(entry.AmountInstalled)).Float64()
		if err := f.SetCellFloat(sheetNameForOperations, "F"+fmt.Sprint(startingRow+index), budgetOfCompletedOperations, 2, 64); err != nil {
			return "", fmt.Errorf("Ошибка при добавление данных в ячейцку %s%d: %v", "F", startingRow+index, err)
		}

		budgetOfPlannedOperationsForProject, _ := entry.CostWithCustomer.Mul(decimal.NewFromFloat(entry.AmountPlannedForProject - entry.AmountInstalled)).Float64()
		if err := f.SetCellFloat(sheetNameForOperations, "G"+fmt.Sprint(startingRow+index), budgetOfPlannedOperationsForProject, 2, 64); err != nil {
			return "", fmt.Errorf("Ошибка при добавление данных в ячейцку %s%d: %v", "G", startingRow+index, err)
		}
	}

	currentTime := time.Now()
	progressReportTmpFileName := fmt.Sprintf(
		"Прогресс Проекта - %s.xlsx",
		currentTime.Format("02-01-2006 15-04-05"),
	)

	progressReportTmpFilePath := filepath.Join("./pkg/excels/temp/", progressReportTmpFileName)
	if err := f.SaveAs(progressReportTmpFilePath); err != nil {
		return "", err
	}

	return progressReportTmpFilePath, nil
}

func (service *mainReportService) RemainingMaterialAnalysis(projectID uint) (string, error) {
	materialLocationData, err := service.mainReportRepository.MaterialDataForRemainingMaterialAnalysis(projectID)
	if err != nil {
		return "", err
	}

	materialInstalledOnObject, err := service.mainReportRepository.MaterialsInstalledOnObjectForRemainingMaterialAnalysis(projectID)
	if err != nil {
		return "", err
	}

	type RemainingMaterialForAnalysis struct {
		MaterialID                             uint
		MaterialCode                           string
		MaterialName                           string
		MaterialUnit                           string
		MaterialAmountPlannedForProject        float64
		MaterialAmountInBothWarehouseAndTeam   float64
		MaterialTotalAmountInstalledIn10Days   float64
		AverageMaterialAmountInstalledIn10Days float64
		DaysRemainingForMaterialToBeSufficient float64
		MaterialsAmountInObject                float64
		MaterialAmountWaitingToBeInstalled     float64
		MaterialAmountWaitingToBeBought        float64
	}

	data := []RemainingMaterialForAnalysis{}
	dataIndex := -1
	for _, material := range materialLocationData {
		entry := RemainingMaterialForAnalysis{
			MaterialID:                      material.ID,
			MaterialCode:                    material.Code,
			MaterialName:                    material.Name,
			MaterialUnit:                    material.Unit,
			MaterialAmountPlannedForProject: material.PlannedAmountForProject,
		}

		entry.MaterialAmountInBothWarehouseAndTeam = material.LocationAmount

		if dataIndex == -1 {
			data = append(data, entry)
			dataIndex++
			continue
		}

		if data[dataIndex].MaterialID == entry.MaterialID {
			data[dataIndex].MaterialAmountInBothWarehouseAndTeam += entry.MaterialAmountInBothWarehouseAndTeam
			data[dataIndex].MaterialsAmountInObject += entry.MaterialsAmountInObject
		} else {
			data = append(data, entry)
			dataIndex++
		}
	}

	loc, _ := time.LoadLocation("UTC")
	dateNow := time.Now().In(loc)
	date10DaysAgo := dateNow.AddDate(0, 0, -10)

	for _, material := range materialInstalledOnObject {
		if data[dataIndex].MaterialID != material.ID {
			dataIndex = -1
			for index, oneEntry := range data {
				if oneEntry.MaterialID == material.ID {
					dataIndex = index
					break
				}
			}

			if dataIndex == -1 {
				return "", fmt.Errorf("Обнаружен не существующий материал который был использован в накладной")
			}
		}

		data[dataIndex].MaterialsAmountInObject += material.Amount
		if material.DateOfCorrection.In(loc).After(date10DaysAgo) && material.DateOfCorrection.In(loc).Before(dateNow) {
			data[dataIndex].MaterialTotalAmountInstalledIn10Days += material.Amount
		}
	}

	for index := range data {
		data[index].AverageMaterialAmountInstalledIn10Days = data[index].MaterialTotalAmountInstalledIn10Days / 10
		if data[index].AverageMaterialAmountInstalledIn10Days != 0 {
			data[index].DaysRemainingForMaterialToBeSufficient = data[index].MaterialAmountInBothWarehouseAndTeam / data[index].AverageMaterialAmountInstalledIn10Days
		} else {
			data[index].DaysRemainingForMaterialToBeSufficient = 99999
		}
		data[index].MaterialAmountWaitingToBeInstalled = data[index].MaterialAmountPlannedForProject - data[index].MaterialsAmountInObject
		data[index].MaterialAmountWaitingToBeBought = data[index].MaterialAmountPlannedForProject - data[index].MaterialAmountInBothWarehouseAndTeam - data[index].MaterialsAmountInObject
	}

	remainingMaterialAnalysisPath := filepath.Join("./pkg/excels/templates", "Анализ Остатка Материалов.xlsx")
	f, err := excelize.OpenFile(remainingMaterialAnalysisPath)
	if err != nil {
		f.Close()
		return "", fmt.Errorf("Не смог открыть файл: %v", err)
	}

	sheetName := "Анализ"
	startingRow := 2

	for index, entry := range data {
		if err := f.SetCellStr(sheetName, "A"+fmt.Sprint(startingRow+index), entry.MaterialCode); err != nil {
			return "", fmt.Errorf("Ошибка при добавление данных в ячейцку %s%d: %v", "A", startingRow+index, err)
		}
		if err := f.SetCellStr(sheetName, "B"+fmt.Sprint(startingRow+index), entry.MaterialName); err != nil {
			return "", fmt.Errorf("Ошибка при добавление данных в ячейцку %s%d: %v", "B", startingRow+index, err)
		}
		if err := f.SetCellStr(sheetName, "C"+fmt.Sprint(startingRow+index), entry.MaterialUnit); err != nil {
			return "", fmt.Errorf("Ошибка при добавление данных в ячейцку %s%d: %v", "C", startingRow+index, err)
		}
		if err := f.SetCellFloat(sheetName, "D"+fmt.Sprint(startingRow+index), entry.MaterialAmountPlannedForProject, 2, 64); err != nil {
			return "", fmt.Errorf("Ошибка при добавление данных в ячейцку %s%d: %v", "D", startingRow+index, err)
		}
		if err := f.SetCellFloat(sheetName, "E"+fmt.Sprint(startingRow+index), entry.MaterialAmountInBothWarehouseAndTeam, 2, 64); err != nil {
			return "", fmt.Errorf("Ошибка при добавление данных в ячейцку %s%d: %v", "E", startingRow+index, err)
		}
		if err := f.SetCellFloat(sheetName, "F"+fmt.Sprint(startingRow+index), entry.AverageMaterialAmountInstalledIn10Days, 2, 64); err != nil {
			return "", fmt.Errorf("Ошибка при добавление данных в ячейцку %s%d: %v", "F", startingRow+index, err)
		}
		if err := f.SetCellFloat(sheetName, "G"+fmt.Sprint(startingRow+index), entry.DaysRemainingForMaterialToBeSufficient, 2, 64); err != nil {
			return "", fmt.Errorf("Ошибка при добавление данных в ячейцку %s%d: %v", "G", startingRow+index, err)
		}
		if err := f.SetCellFloat(sheetName, "H"+fmt.Sprint(startingRow+index), entry.MaterialsAmountInObject, 2, 64); err != nil {
			return "", fmt.Errorf("Ошибка при добавление данных в ячейцку %s%d: %v", "H", startingRow+index, err)
		}
		if err := f.SetCellFloat(sheetName, "I"+fmt.Sprint(startingRow+index), entry.MaterialAmountWaitingToBeInstalled, 2, 64); err != nil {
			return "", fmt.Errorf("Ошибка при добавление данных в ячейцку %s%d: %v", "I", startingRow+index, err)
		}
		if err := f.SetCellFloat(sheetName, "J"+fmt.Sprint(startingRow+index), entry.MaterialAmountWaitingToBeBought, 2, 64); err != nil {
			return "", fmt.Errorf("Ошибка при добавление данных в ячейцку %s%d: %v", "J", startingRow+index, err)
		}
	}

	currentTime := time.Now()
	remainingMaterialAnalysisTmpFileName := fmt.Sprintf(
		"Анализ Остатка Материалов - %s.xlsx",
		currentTime.Format("02-01-2006"),
	)

	remainingMaterialAnalysisTmpFilePath := filepath.Join("./pkg/excels/temp/", remainingMaterialAnalysisTmpFileName)
	if err := f.SaveAs(remainingMaterialAnalysisTmpFilePath); err != nil {
		return "", err
	}

	return remainingMaterialAnalysisTmpFilePath, err
}
