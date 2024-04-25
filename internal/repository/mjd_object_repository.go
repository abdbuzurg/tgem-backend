package repository

import (
	"backend-v2/internal/dto"
	"backend-v2/model"

	"gorm.io/gorm"
)

type mjdObjectRepository struct {
	db *gorm.DB
}

func InitMJDObjectRepository(db *gorm.DB) IMJDObjectRepository {
	return &mjdObjectRepository{
		db: db,
	}
}

type IMJDObjectRepository interface {
	GetPaginated(page, limit int, projectID uint) ([]dto.MJDObjectPaginatedQuery, error)
	Count(projectID uint) (int64, error)
	Create(data dto.MJDObjectCreate) (model.MJD_Object, error)
	Update(data dto.MJDObjectCreate) (model.MJD_Object, error)
	Delete(id, projectID uint) error
}

func (repo *mjdObjectRepository) GetPaginated(page, limit int, projectID uint) ([]dto.MJDObjectPaginatedQuery, error) {
	data := []dto.MJDObjectPaginatedQuery{}
	err := repo.db.Raw(`
      SELECT 
        objects.id as object_id,
        objects.object_detailed_id as object_detailed_id,
        objects.name as name,
        objects.status as status,
        mjd_objects.model as model,
        mjd_objects.amount_stores as amount_stores,
        mjd_objects.amount_entrances as amount_entrances,
        mjd_objects.has_basement as has_basement,
        workers.name as supervisor_name
      FROM objects
        INNER JOIN mjd_objects ON objects.object_detailed_id = mjd_objects.id
        INNER JOIN supervisor_objects ON objects.id = supervisor_objects.object_id
        INNER JOIN workers ON workers.id = supervisor_objects.supervisor_worker_id
      WHERE
        objects.type = 'mjd_objects' AND
        objects.project_id = ?
      ORDER BY mjd_objects.id DESC 
      LIMIT ? 
      OFFSET ?;
    `, projectID, limit, (page-1)*limit).Scan(&data).Error

	return data, err
}

func (repo *mjdObjectRepository) Count(projectID uint) (int64, error) {
	var count int64
	err := repo.db.Raw(`
    SELECT COUNT(*)
    FROM objects
    WHERE
      objects.type = 'mjd_objects' AND
      objects.project_id = ?
    `, projectID).Scan(&count).Error
	return count, err
}

func (repo *mjdObjectRepository) Create(data dto.MJDObjectCreate) (model.MJD_Object, error) {
	mjd := model.MJD_Object{
		ID:              0,
		Model:           data.DetailedInfo.Model,
		AmountStores:    data.DetailedInfo.AmountStores,
		AmountEntrances: data.DetailedInfo.AmountEntrances,
		HasBasement:     data.DetailedInfo.HasBasement,
	}

	err := repo.db.Transaction(func(tx *gorm.DB) error {

		if err := tx.Create(&mjd).Error; err != nil {
			return err
		}

		object := model.Object{
			ID:               0,
			ObjectDetailedID: mjd.ID,
			ProjectID:        data.BaseInfo.ProjectID,
			Name:             data.BaseInfo.Name,
			Status:           data.BaseInfo.Status,
			Type:             "mjd_objects",
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

	return mjd, err
}

func (repo *mjdObjectRepository) Update(data dto.MJDObjectCreate) (model.MJD_Object, error) {
	mjd := model.MJD_Object{
		ID:              data.BaseInfo.ObjectDetailedID,
		Model:           data.DetailedInfo.Model,
		AmountStores:    data.DetailedInfo.AmountStores,
		AmountEntrances: data.DetailedInfo.AmountEntrances,
		HasBasement:     data.DetailedInfo.HasBasement,
	}

	err := repo.db.Transaction(func(tx *gorm.DB) error {

		if err := tx.Model(&model.MJD_Object{}).Where("id = ?", mjd.ID).Updates(&mjd).Error; err != nil {
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

	return mjd, err
}

func (repo *mjdObjectRepository) Delete(id, projectID uint) error {
	return repo.db.Transaction(func(tx *gorm.DB) error {
		err := tx.Exec(`
      DELETE FROM supervisor_objects
      WHERE supervisor_objects.object_id = (
        SELECT DISTINCT(objects.id)
        FROM objects
          INNER JOIN mjd_objects ON mjd_objects.id = objects.object_detailed_id
        WHERE
          objects.project_id = ? AND
          mjd_objects.id = ? AND
          objects.type = 'mjd_objects'
      );
    `, projectID, id).Error

		if err != nil {
			return err
		}

		if err := tx.Table("mjd_objects").Delete(&model.MJD_Object{}, "id = ?", id).Error; err != nil {
			return err
		}

		if err := tx.Table("objects").Delete(&model.Object{}, "object_detailed_id = ? AND type = 'mjd_objects'", id).Error; err != nil {
			return err
		}

		return nil
	})
}
