package controller

import (
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

type districtController struct {
	districtService   service.IDistrictService
	userActionService service.IUserActionService
}

func InitDistrictController(
	districtService service.IDistrictService,
	userActionService service.IUserActionService,
) IDistictController {
	return &districtController{
		districtService:   districtService,
		userActionService: userActionService,
	}
}

type IDistictController interface {
	GetAll(c *gin.Context)
	GetPaginated(c *gin.Context)
	GetByID(c *gin.Context)
	Create(c *gin.Context)
	Update(c *gin.Context)
	Delete(c *gin.Context)
}

func (controller *districtController) GetAll(c *gin.Context) {

	userID := c.GetUint("userID")
	projectID := c.GetUint("projectID")

	data, err := controller.districtService.GetAll()
	if err != nil {

		controller.userActionService.Create(model.UserAction{
			UserID:              userID,
			ProjectID:           projectID,
			DateOfAction:        time.Now(),
			ActionURL:           c.Request.URL.Path,
			ActionStatus:        false,
			ActionStatusMessage: useraction.INTERNAL_SERVER_ERROR,
			ActionID:            0,
			ActionType:          "Запрос данных всех районов",
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
		ActionType:          "Запрос данных всех районов",
	})
	response.ResponseSuccess(c, data)
}

func (controller *districtController) GetPaginated(c *gin.Context) {

	userID := c.GetUint("userID")
	projectID := c.GetUint("projectID")

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
			ActionType:          fmt.Sprintf("Запрос данных с разбивкой на страницы для справочника районов: страница %d", page),
		})
		response.ResponseError(c, fmt.Sprintf("Сервер получил неправильные данные: %v", err))
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
			ActionType:          fmt.Sprintf("Запрос данных с разбивкой на страницы для справочника районов: страница %d", page),
		})
		response.ResponseError(c, fmt.Sprintf("Сервер получил неправильные данные: %v", err))
		return

	}

	name := c.DefaultQuery("name", "")
	name, err = url.QueryUnescape(name)
	if err != nil {

		controller.userActionService.Create(model.UserAction{
			UserID:              userID,
			ProjectID:           projectID,
			DateOfAction:        time.Now(),
			ActionURL:           c.Request.URL.Path,
			ActionStatus:        false,
			ActionStatusMessage: fmt.Sprintf(useraction.INCORRECT_PARAMETER, "NAME"),
			ActionID:            0,
			ActionType:          fmt.Sprintf("Запрос данных с разбивкой на страницы для справочника районов: страница %d", page),
		})
		response.ResponseError(c, fmt.Sprintf("Сервер получил неправильные данные: %v", err))
		return

	}

	filter := model.District{
		Name: name,
	}

	data, err := controller.districtService.GetPaginated(page, limit, filter)
	if err != nil {

		controller.userActionService.Create(model.UserAction{
			UserID:              userID,
			ProjectID:           projectID,
			DateOfAction:        time.Now(),
			ActionURL:           c.Request.URL.Path,
			ActionStatus:        false,
			ActionStatusMessage: useraction.INTERNAL_SERVER_ERROR,
			ActionID:            0,
			ActionType:          fmt.Sprintf("Запрос данных с разбивкой на страницы для справочника районов: страница %d", page),
		})
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return

	}

	dataCount, err := controller.districtService.Count()
	if err != nil {

		controller.userActionService.Create(model.UserAction{
			UserID:              userID,
			ProjectID:           projectID,
			DateOfAction:        time.Now(),
			ActionURL:           c.Request.URL.Path,
			ActionStatus:        false,
			ActionStatusMessage: useraction.INTERNAL_SERVER_ERROR,
			ActionID:            0,
			ActionType:          fmt.Sprintf("Запрос данных с разбивкой на страницы для справочника районов: страница %d", page),
		})
		response.ResponseError(c,  fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
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
		ActionType:          fmt.Sprintf("Запрос данных с разбивкой на страницы для справочника районов: страница %d", page),
	})
	response.ResponsePaginatedData(c, data, dataCount)

}

func (controller *districtController) GetByID(c *gin.Context) {

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
			ActionType:          fmt.Sprintf("Запрос на одиночный поиск данных из справочника районов: id = %d", id),
		})
		response.ResponseError(c, fmt.Sprintf("Сервер получил неправильные данные: %v", err))
		return

	}

	data, err := controller.districtService.GetByID(uint(id))
	if err != nil {

		controller.userActionService.Create(model.UserAction{
			UserID:              userID,
			ProjectID:           projectID,
			DateOfAction:        time.Now(),
			ActionURL:           c.Request.URL.Path,
			ActionStatus:        false,
			ActionStatusMessage: useraction.INTERNAL_SERVER_ERROR,
			ActionID:            0,
			ActionType:          fmt.Sprintf("Запрос на одиночный поиск данных из справочника районов: id = %d", id),
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
		ActionType:          fmt.Sprintf("Запрос на одиночный поиск данных из справочника районов: id = %d", id),
	})
	response.ResponseSuccess(c, data)

}

func (controller *districtController) Create(c *gin.Context) {

	userID := c.GetUint("userID")
	projectID := c.GetUint("projectID")

	var createData model.District
	if err := c.ShouldBindJSON(&createData); err != nil {

		controller.userActionService.Create(model.UserAction{
			UserID:              userID,
			ProjectID:           projectID,
			DateOfAction:        time.Now(),
			ActionURL:           c.Request.URL.Path,
			ActionStatus:        false,
			ActionStatusMessage: useraction.INCORRECT_BODY,
			ActionID:            0,
			ActionType:          "Запрос на создание нового райнона в справочнике районов",
		})
    response.ResponseError(c, fmt.Sprintf("Неверное тело запроса: %v", err))
		return
	}

	data, err := controller.districtService.Create(createData)
	if err != nil {

		controller.userActionService.Create(model.UserAction{
			UserID:              userID,
			ProjectID:           projectID,
			DateOfAction:        time.Now(),
			ActionURL:           c.Request.URL.Path,
			ActionStatus:        false,
			ActionStatusMessage: useraction.INTERNAL_SERVER_ERROR,
			ActionID:            0,
			ActionType:          "Запрос на создание нового райнона в справочнике районов",
		})
    response.ResponseError(c, fmt.Sprintf("Внутрення ошибка сервера: %v", err))
		return
	}

  controller.userActionService.Create(model.UserAction{
    UserID:              userID,
    ProjectID:           projectID,
    DateOfAction:        time.Now(),
    ActionURL:           c.Request.URL.Path,
    ActionStatus:        true,
    ActionStatusMessage: useraction.POST_SUCCESS,
    ActionID:            data.ID,
    ActionType:          "Запрос на создание нового райнона в справочнике районов",
  })
	response.ResponseSuccess(c, data)

}

func (controller *districtController) Update(c *gin.Context) {

  projectID := c.GetUint("projectID")
  userID := c.GetUint("userID")

	var updateData model.District
	if err := c.ShouldBindJSON(&updateData); err != nil {

    controller.userActionService.Create(model.UserAction{
      UserID:              userID,
      ProjectID:           projectID,
      DateOfAction:        time.Now(),
      ActionURL:           c.Request.URL.Path,
      ActionStatus:        false,
      ActionStatusMessage: useraction.INCORRECT_BODY,
      ActionID:            0,
      ActionType:          "Запрос на изменение данных одного райнона в справочнике районов",
    })
    response.ResponseError(c, fmt.Sprintf("Неверное тело запроса: %v", err))
		return
	}

	data, err := controller.districtService.Update(updateData)
	if err != nil {

    controller.userActionService.Create(model.UserAction{
      UserID:              userID,
      ProjectID:           projectID,
      DateOfAction:        time.Now(),
      ActionURL:           c.Request.URL.Path,
      ActionStatus:        false,
      ActionStatusMessage: useraction.INTERNAL_SERVER_ERROR,
      ActionID:            updateData.ID,
      ActionType:          "Запрос на изменение данных одного райнона в справочнике районов",
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
    ActionStatusMessage: useraction.PATCH_SUCCESS,
    ActionID:            data.ID,
    ActionType:          "Запрос на изменение данных одного райнона в справочнике районов",
  })
	response.ResponseSuccess(c, data)

}

func (controller *districtController) Delete(c *gin.Context) {
	
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
      ActionType:          "Запрос на удаление одиночных данных из справочника районов",
    })
    response.ResponseError(c, fmt.Sprintf("Неправильный параметер запроса: %v", err))
		return

	}

	err = controller.districtService.Delete(uint(id))
	if err != nil {

    controller.userActionService.Create(model.UserAction{
      UserID:              userID,
      ProjectID:           projectID,
      DateOfAction:        time.Now(),
      ActionURL:           c.Request.URL.Path,
      ActionStatus:        false,
      ActionStatusMessage: useraction.INTERNAL_SERVER_ERROR,
      ActionID:            uint(id),
      ActionType:          "Запрос на удаление одиночных данных из справочника районов",
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
    ActionType:          "Запрос на удаление одиночных данных из справочника районов",
  })
	response.ResponseSuccess(c, "deleted")
}
