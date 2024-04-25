package controller

import (
	"backend-v2/internal/dto"
	"backend-v2/internal/service"
	"backend-v2/model"
	"backend-v2/pkg/response"
	"backend-v2/pkg/useraction"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type invoiceInputController struct {
	invoiceInputService service.IInvoiceInputService
	userActionService   service.IUserActionService
}

func InitInvoiceInputController(
	invoiceInputService service.IInvoiceInputService,
	userActionService service.IUserActionService,
) IInvoiceInputController {
	return &invoiceInputController{
		invoiceInputService: invoiceInputService,
		userActionService:   userActionService,
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
  NewMaterial(c *gin.Context)
  NewMaterialCost(c *gin.Context)
}

func (controller *invoiceInputController) GetAll(c *gin.Context) {

	projectID := c.GetUint("projectID")
	userID := c.GetUint("userID")

	data, err := controller.invoiceInputService.GetAll()
	if err != nil {
		controller.userActionService.Create(model.UserAction{
			UserID:              userID,
			ProjectID:           projectID,
			DateOfAction:        time.Now(),
			ActionURL:           c.Request.URL.Path,
			ActionStatus:        false,
			ActionStatusMessage: useraction.INTERNAL_SERVER_ERROR,
			ActionID:            0,
			ActionType:          "Запрос всех данных накладной приход",
		})
		response.ResponseError(c, fmt.Sprintf("Could not get Invoice Input data: %v", err))
		return
	}

	controller.userActionService.Create(model.UserAction{
		UserID:              userID,
		ProjectID:           projectID,
		DateOfAction:        time.Now(),
		ActionURL:           c.Request.URL.Path,
		ActionStatus:        true,
		ActionStatusMessage: useraction.GET_SUCCESS,
		ActionID:            0,
		ActionType:          "Запрос всех данных накладной приход",
	})
	response.ResponseSuccess(c, data)

}

func (controller *invoiceInputController) GetPaginated(c *gin.Context) {

	projectID := c.GetUint("projectID")
	userID := c.GetUint("userID")

	pageStr := c.DefaultQuery("page", "1")
	page, err := strconv.Atoi(pageStr)
	if err != nil || page <= 0 {

		controller.userActionService.Create(model.UserAction{
			UserID:              userID,
			ProjectID:           projectID,
			DateOfAction:        time.Now(),
			ActionURL:           c.Request.URL.Path,
			ActionStatus:        false,
			ActionStatusMessage: fmt.Sprintf(useraction.INCORRECT_PARAMETER, "PAGE"),
			ActionID:            0,
			ActionType:          fmt.Sprintf("Запрос данных с разбивкой на страницы для накладной приход: страница %d", page),
		})
		response.ResponseError(c, fmt.Sprintf("Неверное тело запроса: %v", err))
		return

	}

	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {

		controller.userActionService.Create(model.UserAction{
			UserID:              userID,
			ProjectID:           projectID,
			DateOfAction:        time.Now(),
			ActionURL:           c.Request.URL.Path,
			ActionStatus:        false,
			ActionStatusMessage: fmt.Sprintf(useraction.INCORRECT_PARAMETER, "LIMIT"),
			ActionID:            0,
			ActionType:          fmt.Sprintf("Запрос данных с разбивкой на страницы для накладной приход: страница %d", page),
		})
		response.ResponseError(c, fmt.Sprintf("Неверное тело запроса: %v", err))
		return

	}

	warehouseManagerWorkerIDStr := c.DefaultQuery("warehouseManagerWorkerID", "")
	warehouseManagerWorkerID := 0
	if warehouseManagerWorkerIDStr != "" {

		warehouseManagerWorkerID, err = strconv.Atoi(warehouseManagerWorkerIDStr)
		if err != nil {

			controller.userActionService.Create(model.UserAction{
				UserID:              userID,
				ProjectID:           projectID,
				DateOfAction:        time.Now(),
				ActionURL:           c.Request.URL.Path,
				ActionStatus:        false,
				ActionStatusMessage: fmt.Sprintf(useraction.INCORRECT_PARAMETER, "WAREHOUSEMANAGER_WORKER_ID"),
				ActionID:            0,
				ActionType:          fmt.Sprintf("Запрос данных с разбивкой на страницы для накладной приход: страница %d", page),
			})
			response.ResponseError(c, fmt.Sprintf("Неверное тело запроса: %v", err))
			return

		}

	}

	releasedWorkerIDStr := c.DefaultQuery("releasedWorkerID", "")
	releasedWorkerID := 0
	if releasedWorkerIDStr != "" {

		releasedWorkerID, err = strconv.Atoi(releasedWorkerIDStr)
		if err != nil {

			controller.userActionService.Create(model.UserAction{
				UserID:              userID,
				ProjectID:           projectID,
				DateOfAction:        time.Now(),
				ActionURL:           c.Request.URL.Path,
				ActionStatus:        false,
				ActionStatusMessage: fmt.Sprintf(useraction.INCORRECT_PARAMETER, "RELEASED_WORKER_ID"),
				ActionID:            0,
				ActionType:          fmt.Sprintf("Запрос данных с разбивкой на страницы для накладной приход: страница %d", page),
			})
			response.ResponseError(c, fmt.Sprintf("Неверное тело запроса: %v", err))
			return

		}

	}

	deliveryCode := c.DefaultQuery("deliveryCode", "")
	deliveryCode, err = url.QueryUnescape(deliveryCode)
	if err != nil {

		controller.userActionService.Create(model.UserAction{
			UserID:              userID,
			ProjectID:           projectID,
			DateOfAction:        time.Now(),
			ActionURL:           c.Request.URL.Path,
			ActionStatus:        false,
			ActionStatusMessage: fmt.Sprintf(useraction.INCORRECT_PARAMETER, "DELIVERY_CODE"),
			ActionID:            0,
			ActionType:          fmt.Sprintf("Запрос данных с разбивкой на страницы для накладной приход: страница %d", page),
		})
		response.ResponseError(c, fmt.Sprintf("Неверное тело запроса: %v", err))
		return

	}

	filter := model.InvoiceInput{
		WarehouseManagerWorkerID: uint(warehouseManagerWorkerID),
		ReleasedWorkerID:         uint(releasedWorkerID),
		ProjectID:                projectID,
		DeliveryCode:             deliveryCode,
	}

	data, err := controller.invoiceInputService.GetPaginated(page, limit, filter)
	if err != nil {

		controller.userActionService.Create(model.UserAction{
			UserID:              userID,
			ProjectID:           projectID,
			DateOfAction:        time.Now(),
			ActionURL:           c.Request.URL.Path,
			ActionStatus:        false,
			ActionStatusMessage: useraction.INTERNAL_SERVER_ERROR,
			ActionID:            0,
			ActionType:          fmt.Sprintf("Запрос данных с разбивкой на страницы для накладной приход: страница %d", page),
		})
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return

	}

	dataCount, err := controller.invoiceInputService.Count(projectID)
	if err != nil {

		controller.userActionService.Create(model.UserAction{
			UserID:              userID,
			ProjectID:           projectID,
			DateOfAction:        time.Now(),
			ActionURL:           c.Request.URL.Path,
			ActionStatus:        false,
			ActionStatusMessage: useraction.INTERNAL_SERVER_ERROR,
			ActionID:            0,
			ActionType:          fmt.Sprintf("Запрос данных с разбивкой на страницы для накладной приход: страница %d", page),
		})
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return

	}

	controller.userActionService.Create(model.UserAction{
		UserID:              userID,
		ProjectID:           projectID,
		DateOfAction:        time.Now(),
		ActionURL:           c.Request.URL.Path,
		ActionStatus:        true,
		ActionStatusMessage: useraction.GET_SUCCESS,
		ActionID:            0,
		ActionType:          fmt.Sprintf("Запрос данных с разбивкой на страницы для накладной приход: страница %d", page),
	})
	response.ResponsePaginatedData(c, data, dataCount)

}

func (controller *invoiceInputController) Create(c *gin.Context) {

	workerID := c.GetUint("workerID")
	projectID := c.GetUint("projectID")
	userID := c.GetUint("userID")

	var createData dto.InvoiceInput
	if err := c.ShouldBindJSON(&createData); err != nil {

		controller.userActionService.Create(model.UserAction{
			UserID:              userID,
			ProjectID:           projectID,
			DateOfAction:        time.Now(),
			ActionURL:           c.Request.URL.Path,
			ActionStatus:        false,
			ActionStatusMessage: useraction.INCORRECT_BODY,
			ActionID:            0,
			ActionType:          "Запрос на создание накладной приход",
		})
		response.ResponseError(c, fmt.Sprintf("Неверное тело запроса: %v", err))
		return

	}

	createData.Details.ProjectID = projectID
	createData.Details.ReleasedWorkerID = workerID

	data, err := controller.invoiceInputService.Create(createData)
	if err != nil {

		controller.userActionService.Create(model.UserAction{
			UserID:              userID,
			ProjectID:           projectID,
			DateOfAction:        time.Now(),
			ActionURL:           c.Request.URL.Path,
			ActionStatus:        false,
			ActionStatusMessage: useraction.INTERNAL_SERVER_ERROR,
			ActionID:            0,
			ActionType:          "Запрос на создание накладной приход",
		})
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return

	}

	controller.userActionService.Create(model.UserAction{
		UserID:              userID,
		ProjectID:           projectID,
		DateOfAction:        time.Now(),
		ActionURL:           c.Request.URL.Path,
		ActionStatus:        true,
		ActionStatusMessage: useraction.POST_SUCCESS,
		ActionID:            data.Details.ID,
		ActionType:          "Запрос на создание накладной приход",
	})
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

	projectID := c.GetUint("projectID")
	userID := c.GetUint("userID")

	idRaw := c.Param("id")
	id, err := strconv.ParseUint(idRaw, 10, 64)
	if err != nil {

		controller.userActionService.Create(model.UserAction{
			UserID:              userID,
			ProjectID:           projectID,
			DateOfAction:        time.Now(),
			ActionURL:           c.Request.URL.Path,
			ActionStatus:        false,
			ActionStatusMessage: fmt.Sprintf(useraction.INCORRECT_PARAMETER, "ID"),
			ActionID:            0,
			ActionType:          "Запрос на удаление накладной приход",
		})
		response.ResponseError(c, fmt.Sprintf("Неверное тело запроса: %v", err))
		return

	}

	err = controller.invoiceInputService.Delete(uint(id))
	if err != nil {

		controller.userActionService.Create(model.UserAction{
			UserID:              userID,
			ProjectID:           projectID,
			DateOfAction:        time.Now(),
			ActionURL:           c.Request.URL.Path,
			ActionStatus:        false,
			ActionStatusMessage: useraction.INTERNAL_SERVER_ERROR,
			ActionID:            uint(id),
			ActionType:          "Запрос на удаление накладной приход",
		})
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return

	}
	controller.userActionService.Create(model.UserAction{
		UserID:              userID,
		ProjectID:           projectID,
		DateOfAction:        time.Now(),
		ActionURL:           c.Request.URL.Path,
		ActionStatus:        true,
		ActionStatusMessage: useraction.DELETE_SUCCESS,
		ActionID:            uint(id),
		ActionType:          "Запрос на удаление накладной приход",
	})
	response.ResponseSuccess(c, "deleted")
}

func (controller *invoiceInputController) Confirmation(c *gin.Context) {

	userID := c.GetUint("userID")
	projectID := c.GetUint("projectID")

	idRaw := c.Param("id")
	id, err := strconv.ParseUint(idRaw, 10, 64)
	if err != nil {

		controller.userActionService.Create(model.UserAction{
			UserID:              userID,
			ProjectID:           projectID,
			DateOfAction:        time.Now(),
			ActionURL:           c.Request.URL.Path,
			ActionStatus:        false,
			ActionStatusMessage: fmt.Sprintf(useraction.INCORRECT_PARAMETER, "ID"),
			ActionID:            0,
			ActionType:          "Запрос на подтверждение накладной приход",
		})
		response.ResponseError(c, fmt.Sprintf("Неверное тело запроса: %v", err))
		return

	}

	invoiceInput, err := controller.invoiceInputService.GetByID(uint(id))
	if err != nil {

		controller.userActionService.Create(model.UserAction{
			UserID:              userID,
			ProjectID:           projectID,
			DateOfAction:        time.Now(),
			ActionURL:           c.Request.URL.Path,
			ActionStatus:        false,
			ActionStatusMessage: useraction.INTERNAL_SERVER_ERROR,
			ActionID:            uint(id),
			ActionType:          "Запрос на подтверждение накладной приход",
		})
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return

	}

	file, err := c.FormFile("file")
	if err != nil {

		controller.userActionService.Create(model.UserAction{
			UserID:              userID,
			ProjectID:           projectID,
			DateOfAction:        time.Now(),
			ActionURL:           c.Request.URL.Path,
			ActionStatus:        false,
			ActionStatusMessage: "Ошибка при получении файла",
			ActionID:            uint(id),
			ActionType:          "Запрос на подтверждение накладной приход",
		})
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return

	}

	file.Filename = invoiceInput.DeliveryCode
	filePath := "./pkg/excels/input/" + file.Filename + ".xlsx"
	err = c.SaveUploadedFile(file, filePath)
	if err != nil {

		controller.userActionService.Create(model.UserAction{
			UserID:              userID,
			ProjectID:           projectID,
			DateOfAction:        time.Now(),
			ActionURL:           c.Request.URL.Path,
			ActionStatus:        false,
			ActionStatusMessage: "Ошибка при сохранения файла на сервер",
			ActionID:            uint(id),
			ActionType:          "Запрос на подтверждение накладной приход",
		})
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return

	}

	err = controller.invoiceInputService.Confirmation(uint(id), projectID)
	if err != nil {

		controller.userActionService.Create(model.UserAction{
			UserID:              userID,
			ProjectID:           projectID,
			DateOfAction:        time.Now(),
			ActionURL:           c.Request.URL.Path,
			ActionStatus:        false,
			ActionStatusMessage: "Ошибка при привязки(подтверждении) файла к накладной",
			ActionID:            uint(id),
			ActionType:          "Запрос на подтверждение накладной приход",
		})
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return

	}

	controller.userActionService.Create(model.UserAction{
		UserID:              userID,
		ProjectID:           projectID,
		DateOfAction:        time.Now(),
		ActionURL:           c.Request.URL.Path,
		ActionStatus:        true,
		ActionStatusMessage: "Файл успешной привязан к накладой",
		ActionID:            uint(id),
		ActionType:          "Запрос на подтверждение накладной приход",
	})
	response.ResponseSuccess(c, true)
}

func (controller *invoiceInputController) GetDocument(c *gin.Context) {

	projectID := c.GetUint("projectID")
	userID := c.GetUint("userID")

	deliveryCode := c.Param("deliveryCode")

	controller.userActionService.Create(model.UserAction{
		UserID:              userID,
		ProjectID:           projectID,
		DateOfAction:        time.Now(),
		ActionURL:           c.Request.URL.Path,
		ActionStatus:        true,
		ActionStatusMessage: "Получение файла накладной приход с Кодом " + deliveryCode,
		ActionID:            0,
		ActionType:          "Запрос на получение файла",
	})
	c.FileAttachment("./pkg/excels/input/"+deliveryCode+".xlsx", deliveryCode+".xlsx")

}

func (controller *invoiceInputController) UniqueCode(c *gin.Context) {

	projectID := c.GetUint("projectID")
	userID := c.GetUint("userID")

	data, err := controller.invoiceInputService.UniqueCode(projectID)
	if err != nil {

		controller.userActionService.Create(model.UserAction{
			UserID:              userID,
			ProjectID:           projectID,
			DateOfAction:        time.Now(),
			ActionURL:           c.Request.URL.Path,
			ActionStatus:        false,
			ActionStatusMessage: useraction.INTERNAL_SERVER_ERROR,
			ActionID:            0,
			ActionType:          "Запрос на получение уникальных кодов накладной приход",
		})
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return

	}

	controller.userActionService.Create(model.UserAction{
		UserID:              userID,
		ProjectID:           projectID,
		DateOfAction:        time.Now(),
		ActionURL:           c.Request.URL.Path,
		ActionStatus:        true,
		ActionStatusMessage: useraction.GET_SUCCESS,
		ActionID:            0,
		ActionType:          "Запрос на получение уникальных кодов накладной приход",
	})
	response.ResponseSuccess(c, data)
}

func (controller *invoiceInputController) UniqueWarehouseManager(c *gin.Context) {

	projectID := c.GetUint("projectID")
	userID := c.GetUint("userID")

	data, err := controller.invoiceInputService.UniqueWarehouseManager(projectID)
	if err != nil {

		controller.userActionService.Create(model.UserAction{
			UserID:              userID,
			ProjectID:           projectID,
			DateOfAction:        time.Now(),
			ActionURL:           c.Request.URL.Path,
			ActionStatus:        false,
			ActionStatusMessage: useraction.INTERNAL_SERVER_ERROR,
			ActionID:            0,
			ActionType:          "Запрос на получение уникальных заведующих складом присутсвующих в накладой приход",
		})
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return

	}

	controller.userActionService.Create(model.UserAction{
		UserID:              userID,
		ProjectID:           projectID,
		DateOfAction:        time.Now(),
		ActionURL:           c.Request.URL.Path,
		ActionStatus:        true,
		ActionStatusMessage: useraction.GET_SUCCESS,
		ActionID:            0,
		ActionType:          "Запрос на получение уникальных заведующих складом присутсвующих в накладой приход",
	})
	response.ResponseSuccess(c, data)

}

func (controller *invoiceInputController) UniqueReleased(c *gin.Context) {

	projectID := c.GetUint("projectID")
	userID := c.GetUint("userID")

	data, err := controller.invoiceInputService.UniqueReleased(projectID)
	if err != nil {

		controller.userActionService.Create(model.UserAction{
			UserID:              userID,
			ProjectID:           projectID,
			DateOfAction:        time.Now(),
			ActionURL:           c.Request.URL.Path,
			ActionStatus:        false,
			ActionStatusMessage: useraction.INTERNAL_SERVER_ERROR,
			ActionID:            0,
			ActionType:          "Запрос на получение уникальных составителей присутствующих в накладной приход",
		})
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return

	}

	controller.userActionService.Create(model.UserAction{
		UserID:              userID,
		ProjectID:           projectID,
		DateOfAction:        time.Now(),
		ActionURL:           c.Request.URL.Path,
		ActionStatus:        true,
		ActionStatusMessage: useraction.GET_SUCCESS,
		ActionID:            0,
		ActionType:          "Запрос на получение уникальных составителей присутствующих в накладной приход",
	})
	response.ResponseSuccess(c, data)
}

func (controller *invoiceInputController) Report(c *gin.Context) {

	projectID := c.GetUint("projectID")
	userID := c.GetUint("userID")

	var filter dto.InvoiceInputReportFilterRequest
	if err := c.ShouldBindJSON(&filter); err != nil {

		controller.userActionService.Create(model.UserAction{
			UserID:              userID,
			ProjectID:           projectID,
			DateOfAction:        time.Now(),
			ActionURL:           c.Request.URL.Path,
			ActionStatus:        false,
			ActionStatusMessage: useraction.INCORRECT_BODY,
			ActionID:            0,
			ActionType:          "Запрос на получение отсчетного файла накладной приход",
		})
		response.ResponseError(c, fmt.Sprintf("Неверное тело запроса: %v", err))
		return

	}

  filter.ProjectID = projectID
	filename, err := controller.invoiceInputService.Report(filter, projectID)
	if err != nil {

		controller.userActionService.Create(model.UserAction{
			UserID:              userID,
			ProjectID:           projectID,
			DateOfAction:        time.Now(),
			ActionURL:           c.Request.URL.Path,
			ActionStatus:        false,
			ActionStatusMessage: useraction.INTERNAL_SERVER_ERROR,
			ActionID:            0,
			ActionType:          "Запрос на получение отсчетного файла накладной приход",
		})
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return

	}

	controller.userActionService.Create(model.UserAction{
		UserID:              userID,
		ProjectID:           projectID,
		DateOfAction:        time.Now(),
		ActionURL:           c.Request.URL.Path,
		ActionStatus:        true,
		ActionStatusMessage: useraction.INTERNAL_SERVER_ERROR,
		ActionID:            0,
		ActionType:          "Запрос на получение отсчетного файла накладной приход",
	})
	c.FileAttachment("./pkg/excels/report/"+filename, filename)
	// response.ResponseSuccess(c, true)
}

func(controller *invoiceInputController) NewMaterial(c *gin.Context) {
  var data dto.NewMaterialDataFromInvoiceInput
  if err := c.ShouldBindJSON(&data); err != nil {
    response.ResponseError(c, fmt.Sprintf("Неверное тело запроса %v", err))
    return
  }

  projectID := c.GetUint("projectID")
  data.ProjectID = projectID

  err := controller.invoiceInputService.NewMaterialAndItsCost(data)
  if err != nil {
    response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
    return
  }

  response.ResponseSuccess(c, true)
}

func(controller *invoiceInputController) NewMaterialCost(c *gin.Context) {
  var data model.MaterialCost
  if err := c.ShouldBindJSON(&data); err != nil {
    response.ResponseError(c, fmt.Sprintf("Неверное тело запроса %v", err))
    return
  }

  err := controller.invoiceInputService.NewMaterialCost(data)
  if err != nil {
    response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
    return
  }

  response.ResponseSuccess(c, true)
}
