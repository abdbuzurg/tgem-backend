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

type sipObjectService struct {
	sipObjectRepo         repository.ISIPObjectRepository
	workerRepo            repository.IWorkerRepository
	objectSupervisorsRepo repository.IObjectSupervisorsRepository
	objectTeamsRepo       repository.IObjectTeamsRepository
}

func InitSIPObjectService(
	sipObjectRepo repository.ISIPObjectRepository,
	workerRepo repository.IWorkerRepository,
	objectSupervisorsRepo repository.IObjectSupervisorsRepository,
	objectTeamsRepo repository.IObjectTeamsRepository,
) ISIPObjectService {
	return &sipObjectService{
		sipObjectRepo:         sipObjectRepo,
		workerRepo:            workerRepo,
		objectSupervisorsRepo: objectSupervisorsRepo,
		objectTeamsRepo:       objectTeamsRepo,
	}
}

type ISIPObjectService interface {
	GetPaginated(page, limit int, projectID uint) ([]dto.SIPObjectPaginated, error)
	Count(projectID uint) (int64, error)
	Create(data dto.SIPObjectCreate) (model.SIP_Object, error)
	Update(data dto.SIPObjectCreate) (model.SIP_Object, error)
	Delete(id, projectID uint) error
	TemplateFile(filepath string) error
	Import(projectID uint, filepath string) error
}

func (service *sipObjectService) GetPaginated(page, limit int, projectID uint) ([]dto.SIPObjectPaginated, error) {

	data, err := service.sipObjectRepo.GetPaginated(page, limit, projectID)
	if err != nil {
		return []dto.SIPObjectPaginated{}, err
	}

	result := []dto.SIPObjectPaginated{}
	for _, object := range data {

		supervisorNames, err := service.objectSupervisorsRepo.GetSupervisorsNameByObjectID(object.ObjectID)
		if err != nil {
			return []dto.SIPObjectPaginated{}, err
		}

		teamNumbers, err := service.objectTeamsRepo.GetTeamsNumberByObjectID(object.ObjectID)
		if err != nil {
			return []dto.SIPObjectPaginated{}, err
		}

		result = append(result, dto.SIPObjectPaginated{
			ObjectID:         object.ObjectID,
			ObjectDetailedID: object.ObjectDetailedID,
			Name:             object.Name,
			Status:           object.Status,
			AmountFeeders:    object.AmountFeeders,
			Supervisors:      supervisorNames,
			Teams:            teamNumbers,
		})
	}

	return result, nil
}

func (service *sipObjectService) Count(projectID uint) (int64, error) {
	return service.sipObjectRepo.Count(projectID)
}

func (service *sipObjectService) Create(data dto.SIPObjectCreate) (model.SIP_Object, error) {
	return service.sipObjectRepo.Create(data)
}

func (service *sipObjectService) Update(data dto.SIPObjectCreate) (model.SIP_Object, error) {
	return service.sipObjectRepo.Update(data)
}

func (service *sipObjectService) Delete(id, projectID uint) error {
	return service.sipObjectRepo.Delete(id, projectID)
}

func (service *sipObjectService) TemplateFile(filepath string) error {
	f, err := excelize.OpenFile(filepath)
	if err != nil {
		f.Close()
		return fmt.Errorf("Не смог открыть шаблонный файл: %v", err)
	}

	sheetName := "Супервайзеры"
	allSupervisors, err := service.workerRepo.GetByJobTitle("Супервайзер")
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

func (service *sipObjectService) Import(projectID uint, filepath string) error {

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
	sips := []model.SIP_Object{}
	supervisorIDs := []uint{}
	index := 1
	for len(rows) > index {
		object := model.Object{
			ProjectID: projectID,
			Type:      "sip_objects",
		}

		sip := model.SIP_Object{}

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

		amountFeedersSTR, err := f.GetCellValue(sheetName, "D"+fmt.Sprint(index+1))
		if err != nil {
			f.Close()
			os.Remove(filepath)
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке D%d: %v", index+1, err)
		}

		amountFeedersUINT64, err := strconv.ParseUint(amountFeedersSTR, 10, 64)
		if err != nil {
			f.Close()
			os.Remove(filepath)
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке D%d: %v", index+1, err)
		}
		sip.AmountFeeders = uint(amountFeedersUINT64)

		supervisorIDs = append(supervisorIDs, supervisorWorker.ID)
		objects = append(objects, object)
		sips = append(sips, sip)
		index++
	}

	if err := f.Close(); err != nil {
		return fmt.Errorf("Ошибка при закрытии файла: %v", err)
	}

	if err := os.Remove(filepath); err != nil {
		return fmt.Errorf("Ошибка при удалении временного файла: %v", err)
	}

	_, err = service.sipObjectRepo.CreateInBatches(objects, sips, supervisorIDs)
	if err != nil {
		return err
	}

	return nil
}
