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
	CreateInBatches(objects []model.Object, stvts []model.STVT_Object, supervisors []uint) ([]model.STVT_Object, error)
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
        stvt_objects.tt_coefficient as tt_coefficient
      FROM objects
        INNER JOIN stvt_objects ON objects.object_detailed_id = stvt_objects.id
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

	return stvt, err
}

func (repo *stvtObjectRepository) Delete(id, projectID uint) error {
	return repo.db.Transaction(func(tx *gorm.DB) error {
		err := tx.Exec(`
      DELETE FROM object_supervisors
      WHERE object_supervisors.object_id = (
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

		err = tx.Exec(`
      DELETE FROM object_teams
      WHERE object_teams.object_id = (
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

func (repo *stvtObjectRepository) CreateInBatches(objects []model.Object, stvts []model.STVT_Object, supervisors []uint) ([]model.STVT_Object, error) {
	err := repo.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.CreateInBatches(&stvts, 10).Error; err != nil {
			return err
		}

		for index := range objects {
			objects[index].ObjectDetailedID = stvts[index].ID
		}

		if err := tx.CreateInBatches(&objects, 10).Error; err != nil {
			return err
		}

		return nil
	})

	return stvts, err
}
