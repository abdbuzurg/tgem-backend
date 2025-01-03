package repository

import (
	"backend-v2/internal/dto"
	"backend-v2/model"

	"gorm.io/gorm"
)

type sipObjectRepository struct {
	db *gorm.DB
}

func InitSIPObjectRepository(db *gorm.DB) ISIPObjectRepository {
	return &sipObjectRepository{
		db: db,
	}
}

type ISIPObjectRepository interface {
	GetPaginated(page, limit int, filter dto.SIPObjectSearchParameters) ([]dto.SIPObjectPaginatedQuery, error)
	Count(filter dto.SIPObjectSearchParameters) (int64, error)
	Create(data dto.SIPObjectCreate) (model.SIP_Object, error)
	Update(data dto.SIPObjectCreate) (model.SIP_Object, error)
	Delete(id, projectID uint) error
	CreateInBatches(objects []model.Object, sips []model.SIP_Object, supervisors []uint) ([]model.SIP_Object, error)
	Import(data []dto.SIPObjectImportData) error
	GetObjectNamesForSearch(projectID uint) ([]dto.DataForSelect[string], error)
}

func (repo *sipObjectRepository) GetPaginated(page, limit int, filter dto.SIPObjectSearchParameters) ([]dto.SIPObjectPaginatedQuery, error) {
	data := []dto.SIPObjectPaginatedQuery{}
	err := repo.db.Raw(`
    SELECT DISTINCT 
      objects.id as object_id,
      s_ip_objects.id as object_detailed_id,
      objects.name as name,
      objects.status as status,
      s_ip_objects.amount_feeders
    FROM objects
    INNER JOIN s_ip_objects ON objects.object_detailed_id = s_ip_objects.id
    FULL JOIN object_teams ON object_teams.object_id = objects.id
    FULL JOIN object_supervisors ON object_supervisors.object_id = objects.id
    FULL JOIN tp_nourashes_objects ON tp_nourashes_objects.target_id = objects.id
    WHERE
      objects.type = 'sip_objects' AND
      objects.project_id = ? AND
      (nullif(?, '') IS NULL OR objects.name = ?) AND
      (nullif(?, 0) IS NULL OR object_teams.team_id = ?) AND
      (nullif(?, 0) IS NULL OR object_supervisors.supervisor_worker_id = ?)
    ORDER BY s_ip_objects.id DESC 
    LIMIT ? 
    OFFSET ?;
    `, filter.ProjectID,
		filter.ObjectName, filter.ObjectName,
		filter.TeamID, filter.TeamID,
		filter.SupervisorWorkerID, filter.SupervisorWorkerID,
		limit, (page-1)*limit,
	).Scan(&data).Error

	return data, err
}

func (repo *sipObjectRepository) Count(filter dto.SIPObjectSearchParameters) (int64, error) {
	var count int64
	err := repo.db.Raw(`
    SELECT DISTINCT COUNT(*)
    FROM objects
    FULL JOIN object_teams ON object_teams.object_id = objects.id
    FULL JOIN object_supervisors ON object_supervisors.object_id = objects.id
    FULL JOIN tp_nourashes_objects ON tp_nourashes_objects.target_id = objects.id
    WHERE
      objects.type = 'sip_objects' AND
      objects.project_id = ? AND
      (nullif(?, '') IS NULL OR objects.name = ?) AND
      (nullif(?, 0) IS NULL OR object_teams.team_id = ?) AND
      (nullif(?, 0) IS NULL OR object_supervisors.supervisor_worker_id = ?)
    `,
		filter.ProjectID,
		filter.ObjectName, filter.ObjectName,
		filter.TeamID, filter.TeamID,
		filter.SupervisorWorkerID, filter.SupervisorWorkerID,
	).Scan(&count).Error
	return count, err
}

func (repo *sipObjectRepository) Create(data dto.SIPObjectCreate) (model.SIP_Object, error) {
	sip := model.SIP_Object{
		AmountFeeders: data.DetailedInfo.AmountFeeders,
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
			Type:             "sip_objects",
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
					TargetType:  "sip_objects",
				})

				if err := tx.CreateInBatches(&tpNourashesObjects, 5).Error; err != nil {
					return err
				}
			}
		}

		return nil
	})

	return sip, err

}

func (repo *sipObjectRepository) Update(data dto.SIPObjectCreate) (model.SIP_Object, error) {
	sip := model.SIP_Object{
		ID:            data.BaseInfo.ObjectDetailedID,
		AmountFeeders: data.DetailedInfo.AmountFeeders,
	}

	err := repo.db.Transaction(func(tx *gorm.DB) error {

		if err := tx.Model(&model.SIP_Object{}).Where("id = ?", sip.ID).Updates(&sip).Error; err != nil {
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

	return sip, err
}

func (repo *sipObjectRepository) Delete(id, projectID uint) error {
	return repo.db.Transaction(func(tx *gorm.DB) error {
		err := tx.Exec(`
      DELETE FROM object_supervisors
      WHERE object_supervisors.object_id = (
        SELECT DISTINCT(objects.id)
        FROM objects
          INNER JOIN s_ip_objects ON s_ip_objects.id = objects.object_detailed_id
        WHERE
          objects.project_id = ? AND
          s_ip_objects.id = ? AND
          objects.type = 'sip_objects'
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
          INNER JOIN s_ip_objects ON s_ip_objects.id = objects.object_detailed_id
        WHERE
          objects.project_id = ? AND
          s_ip_objects.id = ? AND
          objects.type = 'sip_objects'
      );
    `, projectID, id).Error

		if err != nil {
			return err
		}

		if err := tx.Table("s_ip_objects").Delete(&model.SIP_Object{}, "id = ?", id).Error; err != nil {
			return err
		}

		if err := tx.Table("objects").Delete(&model.Object{}, "object_detailed_id = ? AND type = 'sip_objects'", id).Error; err != nil {
			return err
		}

		return nil
	})
}

func (repo *sipObjectRepository) CreateInBatches(objects []model.Object, sips []model.SIP_Object, supervisors []uint) ([]model.SIP_Object, error) {
	err := repo.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.CreateInBatches(&sips, 10).Error; err != nil {
			return err
		}

		for index := range objects {
			objects[index].ObjectDetailedID = sips[index].ID
		}

		if err := tx.CreateInBatches(&objects, 10).Error; err != nil {
			return err
		}

		return nil
	})

	return sips, err
}

func (repo *sipObjectRepository) Import(data []dto.SIPObjectImportData) error {
	return repo.db.Transaction(func(tx *gorm.DB) error {
		for index, row := range data {
			sip := row.SIP
			if err := tx.Create(&sip).Error; err != nil {
				return err
			}

			object := row.Object
			object.ObjectDetailedID = sip.ID
			data[index].Object.ObjectDetailedID = sip.ID
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

func (repo *sipObjectRepository) GetObjectNamesForSearch(projectID uint) ([]dto.DataForSelect[string], error) {
	data := []dto.DataForSelect[string]{}
	err := repo.db.Raw(`
    SELECT 
      objects.name as "label",
      objects.name as "value"
    FROM objects
    INNER JOIN s_ip_objects ON s_ip_objects.id = objects.object_detailed_id
    WHERE
      objects.project_id = ? AND
      objects.type = 'sip_objects'
    `, projectID).Scan(&data).Error

	return data, err
}
