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

type materialController struct {
	materialService service.IMaterialService
}

func InitMaterialController(materialService service.IMaterialService) IMaterialController {
	return &materialController{
		materialService: materialService,
	}
}

type IMaterialController interface {
	GetAll(c *gin.Context)
	GetPaginated(c *gin.Context)
	GetByID(c *gin.Context)
	Create(c *gin.Context)
	Update(c *gin.Context)
	Delete(c *gin.Context)
}

func (controller *materialController) GetAll(c *gin.Context) {
	data, err := controller.materialService.GetAll()
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not get Material data: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (controller *materialController) GetPaginated(c *gin.Context) {
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

	category := c.DefaultQuery("category", "")
	category, err = url.QueryUnescape(category)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Wrong query parameter provided for category: %v", err))
		return
	}

	code := c.DefaultQuery("code", "")
	code, err = url.QueryUnescape(code)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Wrong query parameter provided for code: %v", err))
		return
	}

	name := c.DefaultQuery("name", "")
	name, err = url.QueryUnescape(name)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Wrong query parameter provided for name: %v", err))
		return
	}

	unit := c.DefaultQuery("unit", "")
	unit, err = url.QueryUnescape(unit)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Wrong query parameter provided for unit: %v", err))
		return
	}

	filter := model.Material{
		Category: category,
		Code:     code,
		Name:     name,
		Unit:     unit,
	}

	data, err := controller.materialService.GetPaginated(page, limit, filter)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not get the paginated data of Material: %v", err))
		return
	}

	dataCount, err := controller.materialService.Count()
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not get the total amount of Materials: %v", err))
		return
	}

	response.ResponsePaginatedData(c, data, dataCount)
}

func (controller *materialController) GetByID(c *gin.Context) {
	idRaw := c.Param("id")
	id, err := strconv.ParseUint(idRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Incorrect parameter provided: %v", err))
		return
	}

	data, err := controller.materialService.GetByID(uint(id))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not get the data with ID(%d): %v", id, err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (controller *materialController) Create(c *gin.Context) {
	var createData model.Material
	if err := c.ShouldBindJSON(&createData); err != nil {
		response.ResponseError(c, fmt.Sprintf("Invalid data recieved by server: %v", err))
		return
	}

	data, err := controller.materialService.Create(createData)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could perform the creation of Material: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (controller *materialController) Update(c *gin.Context) {
	var updateData model.Material
	if err := c.ShouldBindJSON(&updateData); err != nil {
		response.ResponseError(c, fmt.Sprintf("Invalid data recieved by server: %v", err))
		return
	}

	data, err := controller.materialService.Update(updateData)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not perform the updation of Material: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (controller *materialController) Delete(c *gin.Context) {
	idRaw := c.Param("id")
	id, err := strconv.ParseUint(idRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Incorrect parameter provided: %v", err))
		return
	}

	err = controller.materialService.Delete(uint(id))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not perform the deletion of Material: %v", err))
		return
	}

	response.ResponseSuccess(c, "deleted")
}
