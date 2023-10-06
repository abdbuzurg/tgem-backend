package controller

import (
	"backend-v2/internal/service"
	"backend-v2/model"
	"backend-v2/pkg/response"
	"fmt"
	"net/url"
	"strconv"

	"github.com/gin-gonic/gin"
)

type userController struct {
	userService service.IUserService
}

func InitUserController(userService service.IUserService) IUserController {
	return &userController{
		userService: userService,
	}
}

type IUserController interface {
	GetAll(c *gin.Context)
	GetPaginated(c *gin.Context)
	GetByID(c *gin.Context)
	Create(c *gin.Context)
	Update(c *gin.Context)
	Delete(c *gin.Context)
}

func (controller *userController) GetAll(c *gin.Context) {
	data, err := controller.userService.GetAll()
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not get User data: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (controller *userController) GetPaginated(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	page, err := strconv.Atoi(pageStr)
	if err != nil || page <= 0 {
		response.ResponseError(c, fmt.Sprintf("Wrong query parameter provided for page: %v", err))
		return
	}

	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		response.ResponseError(c, fmt.Sprintf("Wrong query parameter provided for limit: %v", err))
		return
	}

	workerIDStr := c.DefaultQuery("workerID", "")
	workerID := 0
	if workerIDStr != "" {
		workerID, err = strconv.Atoi(workerIDStr)
		if err != nil {
			response.ResponseError(c, fmt.Sprintf("Cannot decode workerID parameter: %v", err))
			return
		}
	}

	username := c.DefaultQuery("username", "")
	username, err = url.QueryUnescape(username)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Wrong query parameter provided for username: %v", err))
		return
	}

	filter := model.User{
		WorkerID: uint(workerID),
		Username: username,
	}

	data, err := controller.userService.GetPaginated(page, limit, filter)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not get the paginated data of User: %v", err))
		return
	}

	dataCount, err := controller.userService.Count()
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not get the total amount of User: %v", err))
		return
	}

	response.ResponsePaginatedData(c, data, dataCount)
}

func (controller *userController) GetByID(c *gin.Context) {
	idRaw := c.Param("id")
	id, err := strconv.ParseUint(idRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Incorrect parameter provided: %v", err))
		return
	}

	data, err := controller.userService.GetByID(uint(id))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not get the data with ID(%d): %v", id, err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (controller *userController) Create(c *gin.Context) {
	var createData model.User
	if err := c.ShouldBindJSON(&createData); err != nil {
		response.ResponseError(c, fmt.Sprintf("Invalid data recieved by server: %v", err))
		return
	}

	data, err := controller.userService.Create(createData)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could perform the creation of User: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (controller *userController) Update(c *gin.Context) {
	var updateData model.User
	if err := c.ShouldBindJSON(&updateData); err != nil {
		response.ResponseError(c, fmt.Sprintf("Invalid data recieved by server: %v", err))
		return
	}

	data, err := controller.userService.Update(updateData)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not perform the updation of User: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (controller *userController) Delete(c *gin.Context) {
	idRaw := c.Param("id")
	id, err := strconv.ParseUint(idRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Incorrect parameter provided: %v", err))
		return
	}

	err = controller.userService.Delete(uint(id))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not perform the deletion of User: %v", err))
		return
	}

	response.ResponseSuccess(c, "deleted")
}
