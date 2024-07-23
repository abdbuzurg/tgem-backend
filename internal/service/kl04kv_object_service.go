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

type kl04kvObjectService struct {
	kl04kvObjectRepo      repository.IKL04KVObjectRepository
	workerRepo            repository.IWorkerRepository
	objectSupervisorsRepo repository.IObjectSupervisorsRepository
	objectTeamsRepo       repository.IObjectTeamsRepository
	tpNourashesObjects    repository.ITPNourashesObjectsRepository
	teamRepo              repository.ITeamRepository
	tpObjectRepo          repository.ITPObjectRepository
	objectRepo            repository.IObjectRepository
}

func InitKL04KVObjectService(
	kl04kvObjectRepo repository.IKL04KVObjectRepository,
	workerRepo repository.IWorkerRepository,
	objectSupervisorsRepo repository.IObjectSupervisorsRepository,
	objectTeamsRepo repository.IObjectTeamsRepository,
	tpNourashesObjects repository.ITPNourashesObjectsRepository,
	teamRepo repository.ITeamRepository,
	tpObjectRepo repository.ITPObjectRepository,
	objectRepo repository.IObjectRepository,
) IKL04KVObjectService {
	return &kl04kvObjectService{
		kl04kvObjectRepo:      kl04kvObjectRepo,
		workerRepo:            workerRepo,
		objectSupervisorsRepo: objectSupervisorsRepo,
		objectTeamsRepo:       objectTeamsRepo,
		tpNourashesObjects:    tpNourashesObjects,
		teamRepo:              teamRepo,
		tpObjectRepo:          tpObjectRepo,
		objectRepo:            objectRepo,
	}
}

type IKL04KVObjectService interface {
	GetPaginated(page, limit int, filter dto.KL04KVObjectSearchParameters) ([]dto.KL04KVObjectPaginated, error)
	Count(filter dto.KL04KVObjectSearchParameters) (int64, error)
	Create(data dto.KL04KVObjectCreate) (model.KL04KV_Object, error)
	Delete(projectID, id uint) error
	Update(data dto.KL04KVObjectCreate) (model.KL04KV_Object, error)
	TemplateFile(filepath string, projectID uint) (string, error)
	Import(projectID uint, filepath string) error
	Export(projectID uint) (string, error)
	GetObjectNamesForSearch(projectID uint) ([]dto.DataForSelect[string], error)
}

func (service *kl04kvObjectService) GetPaginated(page, limit int, filter dto.KL04KVObjectSearchParameters) ([]dto.KL04KVObjectPaginated, error) {

	data, err := service.kl04kvObjectRepo.GetPaginated(page, limit, filter)
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

func (service *kl04kvObjectService) Count(filter dto.KL04KVObjectSearchParameters) (int64, error) {
	return service.kl04kvObjectRepo.Count(filter)
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

func (service *kl04kvObjectService) TemplateFile(filePath string, projectID uint) (string, error) {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		f.Close()
		return "", fmt.Errorf("Не смог открыть шаблонный файл: %v", err)
	}

	supervisorSheetName := "Супервайзеры"
	allSupervisors, err := service.workerRepo.GetByJobTitleInProject("Супервайзер", projectID)
	if err != nil {
		f.Close()
		return "", fmt.Errorf("Данные супервайзеров недоступны: %v", err)
	}

	for index, supervisor := range allSupervisors {
		f.SetCellStr(supervisorSheetName, "A"+fmt.Sprint(index+2), supervisor.Name)
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

	allTPObjects, err := service.tpObjectRepo.GetAll(projectID)
	if err != nil {
		f.Close()
		return "", fmt.Errorf("Данны бригад не доступны: %v", err)
	}

	tpObjectSheetName := "ТП"
	for index, tp := range allTPObjects {
		f.SetCellStr(tpObjectSheetName, "A"+fmt.Sprint(index+2), tp.Name)
	}

  date := time.Now()
  temporaryFilePath := filepath.Join("./pkg/excels/temp/", date.String() + " Шаблон для импорта КЛ 04 КВ.xlsx")
	if err := f.SaveAs(temporaryFilePath); err != nil {
		return "", fmt.Errorf("Не удалось обновить шаблон с новыми данными: %v", err)
	}

	f.Close()

	return temporaryFilePath, nil
}

func (service *kl04kvObjectService) Import(projectID uint, filepath string) error {
	f, err := excelize.OpenFile(filepath)
	if err != nil {
		f.Close()
		os.Remove(filepath)
		return fmt.Errorf("Не смог открыть файл: %v", err)
	}

	sheetName := "КЛ 04 КВ"
	rows, err := f.GetRows(sheetName)
	if err != nil {
		f.Close()
		os.Remove(filepath)
		return fmt.Errorf("Не смог найти таблицу 'КЛ 04 КВ': %v", err)
	}

	if len(rows) == 1 {
		f.Close()
		os.Remove(filepath)
		return fmt.Errorf("Файл не имеет данных")
	}

	kl04kvs := []dto.KL04KVObjectImportData{}
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
		if object.Name == "" {
			f.Close()
			os.Remove(filepath)
			return fmt.Errorf("Ошибка в файле, ячейкa А%d должна иметь данные: %v", index+1, err)
		}

		object.Status, err = f.GetCellValue(sheetName, "B"+fmt.Sprint(index+1))
		if err != nil {
			f.Close()
			os.Remove(filepath)
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке B%d: %v", index+1, err)
		}
		if object.Status == "" {
			f.Close()
			os.Remove(filepath)
			return fmt.Errorf("Ошибка в файле, ячейкa B%d должна иметь данные: %v", index+1, err)
		}

		kl04kv.Nourashes, err = f.GetCellValue(sheetName, "C"+fmt.Sprint(index+1))
		if err != nil {
			f.Close()
			os.Remove(filepath)
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке C%d: %v", index+1, err)
		}
		if kl04kv.Nourashes == "" {
			f.Close()
			os.Remove(filepath)
			return fmt.Errorf("Ошибка в файле, ячейкa C%d должна иметь данные: %v", index+1, err)
		}

		lengthSTR, err := f.GetCellValue(sheetName, "D"+fmt.Sprint(index+1))
		if err != nil {
			f.Close()
			os.Remove(filepath)
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке D%d: %v", index+1, err)
		}
		if lengthSTR == "" {
			f.Close()
			os.Remove(filepath)
			return fmt.Errorf("Ошибка в файле, ячейкa D%d должна иметь данные: %v", index+1, err)
		}

		kl04kv.Length, err = strconv.ParseFloat(lengthSTR, 64)
		if err != nil {
			f.Close()
			os.Remove(filepath)
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке D%d: %v", index+1, err)
		}

		supervisorName, err := f.GetCellValue(sheetName, "E"+fmt.Sprint(index+1))
		if err != nil {
			f.Close()
			os.Remove(filepath)
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке E%d: %v", index+1, err)
		}
		supervisorWorker := model.Worker{}
		if supervisorName != "" {
			supervisorWorker, err = service.workerRepo.GetByName(supervisorName)
			if err != nil {
				f.Close()
				os.Remove(filepath)
				return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке E%d: %v", index+1, err)
			}
		}

		teamNumber, err := f.GetCellValue(sheetName, "F"+fmt.Sprint(index+1))
		if err != nil {
			f.Close()
			os.Remove(filepath)
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке F%d: %v", index+1, err)
		}
		team := model.Team{}
		if teamNumber != "" {
			team, err = service.teamRepo.GetByNumber(teamNumber)
			if err != nil {
				f.Close()
				os.Remove(filepath)
				return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке F%d: %v", index+1, err)
			}
		}

		tpName, err := f.GetCellValue(sheetName, "G"+fmt.Sprint(index+1))
		if err != nil {
			f.Close()
			os.Remove(filepath)
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке G%d: %v", index+1, err)
		}
		tpObject := model.Object{}
		if tpName != "" {
			tpObject, err = service.objectRepo.GetByName(tpName)
			if err != nil {
				f.Close()
				os.Remove(filepath)
				return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке G%d: %v", index+1, err)
			}
		}

		kl04kvs = append(kl04kvs, dto.KL04KVObjectImportData{
			Object: object,
			Kl04KV: kl04kv,
			ObjectSupervisors: model.ObjectSupervisors{
				SupervisorWorkerID: supervisorWorker.ID,
			},
			ObjectTeam: model.ObjectTeams{
				TeamID: team.ID,
			},
			NourashedByTP: model.TPNourashesObjects{
				TP_ObjectID: tpObject.ID,
				TargetType:  "kl04kv_objects",
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

	err = service.kl04kvObjectRepo.Import(kl04kvs)
	if err != nil {
		return err
	}

	return nil
}

func (service *kl04kvObjectService) Export(projectID uint) (string, error) {
	kl04kvTempalteFilePath := filepath.Join("./pkg/excels/templates", "Шаблон для импорта КЛ 04 КВ.xlsx")
	f, err := excelize.OpenFile(kl04kvTempalteFilePath)
	if err != nil {
		f.Close()
		return "", fmt.Errorf("Не смог открыть файл: %v", err)
	}
	sheetName := "Материалы"
	startingRow := 2

	kl04kvCount, err := service.kl04kvObjectRepo.Count(dto.KL04KVObjectSearchParameters{ProjectID: projectID})
	if err != nil {
		return "", err
	}

	limit := 100
	page := 1

	for kl04kvCount > 0 {
		kl04kvs, err := service.kl04kvObjectRepo.GetPaginated(page, limit, dto.KL04KVObjectSearchParameters{ProjectID: projectID})
		if err != nil {
			return "", err
		}

		for index, kl04kv := range kl04kvs {
			supervisorNames, err := service.objectSupervisorsRepo.GetSupervisorsNameByObjectID(kl04kv.ObjectID)
			if err != nil {
				return "", err
			}

			teamNumbers, err := service.objectTeamsRepo.GetTeamsNumberByObjectID(kl04kv.ObjectID)
			if err != nil {
				return "", err
			}

			tpNames, err := service.tpNourashesObjects.GetTPObjectNames(kl04kv.ObjectID, "kl04kv_objects")
			if err != nil {
				return "", err
			}

			f.SetCellStr(sheetName, "A"+fmt.Sprint(startingRow+index), kl04kv.Name)
			f.SetCellStr(sheetName, "B"+fmt.Sprint(startingRow+index), kl04kv.Status)
			f.SetCellStr(sheetName, "C"+fmt.Sprint(startingRow+index), kl04kv.Nourashes)
			f.SetCellFloat(sheetName, "D"+fmt.Sprint(startingRow+index), kl04kv.Length, 2, 64)

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

			tpNamesCombined := ""
			for index, tpName := range tpNames {
				if index == 0 {
					tpNamesCombined += tpName
					continue
				}

				tpNamesCombined += ", " + tpName
			}
			f.SetCellStr(sheetName, "G"+fmt.Sprint(startingRow+index), tpNamesCombined)
		}

		startingRow = page*limit + 2
		page++
		kl04kvCount -= int64(limit)
	}

	exportFileName := "Экспорт КЛ04КВ.xlsx"
	exportFilePath := filepath.Join("./pkg/excels/temp/", exportFileName)
	if err := f.SaveAs(exportFilePath); err != nil {
		return "", err
	}

	return exportFileName, nil
}

func (service *kl04kvObjectService) GetObjectNamesForSearch(projectID uint) ([]dto.DataForSelect[string], error) {
	return service.kl04kvObjectRepo.GetObjectNamesForSearch(projectID)
}
