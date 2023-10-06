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

type objectController struct {
	objectService service.IObjectService
}

func InitObjectController(objectService service.IObjectService) IObjectController {
	return &objectController{
		objectService: objectService,
	}
}

type IObjectController interface {
	GetAll(c *gin.Context)
	GetPaginated(c *gin.Context)
	GetByID(c *gin.Context)
	Create(c *gin.Context)
	Update(c *gin.Context)
	Delete(c *gin.Context)
}

func (controller *objectController) GetAll(c *gin.Context) {
	data, err := controller.objectService.GetAll()
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not get Object data: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (controller *objectController) GetPaginated(c *gin.Context) {
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

	objectDetailedIDStr := c.DefaultQuery("objectDetailedID", "")
	objectDetailedID := 0
	if objectDetailedIDStr != "" {
		objectDetailedID, err = strconv.Atoi(objectDetailedIDStr)
		if err != nil {
			response.ResponseError(c, fmt.Sprintf("Cannot decode objectDetailedID parameter: %v", err))
			return
		}
	}

	supervisorWorkerIDStr := c.DefaultQuery("supervisorWorkerID", "")
	supervisorWorkerID := 0
	if supervisorWorkerIDStr != "" {
		supervisorWorkerID, err = strconv.Atoi(supervisorWorkerIDStr)
		if err != nil {
			response.ResponseError(c, fmt.Sprintf("Cannot decode supervisorWorkerID parameter: %v", err))
			return
		}
	}

	objectType := c.DefaultQuery("objectType", "")
	objectType, err = url.QueryUnescape(objectType)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Wrong query parameter provided for objectType: %v", err))
		return
	}

	name := c.DefaultQuery("name", "")
	name, err = url.QueryUnescape(name)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Wrong query parameter provided for objectType: %v", err))
		return
	}

	status := c.DefaultQuery("status", "")
	status, err = url.QueryUnescape(status)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Wrong query parameter provided for status: %v", err))
		return
	}

	filter := model.Object{
		ObjectDetailedID:   uint(objectDetailedID),
		SupervisorWorkerID: uint(supervisorWorkerID),
		Type:               objectType,
		Name:               name,
		Status:             status,
	}

	data, err := controller.objectService.GetPaginated(page, limit, filter)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not get the paginated data of Object: %v", err))
		return
	}

	dataCount, err := controller.objectService.Count()
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not get the total amount of Object: %v", err))
		return
	}

	response.ResponsePaginatedData(c, data, dataCount)
}

func (controller *objectController) GetByID(c *gin.Context) {
	idRaw := c.Param("id")
	id, err := strconv.ParseUint(idRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Incorrect parameter provided: %v", err))
		return
	}

	data, err := controller.objectService.GetByID(uint(id))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not get the data with ID(%d): %v", id, err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (controller *objectController) Create(c *gin.Context) {
	var createData model.Object
	if err := c.ShouldBindJSON(&createData); err != nil {
		response.ResponseError(c, fmt.Sprintf("Invalid data recieved by server: %v", err))
		return
	}

	data, err := controller.objectService.Create(createData)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could perform the creation of Object: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (controller *objectController) Update(c *gin.Context) {
	var updateData model.Object
	if err := c.ShouldBindJSON(&updateData); err != nil {
		response.ResponseError(c, fmt.Sprintf("Invalid data recieved by server: %v", err))
		return
	}

	data, err := controller.objectService.Update(updateData)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not perform the updation of Object: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (controller *objectController) Delete(c *gin.Context) {
	idRaw := c.Param("id")
	id, err := strconv.ParseUint(idRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Incorrect parameter provided: %v", err))
		return
	}

	err = controller.objectService.Delete(uint(id))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not perform the deletion of Object: %v", err))
		return
	}

	response.ResponseSuccess(c, "deleted")
}
