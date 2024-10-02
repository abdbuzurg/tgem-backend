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

type tpObjectService struct {
	tpObjectRepo          repository.ITPObjectRepository
	workerRepo            repository.IWorkerRepository
	objectSupervisorsRepo repository.IObjectSupervisorsRepository
	objectTeamsRepo       repository.IObjectTeamsRepository
	objectRepo            repository.IObjectRepository
	teamRepo              repository.ITeamRepository
}

func InitTPObjectService(
	tpObjectRepo repository.ITPObjectRepository,
	workerRepo repository.IWorkerRepository,
	objectSupervisorsRepo repository.IObjectSupervisorsRepository,
	objectTeamsRepo repository.IObjectTeamsRepository,
	objectRepo repository.IObjectRepository,
	teamRepo repository.ITeamRepository,
) ITPObjectService {
	return &tpObjectService{
		tpObjectRepo:          tpObjectRepo,
		workerRepo:            workerRepo,
		objectSupervisorsRepo: objectSupervisorsRepo,
		objectTeamsRepo:       objectTeamsRepo,
		objectRepo:            objectRepo,
		teamRepo:              teamRepo,
	}
}

type ITPObjectService interface {
	GetAllOnlyObjects(projectID uint) ([]model.Object, error)
	GetPaginated(page, limit int, filter dto.TPObjectSearchParameters) ([]dto.TPObjectPaginated, error)
	Count(filter dto.TPObjectSearchParameters) (int64, error)
	Create(data dto.TPObjectCreate) (model.TP_Object, error)
	Update(data dto.TPObjectCreate) (model.TP_Object, error)
	Delete(id, projectID uint) error
	TemplateFile(filePath string, projectID uint) (string, error)
	Import(projectID uint, filepath string) error
	GetObjectNamesForSearch(projectID uint) ([]dto.DataForSelect[string], error)
	Export(projectID uint) (string, error)
}

func (service *tpObjectService) GetAllOnlyObjects(projectID uint) ([]model.Object, error) {
	return service.objectRepo.GetAllObjectBasedOnType(projectID, "tp_objects")
}

func (service *tpObjectService) GetPaginated(page, limit int, filter dto.TPObjectSearchParameters) ([]dto.TPObjectPaginated, error) {

	data, err := service.tpObjectRepo.GetPaginated(page, limit, filter)
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
			Supervisors:      supervisorNames,
			Teams:            teamNumbers,
		})
	}

	return result, nil
}

func (service *tpObjectService) Count(filter dto.TPObjectSearchParameters) (int64, error) {
	return service.tpObjectRepo.Count(filter)
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

func (service *tpObjectService) TemplateFile(filePath string, projectID uint) (string, error) {

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
		"Шаблон импорта ТП - %s.xlsx",
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

func (service *tpObjectService) Import(projectID uint, filepath string) error {

	f, err := excelize.OpenFile(filepath)
	if err != nil {
		f.Close()
		os.Remove(filepath)
		return fmt.Errorf("Не смог открыть файл: %v", err)
	}

	sheetName := "ТП"
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

	tps := []dto.TPObjectImportData{}
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

		tp.Model, err = f.GetCellValue(sheetName, "C"+fmt.Sprint(index+1))
		if err != nil {
			f.Close()
			os.Remove(filepath)
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке D%d: %v", index+1, err)
		}

		tp.VoltageClass, err = f.GetCellValue(sheetName, "D"+fmt.Sprint(index+1))
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

		tps = append(tps, dto.TPObjectImportData{
			Object: object,
			TP:     tp,
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

	if err := service.tpObjectRepo.CreateInBatches(tps); err != nil {
		return err
	}

	return nil
}

func (service *tpObjectService) GetObjectNamesForSearch(projectID uint) ([]dto.DataForSelect[string], error) {
	return service.tpObjectRepo.GetObjectNamesForSearch(projectID)
}

func (service *tpObjectService) Export(projectID uint) (string, error) {
	tpTempalteFilePath := filepath.Join("./pkg/excels/templates/", "Шаблон для импорта ТП.xlsx")
	f, err := excelize.OpenFile(tpTempalteFilePath)
	if err != nil {
		f.Close()
		return "", fmt.Errorf("Не смог открыть файл: %v", err)
	}
	sheetName := "ТП"
	startingRow := 2

	tpCount, err := service.tpObjectRepo.Count(dto.TPObjectSearchParameters{ProjectID: projectID})
	if err != nil {
		return "", err
	}
	limit := 100
	page := 1

	for tpCount > 0 {
		tps, err := service.tpObjectRepo.GetPaginated(page, limit, dto.TPObjectSearchParameters{ProjectID: projectID})
		if err != nil {
			return "", err
		}

		for index, tp := range tps {
			supervisorNames, err := service.objectSupervisorsRepo.GetSupervisorsNameByObjectID(tp.ObjectID)
			if err != nil {
				return "", err
			}

			teamNumbers, err := service.objectTeamsRepo.GetTeamsNumberByObjectID(tp.ObjectID)
			if err != nil {
				return "", err
			}

			f.SetCellStr(sheetName, "A"+fmt.Sprint(startingRow+index), tp.Name)
			f.SetCellStr(sheetName, "B"+fmt.Sprint(startingRow+index), tp.Status)
			f.SetCellStr(sheetName, "C"+fmt.Sprint(startingRow+index), tp.Model)
			f.SetCellStr(sheetName, "D"+fmt.Sprint(startingRow+index), tp.VoltageClass)

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
		tpCount -= int64(limit)
	}

	currentTime := time.Now()
	exportFileName := fmt.Sprintf(
		"Экспорт ТП - %s.xlsx",
		currentTime.Format("02-01-2006"),
	)
	exportFilePath := filepath.Join("./pkg/excels/temp/", exportFileName)
	if err := f.SaveAs(exportFilePath); err != nil {
		return "", err
	}

	return exportFileName, nil
}
