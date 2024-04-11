package service

import (
	"backend-v2/internal/dto"
	"backend-v2/internal/repository"
	"backend-v2/model"
	"backend-v2/pkg/utils"
	"fmt"
)

type objectService struct {
	objectRepo            repository.IObjectRepository
	supervisorObjectsRepo repository.ISupervisorObjectsRepository
	kl04kvObjectRepo      repository.IKL04KVObjectRepository
	mjdObjectRepo         repository.IMJDObjectRepository
	sipObjectRepo         repository.ISIPObjectRepository
	stvtObjectRepo        repository.ISTVTObjectRepository
	tpObjectRepo          repository.ITPObjectRepository
}

func InitObjectService(
	objectRepo repository.IObjectRepository,
	supervisorObjectsRepo repository.ISupervisorObjectsRepository,
	kl04kvObjectRepo repository.IKL04KVObjectRepository,
	mjdObjectRepo repository.IMJDObjectRepository,
	sipObjectRepo repository.ISIPObjectRepository,
	stvtObjectRepo repository.ISTVTObjectRepository,
	tpObjectRepo repository.ITPObjectRepository,
) IObjectService {
	return &objectService{
		objectRepo:            objectRepo,
		supervisorObjectsRepo: supervisorObjectsRepo,
		kl04kvObjectRepo:      kl04kvObjectRepo,
		mjdObjectRepo:         mjdObjectRepo,
		sipObjectRepo:         sipObjectRepo,
		stvtObjectRepo:        stvtObjectRepo,
		tpObjectRepo:          tpObjectRepo,
	}
}

type IObjectService interface {
	GetAll() ([]model.Object, error)
	GetPaginated(page, limit int, data model.Object) ([]dto.ObjectPaginated, error)
	GetByID(id uint) (model.Object, error)
	Create(data dto.ObjectCreate) (model.Object, error)
	Update(data model.Object) (model.Object, error)
	Delete(id uint) error
	Count() (int64, error)
}

func (service *objectService) GetAll() ([]model.Object, error) {
	return service.objectRepo.GetAll()
}

func (service *objectService) GetPaginated(page, limit int, filter model.Object) ([]dto.ObjectPaginated, error) {

	var data []model.Object
	var err error
	if !(utils.IsEmptyFields(data)) {
		data, err = service.objectRepo.GetPaginatedFiltered(page, limit, filter)
	} else {
		data, err = service.objectRepo.GetPaginated(page, limit)
	}
	if err != nil {
		return []dto.ObjectPaginated{}, err
	}

	result := []dto.ObjectPaginated{}
	for _, object := range data {
		_, err := service.supervisorObjectsRepo.GetByObjectID(object.ID)
		if err != nil {
			return []dto.ObjectPaginated{}, err
		}

		result = append(result, dto.ObjectPaginated{
			ID:     object.ID,
			Type:   object.Type,
			Name:   object.Name,
			Status: object.Status,
		})
	}

	return result, nil

}

func (service *objectService) GetByID(id uint) (model.Object, error) {
	return service.objectRepo.GetByID(id)
}

func (service *objectService) Create(data dto.ObjectCreate) (model.Object, error) {

	object := model.Object{
		ID:               0,
		ObjectDetailedID: 0,
		Type:             data.Type,
		Name:             data.Name,
		Status:           data.Status,
	}

	switch data.Type {
	case "kl04kv_objects":

		objectDetail, err := service.kl04kvObjectRepo.Create(model.KL04KV_Object{
			Length:    data.Length,
			Nourashes: data.Nourashes,
		})
		if err != nil {
			return model.Object{}, err
		}

		object.ObjectDetailedID = objectDetail.ID
		break

	case "mjd_objects":

		objectDetail, err := service.mjdObjectRepo.Create(model.MJD_Object{
			Model:          data.Model,
			AmountStores:   data.AmountStores,
			AmountEntraces: data.AmountEntraces,
			HasBasement:    data.HasBasement,
		})
		if err != nil {
			return model.Object{}, err
		}

		object.ObjectDetailedID = objectDetail.ID
		break

	case "sip_objects":

		objectDetail, err := service.sipObjectRepo.Create(model.SIP_Object{
			AmountFeeders: data.AmountFeeders,
		})
		if err != nil {
			return model.Object{}, err
		}

		object.ObjectDetailedID = objectDetail.ID
		break

	case "stvt_objects":

		objectDetail, err := service.stvtObjectRepo.Create(model.STVT_Object{
			VoltageClass:  data.VoltageClass,
			TTCoefficient: data.TTCoefficient,
		})
		if err != nil {
			return model.Object{}, err
		}

		object.ObjectDetailedID = objectDetail.ID
		break

	case "tp_objects":

		objectDetail, err := service.tpObjectRepo.Create(model.TP_Object{
			Model:        data.Model,
			VoltageClass: data.VoltageClass,
			Nourashes:    data.Nourashes,
		})
		if err != nil {
			return model.Object{}, err
		}

		object.ObjectDetailedID = objectDetail.ID
		break

	default:
		return model.Object{}, fmt.Errorf("Неправильный тип объекта")

	}

	object, err := service.objectRepo.Create(object)
	if err != nil {
		return model.Object{}, err
	}

	var supervisorObjects []model.SupervisorObjects
	for _, supervisorID := range data.Supervisors {
		supervisorObjects = append(supervisorObjects, model.SupervisorObjects{
			ObjectID:           object.ID,
			SupervisorWorkerID: supervisorID,
		})
	}

  _, err = service.supervisorObjectsRepo.CreateBatch(supervisorObjects)
  if err != nil {
    return model.Object{}, err
  }

	return object, nil
}

func (service *objectService) Update(data model.Object) (model.Object, error) {
	return service.objectRepo.Update(data)
}

func (service *objectService) Delete(id uint) error {
	return service.objectRepo.Delete(id)
}

func (service *objectService) Count() (int64, error) {
	return service.objectRepo.Count()
}
