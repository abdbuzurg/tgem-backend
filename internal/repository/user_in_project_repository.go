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
}

func(repo *userInProjectRepository) GetByUserID(userID uint) ([]model.UserInProject, error) {
  var data []model.UserInProject
  err := repo.db.Find(&data, "user_id = ?", userID).Error
  return data, err
}

func(repo *userInProjectRepository) AddUserToProjects(userID uint, projectIDs []uint) error {
  var data []model.UserInProject
  for _, projectID := range projectIDs {
    data = append(data, model.UserInProject{
      ProjectID: uint(projectID),
      UserID: userID,
    })
  }

  err := repo.db.CreateInBatches(data, 10).Error
  return err
}
