package controller

import (
	"backend-v2/internal/service"
	"backend-v2/model"
	"backend-v2/pkg/response"
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
)

type materialForProjectRepository struct {
	materialForProjectService service.IMaterialForProjectService
}

func InitMaterialForProjectController(materialForProjectService service.IMaterialForProjectService) IMaterialForProjectController {
	return &materialForProjectRepository{
		materialForProjectService: materialForProjectService,
	}
}

type IMaterialForProjectController interface {
	GetAll(c *gin.Context)
	Create(c *gin.Context)
	Update(c *gin.Context)
	Delete(c *gin.Context)
}

func (controller *materialForProjectRepository) GetAll(c *gin.Context) {
	data, err := controller.materialForProjectService.GetAll()
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not get MaterialForProject data: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (controller *materialForProjectRepository) Create(c *gin.Context) {
	var createData model.MaterialForProject
	if err := c.ShouldBindJSON(&createData); err != nil {
		response.ResponseError(c, fmt.Sprintf("Invalid data recieved by server: %v", err))
		return
	}

	data, err := controller.materialForProjectService.Create(createData)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could perform the creation of MaterialForProject: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (controller *materialForProjectRepository) Update(c *gin.Context) {
	var updateData model.MaterialForProject
	if err := c.ShouldBindJSON(&updateData); err != nil {
		response.ResponseError(c, fmt.Sprintf("Invalid data recieved by server: %v", err))
		return
	}

	data, err := controller.materialForProjectService.Update(updateData)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not perform the updation of MaterialForProject: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (controller *materialForProjectRepository) Delete(c *gin.Context) {
	idRaw := c.Param("id")
	id, err := strconv.ParseUint(idRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Incorrect parameter provided: %v", err))
		return
	}

	err = controller.materialForProjectService.Delete(uint(id))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not perform the deletion of MaterialForProject: %v", err))
		return
	}

	response.ResponseSuccess(c, "deleted")
}
