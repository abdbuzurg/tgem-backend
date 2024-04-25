package repository

import (
	"backend-v2/internal/dto"
	"backend-v2/model"

	"gorm.io/gorm"
)

type tpObjectRepository struct {
	db *gorm.DB
}

func InitTPObjectRepository(db *gorm.DB) ITPObjectRepository {
	return &tpObjectRepository{
		db: db,
	}
}

type ITPObjectRepository interface {
	GetPaginated(page, limit int, projectID uint) ([]dto.TPObjectPaginatedQuery, error)
	Count(projectID uint) (int64, error)
	Create(data dto.TPObjectCreate) (model.TP_Object, error)
	Update(data dto.TPObjectCreate) (model.TP_Object, error)
	Delete(id, projectID uint) error
}

func (repo *tpObjectRepository) GetPaginated(page, limit int, projectID uint) ([]dto.TPObjectPaginatedQuery, error) {
	data := []dto.TPObjectPaginatedQuery{}
	err := repo.db.Raw(`
      SELECT 
        objects.id as object_id,
        objects.object_detailed_id as object_detailed_id,
        objects.name as name,
        objects.status as status,
        tp_objects.model as model,
        tp_objects.voltage_class as voltage_class,
        tp_objects.nourashes as nourashes,
        workers.name as supervisor_name
      FROM objects
        INNER JOIN tp_objects ON objects.object_detailed_id = tp_objects.id
        INNER JOIN supervisor_objects ON objects.id = supervisor_objects.object_id
        INNER JOIN workers ON workers.id = supervisor_objects.supervisor_worker_id
      WHERE
        objects.type = 'tp_objects' AND
        objects.project_id = ?
      ORDER BY tp_objects.id DESC 
      LIMIT ? 
      OFFSET ?;
    `, projectID, limit, (page-1)*limit).Scan(&data).Error

	return data, err
}

func (repo *tpObjectRepository) Count(projectID uint) (int64, error) {
	var count int64
	err := repo.db.Raw(`
    SELECT COUNT(*)
    FROM objects
    WHERE
      objects.type = 'tp_objects' AND
      objects.project_id = ?
    `, projectID).Scan(&count).Error
	return count, err
}

func (repo *tpObjectRepository) Create(data dto.TPObjectCreate) (model.TP_Object, error) {
	tp := model.TP_Object{
		Model:        data.DetailedInfo.Model,
		VoltageClass: data.DetailedInfo.VoltageClass,
		Nourashes:    data.DetailedInfo.Nourashes,
	}

	err := repo.db.Transaction(func(tx *gorm.DB) error {

		if err := tx.Create(&tp).Error; err != nil {
			return err
		}

		object := model.Object{
			ID:               0,
			ObjectDetailedID: tp.ID,
			ProjectID:        data.BaseInfo.ProjectID,
			Name:             data.BaseInfo.Name,
			Status:           data.BaseInfo.Status,
			Type:             "tp_objects",
		}

		if err := tx.Create(&object).Error; err != nil {
			return err
		}

		supervisors_object := []model.SupervisorObjects{}
		for _, supervisorID := range data.Supervisors {
			supervisors_object = append(supervisors_object, model.SupervisorObjects{
				ObjectID:           object.ID,
				SupervisorWorkerID: supervisorID,
			})
		}

		if err := tx.CreateInBatches(&supervisors_object, 5).Error; err != nil {
			return err
		}

		return nil
	})

	return tp, err
}

func (repo *tpObjectRepository) Update(data dto.TPObjectCreate) (model.TP_Object, error) {
	tp := model.TP_Object{
		ID:           data.BaseInfo.ObjectDetailedID,
		Model:        data.DetailedInfo.Model,
		VoltageClass: data.DetailedInfo.VoltageClass,
		Nourashes:    data.DetailedInfo.Nourashes,
	}

	err := repo.db.Transaction(func(tx *gorm.DB) error {

		if err := tx.Model(&model.TP_Object{}).Where("id = ?", tp.ID).Updates(&tp).Error; err != nil {
			return err
		}

		object := model.Object{
			ID:               data.BaseInfo.ID,
			ProjectID:        data.BaseInfo.ProjectID,
			ObjectDetailedID: data.BaseInfo.ObjectDetailedID,
			Name:             data.BaseInfo.Name,
			Type:             "tp_objects",
			Status:           data.BaseInfo.Status,
		}

		if err := tx.Model(&model.Object{}).Where("id = ?", object.ID).Updates(&object).Error; err != nil {
			return err
		}

		if err := tx.Delete(&model.SupervisorObjects{}, "object_id = ?", object.ID).Error; err != nil {
			return err
		}

		supervisorsObject := []model.SupervisorObjects{}
		for _, supervisorWorkerID := range data.Supervisors {
			supervisorsObject = append(supervisorsObject, model.SupervisorObjects{
				ObjectID:           object.ID,
				SupervisorWorkerID: supervisorWorkerID,
			})
		}

		if err := tx.CreateInBatches(&supervisorsObject, 5).Error; err != nil {
			return err
		}

		return nil
	})

	return tp, err
}

func (repo *tpObjectRepository) Delete(id, projectID uint) error {
	return repo.db.Transaction(func(tx *gorm.DB) error {
		err := tx.Exec(`
      DELETE FROM supervisor_objects
      WHERE supervisor_objects.object_id = (
        SELECT DISTINCT(objects.id)
        FROM objects
          INNER JOIN tp_objects ON tp_objects.id = objects.object_detailed_id
        WHERE
          objects.project_id = ? AND
          tp_objects.id = ? AND
          objects.type = 'tp_objects'
      );
    `, projectID, id).Error

		if err != nil {
			return err
		}

		if err := tx.Table("tp_objects").Delete(&model.TP_Object{}, "id = ?", id).Error; err != nil {
			return err
		}

		if err := tx.Table("objects").Delete(&model.Object{}, "object_detailed_id = ? AND type = 'tp_objects'", id).Error; err != nil {
			return err
		}

		return nil
	})
}
