package service

import (
	"backend-v2/internal/repository"
	"backend-v2/model"
)

type serialNumberService struct {
	serialNumberRepo repository.ISerialNumberRepository
	materialCostRepo repository.IMaterialCostRepository
}

func InitSerialNumberService(
	serialNumberRepo repository.ISerialNumberRepository,
	materialCostRepo repository.IMaterialCostRepository,
) ISerialNumberService {
	return &serialNumberService{
		serialNumberRepo: serialNumberRepo,
		materialCostRepo: materialCostRepo,
	}
}

type ISerialNumberService interface {
	GetAll() ([]model.SerialNumber, error)
	GetCodesByMaterialID(materialID uint) ([]string, error)
	Create(data model.SerialNumber) (model.SerialNumber, error)
	Update(data model.SerialNumber) (model.SerialNumber, error)
	Delete(id uint) error
}

func (service *serialNumberService) GetAll() ([]model.SerialNumber, error) {
	return service.serialNumberRepo.GetAll()
}

func (service *serialNumberService) GetCodesByMaterialID(materialID uint) ([]string, error) {
	var codes []string

	materialCosts, err := service.materialCostRepo.GetByMaterialID(materialID)
	if err != nil {
		return codes, err
	}

	for _, materialCost := range materialCosts {
		serialNumbers, err := service.serialNumberRepo.GetByMaterialCostID(materialCost.ID)
		if err != nil {
			return codes, err
		}

		for _, serialNumber := range serialNumbers {
			codes = append(codes, serialNumber.Code)
		}
	}
	return codes, nil
}

func (service *serialNumberService) Create(data model.SerialNumber) (model.SerialNumber, error) {
	return service.serialNumberRepo.Create(data)
}

func (service *serialNumberService) Update(data model.SerialNumber) (model.SerialNumber, error) {
	return service.serialNumberRepo.Update(data)
}

func (service *serialNumberService) Delete(id uint) error {
	return service.serialNumberRepo.Delete(id)
}
