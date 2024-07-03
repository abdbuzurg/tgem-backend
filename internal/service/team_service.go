package service

import (
	"backend-v2/internal/dto"
	"backend-v2/internal/repository"
	"backend-v2/model"
	"backend-v2/pkg/utils"
	"fmt"
	"os"

	"github.com/xuri/excelize/v2"
)

type teamService struct {
	teamRepo   repository.ITeamRepository
	workerRepo repository.IWorkerRepository
	objectRepo repository.IObjectRepository
}

func InitTeamService(
	teamRepo repository.ITeamRepository,
	workerRepo repository.IWorkerRepository,
	objectRepo repository.IObjectRepository,
) ITeamService {
	return &teamService{
		teamRepo:   teamRepo,
		workerRepo: workerRepo,
		objectRepo: objectRepo,
	}
}

type ITeamService interface {
	GetAll(projectID uint) ([]model.Team, error)
	GetPaginated(page, limit int, projectID uint) ([]dto.TeamPaginated, error)
	GetByID(id uint) (model.Team, error)
	Create(data dto.TeamMutation) (model.Team, error)
	Update(data dto.TeamMutation) (model.Team, error)
	Delete(id uint) error
	Count(projectID uint) (int64, error)
	TemplateFile(projectID uint, filepath string) error
	Import(projectID uint, filepath string) error
  DoesTeamNumberAlreadyExistForCreate(teamNumber string) (bool, error)
  DoesTeamNumberAlreadyExistForUpdate(teamNumber string, id uint) (bool, error)
}

func (service *teamService) GetAll(projectID uint) ([]model.Team, error) {
	return service.teamRepo.GetAll(projectID)
}

func (service *teamService) GetPaginated(page, limit int, projectID uint) ([]dto.TeamPaginated, error) {
	teamPaginatedQueryData, err := service.teamRepo.GetPaginated(page, limit, projectID)
	if err != nil {
		return []dto.TeamPaginated{}, err
	}

	result := []dto.TeamPaginated{}
	latestEntry := dto.TeamPaginated{}
	for index, team := range teamPaginatedQueryData {
		if latestEntry.ID == team.ID {

			if !utils.DoesExist(latestEntry.LeaderNames, team.LeaderName) {
				latestEntry.LeaderNames = append(latestEntry.LeaderNames, team.LeaderName)
			}

		} else {

			if index != 0 {
				result = append(result, latestEntry)
			}

			latestEntry = dto.TeamPaginated{
				ID:           team.ID,
				Number:       team.TeamNumber,
				MobileNumber: team.TeamMobileNumber,
				Company:      team.TeamCompany,
				LeaderNames: []string{
					team.LeaderName,
				},
			}

		}
	}

	if len(teamPaginatedQueryData) > 0 {
		result = append(result, latestEntry)
	}

	return result, nil
}

func (service *teamService) GetByID(id uint) (model.Team, error) {
	return service.teamRepo.GetByID(id)
}

func (service *teamService) Create(data dto.TeamMutation) (model.Team, error) {
	return service.teamRepo.Create(data)
}

func (service *teamService) Update(data dto.TeamMutation) (model.Team, error) {
	return service.teamRepo.Update(data)
}

func (service *teamService) Delete(id uint) error {
	return service.teamRepo.Delete(id)
}

func (service *teamService) Count(projectID uint) (int64, error) {
	return service.teamRepo.Count(projectID)
}

func (service *teamService) TemplateFile(projectID uint, filepath string) error {

	f, err := excelize.OpenFile(filepath)
	if err != nil {
		f.Close()
		return fmt.Errorf("Не смог открыть шаблонный файл: %v", err)
	}

	teamLeaderSheetName := "Бригадиры"
	teamLeaders, err := service.workerRepo.GetByJobTitleInProject("Бригадир")
	if err != nil {
		f.Close()
		return fmt.Errorf("Данные бригадиров недоступны: %v", err)
	}

	for index, teamLeader := range teamLeaders {
		f.SetCellValue(teamLeaderSheetName, "A"+fmt.Sprint(index+2), teamLeader.Name)
	}

	objectSheetName := "Объекты"
	allObjects, err := service.objectRepo.GetAll(projectID)
	if err != nil {
		f.Close()
		return fmt.Errorf("Данные объектов недоступны: %v", err)
	}

	for index, object := range allObjects {
		f.SetCellValue(objectSheetName, "A"+fmt.Sprint(index+2), object.Name)

		objectType := ""
		switch object.Type {
		case "tp_objects":
			objectType = "ТП"
			break
		case "kl04kv_objects":
			objectType = "КЛ 04 КВ"
			break
		case "mjd_objects":
			objectType = "МЖД"
			break
		case "sip_objects":
			objectType = "СИП"
			break
		case "stvt_objects":
			objectType = "СТВТ"
			break
		}
		f.SetCellValue(objectSheetName, "B"+fmt.Sprint(index+2), objectType)
	}

	if err := f.Save(); err != nil {
		return fmt.Errorf("Не удалось обновить шаблон с новыми данными: %v", err)
	}

	f.Close()

	return nil
}

func (service *teamService) Import(projectID uint, filepath string) error {

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

	mutationData := []dto.TeamMutation{}
	index := 1
	for len(rows) > index {
		oneEntry := dto.TeamMutation{
			ProjectID: projectID,
		}

		oneEntry.Number, err = f.GetCellValue(sheetName, "A"+fmt.Sprint(index+1))
		if err != nil {
			f.Close()
			os.Remove(filepath)
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке А%d: %v", index+1, err)
		}

		teamLeader, err := f.GetCellValue(sheetName, "B"+fmt.Sprint(index+1))
		if err != nil {
			f.Close()
			os.Remove(filepath)
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке B%d: %v", index+1, err)
		}
		teamLeaderDataFromDB, err := service.workerRepo.GetByName(teamLeader)
		if err != nil {
			f.Close()
			os.Remove(filepath)
			return fmt.Errorf("Ошибка в файле, заданный бригадир в ячейке B%d отсутствует в базе: %v", index+1, err)
		}
		oneEntry.LeaderWorkerIDs = append(oneEntry.LeaderWorkerIDs, teamLeaderDataFromDB.ID)

		oneEntry.MobileNumber, err = f.GetCellValue(sheetName, "C"+fmt.Sprint(index+1))
		if err != nil {
			f.Close()
			os.Remove(filepath)
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке C%d: %v", index+1, err)
		}

		oneEntry.Company, err = f.GetCellValue(sheetName, "D"+fmt.Sprint(index+1))
		if err != nil {
			f.Close()
			os.Remove(filepath)
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке D%d: %v", index+1, err)
		}

		mutationData = append(mutationData, oneEntry)
		index++
	}

	if err := f.Close(); err != nil {
		return fmt.Errorf("Ошибка при закрытии файла: %v", err)
	}

	if err := os.Remove(filepath); err != nil {
		return fmt.Errorf("Ошибка при удалении временного файла: %v", err)
	}

	_, err = service.teamRepo.CreateInBatches(mutationData)
	if err != nil {
		return err
	}

	return nil
}

func(service *teamService) DoesTeamNumberAlreadyExistForCreate(teamNumber string) (bool, error) {
  return service.teamRepo.DoesTeamNumberAlreadyExistForCreate(teamNumber)
}
func(service *teamService) DoesTeamNumberAlreadyExistForUpdate(teamNumber string, id uint) (bool, error) {
  return service.teamRepo.DoesTeamNumberAlreadyExistForUpdate(teamNumber, id)
}
