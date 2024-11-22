package service

import (
	"backend-v2/internal/dto"
	"backend-v2/internal/repository"
	"backend-v2/model"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/xuri/excelize/v2"
)

type substationCellObjectService struct {
	substationCellObjectRepo repository.ISubstationCellObjectRepository
	objectSupervisorsRepo    repository.IObjectSupervisorsRepository
	objectTeamsRepo          repository.IObjectTeamsRepository
	workerRepo               repository.IWorkerRepository
	teamRepo                 repository.ITeamRepository
	substationRepo           repository.ISubstationObjectRepository
}

func InitSubstationCellObjectService(
	substationCellObjectRepo repository.ISubstationCellObjectRepository,
	objectSupervisorsRepo repository.IObjectSupervisorsRepository,
	objectTeamsRepo repository.IObjectTeamsRepository,
	workerRepo repository.IWorkerRepository,
	teamRepo repository.ITeamRepository,
	substationRepo repository.ISubstationObjectRepository,
) ISubstationCellObjectService {
	return &substationCellObjectService{
		substationCellObjectRepo: substationCellObjectRepo,
		objectSupervisorsRepo:    objectSupervisorsRepo,
		objectTeamsRepo:          objectTeamsRepo,
		workerRepo:               workerRepo,
		teamRepo:                 teamRepo,
		substationRepo:           substationRepo,
	}
}

type ISubstationCellObjectService interface {
	GetPaginated(int, int, dto.SubstationCellObjectSearchParameters) ([]dto.SubstationCellObjectPaginated, error)
	Count(dto.SubstationCellObjectSearchParameters) (int64, error)
	Create(dto.SubstationCellObjectCreate) (model.SubstationCellObject, error)
	Update(dto.SubstationCellObjectCreate) (model.SubstationCellObject, error)
	Delete(id, projectID uint) error
	TemplateFile(filePath string, projectID uint) (string, error)
	Import(projectID uint, filepath string) error
	Export(projectID uint) (string, error)
	GetObjectNamesForSearch(projectID uint) ([]dto.DataForSelect[string], error)
}

func (service *substationCellObjectService) GetPaginated(page, limit int, filter dto.SubstationCellObjectSearchParameters) ([]dto.SubstationCellObjectPaginated, error) {

	data, err := service.substationCellObjectRepo.GetPaginated(page, limit, filter)
	if err != nil {
		return []dto.SubstationCellObjectPaginated{}, err
	}

	result := []dto.SubstationCellObjectPaginated{}
	for _, object := range data {

		supervisorNames, err := service.objectSupervisorsRepo.GetSupervisorsNameByObjectID(object.ObjectID)
		if err != nil {
			return []dto.SubstationCellObjectPaginated{}, err
		}

		teamNumbers, err := service.objectTeamsRepo.GetTeamsNumberByObjectID(object.ObjectID)
		if err != nil {
			return []dto.SubstationCellObjectPaginated{}, err
		}

		substationName, err := service.substationCellObjectRepo.GetSubstationName(object.ObjectID)
		if err != nil {
			return []dto.SubstationCellObjectPaginated{}, err
		}

		result = append(result, dto.SubstationCellObjectPaginated{
			ObjectID:         object.ObjectID,
			ObjectDetailedID: object.ObjectDetailedID,
			Name:             object.Name,
			Status:           object.Status,
			Supervisors:      supervisorNames,
			Teams:            teamNumbers,
			SubstationName:   substationName,
		})
	}

	return result, nil
}

func (service *substationCellObjectService) Count(filter dto.SubstationCellObjectSearchParameters) (int64, error) {
	return service.substationCellObjectRepo.Count(filter)
}

func (service *substationCellObjectService) Create(data dto.SubstationCellObjectCreate) (model.SubstationCellObject, error) {
	return service.substationCellObjectRepo.Create(data)
}

func (service *substationCellObjectService) Update(data dto.SubstationCellObjectCreate) (model.SubstationCellObject, error) {
	return service.substationCellObjectRepo.Update(data)
}

func (service *substationCellObjectService) Delete(id, projectID uint) error {
	return service.substationCellObjectRepo.Delete(id, projectID)
}

func (service *substationCellObjectService) TemplateFile(filePath string, projectID uint) (string, error) {
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

	substationSheetName := "Подстанции"
	allSubstationObjects, err := service.substationRepo.GetAllNames(projectID)
	if err != nil {
		f.Close()
		return "", fmt.Errorf("Данны подстанций не доступны: %v", err)
	}

	for index, name := range allSubstationObjects {
		f.SetCellStr(substationSheetName, "A"+fmt.Sprint(index+2), name)
	}

	currentTime := time.Now()
	temporaryFileName := fmt.Sprintf(
		"Шаблон для импорт Ячеек Подстанции - %s.xlsx",
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

func (service *substationCellObjectService) Import(projectID uint, filepath string) error {
	f, err := excelize.OpenFile(filepath)
	if err != nil {
		f.Close()
		os.Remove(filepath)
		return fmt.Errorf("Не смог открыть файл: %v", err)
	}

	sheetName := "Ячейка Подстанции"
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

	substationCells := []dto.SubstationCellObjectImportData{}
	index := 1
	for len(rows) > index {
		object := model.Object{
			ProjectID: projectID,
			Type:      "substation_cell_objects",
		}

		substationCell := model.SubstationCellObject{}

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

		supervisorWorker := model.Worker{}
		if supervisorName != "" {
			supervisorWorker, err = service.workerRepo.GetByName(supervisorName)
			if err != nil {
				f.Close()
				os.Remove(filepath)
				return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке C%d: %v", index+1, err)
			}
		}

		teamNumber, err := f.GetCellValue(sheetName, "D"+fmt.Sprint(index+1))
		if err != nil {
			f.Close()
			os.Remove(filepath)
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке D%d: %v", index+1, err)
		}
		team := model.Team{}
		if teamNumber != "" {
			team, err = service.teamRepo.GetByNumber(teamNumber)
			if err != nil {
				f.Close()
				os.Remove(filepath)
				return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке D%d: %v", index+1, err)
			}
		}

		substationName, err := f.GetCellValue(sheetName, "E"+fmt.Sprint(index+1))
		if err != nil {
			f.Close()
			os.Remove(filepath)
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке E%d: %v", index+1, err)
		}
		substation := model.Object{}
		if substationName != "" {
			substation, err = service.substationRepo.GetByName(substationName)
			if err != nil {
				f.Close()
				os.Remove(filepath)
				return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке D%d: %v", index+1, err)
			}
		}

		substationCells = append(substationCells, dto.SubstationCellObjectImportData{
			Object:         object,
			SubstationCell: substationCell,
			ObjectTeam: model.ObjectTeams{
				TeamID: team.ID,
			},
			ObjectSupervisors: model.ObjectSupervisors{
				SupervisorWorkerID: supervisorWorker.ID,
			},
			Nourashes: model.SubstationCellNourashesSubstationObject{
				SubstationObjectID: substation.ID,
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

	if err := service.substationCellObjectRepo.CreateInBatches(substationCells); err != nil {
		return err
	}

	return nil
}

func (service *substationCellObjectService) Export(projectID uint) (string, error) {
	substationCellTempalteFilePath := filepath.Join("./pkg/excels/templates/", "Шаблон для импорт Ячеек Подстанции.xlsx")
	f, err := excelize.OpenFile(substationCellTempalteFilePath)
	if err != nil {
		f.Close()
		return "", fmt.Errorf("Не смог открыть файл: %v", err)
	}
	sheetName := "Ячейка Подстанции"
	startingRow := 2

	substationCellCount, err := service.substationCellObjectRepo.Count(dto.SubstationCellObjectSearchParameters{ProjectID: projectID})
	if err != nil {
		return "", err
	}
	limit := 100
	page := 1

	for substationCellCount > 0 {
		substationCells, err := service.substationCellObjectRepo.GetPaginated(page, limit, dto.SubstationCellObjectSearchParameters{ProjectID: projectID})
		if err != nil {
			return "", err
		}

		for index, substationCell := range substationCells {
			supervisorNames, err := service.objectSupervisorsRepo.GetSupervisorsNameByObjectID(substationCell.ObjectID)
			if err != nil {
				return "", err
			}

			teamNumbers, err := service.objectTeamsRepo.GetTeamsNumberByObjectID(substationCell.ObjectID)
			if err != nil {
				return "", err
			}

			f.SetCellStr(sheetName, "A"+fmt.Sprint(startingRow+index), substationCell.Name)
			f.SetCellStr(sheetName, "B"+fmt.Sprint(startingRow+index), substationCell.Status)

			supervisorsCombined := ""
			for index, supervisor := range supervisorNames {
				if index == 0 {
					supervisorsCombined += supervisor
					continue
				}

				supervisorsCombined += ", " + supervisor
			}
			f.SetCellStr(sheetName, "C"+fmt.Sprint(startingRow+index), supervisorsCombined)

			teamNumbersCombined := ""
			for index, teamNumber := range teamNumbers {
				if index == 0 {
					teamNumbersCombined += teamNumber
					continue
				}

				teamNumbersCombined += ", " + teamNumber
			}
			f.SetCellStr(sheetName, "D"+fmt.Sprint(startingRow+index), teamNumbersCombined)

			substationName, err := service.substationCellObjectRepo.GetSubstationName(substationCell.ObjectID)
			if err != nil {
				return "", err
			}
			f.SetCellValue(sheetName, "E"+fmt.Sprint(startingRow+index), substationName)
		}

		startingRow = page*limit + 2
		page++
		substationCellCount -= int64(limit)
	}

	currentTime := time.Now()
	exportFileName := fmt.Sprintf(
		"Экспорт СТВТ - %s.xlsx",
		currentTime.Format("02-01-2006"),
	)
	exportFilePath := filepath.Join("./pkg/excels/temp/", exportFileName)
	if err := f.SaveAs(exportFilePath); err != nil {
		return "", err
	}
	return exportFileName, nil
}

func (service *substationCellObjectService) GetObjectNamesForSearch(projectID uint) ([]dto.DataForSelect[string], error) {
	return service.substationCellObjectRepo.GetObjectNamesForSearch(projectID)
}
