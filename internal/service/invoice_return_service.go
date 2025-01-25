package service

import (
	"backend-v2/internal/dto"
	"backend-v2/internal/repository"
	"backend-v2/model"
	"backend-v2/pkg/utils"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/xuri/excelize/v2"
)

type invoiceReturnService struct {
	invoiceReturnRepo     repository.IInvoiceReturnRepository
	workerRepo            repository.IWorkerRepository
	objectRepo            repository.IObjectRepository
	teamRepo              repository.ITeamRepository
	materialLocationRepo  repository.IMaterialLocationRepository
	invoiceMaterialsRepo  repository.IInvoiceMaterialsRepository
	materialRepo          repository.IMaterialRepository
	materialCostRepo      repository.IMaterialCostRepository
	materialDefectRepo    repository.IMaterialDefectRepository
	serialNumberRepo      repository.ISerialNumberRepository
	projectRepo           repository.IProjectRepository
	districtRepo          repository.IDistrictRepository
	objectSupervisorsRepo repository.IObjectSupervisorsRepository
	invoiceCountRepo      repository.IInvoiceCountRepository
}

func InitInvoiceReturnService(
	invoiceReturnRepo repository.IInvoiceReturnRepository,
	workerRepo repository.IWorkerRepository,
	objectRepo repository.IObjectRepository,
	teamRepo repository.ITeamRepository,
	materialLocationRepo repository.IMaterialLocationRepository,
	invoiceMaterialsRepo repository.IInvoiceMaterialsRepository,
	materialRepo repository.IMaterialRepository,
	materialCostRepo repository.IMaterialCostRepository,
	materialDefectRepo repository.IMaterialDefectRepository,
	serialNumberRepo repository.ISerialNumberRepository,
	projectRepo repository.IProjectRepository,
	districtRepo repository.IDistrictRepository,
	objectSupervisorsRepo repository.IObjectSupervisorsRepository,
	invoiceCountRepo repository.IInvoiceCountRepository,
) IInvoiceReturnService {
	return &invoiceReturnService{
		invoiceReturnRepo:     invoiceReturnRepo,
		workerRepo:            workerRepo,
		objectRepo:            objectRepo,
		teamRepo:              teamRepo,
		materialLocationRepo:  materialLocationRepo,
		invoiceMaterialsRepo:  invoiceMaterialsRepo,
		materialRepo:          materialRepo,
		materialCostRepo:      materialCostRepo,
		materialDefectRepo:    materialDefectRepo,
		serialNumberRepo:      serialNumberRepo,
		projectRepo:           projectRepo,
		districtRepo:          districtRepo,
		objectSupervisorsRepo: objectSupervisorsRepo,
		invoiceCountRepo:      invoiceCountRepo,
	}
}

type IInvoiceReturnService interface {
	GetAll() ([]model.InvoiceReturn, error)
	GetByID(id uint) (model.InvoiceReturn, error)
	GetPaginatedTeam(page, limit int, projectID uint) ([]dto.InvoiceReturnTeamPaginatedQueryData, error)
	GetPaginatedObject(page, limit int, projectID uint) ([]dto.InvoiceReturnObjectPaginated, error)
	GetDocument(deliveryCode string) (string, error)
	Create(data dto.InvoiceReturn) (model.InvoiceReturn, error)
	Update(data dto.InvoiceReturn) (model.InvoiceReturn, error)
	Delete(id uint) error
	CountBasedOnType(projectID uint, invoiceType string) (int64, error)
	Confirmation(id uint) error
	UniqueCode(projectID uint) ([]string, error)
	UniqueTeam(projectID uint) ([]string, error)
	UniqueObject(projectID uint) ([]string, error)
	Report(filter dto.InvoiceReturnReportFilterRequest, projectID uint) (string, error)
	GetMaterialsInLocation(projectID, locationID uint, locationType string) ([]dto.InvoiceReturnMaterialForSelect, error)
	GetMaterialCostInLocation(projectID, locationID, materialID uint, locationType string) ([]model.MaterialCost, error)
	GetMaterialAmountInLocation(projectID, locationID, materialCostID uint, locationType string) (float64, error)
	GetSerialNumberCodesInLocation(projectID, materialID uint, locationType string, locationID uint) ([]string, error)
	GetInvoiceMaterialsWithoutSerialNumbers(id uint) ([]dto.InvoiceMaterialsWithoutSerialNumberView, error)
	GetInvoiceMaterialsWithSerialNumbers(id uint) ([]dto.InvoiceMaterialsWithSerialNumberView, error)
	GetMaterialsForEdit(id uint, locationType string, locationID uint) ([]dto.InvoiceReturnMaterialForEdit, error)
	GetMaterialAmountByMaterialID(projectID, materialID, locationID uint, locationType string) (float64, error)
}

func (service *invoiceReturnService) GetAll() ([]model.InvoiceReturn, error) {
	return service.invoiceReturnRepo.GetAll()
}

func (service *invoiceReturnService) GetByID(id uint) (model.InvoiceReturn, error) {
	return service.invoiceReturnRepo.GetByID(id)
}

func (service *invoiceReturnService) GetPaginatedTeam(page, limit int, projectID uint) ([]dto.InvoiceReturnTeamPaginatedQueryData, error) {
	return service.invoiceReturnRepo.GetPaginatedTeam(page, limit, projectID)
}

func (service *invoiceReturnService) GetPaginatedObject(page, limit int, projectID uint) ([]dto.InvoiceReturnObjectPaginated, error) {
	invoiceReturnQueryData, err := service.invoiceReturnRepo.GetPaginatedObject(page, limit, projectID)
	if err != nil {
		return []dto.InvoiceReturnObjectPaginated{}, err
	}

	result := []dto.InvoiceReturnObjectPaginated{}
	currentInvoice := dto.InvoiceReturnObjectPaginated{}
	for index, entry := range invoiceReturnQueryData {
		if index == 0 {
			currentInvoice = dto.InvoiceReturnObjectPaginated{
				ID:                    entry.ID,
				DeliveryCode:          entry.DeliveryCode,
				DateOfInvoice:         entry.DateOfInvoice,
				ObjectName:            entry.ObjectName,
				ObjectType:            entry.ObjectType,
				AcceptorName:          entry.AcceptorName,
				DistrictName:          entry.DistrictName,
				TeamNumber:            entry.TeamNumber,
				TeamLeaderName:        entry.TeamLeaderName,
				ObjectSupervisorNames: []string{},
				Confirmation:          entry.Confirmation,
			}
		}

		if currentInvoice.ID == entry.ID {
			currentInvoice.ObjectSupervisorNames = append(currentInvoice.ObjectSupervisorNames, entry.ObjectSupervisorName)
		} else {
			result = append(result, currentInvoice)
			currentInvoice = dto.InvoiceReturnObjectPaginated{
				ID:                    entry.ID,
				DeliveryCode:          entry.DeliveryCode,
				DateOfInvoice:         entry.DateOfInvoice,
				AcceptorName:          entry.AcceptorName,
				DistrictName:          entry.DistrictName,
				ObjectName:            entry.ObjectName,
				ObjectSupervisorNames: []string{entry.ObjectSupervisorName},
				ObjectType:            entry.ObjectType,
				TeamNumber:            entry.TeamNumber,
				TeamLeaderName:        entry.TeamLeaderName,
				Confirmation:          entry.Confirmation,
			}
		}
	}

	if len(invoiceReturnQueryData) != 0 {
		result = append(result, currentInvoice)
	}

	return result, nil
}

func (service *invoiceReturnService) Create(data dto.InvoiceReturn) (model.InvoiceReturn, error) {

	count, err := service.invoiceCountRepo.CountInvoice("return", data.Details.ProjectID)
	if err != nil {
		return model.InvoiceReturn{}, err
	}

	data.Details.DeliveryCode = utils.UniqueCodeGeneration("В", int64(count+1), data.Details.ProjectID)

	invoiceMaterialsForCreate := []model.InvoiceMaterials{}
	serialNumberMovements := []model.SerialNumberMovement{}
	for _, invoiceMaterial := range data.Items {
		if len(invoiceMaterial.SerialNumbers) == 0 {
			materialCostsReverseSorted, err := service.materialLocationRepo.GetMaterialAmountReverseSortedByCostM19InLocation(data.Details.ProjectID, invoiceMaterial.MaterialID, data.Details.ReturnerType, data.Details.ReturnerID)
			if err != nil {
				return model.InvoiceReturn{}, err
			}

			fmt.Println(materialCostsReverseSorted)

			index := 0
			for invoiceMaterial.Amount > 0 {
				invoiceMaterialCreate := model.InvoiceMaterials{
					MaterialCostID: materialCostsReverseSorted[index].MaterialCostID,
					ProjectID:      data.Details.ProjectID,
					InvoiceID:      0,
					InvoiceType:    "return",
					IsDefected:     invoiceMaterial.IsDefected,
					Amount:         0,
					Notes:          invoiceMaterial.Notes,
				}

				if materialCostsReverseSorted[index].MaterialAmount <= invoiceMaterial.Amount {
					invoiceMaterialCreate.Amount = materialCostsReverseSorted[index].MaterialAmount
					invoiceMaterial.Amount -= materialCostsReverseSorted[index].MaterialAmount
				} else {
					invoiceMaterialCreate.Amount = invoiceMaterial.Amount
					invoiceMaterial.Amount = 0
				}

				invoiceMaterialsForCreate = append(invoiceMaterialsForCreate, invoiceMaterialCreate)
				index++
			}
		}

		if len(invoiceMaterial.SerialNumbers) != 0 {
			return model.InvoiceReturn{}, fmt.Errorf("Операция возврат через серийный номер тестируется")
		}
	}

	if err := service.GenerateExcel(data); err != nil {
		return model.InvoiceReturn{}, err
	}

	invoiceReturn, err := service.invoiceReturnRepo.Create(dto.InvoiceReturnCreateQueryData{
		Invoice:               data.Details,
		InvoiceMaterials:      invoiceMaterialsForCreate,
		SerialNumberMovements: serialNumberMovements,
	})
	if err != nil {
		return model.InvoiceReturn{}, err
	}

	return invoiceReturn, nil
}

func (service *invoiceReturnService) Update(data dto.InvoiceReturn) (model.InvoiceReturn, error) {

	invoiceMaterialsForCreate := []model.InvoiceMaterials{}
	serialNumberMovements := []model.SerialNumberMovement{}
	for _, invoiceMaterial := range data.Items {
		if len(invoiceMaterial.SerialNumbers) == 0 {
			materialCostsReverseSorted, err := service.materialLocationRepo.GetMaterialAmountReverseSortedByCostM19InLocation(data.Details.ProjectID, invoiceMaterial.MaterialID, data.Details.ReturnerType, data.Details.ReturnerID)
			if err != nil {
				return model.InvoiceReturn{}, err
			}

			index := 0
			for invoiceMaterial.Amount > 0 {
				invoiceMaterialCreate := model.InvoiceMaterials{
					MaterialCostID: materialCostsReverseSorted[index].MaterialCostID,
					ProjectID:      data.Details.ProjectID,
					InvoiceID:      0,
					InvoiceType:    "return",
					IsDefected:     invoiceMaterial.IsDefected,
					Amount:         0,
					Notes:          invoiceMaterial.Notes,
				}

				if materialCostsReverseSorted[index].MaterialAmount <= invoiceMaterial.Amount {
					invoiceMaterialCreate.Amount = materialCostsReverseSorted[index].MaterialAmount
					invoiceMaterial.Amount -= materialCostsReverseSorted[index].MaterialAmount
				} else {
					invoiceMaterialCreate.Amount = invoiceMaterial.Amount
					invoiceMaterial.Amount = 0
				}

				invoiceMaterialsForCreate = append(invoiceMaterialsForCreate, invoiceMaterialCreate)
				index++
			}
		}

		if len(invoiceMaterial.SerialNumbers) == 0 {
			return model.InvoiceReturn{}, fmt.Errorf("Операция возврат через серийный номер тестируется")
		}

	}

	excelFilePath := filepath.Join("./pkg/excels/return/", data.Details.DeliveryCode+".xlsx")
	if err := os.Remove(excelFilePath); err != nil {
		return model.InvoiceReturn{}, err
	}

	if err := service.GenerateExcel(data); err != nil {
		return model.InvoiceReturn{}, err
	}

	invoiceReturn, err := service.invoiceReturnRepo.Update(dto.InvoiceReturnCreateQueryData{
		Invoice:               data.Details,
		InvoiceMaterials:      invoiceMaterialsForCreate,
		SerialNumberMovements: serialNumberMovements,
	})
	if err != nil {
		return model.InvoiceReturn{}, err
	}

	return invoiceReturn, nil
}

func (service *invoiceReturnService) Delete(id uint) error {
	return service.invoiceReturnRepo.Delete(id)
}

func (service *invoiceReturnService) CountBasedOnType(projectID uint, invoiceType string) (int64, error) {
	return service.invoiceReturnRepo.CountBasedOnType(projectID, invoiceType)
}

func (service *invoiceReturnService) Confirmation(id uint) error {
	invoiceReturn, err := service.invoiceReturnRepo.GetByID(id)
	if err != nil {
		return err
	}
	invoiceReturn.Confirmation = true

	invoiceMaterials, err := service.invoiceMaterialsRepo.GetByInvoice(invoiceReturn.ProjectID, invoiceReturn.ID, "return")
	if err != nil {
		return err
	}

	materialsInReturnerLocation, err := service.materialLocationRepo.GetMaterialsInLocationBasedOnInvoiceID(invoiceReturn.ReturnerID, invoiceReturn.ReturnerType, id, "return")
	if err != nil {
		return err
	}

	materialsInAcceptorLocation, err := service.materialLocationRepo.GetMaterialsInLocationBasedOnInvoiceID(invoiceReturn.AcceptorID, invoiceReturn.AcceptorType, id, "return")
	if err != nil {
		return err
	}

	materialsDefected := []model.MaterialDefect{}
	newMaterialsDefected := []model.MaterialDefect{}
	newMaterialsInAcceptorLocationWithDefect := []model.MaterialLocation{}
	for _, invoiceMaterial := range invoiceMaterials {
		materialInReturnerLocationIndex := -1
		for index, materialInReturnerLocation := range materialsInReturnerLocation {
			if materialInReturnerLocation.MaterialCostID == invoiceMaterial.MaterialCostID {
				materialInReturnerLocationIndex = index
				break
			}
		}

		materialsInReturnerLocation[materialInReturnerLocationIndex].Amount -= invoiceMaterial.Amount

		materialInAcceptorLocationIndex := -1
		for index, materialInAcceptorLocation := range materialsInAcceptorLocation {
			if materialInAcceptorLocation.MaterialCostID == invoiceMaterial.MaterialCostID {
				materialInAcceptorLocationIndex = index
				break
			}
		}

		if materialInAcceptorLocationIndex != -1 {
			materialsInAcceptorLocation[materialInAcceptorLocationIndex].Amount += invoiceMaterial.Amount
		} else {
			materialsInAcceptorLocation = append(materialsInAcceptorLocation, model.MaterialLocation{
				ProjectID:      invoiceReturn.ProjectID,
				MaterialCostID: invoiceMaterial.MaterialCostID,
				LocationType:   invoiceReturn.AcceptorType,
				LocationID:     invoiceReturn.AcceptorID,
				Amount:         invoiceMaterial.Amount,
			})
			materialInAcceptorLocationIndex = len(materialsInAcceptorLocation) - 1
		}

		if invoiceMaterial.IsDefected {
			materialDefectInDatabase, err := service.materialDefectRepo.GetByMaterialLocationID(materialsInReturnerLocation[materialInReturnerLocationIndex].ID)
			if err != nil {
				return err
			}

			if materialsInAcceptorLocation[materialInAcceptorLocationIndex].ID == 0 {
				newMaterialInAcceptorLocation := model.MaterialLocation{}

				if materialInAcceptorLocationIndex != -1 {
					newMaterialInAcceptorLocation = model.MaterialLocation{
						ProjectID:      invoiceReturn.ProjectID,
						MaterialCostID: invoiceMaterial.MaterialCostID,
						LocationType:   invoiceReturn.AcceptorType,
						LocationID:     invoiceReturn.AcceptorID,
						Amount:         invoiceMaterial.Amount,
					}
				} else {
					newMaterialInAcceptorLocation = materialsInAcceptorLocation[materialInAcceptorLocationIndex]

					materialsInAcceptorLocation = materialsInAcceptorLocation[:len(materialsInAcceptorLocation)-1]
				}
				materialDefectInDatabase.Amount = invoiceMaterial.Amount

				newMaterialsDefected = append(newMaterialsDefected, materialDefectInDatabase)
				newMaterialsInAcceptorLocationWithDefect = append(newMaterialsInAcceptorLocationWithDefect, newMaterialInAcceptorLocation)
			}

			if materialsInAcceptorLocation[materialInAcceptorLocationIndex].ID != 0 {
				materialDefectInDatabase.MaterialLocationID = materialsInAcceptorLocation[materialInAcceptorLocationIndex].ID
				materialDefectInDatabase.Amount += invoiceMaterial.Amount
				materialsDefected = append(materialsDefected, materialDefectInDatabase)
			}
		}
	}

	err = service.invoiceReturnRepo.Confirmation(dto.InvoiceReturnConfirmDataQuery{
		Invoice:                     invoiceReturn,
		MaterialsInReturnerLocation: materialsInReturnerLocation,
		MaterialsInAcceptorLocation: materialsInAcceptorLocation,
		MaterialsDefected:           materialsDefected,
		NewMaterialsInAcceptorLocationWithNewDefect: newMaterialsInAcceptorLocationWithDefect,
		NewMaterialsDefected:                        newMaterialsDefected,
	})

	return err
}

func (service *invoiceReturnService) UniqueCode(projectID uint) ([]string, error) {
	return service.invoiceReturnRepo.UniqueCode(projectID)
}

func (service *invoiceReturnService) UniqueTeam(projectID uint) ([]string, error) {
	var data []string
	teamIDs, err := service.invoiceReturnRepo.UniqueTeam(projectID)
	if err != nil {
		return data, err
	}

	for _, teamID := range teamIDs {
		team, err := service.teamRepo.GetByID(teamID)
		if err != nil {
			return []string{}, err
		}

		data = append(data, team.Number)
	}

	return data, err
}

func (service *invoiceReturnService) UniqueObject(projectID uint) ([]string, error) {
	var data []string
	objectIDs, err := service.invoiceReturnRepo.UniqueObject(projectID)
	if err != nil {
		return data, err
	}

	for _, objectID := range objectIDs {
		object, err := service.teamRepo.GetByID(objectID)
		if err != nil {
			return []string{}, err
		}

		data = append(data, object.Number)
	}

	return data, err
}

func (service *invoiceReturnService) Report(filter dto.InvoiceReturnReportFilterRequest, projectID uint) (string, error) {
	newFilter := dto.InvoiceReturnReportFilter{
		Code:     filter.Code,
		DateFrom: filter.DateFrom,
		DateTo:   filter.DateTo,
	}

	if filter.ReturnerType == "team" {
		newFilter.ReturnerType = "team"
		if filter.Returner != "" {
			team, err := service.teamRepo.GetByNumber(filter.Returner)
			if err != nil {
				return "", err
			}

			newFilter.ReturnerID = team.ID
		} else {
			newFilter.ReturnerID = 0
		}
	}

	if filter.ReturnerType == "object" {
		newFilter.ReturnerType = "object"
		if filter.Returner != "" {
			object, err := service.objectRepo.GetByName(filter.Returner)
			if err != nil {
				return "", err
			}

			newFilter.ReturnerID = object.ID
		} else {
			newFilter.ReturnerID = 0
		}
	}

	if filter.ReturnerType == "all" {
		newFilter.ReturnerType = ""
		newFilter.ReturnerID = 0
	}

	invoices, err := service.invoiceReturnRepo.ReportFilterData(newFilter, projectID)
	if err != nil {
		return "", err
	}

	templateFilePath := filepath.Join("./pkg/excels/templates/", "Invoice Return Report.xlsx")
	f, err := excelize.OpenFile(templateFilePath)
  defer f.Close()
	if err != nil {
		return "", err
	}

	sheetName := "Sheet1"
	rowCount := 2

	type InvoiceReturnReportData struct {
		DeliveryCode      string
		InvoiceReturnType string
		Returner          string
		DateOfInvoice     time.Time
		MaterialName      string
		MaterialUnit      string
		Amount            float64
		Price             float64
		IsDefected        string
		Notes             string
	}

	reportData := []InvoiceReturnReportData{}
	for _, invoice := range invoices {
		invoiceMaterialRepo, err := service.invoiceMaterialsRepo.GetByInvoice(projectID, invoice.ID, "return")
		if err != nil {
			return "", err
		}

		for _, invoiceMaterial := range invoiceMaterialRepo {
			oneEntry := InvoiceReturnReportData{
				DeliveryCode:  invoice.DeliveryCode,
				DateOfInvoice: invoice.DateOfInvoice,
        Amount: invoiceMaterial.Amount,
			}

			if invoice.ReturnerType == "team" {
				team, err := service.teamRepo.GetByID(invoice.ReturnerID)
				if err != nil {
					return "", err
				}

				oneEntry.InvoiceReturnType = "Бригада"
				oneEntry.Returner = team.Number
			}

			if invoice.ReturnerType == "object" {
				object, err := service.objectRepo.GetByID(invoice.ReturnerID)
				if err != nil {
					return "", err
				}

				oneEntry.InvoiceReturnType = "Бригада"
				oneEntry.Returner = object.Name
			}

			materialCost, err := service.materialCostRepo.GetByID(invoiceMaterial.MaterialCostID)
			if err != nil {
				return "", nil
			}

			material, err := service.materialRepo.GetByID(materialCost.MaterialID)
			if err != nil {
				return "", nil
			}

			oneEntry.MaterialName = material.Name
			oneEntry.MaterialUnit = material.Unit
			oneEntry.Price, _ = materialCost.CostM19.Float64()
			if invoiceMaterial.IsDefected {
				oneEntry.IsDefected = "Да"
			} else {
				oneEntry.IsDefected = "Нет"
			}
			oneEntry.Notes = invoiceMaterial.Notes

			reportData = append(reportData, oneEntry)
		}
	}

	for _, oneEntry := range reportData {
		f.SetCellValue(sheetName, "A"+fmt.Sprint(rowCount), oneEntry.DeliveryCode)
		f.SetCellValue(sheetName, "B"+fmt.Sprint(rowCount), oneEntry.InvoiceReturnType)
		f.SetCellValue(sheetName, "C"+fmt.Sprint(rowCount), oneEntry.Returner)

		dateOfInvoice := oneEntry.DateOfInvoice.String()
		dateOfInvoice = dateOfInvoice[:len(dateOfInvoice)-10]
		f.SetCellValue(sheetName, "D"+fmt.Sprint(rowCount), dateOfInvoice)

		f.SetCellValue(sheetName, "E"+fmt.Sprint(rowCount), oneEntry.MaterialName)
		f.SetCellValue(sheetName, "F"+fmt.Sprint(rowCount), oneEntry.MaterialUnit)
		f.SetCellValue(sheetName, "G"+fmt.Sprint(rowCount), oneEntry.Amount)
		f.SetCellValue(sheetName, "H"+fmt.Sprint(rowCount), oneEntry.Price)
		f.SetCellValue(sheetName, "I"+fmt.Sprint(rowCount), oneEntry.IsDefected)
		f.SetCellValue(sheetName, "J"+fmt.Sprint(rowCount), oneEntry.Notes)
		rowCount++
	}

	currentTime := time.Now()
	fileName := fmt.Sprintf(
		"Отсчет накладной возврат - %s.xlsx",
		currentTime.Format("02-01-2006"),
	)

	tempFilePath := filepath.Join("./pkg/excels/temp/", fileName)

	f.SaveAs(tempFilePath)

	return fileName, nil
}

func (service *invoiceReturnService) GetMaterialsInLocation(projectID, locationID uint, locationType string) ([]dto.InvoiceReturnMaterialForSelect, error) {

	materialsInLocation, err := service.materialLocationRepo.GetUniqueMaterialsFromLocation(projectID, locationID, locationType)
	if err != nil {
		return []dto.InvoiceReturnMaterialForSelect{}, err
	}

	var result []dto.InvoiceReturnMaterialForSelect
	for _, entry := range materialsInLocation {
		amount, err := service.materialLocationRepo.GetTotalAmountInLocation(projectID, entry.ID, locationID, locationType)
		if err != nil {
			return []dto.InvoiceReturnMaterialForSelect{}, err
		}

		result = append(result, dto.InvoiceReturnMaterialForSelect{
			MaterialID:      entry.ID,
			MaterialName:    entry.Name,
			MaterialUnit:    entry.Unit,
			Amount:          amount,
			HasSerialNumber: entry.HasSerialNumber,
		})
	}

	return result, nil
}

func (service *invoiceReturnService) GetMaterialCostInLocation(projectID, locationID, materialID uint, locationType string) ([]model.MaterialCost, error) {
	return service.materialLocationRepo.GetUniqueMaterialCostsFromLocation(projectID, materialID, locationID, locationType)
}

func (service *invoiceReturnService) GetMaterialAmountInLocation(projectID, locationID, materialCostID uint, locationType string) (float64, error) {
	return service.materialLocationRepo.GetUniqueMaterialTotalAmount(projectID, materialCostID, locationID, locationType)
}

func (service *invoiceReturnService) GetSerialNumberCodesInLocation(projectID, materialID uint, locationType string, locationID uint) ([]string, error) {
	return service.serialNumberRepo.GetCodesByMaterialIDAndLocation(projectID, materialID, locationType, locationID)
}

func (service *invoiceReturnService) GetInvoiceMaterialsWithoutSerialNumbers(id uint) ([]dto.InvoiceMaterialsWithoutSerialNumberView, error) {
	return service.invoiceMaterialsRepo.GetInvoiceMaterialsWithoutSerialNumbers(id, "return")
}

func (service *invoiceReturnService) GetInvoiceMaterialsWithSerialNumbers(id uint) ([]dto.InvoiceMaterialsWithSerialNumberView, error) {
	queryData, err := service.invoiceMaterialsRepo.GetInvoiceMaterialsWithSerialNumbers(id, "return")
	if err != nil {
		return []dto.InvoiceMaterialsWithSerialNumberView{}, err
	}

	result := []dto.InvoiceMaterialsWithSerialNumberView{}
	current := dto.InvoiceMaterialsWithSerialNumberView{}
	for index, materialInfo := range queryData {
		if index == 0 {
			current = dto.InvoiceMaterialsWithSerialNumberView{
				ID:            materialInfo.ID,
				MaterialName:  materialInfo.MaterialName,
				MaterialUnit:  materialInfo.MaterialUnit,
				IsDefected:    materialInfo.IsDefected,
				SerialNumbers: []string{},
				Amount:        materialInfo.Amount,
				CostM19:       materialInfo.CostM19,
				Notes:         materialInfo.Notes,
			}
		}

		if current.MaterialName == materialInfo.MaterialName && current.CostM19.Equal(materialInfo.CostM19) && current.IsDefected == materialInfo.IsDefected {
			if len(current.SerialNumbers) == 0 {
				current.SerialNumbers = append(current.SerialNumbers, materialInfo.SerialNumber)
				continue
			}

			if current.SerialNumbers[len(current.SerialNumbers)-1] != materialInfo.SerialNumber {
				current.SerialNumbers = append(current.SerialNumbers, materialInfo.SerialNumber)
			}

		} else {
			result = append(result, current)
			current = dto.InvoiceMaterialsWithSerialNumberView{
				ID:            materialInfo.ID,
				MaterialName:  materialInfo.MaterialName,
				MaterialUnit:  materialInfo.MaterialUnit,
				IsDefected:    materialInfo.IsDefected,
				SerialNumbers: []string{materialInfo.SerialNumber},
				Amount:        materialInfo.Amount,
				CostM19:       materialInfo.CostM19,
				Notes:         materialInfo.Notes,
			}
		}
	}

	if len(queryData) != 0 {
		result = append(result, current)
	}

	return result, nil
}

func (service *invoiceReturnService) GetMaterialsForEdit(id uint, locationType string, locationID uint) ([]dto.InvoiceReturnMaterialForEdit, error) {
	data, err := service.invoiceReturnRepo.GetMaterialsForEdit(id, locationType, locationID)
	if err != nil {
		return []dto.InvoiceReturnMaterialForEdit{}, nil
	}

	fmt.Println(data)

	var result []dto.InvoiceReturnMaterialForEdit
	for index, entry := range data {
		if index == 0 {
			result = append(result, entry)
			continue
		}

		lastItemIndex := len(result) - 1
		if result[lastItemIndex].MaterialID == entry.MaterialID {
			result[lastItemIndex].Amount += entry.Amount
			result[lastItemIndex].HolderAmount += entry.HolderAmount
		} else {
			result = append(result, entry)
		}
	}

	return result, nil

}

func (service *invoiceReturnService) GetMaterialAmountByMaterialID(projectID, materialID, locationID uint, locationType string) (float64, error) {
	return service.materialLocationRepo.GetTotalAmountInLocation(projectID, materialID, locationID, locationType)
}

func (service *invoiceReturnService) GenerateExcel(data dto.InvoiceReturn) error {

	templateFilePath := filepath.Join("./pkg/excels/templates/return.xlsx")
	f, err := excelize.OpenFile(templateFilePath)
	if err != nil {
		return err
	}
	sheetName := "Возврат"
	startingRow := 5

	f.InsertRows(sheetName, startingRow, len(data.Items))

	defaultStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Size:      8,
			VertAlign: "center",
			Family:    "Times New Roman",
		},
		Border: []excelize.Border{
			{Type: "left", Color: "#000000", Style: 1},
			{Type: "top", Color: "#000000", Style: 1},
			{Type: "right", Color: "#000000", Style: 1},
			{Type: "bottom", Color: "#000000", Style: 1},
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			WrapText:   true,
			Vertical:   "center",
		},
	})

	namingStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Size:      8,
			VertAlign: "center",
			Family:    "Times New Roman",
		},
		Border: []excelize.Border{
			{Type: "left", Color: "#000000", Style: 1},
			{Type: "top", Color: "#000000", Style: 1},
			{Type: "bottom", Color: "#000000", Style: 1},
		},
		Alignment: &excelize.Alignment{
			Horizontal: "left",
			Vertical:   "center",
			WrapText:   true,
		},
	})

	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Size:      10,
			VertAlign: "center",
			Bold:      true,
			Family:    "Times New Roman",
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "top",
			WrapText:   true,
		},
	})

	project, err := service.projectRepo.GetByID(data.Details.ProjectID)
	if err != nil {
		return err
	}

	district, err := service.districtRepo.GetByID(data.Details.DistrictID)
	if err != nil {
		return err
	}

	f.SetCellStyle(sheetName, "C1", "C1", headerStyle)
	f.SetCellStr(sheetName, "C1", fmt.Sprintf(`НАКЛАДНАЯ № %s
от %s года       
на возврат материала 
      `, data.Details.DeliveryCode, utils.DateConverter(data.Details.DateOfInvoice)))

	f.MergeCell(sheetName, "G1", "I1")
	f.SetCellStyle(sheetName, "G1", "I1", headerStyle)
	f.SetCellStr(sheetName, "G1", fmt.Sprintf(`%s
в г. Душанбе
Регион: %s 
      `, project.Name, district.Name))

	if data.Details.AcceptorType == "warehouse" {
		teamData, err := service.teamRepo.GetTeamNumberAndTeamLeadersByID(data.Details.ProjectID, data.Details.ReturnerID)
		if err != nil {
			return err
		}

		f.SetCellStr(sheetName, "B2", "")
		f.SetCellStr(sheetName, "B3", "")
		f.SetCellStr(sheetName, "C"+fmt.Sprint(6+len(data.Items)), fmt.Sprint(teamData[0].TeamLeaderName))
		f.SetCellStr(sheetName, "C"+fmt.Sprint(8+len(data.Items)), fmt.Sprint(teamData[0].TeamLeaderName))

		acceptor, err := service.workerRepo.GetByID(data.Details.AcceptedByWorkerID)
		if err != nil {
			return err
		}

		f.SetCellValue(sheetName, "C"+fmt.Sprint(10+len(data.Items)), fmt.Sprint(acceptor.Name))
	}

	if data.Details.AcceptorType == "team" {
		object, err := service.objectRepo.GetByID(data.Details.ReturnerID)
		if err != nil {
			return err
		}

		f.SetCellStr(sheetName, "D2", utils.ObjectTypeConverter(object.Type))
		f.SetCellStr(sheetName, "C3", object.Name)

		teamData, err := service.teamRepo.GetTeamNumberAndTeamLeadersByID(data.Details.ProjectID, data.Details.AcceptorID)
		if err != nil {
			return err
		}

		f.SetCellStr(sheetName, "C"+fmt.Sprint(6+len(data.Items)), fmt.Sprint(teamData[0].TeamLeaderName))
		f.SetCellStr(sheetName, "C"+fmt.Sprint(10+len(data.Items)), fmt.Sprint(teamData[0].TeamLeaderName))

		supervisorNames, err := service.objectSupervisorsRepo.GetSupervisorsNameByObjectID(data.Details.ReturnerID)
		if err != nil {
			return err
		}
		f.SetCellStr(sheetName, "C"+fmt.Sprint(8+len(data.Items)), supervisorNames[len(supervisorNames)-1])
	}

	for index, oneEntry := range data.Items {
		f.MergeCell(sheetName, "G"+fmt.Sprint(startingRow+index), "I"+fmt.Sprint(startingRow+index))

		f.SetCellStyle(sheetName, "A"+fmt.Sprint(startingRow+index), "I"+fmt.Sprint(startingRow+index), defaultStyle)
		f.SetCellStyle(sheetName, "B"+fmt.Sprint(startingRow+index), "B"+fmt.Sprint(startingRow+index), namingStyle)

		material, err := service.materialRepo.GetByID(oneEntry.MaterialID)
		if err != nil {
			return err
		}

		f.SetCellValue(sheetName, "A"+fmt.Sprint(startingRow+index), index+1)
		f.SetCellValue(sheetName, "B"+fmt.Sprint(startingRow+index), material.Code)
		f.SetCellValue(sheetName, "C"+fmt.Sprint(startingRow+index), material.Name)
		f.SetCellValue(sheetName, "D"+fmt.Sprint(startingRow+index), material.Unit)
		f.SetCellValue(sheetName, "E"+fmt.Sprint(startingRow+index), oneEntry.Amount)
		materialDefect := "Нет"
		if oneEntry.IsDefected {
			materialDefect = "Да"
		}
		f.SetCellValue(sheetName, "F"+fmt.Sprint(startingRow+index), materialDefect)
		f.SetCellValue(sheetName, "G"+fmt.Sprint(startingRow+index), oneEntry.Notes)
	}

	savePath := filepath.Join("./pkg/excels/return/", data.Details.DeliveryCode+".xlsx")
	f.SaveAs(savePath)
	if err := f.Close(); err != nil {
		fmt.Println(err)
	}

	return nil
}

func (service *invoiceReturnService) GetDocument(deliveryCode string) (string, error) {
	invoiceReturn, err := service.invoiceReturnRepo.GetByDeliveryCode(deliveryCode)
	if err != nil {
		return "", err
	}

	if invoiceReturn.Confirmation {
		return ".pdf", nil
	} else {
		return ".xlsx", nil
	}
}
