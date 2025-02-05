package controller

import (
	"backend-v2/internal/service"
	"backend-v2/pkg/response"
	"fmt"

	"github.com/gin-gonic/gin"
)

type statisticsController struct {
	statService service.IStatisticsService
}

type IStatisticsController interface {
	InvoiceCountStat(c *gin.Context)
	InvoiceInputCreatorStat(c *gin.Context)
	InvoiceOutputCreatorStat(c *gin.Context)
}

func NewStatisticsController(statService service.IStatisticsService) IStatisticsController {
	return &statisticsController{
		statService: statService,
	}
}

func (controller *statisticsController) InvoiceCountStat(c *gin.Context) {
	data, err := controller.statService.InvoiceCountStat(c.GetUint("projectID"))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Система не смогла собрать данные: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (controller *statisticsController) InvoiceInputCreatorStat(c *gin.Context) {
	data, err := controller.statService.InvoiceInputCreatorStat(c.GetUint("projectID"))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Система не смогла собрать данные: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (controller *statisticsController) InvoiceOutputCreatorStat(c *gin.Context) {
	data, err := controller.statService.InvoiceOutputCreatorStat(c.GetUint("projectID"))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Система не смогла собрать данные: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}
