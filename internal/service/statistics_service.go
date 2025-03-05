package service

import (
	"backend-v2/internal/dto"
	"backend-v2/internal/repository"
	"fmt"
)

type statisticsService struct {
	statRepo   repository.IStatisticsRepository
	workerRepo repository.IWorkerRepository
}

type IStatisticsService interface {
	InvoiceCountStat(projectID uint) ([]dto.PieChartData, error)
	InvoiceInputCreatorStat(projectID uint) ([]dto.PieChartData, error)
	InvoiceOutputCreatorStat(projectID uint) ([]dto.PieChartData, error)
	CountMaterialInInvoices(materialID uint) ([]dto.PieChartData, error)
	LocationMaterial(materialID uint) ([]dto.PieChartData, error)
}

func NewStatisticsService(
	statRepo repository.IStatisticsRepository,
	workerRepo repository.IWorkerRepository,
) IStatisticsService {
	return &statisticsService{
		statRepo:   statRepo,
		workerRepo: workerRepo,
	}
}

func (service *statisticsService) InvoiceCountStat(projectID uint) ([]dto.PieChartData, error) {
	invoiceInputCount, err := service.statRepo.CountInvoiceInputs(projectID)
	if err != nil {
		return []dto.PieChartData{}, err
	}

	invoiceOutputCount, err := service.statRepo.CountInvoiceOutputs(projectID)
	if err != nil {
		return []dto.PieChartData{}, err
	}

	invoiceReturnCount, err := service.statRepo.CountInvoiceReturns(projectID)
	if err != nil {
		return []dto.PieChartData{}, err
	}

	invoiceWriteOffCount, err := service.statRepo.CountInvoiceWriteOffs(projectID)
	if err != nil {
		return []dto.PieChartData{}, err
	}

	result := []dto.PieChartData{
		{ID: 0, Value: float64(invoiceInputCount), Label: "Приход"},
		{ID: 1, Value: float64(invoiceOutputCount), Label: "Отпуск"},
		{ID: 2, Value: float64(invoiceReturnCount), Label: "Возврат"},
		{ID: 3, Value: float64(invoiceWriteOffCount), Label: "Списание"},
	}

	return result, nil
}

func (service *statisticsService) InvoiceInputCreatorStat(projectID uint) ([]dto.PieChartData, error) {
	invoiceInputUnuqueCreators, err := service.statRepo.CountInvoiceInputUniqueCreators(projectID)
	if err != nil {
		return []dto.PieChartData{}, err
	}

	result := []dto.PieChartData{}
	fmt.Println(invoiceInputUnuqueCreators)
	for index, workerID := range invoiceInputUnuqueCreators {
		invoiceInputCountForCreator, err := service.statRepo.CountInvoiceInputCreatorInvoices(projectID, uint(workerID))
		if err != nil {
			return []dto.PieChartData{}, err
		}

		worker, err := service.workerRepo.GetByID(uint(workerID))
		if err != nil {
			return []dto.PieChartData{}, err
		}

		result = append(result, dto.PieChartData{
			ID:    uint(index),
			Value: float64(invoiceInputCountForCreator),
			Label: worker.Name,
		})
	}

	return result, nil
}

func (service *statisticsService) InvoiceOutputCreatorStat(projectID uint) ([]dto.PieChartData, error) {
	invoiceOutputUnuqueCreators, err := service.statRepo.CountInvoiceOutputUniqueCreators(projectID)
	if err != nil {
		return []dto.PieChartData{}, err
	}

	result := []dto.PieChartData{}
	for index, workerID := range invoiceOutputUnuqueCreators {
		invoiceOutputCountForCreator, err := service.statRepo.CountInvoiceOutputCreatorInvoices(projectID, uint(workerID))
		if err != nil {
			return []dto.PieChartData{}, err
		}

		worker, err := service.workerRepo.GetByID(uint(workerID))
		if err != nil {
			return []dto.PieChartData{}, err
		}

		result = append(result, dto.PieChartData{
			ID:    uint(index),
			Value: float64(invoiceOutputCountForCreator),
			Label: worker.Name,
		})
	}

	return result, nil
}

func (service *statisticsService) CountMaterialInInvoices(materialID uint) ([]dto.PieChartData, error) {
	result := []dto.PieChartData{
		{ID: 0, Value: 0, Label: "Приход"},
		{ID: 1, Value: 0, Label: "Отпуск"},
		{ID: 2, Value: 0, Label: "Возврат"},
		{ID: 3, Value: 0, Label: "В процессе корректировки"},
		{ID: 4, Value: 0, Label: "Прошел корректировку"},
		{ID: 5, Value: 0, Label: "Списание"},
		{ID: 6, Value: 0, Label: "Отпуск вне проекта"},
	}

	materialInInvoices, err := service.statRepo.CountMaterialInInvoices(materialID)
	if err != nil {
		return []dto.PieChartData{}, err
	}

	for _, materialInInvoice := range materialInInvoices {
		switch materialInInvoice.InvoiceType {
		case "input":
			result[0].Value += materialInInvoice.Amount
			break
		case "output":
			result[1].Value += materialInInvoice.Amount
			break
		case "return":
			result[2].Value += materialInInvoice.Amount
			break
		case "object":
			result[3].Value += materialInInvoice.Amount
			break
		case "object-correction":
			result[4].Value += materialInInvoice.Amount
			break
		case "writeoff":
			result[5].Value += materialInInvoice.Amount
			break
		case "output-out-of-project":
			result[6].Value += materialInInvoice.Amount
			break

		default:
			fmt.Println("Unknown InvoiceType")
		}
	}

	return result, nil
}

func (service *statisticsService) LocationMaterial(materialID uint) ([]dto.PieChartData, error) {
	result := []dto.PieChartData{
		{ID: 0, Value: 0, Label: "Склад"},
		{ID: 1, Value: 0, Label: "Бригада"},
		{ID: 2, Value: 0, Label: "Объект"},
		{ID: 3, Value: 0, Label: "Списание Склада"},
		{ID: 4, Value: 0, Label: "Потеря Склада"},
		{ID: 5, Value: 0, Label: "Потеря Бригады"},
		{ID: 6, Value: 0, Label: "Списание Объекта"},
		{ID: 7, Value: 0, Label: "Потеря Объекта"},
		{ID: 8, Value: 0, Label: "Вышло из проекта"},
	}

	locationsOfMaterial, err := service.statRepo.CountMaterialInLocations(materialID)
	if err != nil {
		return []dto.PieChartData{}, err
	}

	for _, locationOfMaterial := range locationsOfMaterial {
		switch locationOfMaterial.LocationType {
		case "warehouse":
			result[0].Value += locationOfMaterial.Amount
			break
		case "team":
			result[1].Value += locationOfMaterial.Amount
			break
		case "object":
			result[2].Value += locationOfMaterial.Amount
			break
		case "writeoff-warehouse":
			result[3].Value += locationOfMaterial.Amount
			break
		case "loss-warehouse":
			result[4].Value += locationOfMaterial.Amount
			break
		case "loss-team":
			result[5].Value += locationOfMaterial.Amount
			break
		case "writeoff-object":
			result[6].Value += locationOfMaterial.Amount
			break
		case "loss-object":
			result[7].Value += locationOfMaterial.Amount
			break
		case "out-of-project":
			result[8].Value += locationOfMaterial.Amount
			break
		default:
			fmt.Println("Unknown Storage")
		}
	}

	return result, nil
}
