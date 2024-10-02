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

type stvtObjectService struct {
	stvtObjectRepo        repository.ISTVTObjectRepository
	workerRepo            repository.IWorkerRepository
	objectSupervisorsRepo repository.IObjectSupervisorsRepository
	objectTeamsRepo       repository.IObjectTeamsRepository
	teamRepo              repository.ITeamRepository
}

func InitSTVTObjectService(
	stvtObjectRepo repository.ISTVTObjectRepository,
	workerRepo repository.IWorkerRepository,
	objectSupervisorsRepo repository.IObjectSupervisorsRepository,
	objectTeamsRepo repository.IObjectTeamsRepository,
	teamRepo repository.ITeamRepository,
) ISTVTObjectService {
	return &stvtObjectService{
		stvtObjectRepo:        stvtObjectRepo,
		workerRepo:            workerRepo,
		objectSupervisorsRepo: objectSupervisorsRepo,
		objectTeamsRepo:       objectTeamsRepo,
		teamRepo:              teamRepo,
	}
}

type ISTVTObjectService interface {
	GetPaginated(page, limit int, filter dto.STVTObjectSearchParameters) ([]dto.STVTObjectPaginated, error)
	Count(filter dto.STVTObjectSearchParameters) (int64, error)
	Create(data dto.STVTObjectCreate) (model.STVT_Object, error)
	Update(data dto.STVTObjectCreate) (model.STVT_Object, error)
	Delete(id, projectID uint) error
	TemplateFile(filePath string, projectID uint) (string, error)
	Import(projectID uint, filePath string) error
	GetObjectNamesForSearch(projectID uint) ([]dto.DataForSelect[string], error)
	Export(projectID uint) (string, error)
}

func (service *stvtObjectService) GetPaginated(page, limit int, filter dto.STVTObjectSearchParameters) ([]dto.STVTObjectPaginated, error) {

	data, err := service.stvtObjectRepo.GetPaginated(page, limit, filter)
	if err != nil {
		return []dto.STVTObjectPaginated{}, err
	}

	result := []dto.STVTObjectPaginated{}
	for _, object := range data {

		supervisorNames, err := service.objectSupervisorsRepo.GetSupervisorsNameByObjectID(object.ObjectID)
		if err != nil {
			return []dto.STVTObjectPaginated{}, err
		}

		teamNumbers, err := service.objectTeamsRepo.GetTeamsNumberByObjectID(object.ObjectID)
		if err != nil {
			return []dto.STVTObjectPaginated{}, err
		}

		result = append(result, dto.STVTObjectPaginated{
			ObjectID:         object.ObjectID,
			ObjectDetailedID: object.ObjectDetailedID,
			Name:             object.Name,
			Status:           object.Status,
			VoltageClass:     object.VoltageClass,
			TTCoefficient:    object.TTCoefficient,
			Supervisors:      supervisorNames,
			Teams:            teamNumbers,
		})
	}

	return result, nil

}

func (service *stvtObjectService) Count(filter dto.STVTObjectSearchParameters) (int64, error) {
	return service.stvtObjectRepo.Count(filter)
}

func (service *stvtObjectService) Create(data dto.STVTObjectCreate) (model.STVT_Object, error) {
	return service.stvtObjectRepo.Create(data)
}

func (service *stvtObjectService) Update(data dto.STVTObjectCreate) (model.STVT_Object, error) {
	return service.stvtObjectRepo.Update(data)
}

func (service *stvtObjectService) Delete(id, projectID uint) error {
	return service.stvtObjectRepo.Delete(id, projectID)
}

func (service *stvtObjectService) TemplateFile(filePath string, projectID uint) (string, error) {
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
		"Шаблон импорта СТВТ - %s.xlsx",
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

func (service *stvtObjectService) Import(projectID uint, filepath string) error {

	f, err := excelize.OpenFile(filepath)
	if err != nil {
		f.Close()
		os.Remove(filepath)
		return fmt.Errorf("Не смог открыть файл: %v", err)
	}

	sheetName := "СТВТ"
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

	stvts := []dto.STVTObjectImportData{}
	index := 1
	for len(rows) > index {
		object := model.Object{
			ProjectID: projectID,
			Type:      "stvt_objects",
		}

		stvt := model.STVT_Object{}

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

		stvt.VoltageClass, err = f.GetCellValue(sheetName, "C"+fmt.Sprint(index+1))
		if err != nil {
			f.Close()
			os.Remove(filepath)
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке D%d: %v", index+1, err)
		}

		stvt.TTCoefficient, err = f.GetCellValue(sheetName, "D"+fmt.Sprint(index+1))
		if err != nil {
			f.Close()
			os.Remove(filepath)
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке E%d: %v", index+1, err)
		}

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

		stvts = append(stvts, dto.STVTObjectImportData{
			Object: object,
			STVT:   stvt,
			ObjectTeam: model.ObjectTeams{
				TeamID: team.ID,
			},
			ObjectSupervisors: model.ObjectSupervisors{
				SupervisorWorkerID: supervisorWorker.ID,
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

	if err := service.stvtObjectRepo.CreateInBatches(stvts); err != nil {
		return err
	}

	return nil
}

func (service *stvtObjectService) GetObjectNamesForSearch(projectID uint) ([]dto.DataForSelect[string], error) {
	return service.stvtObjectRepo.GetObjectNamesForSearch(projectID)
}

func (service *stvtObjectService) Export(projectID uint) (string, error) {
	stvtTempalteFilePath := filepath.Join("./pkg/excels/templates/", "Шаблон для импорта СТВТ.xlsx")
	f, err := excelize.OpenFile(stvtTempalteFilePath)
	if err != nil {
		f.Close()
		return "", fmt.Errorf("Не смог открыть файл: %v", err)
	}
	sheetName := "СТВТ"
	startingRow := 2

	stvtCount, err := service.stvtObjectRepo.Count(dto.STVTObjectSearchParameters{ProjectID: projectID})
	if err != nil {
		return "", err
	}
	limit := 100
	page := 1

	for stvtCount > 0 {
		stvts, err := service.stvtObjectRepo.GetPaginated(page, limit, dto.STVTObjectSearchParameters{ProjectID: projectID})
		if err != nil {
			return "", err
		}

		for index, stvt := range stvts {
			supervisorNames, err := service.objectSupervisorsRepo.GetSupervisorsNameByObjectID(stvt.ObjectID)
			if err != nil {
				return "", err
			}

			teamNumbers, err := service.objectTeamsRepo.GetTeamsNumberByObjectID(stvt.ObjectID)
			if err != nil {
				return "", err
			}

			f.SetCellStr(sheetName, "A"+fmt.Sprint(startingRow+index), stvt.Name)
			f.SetCellStr(sheetName, "B"+fmt.Sprint(startingRow+index), stvt.Status)
			f.SetCellStr(sheetName, "C"+fmt.Sprint(startingRow+index), stvt.VoltageClass)
			f.SetCellStr(sheetName, "D"+fmt.Sprint(startingRow+index), stvt.TTCoefficient)

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
		stvtCount -= int64(limit)
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
