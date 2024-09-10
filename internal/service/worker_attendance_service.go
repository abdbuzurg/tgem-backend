package service

import (
	"backend-v2/internal/dto"
	"backend-v2/internal/repository"
	"backend-v2/model"
	"fmt"
	"os"
	"time"

	"github.com/xuri/excelize/v2"
)

type workerAttendanceService struct {
	workerAttendanceRepo repository.IWorkerAttendanceRepository
	workerRepo           repository.IWorkerRepository
}

type IWorkerAttendanceService interface {
	Import(projectID uint, filePath string) error
	GetPaginated(projectID uint) ([]dto.WorkerAttendancePaginated, error)
	Count(projectID uint) (int64, error)
}

func InitWorkerAttendanceService(
	workerAttendanceRepo repository.IWorkerAttendanceRepository,
	workerRepo repository.IWorkerRepository,
) IWorkerAttendanceService {
	return &workerAttendanceService{
		workerAttendanceRepo: workerAttendanceRepo,
		workerRepo:           workerRepo,
	}
}

func (service *workerAttendanceService) Import(projectID uint, filepath string) error {

	f, err := excelize.OpenFile(filepath)
	if err != nil {
		f.Close()
		os.Remove(filepath)
		return fmt.Errorf("Не смог открыть файл: %v", err)
	}

	sheetName := "морфо"
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

	type workerAttendanceRaw struct {
		CompanyWorkerID string
		Date            time.Time
	}

	excelData := []workerAttendanceRaw{}
	index := 1
	dateTimeLayout := "01-02-06 3:04:05 PM"
	for len(rows) > index {
		dateExcel, err := f.GetCellValue(sheetName, "A"+fmt.Sprint(index+1))
		if err != nil {
			f.Close()
			os.Remove(filepath)
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке А%d: %v", index+1, err)
		}

		timeExcel, err := f.GetCellValue(sheetName, "B"+fmt.Sprint(index+1))
		if err != nil {
			f.Close()
			os.Remove(filepath)
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке B%d: %v", index+1, err)
		}

		companyWorkerID, err := f.GetCellValue(sheetName, "E"+fmt.Sprint(index+1))
		if err != nil {
			f.Close()
			os.Remove(filepath)
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке B%d: %v", index+1, err)
		}

		dateTime, err := time.Parse(dateTimeLayout, dateExcel+" "+timeExcel)
		if err != nil {
			f.Close()
			os.Remove(filepath)
			return err
		}

		excelData = append(excelData, workerAttendanceRaw{
			CompanyWorkerID: companyWorkerID,
			Date:            dateTime,
		})

		index++
	}
	f.Close()
	os.Remove(filepath)

	workerAttendance := []model.WorkerAttendance{}
	for _, entry := range excelData {
		worker, err := service.workerRepo.GetByCompanyID(entry.CompanyWorkerID)
		if err != nil {
			return err
		}

		isNewWorkerAttendance := true
		for index, attendance := range workerAttendance {
			if attendance.WorkerID == worker.ID {
				if attendance.Start.Day() == entry.Date.Day() {
					workerAttendance[index].End = entry.Date
					isNewWorkerAttendance = false
					break
				}
			}
		}

		if isNewWorkerAttendance {
			workerAttendance = append(workerAttendance, model.WorkerAttendance{
				WorkerID:  worker.ID,
				ProjectID: projectID,
				Start:     entry.Date,
			})
		}
	}

	fmt.Println(workerAttendance)

	return service.workerAttendanceRepo.CreateBatch(workerAttendance)
}

func (service *workerAttendanceService) GetPaginated(projectID uint) ([]dto.WorkerAttendancePaginated, error) {
  data, err :=  service.workerAttendanceRepo.GetPaginated(projectID)
  if err != nil {
    return []dto.WorkerAttendancePaginated{}, err
  }

  location, _ := time.LoadLocation("UTC")
  for index, entry := range data {
    data[index].Start = entry.Start.In(location)
    data[index].End = entry.End.In(location)
  }

  return data, nil
}

func (service *workerAttendanceService) Count(projectID uint) (int64, error) {
	return service.workerAttendanceRepo.Count(projectID)
}
