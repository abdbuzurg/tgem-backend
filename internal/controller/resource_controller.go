package controller

import (
	"backend-v2/internal/service"
	"backend-v2/pkg/response"
	"fmt"

	"github.com/gin-gonic/gin"
)

type resourceController struct {
	resourceService service.IResourceService
}

func InitResourceController(
	resourceService service.IResourceService,
) IResourceController {
	return &resourceController{
		resourceService: resourceService,
	}
}

type IResourceController interface {
	GetAll(c *gin.Context)
}

func (controller *resourceController) GetAll(c *gin.Context) {
	data, err := controller.resourceService.GetAll()
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутреняя ошибка сервера: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}
