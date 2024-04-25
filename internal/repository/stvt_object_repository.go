package repository

import (
	"backend-v2/internal/dto"
	"backend-v2/model"

	"gorm.io/gorm"
)

type stvtObjectRepository struct {
	db *gorm.DB
}

func InitSTVTObjectRepository(db *gorm.DB) ISTVTObjectRepository {
	return &stvtObjectRepository{
		db: db,
	}
}

type ISTVTObjectRepository interface {
	GetPaginated(page, limit int, projectID uint) ([]dto.STVTObjectPaginatedQuery, error)
	Count(projectID uint) (int64, error)
	Create(data dto.STVTObjectCreate) (model.STVT_Object, error)
	Update(data dto.STVTObjectCreate) (model.STVT_Object, error)
	Delete(id, projectID uint) error
}

func (repo *stvtObjectRepository) GetPaginated(page, limit int, projectID uint) ([]dto.STVTObjectPaginatedQuery, error) {
	data := []dto.STVTObjectPaginatedQuery{}
	err := repo.db.Raw(`
      SELECT 
        objects.id as object_id,
        objects.object_detailed_id as object_detailed_id,
        objects.name as name,
        objects.status as status,
        stvt_objects.voltage_class as voltage_class,
        stvt_objects.tt_coefficient as tt_coefficient,
        workers.name as supervisor_name
      FROM objects
        INNER JOIN stvt_objects ON objects.object_detailed_id = stvt_objects.id
        INNER JOIN supervisor_objects ON objects.id = supervisor_objects.object_id
        INNER JOIN workers ON workers.id = supervisor_objects.supervisor_worker_id
      WHERE
        objects.type = 'stvt_objects' AND
        objects.project_id = ?
      ORDER BY stvt_objects.id DESC 
      LIMIT ? 
      OFFSET ?;
    `, projectID, limit, (page-1)*limit).Scan(&data).Error

	return data, err
}

func (repo *stvtObjectRepository) Count(projectID uint) (int64, error) {
	var count int64
	err := repo.db.Raw(`
    SELECT COUNT(*)
    FROM objects
    WHERE
      objects.type = 'stvt_objects' AND
      objects.project_id = ?
    `, projectID).Scan(&count).Error
	return count, err
}

func (repo *stvtObjectRepository) Create(data dto.STVTObjectCreate) (model.STVT_Object, error) {
	sip := model.STVT_Object{
		VoltageClass:  data.DetailedInfo.VoltageClass,
		TTCoefficient: data.DetailedInfo.TTCoefficient,
	}

	err := repo.db.Transaction(func(tx *gorm.DB) error {

		if err := tx.Create(&sip).Error; err != nil {
			return err
		}

		object := model.Object{
			ID:               0,
			ObjectDetailedID: sip.ID,
			ProjectID:        data.BaseInfo.ProjectID,
			Name:             data.BaseInfo.Name,
			Status:           data.BaseInfo.Status,
			Type:             "stvt_objects",
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

	return sip, err

}

func (repo *stvtObjectRepository) Update(data dto.STVTObjectCreate) (model.STVT_Object, error) {
	stvt := model.STVT_Object{
		ID:            data.BaseInfo.ObjectDetailedID,
		VoltageClass:  data.DetailedInfo.VoltageClass,
		TTCoefficient: data.DetailedInfo.TTCoefficient,
	}

	err := repo.db.Transaction(func(tx *gorm.DB) error {

		if err := tx.Model(&model.STVT_Object{}).Where("id = ?", stvt.ID).Updates(&stvt).Error; err != nil {
			return err
		}

		object := model.Object{
			ID:               data.BaseInfo.ID,
			ProjectID:        data.BaseInfo.ProjectID,
			ObjectDetailedID: data.BaseInfo.ObjectDetailedID,
			Name:             data.BaseInfo.Name,
			Type:             "stvt_objects",
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

	return stvt, err
}

func (repo *stvtObjectRepository) Delete(id, projectID uint) error {
	return repo.db.Transaction(func(tx *gorm.DB) error {
		err := tx.Exec(`
      DELETE FROM supervisor_objects
      WHERE supervisor_objects.object_id = (
        SELECT DISTINCT(objects.id)
        FROM objects
          INNER JOIN stvt_objects ON stvt_objects.id = objects.object_detailed_id
        WHERE
          objects.project_id = ? AND
          stvt_objects.id = ? AND
          objects.type = 'stvt_objects'
      );
    `, projectID, id).Error

		if err != nil {
			return err
		}

		if err := tx.Table("stvt_objects").Delete(&model.STVT_Object{}, "id = ?", id).Error; err != nil {
			return err
		}

		if err := tx.Table("objects").Delete(&model.Object{}, "object_detailed_id = ? AND type = 'stvt_objects'", id).Error; err != nil {
			return err
		}

		return nil
	})
}
