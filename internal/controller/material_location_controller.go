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

type materialLocationController struct {
	materialLocationService service.IMaterialLocationService
}

func InitMaterialLocationController(materialLocationService service.IMaterialLocationService) IMaterialLocationController {
	return &materialLocationController{
		materialLocationService: materialLocationService,
	}
}

type IMaterialLocationController interface {
	GetAll(c *gin.Context)
	GetPaginated(c *gin.Context)
	GetByID(c *gin.Context)
	Create(c *gin.Context)
	Update(c *gin.Context)
	Delete(c *gin.Context)
}

func (controller *materialLocationController) GetAll(c *gin.Context) {
	data, err := controller.materialLocationService.GetAll()
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not get MaterialLocation data: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (controller *materialLocationController) GetPaginated(c *gin.Context) {
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

	materialCostIDStr := c.DefaultQuery("materialCostID", "")
	materialCostID := 0
	if materialCostIDStr != "" {
		materialCostID, err = strconv.Atoi(materialCostIDStr)
		if err != nil {
			response.ResponseError(c, fmt.Sprintf("Cannot decode materialCostID parameter: %v", err))
			return
		}
	}

	materialDetailLocationIDStr := c.DefaultQuery("materialDetailLocationIDStr", "")
	materialDetailLocationID := 0
	if materialDetailLocationIDStr != "" {
		materialDetailLocationID, err = strconv.Atoi(materialCostIDStr)
		if err != nil {
			response.ResponseError(c, fmt.Sprintf("Cannot decode materialDetailLocationID parameter: %v", err))
			return
		}
	}

	locationType := c.DefaultQuery("locationType", "")
	locationType, err = url.QueryUnescape(locationType)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Wrong query parameter provided for locationType: %v", err))
		return
	}

	amountStr := c.DefaultQuery("amount", "")
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Wrong query parameter provided for amount: %v", err))
		return
	}
	var amount float64
	if amountStr != "" {
		amount, err = strconv.ParseFloat(amountStr, 64)
		if err != nil {
			response.ResponseError(c, fmt.Sprintf("Cannot decode amount parameter: %v", err))
			return
		}
	} else {
		amount = 0
	}

	filter := model.MaterialLocation{
		MaterialCostID:           uint(materialCostID),
		MaterialDetailLocationID: uint(materialDetailLocationID),
		LocationType:             locationType,
		Amount:                   amount,
	}

	data, err := controller.materialLocationService.GetPaginated(page, limit, filter)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not get the paginated data of MaterialLocation: %v", err))
		return
	}

	dataCount, err := controller.materialLocationService.Count()
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not get the total amount of MaterialLocation: %v", err))
		return
	}

	response.ResponsePaginatedData(c, data, dataCount)
}

func (controller *materialLocationController) GetByID(c *gin.Context) {
	idRaw := c.Param("id")
	id, err := strconv.ParseUint(idRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Incorrect parameter provided: %v", err))
		return
	}

	data, err := controller.materialLocationService.GetByID(uint(id))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not get the data with ID(%d): %v", id, err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (controller *materialLocationController) Create(c *gin.Context) {
	var createData model.MaterialLocation
	if err := c.ShouldBindJSON(&createData); err != nil {
		response.ResponseError(c, fmt.Sprintf("Invalid data recieved by server: %v", err))
		return
	}

	data, err := controller.materialLocationService.Create(createData)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could perform the creation of MaterialLocation: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (controller *materialLocationController) Update(c *gin.Context) {
	var updateData model.MaterialLocation
	if err := c.ShouldBindJSON(&updateData); err != nil {
		response.ResponseError(c, fmt.Sprintf("Invalid data recieved by server: %v", err))
		return
	}

	data, err := controller.materialLocationService.Update(updateData)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not perform the updation of MaterialLocation: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (controller *materialLocationController) Delete(c *gin.Context) {
	idRaw := c.Param("id")
	id, err := strconv.ParseUint(idRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Incorrect parameter provided: %v", err))
		return
	}

	err = controller.materialLocationService.Delete(uint(id))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not perform the deletion of MaterialLocation: %v", err))
		return
	}

	response.ResponseSuccess(c, "deleted")
}
