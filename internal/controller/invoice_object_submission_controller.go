package controller

import (
	"backend-v2/internal/service"
	"backend-v2/model"
	"backend-v2/pkg/response"
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
)

type invoiceObjectController struct {
	invoiceObjectService service.IInvoiceObjectService
}

func InitInvoiceObjectController(
	invoiceObjectService service.IInvoiceObjectService,
) IInvoiceObjectController {
	return &invoiceObjectController{
		invoiceObjectService: invoiceObjectService,
	}
}

type IInvoiceObjectController interface{}

func (controller *invoiceObjectController) GetPaginated(c *gin.Context) {
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
  roleID := c.GetUint("roleID")
  workerID := c.GetUint("workerID")

	data, err := controller.invoiceObjectService.GetPaginated(limit, page, projectID, roleID, workerID)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Internal Server Error: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (controller *invoiceObjectController) Create(c *gin.Context) {
	var data model.InvoiceObject
	if err := c.ShouldBindJSON(&data); err != nil {
		response.ResponseError(c, fmt.Sprintf("Incorrect body request: %v", err))
		return
	}

  projectID := c.GetUint("projectID")
  data.ProjectID = projectID

	data, err := controller.invoiceObjectService.Create(data)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Internal Server Error: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (controller *invoiceObjectController) Delete(c *gin.Context) {
	idRaw := c.Param("id")
	id, err := strconv.ParseUint(idRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Incorrect parameter provided: %v", err))
		return
	}

	err = controller.invoiceObjectService.Delete(uint(id))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Internal Server Error: %v", err))
		return
	}

	response.ResponseSuccess(c, true)
}
