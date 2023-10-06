package service

import (
	"backend-v2/internal/repository"
	"backend-v2/model"
	"backend-v2/pkg/utils"
)

type kl04kvObjectService struct {
	kl04kvObjectRepo repository.IKL04KVObjectRepository
}

func InitKL04KVObjectService(kl04kvObjectRepo repository.IKL04KVObjectRepository) IKL04KVObjectService {
	return &kl04kvObjectService{
		kl04kvObjectRepo: kl04kvObjectRepo,
	}
}

type IKL04KVObjectService interface {
	GetAll() ([]model.KL04KV_Object, error)
	GetPaginated(page, limit int, data model.KL04KV_Object) ([]model.KL04KV_Object, error)
	GetByID(id uint) (model.KL04KV_Object, error)
	Create(data model.KL04KV_Object) (model.KL04KV_Object, error)
	Update(data model.KL04KV_Object) (model.KL04KV_Object, error)
	Delete(id uint) error
	Count() (int64, error)
}

func (service *kl04kvObjectService) GetAll() ([]model.KL04KV_Object, error) {
	return service.kl04kvObjectRepo.GetAll()
}

func (service *kl04kvObjectService) GetPaginated(page, limit int, data model.KL04KV_Object) ([]model.KL04KV_Object, error) {
	if !(utils.IsEmptyFields(data)) {
		return service.kl04kvObjectRepo.GetPaginatedFiltered(page, limit, data)
	}

	return service.kl04kvObjectRepo.GetPaginated(page, limit)
}

func (service *kl04kvObjectService) GetByID(id uint) (model.KL04KV_Object, error) {
	return service.kl04kvObjectRepo.GetByID(id)
}

func (service *kl04kvObjectService) Create(data model.KL04KV_Object) (model.KL04KV_Object, error) {
	return service.kl04kvObjectRepo.Create(data)
}

func (service *kl04kvObjectService) Update(data model.KL04KV_Object) (model.KL04KV_Object, error) {
	return service.kl04kvObjectRepo.Update(data)
}

func (service *kl04kvObjectService) Delete(id uint) error {
	return service.kl04kvObjectRepo.Delete(id)
}

func (service *kl04kvObjectService) Count() (int64, error) {
	return service.kl04kvObjectRepo.Count()
}
