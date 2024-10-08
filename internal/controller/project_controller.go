package controller

import (
	"backend-v2/internal/service"
	"backend-v2/model"
	"backend-v2/pkg/response"
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
)

type projectController struct {
	projectService service.IProjectService
}

func InitProjectController(projectService service.IProjectService) IProjectController {
	return &projectController{
		projectService: projectService,
	}
}

type IProjectController interface {
	GetAll(c *gin.Context)
	GetPaginated(c *gin.Context)
	GetByID(c *gin.Context)
	Create(c *gin.Context)
	Update(c *gin.Context)
	Delete(c *gin.Context)
	GetProjectName(c *gin.Context)
}

func (controller *projectController) GetAll(c *gin.Context) {
	data, err := controller.projectService.GetAll()
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not get Project data: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (controller *projectController) GetPaginated(c *gin.Context) {
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

	data, err := controller.projectService.GetPaginated(page, limit)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not get the paginated data of Project: %v", err))
		return
	}

	dataCount, err := controller.projectService.Count()
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not get the total amount of Project: %v", err))
		return
	}

	response.ResponsePaginatedData(c, data, dataCount)
}

func (controller *projectController) GetByID(c *gin.Context) {
	idRaw := c.Param("id")
	id, err := strconv.ParseUint(idRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Incorrect parameter provided: %v", err))
		return
	}

	data, err := controller.projectService.GetByID(uint(id))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not get the data with ID(%d): %v", id, err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (controller *projectController) Create(c *gin.Context) {
	var createData model.Project
	if err := c.ShouldBindJSON(&createData); err != nil {
		response.ResponseError(c, fmt.Sprintf("Invalid data recieved by server: %v", err))
		return
	}

	data, err := controller.projectService.Create(createData)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could perform the creation of Project: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (controller *projectController) Update(c *gin.Context) {
	var updateData model.Project
	if err := c.ShouldBindJSON(&updateData); err != nil {
		response.ResponseError(c, fmt.Sprintf("Invalid data recieved by server: %v", err))
		return
	}

	data, err := controller.projectService.Update(updateData)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not perform the updation of Project: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (controller *projectController) Delete(c *gin.Context) {
	idRaw := c.Param("id")
	id, err := strconv.ParseUint(idRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Incorrect parameter provided: %v", err))
		return
	}

	err = controller.projectService.Delete(uint(id))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not perform the deletion of Project: %v", err))
		return
	}

	response.ResponseSuccess(c, "deleted")
}

func (controller *projectController) GetProjectName(c *gin.Context) {
	projectName, err := controller.projectService.GetProjectName(c.GetUint("projectID"))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not perform the deletion of Project: %v", err))
		return
	}

  response.ResponseSuccess(c, projectName)
}
