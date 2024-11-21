package service

import (
	"backend-v2/internal/dto"
	"backend-v2/internal/repository"
	"backend-v2/model"
	"fmt"
	"os"

	"github.com/xuri/excelize/v2"
)

type workerService struct {
	workerRepo repository.IWorkerRepository
}

func InitWorkerService(workerRepo repository.IWorkerRepository) IWorkerService {
	return &workerService{
		workerRepo: workerRepo,
	}
}

type IWorkerService interface {
	GetAll(projectID uint) ([]model.Worker, error)
	GetPaginated(page, limit int, data model.Worker) ([]model.Worker, error)
	GetByID(id uint) (model.Worker, error)
	GetByJobTitleInProject(jobTitleInProject string, projectID uint) ([]model.Worker, error)
	Create(data model.Worker) (model.Worker, error)
	Update(data model.Worker) (model.Worker, error)
	Delete(id uint) error
	Count() (int64, error)
	Import(filepath string, projectID uint) error
	GetWorkerInformationForSearch(projectID uint) (dto.WorkerInformationForSearch, error)
}

func (service *workerService) GetAll(projectID uint) ([]model.Worker, error) {
	return service.workerRepo.GetAll(projectID)
}

func (service *workerService) GetPaginated(page, limit int, data model.Worker) ([]model.Worker, error) {
	// if !(utils.IsEmptyFields(data)) {
	return service.workerRepo.GetPaginatedFiltered(page, limit, data)
	// }

	// return service.workerRepo.GetPaginated(page, limit)
}

func (service *workerService) GetByID(id uint) (model.Worker, error) {
	return service.workerRepo.GetByID(id)
}

func (service *workerService) GetByJobTitleInProject(jobTitleInProject string, projectID uint) ([]model.Worker, error) {
	return service.workerRepo.GetByJobTitleInProject(jobTitleInProject, projectID)
}

func (service *workerService) Create(data model.Worker) (model.Worker, error) {
	return service.workerRepo.Create(data)
}

func (service *workerService) Update(data model.Worker) (model.Worker, error) {
	return service.workerRepo.Update(data)
}

func (service *workerService) Delete(id uint) error {
	return service.workerRepo.Delete(id)
}

func (service *workerService) Count() (int64, error) {
	return service.workerRepo.Count()
}

func (service *workerService) Import(filepath string, projectID uint) error {

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

	workers := []model.Worker{}
	index := 1
	for len(rows) > index {
		worker := model.Worker{
			ProjectID: projectID,
		}

		worker.Name, err = f.GetCellValue(sheetName, "A"+fmt.Sprint(index+1))
		if err != nil {
			f.Close()
			os.Remove(filepath)
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке А%d: %v", index+1, err)
		}
		worker.JobTitleInProject, err = f.GetCellValue(sheetName, "B"+fmt.Sprint(index+1))
		if err != nil {
			f.Close()
			os.Remove(filepath)
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке B%d: %v", index+1, err)
		}

		worker.MobileNumber, err = f.GetCellValue(sheetName, "C"+fmt.Sprint(index+1))
		if err != nil {
			f.Close()
			os.Remove(filepath)
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке C%d: %v", index+1, err)
		}

		worker.JobTitleInCompany, err = f.GetCellValue(sheetName, "D"+fmt.Sprint(index+1))
		if err != nil {
			f.Close()
			os.Remove(filepath)
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке B%d: %v", index+1, err)
		}

		worker.CompanyWorkerID, err = f.GetCellValue(sheetName, "E"+fmt.Sprint(index+1))
		if err != nil {
			f.Close()
			os.Remove(filepath)
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке C%d: %v", index+1, err)
		}

		workers = append(workers, worker)
		index++
	}

	if err := f.Close(); err != nil {
		return fmt.Errorf("Ошибка при закрытии файла: %v", err)
	}

	if err := os.Remove(filepath); err != nil {
		return fmt.Errorf("Ошибка при удалении временного файла: %v", err)
	}

	_, err = service.workerRepo.CreateInBatches(workers)
	if err != nil {
		return err
	}

	return nil
}

func (service *workerService) GetWorkerInformationForSearch(projectID uint) (dto.WorkerInformationForSearch, error) {
  return service.workerRepo.GetFullWorkerInformationForSearch(projectID)
}
