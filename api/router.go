package api

import (
	"backend-v2/api/middleware"
	"backend-v2/internal/controller"
	"backend-v2/internal/repository"
	"backend-v2/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupRouter(db *gorm.DB) *gin.Engine {

	router := gin.Default()

	router.Use(gin.Recovery())

	router.Use(middleware.CORSMiddleware())

	//Initialization of Repositories
	invoiceMaterialRepo := repository.InitInvoiceMaterialsRepository(db)
	invoiceRepo := repository.InitInvoiceRepository(db)
	// kl04kvObjectRepo := repository.InitKL04KVObjectRepository(db)
	materialCostRepo := repository.InitMaterialCostRepository(db)
	// materialForProjectRepo := repository.InitMaterialForProjectRepository(db)
	materialLocationRepo := repository.InitMaterialLocationRepository(db)
	materialRepo := repository.InitMaterialRepository(db)
	// mjdObjectRepo := repository.InitMJDObjectRepository(db)
	// objectOperationRepo := repository.InitObjectOperationRepository(db)
	objectRepo := repository.InitObjectRepository(db)
	// operationRepo := repository.InitOperationRepository(db)
	projectRepo := repository.InitProjectRepository(db)
	// sipObjectRepo := repository.InitSIPObjectRepository(db)
	// stvtObjectRepo := repository.InitSTVTObjectRepository(db)
	teamRepo := repository.InitTeamRepostory(db)
	// tpObjectRepo := repository.InitTPObjectRepository(db)
	userRepo := repository.InitUserRepository(db)
	workerRepo := repository.InitWorkerRepository(db)

	//Initialization of Services
	// invoiceMaterialsService := service.InitInvoiceMaterialsService(invoiceMaterialRepo)
	invoiceService := service.InitInvoiceService(
		projectRepo,
		invoiceRepo,
		invoiceMaterialRepo,
		objectRepo,
		materialCostRepo,
		materialLocationRepo,
		materialRepo,
		workerRepo,
		teamRepo,
	)
	// kl04kvObjectService := service.InitKL04KVObjectService(kl04kvObjectRepo)
	// materialCostService := service.InitMaterialCostService(materialCostRepo)
	// materialForProjectService := service.InitMaterialForProjectService(materialForProjectRepo)
	// materialLocationService := service.InitMaterialLocationService(materialLocationRepo)
	// materialService := service.InitMaterialService(materialRepo)
	// mjdObjectService := service.InitMJDObjectService(mjdObjectRepo)
	// objectOperationService := service.InitObjectOperationService(objectOperationRepo)
	// objectService := service.InitObjectService(objectRepo)
	// operationService := service.InitOperationService(operationRepo)
	// projectService := service.InitProjectService(projectRepo)
	// sipObjectService := service.InitSIPObjectService(sipObjectRepo)
	// stvtObjectService := service.InitSTVTObjectService(stvtObjectRepo)
	// teamService := service.InitTeamService(teamRepo)
	// tpObjctService := service.InitTPObjectService(tpObjectRepo)
	userService := service.InitUserService(userRepo)
	// workerService := service.InitWorkerService(workerRepo)

	//Initialization of Controllers
	invoiceController := controller.InitInvoiceController(invoiceService)
	// invoiceMaterialController := controller.InitInvoiceMaterialsController(invoiceMaterialsService)
	// materialController := controller.InitMaterialController(materialService)
	// materialCostController := controller.InitMaterialCostController(materialCostService)
	// materialForProjectController := controller.InitMaterialForProjectController(materialForProjectService)
	// materialLocationController := controller.InitMaterialLocationController(materialLocationService)
	// objectController := controller.InitObjectController(objectService)
	// objectOperationController := controller.InitObjectOperationController(objectOperationService)
	// operationController := controller.InitOperationController(operationService)
	// projectController := controller.InitProjectController(projectService)
	// teamController := controller.InitTeamController(teamService)
	userController := controller.InitUserController(userService)
	// workerController := controller.InitWorkerController(workerService)

	//Initialization of Routes
	InitInvoiceRoutes(router, invoiceController)
	InitUserRoutes(router, userController)

	return router

}

func InitInvoiceRoutes(router *gin.Engine, controller controller.IInvoiceController) {
	invoiceRoutes := router.Group("/invoice")
	invoiceRoutes.GET("/:invoiceType/all", controller.GetAll)
	invoiceRoutes.GET("/:invoiceType/pagniated", controller.GetPaginated)
	invoiceRoutes.GET("/:id", controller.GetByID)
	invoiceRoutes.POST("/", controller.Create)
	invoiceRoutes.PATCH("/", controller.Update)
	invoiceRoutes.DELETE("/:id", controller.Delete)
}

func InitUserRoutes(router *gin.Engine, controller controller.IUserController) {
	userRoutes := router.Group("/user")
	userRoutes.GET("/all", controller.GetAll)
	userRoutes.GET("/:id", controller.GetByID)
	userRoutes.GET("/paginated", controller.GetPaginated)
	userRoutes.GET("/is-authenticated", controller.IsAuthenticated)
	userRoutes.GET("/permissions", controller.GetPermissions)
	userRoutes.POST("/", controller.Create)
	userRoutes.POST("/login", controller.Login)
	userRoutes.PATCH("/", controller.Update)
	userRoutes.DELETE("/:id", controller.Delete)
}
