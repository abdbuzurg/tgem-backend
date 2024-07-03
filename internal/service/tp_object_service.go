package service

import (
	"backend-v2/internal/dto"
	"backend-v2/internal/repository"
	"backend-v2/model"
	"fmt"
	"os"

	"github.com/xuri/excelize/v2"
)

type tpObjectService struct {
	tpObjectRepo          repository.ITPObjectRepository
	workerRepo            repository.IWorkerRepository
	objectSupervisorsRepo repository.IObjectSupervisorsRepository
	objectTeamsRepo       repository.IObjectTeamsRepository
  objectRepo repository.IObjectRepository
}

func InitTPObjectService(
	tpObjectRepo repository.ITPObjectRepository,
	workerRepo repository.IWorkerRepository,
	objectSupervisorsRepo repository.IObjectSupervisorsRepository,
	objectTeamsRepo repository.IObjectTeamsRepository,
  objectRepo repository.IObjectRepository,
) ITPObjectService {
	return &tpObjectService{
		tpObjectRepo:          tpObjectRepo,
		workerRepo:            workerRepo,
		objectSupervisorsRepo: objectSupervisorsRepo,
		objectTeamsRepo:       objectTeamsRepo,
    objectRepo: objectRepo,
	}
}

type ITPObjectService interface {
  GetAllOnlyObjects(projectID uint) ([]model.Object, error)
	GetPaginated(page, limit int, projectID uint) ([]dto.TPObjectPaginated, error)
	Count(projectID uint) (int64, error)
	Create(data dto.TPObjectCreate) (model.TP_Object, error)
	Update(data dto.TPObjectCreate) (model.TP_Object, error)
	Delete(id, projectID uint) error
	TemplateFile(filepath string) error
	Import(projectID uint, filepath string) error
}

func(service *tpObjectService) GetAllOnlyObjects(projectID  uint) ([]model.Object, error) {
  return service.objectRepo.GetAllObjectBasedOnType(projectID, "tp_objects") 
}

func (service *tpObjectService) GetPaginated(page, limit int, projectID uint) ([]dto.TPObjectPaginated, error) {

	data, err := service.tpObjectRepo.GetPaginated(page, limit, projectID)
	if err != nil {
		return []dto.TPObjectPaginated{}, err
	}

	result := []dto.TPObjectPaginated{}
	for _, object := range data {
		supervisorNames, err := service.objectSupervisorsRepo.GetSupervisorsNameByObjectID(object.ObjectID)
		if err != nil {
			return []dto.TPObjectPaginated{}, err
		}

		teamNumbers, err := service.objectTeamsRepo.GetTeamsNumberByObjectID(object.ObjectID)
		if err != nil {
			return []dto.TPObjectPaginated{}, err
		}

		result = append(result, dto.TPObjectPaginated{
			ObjectID:         object.ObjectID,
			ObjectDetailedID: object.ObjectDetailedID,
			Name:             object.Name,
			Status:           object.Status,
			Model:            object.Model,
			VoltageClass:     object.VoltageClass,
			Nourashes:        object.Nourashes,
			Supervisors:      supervisorNames,
			Teams:            teamNumbers,
		})
	}

	return result, nil
}

func (service *tpObjectService) Count(projectID uint) (int64, error) {
	return service.tpObjectRepo.Count(projectID)
}

func (service *tpObjectService) Create(data dto.TPObjectCreate) (model.TP_Object, error) {
	return service.tpObjectRepo.Create(data)
}

func (service *tpObjectService) Update(data dto.TPObjectCreate) (model.TP_Object, error) {
	return service.tpObjectRepo.Update(data)
}

func (service *tpObjectService) Delete(id, projectID uint) error {
	return service.tpObjectRepo.Delete(id, projectID)
}

func (service *tpObjectService) TemplateFile(filepath string) error {

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

func (service *tpObjectService) Import(projectID uint, filepath string) error {

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
	tps := []model.TP_Object{}
	supervisorIDs := []uint{}
	index := 1
	for len(rows) > index {
		object := model.Object{
			ProjectID: projectID,
			Type:      "tp_objects",
		}

		tp := model.TP_Object{}

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

		tp.Model, err = f.GetCellValue(sheetName, "D"+fmt.Sprint(index+1))
		if err != nil {
			f.Close()
			os.Remove(filepath)
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке D%d: %v", index+1, err)
		}

		tp.VoltageClass, err = f.GetCellValue(sheetName, "E"+fmt.Sprint(index+1))
		if err != nil {
			f.Close()
			os.Remove(filepath)
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке E%d: %v", index+1, err)
		}

		tp.Nourashes, err = f.GetCellValue(sheetName, "F"+fmt.Sprint(index+1))
		if err != nil {
			f.Close()
			os.Remove(filepath)
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке F%d: %v", index+1, err)
		}

		supervisorIDs = append(supervisorIDs, supervisorWorker.ID)
		objects = append(objects, object)
		tps = append(tps, tp)
		index++
	}

	if err := f.Close(); err != nil {
		return fmt.Errorf("Ошибка при закрытии файла: %v", err)
	}

	if err := os.Remove(filepath); err != nil {
		return fmt.Errorf("Ошибка при удалении временного файла: %v", err)
	}

	_, err = service.tpObjectRepo.CreateInBatches(objects, tps, supervisorIDs)
	if err != nil {
		return err
	}

	return nil
}
