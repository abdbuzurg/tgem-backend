package controller

import (
	"backend-v2/internal/service"
	"backend-v2/model"
	"backend-v2/pkg/response"
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
)

type materialCostController struct {
	materialCostService service.IMaterialCostService
}

func InitMaterialCostController(materialCostService service.IMaterialCostService) IMaterialCostController {
	return &materialCostController{
		materialCostService: materialCostService,
	}
}

type IMaterialCostController interface {
	GetAll(c *gin.Context)
	GetPaginated(c *gin.Context)
	GetByID(c *gin.Context)
	Create(c *gin.Context)
	Update(c *gin.Context)
	Delete(c *gin.Context)
	GetAllMaterialCostByMaterialID(c *gin.Context)
}

func (controller *materialCostController) GetAll(c *gin.Context) {
	data, err := controller.materialCostService.GetAll()
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not get MaterialCost data: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (controller *materialCostController) GetPaginated(c *gin.Context) {
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

  projectID := c.GetUint("projectID")

	data, err := controller.materialCostService.GetPaginated(page, limit, projectID)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not get the paginated data of MaterialCost: %v", err))
		return
	}

	dataCount, err := controller.materialCostService.Count()
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not get the total amount of MaterialCost: %v", err))
		return
	}

	response.ResponsePaginatedData(c, data, dataCount)
}

func (controller *materialCostController) GetByID(c *gin.Context) {
	idRaw := c.Param("id")
	id, err := strconv.ParseUint(idRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Incorrect parameter provided: %v", err))
		return
	}

	data, err := controller.materialCostService.GetByID(uint(id))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not get the data with ID(%d): %v", id, err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (controller *materialCostController) GetAllMaterialCostByMaterialID(c *gin.Context) {
	materialIDRaw := c.Param("materialID")
	materialID, err := strconv.ParseUint(materialIDRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Incorrect parameter provided: %v", err))
		return
	}

	data, err := controller.materialCostService.GetByMaterialID(uint(materialID))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Internal Server Error: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (controller *materialCostController) Create(c *gin.Context) {
	var createData model.MaterialCost
	if err := c.ShouldBindJSON(&createData); err != nil {
		response.ResponseError(c, fmt.Sprintf("Invalid data recieved by server: %v", err))
		return
	}

	data, err := controller.materialCostService.Create(createData)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could perform the creation of MaterialCost: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (controller *materialCostController) Update(c *gin.Context) {
	var updateData model.MaterialCost
	if err := c.ShouldBindJSON(&updateData); err != nil {
		response.ResponseError(c, fmt.Sprintf("Invalid data recieved by server: %v", err))
		return
	}

	data, err := controller.materialCostService.Update(updateData)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not perform the updation of MaterialCost: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (controller *materialCostController) Delete(c *gin.Context) {
	idRaw := c.Param("id")
	id, err := strconv.ParseUint(idRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Incorrect parameter provided: %v", err))
		return
	}

	err = controller.materialCostService.Delete(uint(id))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not perform the deletion of MaterialCost: %v", err))
		return
	}

	response.ResponseSuccess(c, "deleted")
}
