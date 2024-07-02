package repository

import (
	"backend-v2/internal/dto"
	"backend-v2/model"

	"gorm.io/gorm"
)

type kl04kvObjectRepository struct {
	db *gorm.DB
}

func InitKL04KVObjectRepository(db *gorm.DB) IKL04KVObjectRepository {
	return &kl04kvObjectRepository{
		db: db,
	}
}

type IKL04KVObjectRepository interface {
	GetAll() ([]model.KL04KV_Object, error)
	GetPaginated(page, limit int, projectID uint) ([]dto.KL04KVObjectPaginatedQuery, error)
	GetByID(id uint) (model.KL04KV_Object, error)
	Create(data dto.KL04KVObjectCreate) (model.KL04KV_Object, error)
	CreateInBatches(objects []model.Object, kl04kvs []model.KL04KV_Object, supervisors []uint) ([]model.KL04KV_Object, error)
	Delete(projectID, id uint) error
	Count(projectID uint) (int64, error)
	Update(data dto.KL04KVObjectCreate) (model.KL04KV_Object, error)
}

func (repo *kl04kvObjectRepository) GetAll() ([]model.KL04KV_Object, error) {
	data := []model.KL04KV_Object{}
	err := repo.db.Order("id desc").Find(&data).Error
	return data, err
}

func (repo *kl04kvObjectRepository) GetPaginated(page, limit int, projectID uint) ([]dto.KL04KVObjectPaginatedQuery, error) {
	data := []dto.KL04KVObjectPaginatedQuery{}
	err := repo.db.Raw(`
    SELECT 
      objects.id as object_id,
      kl04_kv_objects.id as object_detailed_id,
      objects.name as name,
      objects.status as status,
      kl04_kv_objects.length as length,
      kl04_kv_objects.nourashes as nourashes
    FROM objects
      INNER JOIN kl04_kv_objects ON objects.object_detailed_id = kl04_kv_objects.id
    WHERE
      objects.type = 'kl04kv_objects' AND
      objects.project_id = ?
    ORDER BY kl04_kv_objects.id DESC 
    LIMIT ? 
    OFFSET ?;

    `, projectID, limit, (page-1)*limit).Scan(&data).Error
	return data, err
}

func (repo *kl04kvObjectRepository) GetByID(id uint) (model.KL04KV_Object, error) {
	data := model.KL04KV_Object{}
	err := repo.db.Find(&data, "id = ?", id).Error
	return data, err
}

func (repo *kl04kvObjectRepository) Create(data dto.KL04KVObjectCreate) (model.KL04KV_Object, error) {
	result := data.DetailedInfo
	err := repo.db.Transaction(func(tx *gorm.DB) error {

		if err := tx.Create(&result).Error; err != nil {
			return err
		}

		object := data.BaseInfo
		object.ObjectDetailedID = result.ID
		if err := tx.Create(&object).Error; err != nil {
			return err
		}

		if len(data.Supervisors) != 0 {
			objectSupervisors := []model.ObjectSupervisors{}
			for _, supervisorWorkerID := range data.Supervisors {
				objectSupervisors = append(objectSupervisors, model.ObjectSupervisors{
					ObjectID:           object.ID,
					SupervisorWorkerID: supervisorWorkerID,
				})
			}

			if err := tx.CreateInBatches(&objectSupervisors, 5).Error; err != nil {
				return err
			}
		}

		if len(data.Teams) != 0 {
			objectTeams := []model.ObjectTeams{}
			for _, teamID := range data.Teams {
				objectTeams = append(objectTeams, model.ObjectTeams{
					TeamID:   teamID,
					ObjectID: object.ID,
				})
			}

			if err := tx.CreateInBatches(&objectTeams, 5).Error; err != nil {
				return err
			}
		}

		return nil
	})

	return result, err
}

func (repo *kl04kvObjectRepository) Delete(projectID, id uint) error {
	err := repo.db.Transaction(func(tx *gorm.DB) error {
		err := tx.Exec(`
      DELETE FROM object_supervisors
      WHERE object_supervisors.object_id = (
        SELECT DISTINCT(objects.id)
        FROM objects
          INNER JOIN kl04_kv_objects ON kl04_kv_objects.id = objects.object_detailed_id
        WHERE
          objects.project_id = ? AND
          kl04_kv_objects.id = ? AND
          objects.type = 'kl04kv_objects'
      );
    `, projectID, id).Error

		if err != nil {
			return err
		}

		err = tx.Exec(`
      DELETE FROM object_teams
      WHERE object_teams.object_id = (
        SELECT DISTINCT(objects.id)
        FROM objects
          INNER JOIN kl04_kv_objects ON kl04_kv_objects.id = objects.object_detailed_id
        WHERE
          objects.project_id = ? AND
          kl04_kv_objects.id = ? AND
          objects.type = 'kl04kv_objects'
      );
    `, projectID, id).Error
		if err != nil {
			return err
		}

		if err := tx.Table("kl04_kv_objects").Delete(&model.KL04KV_Object{}, "id = ?", id).Error; err != nil {
			return err
		}

		if err := tx.Table("objects").Delete(&model.Object{}, "object_detailed_id = ? AND type = 'kl04_kv_objects'", id).Error; err != nil {
			return err
		}

		return nil
	})

	return err
}

func (repo *kl04kvObjectRepository) Count(projectID uint) (int64, error) {
	var count int64
	err := repo.db.Raw(`
    SELECT COUNT(*)
    FROM objects
    WHERE
      objects.type = 'kl04kv_objects' AND
      objects.project_id = ?
    `, projectID).Scan(&count).Error
	return count, err
}

func (repo *kl04kvObjectRepository) Update(data dto.KL04KVObjectCreate) (model.KL04KV_Object, error) {
	kl04kv := model.KL04KV_Object{
		ID:        data.BaseInfo.ObjectDetailedID,
		Length:    data.DetailedInfo.Length,
		Nourashes: data.DetailedInfo.Nourashes,
	}

	err := repo.db.Transaction(func(tx *gorm.DB) error {

		if err := tx.Model(&model.KL04KV_Object{}).Where("id = ?", kl04kv.ID).Updates(&kl04kv).Error; err != nil {
			return err
		}

		object := model.Object{
			ID:               data.BaseInfo.ID,
			ProjectID:        data.BaseInfo.ProjectID,
			ObjectDetailedID: data.BaseInfo.ObjectDetailedID,
			Name:             data.BaseInfo.Name,
			Type:             data.BaseInfo.Type,
			Status:           data.BaseInfo.Status,
		}

		if err := tx.Model(&model.Object{}).Where("id = ?", object.ID).Updates(&object).Error; err != nil {
			return err
		}

		if err := tx.Delete(&model.ObjectSupervisors{}, "object_id = ?", object.ID).Error; err != nil {
			return err
		}

		if len(data.Supervisors) != 0 {
			objectSupervisors := []model.ObjectSupervisors{}
			for _, supervisorWorkerID := range data.Supervisors {
				objectSupervisors = append(objectSupervisors, model.ObjectSupervisors{
					ObjectID:           object.ID,
					SupervisorWorkerID: supervisorWorkerID,
				})
			}

			if err := tx.CreateInBatches(&objectSupervisors, 5).Error; err != nil {
				return err
			}
		}

		if err := tx.Delete(&model.ObjectTeams{}, "object_id = ?", object.ID).Error; err != nil {
			return err
		}

		if len(data.Teams) != 0 {
			objectTeams := []model.ObjectTeams{}
			for _, teamID := range data.Teams {
				objectTeams = append(objectTeams, model.ObjectTeams{
					ObjectID: object.ID,
					TeamID:   teamID,
				})
			}

			if err := tx.CreateInBatches(&objectTeams, 5).Error; err != nil {
				return err
			}
		}

		return nil
	})

	return kl04kv, err
}

func (repo *kl04kvObjectRepository) CreateInBatches(objects []model.Object, kl04kvs []model.KL04KV_Object, supervisors []uint) ([]model.KL04KV_Object, error) {
	err := repo.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.CreateInBatches(&kl04kvs, 10).Error; err != nil {
			return err
		}

		for index := range objects {
			objects[index].ObjectDetailedID = kl04kvs[index].ID
		}

		if err := tx.CreateInBatches(&objects, 10).Error; err != nil {
			return err
		}
		
		return nil
	})

	return kl04kvs, err
}
