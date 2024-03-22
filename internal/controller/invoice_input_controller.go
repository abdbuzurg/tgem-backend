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

type invoiceInputController struct {
	invoiceInputService service.IInvoiceInputService
}

func InitInvoiceInputController(invoiceInputService service.IInvoiceInputService) IInvoiceInputController {
	return &invoiceInputController{
		invoiceInputService: invoiceInputService,
	}
}

type IInvoiceInputController interface {
	GetAll(c *gin.Context)
	GetPaginated(c *gin.Context)
	Create(c *gin.Context)
	// Update(c *gin.Context)
	Delete(c *gin.Context)
	Confirmation(c *gin.Context)
	GetDocument(c *gin.Context)
	UniqueCode(c *gin.Context)
	UniqueReleased(c *gin.Context)
	UniqueWarehouseManager(c *gin.Context)
	Report(c *gin.Context)
}

func (controller *invoiceInputController) GetAll(c *gin.Context) {
	data, err := controller.invoiceInputService.GetAll()
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not get Invoice Input data: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (controller *invoiceInputController) GetPaginated(c *gin.Context) {
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

	deliveryCode := c.DefaultQuery("deliveryCode", "")
	deliveryCode, err = url.QueryUnescape(deliveryCode)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Wrong query parameter provided for deliveryCode: %v", err))
		return
	}

	projectID := c.GetUint("projectID")
	filter := model.InvoiceInput{
		WarehouseManagerWorkerID: uint(warehouseManagerWorkerID),
		ReleasedWorkerID:         uint(releasedWorkerID),
		OperatorAddWorkerID:      uint(operatorAddWorkerID),
		OperatorEditWorkerID:     uint(operatorEditWorkerID),
		ProjectID:                projectID,
		DeliveryCode:             deliveryCode,
	}

	data, err := controller.invoiceInputService.GetPaginated(page, limit, filter)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not get the paginated data of Invoice: %v", err))
		return
	}

	dataCount, err := controller.invoiceInputService.Count(projectID)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not get the total amount of Invoice: %v", err))
		return
	}

	response.ResponsePaginatedData(c, data, dataCount)
}

func (controller *invoiceInputController) Create(c *gin.Context) {
	var createData dto.InvoiceInput
	if err := c.ShouldBindJSON(&createData); err != nil {
		response.ResponseError(c, fmt.Sprintf("Invalid data recieved by server: %v", err))
		return
	}

	workerID := c.GetUint("workerID")
	createData.Details.OperatorAddWorkerID = workerID
	createData.Details.OperatorEditWorkerID = workerID

	projectID := c.GetUint("projectID")
	createData.Details.ProjectID = projectID

	data, err := controller.invoiceInputService.Create(createData)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could perform the creation of Invoice: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

// func (controller *invoiceInputController) Update(c *gin.Context) {
// 	var updateData dto.InvoiceInput
// 	if err := c.ShouldBindJSON(&updateData); err != nil {
// 		response.ResponseError(c, fmt.Sprintf("Invalid data recieved by server: %v", err))
// 		return
// 	}
//
// 	workerID := c.GetUint("workerID")
// 	updateData.Details.OperatorEditWorkerID = workerID
//
// 	data, err := controller.invoiceInputService.Update(updateData)
// 	if err != nil {
// 		response.ResponseError(c, fmt.Sprintf("Could not perform the updation of Invoice: %v", err))
// 		return
// 	}
//
// 	response.ResponseSuccess(c, data)
// }

func (controller *invoiceInputController) Delete(c *gin.Context) {
	idRaw := c.Param("id")
	id, err := strconv.ParseUint(idRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Incorrect parameter provided: %v", err))
		return
	}

	err = controller.invoiceInputService.Delete(uint(id))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not perform the deletion of Invoice: %v", err))
		return
	}

	response.ResponseSuccess(c, "deleted")
}

func (controller *invoiceInputController) Confirmation(c *gin.Context) {
	idRaw := c.Param("id")
	id, err := strconv.ParseUint(idRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Incorrect parameter provided: %v", err))
		return
	}

	invoiceInput, err := controller.invoiceInputService.GetByID(uint(id))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Cannot find invoice Input by id %v: %v", id, err))
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("cannot form file: %v", err))
		return
	}
	file.Filename = invoiceInput.DeliveryCode

	filePath := "./pkg/excels/input/" + file.Filename + ".xlsx"
	err = c.SaveUploadedFile(file, filePath)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("cannot save file: %v", err))
		return
	}

	err = controller.invoiceInputService.Confirmation(uint(id))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("cannot confirm invoice input with id %v: %v", id, err))
		return
	}

	response.ResponseSuccess(c, true)
}

func (controller *invoiceInputController) GetDocument(c *gin.Context) {
	deliveryCode := c.Param("deliveryCode")
	c.FileAttachment("./pkg/excels/input/"+deliveryCode+".xlsx", deliveryCode+".xlsx")
}

func (controller *invoiceInputController) UniqueCode(c *gin.Context) {
	projectID := c.GetUint("projectID")
	data, err := controller.invoiceInputService.UniqueCode(projectID)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Internal server error: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (controller *invoiceInputController) UniqueWarehouseManager(c *gin.Context) {
	projectID := c.GetUint("projectID")
	data, err := controller.invoiceInputService.UniqueWarehouseManager(projectID)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Internal server error: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (controller *invoiceInputController) UniqueReleased(c *gin.Context) {
	projectID := c.GetUint("projectID")
	data, err := controller.invoiceInputService.UniqueReleased(projectID)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Internal server error: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (controller *invoiceInputController) Report(c *gin.Context) {
	var filter dto.InvoiceInputReportFilterRequest
	if err := c.ShouldBindJSON(&filter); err != nil {
		response.ResponseError(c, fmt.Sprintf("Invalid data recieved by server: %v", err))
		return
	}

	projectID := c.GetUint("projectID")
	filename, err := controller.invoiceInputService.Report(filter, projectID)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Internal server error: %v", err))
		return
	}

	c.FileAttachment("./pkg/excels/report/"+filename, filename)
	// response.ResponseSuccess(c, true)
}
