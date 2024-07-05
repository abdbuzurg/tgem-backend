package repository

import (
	"backend-v2/model"

	"gorm.io/gorm"
)

type projectRepository struct {
	db *gorm.DB
}

func InitProjectRepository(db *gorm.DB) IProjectRepository {
	return &projectRepository{
		db: db,
	}
}

type IProjectRepository interface {
	GetAll() ([]model.Project, error)
	GetPaginated(page, limit int) ([]model.Project, error)
	GetPaginatedFiltered(page, limit int, filter model.Project) ([]model.Project, error)
	GetByID(id uint) (model.Project, error)
	Create(data model.Project) (model.Project, error)
	Update(data model.Project) (model.Project, error)
	Delete(id uint) error
	Count() (int64, error)
}

func (repo *projectRepository) GetAll() ([]model.Project, error) {
	data := []model.Project{}
	err := repo.db.Order("id desc").Find(&data).Error
	return data, err
}

func (repo *projectRepository) GetPaginated(page, limit int) ([]model.Project, error) {
	data := []model.Project{}
	err := repo.db.Order("id desc").Offset((page - 1) * limit).Limit(limit).Find(&data).Error
	return data, err
}

func (repo *projectRepository) GetPaginatedFiltered(page, limit int, filter model.Project) ([]model.Project, error) {
	data := []model.Project{}
	err := repo.db.
		Raw(`SELECT * FROM projects WHERE
			(nullif(?, '') IS NULL OR name = ?) AND
			(nullif(?, '') IS NULL OR client = ?) ORDER BY id DESC LIMIT ? OFFSET ?`,
			filter.Name, filter.Name,
			filter.Client, filter.Client,
			limit, (page-1)*limit,
		).
		Scan(&data).Error

	return data, err
}

func (repo *projectRepository) GetByID(id uint) (model.Project, error) {
	data := model.Project{}
	err := repo.db.Find(&data, "id = ?", id).Error
	return data, err
}

func (repo *projectRepository) Create(data model.Project) (model.Project, error) {

	err := repo.db.Transaction(func(tx *gorm.DB) error {
		err := repo.db.Create(&data).Error
    if err != nil {
      return err
    }

    err = repo.db.Create(&model.UserInProject{
      UserID: 1,
      ProjectID: data.ID,
    }).Error
    if err != nil {
      return err
    }

    return nil
	})
	return data, err
}

func (repo *projectRepository) Update(data model.Project) (model.Project, error) {
	err := repo.db.Model(&model.Project{}).Select("*").Updates(&data).Error
	return data, err
}

func (repo *projectRepository) Delete(id uint) error {
	return repo.db.Delete(&model.Project{}, "id = ?", id).Error
}

func (repo *projectRepository) Count() (int64, error) {
	var count int64
	err := repo.db.Model(&model.Project{}).Count(&count).Error
	return count, err
}
