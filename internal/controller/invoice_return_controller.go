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

type invoiceReturnController struct {
	invoiceReturnService service.IInvoiceReturnService
}

func InitInvoiceReturnController(invoiceReturnService service.IInvoiceReturnService) IInvoiceReturnController {
	return &invoiceReturnController{
		invoiceReturnService: invoiceReturnService,
	}
}

type IInvoiceReturnController interface {
	GetAll(c *gin.Context)
	GetPaginated(c *gin.Context)
	Create(c *gin.Context)
	// Update(c *gin.Context)
	Delete(c *gin.Context)
	GetDocument(c *gin.Context)
	Confirmation(c *gin.Context)
	UniqueCode(c *gin.Context)
	UniqueTeam(c *gin.Context)
	UniqueObject(c *gin.Context)
	Report(c *gin.Context)
	GetUniqueMaterialCostsFromLocation(c *gin.Context)
	GetMaterialsInLocation(c *gin.Context)
	GetMaterialAmountInLocation(c *gin.Context)
  GetSerialNumberCodesInLocation(c *gin.Context) 
}

func (controller *invoiceReturnController) GetAll(c *gin.Context) {
	data, err := controller.invoiceReturnService.GetAll()
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not get Invoice Input data: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (controller *invoiceReturnController) GetPaginated(c *gin.Context) {
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

	returnerIDStr := c.DefaultQuery("returnerID", "")
	returnerID := 0
	if returnerIDStr != "" {
		returnerID, err = strconv.Atoi(returnerIDStr)
		if err != nil {
			response.ResponseError(c, fmt.Sprintf("Cannot decode returnerID parameter: %v", err))
			return
		}
	}

	deliveryCode := c.DefaultQuery("deliveryCode", "")
	deliveryCode, err = url.QueryUnescape(deliveryCode)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Wrong query parameter provided for deliveryCode: %v", err))
		return
	}

	returnerType := c.DefaultQuery("returnerType", "")
	returnerType, err = url.QueryUnescape(returnerType)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Wrong query parameter provided for returnerType: %v", err))
		return
	}

	projectID := c.GetUint("projectID")
	filter := model.InvoiceReturn{
		ProjectID:    projectID,
		ReturnerType: returnerType,
		ReturnerID:   uint(returnerID),
		DeliveryCode: deliveryCode,
	}

	data, err := controller.invoiceReturnService.GetPaginated(page, limit, filter)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not get the paginated data of Invoice: %v", err))
		return
	}

	dataCount, err := controller.invoiceReturnService.Count(projectID)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not get the total amount of Invoice: %v", err))
		return
	}

	response.ResponsePaginatedData(c, data, dataCount)
}

func (controller *invoiceReturnController) Create(c *gin.Context) {
	var createData dto.InvoiceReturn
	if err := c.ShouldBindJSON(&createData); err != nil {
		response.ResponseError(c, fmt.Sprintf("Invalid data recieved by server: %v", err))
		return
	}

	projectID := c.GetUint("projectID")
	createData.Details.ProjectID = projectID
	data, err := controller.invoiceReturnService.Create(createData)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could perform the creation of Invoice: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (controller *invoiceReturnController) Delete(c *gin.Context) {
	idRaw := c.Param("id")
	id, err := strconv.ParseUint(idRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Incorrect parameter provided: %v", err))
		return
	}

	err = controller.invoiceReturnService.Delete(uint(id))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not perform the deletion of Invoice: %v", err))
		return
	}

	response.ResponseSuccess(c, "deleted")
}

func (controller *invoiceReturnController) Confirmation(c *gin.Context) {
	idRaw := c.Param("id")
	id, err := strconv.ParseUint(idRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Incorrect parameter provided: %v", err))
		return
	}

	invoiceReturn, err := controller.invoiceReturnService.GetByID(uint(id))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Cannot find invoice Return by id %v: %v", id, err))
	}

	file, err := c.FormFile("file")
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("cannot form file: %v", err))
		return
	}
	file.Filename = invoiceReturn.DeliveryCode

	filePath := "./pkg/excels/return/" + file.Filename + ".xlsx"
	err = c.SaveUploadedFile(file, filePath)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("cannot save file: %v", err))
		return
	}

	err = controller.invoiceReturnService.Confirmation(uint(id))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("cannot confirm invoice input with id %v: %v", id, err))
		return
	}

	response.ResponseSuccess(c, true)
}

func (controller *invoiceReturnController) GetDocument(c *gin.Context) {
	deliveryCode := c.Param("deliveryCode")
	c.FileAttachment("./pkg/excels/return/"+deliveryCode+".xlsx", deliveryCode+".xlsx")
}

func (controller *invoiceReturnController) UniqueCode(c *gin.Context) {
	projectID := c.GetUint("projectID")
	data, err := controller.invoiceReturnService.UniqueCode(projectID)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Internal Server Error: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (controller *invoiceReturnController) UniqueTeam(c *gin.Context) {
	projectID := c.GetUint("projectID")
	data, err := controller.invoiceReturnService.UniqueTeam(projectID)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Internal Server Error: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (controller *invoiceReturnController) UniqueObject(c *gin.Context) {
	projectID := c.GetUint("projectID")
	data, err := controller.invoiceReturnService.UniqueObject(projectID)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Internal Server Error: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (controller *invoiceReturnController) Report(c *gin.Context) {
	var filter dto.InvoiceReturnReportFilterRequest
	if err := c.ShouldBindJSON(&filter); err != nil {
		response.ResponseError(c, fmt.Sprintf("Invalid body request: %v", err))
		return
	}

	projectID := c.GetUint("projectID")
	filename, err := controller.invoiceReturnService.Report(filter, projectID)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Internal server error: %v", err))
		return
	}

	c.FileAttachment("./pkg/excels/report/"+filename, filename)
}

func (controller *invoiceReturnController) GetUniqueMaterialCostsFromLocation(c *gin.Context) {

	projectID := c.GetUint("projectID")

	locationIDRaw := c.Param("locationID")
	locationID, err := strconv.Atoi(locationIDRaw)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Invalid parameters in request: %v", err))
		return
	}

	locationType := c.Param("locationType")

	materialIDRaw := c.Param("materialID")
	materialID, err := strconv.Atoi(materialIDRaw)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Invalid parameters in request: %v", err))
		return
	}

	data, err := controller.invoiceReturnService.GetMaterialCostInLocation(projectID, uint(locationID), uint(materialID), locationType)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Internal server error: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (controller *invoiceReturnController) GetMaterialsInLocation(c *gin.Context) {

	projectID := c.GetUint("projectID")

	locationIDRaw := c.Param("locationID")
	locationID, err := strconv.Atoi(locationIDRaw)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Invalid parameters in request: %v", err))
		return
	}

	locationType := c.Param("locationType")

	data, err := controller.invoiceReturnService.GetMaterialsInLocation(projectID, uint(locationID), locationType)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Internal server error: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (controller *invoiceReturnController) GetMaterialAmountInLocation(c *gin.Context) {

	projectID := c.GetUint("projectID")

	locationIDRaw := c.Param("locationID")
	locationID, err := strconv.Atoi(locationIDRaw)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Invalid parameters in request: %v", err))
		return
	}

	locationType := c.Param("locationType")

	materialCostIDRaw := c.Param("materialCostID")
	materialCostID, err := strconv.Atoi(materialCostIDRaw)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Invalid parameters in request: %v", err))
		return
	}

	data, err := controller.invoiceReturnService.GetMaterialAmountInLocation(projectID, uint(locationID), uint(materialCostID), locationType)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Internal server error: %v", err))
		return
	}

	response.ResponseSuccess(c, data)

}

func (controller *invoiceReturnController) GetSerialNumberCodesInLocation(c *gin.Context) {

	materialIDRaw := c.Param("materialID")
	materialID, err := strconv.Atoi(materialIDRaw)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Invalid parameters in request: %v", err))
		return
	}

  status := c.Param("status")
  projectID := c.GetUint("projectID")

  data, err := controller.invoiceReturnService.GetSerialNumberCodesInLocation(projectID, uint(materialID), status)
  if err != nil {
		response.ResponseError(c, fmt.Sprintf("Internal server error: %v", err))
		return
	}

  response.ResponseSuccess(c, data)
}
