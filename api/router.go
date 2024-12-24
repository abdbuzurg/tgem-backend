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

	mainRouter := gin.Default()

	mainRouter.Use(gin.Recovery())

	mainRouter.Use(cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Content-Length", "Accept-Encoding", "Authorization", "Cache-Control"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		AllowFiles:       true,
		MaxAge:           12 * time.Hour,
	}))

	router := mainRouter.Group("/api")

	//Initialization of Repositories
  auctionRepository := repository.InitAuctionRepository(db)
	invoiceInputRepo := repository.InitInvoiceInputRepository(db)
	invoiceOutputRepo := repository.InitInvoiceOutputRepository(db)
	invoiceReturnRepo := repository.InitInvoiceReturnRepository(db)
	invoiceMaterialRepo := repository.InitInvoiceMaterialsRepository(db)
	invoiceOutputOutOfProjectRepo := repository.InitInvoiceOutputOutOfProjectRepository(db)
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
	serialNumberMovementRepo := repository.InitSerialNumberMovementRepository(db)
	districtRepo := repository.InitDistrictRepository(db)
	permissionRepo := repository.InitPermissionRepository(db)
	roleRepo := repository.InitRoleRepository(db)
	materialDefectRepo := repository.InitMaterialDefectRepository(db)
	userActionRepo := repository.InitUserActionRepository(db)
	objectSupervisorsRepo := repository.InitObjectSupervisorsRepository(db)
	objectTeamsRepo := repository.InitObjectTeamsRepository(db)
	resourceRepo := repository.InitResourceRepository(db)
	invoiceObjectRepo := repository.InitInvoiceObjectRepository(db)
	invoiceCorrectionRepo := repository.InitInvoiceCorrectionRepository(db)
	substationObjectRepo := repository.InitSubstationObjectRepository(db)
	tpNourashesObjectsRepo := repository.InitTPNourashesObjectsRepository(db)
	invoiceCountRepo := repository.InitInvoiceCountRepository(db)
	operationRepo := repository.InitOperationRepository(db)
	operationMaterialRepo := repository.InitOperationMaterialRepository(db)
	invoiceWriteOffRepo := repository.InitInvoiceWriteOffRepository(db)
	workerAttendanceRepo := repository.InitWorkerAttendanceRepository(db)
	mainReportRepository := repository.InitMainReportRepository(db)
	substationCellRepository := repository.NewSubstationCellObjectRepository(db)

	//Initialization of Services
  auctionService := service.InitAuctionService(auctionRepository)
	invoiceInputService := service.InitInvoiceInputService(
		invoiceInputRepo,
		invoiceMaterialRepo,
		materialLocationRepo,
		workerRepo,
		materialCostRepo,
		materialRepo,
		serialNumberRepo,
		serialNumberMovementRepo,
		invoiceCountRepo,
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
		invoiceCountRepo,
		projectRepo,
	)
	invoiceOutputOutOfProjectService := service.InitInvoiceOutputOutOfProjectService(
		invoiceOutputOutOfProjectRepo,
		invoiceOutputRepo,
		materialLocationRepo,
		invoiceCountRepo,
		invoiceMaterialRepo,
		materialRepo,
		workerRepo,
		materialCostRepo,
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
		materialDefectRepo,
		serialNumberRepo,
		projectRepo,
		districtRepo,
		objectSupervisorsRepo,
		invoiceCountRepo,
	)
	invoiceObjectService := service.InitInvoiceObjectService(
		invoiceObjectRepo,
		objectRepo,
		workerRepo,
		teamRepo,
		materialLocationRepo,
		serialNumberRepo,
		materialCostRepo,
		invoiceMaterialRepo,
		objectTeamsRepo,
		operationMaterialRepo,
		operationRepo,
		materialRepo,
	)
	invoiceCorrectionService := service.InitInvoiceCorrectionService(
		invoiceCorrectionRepo,
		invoiceObjectRepo,
		invoiceMaterialRepo,
		materialLocationRepo,
	)

	// invoiceMaterialsService := service.InitInvoiceMaterialsService(invoiceMaterialRepo)
	kl04kvObjectService := service.InitKL04KVObjectService(
		kl04kvObjectRepo,
		workerRepo,
		objectSupervisorsRepo,
		objectTeamsRepo,
		tpNourashesObjectsRepo,
		teamRepo,
		tpObjectRepo,
		objectRepo,
	)
	materialCostService := service.InitMaterialCostService(
		materialCostRepo,
		materialRepo,
	)
	// materialForProjectService := service.InitMaterialForProjectService(materialForProjectRepo)
	materialLocationService := service.InitMaterialLocationService(
		materialLocationRepo,
		materialCostRepo,
		materialRepo,
		teamRepo,
		objectRepo,
		materialDefectRepo,
		objectSupervisorsRepo,
	)

	materialService := service.InitMaterialService(materialRepo)
	mjdObjectService := service.InitMJDObjectService(
		mjdObjectRepo,
		workerRepo,
		objectSupervisorsRepo,
		objectTeamsRepo,
		tpNourashesObjectsRepo,
		teamRepo,
		tpObjectRepo,
		objectRepo,
	)
	// objectOperationService := service.InitObjectOperationService(objectOperationRepo)
	objectService := service.InitObjectService(
		objectRepo,
		objectSupervisorsRepo,
		kl04kvObjectRepo,
		mjdObjectRepo,
		sipObjectRepo,
		stvtObjectRepo,
		tpObjectRepo,
		objectTeamsRepo,
	)
	operationService := service.InitOperationService(
		operationRepo,
		materialRepo,
	)
	projectService := service.InitProjectService(projectRepo)
	sipObjectService := service.InitSIPObjectService(
		sipObjectRepo,
		workerRepo,
		objectSupervisorsRepo,
		objectTeamsRepo,
		teamRepo,
		tpObjectRepo,
		tpNourashesObjectsRepo,
		objectRepo,
	)
	stvtObjectService := service.InitSTVTObjectService(
		stvtObjectRepo,
		workerRepo,
		objectSupervisorsRepo,
		objectTeamsRepo,
		teamRepo,
	)
	teamService := service.InitTeamService(
		teamRepo,
		workerRepo,
		objectRepo,
	)
	tpObjctService := service.InitTPObjectService(
		tpObjectRepo,
		workerRepo,
		objectSupervisorsRepo,
		objectTeamsRepo,
		objectRepo,
		teamRepo,
	)
	userService := service.InitUserService(
		userRepo,
		userInProjects,
		workerRepo,
		roleRepo,
		userInProjects,
		projectRepo,
	)
	workerService := service.InitWorkerService(workerRepo)
	districtService := service.InitDistrictService(districtRepo)
	permissionService := service.InitPermissionService(
		permissionRepo,
		roleRepo,
		resourceRepo,
	)
	roleService := service.InitRoleService(roleRepo)
	userActionService := service.InitUserActionService(userActionRepo, userRepo)
	resourceService := service.InitResourceService(resourceRepo)
	substationObjectService := service.InitSubstationObjectService(
		substationObjectRepo,
		workerRepo,
		objectSupervisorsRepo,
		objectTeamsRepo,
		teamRepo,
	)
	invoiceWriteOffService := service.InitInvoiceWriteOffService(
		invoiceWriteOffRepo,
		workerRepo,
		objectRepo,
		teamRepo,
		materialLocationRepo,
		invoiceMaterialRepo,
		materialRepo,
		materialCostRepo,
		invoiceCountRepo,
	)
	workerAttendanceService := service.InitWorkerAttendanceService(
		workerAttendanceRepo,
		workerRepo,
	)
	mainReportService := service.InitMainReportService(mainReportRepository)
	substationCellObjectService := service.InitSubstationCellObjectService(
		substationCellRepository,
		objectSupervisorsRepo,
		objectTeamsRepo,
		workerRepo,
		teamRepo,
		substationObjectRepo,
	)

	//Initialization of Controllers
  auctionController := controller.InitAuctionController(auctionService)
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
	operationController := controller.InitOperationController(operationService)
	projectController := controller.InitProjectController(projectService)
	teamController := controller.InitTeamController(teamService)
	userController := controller.InitUserController(userService)
	workerController := controller.InitWorkerController(workerService)
	districtController := controller.InitDistrictController(districtService, userActionService)
	permissionController := controller.InitPermissionController(permissionService)
	roleController := controller.InitRoleController(roleService)
	resourceController := controller.InitResourceController(resourceService)
	invoiceObjectController := controller.InitInvoiceObjectController(invoiceObjectService)
	invoiceCorrectionController := controller.InitInvoiceCorrectionController(invoiceCorrectionService)
	kl04kvObjectController := controller.InitKl04KVObjectController(kl04kvObjectService)
	mjdObjectController := controller.InitMJDObjectController(mjdObjectService)
	sipObjectController := controller.InitSIPObjectController(sipObjectService)
	stvtObjectController := controller.InitSTVTObjectController(stvtObjectService)
	tpObjectController := controller.InitTPObjectController(tpObjctService)
	substationObjectController := controller.InitSubstationObjectController(substationObjectService)
	invoiceOutputOutOfProjectController := controller.InitInvoiceOutputOutOfProjectController(invoiceOutputOutOfProjectService)
	invoiceWriteOffController := controller.InitInvoiceWriteOffController(invoiceWriteOffService)
	workerAttendanceController := controller.InitWorkerAttendanceController(workerAttendanceService)
	mainReportController := controller.InitMainReportController(mainReportService)
	substationCellController := controller.InitSubstationCellObjectController(substationCellObjectService)

	//Initialization of Routes
  InitAuctionRoutes(router, auctionController)
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
	InitResourceRoutes(router, resourceController, db)
	InitInvoiceObjectRoutes(router, invoiceObjectController, db)
	InitInvoiceCorrectionRoutes(router, invoiceCorrectionController)
	InitKL04KVObjectRoutes(router, kl04kvObjectController)
	InitMJDObjectRoutes(router, mjdObjectController)
	InitSIPObjectRoutes(router, sipObjectController)
	InitSTVTObjectRoutes(router, stvtObjectController)
	InitTPObjectRoutes(router, tpObjectController)
	InitSubstationObjectRoutes(router, substationObjectController)
	InitInvoiceOutputOutOfProjectRoutes(router, invoiceOutputOutOfProjectController)
	InitOperationRoutes(router, operationController)
	InitInvoiceWriteOffRoutes(router, invoiceWriteOffController)
	InitWorkerAttendanceRoutes(router, workerAttendanceController)
	InitMainReports(router, mainReportController)
	InitSubstationCellRoutes(router, substationCellController)

	return mainRouter
}

func InitAuctionRoutes(router *gin.RouterGroup, controller controller.IAuctionController) {
  auctionRoutes := router.Group("/auction")
  auctionRoutes.GET("/:auctionID", controller.GetAuctionDataForPublic)
  auctionRoutes.GET("/private/:auctionID", middleware.Authentication(), controller.GetAuctionDataForPrivate)
  auctionRoutes.POST("/private", middleware.Authentication(), controller.SaveParticipantChanges)
}

func InitMainReports(router *gin.RouterGroup, controller controller.IMainReportController) {
	mainReportRoutes := router.Group("/main-reports")
	mainReportRoutes.Use(
		middleware.Authentication(),
	)
	mainReportRoutes.POST("/project-progress", controller.ProjectProgress)
	mainReportRoutes.GET("/analysis-of-remaining-materials", controller.RemainingMaterialAnalysis)
}

func InitWorkerAttendanceRoutes(router *gin.RouterGroup, controller controller.IWorkerAttendanceController) {
	workerAttendanceRoutes := router.Group("/worker-attendance")
	workerAttendanceRoutes.Use(
		middleware.Authentication(),
	)

	workerAttendanceRoutes.GET("/paginated", controller.GetPaginated)
	workerAttendanceRoutes.POST("/", controller.Import)
}

func InitInvoiceWriteOffRoutes(router *gin.RouterGroup, controller controller.IInvoiceWriteOffController) {
	invoiceWriteOffRoutes := router.Group("/invoice-writeoff")
	invoiceWriteOffRoutes.Use(
		middleware.Authentication(),
	)

	invoiceWriteOffRoutes.GET("/paginated", controller.GetPaginated)
	invoiceWriteOffRoutes.GET("/:id/materials/without-serial-number", controller.GetInvoiceMaterialsWithoutSerialNumber)
	invoiceWriteOffRoutes.GET("/invoice-materials/:id/:locationType/:locationID", controller.GetMaterialsForEdit)
	invoiceWriteOffRoutes.GET("/document/:deliveryCode", controller.GetDocument)
	invoiceWriteOffRoutes.GET("/material/:locationType/:locationID", controller.GetMaterialsInLocation)
	invoiceWriteOffRoutes.POST("/", controller.Create)
	invoiceWriteOffRoutes.POST("/confirm/:id", controller.Confirmation)
	invoiceWriteOffRoutes.POST("/report", controller.Report)
	invoiceWriteOffRoutes.PATCH("/", controller.Update)
	invoiceWriteOffRoutes.DELETE("/:id", controller.Delete)
}

func InitInvoiceOutputOutOfProjectRoutes(router *gin.RouterGroup, controller controller.IInvoiceOutputOutOfProjectController) {
	invoiceOutputOutOfProjectRoutes := router.Group("/invoice-output-out-of-project")
	invoiceOutputOutOfProjectRoutes.Use(
		middleware.Authentication(),
	)

	invoiceOutputOutOfProjectRoutes.GET("/paginated", controller.GetPaginated)
	invoiceOutputOutOfProjectRoutes.GET("/:id/materials/without-serial-number", controller.GetInvoiceMaterialsWithoutSerialNumbers)
	invoiceOutputOutOfProjectRoutes.GET("/:id/materials/with-serial-number", controller.GetInvoiceMaterialsWithSerialNumbers)
	invoiceOutputOutOfProjectRoutes.GET("/invoice-materials/:id", controller.GetMaterialsForEdit)
	invoiceOutputOutOfProjectRoutes.GET("/unique/name-of-project", controller.UniqueNameOfProjects)
	invoiceOutputOutOfProjectRoutes.GET("/document/:deliveryCode", controller.GetDocument)
	invoiceOutputOutOfProjectRoutes.POST("/", controller.Create)
	invoiceOutputOutOfProjectRoutes.POST("/report", controller.Report)
	invoiceOutputOutOfProjectRoutes.PATCH("/", controller.Update)
	invoiceOutputOutOfProjectRoutes.POST("/confirm/:id", controller.Confirmation)
	invoiceOutputOutOfProjectRoutes.DELETE("/:id", controller.Delete)
}

func InitInvoiceCorrectionRoutes(router *gin.RouterGroup, controller controller.IInvoiceCorrectionController) {
	invoiceCorrectionRoutes := router.Group("/invoice-correction")
	invoiceCorrectionRoutes.Use(
		middleware.Authentication(),
	)

	invoiceCorrectionRoutes.GET("/paginated", controller.GetPaginated)
	invoiceCorrectionRoutes.GET("/", controller.GetAll)
	invoiceCorrectionRoutes.GET("/materials/:id", controller.GetInvoiceMaterialsByInvoiceObjectID)
	invoiceCorrectionRoutes.GET("/operations/:id", controller.GetOperationsByInvoiceObjectID)
	invoiceCorrectionRoutes.GET("/total-amount/:materialID/team/:teamNumber", controller.GetTotalMaterialInTeamByTeamNumber)
	invoiceCorrectionRoutes.GET("/serial-number/material/:materialID/teams/:teamNumber", controller.GetSerialNumbersOfMaterial)
	invoiceCorrectionRoutes.GET("unique/team", controller.UniqueTeam)
	invoiceCorrectionRoutes.GET("unique/object", controller.UniqueObject)
	invoiceCorrectionRoutes.POST("/report", controller.Report)
	invoiceCorrectionRoutes.POST("/", controller.Create)
}

func InitInvoiceObjectRoutes(router *gin.RouterGroup, controller controller.IInvoiceObjectController, db *gorm.DB) {
	invoiceObjectRoutes := router.Group("/invoice-object")
	invoiceObjectRoutes.Use(
		middleware.Authentication(),
		middleware.Permission(db),
	)

	invoiceObjectRoutes.GET("/:id", controller.GetInvoiceObjectDescriptiveDataByID)
	invoiceObjectRoutes.GET("/paginated", controller.GetPaginated)
	invoiceObjectRoutes.GET("/team-materials/:teamID", controller.GetTeamsMaterials)
	invoiceObjectRoutes.GET("/available-operations-for-team/:teamID", controller.GetOperationsBasedOnMaterialsInTeamID)
	invoiceObjectRoutes.GET("/serial-number/material/:materialID/teams/:teamID", controller.GetSerialNumbersOfMaterial)
	invoiceObjectRoutes.GET("/material/:materialID/team/:teamID", controller.GetMaterialAmountInTeam)
	invoiceObjectRoutes.GET("/object/:objectID", controller.GetTeamsFromObjectID)
	invoiceObjectRoutes.POST("/", controller.Create)
}
func InitInvoiceReturnRoutes(router *gin.RouterGroup, controller controller.IInvoiceReturnController, db *gorm.DB) {
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
	invoiceReturnRoutes.GET("/material/:locationType/:locationID", controller.GetMaterialsInLocation)
	invoiceReturnRoutes.GET("/material-cost/:materialID/:locationType/:locationID", controller.GetUniqueMaterialCostsFromLocation)
	invoiceReturnRoutes.GET("/material-amount/:materialCostID/:locationType/:locationID", controller.GetMaterialAmountInLocation)
	invoiceReturnRoutes.GET("/serial-number/:locationType/:locationID/:materialID", controller.GetSerialNumberCodesInLocation)
	invoiceReturnRoutes.GET("/:id/materials/without-serial-number", controller.GetInvoiceMaterialsWithoutSerialNumbers)
	invoiceReturnRoutes.GET("/:id/materials/with-serial-number", controller.GetInvoiceMaterialsWithSerialNumbers)
	invoiceReturnRoutes.GET("/invoice-materials/:id/:locationType/:locationID", controller.GetMaterialsForEdit)
	invoiceReturnRoutes.GET("/amount/:locationType/:locationID/:materialID", controller.GetMaterialAmountByMaterialID)
	invoiceReturnRoutes.POST("/confirm/:id", controller.Confirmation)
	invoiceReturnRoutes.POST("/", controller.Create)
	invoiceReturnRoutes.POST("/report", controller.Report)
	invoiceReturnRoutes.PATCH("/", controller.Update)
	invoiceReturnRoutes.DELETE("/:id", controller.Delete)
}

func InitInvoiceOutputRoutes(router *gin.RouterGroup, controller controller.IInvoiceOutputController) {
	invoiceOutputRoutes := router.Group("/output")
	invoiceOutputRoutes.Use(
		middleware.Authentication(),
	)
	invoiceOutputRoutes.GET("/paginated", controller.GetPaginated)
	invoiceOutputRoutes.GET("/unique/district", controller.UniqueDistrict)
	invoiceOutputRoutes.GET("/unique/code", controller.UniqueCode)
	invoiceOutputRoutes.GET("/unique/recieved", controller.UniqueRecieved)
	invoiceOutputRoutes.GET("/unique/warehouse-manager", controller.UniqueRecieved)
	invoiceOutputRoutes.GET("/unique/team", controller.UniqueTeam)
	invoiceOutputRoutes.GET("/document/:deliveryCode", controller.GetDocument)
	invoiceOutputRoutes.GET("/material/available-in-warehouse", controller.GetAvailableMaterialsInWarehouse)
	invoiceOutputRoutes.GET("/material/:materialID/total-amount", controller.GetTotalAmountInWarehouse)
	invoiceOutputRoutes.GET("/serial-number/material/:materialID", controller.GetCodesByMaterialID)
	invoiceOutputRoutes.GET("/:id/materials/without-serial-number", controller.GetInvoiceMaterialsWithoutSerialNumbers)
	invoiceOutputRoutes.GET("/:id/materials/with-serial-number", controller.GetInvoiceMaterialsWithSerialNumbers)
	invoiceOutputRoutes.GET("/invoice-materials/:id", controller.GetMaterialsForEdit)
	invoiceOutputRoutes.POST("/report", controller.Report)
	invoiceOutputRoutes.POST("/confirm/:id", controller.Confirmation)
	invoiceOutputRoutes.POST("/", controller.Create)
	invoiceOutputRoutes.POST("/import", controller.Import)
	invoiceOutputRoutes.PATCH("/", controller.Update)
	invoiceOutputRoutes.DELETE("/:id", controller.Delete)
}

func InitInvoiceInputRoutes(router *gin.RouterGroup, controller controller.IInvoiceInputController, db *gorm.DB) {
	invoiceInputRoutes := router.Group("/input")
	invoiceInputRoutes.Use(
		middleware.Authentication(),
		middleware.Permission(db),
	)

	invoiceInputRoutes.GET("/paginated", controller.GetPaginated)
	invoiceInputRoutes.GET("/unique/code", controller.UniqueCode)
	invoiceInputRoutes.GET("/unique/warehouse-manager", controller.UniqueWarehouseManager)
	invoiceInputRoutes.GET("/unique/released", controller.UniqueReleased)
	invoiceInputRoutes.GET("/document/:deliveryCode", controller.GetDocument)
	invoiceInputRoutes.GET("/:id/materials/without-serial-number", controller.GetInvoiceMaterialsWithoutSerialNumbers)
	invoiceInputRoutes.GET("/:id/materials/with-serial-number", controller.GetInvoiceMaterialsWithSerialNumbers)
	invoiceInputRoutes.GET("/invoice-materials/:id", controller.GetMaterialsForEdit)
	invoiceInputRoutes.POST("/", controller.Create)
	invoiceInputRoutes.POST("/report", controller.Report)
	invoiceInputRoutes.POST("/confirm/:id", controller.Confirmation)
	invoiceInputRoutes.POST("/material/new", controller.NewMaterial)
	invoiceInputRoutes.POST("/material-cost/new", controller.NewMaterialCost)
	invoiceInputRoutes.POST("/import", controller.Import)
	invoiceInputRoutes.PATCH("/", controller.Update)
	invoiceInputRoutes.DELETE("/:id", controller.Delete)
}

func InitProjectRoutes(router *gin.RouterGroup, controller controller.IProjectController) {
	projectRoutes := router.Group("/project")
	projectRoutes.GET("/all", controller.GetAll)

	projectRoutes.Use(middleware.Authentication())
	projectRoutes.GET("/paginated", controller.GetPaginated)
	projectRoutes.GET("/name", controller.GetProjectName)
	projectRoutes.POST("/", controller.Create)
	projectRoutes.PATCH("/", controller.Update)
	projectRoutes.DELETE("/:id", controller.Delete)
}

func InitMaterialLocationRoutes(router *gin.RouterGroup, controller controller.IMaterialLocationController) {
	materialLocationRoutes := router.Group("/material-location")
	materialLocationRoutes.Use(middleware.Authentication())
	materialLocationRoutes.GET("/available/:locationType/:locationID", controller.GetMaterialInLocation)
	materialLocationRoutes.GET("/costs/:materialID/:locationType/:locationID", controller.GetMaterialCostsInLocation)
	materialLocationRoutes.GET("/amount/:materialCostID/:locationType/:locationID", controller.GetMaterialAmountBasedOnCost)
	materialLocationRoutes.GET("/unique/team", controller.UniqueTeams)
	materialLocationRoutes.GET("/unique/object", controller.UniqueObjects)
	materialLocationRoutes.GET("/live", controller.Live)
	materialLocationRoutes.POST("/report/balance", controller.ReportBalance)
	materialLocationRoutes.POST("/report/balance/writeoff", controller.ReportBalanceWriteOff)
	materialLocationRoutes.POST("/report/balance/out-of-project", controller.ReportBalanceOutOfProject)
}

func InitMaterialCostRoutes(router *gin.RouterGroup, controller controller.IMaterialCostController) {
	materialCostRoutes := router.Group("/material-cost")
	materialCostRoutes.Use(middleware.Authentication())
	materialCostRoutes.GET("/paginated", controller.GetPaginated)
	materialCostRoutes.GET("/material-id/:materialID", controller.GetAllMaterialCostByMaterialID)
	materialCostRoutes.GET("/document/template", controller.ImportTemplate)
	materialCostRoutes.GET("/document/export", controller.Export)
	materialCostRoutes.POST("/document/import", controller.Import)
	materialCostRoutes.POST("/", controller.Create)
	materialCostRoutes.PATCH("/", controller.Update)
	materialCostRoutes.DELETE("/:id", controller.Delete)
}

func InitMaterialRoutes(router *gin.RouterGroup, controller controller.IMaterialController) {
	materialRoutes := router.Group("/material")
	materialRoutes.Use(
		middleware.Authentication(),
	)
	materialRoutes.GET("/all", controller.GetAll)
	materialRoutes.GET("/paginated", controller.GetPaginated)
	materialRoutes.GET("/document/template", controller.GetTemplateFile)
	materialRoutes.GET("/document/export", controller.Export)
	materialRoutes.POST("/", controller.Create)
	materialRoutes.POST("/document/import", controller.Import)
	materialRoutes.PATCH("/", controller.Update)
	materialRoutes.DELETE("/:id", controller.Delete)
}

func InitDistrictRoutes(router *gin.RouterGroup, controller controller.IDistictController) {
	districtRoutes := router.Group("/district")
	districtRoutes.Use(
		middleware.Authentication(),
	)
	districtRoutes.GET("/all", controller.GetAll)
	districtRoutes.GET("/paginated", controller.GetPaginated)
	districtRoutes.POST("/", controller.Create)
	districtRoutes.PATCH("/", controller.Update)
	districtRoutes.DELETE("/:id", controller.Delete)
}

func InitTeamRoutes(router *gin.RouterGroup, controller controller.ITeamController) {
	teamRoutes := router.Group("/team")
	teamRoutes.Use(
		middleware.Authentication(),
	)
	teamRoutes.GET("/all", controller.GetAll)
	teamRoutes.GET("/all/for-select", controller.GetAllForSelect)
	teamRoutes.GET("/paginated", controller.GetPaginated)
	teamRoutes.GET("/:id", controller.GetByID)
	teamRoutes.GET("/document/template", controller.GetTemplateFile)
	teamRoutes.GET("/unique/team-number", controller.GetAllUniqueTeamNumbers)
	teamRoutes.GET("/unique/mobile-number", controller.GetAllUniqueMobileNumber)
	teamRoutes.GET("/unique/team-company", controller.GetAllUniqueCompanies)
	teamRoutes.GET("/document/export", controller.Export)
	teamRoutes.POST("/", controller.Create)
	teamRoutes.POST("/document/import", controller.Import)
	teamRoutes.PATCH("/", controller.Update)
	teamRoutes.DELETE("/:id", controller.Delete)
}

func InitObjectRoutes(router *gin.RouterGroup, controller controller.IObjectController) {
	objectRoutes := router.Group("/object")
	objectRoutes.Use(
		middleware.Authentication(),
	)
	objectRoutes.GET("/all", controller.GetAll)
	objectRoutes.GET("/paginated", controller.GetPaginated)
	objectRoutes.GET("/:id", controller.GetByID)
	objectRoutes.GET("/teams/:objectID", controller.GetTeamsByObject)
	objectRoutes.POST("/", controller.Create)
	objectRoutes.PATCH("/", controller.Update)
	objectRoutes.DELETE("/:id", controller.Delete)
}

func InitTPObjectRoutes(router *gin.RouterGroup, controller controller.ITPObjectController) {
	tpObjectRoutes := router.Group("/tp")
	tpObjectRoutes.Use(
		middleware.Authentication(),
	)
	tpObjectRoutes.GET("/paginated", controller.GetPaginated)
	tpObjectRoutes.GET("/document/template", controller.GetTemplateFile)
	tpObjectRoutes.GET("/all", controller.GetAll)
	tpObjectRoutes.GET("/search/object-names", controller.GetObjectNamesForSearch)
	tpObjectRoutes.GET("/document/export", controller.Export)
	tpObjectRoutes.POST("/", controller.Create)
	tpObjectRoutes.POST("/document/import", controller.Import)
	tpObjectRoutes.PATCH("/", controller.Update)
	tpObjectRoutes.DELETE("/:id", controller.Delete)
}

func InitSubstationCellRoutes(router *gin.RouterGroup, controller controller.ISubstationCellObjectController) {
	substationCellObjectRoutes := router.Group("/cell-substation")
	substationCellObjectRoutes.Use(
		middleware.Authentication(),
	)
	substationCellObjectRoutes.GET("/paginated", controller.GetPaginated)
	substationCellObjectRoutes.GET("/document/template", controller.GetTemplateFile)
	substationCellObjectRoutes.GET("/document/export", controller.Export)
	substationCellObjectRoutes.GET("/search/object-names", controller.GetObjectNamesForSearch)
	substationCellObjectRoutes.POST("/document/import", controller.Import)
	substationCellObjectRoutes.POST("/", controller.Create)
	substationCellObjectRoutes.PATCH("/", controller.Update)
	substationCellObjectRoutes.DELETE("/:id", controller.Delete)

}

func InitSubstationObjectRoutes(router *gin.RouterGroup, controller controller.ISubstationObjectController) {
	substationObjectRoutes := router.Group("/substation")
	substationObjectRoutes.Use(
		middleware.Authentication(),
	)
	substationObjectRoutes.GET("/all", controller.GetAll)
	substationObjectRoutes.GET("/paginated", controller.GetPaginated)
	substationObjectRoutes.GET("/document/template", controller.GetTemplateFile)
	substationObjectRoutes.GET("/search/object-names", controller.GetObjectNamesForSearch)
	substationObjectRoutes.GET("/document/export", controller.Export)
	substationObjectRoutes.POST("/", controller.Create)
	substationObjectRoutes.POST("/document/import", controller.Import)
	substationObjectRoutes.PATCH("/", controller.Update)
	substationObjectRoutes.DELETE("/:id", controller.Delete)
}

func InitSTVTObjectRoutes(router *gin.RouterGroup, controller controller.ISTVTObjectController) {
	stvtObjectRoutes := router.Group("/stvt")
	stvtObjectRoutes.Use(
		middleware.Authentication(),
	)
	stvtObjectRoutes.GET("/paginated", controller.GetPaginated)
	stvtObjectRoutes.GET("/document/template", controller.GetTemplateFile)
	stvtObjectRoutes.GET("/search/object-names", controller.GetObjectNamesForSearch)
	stvtObjectRoutes.GET("/document/export", controller.Export)
	stvtObjectRoutes.POST("/", controller.Create)
	stvtObjectRoutes.POST("/document/import", controller.Import)
	stvtObjectRoutes.PATCH("/", controller.Update)
	stvtObjectRoutes.DELETE("/:id", controller.Delete)
}

func InitSIPObjectRoutes(router *gin.RouterGroup, controller controller.ISIPObjectController) {
	sipObjectRoutes := router.Group("/sip")
	sipObjectRoutes.Use(
		middleware.Authentication(),
	)
	sipObjectRoutes.GET("/paginated", controller.GetPaginated)
	sipObjectRoutes.GET("/document/template", controller.GetTemplateFile)
	sipObjectRoutes.GET("/tp-names", controller.GetTPNames)
	sipObjectRoutes.GET("/search/object-names", controller.GetObjectNamesForSearch)
	sipObjectRoutes.GET("/document/export", controller.Export)
	sipObjectRoutes.POST("/", controller.Create)
	sipObjectRoutes.POST("/document/import", controller.Import)
	sipObjectRoutes.PATCH("/", controller.Update)
	sipObjectRoutes.DELETE("/:id", controller.Delete)
}

func InitMJDObjectRoutes(router *gin.RouterGroup, controller controller.IMJDObjectController) {
	mjdObjectRoutes := router.Group("/mjd")
	mjdObjectRoutes.Use(
		middleware.Authentication(),
	)
	mjdObjectRoutes.GET("/paginated", controller.GetPaginated)
	mjdObjectRoutes.GET("/document/template", controller.GetTemplateFile)
	mjdObjectRoutes.GET("/document/export", controller.Export)
	mjdObjectRoutes.GET("/search/object-names", controller.GetObjectNamesForSearch)
	mjdObjectRoutes.POST("/", controller.Create)
	mjdObjectRoutes.POST("/document/import", controller.Import)
	mjdObjectRoutes.PATCH("/", controller.Update)
	mjdObjectRoutes.DELETE("/:id", controller.Delete)
}

func InitKL04KVObjectRoutes(router *gin.RouterGroup, controller controller.IKL04KVObjectController) {
	kl04kvObjectRoutes := router.Group("/kl04kv")
	kl04kvObjectRoutes.Use(
		middleware.Authentication(),
	)
	kl04kvObjectRoutes.GET("/paginated", controller.GetPaginated)
	kl04kvObjectRoutes.GET("/document/export", controller.Export)
	kl04kvObjectRoutes.GET("/document/template", controller.GetTemplateFile)
	kl04kvObjectRoutes.GET("/search/object-names", controller.GetObjectNamesForSearch)
	kl04kvObjectRoutes.POST("/", controller.Create)
	kl04kvObjectRoutes.POST("/document/import", controller.Import)
	kl04kvObjectRoutes.PATCH("/", controller.Update)
	kl04kvObjectRoutes.DELETE("/:id", controller.Delete)
}

func InitWorkerRoutes(router *gin.RouterGroup, controller controller.IWorkerController) {
	workerRoutes := router.Group("/worker")
	workerRoutes.Use(
		middleware.Authentication(),
	)
	workerRoutes.GET("/all", controller.GetAll)
	workerRoutes.GET("/paginated", controller.GetPaginated)
	workerRoutes.GET("/:id", controller.GetByID)
	workerRoutes.GET("/job-title/:jobTitleInProject", controller.GetByJobTitleInProject)
	workerRoutes.GET("/document/template", controller.GetTemplateFile)
	workerRoutes.GET("/unique/worker-information", controller.GetWorkerInformationForSearch)
	workerRoutes.POST("/", controller.Create)
	workerRoutes.POST("/document/import", controller.Import)
	workerRoutes.PATCH("/", controller.Update)
	workerRoutes.DELETE("/:id", controller.Delete)
}

func InitUserRoutes(router *gin.RouterGroup, controller controller.IUserController) {
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

func InitPermissionRoutes(router *gin.RouterGroup, controller controller.IPermissionController) {
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

func InitRoleRoutes(router *gin.RouterGroup, controller controller.IRoleController) {
	roleRoutes := router.Group("/role")
	roleRoutes.Use(middleware.Authentication())

	roleRoutes.GET("/all", controller.GetAll)
	roleRoutes.POST("/", controller.Create)
	roleRoutes.PATCH("/", controller.Update)
	roleRoutes.DELETE("/:id", controller.Delete)
}

func InitResourceRoutes(router *gin.RouterGroup, controller controller.IResourceController, db *gorm.DB) {
	resourceRoutes := router.Group("/resource")
	resourceRoutes.Use(middleware.Authentication(), middleware.Permission(db))

	resourceRoutes.GET("/", controller.GetAll)
}

func InitOperationRoutes(router *gin.RouterGroup, controller controller.IOperationController) {
	operationRoutes := router.Group("/operation")
	operationRoutes.Use(
		middleware.Authentication(),
	)
	operationRoutes.GET("/paginated", controller.GetPaginated)
	operationRoutes.GET("/all", controller.GetAll)
	operationRoutes.GET("/document/template", controller.GetTemplateFile)
	operationRoutes.POST("/", controller.Create)
	operationRoutes.POST("/document/import", controller.Import)
	operationRoutes.PATCH("/", controller.Update)
	operationRoutes.DELETE("/:id", controller.Delete)
}
