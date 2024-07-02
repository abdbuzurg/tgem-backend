package service

import (
	"backend-v2/internal/dto"
	"backend-v2/internal/repository"
	"backend-v2/model"
	"fmt"
	"os"
	"strconv"

	"github.com/xuri/excelize/v2"
)

type mjdObjectService struct {
	mjdObjectRepo         repository.IMJDObjectRepository
	workerRepo            repository.IWorkerRepository
	objectSupervisorsRepo repository.IObjectSupervisorsRepository
	objectTeamsRepo       repository.IObjectTeamsRepository
}

func InitMJDObjectService(
	mjdObjectRepo repository.IMJDObjectRepository,
	workerRepo repository.IWorkerRepository,
	objectSupervisorsRepo repository.IObjectSupervisorsRepository,
	objectTeamsRepo repository.IObjectTeamsRepository,
) IMJDObjectService {
	return &mjdObjectService{
		mjdObjectRepo:         mjdObjectRepo,
		workerRepo:            workerRepo,
		objectSupervisorsRepo: objectSupervisorsRepo,
		objectTeamsRepo:       objectTeamsRepo,
	}
}

type IMJDObjectService interface {
	GetPaginated(page, limit int, projectID uint) ([]dto.MJDObjectPaginated, error)
	Count(projectID uint) (int64, error)
	Create(data dto.MJDObjectCreate) (model.MJD_Object, error)
	Update(data dto.MJDObjectCreate) (model.MJD_Object, error)
	Delete(id, projectID uint) error
	TemplateFile(filepath string) error
	Import(projectID uint, filepath string) error
}

func (service *mjdObjectService) GetPaginated(page, limit int, projectID uint) ([]dto.MJDObjectPaginated, error) {

	data, err := service.mjdObjectRepo.GetPaginated(page, limit, projectID)
	if err != nil {
		return []dto.MJDObjectPaginated{}, err
	}

	result := []dto.MJDObjectPaginated{}
	for _, oneEntry := range data {

		supervisorNames, err := service.objectSupervisorsRepo.GetSupervisorsNameByObjectID(oneEntry.ObjectID)
		if err != nil {
			return []dto.MJDObjectPaginated{}, err
		}

		teamNumbers, err := service.objectTeamsRepo.GetTeamsNumberByObjectID(oneEntry.ObjectID)
		if err != nil {
			return []dto.MJDObjectPaginated{}, err
		}

		result = append(result, dto.MJDObjectPaginated{
			ObjectID:         oneEntry.ObjectID,
			ObjectDetailedID: oneEntry.ObjectDetailedID,
			Name:             oneEntry.Name,
			Status:           oneEntry.Status,
			Model:            oneEntry.Model,
			AmountStores:     oneEntry.AmountStores,
			AmountEntrances:  oneEntry.AmountStores,
			HasBasement:      oneEntry.HasBasement,
			Supervisors:      supervisorNames,
			Teams:            teamNumbers,
		})
	}
	return result, nil
}

func (service *mjdObjectService) Count(projectID uint) (int64, error) {
	return service.mjdObjectRepo.Count(projectID)
}

func (service *mjdObjectService) Create(data dto.MJDObjectCreate) (model.MJD_Object, error) {
	return service.mjdObjectRepo.Create(data)
}

func (service *mjdObjectService) Update(data dto.MJDObjectCreate) (model.MJD_Object, error) {
	return service.mjdObjectRepo.Update(data)
}

func (service *mjdObjectService) Delete(id, projectID uint) error {
	return service.mjdObjectRepo.Delete(id, projectID)
}

func (service *mjdObjectService) TemplateFile(filepath string) error {
	f, err := excelize.OpenFile(filepath)
	if err != nil {
		f.Close()
		return fmt.Errorf("Не смог открыть шаблонный файл: %v", err)
	}

	sheetName := "Супервайзеры"
	allSupervisors, err := service.workerRepo.GetByJobTitleInProject("Супервайзер")
	if err != nil {
		f.Close()
		return fmt.Errorf("Данные супервайзеров недоступны: %v", err)
	}

	for index, supervisor := range allSupervisors {
		f.SetCellValue(sheetName, "A"+fmt.Sprint(index+2), supervisor.Name)
	}

	if err := f.Save(); err != nil {
		return fmt.Errorf("Не удалось обновить шаблон с новыми данными: %v", err)
	}

	f.Close()

	return nil
}

func (service *mjdObjectService) Import(projectID uint, filepath string) error {

	f, err := excelize.OpenFile(filepath)
	if err != nil {
		f.Close()
		os.Remove(filepath)
		return fmt.Errorf("Не смог открыть файл: %v", err)
	}

	sheetName := "Импорт"
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

	objects := []model.Object{}
	mjds := []model.MJD_Object{}
	supervisorIDs := []uint{}
	index := 1
	for len(rows) > index {
		object := model.Object{
			ProjectID: projectID,
			Type:      "mjd_objects",
		}

		mjd := model.MJD_Object{
			HasBasement: false,
		}

		object.Name, err = f.GetCellValue(sheetName, "A"+fmt.Sprint(index+1))
		if err != nil {
			f.Close()
			os.Remove(filepath)
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке А%d: %v", index+1, err)
		}

		object.Status, err = f.GetCellValue(sheetName, "B"+fmt.Sprint(index+1))
		if err != nil {
			f.Close()
			os.Remove(filepath)
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке B%d: %v", index+1, err)
		}

		supervisorName, err := f.GetCellValue(sheetName, "C"+fmt.Sprint(index+1))
		if err != nil {
			f.Close()
			os.Remove(filepath)
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке C%d: %v", index+1, err)
		}

		supervisorWorker, err := service.workerRepo.GetByName(supervisorName)
		if err != nil {
			f.Close()
			os.Remove(filepath)
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке C%d: %v", index+1, err)
		}

		mjd.Model, err = f.GetCellValue(sheetName, "D"+fmt.Sprint(index+1))
		if err != nil {
			f.Close()
			os.Remove(filepath)
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке D%d: %v", index+1, err)
		}

		amountEntrancesSTR, err := f.GetCellValue(sheetName, "E"+fmt.Sprint(index+1))
		if err != nil {
			f.Close()
			os.Remove(filepath)
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке E%d: %v", index+1, err)
		}

		amountEntrancesUINT64, err := strconv.ParseUint(amountEntrancesSTR, 10, 64)
		if err != nil {
			f.Close()
			os.Remove(filepath)
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке E%d: %v", index+1, err)
		}
		mjd.AmountEntrances = uint(amountEntrancesUINT64)

		amountStoresSTR, err := f.GetCellValue(sheetName, "F"+fmt.Sprint(index+1))
		if err != nil {
			f.Close()
			os.Remove(filepath)
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке F%d: %v", index+1, err)
		}

		amountStoresUINT64, err := strconv.ParseUint(amountStoresSTR, 10, 64)
		if err != nil {
			f.Close()
			os.Remove(filepath)
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке F%d: %v", index+1, err)
		}
		mjd.AmountStores = uint(amountStoresUINT64)

		hasBasement, err := f.GetCellValue(sheetName, "G"+fmt.Sprint(index+1))
		if err != nil {
			f.Close()
			os.Remove(filepath)
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке E%d: %v", index+1, err)
		}

		if hasBasement == "Да" {
			mjd.HasBasement = true
		}

		supervisorIDs = append(supervisorIDs, supervisorWorker.ID)
		objects = append(objects, object)
		mjds = append(mjds, mjd)
		index++
	}

	if err := f.Close(); err != nil {
		return fmt.Errorf("Ошибка при закрытии файла: %v", err)
	}

	if err := os.Remove(filepath); err != nil {
		return fmt.Errorf("Ошибка при удалении временного файла: %v", err)
	}

	_, err = service.mjdObjectRepo.CreateInBatches(objects, mjds, supervisorIDs)
	if err != nil {
		return err
	}

	return nil
}
