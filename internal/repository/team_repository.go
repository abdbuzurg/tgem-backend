package repository

import (
	"backend-v2/internal/dto"
	"backend-v2/model"

	"gorm.io/gorm"
)

type teamRepository struct {
	db *gorm.DB
}

func InitTeamRepostory(db *gorm.DB) ITeamRepository {
	return &teamRepository{
		db: db,
	}
}

type ITeamRepository interface {
	GetAll(projectID uint) ([]model.Team, error)
	GetPaginated(page, limit int, projectID uint) ([]dto.TeamPaginatedQuery, error)
	GetByID(id uint) (model.Team, error)
	GetByRangeOfIDs(ids []uint) ([]model.Team, error)
	GetByNumber(number string) (model.Team, error)
	Create(data dto.TeamMutation) (model.Team, error)
	CreateInBatches(data []dto.TeamMutation) ([]model.Team, error)
	Update(data dto.TeamMutation) (model.Team, error)
	Delete(id uint) error
	Count(projectID uint) (int64, error)
	GetTeamNumberAndTeamLeadersByID(projectID, id uint) ([]dto.TeamNumberAndTeamLeaderNameQueryResult, error)
	DoesTeamNumberAlreadyExistForCreate(teamNumber string, projectID uint) (bool, error)
	DoesTeamNumberAlreadyExistForUpdate(teamNumber string, id uint, projectID uint) (bool, error)
	GetAllForSelect(projectID uint) ([]dto.TeamDataForSelect, error)
}

func (repo *teamRepository) GetAll(projectID uint) ([]model.Team, error) {
	data := []model.Team{}
	err := repo.db.Order("id desc").Find(&data, "project_id = ?", projectID).Error
	return data, err
}

func (repo *teamRepository) GetPaginated(page, limit int, projectID uint) ([]dto.TeamPaginatedQuery, error) {
	data := []dto.TeamPaginatedQuery{}
	err := repo.db.
		Raw(`SELECT 
          teams.id as id,
          teams.number AS team_number, 
          workers.id AS leader_id,
          workers.name AS leader_name,
          teams.mobile_number AS team_mobile_number,
          teams.company AS team_company
        FROM teams
        INNER JOIN team_leaders ON team_leaders.team_id = teams.id
        INNER JOIN workers ON team_leaders.leader_worker_id = workers.id
        WHERE
          teams.project_id = ?
        ORDER BY teams.id DESC LIMIT ? OFFSET ?`,
			projectID, limit, (page-1)*limit,
		).
		Scan(&data).Error

	return data, err
}

func (repo *teamRepository) GetByID(id uint) (model.Team, error) {
	data := model.Team{}
	err := repo.db.Find(&data, "id = ?", id).Error
	return data, err
}

func (repo *teamRepository) Create(data dto.TeamMutation) (model.Team, error) {
	team := model.Team{
		ProjectID:    data.ProjectID,
		Number:       data.Number,
		Company:      data.Company,
		MobileNumber: data.MobileNumber,
	}

	err := repo.db.Transaction(func(tx *gorm.DB) error {

		if err := tx.Create(&team).Error; err != nil {
			return err
		}

		teamLeaders := []model.TeamLeaders{}
		for _, teamLeaderID := range data.LeaderWorkerIDs {
			teamLeaders = append(teamLeaders, model.TeamLeaders{
				TeamID:         team.ID,
				LeaderWorkerID: teamLeaderID,
			})
		}

		if err := tx.CreateInBatches(&teamLeaders, 5).Error; err != nil {
			return err
		}

		return nil
	})

	return team, err
}

func (repo *teamRepository) Update(data dto.TeamMutation) (model.Team, error) {
	team := model.Team{
		ID:           data.ID,
		Number:       data.Number,
		MobileNumber: data.MobileNumber,
		Company:      data.Company,
	}

	err := repo.db.Transaction(func(tx *gorm.DB) error {

		if err := tx.Model(&model.Team{}).Where("id = ?", team.ID).Updates(&team).Error; err != nil {
			return err
		}

		if err := tx.Delete(&model.TeamLeaders{}, "team_id = ?", team.ID).Error; err != nil {
			return err
		}

		teamLeaders := []model.TeamLeaders{}
		for _, teamLeaderID := range data.LeaderWorkerIDs {
			teamLeaders = append(teamLeaders, model.TeamLeaders{
				TeamID:         team.ID,
				LeaderWorkerID: teamLeaderID,
			})
		}

		if err := tx.CreateInBatches(&teamLeaders, 5).Error; err != nil {
			return err
		}

		return nil
	})

	return team, err
}

func (repo *teamRepository) Delete(id uint) error {
	return repo.db.Transaction(func(tx *gorm.DB) error {

		if err := tx.Delete(&model.TeamLeaders{}, "team_id = ?", id).Error; err != nil {
			return err
		}

		if err := tx.Delete(&model.Team{}, "id = ?", id).Error; err != nil {
			return err
		}

		return nil

	})
}

func (repo *teamRepository) Count(projectID uint) (int64, error) {
	var count int64
	err := repo.db.Model(&model.Team{}).Where("project_id = ?", projectID).Count(&count).Error
	return count, err
}

func (repo *teamRepository) GetByNumber(number string) (model.Team, error) {
	data := model.Team{}
	err := repo.db.
		Raw(`SELECT * FROM teams WHERE number = ?`, number).
    Scan(&data).
		Error
	return data, err
}

func (repo *teamRepository) GetByRangeOfIDs(ids []uint) ([]model.Team, error) {
	var data []model.Team
	err := repo.db.Model(model.Team{}).Select("*").Where("id IN ?", ids).Scan(&data).Error
	return data, err
}

func (repo *teamRepository) CreateInBatches(data []dto.TeamMutation) ([]model.Team, error) {
	teams := []model.Team{}

	err := repo.db.Transaction(func(tx *gorm.DB) error {

		for _, oneEntry := range data {
			teams = append(teams, model.Team{
				Number:       oneEntry.Number,
				MobileNumber: oneEntry.MobileNumber,
				Company:      oneEntry.Company,
				ProjectID:    oneEntry.ProjectID,
			})
		}

		if err := tx.CreateInBatches(&teams, 5).Error; err != nil {
			return err
		}

		teamLeaders := []model.TeamLeaders{}
		for index, oneEntry := range data {

			for _, learderWorkerID := range oneEntry.LeaderWorkerIDs {
				teamLeaders = append(teamLeaders, model.TeamLeaders{
					TeamID:         teams[index].ID,
					LeaderWorkerID: learderWorkerID,
				})
			}
		}

		if err := tx.CreateInBatches(&teamLeaders, 10).Error; err != nil {
			return err
		}

		return nil
	})

	return teams, err
}

func (repo *teamRepository) GetTeamNumberAndTeamLeadersByID(projectID, id uint) ([]dto.TeamNumberAndTeamLeaderNameQueryResult, error) {
	data := []dto.TeamNumberAndTeamLeaderNameQueryResult{}
	err := repo.db.Raw(`
    SELECT 
      teams.number as team_number,
      workers.name as team_leader_name
    FROM teams
    INNER JOIN team_leaders ON team_leaders.team_id = teams.id
    INNER JOIN workers ON workers.id = team_leaders.leader_worker_id
    WHERE
      teams.project_id = ? AND
      teams.id = ?
	  `, projectID, id).Scan(&data).Error

	return data, err
}

func (repo *teamRepository) DoesTeamNumberAlreadyExistForCreate(teamNumber string, projectID uint) (bool, error) {
	result := false
	err := repo.db.Raw(`
    SELECT true
    FROM teams
    WHERE teams.number = ? AND teams.project_id = ?;
    `, teamNumber, projectID).Scan(&result).Error

	return result, err
}

func (repo *teamRepository) DoesTeamNumberAlreadyExistForUpdate(teamNumber string, id uint, projectID uint) (bool, error) {
	result := false
	err := repo.db.Raw(`
    SELECT true
    FROM teams
    WHERE teams.number = ? AND teams.id <> ? AND teams.project_id = ?;
    `, teamNumber, id, projectID).Scan(&result).Error

	return result, err
}

func (repo *teamRepository) GetAllForSelect(projectID uint) ([]dto.TeamDataForSelect, error) {
	result := []dto.TeamDataForSelect{}
	err := repo.db.Raw(`
    SELECT 
      teams.id as id,
      teams.number as team_number,
      workers.name as team_leader_name
    FROM teams 
    INNER JOIN team_leaders ON team_leaders.team_id = teams.id
    INNER JOIN workers ON workers.id = team_leaders.leader_worker_id
    WHERE teams.project_id = ?
    `, projectID).Scan(&result).Error

	return result, err
}
