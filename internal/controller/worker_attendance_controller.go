package controller

import (
	"backend-v2/internal/service"
	"backend-v2/pkg/response"
	"fmt"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
)

type workerAttendanceController struct {
	workerAttendanceService service.IWorkerAttendanceService
}

type IWorkerAttendanceController interface {
	Import(c *gin.Context)
	GetPaginated(c *gin.Context)
}

func InitWorkerAttendanceController(workerAttendanceService service.IWorkerAttendanceService) IWorkerAttendanceController {
	return &workerAttendanceController{
		workerAttendanceService: workerAttendanceService,
	}
}

func (controller *workerAttendanceController) Import(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Файл не может быть сформирован, проверьте файл: %v", err))
		return
	}

	date := time.Now()
	importFileName := date.Format("2006-01-02 15-04-05") + file.Filename
	importFilePath := filepath.Join("./pkg/excels/temp/", importFileName)
	err = c.SaveUploadedFile(file, importFilePath)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Файл не может быть сохранен на сервере: %v", err))
		return
	}

	projectID := c.GetUint("projectID")
	err = controller.workerAttendanceService.Import(projectID, importFilePath)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	response.ResponseSuccess(c, true)
}

func (controller *workerAttendanceController) GetPaginated(c *gin.Context) {
	projectID := c.GetUint("projectID")

	data, err := controller.workerAttendanceService.GetPaginated(projectID)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Ошибка при обработке данных: %v", err))
		return
	}

	dataCount, err := controller.workerAttendanceService.Count(projectID)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not get the total amount of Materials: %v", err))
		return
	}

	response.ResponsePaginatedData(c, data, dataCount)
}
