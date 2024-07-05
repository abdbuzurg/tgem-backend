package service

import (
	"backend-v2/internal/dto"
	"backend-v2/internal/repository"
	"backend-v2/model"
	"backend-v2/pkg/jwt"
	"backend-v2/pkg/security"
	"backend-v2/pkg/utils"
	"fmt"
  "errors"

	"gorm.io/gorm"
)

type userService struct {
	userRepo repository.IUserRepository
  userInProjectRepo repository.IUserInProjectRepository
  workerRepo repository.IWorkerRepository
  roleRepo repository.IRoleRepository
  userInProjects repository.IUserInProjectRepository
  projectRepo repository.IProjectRepository
}

func InitUserService(
  userRepo repository.IUserRepository,
  userInProjectRepo repository.IUserInProjectRepository,
  workerRepo repository.IWorkerRepository,
  roleRepo repository.IRoleRepository,
  userInProjects repository.IUserInProjectRepository,
  projectRepo repository.IProjectRepository,
) IUserService {
	return &userService{
		userRepo: userRepo,
    userInProjectRepo: userInProjectRepo,
    workerRepo: workerRepo,
    roleRepo: roleRepo,
    userInProjects: userInProjectRepo,
    projectRepo: projectRepo,
	}
}

type IUserService interface {
	GetAll() ([]model.User, error)
	GetPaginated(page, limit int, data model.User) ([]dto.UserPaginated, error)
	GetByID(id uint) (model.User, error)
	Create(data dto.NewUserData) error
	Update(data model.User) (model.User, error)
	Delete(id uint) error
	Count() (int64, error)
	Login(data dto.LoginData) (dto.LoginResponse, error)
}

func (service *userService) GetAll() ([]model.User, error) {
	return service.userRepo.GetAll()
}

func (service *userService) GetPaginated(page, limit int, data model.User) ([]dto.UserPaginated, error) {
  var userData []model.User
  var err error
	if !(utils.IsEmptyFields(data)) {
		userData, err = service.userRepo.GetPaginatedFiltered(page, limit, data)
	} else {
    userData, err = service.userRepo.GetPaginated(page, limit)
  }

  if err != nil {
    return []dto.UserPaginated{}, err
  }

  var result []dto.UserPaginated
  for _, user := range userData {
    worker, err := service.workerRepo.GetByID(user.WorkerID)
    if err != nil {
      return []dto.UserPaginated{}, err
    }

    role, err := service.roleRepo.GetByID(user.RoleID)
    if err != nil {
      return []dto.UserPaginated{}, err
    }

    result = append(result, dto.UserPaginated{
      Username: user.Username,
      WorkerName: worker.Name,
      WorkerJobTitle: worker.JobTitleInProject,
      WorkerMobileNumber: worker.MobileNumber,
      RoleName: role.Name,
    })
  }	

  return result, nil
}

func (service *userService) GetByID(id uint) (model.User, error) {
	return service.userRepo.GetByID(id)
}

func (service *userService) Create(data dto.NewUserData) error {
  hashedPassword, err := security.Hash(data.UserData.Password)
  if err != nil {
    return  err
  }

  data.UserData.Password = string(hashedPassword)

  user, err := service.userRepo.Create(data.UserData)
  if err != nil {
    return err
  }
  
  err = service.userInProjectRepo.AddUserToProjects(user.ID, data.Projects)
  if err != nil {
    return err
  }

  return nil
}

func (service *userService) Update(data model.User) (model.User, error) {
	return service.userRepo.Update(data)
}

func (service *userService) Delete(id uint) error {
	return service.userRepo.Delete(id)
}

func (service *userService) Count() (int64, error) {
	return service.userRepo.Count()
}

func (service *userService) Login(data dto.LoginData) (dto.LoginResponse, error) {
	user, err := service.userRepo.GetByUsername(data.Username)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return dto.LoginResponse{}, fmt.Errorf("Неправильное имя пользователя")
	}
	if err != nil {
		return dto.LoginResponse{}, err
	}

	err = security.VerifyPassword(user.Password, data.Password)
	if err != nil {
		return dto.LoginResponse{}, fmt.Errorf("Неправильный пароль")
	}

  userInProjects, err := service.userInProjectRepo.GetByUserID(user.ID)
  if err != nil {
    return dto.LoginResponse{}, fmt.Errorf("У вас нету доступа в выбранный проект")
  }
  
  access := false
  for _, userInProject := range userInProjects {
    if userInProject.ProjectID == data.ProjectID {
      access = true
      break
    }
  }

  if (!access) {
    return dto.LoginResponse{}, fmt.Errorf("У вас нету доступа в выбранный проект")
  }

  result := dto.LoginResponse{
    Admin: false,
  }

	token, err := jwt.CreateToken(user.ID, user.WorkerID, user.RoleID, data.ProjectID)
	if err != nil {
		return dto.LoginResponse{}, err
	}

  result.Token = token

  project, err := service.projectRepo.GetByID(data.ProjectID)
  if err != nil {
    return dto.LoginResponse{}, err
  }

  if project.Name == "Администрирование" {
    result.Admin = true
  }

	return result, nil
}
