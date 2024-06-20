package repository

import "gorm.io/gorm"

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
