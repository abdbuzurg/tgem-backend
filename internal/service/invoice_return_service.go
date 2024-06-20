package service

import (
	"backend-v2/internal/dto"
	"backend-v2/internal/repository"
	"backend-v2/model"
	"backend-v2/pkg/utils"
	"fmt"

	"github.com/xuri/excelize/v2"
)

type invoiceReturnService struct {
	invoiceReturnRepo    repository.IInvoiceReturnRepository
	workerRepo           repository.IWorkerRepository
	objectRepo           repository.IObjectRepository
	teamRepo             repository.ITeamRepository
	materialLocationRepo repository.IMaterialLocationRepository
	invoiceMaterialsRepo repository.IInvoiceMaterialsRepository
	materialRepo         repository.IMaterialRepository
	materialCostRepo     repository.IMaterialCostRepository
	materialDefectRepo   repository.IMaterialDefectRepository
	serialNumberRepo     repository.ISerialNumberRepository
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
) IInvoiceReturnService {
	return &invoiceReturnService{
		invoiceReturnRepo:    invoiceReturnRepo,
		workerRepo:           workerRepo,
		objectRepo:           objectRepo,
		teamRepo:             teamRepo,
		materialLocationRepo: materialLocationRepo,
		invoiceMaterialsRepo: invoiceMaterialsRepo,
		materialRepo:         materialRepo,
		materialCostRepo:     materialCostRepo,
		materialDefectRepo:   materialDefectRepo,
		serialNumberRepo:     serialNumberRepo,
	}
}

type IInvoiceReturnService interface {
	GetAll() ([]model.InvoiceReturn, error)
	GetByID(id uint) (model.InvoiceReturn, error)
	GetPaginatedTeam(page, limit int, projectID uint) ([]dto.InvoiceReturnTeamPaginated, error)
	GetPaginatedObject(page, limit int, projectID uint) ([]dto.InvoiceReturnObjectPaginated, error)
	Create(data dto.InvoiceReturn) (model.InvoiceReturn, error)
	Delete(id uint) error
	CountBasedOnType(projectID uint, invoiceType string) (int64, error)
	Confirmation(id uint) error
	UniqueCode(projectID uint) ([]string, error)
	UniqueTeam(projectID uint) ([]string, error)
	UniqueObject(projectID uint) ([]string, error)
	Report(filter dto.InvoiceReturnReportFilterRequest, projectID uint) (string, error)
	GetMaterialsInLocation(projectID, locationID uint, locationType string) ([]model.Material, error)
	GetMaterialCostInLocation(projectID, locationID, materialID uint, locationType string) ([]model.MaterialCost, error)
	GetMaterialAmountInLocation(projectID, locationID, materialCostID uint, locationType string) (float64, error)
	GetSerialNumberCodesInLocation(projectID, materialID uint, locationType string, locationID uint) ([]string, error)
	GetInvoiceMaterialsWithoutSerialNumbers(id uint) ([]dto.InvoiceMaterialsWithoutSerialNumberView, error)
	GetInvoiceMaterialsWithSerialNumbers(id uint) ([]dto.InvoiceMaterialsWithSerialNumberView, error)
}

func (service *invoiceReturnService) GetAll() ([]model.InvoiceReturn, error) {
	return service.invoiceReturnRepo.GetAll()
}

func (service *invoiceReturnService) GetByID(id uint) (model.InvoiceReturn, error) {
	return service.invoiceReturnRepo.GetByID(id)
}

func (service *invoiceReturnService) GetPaginatedTeam(page, limit int, projectID uint) ([]dto.InvoiceReturnTeamPaginated, error) {
	invoiceReturnQueryData, err := service.invoiceReturnRepo.GetPaginatedTeam(page, limit, projectID)
	if err != nil {
		return []dto.InvoiceReturnTeamPaginated{}, err
	}

	result := []dto.InvoiceReturnTeamPaginated{}
	currentInvoice := dto.InvoiceReturnTeamPaginated{}
	for index, entry := range invoiceReturnQueryData {
		if index == 0 {
			currentInvoice = dto.InvoiceReturnTeamPaginated{
				ID:              entry.ID,
				DeliveryCode:    entry.DeliveryCode,
				DateOfInvoice:   entry.DateOfInvoice,
				TeamNumber:      entry.TeamNumber,
				TeamLeaderNames: []string{},
				Confirmation:    entry.Confirmation,
			}
		}

		if currentInvoice.ID == entry.ID {
			currentInvoice.TeamLeaderNames = append(currentInvoice.TeamLeaderNames, entry.TeamLeaderName)
		} else {
			result = append(result, currentInvoice)
			currentInvoice = dto.InvoiceReturnTeamPaginated{
				ID:              entry.ID,
				DeliveryCode:    entry.DeliveryCode,
				DateOfInvoice:   entry.DateOfInvoice,
				TeamNumber:      entry.TeamNumber,
				TeamLeaderNames: []string{entry.TeamLeaderName},
				Confirmation:    entry.Confirmation,
			}
		}
	}

	if len(invoiceReturnQueryData) != 0 {
		result = append(result, currentInvoice)
	}

	return result, nil
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
				ObjectName:            entry.ObjectName,
				ObjectSupervisorNames: []string{entry.ObjectSupervisorName},
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

	count, err := service.invoiceReturnRepo.Count(data.Details.ProjectID)
	if err != nil {
		return model.InvoiceReturn{}, err
	}

	data.Details.DeliveryCode = utils.UniqueCodeGeneration("В", count+1, data.Details.ProjectID)

	invoiceMaterialsForCreate := []model.InvoiceMaterials{}
	serialNumberMovements := []model.SerialNumberMovement{}
	for _, invoiceMaterial := range data.Items {
		invoiceMaterialsForCreate = append(invoiceMaterialsForCreate, model.InvoiceMaterials{
			ProjectID:      data.Details.ProjectID,
			MaterialCostID: invoiceMaterial.MaterialCostID,
			InvoiceID:      0,
			InvoiceType:    "return",
			Amount:         invoiceMaterial.Amount,
			IsDefected:     invoiceMaterial.IsDefected,
			Notes:          invoiceMaterial.Notes,
		})

		if len(invoiceMaterial.SerialNumbers) != 0 {
			serialNumbers, err := service.serialNumberRepo.GetSerialNumberIDsBySerialNumberCodes(invoiceMaterial.SerialNumbers)
			if err != nil {
				return model.InvoiceReturn{}, err
			}

			for _, serialNumber := range serialNumbers {
				serialNumberMovements = append(serialNumberMovements, model.SerialNumberMovement{
					ID:             0,
					SerialNumberID: serialNumber.ID,
					ProjectID:      data.Details.ProjectID,
					InvoiceID:      0,
					InvoiceType:    "return",
					IsDefected:     invoiceMaterial.IsDefected,
				})
			}
		}
	}

	invoiceReturn, err := service.invoiceReturnRepo.Create(dto.InvoiceReturnCreateQueryData{
		Invoice:               data.Details,
		InvoiceMaterials:      invoiceMaterialsForCreate,
		SerialNumberMovements: serialNumberMovements,
	})
	if err != nil {
		return model.InvoiceReturn{}, err
	}

	f, err := excelize.OpenFile("./pkg/excels/templates/return.xlsx")
	if err != nil {
		return model.InvoiceReturn{}, err
	}
	sheetName := "Возврат"
	startingRow := 5

	materialsForExcel, err := service.invoiceReturnRepo.GetInvoiceReturnMaterialsForExcel(invoiceReturn.ID)
	if err != nil {
		return model.InvoiceReturn{}, err
	}

	f.InsertRows(sheetName, startingRow, len(materialsForExcel))

	defaultStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Size:      8,
			VertAlign: "center",
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

	if invoiceReturn.AcceptorType == "warehouse" {
		invoiceReturnTeamDataExcel, err := service.invoiceReturnRepo.GetInvoiceReturnTeamDataForExcel(invoiceReturn.ID)
		if err != nil {
			return model.InvoiceReturn{}, err
		}

		f.SetCellValue(sheetName, "E1", fmt.Sprintf(`
НАКЛАДНАЯ № %s
от %s года       
на возврат материала 
      `, invoiceReturnTeamDataExcel.DeliveryCode, utils.DateConverter(invoiceReturnTeamDataExcel.DateOfInvoice)))
		f.SetCellValue(sheetName, "I1", fmt.Sprintf(`
%s
в г. Душанбе
Регион: %s 
      `, invoiceReturnTeamDataExcel.ProjectName, invoiceReturnTeamDataExcel.DistrictName))

		f.SetCellValue(sheetName, "A2", "")
		f.SetCellValue(sheetName, "A3", "")
		f.SetCellValue(sheetName, "E2", "")
		f.SetCellValue(sheetName, "I"+fmt.Sprint(6+len(materialsForExcel)), fmt.Sprint(invoiceReturnTeamDataExcel.TeamLeaderName))
		f.SetCellValue(sheetName, "I"+fmt.Sprint(9+len(materialsForExcel)), fmt.Sprint(invoiceReturnTeamDataExcel.TeamLeaderName))
		f.SetCellValue(sheetName, "D"+fmt.Sprint(9+len(materialsForExcel)), fmt.Sprint(invoiceReturnTeamDataExcel.AcceptorName))
	}

	if invoiceReturn.AcceptorType == "team" {
		invoiceReturnObjectDataExcel, err := service.invoiceReturnRepo.GetInvoiceReturnObjectDataForExcel(invoiceReturn.ID)
		if err != nil {
			return model.InvoiceReturn{}, err
		}

		f.SetCellValue(sheetName, "E1", fmt.Sprintf(`
НАКЛАДНАЯ № %s
от %s года       
на возврат материала 
      `, invoiceReturnObjectDataExcel.DeliveryCode, utils.DateConverter(invoiceReturnObjectDataExcel.DateOfInvoice)))
		f.SetCellValue(sheetName, "I1", fmt.Sprintf(`
%s
в г. Душанбе
Регион: %s 
      `, invoiceReturnObjectDataExcel.ProjectName, invoiceReturnObjectDataExcel.DistrictName))

		f.SetCellValue(sheetName, "D2", invoiceReturnObjectDataExcel.ObjectType)
		f.SetCellValue(sheetName, "D3", invoiceReturnObjectDataExcel.ObjectName)
		f.SetCellValue(sheetName, "I"+fmt.Sprint(6+len(materialsForExcel)), fmt.Sprint(invoiceReturnObjectDataExcel.TeamLeaderName))
		f.SetCellValue(sheetName, "I"+fmt.Sprint(9+len(materialsForExcel)), fmt.Sprint(invoiceReturnObjectDataExcel.SupervisorName))
		f.SetCellValue(sheetName, "D"+fmt.Sprint(9+len(materialsForExcel)), fmt.Sprint(invoiceReturnObjectDataExcel.TeamLeaderName))
	}

	for index, oneEntry := range materialsForExcel {
		f.MergeCell(sheetName, "D"+fmt.Sprint(startingRow+index), "F"+fmt.Sprint(startingRow+index))
		f.MergeCell(sheetName, "I"+fmt.Sprint(startingRow+index), "K"+fmt.Sprint(startingRow+index))

		f.SetCellStyle(sheetName, "A"+fmt.Sprint(startingRow+index), "K"+fmt.Sprint(startingRow+index), defaultStyle)
		f.SetCellStyle(sheetName, "B"+fmt.Sprint(startingRow+index), "B"+fmt.Sprint(startingRow+index), namingStyle)

		f.SetCellValue(sheetName, "A"+fmt.Sprint(startingRow+index), index+1)
		f.SetCellValue(sheetName, "B"+fmt.Sprint(startingRow+index), oneEntry.MaterialCode)
		f.SetCellValue(sheetName, "D"+fmt.Sprint(startingRow+index), oneEntry.MaterialName)
		f.SetCellValue(sheetName, "G"+fmt.Sprint(startingRow+index), oneEntry.MaterialUnit)
		f.SetCellValue(sheetName, "H"+fmt.Sprint(startingRow+index), oneEntry.MaterialAmount)
		materialDefect := ""
		if oneEntry.MaterialDefected {
			materialDefect = "Да"
		} else {
			materialDefect = "Нет"
		}
		f.SetCellValue(sheetName, "I"+fmt.Sprint(startingRow+index), materialDefect)
		f.SetCellValue(sheetName, "J"+fmt.Sprint(startingRow+index), oneEntry.MaterialNotes)
	}

	f.SaveAs("./pkg/excels/return/" + data.Details.DeliveryCode + ".xlsx")
	if err := f.Close(); err != nil {
		fmt.Println(err)
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

	if filter.ReturnerType == "teams" {
		newFilter.ReturnerType = "teams"
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

	if filter.ReturnerType == "objects" {
		newFilter.ReturnerType = "objects"
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

	f, err := excelize.OpenFile("./pkg/excels/report/Invoice Return Report.xlsx")
	if err != nil {
		return "", err
	}
	sheetName := "Sheet1"

	rowCount := 2
	for _, invoice := range invoices {
		invoiceMaterialRepo, err := service.invoiceMaterialsRepo.GetByInvoice(projectID, invoice.ID, "output")
		if err != nil {
			return "", err
		}

		fmt.Println(invoiceMaterialRepo)
		for _, invoiceMaterial := range invoiceMaterialRepo {
			materialCost, err := service.materialCostRepo.GetByID(invoiceMaterial.MaterialCostID)
			if err != nil {
				return "", nil
			}

			material, err := service.materialRepo.GetByID(materialCost.MaterialID)
			if err != nil {
				return "", nil
			}

			f.SetCellValue(sheetName, "A"+fmt.Sprint(rowCount), invoice.DeliveryCode)

			if invoice.ReturnerType == "teams" {
				f.SetCellValue(sheetName, "B"+fmt.Sprint(rowCount), "Бригада")

				team, err := service.teamRepo.GetByID(invoice.ReturnerID)
				if err != nil {
					return "", err
				}

				f.SetCellValue(sheetName, "C"+fmt.Sprint(rowCount), team.Number)
			} else {
				f.SetCellValue(sheetName, "B"+fmt.Sprint(rowCount), "Бригада")

				object, err := service.objectRepo.GetByID(invoice.ReturnerID)
				if err != nil {
					return "", err
				}

				f.SetCellValue(sheetName, "B"+fmt.Sprint(rowCount), object.Name)
			}

			dateOfInvoice := invoice.DateOfInvoice.String()
			dateOfInvoice = dateOfInvoice[:len(dateOfInvoice)-10]
			f.SetCellValue(sheetName, "D"+fmt.Sprint(rowCount), dateOfInvoice)

			f.SetCellValue(sheetName, "E"+fmt.Sprint(rowCount), material.Name)
			f.SetCellValue(sheetName, "F"+fmt.Sprint(rowCount), material.Unit)
			f.SetCellValue(sheetName, "G"+fmt.Sprint(rowCount), invoiceMaterial.Amount)
			f.SetCellValue(sheetName, "H"+fmt.Sprint(rowCount), materialCost.CostM19)
			f.SetCellValue(sheetName, "I"+fmt.Sprint(rowCount), invoiceMaterial.Notes)
			rowCount++
		}
	}

	fileName := "Invoice Return Report " + fmt.Sprint(rowCount) + ".xlsx"
	f.SaveAs("./pkg/excels/report/" + fileName)
	if err := f.Close(); err != nil {
		fmt.Println(err)
	}

	return fileName, nil
}

func (service *invoiceReturnService) GetMaterialsInLocation(projectID, locationID uint, locationType string) ([]model.Material, error) {
	return service.materialLocationRepo.GetUniqueMaterialsFromLocation(projectID, locationID, locationType)
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
