package controller

import (
	"backend-v2/internal/service"
	"backend-v2/model"
	"backend-v2/pkg/response"
	"fmt"
	"net/url"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

type operationController struct {
	operationService service.IOperationService
}

func InitOperationController(operationService service.IOperationService) IOperationController {
	return &operationController{
		operationService: operationService,
	}
}

type IOperationController interface {
	GetAll(c *gin.Context)
	GetPaginated(c *gin.Context)
	GetByID(c *gin.Context)
	Create(c *gin.Context)
	Update(c *gin.Context)
	Delete(c *gin.Context)
}

func (controller *operationController) GetAll(c *gin.Context) {
	data, err := controller.operationService.GetAll()
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not get Operation data: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (controller *operationController) GetPaginated(c *gin.Context) {
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

	costPrimerStr := c.DefaultQuery("costPrimer", "")
	costPrime, err := decimal.NewFromString(costPrimerStr)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Cannot get the costPrime parameter: %v", err))
		return
	}

	costWithCustomerStr := c.DefaultQuery("costWithCustomer", "")
	costWithCustomer, err := decimal.NewFromString(costWithCustomerStr)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Cannot get the costWithCustomer parameter: %v", err))
		return
	}

	name := c.DefaultQuery("name", "")
	name, err = url.QueryUnescape(name)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Wrong query parameter provided for name: %v", err))
		return
	}

	code := c.DefaultQuery("code", "")
	code, err = url.QueryUnescape(code)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Wrong query parameter provided for code: %v", err))
		return
	}

	filter := model.Operation{
		Name:             name,
		Code:             code,
		CostPrime:        costPrime,
		CostWithCustomer: costWithCustomer,
	}

	data, err := controller.operationService.GetPaginated(page, limit, filter)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not get the paginated data of Operation: %v", err))
		return
	}

	dataCount, err := controller.operationService.Count()
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not get the total amount of Operation: %v", err))
		return
	}

	response.ResponsePaginatedData(c, data, dataCount)
}

func (controller *operationController) GetByID(c *gin.Context) {
	idRaw := c.Param("id")
	id, err := strconv.ParseUint(idRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Incorrect parameter provided: %v", err))
		return
	}

	data, err := controller.operationService.GetByID(uint(id))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not get the data with ID(%d): %v", id, err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (controller *operationController) Create(c *gin.Context) {
	var createData model.Operation
	if err := c.ShouldBindJSON(&createData); err != nil {
		response.ResponseError(c, fmt.Sprintf("Invalid data recieved by server: %v", err))
		return
	}

	data, err := controller.operationService.Create(createData)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could perform the creation of Operation: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (controller *operationController) Update(c *gin.Context) {
	var updateData model.Operation
	if err := c.ShouldBindJSON(&updateData); err != nil {
		response.ResponseError(c, fmt.Sprintf("Invalid data recieved by server: %v", err))
		return
	}

	data, err := controller.operationService.Update(updateData)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not perform the updation of Operation: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (controller *operationController) Delete(c *gin.Context) {
	idRaw := c.Param("id")
	id, err := strconv.ParseUint(idRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Incorrect parameter provided: %v", err))
		return
	}

	err = controller.operationService.Delete(uint(id))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not perform the deletion of Operation: %v", err))
		return
	}

	response.ResponseSuccess(c, "deleted")
}
