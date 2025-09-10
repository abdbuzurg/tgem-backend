package repository

import (
	"backend-v2/model"

	"gorm.io/gorm"
)

type userInProjectRepository struct {
	db *gorm.DB
}

func InitUserInProjectRepository(db *gorm.DB) IUserInProjectRepository {
	return &userInProjectRepository{
		db: db,
	}
}

type IUserInProjectRepository interface {
	GetByUserID(userID uint) ([]model.UserInProject, error)
	AddUserToProjects(userID uint, projectIDs []uint) error
	GetProjectNamesByUserID(userID uint) ([]string, error)
	UpdateUserInProjectsWithGivenArray(projectIDs []uint, userID uint) error
}

func (repo *userInProjectRepository) GetByUserID(userID uint) ([]model.UserInProject, error) {
	var data []model.UserInProject
	err := repo.db.Find(&data, "user_id = ?", userID).Error
	return data, err
}

func (repo *userInProjectRepository) AddUserToProjects(userID uint, projectIDs []uint) error {
	var data []model.UserInProject
	for _, projectID := range projectIDs {
		data = append(data, model.UserInProject{
			ProjectID: uint(projectID),
			UserID:    userID,
		})
	}

	err := repo.db.CreateInBatches(data, 10).Error
	return err
}

func (repo *userInProjectRepository) GetProjectNamesByUserID(userID uint) ([]string, error) {
	result := []string{}
	err := repo.db.Raw(`
    SELECT name 
    FROM projects
    WHERE projects.id IN (
	    SELECT project_id
	    FROM user_in_projects
	    WHERE user_id = ?
    )
  `, userID).Scan(&result).Error

	return result, err
}

func (repo *userInProjectRepository) UpdateUserInProjectsWithGivenArray(projectIDs []uint, userID uint) error {
	err := repo.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(`DELETE FROM user_in_projects WHERE user_id = ?`, userID).Error; err != nil {
			return err
		}

		userInProjects := []model.UserInProject{}
		for _, projectID := range projectIDs {
			userInProjects = append(userInProjects, model.UserInProject{
				ProjectID: projectID,
				UserID:    userID,
			})
		}

		if err := tx.CreateInBatches(&userInProjects, 10).Error; err != nil {
			return err
		}

		return nil
	})

	return err
}
