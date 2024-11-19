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

type ISubstationObjectRepository interface {
	GetPaginated(page, limit int, filter dto.SubstationObjectSearchParameters) ([]dto.SubstationObjectPaginatedQuery, error)
	Count(filter dto.SubstationObjectSearchParameters) (int64, error)
	Create(data dto.SubstationObjectCreate) (model.Substation_Object, error)
	Update(data dto.SubstationObjectCreate) (model.Substation_Object, error)
	Delete(id, projectID uint) error
	CreateInBatches(objects []model.Object, tps []model.Substation_Object, supervisors []uint) ([]model.Substation_Object, error)
	GetObjectNamesForSearch(projectID uint) ([]dto.DataForSelect[string], error)
	Import(data []dto.SubstationObjectImportData) error
	GetAllNames(projectID uint) ([]string, error)
	GetByName(name string) (model.Object, error)
	GetAll(projectID uint) ([]model.Object, error)
}

func (repo *substationObjectRepository) GetAll(projectID uint) ([]model.Object, error) {
  result := []model.Object{}
  err := repo.db.Find(&result, "type = 'substation_objects' AND project_id = ?", projectID).Error
  return result, err
}

func (repo *substationObjectRepository) GetPaginated(page, limit int, filter dto.SubstationObjectSearchParameters) ([]dto.SubstationObjectPaginatedQuery, error) {
	data := []dto.SubstationObjectPaginatedQuery{}
	err := repo.db.Raw(`
    SELECT DISTINCT
      objects.id as object_id,
      substation_objects.id as object_detailed_id,
      objects.name as name,
      objects.status as status,
      substation_objects.voltage_class as voltage_class,
      substation_objects.number_of_transformers as number_of_transformers
    FROM objects
    INNER JOIN substation_objects ON objects.object_detailed_id = substation_objects.id
    FULL JOIN object_teams ON object_teams.object_id = objects.id
    FULL JOIN object_supervisors ON object_supervisors.object_id = objects.id
    WHERE
      objects.type = 'substation_objects' AND
      objects.project_id = ? AND
      (nullif(?, '') IS NULL OR objects.name = ?) AND
      (nullif(?, 0) IS NULL OR object_teams.team_id = ?) AND
      (nullif(?, 0) IS NULL OR object_supervisors.supervisor_worker_id = ?)
    ORDER BY substation_objects.id DESC 
    LIMIT ? 
    OFFSET ?;
    `, filter.ProjectID,
		filter.ObjectName, filter.ObjectName,
		filter.TeamID, filter.TeamID,
		filter.SupervisorWorkerID, filter.SupervisorWorkerID,
		limit, (page-1)*limit).Scan(&data).Error

	return data, err
}

func (repo *substationObjectRepository) Count(filter dto.SubstationObjectSearchParameters) (int64, error) {
	var count int64
	err := repo.db.Raw(`
    SELECT DISTINCT COUNT(*)
    FROM objects
    FULL JOIN object_teams ON object_teams.object_id = objects.id
    FULL JOIN object_supervisors ON object_supervisors.object_id = objects.id
    WHERE
      objects.type = 'substation_objects' AND
      objects.project_id = ? AND
      (nullif(?, '') IS NULL OR objects.name = ?) AND
      (nullif(?, 0) IS NULL OR object_teams.team_id = ?) AND
      (nullif(?, 0) IS NULL OR object_supervisors.supervisor_worker_id = ?)
    `, filter.ProjectID,
		filter.ObjectName, filter.ObjectName,
		filter.TeamID, filter.TeamID,
		filter.SupervisorWorkerID, filter.SupervisorWorkerID).Scan(&count).Error
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

func (repo *substationObjectRepository) GetObjectNamesForSearch(projectID uint) ([]dto.DataForSelect[string], error) {
	data := []dto.DataForSelect[string]{}
	err := repo.db.Raw(`
    SELECT 
      objects.name as "label",
      objects.name as "value"
    FROM objects
    INNER JOIN substation_objects ON substation_objects.id = objects.object_detailed_id
    WHERE
      objects.project_id = ? AND
      objects.type = 'substation_objects'
    `, projectID).Scan(&data).Error

	return data, err
}

func (repo *substationObjectRepository) Import(data []dto.SubstationObjectImportData) error {
	return repo.db.Transaction(func(tx *gorm.DB) error {
		for index, row := range data {
			substation := row.Substation
			if err := tx.Create(&substation).Error; err != nil {
				return err
			}

			object := row.Object
			object.ObjectDetailedID = substation.ID
			data[index].Object.ObjectDetailedID = substation.ID
			if err := tx.Create(&object).Error; err != nil {
				return err
			}

			if row.ObjectSupervisors.SupervisorWorkerID != 0 {
				data[index].ObjectSupervisors.ObjectID = object.ID
				if err := tx.Create(&data[index].ObjectSupervisors).Error; err != nil {
					return err
				}
			}

			if row.ObjectTeam.TeamID != 0 {
				data[index].ObjectTeam.ObjectID = object.ID
				if err := tx.Create(&data[index].ObjectTeam).Error; err != nil {
					return err
				}
			}
		}

		return nil
	})
}

func (repo *substationObjectRepository) GetAllNames(projectID uint) ([]string, error) {
	result := []string{}
	err := repo.db.Raw(`SELECT name FROM objects WHERE objects.type='substation_objects' AND objects.project_id = ?`, projectID).Scan(&result).Error
	return result, err
}

func (repo *substationObjectRepository) GetByName(name string) (model.Object, error) {
	result := model.Object{}
	err := repo.db.First(&result, "name = ?", name).Error
	return result, err
}
