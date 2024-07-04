package repository

import (
	"backend-v2/model"

	"gorm.io/gorm"
)

type objectTeamsRepository struct {
	db *gorm.DB
}

func InitObjectTeamsRepository(db *gorm.DB) IObjectTeamsRepository {
	return &objectTeamsRepository{
		db: db,
	}
}

type IObjectTeamsRepository interface {
	GetTeamsNumberByObjectID(objectID uint) ([]string, error)
  GetTeamsByObjectID(objectID uint) ([]model.Team, error)
}

func (repo *objectTeamsRepository) GetTeamsNumberByObjectID(objectID uint) ([]string, error) {
	data := []string{}
	err := repo.db.Raw(`
    SELECT 
      teams.number as team_number
    FROM object_teams
    INNER JOIN objects ON objects.id = object_teams.object_id
    INNER JOIN teams ON teams.id = object_teams.team_id
    WHERE 
      objects.id = ?
    `, objectID).Scan(&data).Error

	return data, err
}

func (repo *objectTeamsRepository) GetTeamsByObjectID(objectID uint) ([]model.Team, error) {
  data := []model.Team{}
  err := repo.db.Raw(`
     SELECT 
      teams.id as id,
      teams.number as number,
      teams.mobile_number as mobile_number,
      teams.company as company,
      teams.project_id as project_id
    FROM object_teams
    INNER JOIN teams ON teams.id = object_teams.team_id
    WHERE object_teams.object_id = ?;
  `, objectID).Scan(&data).Error

  return data, err
}
