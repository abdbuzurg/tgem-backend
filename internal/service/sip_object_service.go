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

type sipObjectService struct {
	sipObjectRepo         repository.ISIPObjectRepository
	workerRepo            repository.IWorkerRepository
	objectSupervisorsRepo repository.IObjectSupervisorsRepository
	objectTeamsRepo       repository.IObjectTeamsRepository
	teamRepo              repository.ITeamRepository
	tpObjectRepo          repository.ITPObjectRepository
}

func InitSIPObjectService(
	sipObjectRepo repository.ISIPObjectRepository,
	workerRepo repository.IWorkerRepository,
	objectSupervisorsRepo repository.IObjectSupervisorsRepository,
	objectTeamsRepo repository.IObjectTeamsRepository,
	teamRepo repository.ITeamRepository,
	tpObjectRepo repository.ITPObjectRepository,
) ISIPObjectService {
	return &sipObjectService{
		sipObjectRepo:         sipObjectRepo,
		workerRepo:            workerRepo,
		objectSupervisorsRepo: objectSupervisorsRepo,
		objectTeamsRepo:       objectTeamsRepo,
		teamRepo:              teamRepo,
		tpObjectRepo:          tpObjectRepo,
	}
}

type ISIPObjectService interface {
	GetPaginated(page, limit int, filter dto.SIPObjectSearchParameters) ([]dto.SIPObjectPaginated, error)
	Count(filter dto.SIPObjectSearchParameters) (int64, error)
	Create(data dto.SIPObjectCreate) (model.SIP_Object, error)
	Update(data dto.SIPObjectCreate) (model.SIP_Object, error)
	Delete(id, projectID uint) error
	TemplateFile(filepath string, projectID uint) (string, error)
	Import(projectID uint, filepath string) error
	GetTPNames(projectID uint) ([]string, error)
	GetObjectNamesForSearch(projectID uint) ([]dto.DataForSelect[string], error)
	Export(projectID uint) (string, error)
}

func (service *sipObjectService) GetPaginated(page, limit int, filter dto.SIPObjectSearchParameters) ([]dto.SIPObjectPaginated, error) {

	data, err := service.sipObjectRepo.GetPaginated(page, limit, filter)
	if err != nil {
		return []dto.SIPObjectPaginated{}, err
	}

	result := []dto.SIPObjectPaginated{}
	for _, object := range data {
		if len(result) != 0 {
			if object.ObjectID == result[len(result)-1].ObjectID {
				continue
			}
		}

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

func (service *sipObjectService) Count(filter dto.SIPObjectSearchParameters) (int64, error) {
	return service.sipObjectRepo.Count(filter)
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

func (service *sipObjectService) TemplateFile(filePath string, projectID uint) (string, error) {
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
	fileName := fmt.Sprintf(
		"Шаблон для Импорта СИП - %s.xlsx",
		currentTime.Format("02-01-2006"),
	)
	tmpFilePath := filepath.Join("./pkg/excels/temp/", fileName)
	if err := f.SaveAs(tmpFilePath); err != nil {
		return "", fmt.Errorf("Не удалось обновить шаблон с новыми данными: %v", err)
	}

	if err := f.Close(); err != nil {
		return "", err
	}

	return tmpFilePath, nil
}

func (service *sipObjectService) Import(projectID uint, filepath string) error {

	f, err := excelize.OpenFile(filepath)
	if err != nil {
		f.Close()
		os.Remove(filepath)
		return fmt.Errorf("Не смог открыть файл: %v", err)
	}

	sheetName := "СИП"
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

	sips := []dto.SIPObjectImportData{}
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

		amountFeedersSTR, err := f.GetCellValue(sheetName, "C"+fmt.Sprint(index+1))
		if err != nil {
			f.Close()
			os.Remove(filepath)
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке C%d: %v", index+1, err)
		}

		amountFeedersUINT64, err := strconv.ParseUint(amountFeedersSTR, 10, 64)
		if err != nil {
			f.Close()
			os.Remove(filepath)
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке C%d: %v", index+1, err)
		}
		sip.AmountFeeders = uint(amountFeedersUINT64)

		supervisorName, err := f.GetCellValue(sheetName, "D"+fmt.Sprint(index+1))
		if err != nil {
			f.Close()
			os.Remove(filepath)
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке D%d: %v", index+1, err)
		}

		supervisorWorker := model.Worker{}
		if supervisorName != "" {
			supervisorWorker, err = service.workerRepo.GetByName(supervisorName)
			if err != nil {
				f.Close()
				os.Remove(filepath)
				return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке D%d: %v", index+1, err)
			}
		}

		teamNumber, err := f.GetCellValue(sheetName, "E"+fmt.Sprint(index+1))
		if err != nil {
			f.Close()
			os.Remove(filepath)
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке E%d: %v", index+1, err)
		}

		team := model.Team{}
		if teamNumber != "" {
			team, err = service.teamRepo.GetByNumber(teamNumber)
			if err != nil {
				f.Close()
				os.Remove(filepath)
				return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке E%d: %v", index+1, err)
			}
		}

		sips = append(sips, dto.SIPObjectImportData{
			Object: object,
			SIP:    sip,
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

	return service.sipObjectRepo.Import(sips)
}

func (service *sipObjectService) GetTPNames(projectID uint) ([]string, error) {
	return service.tpObjectRepo.GetTPNames(projectID)
}

func (service *sipObjectService) GetObjectNamesForSearch(projectID uint) ([]dto.DataForSelect[string], error) {
	return service.sipObjectRepo.GetObjectNamesForSearch(projectID)
}

func (service *sipObjectService) Export(projectID uint) (string, error) {
	sipTempalteFilePath := filepath.Join("./pkg/excels/templates/", "Шаблон для импорта СИП.xlsx")
	f, err := excelize.OpenFile(sipTempalteFilePath)
	if err != nil {
		f.Close()
		return "", fmt.Errorf("Не смог открыть файл: %v", err)
	}
	sheetName := "СИП"
	startingRow := 2

	sipCount, err := service.sipObjectRepo.Count(dto.SIPObjectSearchParameters{ProjectID: projectID})
	if err != nil {
		return "", err
	}
	limit := 100
	page := 1

	for sipCount > 0 {
		sips, err := service.sipObjectRepo.GetPaginated(page, limit, dto.SIPObjectSearchParameters{ProjectID: projectID})
		if err != nil {
			return "", err
		}

		for index, sip := range sips {
			supervisorNames, err := service.objectSupervisorsRepo.GetSupervisorsNameByObjectID(sip.ObjectID)
			if err != nil {
				return "", err
			}

			teamNumbers, err := service.objectTeamsRepo.GetTeamsNumberByObjectID(sip.ObjectID)
			if err != nil {
				return "", err
			}

			f.SetCellStr(sheetName, "A"+fmt.Sprint(startingRow+index), sip.Name)
			f.SetCellStr(sheetName, "B"+fmt.Sprint(startingRow+index), sip.Status)
			f.SetCellUint(sheetName, "C"+fmt.Sprint(startingRow+index), uint64(sip.AmountFeeders))

			supervisorsCombined := ""
			for index, supervisor := range supervisorNames {
				if index == 0 {
					supervisorsCombined += supervisor
					continue
				}
				supervisorsCombined += ", " + supervisor
			}
			f.SetCellStr(sheetName, "D"+fmt.Sprint(startingRow+index), supervisorsCombined)

			teamNumbersCombined := ""
			for index, teamNumber := range teamNumbers {
				if index == 0 {
					teamNumbersCombined += teamNumber
					continue
				}

				teamNumbersCombined += ", " + teamNumber
			}
			f.SetCellStr(sheetName, "E"+fmt.Sprint(startingRow+index), teamNumbersCombined)
		}

		startingRow = page*limit + 2
		page++
		sipCount -= int64(limit)
	}

	

	currentTime := time.Now()
	exportFileName := fmt.Sprintf(
		"Экспорт СИП - %s.xlsx",
		currentTime.Format("02-01-2006"),
	)
	exportFilePath := filepath.Join("./pkg/excels/temp/", exportFileName)
	if err := f.SaveAs(exportFilePath); err != nil {
		return "", err
	}

	if err := f.Close(); err != nil {
		return "", err
	}

	return exportFileName, nil
}
