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
	mjdObjectRepo          repository.IMJDObjectRepository
	workerRepo             repository.IWorkerRepository
	objectSupervisorsRepo  repository.IObjectSupervisorsRepository
	objectTeamsRepo        repository.IObjectTeamsRepository
	tpNourashesObjects     repository.ITPNourashesObjectsRepository
	teamRepo               repository.ITeamRepository
	tpNourashesObjectsRepo repository.ITPNourashesObjectsRepository
	tpObjectRepo           repository.ITPObjectRepository
	objectRepo             repository.IObjectRepository
}

func InitMJDObjectService(
	mjdObjectRepo repository.IMJDObjectRepository,
	workerRepo repository.IWorkerRepository,
	objectSupervisorsRepo repository.IObjectSupervisorsRepository,
	objectTeamsRepo repository.IObjectTeamsRepository,
	tpNourashesObjects repository.ITPNourashesObjectsRepository,
	teamRepo repository.ITeamRepository,
	tpObjectRepo repository.ITPObjectRepository,
	objectRepo repository.IObjectRepository,
) IMJDObjectService {
	return &mjdObjectService{
		mjdObjectRepo:         mjdObjectRepo,
		workerRepo:            workerRepo,
		objectSupervisorsRepo: objectSupervisorsRepo,
		objectTeamsRepo:       objectTeamsRepo,
		tpNourashesObjects:    tpNourashesObjects,
		teamRepo:              teamRepo,
		tpObjectRepo:          tpObjectRepo,
		objectRepo:            objectRepo,
	}
}

type IMJDObjectService interface {
	GetPaginated(page, limit int, projectID uint) ([]dto.MJDObjectPaginated, error)
	Count(projectID uint) (int64, error)
	Create(data dto.MJDObjectCreate) (model.MJD_Object, error)
	Update(data dto.MJDObjectCreate) (model.MJD_Object, error)
	Delete(id, projectID uint) error
	TemplateFile(filepath string, projectID uint) error
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

		tpNames, err := service.tpNourashesObjects.GetTPObjectNames(oneEntry.ObjectID, "kl04kv_objects")
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
			TPNames:          tpNames,
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

func (service *mjdObjectService) TemplateFile(filepath string, projectID uint) error {
	f, err := excelize.OpenFile(filepath)
	if err != nil {
		f.Close()
		return fmt.Errorf("Не смог открыть шаблонный файл: %v", err)
	}

	sheetName := "Супервайзеры"
	allSupervisors, err := service.workerRepo.GetByJobTitleInProject("Супервайзер", projectID)
	if err != nil {
		f.Close()
		return fmt.Errorf("Данные супервайзеров недоступны: %v", err)
	}

	for index, supervisor := range allSupervisors {
		f.SetCellValue(sheetName, "A"+fmt.Sprint(index+2), supervisor.Name)
	}

	allTeams, err := service.teamRepo.GetAll(projectID)
	if err != nil {
		f.Close()
		return fmt.Errorf("Данны бригад не доступны: %v", err)
	}

	teamSheetName := "Бригады"
	for index, team := range allTeams {
		f.SetCellStr(teamSheetName, "A"+fmt.Sprint(index+2), team.Number)
	}

	allTPObjects, err := service.tpObjectRepo.GetAll(projectID)
	if err != nil {
		f.Close()
		return fmt.Errorf("Данны бригад не доступны: %v", err)
	}

	tpObjectSheetName := "ТП"
	for index, tp := range allTPObjects {
		f.SetCellStr(tpObjectSheetName, "A"+fmt.Sprint(index+2), tp.Name)
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

	sheetName := "МЖД"
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

	mjds := []dto.MJDObjectImportData{}
	index := 1
	for len(rows) > index {
		object := model.Object{
      ProjectID: projectID,
      Type: "mjd_objects",
    }
		mjd := model.MJD_Object{}

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

		mjd.Model, err = f.GetCellValue(sheetName, "C"+fmt.Sprint(index+1))
		if err != nil {
			f.Close()
			os.Remove(filepath)
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке C%d: %v", index+1, err)
		}

		amountEntrancesSTR, err := f.GetCellValue(sheetName, "D"+fmt.Sprint(index+1))
		if err != nil {
			f.Close()
			os.Remove(filepath)
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке D%d: %v", index+1, err)
		}

		amountEntrancesUINT64, err := strconv.ParseUint(amountEntrancesSTR, 10, 64)
		if err != nil {
			f.Close()
			os.Remove(filepath)
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке D%d: %v", index+1, err)
		}
		mjd.AmountEntrances = uint(amountEntrancesUINT64)

		amountStoresSTR, err := f.GetCellValue(sheetName, "E"+fmt.Sprint(index+1))
		if err != nil {
			f.Close()
			os.Remove(filepath)
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке E%d: %v", index+1, err)
		}

		amountStoresUINT64, err := strconv.ParseUint(amountStoresSTR, 10, 64)
		if err != nil {
			f.Close()
			os.Remove(filepath)
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке E%d: %v", index+1, err)
		}
		mjd.AmountStores = uint(amountStoresUINT64)

		hasBasement, err := f.GetCellValue(sheetName, "F"+fmt.Sprint(index+1))
		if err != nil {
			f.Close()
			os.Remove(filepath)
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке F%d: %v", index+1, err)
		}

		if hasBasement == "Да" {
			mjd.HasBasement = true
		} else {
			mjd.HasBasement = false
		}

		supervisorName, err := f.GetCellValue(sheetName, "G"+fmt.Sprint(index+1))
		if err != nil {
			f.Close()
			os.Remove(filepath)
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке G%d: %v", index+1, err)
		}
		supervisorWorker := model.Worker{}
		if supervisorName != "" {
			supervisorWorker, err = service.workerRepo.GetByName(supervisorName)
			if err != nil {
				f.Close()
				os.Remove(filepath)
				return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке G%d: %v", index+1, err)
			}
		}

		teamNumber, err := f.GetCellValue(sheetName, "H"+fmt.Sprint(index+1))
		if err != nil {
			f.Close()
			os.Remove(filepath)
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке H%d: %v", index+1, err)
		}
		team := model.Team{}
		if teamNumber != "" {
			team, err = service.teamRepo.GetByNumber(teamNumber)
			if err != nil {
				f.Close()
				os.Remove(filepath)
				return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке H%d: %v", index+1, err)
			}
		}

		tpName, err := f.GetCellValue(sheetName, "I"+fmt.Sprint(index+1))
		if err != nil {
			f.Close()
			os.Remove(filepath)
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке I%d: %v", index+1, err)
		}
		tpObject := model.Object{}
		if tpName != "" {
			tpObject, err = service.objectRepo.GetByName(tpName)
			if err != nil {
				f.Close()
				os.Remove(filepath)
				return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке I%d: %v", index+1, err)
			}
		}

		mjds = append(mjds, dto.MJDObjectImportData{
			Object: object,
			MJD:    mjd,
			ObjectSupervisors: model.ObjectSupervisors{
				SupervisorWorkerID: supervisorWorker.ID,
			},
			ObjectTeam: model.ObjectTeams{
				TeamID: team.ID,
			},
			NourashedByTP: model.TPNourashesObjects{
				TP_ObjectID: tpObject.ID,
				TargetType:  "mjd_objects",
			},
		})
		index++
	}

	if err := f.Close(); err != nil {
		return fmt.Errorf("Ошибка при закрытии файла: %v", err)
	}

	if err := os.Remove(filepath); err != nil {
		return fmt.Errorf("Ошибка при удалении временного файла: %v", err)
	}

	return service.mjdObjectRepo.Import(mjds)
}
