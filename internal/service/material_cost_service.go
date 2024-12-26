package service

import (
	"backend-v2/internal/dto"
	"backend-v2/internal/repository"
	"backend-v2/model"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/shopspring/decimal"
	"github.com/xuri/excelize/v2"
)

type materialCostService struct {
	materialCostRepo repository.IMaterialCostRepository
	materialRepo     repository.IMaterialRepository
}

func InitMaterialCostService(
	materialCostRepo repository.IMaterialCostRepository,
	materialRepo repository.IMaterialRepository,
) IMaterialCostService {
	return &materialCostService{
		materialCostRepo: materialCostRepo,
		materialRepo:     materialRepo,
	}
}

type IMaterialCostService interface {
	GetAll() ([]model.MaterialCost, error)
	GetPaginated(page, limit int, filter dto.MaterialCostSearchFilter) ([]dto.MaterialCostView, error)
	GetByID(id uint) (model.MaterialCost, error)
	Create(data model.MaterialCost) (model.MaterialCost, error)
	Update(data model.MaterialCost) (model.MaterialCost, error)
	Delete(id uint) error
	Count(filter dto.MaterialCostSearchFilter) (int64, error)
	GetByMaterialID(materialID uint) ([]model.MaterialCost, error)
	Import(projectID uint, filePath string) error
	ImportTemplateFile(projectID uint) (string, error)
	Export(projectID uint) (string, error)
}

func (service *materialCostService) GetAll() ([]model.MaterialCost, error) {
	return service.materialCostRepo.GetAll()
}

func (service *materialCostService) GetPaginated(page, limit int, filter dto.MaterialCostSearchFilter) ([]dto.MaterialCostView, error) {
	return service.materialCostRepo.GetPaginatedFiltered(page, limit, filter)
}

func (service *materialCostService) GetByID(id uint) (model.MaterialCost, error) {
	return service.materialCostRepo.GetByID(id)
}

func (service *materialCostService) Create(data model.MaterialCost) (model.MaterialCost, error) {
	return service.materialCostRepo.Create(data)
}

func (service *materialCostService) Update(data model.MaterialCost) (model.MaterialCost, error) {
	return service.materialCostRepo.Update(data)
}

func (service *materialCostService) Delete(id uint) error {
	return service.materialCostRepo.Delete(id)
}

func (service *materialCostService) Count(filter dto.MaterialCostSearchFilter) (int64, error) {
	return service.materialCostRepo.Count(filter)
}

func (service *materialCostService) GetByMaterialID(materialID uint) ([]model.MaterialCost, error) {
	return service.materialCostRepo.GetByMaterialIDSorted(materialID)
}

func (service *materialCostService) ImportTemplateFile(projectID uint) (string, error) {
	materialCostTemplateFilePath := filepath.Join("./pkg/excels/templates/", "Шаблон импорта ценников для материалов.xlsx")
	f, err := excelize.OpenFile(materialCostTemplateFilePath)
	if err != nil {
		f.Close()
		return "", fmt.Errorf("Не смог открыть файл: %v", err)
	}

	sheetName := "Материалы"
	materials, err := service.materialRepo.GetAll(projectID)
	if err != nil {
		f.Close()
		return "", err
	}

	startingRow := 2

	for index, material := range materials {
		f.SetCellStr(sheetName, "A"+fmt.Sprint(startingRow+index), material.Name)
	}

	currentTime := time.Now()
	temporaryFileName := fmt.Sprintf(
		"Шаблон импорта Ценник Материал - %s.xlsx",
		currentTime.Format("02-01-2006"),
	)
	temporaryFilePath := filepath.Join("./pkg/excels/temp/", temporaryFileName)
	if err := f.SaveAs(temporaryFilePath); err != nil {
		return "", fmt.Errorf("Не удалось обновить шаблон с новыми данными: %v", err)
	}

	f.Close()

	return temporaryFilePath, nil
}

func (service *materialCostService) Import(projectID uint, filepath string) error {
	f, err := excelize.OpenFile(filepath)
	if err != nil {
		f.Close()
		os.Remove(filepath)
		return fmt.Errorf("Не смог открыть файл: %v", err)
	}

	sheetName := "Ценники Материалов"
	rows, err := f.GetRows(sheetName)
	if err != nil {
		f.Close()
		os.Remove(filepath)
		return fmt.Errorf("Не смог найти таблицу 'Импорт': %v", err)
	}

	if len(rows) == 1 {
		f.Close()
		os.Remove(filepath)
		return fmt.Errorf("Файл не имеет данных")
	}

	index := 1
	materialCosts := []model.MaterialCost{}
	for len(rows) > index {
		materialCost := model.MaterialCost{}

		materialName, err := f.GetCellValue(sheetName, "A"+fmt.Sprint(index+1))
		if err != nil {
			f.Close()
			os.Remove(filepath)
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке А%d: %v", index+1, err)
		}

		material, err := service.materialRepo.GetByName(projectID, materialName)
		if err != nil {
			f.Close()
			os.Remove(filepath)
			return fmt.Errorf("Ошибка в файле: наименование на строке %v не найдено в базе: %v", index+1, err)
		}

		materialCost.MaterialID = material.ID

		costPrime, err := f.GetCellValue(sheetName, "B"+fmt.Sprint(index+1))
		if err != nil {
			f.Close()
			os.Remove(filepath)
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке B%d: %v", index+1, err)
		}

		if costPrime == "" {
			f.Close()
			os.Remove(filepath)
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке B%d: %v", index+1, err)
		}

		materialCost.CostPrime, err = decimal.NewFromString(costPrime)
		if err != nil {
			f.Close()
			os.Remove(filepath)
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке B%d: %v", index+1, err)
		}

		costM19, err := f.GetCellValue(sheetName, "C"+fmt.Sprint(index+1))
		if err != nil {
			f.Close()
			os.Remove(filepath)
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке C%d: %v", index+1, err)
		}

		if costM19 == "" {
			f.Close()
			os.Remove(filepath)
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке C%d: %v", index+1, err)
		}

		materialCost.CostM19, err = decimal.NewFromString(costM19)
		if err != nil {
			f.Close()
			os.Remove(filepath)
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке C%d: %v", index+1, err)
		}

		costWithCustomer, err := f.GetCellValue(sheetName, "D"+fmt.Sprint(index+1))
		if err != nil {
			f.Close()
			os.Remove(filepath)
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке D%d: %v", index+1, err)
		}

		if costWithCustomer == "" {
			f.Close()
			os.Remove(filepath)
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке D%d: %v", index+1, err)
		}

		materialCost.CostWithCustomer, err = decimal.NewFromString(costWithCustomer)
		if err != nil {
			f.Close()
			os.Remove(filepath)
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке D%d: %v", index+1, err)
		}

		materialCosts = append(materialCosts, materialCost)
		index++

	}

	f.Close()
	os.Remove(filepath)

	_, err = service.materialCostRepo.CreateInBatch(materialCosts)
	if err != nil {
		return fmt.Errorf("Ошибка при сохранение данных: %v", err)
	}

	return nil
}

func (service *materialCostService) Export(projectID uint) (string, error) {

	materialTempalteFilePath := filepath.Join("./pkg/excels/templates", "Шаблон импорта ценников для материалов.xlsx")
	f, err := excelize.OpenFile(materialTempalteFilePath)
	if err != nil {
		f.Close()
		return "", fmt.Errorf("Не смог открыть файл: %v", err)
	}
	sheetName := "Ценники Материалов"
	startingRow := 2

	materialCostCount, err := service.materialCostRepo.Count(dto.MaterialCostSearchFilter{ProjectID: projectID})
	if err != nil {
		return "", err
	}

	limit := 100
	page := 1
	for materialCostCount > 0 {
		materialCosts, err := service.materialCostRepo.GetPaginated(page, limit, projectID)
		if err != nil {
			return "", err
		}

		for index, materialCost := range materialCosts {
			f.SetCellStr(sheetName, "A"+fmt.Sprint(startingRow+index), materialCost.MaterialName)

			costPrime, _ := materialCost.CostPrime.Float64()
			f.SetCellFloat(sheetName, "B"+fmt.Sprint(startingRow+index), costPrime, 2, 64)

			costM19, _ := materialCost.CostM19.Float64()
			f.SetCellFloat(sheetName, "C"+fmt.Sprint(startingRow+index), costM19, 2, 64)

			costWithCustomer, _ := materialCost.CostWithCustomer.Float64()
			f.SetCellFloat(sheetName, "D"+fmt.Sprint(startingRow+index), costWithCustomer, 2, 64)
		}

		startingRow = page*limit + 2
		page++
		materialCostCount -= int64(limit)
	}

	exportFileName := "Экспорт Ценников для Материалов.xlsx"
	exportFilePath := filepath.Join("./pkg/excels/temp/", exportFileName)
	if err := f.SaveAs(exportFilePath); err != nil {
		return "", err
	}

	return exportFileName, nil
}
