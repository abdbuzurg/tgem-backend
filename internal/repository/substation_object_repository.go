package repository

import (
	"backend-v2/internal/dto"
	"backend-v2/model"

	"gorm.io/gorm"
)

type substationObjectRepository struct {
	db *gorm.DB
}

func InitSubstationObjectRepository(db *gorm.DB) ISubstationObjectRepository {
	return &substationObjectRepository{
		db: db,
	}
}

type ISubstationObjectRepository interface{
	GetPaginated(page, limit int, projectID uint) ([]dto.SubstationObjectPaginatedQuery, error)
	Count(projectID uint) (int64, error)
	Create(data dto.SubstationObjectCreate) (model.Substation_Object, error)
	Update(data dto.SubstationObjectCreate) (model.Substation_Object, error)
	Delete(id, projectID uint) error
	CreateInBatches(objects []model.Object, tps []model.Substation_Object, supervisors []uint) ([]model.Substation_Object, error)
}

func (repo *substationObjectRepository) GetPaginated(page, limit int, projectID uint) ([]dto.SubstationObjectPaginatedQuery, error) {
	data := []dto.SubstationObjectPaginatedQuery{}
	err := repo.db.Raw(`
      SELECT 
        objects.id as object_id,
        objects.object_detailed_id as object_detailed_id,
        objects.name as name,
        objects.status as status,
        substation_objects.voltage_class as voltage_class,
        substation_objects.number_of_transformers as number_of_transformers
      FROM objects
        INNER JOIN substation_objects ON objects.object_detailed_id = substation_objects.id
     WHERE
        objects.type = 'substation_objects' AND
        objects.project_id = ?
      ORDER BY substation_objects.id DESC 
      LIMIT ? 
      OFFSET ?;
    `, projectID, limit, (page-1)*limit).Scan(&data).Error

	return data, err
}

func (repo *substationObjectRepository) Count(projectID uint) (int64, error) {
	var count int64
	err := repo.db.Raw(`
    SELECT COUNT(*)
    FROM objects
    WHERE
      objects.type = 'substation_objects' AND
      objects.project_id = ?
    `, projectID).Scan(&count).Error
	return count, err
}

func (repo *substationObjectRepository) Create(data dto.SubstationObjectCreate) (model.Substation_Object, error) {
	substation := model.Substation_Object{
		VoltageClass:         data.DetailedInfo.VoltageClass,
		NumberOfTransformers: data.DetailedInfo.NumberOfTransformers,
	}

	err := repo.db.Transaction(func(tx *gorm.DB) error {

		if err := tx.Create(&substation).Error; err != nil {
			return err
		}

		object := model.Object{
			ID:               0,
			ObjectDetailedID: substation.ID,
			ProjectID:        data.BaseInfo.ProjectID,
			Name:             data.BaseInfo.Name,
			Status:           data.BaseInfo.Status,
			Type:             "substation_objects",
		}

		if err := tx.Create(&object).Error; err != nil {
			return err
		}

		if len(data.Supervisors) != 0 {
			object_supervisors := []model.ObjectSupervisors{}
			for _, supervisorID := range data.Supervisors {
				object_supervisors = append(object_supervisors, model.ObjectSupervisors{
					ObjectID:           object.ID,
					SupervisorWorkerID: supervisorID,
				})
			}

			if err := tx.CreateInBatches(&object_supervisors, 5).Error; err != nil {
				return err
			}
		}

		if len(data.Teams) != 0 {
			object_teams := []model.ObjectTeams{}
			for _, teamID := range data.Teams {
				object_teams = append(object_teams, model.ObjectTeams{
					ObjectID: object.ID,
					TeamID:   teamID,
				})
			}

			if err := tx.CreateInBatches(&object_teams, 5).Error; err != nil {
				return err
			}
		}

		return nil
	})

	return substation, err
}

func (repo *substationObjectRepository) Update(data dto.SubstationObjectCreate) (model.Substation_Object, error) {
	tp := model.Substation_Object{
		ID:                   data.BaseInfo.ObjectDetailedID,
		VoltageClass:         data.DetailedInfo.VoltageClass,
		NumberOfTransformers: data.DetailedInfo.NumberOfTransformers,
	}

	err := repo.db.Transaction(func(tx *gorm.DB) error {

		if err := tx.Model(&model.Substation_Object{}).Where("id = ?", tp.ID).Updates(&tp).Error; err != nil {
			return err
		}

		object := model.Object{
			ID:               data.BaseInfo.ID,
			ProjectID:        data.BaseInfo.ProjectID,
			ObjectDetailedID: data.BaseInfo.ObjectDetailedID,
			Name:             data.BaseInfo.Name,
			Type:             "substation_objects",
			Status:           data.BaseInfo.Status,
		}

		if err := tx.Model(&model.Object{}).Where("id = ?", object.ID).Updates(&object).Error; err != nil {
			return err
		}

		if err := tx.Delete(&model.ObjectSupervisors{}, "object_id = ?", object.ID).Error; err != nil {
			return err
		}

		if len(data.Supervisors) != 0 {
			object_supervisors := []model.ObjectSupervisors{}
			for _, supervisorWorkerID := range data.Supervisors {
				object_supervisors = append(object_supervisors, model.ObjectSupervisors{
					ObjectID:           object.ID,
					SupervisorWorkerID: supervisorWorkerID,
				})
			}

			if err := tx.CreateInBatches(&object_supervisors, 5).Error; err != nil {
				return err
			}
		}

		if err := tx.Delete(&model.ObjectTeams{}, "object_id = ?", object.ID).Error; err != nil {
			return err
		}

		if len(data.Teams) != 0 {
			object_teams := []model.ObjectTeams{}
			for _, teamID := range data.Teams {
				object_teams = append(object_teams, model.ObjectTeams{
					ObjectID: object.ID,
					TeamID:   teamID,
				})
			}

			if err := tx.CreateInBatches(&object_teams, 5).Error; err != nil {
				return err
			}
		}

		return nil
	})

	return tp, err
}

func (repo *substationObjectRepository) Delete(id, projectID uint) error {
	return repo.db.Transaction(func(tx *gorm.DB) error {
		err := tx.Exec(`
      DELETE FROM object_supervisors
      WHERE object_supervisors.object_id = (
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

		err = tx.Exec(`
      DELETE FROM object_teams
      WHERE object_teams.object_id = (
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

		if err := tx.Table("substation_objects").Delete(&model.TP_Object{}, "id = ?", id).Error; err != nil {
			return err
		}

		if err := tx.Table("objects").Delete(&model.Object{}, "object_detailed_id = ? AND type = 'tp_objects'", id).Error; err != nil {
			return err
		}

		return nil
	})
}

func (repo *substationObjectRepository) CreateInBatches(objects []model.Object, substations []model.Substation_Object, supervisors []uint) ([]model.Substation_Object, error) {
	err := repo.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.CreateInBatches(&substations, 10).Error; err != nil {
			return err
		}

		for index := range objects {
			objects[index].ObjectDetailedID = substations[index].ID
		}

		if err := tx.CreateInBatches(&objects, 10).Error; err != nil {
			return err
		}

		return nil
	})

	return substations, err
}
