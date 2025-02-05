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
}

func NewStatisticsService(
	statRepo repository.IStatisticsRepository,
	workerRepo repository.IWorkerRepository,
) IStatisticsService {
	return &statisticsService{
		statRepo: statRepo,
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
		{ID: 0, Value: invoiceInputCount, Label: "Приход"},
		{ID: 1, Value: invoiceOutputCount, Label: "Отпуск"},
		{ID: 2, Value: invoiceReturnCount, Label: "Возврат"},
		{ID: 3, Value: invoiceWriteOffCount, Label: "Списание"},
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
			Value: invoiceInputCountForCreator,
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
			Value: invoiceOutputCountForCreator,
			Label: worker.Name,
		})
	}

	return result, nil
}
