package api

import (
	"backend-v2/api/middleware"
	"backend-v2/internal/controller"
	"backend-v2/internal/repository"
	"backend-v2/internal/service"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupRouter(db *gorm.DB) *gin.Engine {

	router := gin.Default()

	router.Use(gin.Recovery())

	router.Use(cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin, Content-Type, Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		AllowFiles:       true,
		MaxAge:           12 * time.Hour,
	}))

	//Initialization of Repositories
	invoiceInputRepo := repository.InitInvoiceInputRepository(db)
	invoiceOutputRepo := repository.InitInvoiceOutputRepository(db)
	invoiceReturnRepo := repository.InitInvoiceReturnRepository(db)
	invoiceMaterialRepo := repository.InitInvoiceMaterialsRepository(db)
	kl04kvObjectRepo := repository.InitKL04KVObjectRepository(db)
	materialCostRepo := repository.InitMaterialCostRepository(db)
	materialLocationRepo := repository.InitMaterialLocationRepository(db)
	materialRepo := repository.InitMaterialRepository(db)
	mjdObjectRepo := repository.InitMJDObjectRepository(db)
	// objectOperationRepo := repository.InitObjectOperationRepository(db)
	objectRepo := repository.InitObjectRepository(db)
	// operationRepo := repository.InitOperationRepository(db)
	projectRepo := repository.InitProjectRepository(db)
	sipObjectRepo := repository.InitSIPObjectRepository(db)
	stvtObjectRepo := repository.InitSTVTObjectRepository(db)
	teamRepo := repository.InitTeamRepostory(db)
	tpObjectRepo := repository.InitTPObjectRepository(db)
	userRepo := repository.InitUserRepository(db)
	userInProjects := repository.InitUserInProjectRepository(db)
	workerRepo := repository.InitWorkerRepository(db)
	serialNumberRepo := repository.InitSerialNumberRepository(db)
	districtRepo := repository.InitDistrictRepository(db)
	permissionRepo := repository.InitPermissionRepository(db)
	roleRepo := repository.InitRoleRepository(db)
  materialDefect := repository.InitMaterialDefectRepository(db)
  userActionRepo := repository.InitUserActionRepository(db)
  supervisorObjectsRepo := repository.InitSupervisorObjectsRepository(db)

	//Initialization of Services
	invoiceInputService := service.InitInvoiceInputService(
		invoiceInputRepo,
		invoiceMaterialRepo,
		materialLocationRepo,
		workerRepo,
		materialCostRepo,
		materialRepo,
		serialNumberRepo,
	)
	invoiceOutputService := service.InitInvoiceOutputService(
		invoiceOutputRepo,
		invoiceMaterialRepo,
		workerRepo,
		teamRepo,
		objectRepo,
		materialCostRepo,
		materialLocationRepo,
		materialRepo,
		districtRepo,
		serialNumberRepo,
	)
	invoiceReturnService := service.InitInvoiceReturnService(
		invoiceReturnRepo,
		workerRepo,
		objectRepo,
		teamRepo,
		materialLocationRepo,
		invoiceMaterialRepo,
		materialRepo,
		materialCostRepo,
	)
	// invoiceMaterialsService := service.InitInvoiceMaterialsService(invoiceMaterialRepo)
	// kl04kvObjectService := service.InitKL04KVObjectService(kl04kvObjectRepo)
	materialCostService := service.InitMaterialCostService(materialCostRepo)
	// materialForProjectService := service.InitMaterialForProjectService(materialForProjectRepo)
	materialLocationService := service.InitMaterialLocationService(
		materialLocationRepo,
		materialCostRepo,
		materialRepo,
    teamRepo,
    objectRepo,
	  materialDefect,
  )

	materialService := service.InitMaterialService(materialRepo)
	// mjdObjectService := service.InitMJDObjectService(mjdObjectRepo)
	// objectOperationService := service.InitObjectOperationService(objectOperationRepo)
	objectService := service.InitObjectService(
    objectRepo, 
    supervisorObjectsRepo,
    kl04kvObjectRepo,
		mjdObjectRepo,
		sipObjectRepo,
		stvtObjectRepo,
		tpObjectRepo,
  )
	// operationService := service.InitOperationService(operationRepo)
	projectService := service.InitProjectService(projectRepo)
	// sipObjectService := service.InitSIPObjectService(sipObjectRepo)
	// stvtObjectService := service.InitSTVTObjectService(stvtObjectRepo)
	teamService := service.InitTeamService(teamRepo)
	// tpObjctService := service.InitTPObjectService(tpObjectRepo)
	userService := service.InitUserService(
		userRepo,
		userInProjects,
		workerRepo,
		roleRepo,
    userInProjects,
	)
	workerService := service.InitWorkerService(workerRepo)
	districtService := service.InitDistrictService(districtRepo)
	permissionService := service.InitPermissionService(
		permissionRepo,
		roleRepo,
	)
	roleService := service.InitRoleService(roleRepo)
  userActionService := service.InitUserActionService(userActionRepo, userRepo)

	//Initialization of Controllers
	invoiceInputController := controller.InitInvoiceInputController(invoiceInputService, userActionService)
	invoiceOutputController := controller.InitInvoiceOutputController(invoiceOutputService)
	invoiceReturnController := controller.InitInvoiceReturnController(invoiceReturnService)
	// invoiceMaterialController := controller.InitInvoiceMaterialsController(invoiceMaterialsService)
	materialController := controller.InitMaterialController(materialService)
	materialCostController := controller.InitMaterialCostController(materialCostService)
	// materialForProjectController := controller.InitMaterialForProjectController(materialForProjectService)
	materialLocationController := controller.InitMaterialLocationController(materialLocationService)
	objectController := controller.InitObjectController(objectService)
	// objectOperationController := controller.InitObjectOperationController(objectOperationService)
	// operationController := controller.InitOperationController(operationService)
	projectController := controller.InitProjectController(projectService)
	teamController := controller.InitTeamController(teamService)
	userController := controller.InitUserController(userService)
	workerController := controller.InitWorkerController(workerService)
	districtController := controller.InitDistrictController(districtService, userActionService)
	permissionController := controller.InitPermissionController(permissionService)
	roleController := controller.InitRoleController(roleService)

	//Initialization of Routes
	InitInvoiceInputRoutes(router, invoiceInputController, db)
	InitInvoiceOutputRoutes(router, invoiceOutputController)
	InitInvoiceReturnRoutes(router, invoiceReturnController, db)
	InitProjectRoutes(router, projectController)
	InitMaterialRoutes(router, materialController)
	InitMaterialLocationRoutes(router, materialLocationController)
	InitTeamRoutes(router, teamController)
	InitObjectRoutes(router, objectController)
	InitWorkerRoutes(router, workerController)
	InitUserRoutes(router, userController)
	InitDistrictRoutes(router, districtController)
	InitMaterialCostRoutes(router, materialCostController)
	InitPermissionRoutes(router, permissionController)
	InitRoleRoutes(router, roleController)

	return router

}

func InitInvoiceReturnRoutes(router *gin.Engine, controller controller.IInvoiceReturnController, db *gorm.DB) {
	invoiceReturnRoutes := router.Group("/return")
	invoiceReturnRoutes.Use(
    middleware.Authentication(),
    middleware.Permission(db),
  )
	invoiceReturnRoutes.GET("/paginated", controller.GetPaginated)
	invoiceReturnRoutes.GET("/unique/code", controller.UniqueCode)
	invoiceReturnRoutes.GET("/unique/team", controller.UniqueTeam)
	invoiceReturnRoutes.GET("/unique/object", controller.UniqueObject)
  invoiceReturnRoutes.GET("/document/:deliveryCode", controller.GetDocument)
  invoiceReturnRoutes.POST("/confirm/:id", controller.Confirmation)
	invoiceReturnRoutes.POST("/", controller.Create)
  invoiceReturnRoutes.POST("/report", controller.Report)
	invoiceReturnRoutes.DELETE("/:id", controller.Delete)
}

func InitInvoiceOutputRoutes(router *gin.Engine, controller controller.IInvoiceOutputController) {
	invoiceOutputRoutes := router.Group("/output")
	invoiceOutputRoutes.Use(middleware.Authentication())
	invoiceOutputRoutes.GET("/paginated", controller.GetPaginated)
	invoiceOutputRoutes.GET("/unique/district", controller.UniqueDistrict)
	invoiceOutputRoutes.GET("/unique/code", controller.UniqueCode)
	invoiceOutputRoutes.GET("/unique/recieved", controller.UniqueRecieved)
	invoiceOutputRoutes.GET("/unique/warehouse-manager", controller.UniqueRecieved)
	invoiceOutputRoutes.GET("/unique/object", controller.UniqueObject)
	invoiceOutputRoutes.GET("/unique/team", controller.UniqueTeam)
  invoiceOutputRoutes.GET("/document/:deliveryCode", controller.GetDocument)
  invoiceOutputRoutes.POST("/report", controller.Report)
  invoiceOutputRoutes.POST("/confirm/:id", controller.Confirmation)
	invoiceOutputRoutes.POST("/", controller.Create)
	invoiceOutputRoutes.DELETE("/:id", controller.Delete)
}

func InitInvoiceInputRoutes(router *gin.Engine, controller controller.IInvoiceInputController, db *gorm.DB) {
	invoiceInputRoutes := router.Group("/input")
	invoiceInputRoutes.Use(
    middleware.Authentication(),
    middleware.Permission(db),
  )

	invoiceInputRoutes.GET("/paginated", controller.GetPaginated)
	invoiceInputRoutes.GET("/unique/code", controller.UniqueCode)
	invoiceInputRoutes.GET("/unique/warehouse-manager", controller.UniqueWarehouseManager)
	invoiceInputRoutes.GET("/unique/released", controller.UniqueReleased)
  invoiceInputRoutes.GET("/document/:deliveryCode")
	invoiceInputRoutes.POST("/", controller.Create)
  invoiceInputRoutes.POST("/report", controller.Report)
  invoiceInputRoutes.POST("/confirm/:id", controller.Confirmation)
	invoiceInputRoutes.DELETE("/:id", controller.Delete)
}

func InitProjectRoutes(router *gin.Engine, controller controller.IProjectController) {
	projectRoutes := router.Group("/project")
	projectRoutes.GET("/all", controller.GetAll)
}

func InitMaterialLocationRoutes(router *gin.Engine, controller controller.IMaterialLocationController) {
	materialLocationRoutes := router.Group("/material-location")
	materialLocationRoutes.Use(middleware.Authentication())
	materialLocationRoutes.GET("/total-amount/:materialID", controller.GetTotalAmountByMaterialID)
	materialLocationRoutes.GET("/available/:locationType/:locationID", controller.GetMaterialInLocation)
  materialLocationRoutes.GET("/unique/team", controller.UniqueTeams)
  materialLocationRoutes.GET("/unique/object", controller.UniqueObjects)
  materialLocationRoutes.POST("/report/balance", controller.ReportBalance)
}

func InitMaterialCostRoutes(router *gin.Engine, controller controller.IMaterialCostController) {
	materialCostRoutes := router.Group("/material-cost")
	materialCostRoutes.Use(middleware.Authentication())
	materialCostRoutes.GET("/paginated", controller.GetPaginated)
	materialCostRoutes.GET("/material-id/:materialID", controller.GetAllMaterialCostByMaterialID)
	materialCostRoutes.POST("/", controller.Create)
	materialCostRoutes.PATCH("/", controller.Update)
	materialCostRoutes.DELETE("/:id", controller.Delete)

}

func InitMaterialRoutes(router *gin.Engine, controller controller.IMaterialController) {
	materialRoutes := router.Group("/material")
	materialRoutes.GET("/all", controller.GetAll)
	materialRoutes.GET("/paginated", controller.GetPaginated)
	materialRoutes.GET("/:id", controller.GetAll)
	materialRoutes.POST("/", controller.Create)
	materialRoutes.PATCH("/", controller.Update)
	materialRoutes.DELETE("/:id", controller.Delete)
}

func InitDistrictRoutes(router *gin.Engine, controller controller.IDistictController) {
	districtRoutes := router.Group("/district")
	districtRoutes.Use(middleware.Authentication())
	districtRoutes.GET("/all", controller.GetAll)
}

func InitTeamRoutes(router *gin.Engine, controller controller.ITeamController) {
	teamRoutes := router.Group("/team")
	teamRoutes.GET("/all", controller.GetAll)
	teamRoutes.GET("/paginated", controller.GetPaginated)
	teamRoutes.GET("/:id", controller.GetByID)
	teamRoutes.POST("/", controller.Create)
	teamRoutes.PATCH("/", controller.Update)
	teamRoutes.DELETE("/:id", controller.Delete)
}

func InitObjectRoutes(router *gin.Engine, controller controller.IObjectController) {
	objectRoutes := router.Group("/object")
	objectRoutes.GET("/all", controller.GetAll)
	objectRoutes.GET("/paginated", controller.GetPaginated)
	objectRoutes.GET("/:id", controller.GetByID)
	objectRoutes.POST("/", controller.Create)
	objectRoutes.PATCH("/", controller.Update)
	objectRoutes.DELETE("/:id", controller.Delete)
}

func InitWorkerRoutes(router *gin.Engine, controller controller.IWorkerController) {
	workerRoutes := router.Group("/worker")
	workerRoutes.GET("/all", controller.GetAll)
	workerRoutes.GET("/paginated", controller.GetPaginated)
	workerRoutes.GET("/:id", controller.GetByID)
	workerRoutes.GET("/job-title/:jobTitle", controller.GetByJobTitle)
	workerRoutes.POST("/", controller.Create)
	workerRoutes.PATCH("/", controller.Update)
	workerRoutes.DELETE("/:id", controller.Delete)
}

func InitUserRoutes(router *gin.Engine, controller controller.IUserController) {
	userRoutes := router.Group("/user")
	// userRoutes.Use(middleware.)
	userRoutes.GET("/all", controller.GetAll)
	userRoutes.GET("/:id", controller.GetByID)
	userRoutes.GET("/paginated", controller.GetPaginated)
	userRoutes.GET("/is-authenticated", controller.IsAuthenticated)
	userRoutes.POST("/", controller.Create)
	userRoutes.POST("/login", controller.Login)
	userRoutes.PATCH("/", controller.Update)
	userRoutes.DELETE("/:id", controller.Delete)
}

func InitPermissionRoutes(router *gin.Engine, controller controller.IPermissionController) {
	permissionRoutes := router.Group("/permission")
	permissionRoutes.Use(middleware.Authentication())

	permissionRoutes.GET("/all", controller.GetAll)
	permissionRoutes.GET("/role/name/:roleName", controller.GetByRoleName)
  permissionRoutes.GET("/role/url/:resourceURL", controller.GetByResourceURL)
	permissionRoutes.POST("/", controller.Create)
	permissionRoutes.POST("/batch", controller.CreateBatch)
	permissionRoutes.PATCH("/", controller.Update)
	permissionRoutes.DELETE("/:id", controller.Delete)
}

func InitRoleRoutes(router *gin.Engine, controller controller.IRoleController) {
	roleRoutes := router.Group("/role")
	roleRoutes.Use(middleware.Authentication())

	roleRoutes.GET("/all", controller.GetAll)
	roleRoutes.POST("/", controller.Create)
	roleRoutes.PATCH("/", controller.Update)
	roleRoutes.DELETE("/:id", controller.Delete)
}
