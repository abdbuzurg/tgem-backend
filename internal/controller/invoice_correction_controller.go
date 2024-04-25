package controller

import (
	"backend-v2/internal/service"
	"backend-v2/pkg/response"
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
)

type invoiceCorrectionController struct {
	invoiceCorrectionService service.IInvoiceCorrectionService
}

func InitInvoiceCorrectionController(
	invoiceCorrectionService service.IInvoiceCorrectionService,
) IInvoiceCorrectionController {
	return &invoiceCorrectionController{
		invoiceCorrectionService: invoiceCorrectionService,
	}
}

type IInvoiceCorrectionController interface {
	GetAll(c *gin.Context)
	GetMaterialsFromInvoiceObjectForCorrection(c *gin.Context)
	GetTotalMaterialInTeamByTeamNumber(c *gin.Context)
}

func (controller *invoiceCorrectionController) GetAll(c *gin.Context) {

	projectID := c.GetUint("projectID")

	data, err := controller.invoiceCorrectionService.GetAll(projectID)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Internal server error: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (controller *invoiceCorrectionController) GetMaterialsFromInvoiceObjectForCorrection(c *gin.Context) {
	invoiceIDRaw := c.Param("invoiceID")
	invoiceID, err := strconv.ParseUint(invoiceIDRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Incorrect parameter provided: %v", err))
		return
	}

	projectID := c.GetUint("projectID")

	data, err := controller.invoiceCorrectionService.GetMaterialsFromInvoiceObjectForCorrection(projectID, uint(invoiceID))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Internal server error: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (controller *invoiceCorrectionController) GetTotalMaterialInTeamByTeamNumber(c *gin.Context) {

	materialIDRaw := c.Param("materialID")
	materialID, err := strconv.ParseUint(materialIDRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Incorrect parameter provided: %v", err))
		return
	}

	teamNumber := c.Param("teamNumber")
	projectID := c.GetUint("projectID")

	data, err := controller.invoiceCorrectionService.GetTotalAmounInLocationByTeamName(projectID, uint(materialID), teamNumber)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Internal server error: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}
