package service

import (
	"backend-v2/internal/repository"
	"backend-v2/model"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xuri/excelize/v2"
)

type materialService struct {
	materialRepo repository.IMaterialRepository
}

func InitMaterialService(materialRepo repository.IMaterialRepository) IMaterialService {
	return &materialService{
		materialRepo: materialRepo,
	}
}

type IMaterialService interface {
	GetAll(projectID uint) ([]model.Material, error)
	GetPaginated(page, limit int, data model.Material) ([]model.Material, error)
	GetByID(id uint) (model.Material, error)
	Create(data model.Material) (model.Material, error)
	Update(data model.Material) (model.Material, error)
	Delete(id uint) error
	Count(filter model.Material) (int64, error)
	Import(projectID uint, filepath string) error
	Export(projectID uint) (string, error)
}

func (service *materialService) GetAll(projectID uint) ([]model.Material, error) {
	return service.materialRepo.GetAll(projectID)
}

func (service *materialService) GetPaginated(page, limit int, data model.Material) ([]model.Material, error) {
	// if !(utils.IsEmptyFields(data)) {
	return service.materialRepo.GetPaginatedFiltered(page, limit, data)
	// }

	// return service.materialRepo.GetPaginated(page, limit)
}

func (service *materialService) GetByID(id uint) (model.Material, error) {
	return service.materialRepo.GetByID(id)
}

func (service *materialService) Create(data model.Material) (model.Material, error) {
	return service.materialRepo.Create(data)
}

func (service *materialService) Update(data model.Material) (model.Material, error) {
	return service.materialRepo.Update(data)
}

func (service *materialService) Delete(id uint) error {
	return service.materialRepo.Delete(id)
}

func (service *materialService) Count(filter model.Material) (int64, error) {
	return service.materialRepo.Count(filter)
}

func (service *materialService) Import(projectID uint, filepath string) error {
	f, err := excelize.OpenFile(filepath)
	if err != nil {
		f.Close()
		os.Remove(filepath)
		return fmt.Errorf("Не смог открыть файл: %v", err)
	}

	sheetName := "Материалы"
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

	materials := []model.Material{}
	index := 1
	for len(rows) > index {
		material := model.Material{
			ProjectID: projectID,
		}

		material.Name, err = f.GetCellValue(sheetName, "A"+fmt.Sprint(index+1))
		if err != nil {
			f.Close()
			os.Remove(filepath)
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке А%d: %v", index+1, err)
		}

		material.Code, err = f.GetCellValue(sheetName, "B"+fmt.Sprint(index+1))
		if err != nil {
			f.Close()
			os.Remove(filepath)
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке B%d: %v", index+1, err)
		}

		material.Category, err = f.GetCellValue(sheetName, "C"+fmt.Sprint(index+1))
		if err != nil {
			f.Close()
			os.Remove(filepath)
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке C%d: %v", index+1, err)
		}

		material.Unit, err = f.GetCellValue(sheetName, "D"+fmt.Sprint(index+1))
		if err != nil {
			f.Close()
			os.Remove(filepath)
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке D%d: %v", index+1, err)
		}

		material.Article, err = f.GetCellValue(sheetName, "E"+fmt.Sprint(index+1))
		if err != nil {
			f.Close()
			os.Remove(filepath)
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке E%d: %v", index+1, err)
		}

		serialNumberStatus, err := f.GetCellValue(sheetName, "F"+fmt.Sprint(index+1))
		if err != nil {
			f.Close()
			os.Remove(filepath)
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке F%d: %v", index+1, err)
		}

		serialNumberStatus = strings.ToLower(serialNumberStatus)
		if serialNumberStatus == "да" {
			material.HasSerialNumber = true
		} else {
			material.HasSerialNumber = false
		}

		material.Notes, err = f.GetCellValue(sheetName, "G"+fmt.Sprint(index+1))
		if err != nil {
			f.Close()
			os.Remove(filepath)
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке G%d: %v", index+1, err)
		}

		materials = append(materials, material)
		index++
	}

	if err := f.Close(); err != nil {
		return fmt.Errorf("Не удалось закрыть Excel файл: %v", err)
	}

	if err := os.Remove(filepath); err != nil {
		return fmt.Errorf("Не удалось удалить импортированный файл после сохранения данных: %v", err)
	}

	_, err = service.materialRepo.CreateInBatches(materials)
	if err != nil {
		return fmt.Errorf("Не удалось сохранить данные: %v", err)
	}

	return nil
}

func (service *materialService) Export(projectID uint) (string, error) {

	materialTempalteFilePath := filepath.Join("./pkg/excels/templates", "Шаблон для импорта материалов.xlsx")
	f, err := excelize.OpenFile(materialTempalteFilePath)
	if err != nil {
		f.Close()
		return "", fmt.Errorf("Не смог открыть файл: %v", err)
	}
	sheetName := "Материалы"
	startingRow := 2

  materialCount, err := service.materialRepo.Count(model.Material{ProjectID: projectID})
	if err != nil {
		return "", err
	}

	limit := 100
	page := 1
	for materialCount > 0 {
		materials, err := service.materialRepo.GetPaginated(page, limit, projectID)
		if err != nil {
			return "", err
		}

		for index, material := range materials {
			f.SetCellStr(sheetName, "A"+fmt.Sprint(startingRow+index), material.Name)
			f.SetCellStr(sheetName, "B"+fmt.Sprint(startingRow+index), material.Code)
			f.SetCellStr(sheetName, "C"+fmt.Sprint(startingRow+index), material.Category)
			f.SetCellStr(sheetName, "D"+fmt.Sprint(startingRow+index), material.Unit)
			f.SetCellStr(sheetName, "E"+fmt.Sprint(startingRow+index), material.Article)

			serialNumberStatus := "Нет"
			if material.HasSerialNumber {
				serialNumberStatus = "Да"
			}

			f.SetCellStr(sheetName, "F"+fmt.Sprint(startingRow+index), serialNumberStatus)
			f.SetCellStr(sheetName, "G"+fmt.Sprint(startingRow+index), material.Notes)
		}

    startingRow = page * limit + 2 
		page++
		materialCount -= int64(limit)
	}

	exportFileName := "Экспорт Материалов.xlsx"
	exportFilePath := filepath.Join("./pkg/excels/temp/", exportFileName)
	if err := f.SaveAs(exportFilePath); err != nil {
		return "", err
	}

	return exportFileName, nil
}
