package service

import (
	"backend-v2/internal/dto"
	"backend-v2/internal/repository"
	"backend-v2/model"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/shopspring/decimal"
	"github.com/xuri/excelize/v2"
)

type operationService struct {
	operationRepo repository.IOperationRepository
	materialRepo  repository.IMaterialRepository
}

func InitOperationService(
  operationRepo repository.IOperationRepository,
  materialRepo repository.IMaterialRepository,
) IOperationService {
	return &operationService{
		operationRepo: operationRepo,
    materialRepo: materialRepo,
	}
}

type IOperationService interface {
	GetAll(projectID uint) ([]dto.OperationPaginated, error)
	GetPaginated(page, limit int, filter dto.OperationSearchParameters) ([]dto.OperationPaginated, error)
	GetByID(id uint) (model.Operation, error)
	GetByName(name string, projectID uint) (model.Operation, error)
	Create(data dto.Operation) (model.Operation, error)
	Update(data dto.Operation) (model.Operation, error)
	Delete(id uint) error
	Count(filter dto.OperationSearchParameters) (int64, error)
	Import(projectID uint, filepath string) error
	TemplateFile(filepath string, projectID uint) (string, error)
}

func (service *operationService) GetAll(projectID uint) ([]dto.OperationPaginated, error) {
	return service.operationRepo.GetAll(projectID)
}

func (service *operationService) GetPaginated(page, limit int, filter dto.OperationSearchParameters) ([]dto.OperationPaginated, error) {
	return service.operationRepo.GetPaginated(page, limit, filter)
}

func (service *operationService) GetByID(id uint) (model.Operation, error) {
	return service.operationRepo.GetByID(id)
}

func (service *operationService) Create(data dto.Operation) (model.Operation, error) {
	return service.operationRepo.Create(data)
}

func (service *operationService) Update(data dto.Operation) (model.Operation, error) {
	return service.operationRepo.Update(data)
}

func (service *operationService) Delete(id uint) error {
	return service.operationRepo.Delete(id)
}

func (service *operationService) Count(filter dto.OperationSearchParameters) (int64, error) {
	return service.operationRepo.Count(filter)
}

func (service *operationService) GetByName(name string, projectID uint) (model.Operation, error) {
	return service.operationRepo.GetByName(name, projectID)
}

func (service *operationService) Import(projectID uint, filepath string) error {
	f, err := excelize.OpenFile(filepath)
	if err != nil {
		f.Close()
		os.Remove(filepath)
		return fmt.Errorf("Не смог открыть файл: %v", err)
	}

	defer func() {
		f.Close()
		os.Remove(filepath)
	}()

	sheetName := "Услуги"
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return fmt.Errorf("Не смог найти таблицу 'Импорт': %v", err)
	}

	if len(rows) == 1 {
		return fmt.Errorf("Файл не имеет данных")
	}

	operations := []dto.OperationImportDataForInsert{}
	index := 1
	for len(rows) > index {
		operation := dto.OperationImportDataForInsert{
			ProjectID: projectID,
		}

		operation.Code, err = f.GetCellValue(sheetName, "A"+fmt.Sprint(index+1))
		if err != nil {
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке А%d: %v", index+1, err)
		}

		operation.Name, err = f.GetCellValue(sheetName, "B"+fmt.Sprint(index+1))
		if err != nil {
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке B%d: %v", index+1, err)
		}

		costPrimeStr, err := f.GetCellValue(sheetName, "C"+fmt.Sprint(index+1))
		if err != nil {
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке C%d: %v", index+1, err)
		}

		costPrimeFloat64, err := strconv.ParseFloat(costPrimeStr, 64)
		if err != nil {
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке C%d: %v", index+1, err)
		}
		operation.CostPrime = decimal.NewFromFloat(costPrimeFloat64)

		costWithCustomerStr, err := f.GetCellValue(sheetName, "D"+fmt.Sprint(index+1))
		if err != nil {
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке D%d: %v", index+1, err)
		}

		costWithCustomerFloat64, err := strconv.ParseFloat(costWithCustomerStr, 64)
		if err != nil {
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке D%d: %v", index+1, err)
		}
		operation.CostWithCustomer = decimal.NewFromFloat(costWithCustomerFloat64)

		showInReport, err := f.GetCellValue(sheetName, "E"+fmt.Sprint(index+1))
		if err != nil {
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке E%d: %v", index+1, err)
		}

		showInReport = strings.ToLower(showInReport)
		if showInReport == "да" {
			operation.ShowPlannedAmountInReport = true
		} else {
			operation.ShowPlannedAmountInReport = false
		}

		plannedAmountForProject, err := f.GetCellValue(sheetName, "E"+fmt.Sprint(index+1))
		if err != nil {
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке E%d: %v", index+1, err)
		}

		if operation.PlannedAmountForProject, err = strconv.ParseFloat(plannedAmountForProject, 64); err != nil {
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке E%d: %v", index+1, err)
		}

		materialName, err := f.GetCellValue(sheetName, "G"+fmt.Sprint(index+1))
		if err != nil {
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке G%d: %v", index+1, err)
		}

		if materialName != "" {
			material, err := service.materialRepo.GetByName(materialName)
			if err != nil {
				return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке G%d: %v", index+1, err)
			}

			operation.MaterialID = material.ID
		}

		operations = append(operations, operation)
		index++
	}

	if err := f.Close(); err != nil {
		return fmt.Errorf("Не удалось закрыть Excel файл: %v", err)
	}

	if err := os.Remove(filepath); err != nil {
		return fmt.Errorf("Не удалось удалить импортированный файл после сохранения данных: %v", err)
	}

	if err := service.operationRepo.CreateInBatches(operations); err != nil {
		return fmt.Errorf("Не удалось сохранить данные: %v", err)
	}

	return nil
}

func (service *operationService) TemplateFile(filePath string, projectID uint) (string, error) {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		f.Close()
		return "", fmt.Errorf("Не смог открыть шаблонный файл: %v", err)
	}

	materialsSheet := "Материалы"
	allMaterials, err := service.materialRepo.GetAll(projectID)
	if err != nil {
		f.Close()
		return "", fmt.Errorf("Данные материалов недоступны: %v", err)
	}

	for index, materials := range allMaterials {
		f.SetCellStr(materialsSheet, "A"+fmt.Sprint(index+2), materials.Name)
	}

	currentTime := time.Now()
	temporaryFileName := fmt.Sprintf(
		"Шаблон для импорта Услуг - %s.xlsx",
		currentTime.Format("02-01-2006"),
	)
	temporaryFilePath := filepath.Join("./pkg/excels/temp/", temporaryFileName)
	if err := f.SaveAs(temporaryFilePath); err != nil {
		return "", fmt.Errorf("Не удалось обновить шаблон с новыми данными: %v", err)
	}

  if err := f.Close(); err != nil {
    return "", err
  }

	return temporaryFilePath, nil
}
