package controller

import (
	"backend-v2/internal/dto"
	"backend-v2/internal/service"
	"backend-v2/model"
	"backend-v2/pkg/response"
	"fmt"
	"net/url"
	"strconv"

	"github.com/gin-gonic/gin"
)

type invoiceWriteOffController struct {
	invoiceWriteOffService service.IInvoiceWriteOffService
}

func InitInvoiceWriteOffController(invoiceWriteOffService service.IInvoiceWriteOffService) IInvoiceWriteOffController {
	return &invoiceWriteOffController{
		invoiceWriteOffService: invoiceWriteOffService,
	}
}

type IInvoiceWriteOffController interface {
	GetAll(c *gin.Context)
	GetPaginated(c *gin.Context)
	Create(c *gin.Context)
	// Update(c *gin.Context)
	Delete(c *gin.Context)
	GetRawDocument(c *gin.Context)
}

func (controller *invoiceWriteOffController) GetAll(c *gin.Context) {
	data, err := controller.invoiceWriteOffService.GetAll()
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not get Invoice Input data: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (controller *invoiceWriteOffController) GetPaginated(c *gin.Context) {
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

	deliveryCode := c.DefaultQuery("deliveryCode", "")
	deliveryCode, err = url.QueryUnescape(deliveryCode)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Wrong query parameter provided for deliveryCode: %v", err))
		return
	}

	writeOffType := c.DefaultQuery("writeOffType", "")
	writeOffType, err = url.QueryUnescape(writeOffType)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Wrong query parameter provided for returnerType: %v", err))
		return
	}

  projectID := c.GetUint("projectID")
	filter := model.InvoiceWriteOff{
    ProjectID: projectID,
		WriteOffType:         writeOffType,
		DeliveryCode:         deliveryCode,
	}

	data, err := controller.invoiceWriteOffService.GetPaginated(page, limit, filter)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not get the paginated data of Invoice: %v", err))
		return
	}

	dataCount, err := controller.invoiceWriteOffService.Count(projectID)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not get the total amount of Invoice: %v", err))
		return
	}

	response.ResponsePaginatedData(c, data, dataCount)
}

func (controller *invoiceWriteOffController) Create(c *gin.Context) {
	var createData dto.InvoiceWriteOff
	if err := c.ShouldBindJSON(&createData); err != nil {
		response.ResponseError(c, fmt.Sprintf("Invalid data recieved by server: %v", err))
		return
	}

	data, err := controller.invoiceWriteOffService.Create(createData)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could perform the creation of Invoice: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

// func (controller *invoiceWriteOffController) Update(c *gin.Context) {
// 	var updateData dto.InvoiceWriteOff
// 	if err := c.ShouldBindJSON(&updateData); err != nil {
// 		response.ResponseError(c, fmt.Sprintf("Invalid data recieved by server: %v", err))
// 		return
// 	}

// 	workerID := c.GetUint("workerID")
// 	updateData.Details.OperatorEditWorkerID = workerID

// 	data, err := controller.invoiceWriteOffService.Update(updateData)
// 	if err != nil {
// 		response.ResponseError(c, fmt.Sprintf("Could not perform the updation of Invoice: %v", err))
// 		return
// 	}

// 	response.ResponseSuccess(c, data)
// }

func (controller *invoiceWriteOffController) Delete(c *gin.Context) {
	idRaw := c.Param("id")
	id, err := strconv.ParseUint(idRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Incorrect parameter provided: %v", err))
		return
	}

	err = controller.invoiceWriteOffService.Delete(uint(id))
	if err != nil {

		response.ResponseError(c, fmt.Sprintf("Could not perform the deletion of Invoice: %v", err))
		return
	}

	response.ResponseSuccess(c, "deleted")
}

func (controller *invoiceWriteOffController) GetRawDocument(c *gin.Context) {
	deliveryCode := c.Param("deliveryCode")
	c.FileAttachment("./pkg/excels/writeoff/"+deliveryCode+".xlsx", deliveryCode+".xlsx")
}
