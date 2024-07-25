package controller

import (
	"backend-v2/internal/service"
	"backend-v2/model"
	"backend-v2/pkg/response"
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
)

type objectOperationController struct {
	objectOperationService service.IObjectOperationService
}

func InitObjectOperationController(objectOperationService service.IObjectOperationService) IObjectOperationController {
	return &objectOperationController{
		objectOperationService: objectOperationService,
	}
}

type IObjectOperationController interface {
	GetAll(c *gin.Context)
	GetPaginated(c *gin.Context)
	GetByID(c *gin.Context)
	Create(c *gin.Context)
	Update(c *gin.Context)
	Delete(c *gin.Context)
}

func (controller *objectOperationController) GetAll(c *gin.Context) {
	data, err := controller.objectOperationService.GetAll()
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not get ObjectOperation data: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (controller *objectOperationController) GetPaginated(c *gin.Context) {
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

	filter := model.ObjectOperation{}

	data, err := controller.objectOperationService.GetPaginated(page, limit, filter)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not get the paginated data of ObjectOperation: %v", err))
		return
	}

	dataCount, err := controller.objectOperationService.Count()
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not get the total amount of ObjectOperation: %v", err))
		return
	}

	response.ResponsePaginatedData(c, data, dataCount)
}

func (controller *objectOperationController) GetByID(c *gin.Context) {
	idRaw := c.Param("id")
	id, err := strconv.ParseUint(idRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Incorrect parameter provided: %v", err))
		return
	}

	data, err := controller.objectOperationService.GetByID(uint(id))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not get the data with ID(%d): %v", id, err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (controller *objectOperationController) Create(c *gin.Context) {
	var createData model.ObjectOperation
	if err := c.ShouldBindJSON(&createData); err != nil {
		response.ResponseError(c, fmt.Sprintf("Invalid data recieved by server: %v", err))
		return
	}

	data, err := controller.objectOperationService.Create(createData)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could perform the creation of ObjectOperation: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (controller *objectOperationController) Update(c *gin.Context) {
	var updateData model.ObjectOperation
	if err := c.ShouldBindJSON(&updateData); err != nil {
		response.ResponseError(c, fmt.Sprintf("Invalid data recieved by server: %v", err))
		return
	}

	data, err := controller.objectOperationService.Update(updateData)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not perform the updation of ObjectOperation: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (controller *objectOperationController) Delete(c *gin.Context) {
	idRaw := c.Param("id")
	id, err := strconv.ParseUint(idRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Incorrect parameter provided: %v", err))
		return
	}

	err = controller.objectOperationService.Delete(uint(id))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not perform the deletion of ObjectOperation: %v", err))
		return
	}

	response.ResponseSuccess(c, "deleted")
}
