package controller

import (
	"backend-v2/internal/service"
	"backend-v2/model"
	"backend-v2/pkg/response"
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
)

type invoiceMaterialsController struct {
	invoiceMaterialsService service.IInvoiceMaterialsService
}

func InitInvoiceMaterialsController(invoiceMaterialsService service.IInvoiceMaterialsService) IInvoiceMaterialsController {
	return &invoiceMaterialsController{
		invoiceMaterialsService: invoiceMaterialsService,
	}
}

type IInvoiceMaterialsController interface {
	GetAll(c *gin.Context)
	GetPaginated(c *gin.Context)
	GetByID(c *gin.Context)
	Create(c *gin.Context)
	Update(c *gin.Context)
	Delete(c *gin.Context)
}

func (controller *invoiceMaterialsController) GetAll(c *gin.Context) {
	data, err := controller.invoiceMaterialsService.GetAll()
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not get InvoiceMaterials data: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (controller *invoiceMaterialsController) GetPaginated(c *gin.Context) {
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

	invoiceIDStr := c.DefaultQuery("invoiceID", "")
	invoiceID := 0
	if invoiceIDStr != "" {
		invoiceID, err = strconv.Atoi(invoiceIDStr)
		if err != nil {
			response.ResponseError(c, fmt.Sprintf("Cannot decode invoiceID parameter: %v", err))
			return
		}
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

	filter := model.InvoiceMaterials{
		MaterialCostID: uint(materialCostID),
		InvoiceID:      uint(invoiceID),
		Amount:         amount,
	}

	data, err := controller.invoiceMaterialsService.GetPaginated(page, limit, filter)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not get the paginated data of InvoiceMaterials: %v", err))
		return
	}

	dataCount, err := controller.invoiceMaterialsService.Count()
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not get the total amount of InvoiceMaterials: %v", err))
		return
	}

	response.ResponsePaginatedData(c, data, dataCount)
}

func (controller *invoiceMaterialsController) GetByID(c *gin.Context) {
	idRaw := c.Param("id")
	id, err := strconv.ParseUint(idRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Incorrect parameter provided: %v", err))
		return
	}

	data, err := controller.invoiceMaterialsService.GetByID(uint(id))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not get the data with ID(%d): %v", id, err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (controller *invoiceMaterialsController) Create(c *gin.Context) {
	var createData model.InvoiceMaterials
	if err := c.ShouldBindJSON(&createData); err != nil {
		response.ResponseError(c, fmt.Sprintf("Invalid data recieved by server: %v", err))
		return
	}

	data, err := controller.invoiceMaterialsService.Create(createData)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could perform the creation of InvoiceMaterials: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (controller *invoiceMaterialsController) Update(c *gin.Context) {
	var updateData model.InvoiceMaterials
	if err := c.ShouldBindJSON(&updateData); err != nil {
		response.ResponseError(c, fmt.Sprintf("Invalid data recieved by server: %v", err))
		return
	}

	data, err := controller.invoiceMaterialsService.Update(updateData)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not perform the updation of InvoiceMaterials: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (controller *invoiceMaterialsController) Delete(c *gin.Context) {
	idRaw := c.Param("id")
	id, err := strconv.ParseUint(idRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Incorrect parameter provided: %v", err))
		return
	}

	err = controller.invoiceMaterialsService.Delete(uint(id))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not perform the deletion of InvoiceMaterials: %v", err))
		return
	}

	response.ResponseSuccess(c, "deleted")
}
