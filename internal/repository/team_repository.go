package repository

import (
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
	GetAll() ([]model.Team, error)
	GetPaginated(page, limit int) ([]model.Team, error)
	GetPaginatedFiltered(page, limit int, filter model.Team) ([]model.Team, error)
	GetByID(id uint) (model.Team, error)
  GetByRangeOfIDs(ids []uint) ([]model.Team, error)
  GetByNumber(number string) (model.Team, error)
	Create(data model.Team) (model.Team, error)
	Update(data model.Team) (model.Team, error)
	Delete(id uint) error
	Count() (int64, error)
}

func (repo *teamRepository) GetAll() ([]model.Team, error) {
	data := []model.Team{}
	err := repo.db.Order("id desc").Find(&data).Error
	return data, err
}

func (repo *teamRepository) GetPaginated(page, limit int) ([]model.Team, error) {
	data := []model.Team{}
	err := repo.db.Order("id desc").Offset((page - 1) * limit).Limit(limit).Find(&data).Error
	return data, err
}

func (repo *teamRepository) GetPaginatedFiltered(page, limit int, filter model.Team) ([]model.Team, error) {
	data := []model.Team{}
	err := repo.db.
		Raw(`SELECT * FROM teams WHERE
			(nullif(?, '') IS NULL OR leader_worker_id = ?) AND
			(nullif(?, '') IS NULL OR number = ?) AND
			(nullif(?, '') IS NULL OR mobile_number = ?) AND
			(nullif(?, '') IS NULL OR company = ?) ORDER BY id DESC LIMIT ? OFFSET ?`,
			filter.LeaderWorkerID, filter.LeaderWorkerID, filter.Number, filter.Number, filter.MobileNumber, filter.MobileNumber, filter.Company, filter.Company, limit, (page-1)*limit,
		).
		Scan(&data).Error

	return data, err
}

func (repo *teamRepository) GetByID(id uint) (model.Team, error) {
	data := model.Team{}
	err := repo.db.Find(&data, "id = ?", id).Error
	return data, err
}

func (repo *teamRepository) Create(data model.Team) (model.Team, error) {
	err := repo.db.Create(&data).Error
	return data, err
}

func (repo *teamRepository) Update(data model.Team) (model.Team, error) {
	err := repo.db.Model(&model.Team{}).Select("*").Where("id = ?", data.ID).Updates(&data).Error
	return data, err
}

func (repo *teamRepository) Delete(id uint) error {
	return repo.db.Delete(&model.Team{}, "id = ?", id).Error
}

func (repo *teamRepository) Count() (int64, error) {
	var count int64
	err := repo.db.Model(&model.Team{}).Count(&count).Error
	return count, err
}

func (repo *teamRepository)   GetByNumber(number string) (model.Team, error) {
  data := model.Team{}
  err := repo.db.
    Raw(`SELECT * FROM teams WHERE number = ?`, number).
    Error
  return data, err
}

func (repo *teamRepository) GetByRangeOfIDs(ids []uint) ([]model.Team, error) {
  var data []model.Team
  err := repo.db.Model(model.Team{}).Select("*").Where("id IN ?", ids).Scan(&data).Error
  return data, err
}

