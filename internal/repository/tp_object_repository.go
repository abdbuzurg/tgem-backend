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
	GetAll(projectID uint) ([]dto.TPObjectPaginatedQuery, error)
	GetPaginated(page, limit int, filter dto.TPObjectSearchParameters) ([]dto.TPObjectPaginatedQuery, error)
	Count(filter dto.TPObjectSearchParameters) (int64, error)
	Create(data dto.TPObjectCreate) (model.TP_Object, error)
	Update(data dto.TPObjectCreate) (model.TP_Object, error)
	Delete(id, projectID uint) error
	CreateInBatches(data []dto.TPObjectImportData) error
	GetTPNames(projectID uint) ([]string, error)
	GetObjectNamesForSearch(projectID uint) ([]dto.DataForSelect[string], error)
}

func (repo *tpObjectRepository) GetAll(projectID uint) ([]dto.TPObjectPaginatedQuery, error) {
	data := []dto.TPObjectPaginatedQuery{}
	err := repo.db.Raw(`
      SELECT 
        objects.id as object_id,
        objects.object_detailed_id as object_detailed_id,
        objects.name as name,
        objects.status as status,
        tp_objects.model as model,
        tp_objects.voltage_class as voltage_class
      FROM objects
        INNER JOIN tp_objects ON objects.object_detailed_id = tp_objects.id
      WHERE
        objects.type = 'tp_objects' AND
        objects.project_id = ?
      ORDER BY tp_objects.id DESC
    `, projectID).Scan(&data).Error

	return data, err
}

func (repo *tpObjectRepository) GetPaginated(page, limit int, filter dto.TPObjectSearchParameters) ([]dto.TPObjectPaginatedQuery, error) {
	data := []dto.TPObjectPaginatedQuery{}
	err := repo.db.Raw(`
    SELECT DISTINCT 
      objects.id as object_id,
      tp_objects.id as object_detailed_id,
      objects.name as name,
      objects.status as status,
      tp_objects.model as model,
      tp_objects.voltage_class as voltage_class
    FROM objects
    INNER JOIN tp_objects ON objects.object_detailed_id = tp_objects.id
    FULL JOIN object_teams ON object_teams.object_id = objects.id
    FULL JOIN object_supervisors ON object_supervisors.object_id = objects.id
    FULL JOIN tp_nourashes_objects ON tp_nourashes_objects.target_id = objects.id
    WHERE
      objects.type = 'tp_objects' AND
      objects.project_id = ? AND
      (nullif(?, '') IS NULL OR objects.name = ?) AND
      (nullif(?, 0) IS NULL OR object_teams.team_id = ?) AND
      (nullif(?, 0) IS NULL OR object_supervisors.supervisor_worker_id = ?)
    ORDER BY tp_objects.id DESC 
    LIMIT ? 
    OFFSET ?;
    `, filter.ProjectID,
		filter.ObjectName, filter.ObjectName,
		filter.TeamID, filter.TeamID,
		filter.SupervisorWorkerID, filter.SupervisorWorkerID,
		limit, (page-1)*limit).Scan(&data).Error

	return data, err
}

func (repo *tpObjectRepository) Count(filter dto.TPObjectSearchParameters) (int64, error) {
	var count int64
	err := repo.db.Raw(`
    SELECT DISTINCT COUNT(*)
    FROM objects
    FULL JOIN object_teams ON object_teams.object_id = objects.id
    FULL JOIN object_supervisors ON object_supervisors.object_id = objects.id
    FULL JOIN tp_nourashes_objects ON tp_nourashes_objects.target_id = objects.id
    WHERE
      objects.type = 'tp_objects' AND
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

func (repo *tpObjectRepository) Create(data dto.TPObjectCreate) (model.TP_Object, error) {
	tp := model.TP_Object{
		Model:        data.DetailedInfo.Model,
		VoltageClass: data.DetailedInfo.VoltageClass,
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

	return tp, err
}

func (repo *tpObjectRepository) Update(data dto.TPObjectCreate) (model.TP_Object, error) {
	tp := model.TP_Object{
		ID:           data.BaseInfo.ObjectDetailedID,
		Model:        data.DetailedInfo.Model,
		VoltageClass: data.DetailedInfo.VoltageClass,
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

func (repo *tpObjectRepository) Delete(id, projectID uint) error {
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

		if err := tx.Table("tp_objects").Delete(&model.TP_Object{}, "id = ?", id).Error; err != nil {
			return err
		}

		if err := tx.Table("objects").Delete(&model.Object{}, "object_detailed_id = ? AND type = 'tp_objects'", id).Error; err != nil {
			return err
		}

		return nil
	})
}

func (repo *tpObjectRepository) CreateInBatches(data []dto.TPObjectImportData) error {
	return repo.db.Transaction(func(tx *gorm.DB) error {
		for index, row := range data {
			tp := row.TP
			if err := tx.Create(&tp).Error; err != nil {
				return err
			}

			object := row.Object
			object.ObjectDetailedID = tp.ID
			data[index].Object.ObjectDetailedID = tp.ID
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

func (repo *tpObjectRepository) GetTPNames(projectID uint) ([]string, error) {
	var result []string
	err := repo.db.Raw(`SELECT name FROM objects WHERE project_id = ? AND type = 'tp_objects'`, projectID).Scan(&result).Error
	return result, err
}

func (repo *tpObjectRepository) GetObjectNamesForSearch(projectID uint) ([]dto.DataForSelect[string], error) {
	data := []dto.DataForSelect[string]{}
	err := repo.db.Raw(`
    SELECT 
      objects.name as "label",
      objects.name as "value"
    FROM objects
    INNER JOIN tp_objects ON tp_objects.id = objects.object_detailed_id
    WHERE
      objects.project_id = ? AND
      objects.type = 'tp_objects'
    `, projectID).Scan(&data).Error

	return data, err
}
