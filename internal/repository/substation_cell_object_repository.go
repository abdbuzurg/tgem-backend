package repository

import (
	"backend-v2/internal/dto"
	"backend-v2/model"

	"gorm.io/gorm"
)

type substationCellObjectRepository struct {
	db *gorm.DB
}

func NewSubstationCellObjectRepository(db *gorm.DB) ISubstationCellObjectRepository {
	return &substationCellObjectRepository{
		db: db,
	}
}

type ISubstationCellObjectRepository interface {
	GetPaginated(int, int, dto.SubstationCellObjectSearchParameters) ([]dto.SubstationCellObjectPaginatedQuery, error)
	Count(dto.SubstationCellObjectSearchParameters) (int64, error)
	Create(dto.SubstationCellObjectCreate) (model.SubstationCellObject, error)
	Update(dto.SubstationCellObjectCreate) (model.SubstationCellObject, error)
	Delete(id, projectID uint) error
	CreateInBatches(data []dto.SubstationCellObjectImportData) error
	GetSubstationName(substationCellObjectID uint) (string, error)
	GetObjectNamesForSearch(projectID uint) ([]dto.DataForSelect[string], error)
}

func (repo *substationCellObjectRepository) GetPaginated(page, limit int, filter dto.SubstationCellObjectSearchParameters) ([]dto.SubstationCellObjectPaginatedQuery, error) {
	data := []dto.SubstationCellObjectPaginatedQuery{}
	err := repo.db.Raw(`
    SELECT DISTINCT
      objects.id as object_id,
      substation_cell_objects.id as object_detailed_id,
      objects.name as name,
      objects.status as status
    FROM objects
    INNER JOIN substation_cell_objects ON objects.object_detailed_id = substation_cell_objects.id
    FULL JOIN object_teams ON object_teams.object_id = objects.id
    FULL JOIN object_supervisors ON object_supervisors.object_id = objects.id
    FULL JOIN substation_cell_nourashes_substation_objects ON substation_cell_nourashes_substation_objects.substation_cell_object_id = objects.id
    WHERE
      objects.type = 'substation_cell_objects' AND
      objects.project_id = ? AND
      (nullif(?, '') IS NULL OR objects.name = ?) AND
      (nullif(?, 0) IS NULL OR object_teams.team_id = ?) AND
      (nullif(?, 0) IS NULL OR object_supervisors.supervisor_worker_id = ?) AND
      (nullif(?, 0) IS NULL OR substation_cell_nourashes_substation_objects.substation_cell_object_id = ?)
    ORDER BY substation_cell_objects.id DESC 
    LIMIT ? 
    OFFSET ?;
    `, filter.ProjectID,
		filter.ObjectName, filter.ObjectName,
		filter.TeamID, filter.TeamID,
		filter.SupervisorWorkerID, filter.SupervisorWorkerID,
		filter.SubstationObjectID, filter.SubstationObjectID,
		limit, (page-1)*limit).Scan(&data).Error

	return data, err
}

func (repo *substationCellObjectRepository) Count(filter dto.SubstationCellObjectSearchParameters) (int64, error) {
	var count int64
	err := repo.db.Raw(`
    SELECT DISTINCT COUNT(*)
    FROM objects
    INNER JOIN substation_cell_objects ON objects.object_detailed_id = substation_cell_objects.id
    FULL JOIN object_teams ON object_teams.object_id = objects.id
    FULL JOIN object_supervisors ON object_supervisors.object_id = objects.id
    FULL JOIN substation_cell_nourashes_substation_objects ON substation_cell_nourashes_substation_objects.substation_cell_object_id = objects.id
    WHERE
      objects.type = 'substation_cell_objects' AND
      objects.project_id = ? AND
      (nullif(?, '') IS NULL OR objects.name = ?) AND
      (nullif(?, 0) IS NULL OR object_teams.team_id = ?) AND
      (nullif(?, 0) IS NULL OR object_supervisors.supervisor_worker_id = ?) AND
      (nullif(?, 0) IS NULL OR substation_cell_nourashes_substation_objects.substation_cell_object_id = ?)
    `, filter.ProjectID,
		filter.ObjectName, filter.ObjectName,
		filter.TeamID, filter.TeamID,
    filter.SubstationObjectID, filter.SubstationObjectID,
		filter.SupervisorWorkerID, filter.SupervisorWorkerID).Scan(&count).Error
	return count, err
}

func (repo *substationCellObjectRepository) Create(data dto.SubstationCellObjectCreate) (model.SubstationCellObject, error) {
	substationCell := model.SubstationCellObject{}

	err := repo.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&substationCell).Error; err != nil {
			return err
		}

		object := model.Object{
			ID:               0,
			ObjectDetailedID: substationCell.ID,
			ProjectID:        data.BaseInfo.ProjectID,
			Name:             data.BaseInfo.Name,
			Status:           data.BaseInfo.Status,
			Type:             "substation_cell_objects",
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

		if data.SubstationObjectID != 0 {
			nourashes := model.SubstationCellNourashesSubstationObject{
				SubstationObjectID:     data.SubstationObjectID,
				SubstationCellObjectID: substationCell.ID,
			}
			if err := tx.Create(&nourashes).Error; err != nil {
				return err
			}
		}

		return nil
	})

	return substationCell, err
}

func (repo *substationCellObjectRepository) Update(data dto.SubstationCellObjectCreate) (model.SubstationCellObject, error) {
	substationCell := model.SubstationCellObject{
		ID: data.BaseInfo.ObjectDetailedID,
	}

	err := repo.db.Transaction(func(tx *gorm.DB) error {

		if err := tx.Model(&model.STVT_Object{}).Where("id = ?", substationCell.ID).Updates(&substationCell).Error; err != nil {
			return err
		}

		object := model.Object{
			ID:               data.BaseInfo.ID,
			ProjectID:        data.BaseInfo.ProjectID,
			ObjectDetailedID: data.BaseInfo.ObjectDetailedID,
			Name:             data.BaseInfo.Name,
			Type:             "substation_cell_objects",
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

		if err := tx.Delete(&model.SubstationCellNourashesSubstationObject{}, "substation_cell_object_id = ?", object.ID).Error; err != nil {
			return err
		}

		if data.SubstationObjectID != 0 {
			nourashes := model.SubstationCellNourashesSubstationObject{
				SubstationObjectID:     data.SubstationObjectID,
				SubstationCellObjectID: substationCell.ID,
			}
			if err := tx.Create(&nourashes).Error; err != nil {
				return err
			}
		}

		return nil
	})

	return substationCell, err
}

func (repo *substationCellObjectRepository) Delete(id, projectID uint) error {
	return repo.db.Transaction(func(tx *gorm.DB) error {
		err := tx.Exec(`
      DELETE FROM object_supervisors
      WHERE object_supervisors.object_id = (
        SELECT DISTINCT(objects.id)
        FROM objects
          INNER JOIN substation_cell_objects ON substation_cell_objects.id = objects.object_detailed_id
        WHERE
          objects.project_id = ? AND
          stvt_objects.id = ? AND
          objects.type = 'substation_cell_objects'
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
          INNER JOIN substation_cell_objects ON substation_cell_objects.id = objects.object_detailed_id
        WHERE
          objects.project_id = ? AND
          stvt_objects.id = ? AND
          objects.type = 'substation_cell_objects'
      );
    `, projectID, id).Error

		if err != nil {
			return err
		}

		err = tx.Exec(`
      DELETE FROM substation_cell_nourashes_substation_objects
      WHERE substation_cell_nourashes_substation_objects.substation_cell_object_id = (
        SELECT DISTINCT(objects.id)
        FROM objects
          INNER JOIN substation_cell_objects ON substation_cell_objects.id = objects.object_detailed_id
        WHERE
          objects.project_id = ? AND
          substation_cell_objects.id = ? AND
          objects.type = 'substation_cell_objects'
      );
    `, projectID, id).Error

		if err := tx.Table("substation_cell_objects").Delete(&model.SubstationCellObject{}, "id = ?", id).Error; err != nil {
			return err
		}

		if err := tx.Table("objects").Delete(&model.Object{}, "object_detailed_id = ? AND type = 'substation_cell_objects'", id).Error; err != nil {
			return err
		}

		return nil
	})
}

func (repo *substationCellObjectRepository) CreateInBatches(data []dto.SubstationCellObjectImportData) error {
	return repo.db.Transaction(func(tx *gorm.DB) error {
		for index, row := range data {
			substationCell := row.SubstationCell
			if err := tx.Create(&substationCell).Error; err != nil {
				return err
			}

			object := row.Object
			object.ObjectDetailedID = substationCell.ID
			data[index].Object.ObjectDetailedID = substationCell.ID
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

			if row.Nourashes.SubstationObjectID != 0 {
				data[index].Nourashes.SubstationCellObjectID = object.ID
				if err := tx.Create(&data[index].Nourashes).Error; err != nil {
					return err
				}
			}

		}

		return nil
	})
}

func (repo *substationCellObjectRepository) GetSubstationName(substationCellObjectID uint) (string, error) {
	result := ""
	err := repo.db.Raw(`
    SELECT substation_objects.name
    FROM substation_cell_nourashes_substation_objects
    INNER JOIN objects AS substation_cell_objects ON substation_cell_objects.id = substation_cell_nourashes_substation_objects.substation_cell_object_id 
    INNER JOIN objects AS substation_objects ON substation_objects.id = substation_cell_nourashes_substation_objects.substation_object_id
    WHERE
      substation_cell_nourashes_substation_objects.substation_cell_object_id = ?
  `, substationCellObjectID).Scan(&result).Error

	return result, err
}

func (repo *substationCellObjectRepository) GetObjectNamesForSearch(projectID uint) ([]dto.DataForSelect[string], error) {
	result := []dto.DataForSelect[string]{}
	err := repo.db.Raw(`
    SELECT 
      objects.name as "label",
      objects.name as "value"
    FROM objects
    WHERE
      objects.project_id = ? AND
      objects.type = 'substation_cell_objects'`, projectID).Scan(&result).Error
	return result, err
}
