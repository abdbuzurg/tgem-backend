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

type kl04kvObjectService struct {
	kl04kvObjectRepo      repository.IKL04KVObjectRepository
	workerRepo            repository.IWorkerRepository
	objectSupervisorsRepo repository.IObjectSupervisorsRepository
	objectTeamsRepo       repository.IObjectTeamsRepository
	tpNourashesObjects    repository.ITPNourashesObjectsRepository
}

func InitKL04KVObjectService(
	kl04kvObjectRepo repository.IKL04KVObjectRepository,
	workerRepo repository.IWorkerRepository,
	objectSupervisorsRepo repository.IObjectSupervisorsRepository,
	objectTeamsRepo repository.IObjectTeamsRepository,
	tpNourashesObjects repository.ITPNourashesObjectsRepository,
) IKL04KVObjectService {
	return &kl04kvObjectService{
		kl04kvObjectRepo:      kl04kvObjectRepo,
		workerRepo:            workerRepo,
		objectSupervisorsRepo: objectSupervisorsRepo,
		objectTeamsRepo:       objectTeamsRepo,
		tpNourashesObjects:    tpNourashesObjects,
	}
}

type IKL04KVObjectService interface {
	GetPaginated(page, limit int, projectID uint) ([]dto.KL04KVObjectPaginated, error)
	Count(projectID uint) (int64, error)
	Create(data dto.KL04KVObjectCreate) (model.KL04KV_Object, error)
	Delete(projectID, id uint) error
	Update(data dto.KL04KVObjectCreate) (model.KL04KV_Object, error)
	TemplateFile(filepath string) error
	Import(projectID uint, filepath string) error
}

func (service *kl04kvObjectService) GetPaginated(page, limit int, projectID uint) ([]dto.KL04KVObjectPaginated, error) {

	data, err := service.kl04kvObjectRepo.GetPaginated(page, limit, projectID)
	if err != nil {
		return []dto.KL04KVObjectPaginated{}, err
	}

	result := []dto.KL04KVObjectPaginated{}
	for _, oneEntry := range data {
		supervisorNames, err := service.objectSupervisorsRepo.GetSupervisorsNameByObjectID(oneEntry.ObjectID)
		if err != nil {
			return []dto.KL04KVObjectPaginated{}, err
		}

		teamNumbers, err := service.objectTeamsRepo.GetTeamsNumberByObjectID(oneEntry.ObjectID)
		if err != nil {
			return []dto.KL04KVObjectPaginated{}, err
		}

		tpNames, err := service.tpNourashesObjects.GetTPObjectNames(oneEntry.ObjectID, "kl04kv_objects")
		if err != nil {
			return []dto.KL04KVObjectPaginated{}, err
		}

		result = append(result, dto.KL04KVObjectPaginated{
			ObjectID:         oneEntry.ObjectID,
			ObjectDetailedID: oneEntry.ObjectDetailedID,
			Name:             oneEntry.Name,
			Status:           oneEntry.Status,
			Nourashes:        oneEntry.Nourashes,
			Length:           oneEntry.Length,
			Supervisors:      supervisorNames,
			Teams:            teamNumbers,
			TPNames:          tpNames,
		})
	}

	return result, nil
}

func (service *kl04kvObjectService) Count(projectID uint) (int64, error) {
	return service.kl04kvObjectRepo.Count(projectID)
}

func (service *kl04kvObjectService) Create(data dto.KL04KVObjectCreate) (model.KL04KV_Object, error) {
	return service.kl04kvObjectRepo.Create(data)
}

func (service *kl04kvObjectService) Delete(projectID, id uint) error {
	return service.kl04kvObjectRepo.Delete(projectID, id)
}

func (service *kl04kvObjectService) Update(data dto.KL04KVObjectCreate) (model.KL04KV_Object, error) {
	return service.kl04kvObjectRepo.Update(data)
}

func (service *kl04kvObjectService) TemplateFile(filepath string) error {
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

func (service *kl04kvObjectService) Import(projectID uint, filepath string) error {

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
	kl04kvs := []model.KL04KV_Object{}
	supervisorIDs := []uint{}
	index := 1
	for len(rows) > index {
		object := model.Object{
			ProjectID: projectID,
			Type:      "kl04kv_objects",
		}

		kl04kv := model.KL04KV_Object{}

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

		kl04kv.Nourashes, err = f.GetCellValue(sheetName, "D"+fmt.Sprint(index+1))
		if err != nil {
			f.Close()
			os.Remove(filepath)
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке D%d: %v", index+1, err)
		}

		lengthSTR, err := f.GetCellValue(sheetName, "E"+fmt.Sprint(index+1))
		if err != nil {
			f.Close()
			os.Remove(filepath)
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке E%d: %v", index+1, err)
		}

		kl04kv.Length, err = strconv.ParseFloat(lengthSTR, 64)
		if err != nil {
			f.Close()
			os.Remove(filepath)
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке E%d: %v", index+1, err)
		}

		supervisorIDs = append(supervisorIDs, supervisorWorker.ID)
		objects = append(objects, object)
		kl04kvs = append(kl04kvs, kl04kv)
		index++
	}

	if err := f.Close(); err != nil {
		return fmt.Errorf("Ошибка при закрытии файла: %v", err)
	}

	if err := os.Remove(filepath); err != nil {
		return fmt.Errorf("Ошибка при удалении временного файла: %v", err)
	}

	_, err = service.kl04kvObjectRepo.CreateInBatches(objects, kl04kvs, supervisorIDs)
	if err != nil {
		return err
	}

	return nil
}
