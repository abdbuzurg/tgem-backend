package service

import (
	"backend-v2/internal/dto"
	"backend-v2/internal/repository"
	"backend-v2/model"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

type userActionService struct {
	userActionRepo repository.IUserActionRepository
	userRepo       repository.IUserRepository
}

func InitUserActionService(
	userActionRepo repository.IUserActionRepository,
	userRepo repository.IUserRepository,
) IUserActionService {
	return &userActionService{
		userActionRepo: userActionRepo,
		userRepo:       userRepo,
	}
}

type IUserActionService interface {
	GetAllByUserID(userID uint) ([]dto.UserActionView, error)
	Create(data model.UserAction) 
}

func (service *userActionService) GetAllByUserID(userID uint) ([]dto.UserActionView, error) {

	data, err := service.userActionRepo.GetAllByUserID(userID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return []dto.UserActionView{}, err
	}

	result := []dto.UserActionView{}
	for _, userAction := range data {
		result = append(result, dto.UserActionView{
			ID:           userAction.ID,
			ActionType:   userAction.ActionType,
			ActionID:     userAction.ActionID,
			ActionStatus: userAction.ActionStatus,
      ActionStatusMessage: userAction.ActionStatusMessage,
			ActionURL:    userAction.ActionURL,
			DateOfAction: userAction.DateOfAction,
		})
	}

	return result, nil
}

func (service *userActionService) Create(data model.UserAction) {
  
  action, err := service.userActionRepo.Create(data)
  if err != nil {
    fmt.Printf("could not save user action - %v \n with error - %v", action, err)
  }

}
