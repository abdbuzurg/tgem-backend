package controller

import (
	"backend-v2/internal/service"
	"backend-v2/pkg/response"
	"fmt"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

type mainReportController struct {
	mainReportService service.IMainReportService
}

type IMainReportController interface {
	ProjectProgress(c *gin.Context)
	RemainingMaterialAnalysis(c *gin.Context)
}

func InitMainReportController(mainReportService service.IMainReportService) IMainReportController {
	return &mainReportController{
		mainReportService: mainReportService,
	}
}

func (controller *mainReportController) ProjectProgress(c *gin.Context) {
	type projectProgressRequestData struct {
		Date time.Time `json:"date"`
	}
	var data projectProgressRequestData
	if err := c.ShouldBindJSON(&data); err != nil {
		response.ResponseError(c, fmt.Sprintf("Неверное тело запроса: %v", err))
		return
	}

	loc, _ := time.LoadLocation("Asia/Dushanbe")
	dateGiven := data.Date.In(loc)
	dateNow := time.Now().In(loc)
	var progressReportFilePath string
	var err error
	if dateGiven.Day() == dateNow.Day() && dateGiven.Month() == dateNow.Month() && dateGiven.Year() == dateNow.Year() {
		progressReportFilePath, err = controller.mainReportService.ProjectProgress(c.GetUint("projectID"))
		if err != nil {
			response.ResponseError(c, fmt.Sprintf("%v", err))
			return
		}
	} else {
		if dateGiven.After(dateNow) {
			response.ResponseError(c, fmt.Sprint("Невозможно получить прогресс проекта в будущем"))
			return
		}

		progressReportFilePath, err = controller.mainReportService.ProjectProgressByGivenDay(c.GetUint("projectID"), dateGiven)
		if err != nil {
			response.ResponseError(c, fmt.Sprintf("%v", err))
			return
		}
	}

	c.FileAttachment(progressReportFilePath, "Прогресс Проекта.xlsx")
	if err := os.Remove(progressReportFilePath); err != nil {
		fmt.Printf("Error deleting file: %s", progressReportFilePath)
	}
}

func (controller *mainReportController) RemainingMaterialAnalysis(c *gin.Context) {
	remainingMaterialAnalysisFilePath, err := controller.mainReportService.RemainingMaterialAnalysis(c.GetUint("projectID"))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("%v", err))
		return
	}

	c.FileAttachment(remainingMaterialAnalysisFilePath, "Анализ Остатка Материалов.xlsx")
	if err := os.Remove(remainingMaterialAnalysisFilePath); err != nil {
		fmt.Printf("Error deleting file: %s", remainingMaterialAnalysisFilePath)
	}
}
