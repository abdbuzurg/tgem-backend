package controller

import (
	"backend-v2/internal/service"
	"backend-v2/model"
	"backend-v2/pkg/response"
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
)

type invoiceObjectSubmissionController struct {
	invoiceObjectSubmissionService service.IInvoiceObjectSubmissionService
}

func InitInvoiceObjectSubmissionController(
	invoiceObjectSubmissionService service.IInvoiceObjectSubmissionService,
) IInvoiceObjectSubmissionController {
	return &invoiceObjectSubmissionController{
		invoiceObjectSubmissionService: invoiceObjectSubmissionService,
	}
}

type IInvoiceObjectSubmissionController interface{}

func (controller *invoiceObjectSubmissionController) GetPaginated(c *gin.Context) {
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

	data, err := controller.invoiceObjectSubmissionService.GetPaginated(limit, page, projectID)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Internal Server Error: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (controller *invoiceObjectSubmissionController) Create(c *gin.Context) {
	var data model.InvoiceObjectSubmission
	if err := c.ShouldBindJSON(&data); err != nil {
		response.ResponseError(c, fmt.Sprintf("Incorrect body request: %v", err))
		return
	}

	data, err := controller.invoiceObjectSubmissionService.Create(data)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Internal Server Error: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (controller *invoiceObjectSubmissionController) Delete(c *gin.Context) {
	idRaw := c.Param("id")
	id, err := strconv.ParseUint(idRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Incorrect parameter provided: %v", err))
		return
	}

	err = controller.invoiceObjectSubmissionService.Delete(uint(id))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Internal Server Error: %v", err))
		return
	}

	response.ResponseSuccess(c, true)
}
