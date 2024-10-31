package service

import (
	"backend-v2/internal/repository"
	"fmt"
	"path/filepath"
	"time"

	"github.com/xuri/excelize/v2"
)

type mainReportService struct {
	mainReportRepository repository.IMainReportRepository
}

type IMainReportService interface {
	ProjectProgress(projectID uint) (string, error)
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

	type ProgressReportData struct {
		MaterialID                            uint
		MaterialCode                          string
		MaterialName                          string
		MaterialUnit                          string
		MaterialAmountPlannedForProject       float64
		MaterialAmountRecieved                float64
		MaterialAmountWaitingToBeRecieved     float64
		MaterialAmountInstalled               float64
		MaterialAmountWaitingToBeInstalled    float64
		MaterialAmountInWarehouse             float64
		MaterialAmountInTeams                 float64
		MaterialAmountInObjects               float64
		BudgetOfPlannedMaterials              float64
		BudgetOfRecievedMaterials             float64
		BudgetOfInstalledMaterials            float64
		BudgetOfMaterialsWaitingToBeInstalled float64
	}

	data := []ProgressReportData{}
	dataIndex := 0
	for index, material := range materialData {
		oneEntry := ProgressReportData{
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
		}

		if index == 0 {
			data = append(data, oneEntry)
			continue
		}

		if data[dataIndex].MaterialID == material.ID {
			data[dataIndex].MaterialAmountInWarehouse += oneEntry.MaterialAmountInWarehouse
			data[dataIndex].MaterialAmountInTeams += oneEntry.MaterialAmountInTeams
			data[dataIndex].MaterialAmountInObjects += oneEntry.MaterialAmountInObjects
		} else {
			dataIndex++
			data = append(data, oneEntry)
		}
	}

	for _, invoiceMaterial := range invoiceMaterialData {
		if data[dataIndex].MaterialID != invoiceMaterial.MaterialID {
			dataIndex = -1
			for index, oneEntry := range data {
				if oneEntry.MaterialID == invoiceMaterial.MaterialID {
					dataIndex = index
					break
				}
			}

			if dataIndex == -1 {
				return "", fmt.Errorf("Обнаружен не существующий материал который был использован в накладной")
			}
		}

		switch invoiceMaterial.InvoiceType {
		case "input":
			data[dataIndex].MaterialAmountRecieved += invoiceMaterial.Amount
			data[dataIndex].BudgetOfRecievedMaterials += invoiceMaterial.SumInInvoice
			break
		case "object-correction":
			data[dataIndex].MaterialAmountInstalled += invoiceMaterial.Amount
			data[dataIndex].BudgetOfInstalledMaterials += invoiceMaterial.SumInInvoice
			break
		}
	}

	for index := range data {
		data[index].MaterialAmountWaitingToBeRecieved = data[index].MaterialAmountPlannedForProject - data[index].MaterialAmountRecieved
		data[index].MaterialAmountWaitingToBeInstalled = data[index].MaterialAmountRecieved - data[index].MaterialAmountInstalled
		data[index].BudgetOfMaterialsWaitingToBeInstalled = data[index].BudgetOfRecievedMaterials - data[index].BudgetOfInstalledMaterials
	}

	progressReportFilePath := filepath.Join("./pkg/excels/templates", "Прогресс Проекта.xlsx")
	f, err := excelize.OpenFile(progressReportFilePath)
	if err != nil {
		f.Close()
		return "", fmt.Errorf("Не смог открыть файл: %v", err)
	}

	sheetName := "Материалы"
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
		if err := f.SetCellFloat(sheetName, "E"+fmt.Sprint(startingRow+index), entry.MaterialAmountRecieved, 2, 64); err != nil {
			return "", fmt.Errorf("Ошибка при добавление данных в ячейцку %s%d: %v", "E", startingRow+index, err)
		}
		if err := f.SetCellFloat(sheetName, "F"+fmt.Sprint(startingRow+index), entry.MaterialAmountWaitingToBeRecieved, 2, 64); err != nil {
			return "", fmt.Errorf("Ошибка при добавление данных в ячейцку %s%d: %v", "F", startingRow+index, err)
		}
		if err := f.SetCellFloat(sheetName, "G"+fmt.Sprint(startingRow+index), entry.MaterialAmountInstalled, 2, 64); err != nil {
			return "", fmt.Errorf("Ошибка при добавление данных в ячейцку %s%d: %v", "G", startingRow+index, err)
		}
		if err := f.SetCellFloat(sheetName, "H"+fmt.Sprint(startingRow+index), entry.MaterialAmountWaitingToBeInstalled, 2, 64); err != nil {
			return "", fmt.Errorf("Ошибка при добавление данных в ячейцку %s%d: %v", "H", startingRow+index, err)
		}
		if err := f.SetCellFloat(sheetName, "I"+fmt.Sprint(startingRow+index), entry.MaterialAmountInWarehouse, 2, 64); err != nil {
			return "", fmt.Errorf("Ошибка при добавление данных в ячейцку %s%d: %v", "I", startingRow+index, err)
		}
		if err := f.SetCellFloat(sheetName, "J"+fmt.Sprint(startingRow+index), entry.MaterialAmountInTeams, 2, 64); err != nil {
			return "", fmt.Errorf("Ошибка при добавление данных в ячейцку %s%d: %v", "J", startingRow+index, err)
		}
		if err := f.SetCellFloat(sheetName, "K"+fmt.Sprint(startingRow+index), entry.MaterialAmountInObjects, 2, 64); err != nil {
			return "", fmt.Errorf("Ошибка при добавление данных в ячейцку %s%d: %v", "K", startingRow+index, err)
		}
		if err := f.SetCellFloat(sheetName, "L"+fmt.Sprint(startingRow+index), 0, 2, 64); err != nil {
			return "", fmt.Errorf("Ошибка при добавление данных в ячейцку %s%d: %v", "L", startingRow+index, err)
		}
		if err := f.SetCellFloat(sheetName, "M"+fmt.Sprint(startingRow+index), entry.BudgetOfRecievedMaterials, 2, 64); err != nil {
			return "", fmt.Errorf("Ошибка при добавление данных в ячейцку %s%d: %v", "M", startingRow+index, err)
		}
		if err := f.SetCellFloat(sheetName, "N"+fmt.Sprint(startingRow+index), entry.BudgetOfInstalledMaterials, 2, 64); err != nil {
			return "", fmt.Errorf("Ошибка при добавление данных в ячейцку %s%d: %v", "N", startingRow+index, err)
		}
		if err := f.SetCellFloat(sheetName, "O"+fmt.Sprint(startingRow+index), entry.BudgetOfMaterialsWaitingToBeInstalled, 2, 64); err != nil {
			return "", fmt.Errorf("Ошибка при добавление данных в ячейцку %s%d: %v", "O", startingRow+index, err)
		}
	}

	currentTime := time.Now()
	progressReportTmpFileName := fmt.Sprintf(
		"Прогресс Проекта - %s.xlsx",
		currentTime.Format("02-01-2006"),
	)

	progressReportTmpFilePath := filepath.Join("./pkg/excels/temp/", progressReportTmpFileName)
	if err := f.SaveAs(progressReportTmpFilePath); err != nil {
		return "", err
	}

	return progressReportTmpFilePath, nil
}
