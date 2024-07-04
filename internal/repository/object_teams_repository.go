package repository

import (
	"backend-v2/internal/dto"

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
  GetTeamsByObjectID(objectID uint) ([]dto.TeamDataForSelect, error)
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

func (repo *objectTeamsRepository) GetTeamsByObjectID(objectID uint) ([]dto.TeamDataForSelect, error) {
  data := []dto.TeamDataForSelect{}
  err := repo.db.Raw(`
    SELECT 
      teams.id as id,
      teams.number as team_number,
      workers.name as team_leader_name
    FROM object_teams
    INNER JOIN teams ON teams.id = object_teams.team_id
    INNER JOIN team_leaders ON team_leaders.team_id = object_teams.team_id
    INNER JOIN workers ON workers.id = team_leaders.leader_worker_id
    WHERE object_teams.object_id = ?;
  `, objectID).Scan(&data).Error

  return data, err
}
