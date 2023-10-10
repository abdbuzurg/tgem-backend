package controller

import (
	"backend-v2/internal/dto"
	"backend-v2/internal/service"
	"backend-v2/model"
	"backend-v2/pkg/response"
	"backend-v2/pkg/utils"
	"fmt"
	"net/url"
	"strconv"

	"github.com/gin-gonic/gin"
)

type invoiceController struct {
	invoiceService service.IInvoiceService
}

func InitInvoiceController(invoiceService service.IInvoiceService) IInvoiceController {
	return &invoiceController{
		invoiceService: invoiceService,
	}
}

type IInvoiceController interface {
	GetAll(c *gin.Context)
	GetPaginated(c *gin.Context)
	GetByID(c *gin.Context)
	Create(c *gin.Context)
	Update(c *gin.Context)
	Delete(c *gin.Context)
}

func (controller *invoiceController) GetAll(c *gin.Context) {
	invoiceType := c.Param("invoiceType")
	if !utils.IsCorrectInvoiceType(invoiceType) {
		response.ResponseError(c, "Incorrect invoice type provided")
		return
	}

	data, err := controller.invoiceService.GetAll(invoiceType)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not get Invoice data: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (controller *invoiceController) GetPaginated(c *gin.Context) {
	invoiceType := c.Param("invoiceType")
	if !utils.IsCorrectInvoiceType(invoiceType) {
		response.ResponseError(c, "Incorrect invoice type provided")
		return
	}

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

	projectIDStr := c.DefaultQuery("projectID", "")
	projectID := 0
	if projectIDStr != "" {
		projectID, err = strconv.Atoi(projectIDStr)
		if err != nil {
			response.ResponseError(c, fmt.Sprintf("Cannot decode projectID parameter: %v", err))
			return
		}
	}

	teamIDStr := c.DefaultQuery("teamID", "")
	teamID := 0
	if teamIDStr != "" {
		teamID, err = strconv.Atoi(teamIDStr)
		if err != nil {
			response.ResponseError(c, fmt.Sprintf("Cannot decode teamID parameter: %v", err))
			return
		}
	}

	warehouseManagerWorkerIDStr := c.DefaultQuery("warehouseManagerWorkerID", "")
	warehouseManagerWorkerID := 0
	if warehouseManagerWorkerIDStr != "" {
		warehouseManagerWorkerID, err = strconv.Atoi(warehouseManagerWorkerIDStr)
		if err != nil {
			response.ResponseError(c, fmt.Sprintf("Cannot decode invoiceCo parameter: %v", err))
			return
		}
	}

	releasedWorkerIDStr := c.DefaultQuery("releasedWorkerID", "")
	releasedWorkerID := 0
	if releasedWorkerIDStr != "" {
		releasedWorkerID, err = strconv.Atoi(releasedWorkerIDStr)
		if err != nil {
			response.ResponseError(c, fmt.Sprintf("Cannot decode releasedWorkerID parameter: %v", err))
			return
		}
	}

	driverWorkerIDStr := c.DefaultQuery("driverWorkerID", "")
	driverWorkerID := 0
	if driverWorkerIDStr != "" {
		driverWorkerID, err = strconv.Atoi(driverWorkerIDStr)
		if err != nil {
			response.ResponseError(c, fmt.Sprintf("Cannot decode driverWorkerID parameter: %v", err))
			return
		}
	}

	recipientWorkerIDStr := c.DefaultQuery("recipientWorkerID", "")
	recipientWorkerID := 0
	if recipientWorkerIDStr != "" {
		recipientWorkerID, err = strconv.Atoi(recipientWorkerIDStr)
		if err != nil {
			response.ResponseError(c, fmt.Sprintf("Cannot decode recipientWorkerID parameter: %v", err))
			return
		}
	}

	operatorAddWorkerIDStr := c.DefaultQuery("operatorAddWorkerID", "")
	operatorAddWorkerID := 0
	if operatorAddWorkerIDStr != "" {
		operatorAddWorkerID, err = strconv.Atoi(operatorAddWorkerIDStr)
		if err != nil {
			response.ResponseError(c, fmt.Sprintf("Cannot decode operatorAddWorkerID parameter: %v", err))
			return
		}
	}

	operatorEditWorkerIDStr := c.DefaultQuery("operatorEditWorkerID", "")
	operatorEditWorkerID := 0
	if operatorEditWorkerIDStr != "" {
		operatorEditWorkerID, err = strconv.Atoi(operatorEditWorkerIDStr)
		if err != nil {
			response.ResponseError(c, fmt.Sprintf("Cannot decode operatorEditWorkerID parameter: %v", err))
			return
		}
	}

	objectIDStr := c.DefaultQuery("objectID", "")
	objectID := 0
	if objectIDStr != "" {
		objectID, err = strconv.Atoi(objectIDStr)
		if err != nil {
			response.ResponseError(c, fmt.Sprintf("Cannot decode objectID parameter: %v", err))
			return
		}
	}

	deliveryCode := c.DefaultQuery("deliveryCode", "")
	deliveryCode, err = url.QueryUnescape(deliveryCode)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Wrong query parameter provided for deliveryCode: %v", err))
		return
	}

	district := c.DefaultQuery("district", "")
	district, err = url.QueryUnescape(district)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Wrong query parameter provided for district: %v", err))
		return
	}

	carNumber := c.DefaultQuery("carNumber", "")
	carNumber, err = url.QueryUnescape(carNumber)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Wrong query parameter provided for carNumber: %v", err))
		return
	}

	filter := model.Invoice{
		ProjectID:                uint(projectID),
		TeamID:                   uint(teamID),
		WarehouseManagerWorkerID: uint(warehouseManagerWorkerID),
		ReleasedWorkerID:         uint(releasedWorkerID),
		DriverWorkerID:           uint(driverWorkerID),
		RecipientWorkerID:        uint(recipientWorkerID),
		OperatorAddWorkerID:      uint(operatorAddWorkerID),
		OperatorEditWorkerID:     uint(operatorEditWorkerID),
		ObjectID:                 uint(objectID),
		DeliveryCode:             deliveryCode,
		District:                 district,
		CarNumber:                carNumber,
	}

	data, err := controller.invoiceService.GetPaginated(invoiceType, page, limit, filter)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not get the paginated data of Invoice: %v", err))
		return
	}

	dataCount, err := controller.invoiceService.Count(invoiceType)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not get the total amount of Invoice: %v", err))
		return
	}

	response.ResponsePaginatedData(c, data, dataCount)
}

func (controller *invoiceController) GetByID(c *gin.Context) {
	idRaw := c.Param("id")
	id, err := strconv.ParseUint(idRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Incorrect parameter provided: %v", err))
		return
	}

	data, err := controller.invoiceService.GetByID(uint(id))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not get the data with ID(%d): %v", id, err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (controller *invoiceController) Create(c *gin.Context) {
	var createData dto.InvoiceDataUpdateOrCreate
	if err := c.ShouldBindJSON(&createData); err != nil {
		response.ResponseError(c, fmt.Sprintf("Invalid data recieved by server: %v", err))
		return
	}

	data, err := controller.invoiceService.Create(createData)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could perform the creation of Invoice: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (controller *invoiceController) Update(c *gin.Context) {
	var updateData dto.InvoiceDataUpdateOrCreate
	if err := c.ShouldBindJSON(&updateData); err != nil {
		response.ResponseError(c, fmt.Sprintf("Invalid data recieved by server: %v", err))
		return
	}

	data, err := controller.invoiceService.Update(updateData)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not perform the updation of Invoice: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (controller *invoiceController) Delete(c *gin.Context) {
	idRaw := c.Param("id")
	id, err := strconv.ParseUint(idRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Incorrect parameter provided: %v", err))
		return
	}

	err = controller.invoiceService.Delete(uint(id))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not perform the deletion of Invoice: %v", err))
		return
	}

	response.ResponseSuccess(c, "deleted")
}
