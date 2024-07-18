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
	GetPaginated(page, limit int, filter dto.MJDObjectSearchParameters) ([]dto.MJDObjectPaginatedQuery, error)
	Count(filter dto.MJDObjectSearchParameters) (int64, error)
	Create(data dto.MJDObjectCreate) (model.MJD_Object, error)
	Update(data dto.MJDObjectCreate) (model.MJD_Object, error)
	Delete(id, projectID uint) error
	CreateInBatches(objects []model.Object, mjds []model.MJD_Object, supervisors []uint) ([]model.MJD_Object, error)
	Import(data []dto.MJDObjectImportData) error
  GetObjectNamesForSearch(projectID uint) ([]dto.DataForSelect[string], error)
}

func (repo *mjdObjectRepository) GetPaginated(page, limit int, filter dto.MJDObjectSearchParameters) ([]dto.MJDObjectPaginatedQuery, error) {
	data := []dto.MJDObjectPaginatedQuery{}
	err := repo.db.Raw(`
    SELECT 
      objects.id as object_id,
      mjd_objects.id as object_detailed_id,
      objects.name as name,
      objects.status as status,
      mjd_objects.model as model,
      mjd_objects.amount_stores as amount_stores,
      mjd_objects.amount_entrances as amount_entrances
    FROM objects
    INNER JOIN mjd_objects ON mjd_objects.id = objects.object_detailed_id
    FULL JOIN object_teams ON object_teams.object_id = objects.id
    FULL JOIN object_supervisors ON object_supervisors.object_id = objects.id
    FULL JOIN tp_nourashes_objects ON tp_nourashes_objects.target_id = objects.id
    WHERE
      objects.type = 'mjd_objects' AND
      objects.project_id = ? AND
      (nullif(?, '') IS NULL OR objects.name = ?) AND
      (nullif(?, 0) IS NULL OR object_teams.team_id = ?) AND
      (nullif(?, 0) IS NULL OR object_supervisors.supervisor_worker_id = ?) AND
      (nullif(?, 0) IS NULL OR tp_nourashes_objects.tp_object_id = ?)
    ORDER BY mjd_objects.id DESC 
    LIMIT ? 
    OFFSET ?;

    `, filter.ProjectID,
    filter.ObjectName, filter.ObjectName,
    filter.TeamID, filter.TeamID,
    filter.SupervisorWorkerID, filter.SupervisorWorkerID,
    filter.TPObjectID, filter.TPObjectID,
    limit, (page-1)*limit).Scan(&data).Error

	return data, err
}

func (repo *mjdObjectRepository) Count(filter dto.MJDObjectSearchParameters) (int64, error) {
	var count int64
	err := repo.db.Raw(`
    SELECT COUNT(*)
    FROM objects
    FULL JOIN object_teams ON object_teams.object_id = objects.id
    FULL JOIN object_supervisors ON object_supervisors.object_id = objects.id
    FULL JOIN tp_nourashes_objects ON tp_nourashes_objects.target_id = objects.id
    WHERE
      objects.type = 'kl04kv_objects' AND
      objects.project_id = ? AND
      (nullif(?, '') IS NULL OR objects.name = ?) AND
      (nullif(?, 0) IS NULL OR object_teams.team_id = ?) AND
      (nullif(?, 0) IS NULL OR object_supervisors.supervisor_worker_id = ?) AND
      (nullif(?, 0) IS NULL OR tp_nourashes_objects.tp_object_id = ?)
    `,
    filter.ProjectID,
    filter.ObjectName, filter.ObjectName,
    filter.TeamID, filter.TeamID,
    filter.SupervisorWorkerID, filter.SupervisorWorkerID,
    filter.TPObjectID, filter.TPObjectID,
    ).Scan(&count).Error
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

		if len(data.Supervisors) != 0 {

			objectSupervisors := []model.ObjectSupervisors{}
			for _, supervisorID := range data.Supervisors {
				objectSupervisors = append(objectSupervisors, model.ObjectSupervisors{
					ObjectID:           object.ID,
					SupervisorWorkerID: supervisorID,
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
					ObjectID: object.ID,
					TeamID:   teamID,
				})
			}

			if err := tx.CreateInBatches(&objectTeams, 5).Error; err != nil {
				return err
			}
		}

		if len(data.NourashedByTPObjectID) != 0 {
			tpNourashesObjects := []model.TPNourashesObjects{}
			for _, tpObjectID := range data.NourashedByTPObjectID {
				tpNourashesObjects = append(tpNourashesObjects, model.TPNourashesObjects{
					TP_ObjectID: tpObjectID,
					TargetID:    object.ID,
					TargetType:  "kl04kv_objects",
				})

				if err := tx.CreateInBatches(&tpNourashesObjects, 5).Error; err != nil {
					return err
				}
			}
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

		if err := tx.Delete(&model.TPNourashesObjects{}, "target_id = ? AND target_type = 'kl04kv_objects'", object.ID).Error; err != nil {
			return err
		}

		if len(data.NourashedByTPObjectID) != 0 {
			tpNourashesObjects := []model.TPNourashesObjects{}
			for _, tpObjectID := range data.NourashedByTPObjectID {
				tpNourashesObjects = append(tpNourashesObjects, model.TPNourashesObjects{
					TP_ObjectID: tpObjectID,
					TargetID:    object.ID,
					TargetType:  "kl04kv_objects",
				})

				if err := tx.CreateInBatches(&tpNourashesObjects, 5).Error; err != nil {
					return err
				}
			}

		}

		return nil
	})

	return mjd, err
}

func (repo *mjdObjectRepository) Delete(id, projectID uint) error {
	return repo.db.Transaction(func(tx *gorm.DB) error {
		err := tx.Exec(`
      DELETE FROM object_supervisors
      WHERE object_supervisors.object_id = (
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

		err = tx.Exec(`
      DELETE FROM object_teams
      WHERE object_teams.object_id = (
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

		err = tx.Exec(`
      DELETE FROM tp_nourashes_objects
      WHERE tp_nourashes_objects.target_id = (
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

func (repo *mjdObjectRepository) CreateInBatches(objects []model.Object, mjds []model.MJD_Object, supervisors []uint) ([]model.MJD_Object, error) {
	err := repo.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.CreateInBatches(&mjds, 10).Error; err != nil {
			return err
		}

		for index := range objects {
			objects[index].ObjectDetailedID = mjds[index].ID
		}

		if err := tx.CreateInBatches(&objects, 10).Error; err != nil {
			return err
		}

		return nil
	})

	return mjds, err
}

func (repo *mjdObjectRepository) Import(data []dto.MJDObjectImportData) error {
	return repo.db.Transaction(func(tx *gorm.DB) error {
		for index, row := range data {
			mjd := row.MJD
			if err := tx.Create(&mjd).Error; err != nil {
				return err
			}

			object := row.Object
			object.ObjectDetailedID = mjd.ID
			data[index].Object.ObjectDetailedID = mjd.ID
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

			if row.NourashedByTP.TP_ObjectID != 0 {
				data[index].NourashedByTP.TargetID = object.ID
				if err := tx.Create(&data[index].NourashedByTP).Error; err != nil {
					return err
				}
			}

		}

		return nil
	})
}

func (repo *mjdObjectRepository) GetObjectNamesForSearch(projectID uint) ([]dto.DataForSelect[string], error) {
  data := []dto.DataForSelect[string]{}
  err := repo.db.Raw(`
    SELECT 
      objects.name as "label",
      objects.name as "value"
    FROM objects
    INNER JOIN mjd_objects ON mjd_objects.id = objects.object_detailed_id
    WHERE
      objects.project_id = ? AND
      objects.type = 'mjd_objects'
    `, projectID).Scan(&data).Error

  return data, err
}
