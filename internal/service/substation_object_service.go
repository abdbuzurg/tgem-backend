package service

import (
	"backend-v2/internal/dto"
	"backend-v2/internal/repository"
	"backend-v2/model"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/xuri/excelize/v2"
)

type substationObjectService struct {
	substationObjectRepo  repository.ISubstationObjectRepository
	workerRepo            repository.IWorkerRepository
	objectSupervisorsRepo repository.IObjectSupervisorsRepository
	objectTeamsRepo       repository.IObjectTeamsRepository
	teamRepo              repository.ITeamRepository
}

func InitSubstationObjectService(
	substationObjectRepo repository.ISubstationObjectRepository,
	workerRepo repository.IWorkerRepository,
	objectSupervisorsRepo repository.IObjectSupervisorsRepository,
	objectTeamsRepo repository.IObjectTeamsRepository,
	teamRepo repository.ITeamRepository,
) ISubstationObjectService {
	return &substationObjectService{
		substationObjectRepo:  substationObjectRepo,
		workerRepo:            workerRepo,
		objectSupervisorsRepo: objectSupervisorsRepo,
		objectTeamsRepo:       objectTeamsRepo,
		teamRepo:              teamRepo,
	}
}

type ISubstationObjectService interface {
	GetPaginated(page, limit int, filter dto.SubstationObjectSearchParameters) ([]dto.SubstationObjectPaginated, error)
	Count(filter dto.SubstationObjectSearchParameters) (int64, error)
	Create(data dto.SubstationObjectCreate) (model.Substation_Object, error)
	Update(data dto.SubstationObjectCreate) (model.Substation_Object, error)
	Delete(id, projectID uint) error
	TemplateFile(filepath string, projectID uint) (string, error)
	Import(projectID uint, filepath string) error
	GetObjectNamesForSearch(projectID uint) ([]dto.DataForSelect[string], error)
	Export(projectID uint) (string, error)
  GetAll(projectID uint) ([]model.Object, error)
}

func (service *substationObjectService) GetAll(projectID uint) ([]model.Object, error) {
  return service.substationObjectRepo.GetAll(projectID)
}

func (service *substationObjectService) GetPaginated(page, limit int, filter dto.SubstationObjectSearchParameters) ([]dto.SubstationObjectPaginated, error) {

	data, err := service.substationObjectRepo.GetPaginated(page, limit, filter)
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

func (service *substationObjectService) Count(filter dto.SubstationObjectSearchParameters) (int64, error) {
	return service.substationObjectRepo.Count(filter)
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

func (service *substationObjectService) TemplateFile(filePath string, projectID uint) (string, error) {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		f.Close()
		return "", fmt.Errorf("Не смог открыть шаблонный файл: %v", err)
	}

	sheetName := "Супервайзеры"
	allSupervisors, err := service.workerRepo.GetByJobTitleInProject("Супервайзер", projectID)
	if err != nil {
		f.Close()
		return "", fmt.Errorf("Данные супервайзеров недоступны: %v", err)
	}

	for index, supervisor := range allSupervisors {
		f.SetCellValue(sheetName, "A"+fmt.Sprint(index+2), supervisor.Name)
	}

	allTeams, err := service.teamRepo.GetAll(projectID)
	if err != nil {
		f.Close()
		return "", fmt.Errorf("Данны бригад не доступны: %v", err)
	}

	teamSheetName := "Бригады"
	for index, team := range allTeams {
		f.SetCellStr(teamSheetName, "A"+fmt.Sprint(index+2), team.Number)
	}

	currentTime := time.Now()
	temporaryFileName := fmt.Sprintf(
		"Шаблон импорта Подстанция - %s.xlsx",
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

func (service *substationObjectService) Import(projectID uint, filepath string) error {

	f, err := excelize.OpenFile(filepath)
	if err != nil {
		f.Close()
		os.Remove(filepath)
		return fmt.Errorf("Не смог открыть файл: %v", err)
	}

	sheetName := "Подстанция"
	rows, err := f.GetRows(sheetName)
	if err != nil {
		f.Close()
		os.Remove(filepath)
		return fmt.Errorf("Не смог найти таблицу 'Подстанция': %v", err)
	}

	if len(rows) == 1 {
		f.Close()
		os.Remove(filepath)
		return fmt.Errorf("Файл не имеет данных")
	}

	substations := []dto.SubstationObjectImportData{}
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

		substation.VoltageClass, err = f.GetCellValue(sheetName, "C"+fmt.Sprint(index+1))
		if err != nil {
			f.Close()
			os.Remove(filepath)
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке D%d: %v", index+1, err)
		}

		numberOfTransformersStr, err := f.GetCellValue(sheetName, "D"+fmt.Sprint(index+1))
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

		supervisorName, err := f.GetCellValue(sheetName, "E"+fmt.Sprint(index+1))
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

		teamNumber, err := f.GetCellValue(sheetName, "F"+fmt.Sprint(index+1))
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

		substations = append(substations, dto.SubstationObjectImportData{
      Object: object,
      Substation: substation,
      ObjectSupervisors: model.ObjectSupervisors{
        SupervisorWorkerID: supervisorWorker.ID,
      },
      ObjectTeam: model.ObjectTeams{
        TeamID: team.ID,
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

	return service.substationObjectRepo.Import(substations)
}

func (service *substationObjectService) GetObjectNamesForSearch(projectID uint) ([]dto.DataForSelect[string], error) {
	return service.substationObjectRepo.GetObjectNamesForSearch(projectID)
}

func (service *substationObjectService) Export(projectID uint) (string, error) {
	substationTempalteFilePath := filepath.Join("./pkg/excels/templates/", "Шаблон для импорта Подстанции.xlsx")
	f, err := excelize.OpenFile(substationTempalteFilePath)
	if err != nil {
		f.Close()
		return "", fmt.Errorf("Не смог открыть файл: %v", err)
	}
	sheetName := "Подстанция"
	startingRow := 2

	substationCount, err := service.substationObjectRepo.Count(dto.SubstationObjectSearchParameters{ProjectID: projectID})
	if err != nil {
		return "", err
	}
	limit := 100
	page := 1

	for substationCount > 0 {
		substations, err := service.substationObjectRepo.GetPaginated(page, limit, dto.SubstationObjectSearchParameters{ProjectID: projectID})
		if err != nil {
			return "", err
		}

		for index, substation := range substations {
			supervisorNames, err := service.objectSupervisorsRepo.GetSupervisorsNameByObjectID(substation.ObjectID)
			if err != nil {
				return "", err
			}

			teamNumbers, err := service.objectTeamsRepo.GetTeamsNumberByObjectID(substation.ObjectID)
			if err != nil {
				return "", err
			}

			f.SetCellStr(sheetName, "A"+fmt.Sprint(startingRow+index), substation.Name)
			f.SetCellStr(sheetName, "B"+fmt.Sprint(startingRow+index), substation.Status)
			f.SetCellStr(sheetName, "C"+fmt.Sprint(startingRow+index), substation.VoltageClass)
			f.SetCellStr(sheetName, "D"+fmt.Sprint(startingRow+index), substation.NumberOfTransformers)

			supervisorsCombined := ""
			for index, supervisor := range supervisorNames {
				if index == 0 {
					supervisorsCombined += supervisor
					continue
				}

				supervisorsCombined += ", " + supervisor
			}
			f.SetCellStr(sheetName, "E"+fmt.Sprint(startingRow+index), supervisorsCombined)

			teamNumbersCombined := ""
			for index, teamNumber := range teamNumbers {
				if index == 0 {
					teamNumbersCombined += teamNumber
					continue
				}

				teamNumbersCombined += ", " + teamNumber
			}
			f.SetCellStr(sheetName, "F"+fmt.Sprint(startingRow+index), teamNumbersCombined)
		}

		startingRow = page*limit + 2
		page++
		substationCount -= int64(limit)
	}

	currentTime := time.Now()
	exportFileName := fmt.Sprintf(
		"Экспорт Подстанция - %s.xlsx",
		currentTime.Format("02-01-2006"),
	)
	exportFilePath := filepath.Join("./pkg/excels/temp/", exportFileName)
	if err := f.SaveAs(exportFilePath); err != nil {
		return "", err
	}

	return exportFileName, nil
}
