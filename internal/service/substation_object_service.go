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

type substationObjectService struct {
	substationObjectRepo  repository.ISubstationObjectRepository
	workerRepo            repository.IWorkerRepository
	objectSupervisorsRepo repository.IObjectSupervisorsRepository
	objectTeamsRepo       repository.IObjectTeamsRepository
}

func InitSubstationObjectService(
	substationObjectRepo repository.ISubstationObjectRepository,
	workerRepo repository.IWorkerRepository,
	objectSupervisorsRepo repository.IObjectSupervisorsRepository,
	objectTeamsRepo repository.IObjectTeamsRepository,
) ISubstationObjectService {
	return &substationObjectService{
		substationObjectRepo:  substationObjectRepo,
		workerRepo:            workerRepo,
		objectSupervisorsRepo: objectSupervisorsRepo,
		objectTeamsRepo:       objectTeamsRepo,
	}
}

type ISubstationObjectService interface{
	GetPaginated(page, limit int, projectID uint) ([]dto.SubstationObjectPaginated, error)
	Count(projectID uint) (int64, error)
	Create(data dto.SubstationObjectCreate) (model.Substation_Object, error)
	Update(data dto.SubstationObjectCreate) (model.Substation_Object, error)
	Delete(id, projectID uint) error
	TemplateFile(filepath string) error
	Import(projectID uint, filepath string) error
}

func (service *substationObjectService) GetPaginated(page, limit int, projectID uint) ([]dto.SubstationObjectPaginated, error) {

	data, err := service.substationObjectRepo.GetPaginated(page, limit, projectID)
	if err != nil {
		return []dto.SubstationObjectPaginated{}, err
	}

	result := []dto.SubstationObjectPaginated{}
	for _, object := range data {
		supervisorNames, err := service.objectSupervisorsRepo.GetSupervisorsNameByObjectID(object.ObjectID)
		if err != nil {
			return []dto.SubstationObjectPaginated{}, err
		}

		teamNumbers, err := service.objectTeamsRepo.GetTeamsNumberByObjectID(object.ObjectID)
		if err != nil {
			return []dto.SubstationObjectPaginated{}, err
		}

		result = append(result, dto.SubstationObjectPaginated{
			ObjectID:             object.ObjectID,
			ObjectDetailedID:     object.ObjectDetailedID,
			Name:                 object.Name,
			Status:               object.Status,
			VoltageClass:         object.VoltageClass,
			NumberOfTransformers: object.NumberOfTransformers,
			Supervisors:          supervisorNames,
			Teams:                teamNumbers,
		})
	}

	return result, nil
}

func (service *substationObjectService) Count(projectID uint) (int64, error) {
	return service.substationObjectRepo.Count(projectID)
}

func (service *substationObjectService) Create(data dto.SubstationObjectCreate) (model.Substation_Object, error) {
	return service.substationObjectRepo.Create(data)
}

func (service *substationObjectService) Update(data dto.SubstationObjectCreate) (model.Substation_Object, error) {
	return service.substationObjectRepo.Update(data)
}

func (service *substationObjectService) Delete(id, projectID uint) error {
	return service.substationObjectRepo.Delete(id, projectID)
}

func (service *substationObjectService) TemplateFile(filepath string) error {

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

func (service *substationObjectService) Import(projectID uint, filepath string) error {

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
	substations := []model.Substation_Object{}
	supervisorIDs := []uint{}
	index := 1
	for len(rows) > index {
		object := model.Object{
			ProjectID: projectID,
			Type:      "substation_objects",
		}

		substation := model.Substation_Object{}

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

		substation.VoltageClass, err = f.GetCellValue(sheetName, "D"+fmt.Sprint(index+1))
		if err != nil {
			f.Close()
			os.Remove(filepath)
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке D%d: %v", index+1, err)
		}

		numberOfTransformersStr, err := f.GetCellValue(sheetName, "E"+fmt.Sprint(index+1))
		if err != nil {
			f.Close()
			os.Remove(filepath)
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке E%d: %v", index+1, err)
		}

		numberOfTransformersUINT64, err := strconv.ParseUint(numberOfTransformersStr, 10, 64)
		if err != nil {
			f.Close()
			os.Remove(filepath)
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке E%d: %v", index+1, err)
		}
		substation.NumberOfTransformers = uint(numberOfTransformersUINT64)

		supervisorIDs = append(supervisorIDs, supervisorWorker.ID)
		objects = append(objects, object)
		substations = append(substations, substation)
		index++
	}

	if err := f.Close(); err != nil {
		return fmt.Errorf("Ошибка при закрытии файла: %v", err)
	}

	if err := os.Remove(filepath); err != nil {
		return fmt.Errorf("Ошибка при удалении временного файла: %v", err)
	}

	_, err = service.substationObjectRepo.CreateInBatches(objects, substations, supervisorIDs)
	if err != nil {
		return err
	}

	return nil
}
