package service

import (
	"backend-v2/internal/dto"
	"backend-v2/internal/repository"
	"backend-v2/model"
	"backend-v2/pkg/utils"
	"fmt"
	"path/filepath"
	"strconv"
	"time"

	"github.com/xuri/excelize/v2"
)

type invoiceOutputService struct {
	invoiceOutputRepo    repository.IInvoiceOutputRepository
	invoiceMaterialRepo  repository.IInvoiceMaterialsRepository
	workerRepo           repository.IWorkerRepository
	teamRepo             repository.ITeamRepository
	objectRepo           repository.IObjectRepository
	materialCostRepo     repository.IMaterialCostRepository
	materialLocationRepo repository.IMaterialLocationRepository
	materialRepo         repository.IMaterialRepository
	districtRepo         repository.IDistrictRepository
	serialNumberRepo     repository.ISerialNumberRepository
}

func InitInvoiceOutputService(
	invoiceOutputRepo repository.IInvoiceOutputRepository,
	invoiceMaterialRepo repository.IInvoiceMaterialsRepository,
	workerRepo repository.IWorkerRepository,
	teamRepo repository.ITeamRepository,
	objectRepo repository.IObjectRepository,
	materialCostRepo repository.IMaterialCostRepository,
	materialLocationRepo repository.IMaterialLocationRepository,
	materialRepo repository.IMaterialRepository,
	districtRepo repository.IDistrictRepository,
	serialNumberRepo repository.ISerialNumberRepository,
) IInvoiceOutputService {
	return &invoiceOutputService{
		invoiceOutputRepo:    invoiceOutputRepo,
		invoiceMaterialRepo:  invoiceMaterialRepo,
		workerRepo:           workerRepo,
		teamRepo:             teamRepo,
		objectRepo:           objectRepo,
		materialCostRepo:     materialCostRepo,
		materialLocationRepo: materialLocationRepo,
		materialRepo:         materialRepo,
		districtRepo:         districtRepo,
		serialNumberRepo:     serialNumberRepo,
	}
}

type IInvoiceOutputService interface {
	GetAll() ([]model.InvoiceOutput, error)
	GetPaginated(page, limit int, data model.InvoiceOutput) ([]dto.InvoiceOutputPaginated, error)
	GetByID(id uint) (model.InvoiceOutput, error)
	GetInvoiceMaterialsWithoutSerialNumbers(id uint) ([]dto.InvoiceMaterialsWithoutSerialNumberView, error)
	GetInvoiceMaterialsWithSerialNumbers(id uint) ([]dto.InvoiceMaterialsWithSerialNumberView, error)
	Create(data dto.InvoiceOutput) (model.InvoiceOutput, error)
	// Update(data dto.InvoiceOutput) (dto.InvoiceOutput, error)
	Delete(id uint) error
	Count(projectID uint) (int64, error)
	Confirmation(id uint) error
	UniqueCode(projectID uint) ([]string, error)
	UniqueWarehouseManager(projectID uint) ([]string, error)
	UniqueRecieved(projectID uint) ([]string, error)
	UniqueDistrict(projectID uint) ([]string, error)
	UniqueObject(projectID uint) ([]string, error)
	UniqueTeam(projectID uint) ([]string, error)
	Report(filter dto.InvoiceOutputReportFilterRequest, projectID uint) (string, error)
	GetTotalMaterialAmount(projectID, materialID uint) (float64, error)
	GetSerialNumbersByMaterial(projectID, materialID uint) ([]string, error)
	GetAvailableMaterialsInWarehouse(projectID uint) ([]dto.AvailableMaterialsInWarehouse, error)
}

func (service *invoiceOutputService) GetAll() ([]model.InvoiceOutput, error) {
	return service.invoiceOutputRepo.GetAll()
}

func (service *invoiceOutputService) GetByID(id uint) (model.InvoiceOutput, error) {
	return service.invoiceOutputRepo.GetByID(id)
}

func (service *invoiceOutputService) GetPaginated(page, limit int, data model.InvoiceOutput) ([]dto.InvoiceOutputPaginated, error) {
	return service.invoiceOutputRepo.GetPaginatedFiltered(page, limit, data)
}

func (service *invoiceOutputService) GetInvoiceMaterialsWithoutSerialNumbers(id uint) ([]dto.InvoiceMaterialsWithoutSerialNumberView, error) {
	return service.invoiceMaterialRepo.GetInvoiceMaterialsWithoutSerialNumbers(id, "output")
}

func (service *invoiceOutputService) GetInvoiceMaterialsWithSerialNumbers(id uint) ([]dto.InvoiceMaterialsWithSerialNumberView, error) {
	queryData, err := service.invoiceMaterialRepo.GetInvoiceMaterialsWithSerialNumbers(id, "output")
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
				SerialNumbers: []string{},
				Amount:        materialInfo.Amount,
				CostM19:       materialInfo.CostM19,
				Notes:         materialInfo.Notes,
			}
		}

		if current.MaterialName == materialInfo.MaterialName && current.CostM19.Equal(materialInfo.CostM19) {
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

func (service *invoiceOutputService) Create(data dto.InvoiceOutput) (model.InvoiceOutput, error) {
	count, err := service.invoiceOutputRepo.Count(data.Details.ProjectID)
	if err != nil {
		return model.InvoiceOutput{}, err
	}

	data.Details.DeliveryCode = utils.UniqueCodeGeneration("О", count+1, data.Details.ProjectID)

	invoiceMaterialForCreate := []model.InvoiceMaterials{}
	serialNumberMovements := []model.SerialNumberMovement{}
	for _, invoiceMaterial := range data.Items {
		if len(invoiceMaterial.SerialNumbers) == 0 {
			materialInfoSorted, err := service.materialLocationRepo.GetMaterialAmountSortedByCostM19InLocation(data.Details.ProjectID, invoiceMaterial.MaterialID, "warehouse", 0)
			if err != nil {
				return model.InvoiceOutput{}, err
			}

			index := 0
			for invoiceMaterial.Amount > 0 {
				invoiceMaterialCreate := model.InvoiceMaterials{
					ProjectID:      data.Details.ProjectID,
					ID:             0,
					MaterialCostID: materialInfoSorted[index].MaterialCostID,
					InvoiceID:      0,
					InvoiceType:    "output",
					IsDefected:     false,
					Amount:         0,
					Notes:          invoiceMaterial.Notes,
				}

				if materialInfoSorted[index].MaterialAmount <= invoiceMaterial.Amount {
					invoiceMaterialCreate.Amount = materialInfoSorted[index].MaterialAmount
					invoiceMaterial.Amount -= materialInfoSorted[index].MaterialAmount
				} else {
					invoiceMaterialCreate.Amount = invoiceMaterial.Amount
					invoiceMaterial.Amount = 0
				}

				invoiceMaterialForCreate = append(invoiceMaterialForCreate, invoiceMaterialCreate)
				index++
			}

		}

		if len(invoiceMaterial.SerialNumbers) != 0 {
			MC_IDs_AND_SN_IDs, err := service.serialNumberRepo.GetMaterialCostIDsByCodesInLocation(invoiceMaterial.MaterialID, invoiceMaterial.SerialNumbers, "warehouse", 0)
			if err != nil {
				return model.InvoiceOutput{}, err
			}

			var invoiceMaterialCreate model.InvoiceMaterials
			for index, oneEntry := range MC_IDs_AND_SN_IDs {

				serialNumberMovements = append(serialNumberMovements, model.SerialNumberMovement{
					ID:             0,
					SerialNumberID: oneEntry.SerialNumberID,
					ProjectID:      data.Details.ProjectID,
					InvoiceID:      0,
					InvoiceType:    "output",
					IsDefected:     false,
					Confirmation:   false,
				})

				if index == 0 {
					invoiceMaterialCreate = model.InvoiceMaterials{
						ProjectID:      data.Details.ProjectID,
						ID:             0,
						MaterialCostID: oneEntry.MaterialCostID,
						InvoiceID:      data.Details.ID,
						InvoiceType:    "output",
						IsDefected:     false,
						Amount:         0,
						Notes:          invoiceMaterial.Notes,
					}
				}

				if oneEntry.MaterialCostID == invoiceMaterialCreate.MaterialCostID {
					invoiceMaterialCreate.Amount++
				} else {
					invoiceMaterialForCreate = append(invoiceMaterialForCreate, invoiceMaterialCreate)
					invoiceMaterialCreate = model.InvoiceMaterials{
						ProjectID:      data.Details.ProjectID,
						ID:             0,
						MaterialCostID: oneEntry.MaterialCostID,
						InvoiceID:      data.Details.ID,
						InvoiceType:    "output",
						IsDefected:     false,
						Amount:         0,
						Notes:          invoiceMaterial.Notes,
					}

				}

			}

			invoiceMaterialForCreate = append(invoiceMaterialForCreate, invoiceMaterialCreate)
		}

	}

	invoiceOutput, err := service.invoiceOutputRepo.Create(dto.InvoiceOutputCreateQueryData{
		Invoice:               data.Details,
		InvoiceMaterials:      invoiceMaterialForCreate,
		SerialNumberMovements: serialNumberMovements,
	})
	if err != nil {
		return model.InvoiceOutput{}, err
	}

	f, err := excelize.OpenFile("./pkg/excels/templates/output.xlsx")
	if err != nil {
		return model.InvoiceOutput{}, err
	}

	sheetName := "Отпуск"
	startingRow := 5
	f.InsertRows(sheetName, startingRow, len(data.Items))

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
			Horizontal: "left",
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

	invoiceOutputDescriptive, err := service.invoiceOutputRepo.GetDataForExcel(invoiceOutput.ID)
	if err != nil {
		return model.InvoiceOutput{}, err
	}

	f.SetCellValue(sheetName, "E1", fmt.Sprintf(`НАКЛАДНАЯ № %s
от %s года       
на отпуск материала 
`, invoiceOutput.DeliveryCode, utils.DateConverter(invoiceOutputDescriptive.DateOfInvoice)))

	f.SetCellValue(sheetName, "I1", fmt.Sprintf(`%s
Регион: %s `, invoiceOutputDescriptive.ProjectName, invoiceOutputDescriptive.DistrictName))
	f.SetCellValue(sheetName, "D2", fmt.Sprintf("%s", utils.ObjectTypeConverter(invoiceOutputDescriptive.ObjectType)))
	f.SetCellValue(sheetName, "D3", invoiceOutputDescriptive.ObjectName)

	for index, oneEntry := range data.Items {
		material, err := service.materialRepo.GetByID(oneEntry.MaterialID)
		if err != nil {
			return model.InvoiceOutput{}, err
		}

		f.MergeCell(sheetName, "D"+fmt.Sprint(startingRow+index), "F"+fmt.Sprint(startingRow+index))
		f.MergeCell(sheetName, "I"+fmt.Sprint(startingRow+index), "K"+fmt.Sprint(startingRow+index))

		f.SetCellStyle(sheetName, "A"+fmt.Sprint(startingRow+index), "K"+fmt.Sprint(startingRow+index), defaultStyle)
		f.SetCellStyle(sheetName, "B"+fmt.Sprint(startingRow+index), "B"+fmt.Sprint(startingRow+index), namingStyle)

		f.SetCellValue(sheetName, "A"+fmt.Sprint(startingRow+index), index+1)
		f.SetCellValue(sheetName, "B"+fmt.Sprint(startingRow+index), material.Code)
		f.SetCellValue(sheetName, "D"+fmt.Sprint(startingRow+index), material.Name)
		f.SetCellValue(sheetName, "G"+fmt.Sprint(startingRow+index), material.Unit)
		f.SetCellValue(sheetName, "H"+fmt.Sprint(startingRow+index), oneEntry.Amount)
		f.SetCellValue(sheetName, "I"+fmt.Sprint(startingRow+index), oneEntry.Notes)
	}

	f.SetCellValue(sheetName, "D"+fmt.Sprint(9+len(data.Items)), invoiceOutputDescriptive.ReleasedName)
	f.SetCellValue(sheetName, "I"+fmt.Sprint(6+len(data.Items)), invoiceOutputDescriptive.TeamLeaderName)
	f.SetCellValue(sheetName, "I"+fmt.Sprint(9+len(data.Items)), invoiceOutputDescriptive.RecipientName)

	f.SaveAs("./pkg/excels/output/" + data.Details.DeliveryCode + ".xlsx")
	if err := f.Close(); err != nil {
		fmt.Println(err)
	}

	return invoiceOutput, nil
}

func (service *invoiceOutputService) Delete(id uint) error {
	return service.invoiceOutputRepo.Delete(id)
}

func (service *invoiceOutputService) Count(projectID uint) (int64, error) {
	return service.invoiceOutputRepo.Count(projectID)
}

func (service *invoiceOutputService) Confirmation(id uint) error {
	invoiceOutput, err := service.invoiceOutputRepo.GetByID(id)
	if err != nil {
		return err
	}
	invoiceOutput.Confirmation = true

	invoiceMaterials, err := service.invoiceMaterialRepo.GetByInvoice(invoiceOutput.ProjectID, invoiceOutput.ID, "output")
	if err != nil {
		return err
	}

	materialsInWarehouse, err := service.materialLocationRepo.GetMaterialsInLocationBasedOnInvoiceID(0, "warehouse", id, "output")
	if err != nil {
		return err
	}

	materialsInTeam, err := service.materialLocationRepo.GetMaterialsInLocationBasedOnInvoiceID(invoiceOutput.TeamID, "team", id, "output")
	if err != nil {
		return err
	}

	for _, invoiceMaterial := range invoiceMaterials {
		materialInWarehouseIndex := -1
		for index, materialInWarehouse := range materialsInWarehouse {
			if materialInWarehouse.MaterialCostID == invoiceMaterial.MaterialCostID {
				materialInWarehouseIndex = index
				break
			}
		}

		materialsInWarehouse[materialInWarehouseIndex].Amount -= invoiceMaterial.Amount

		materialInTeamIndex := -1
		for index, materialInTeam := range materialsInTeam {
			if materialInTeam.MaterialCostID == invoiceMaterial.MaterialCostID {
				materialInTeamIndex = index
				break
			}
		}

		if materialInTeamIndex != -1 {
			materialsInTeam[materialInTeamIndex].Amount += invoiceMaterial.Amount
		} else {
			materialsInTeam = append(materialsInTeam, model.MaterialLocation{
				ProjectID:      invoiceOutput.ProjectID,
				MaterialCostID: invoiceMaterial.MaterialCostID,
				LocationType:   "team",
				LocationID:     invoiceOutput.TeamID,
				Amount:         invoiceMaterial.Amount,
			})
		}
	}

	err = service.invoiceOutputRepo.Confirmation(dto.InvoiceOutputConfirmationQueryData{
		InvoiceData:        invoiceOutput,
		WarehouseMaterials: materialsInWarehouse,
		TeamMaterials:      materialsInTeam,
	})

	return err
}

func (service *invoiceOutputService) UniqueCode(projectID uint) ([]string, error) {
	return service.invoiceOutputRepo.UniqueCode(projectID)
}

func (service *invoiceOutputService) UniqueWarehouseManager(projectID uint) ([]string, error) {
	ids, err := service.invoiceOutputRepo.UniqueWarehouseManager(projectID)
	if err != nil {
		return []string{}, err
	}

	result := []string{}
	for _, id := range ids {
		idconv, _ := strconv.ParseUint(id, 10, 32)
		warehouseManager, err := service.workerRepo.GetByID(uint(idconv))
		if err != nil {
			return []string{}, err
		}

		result = append(result, warehouseManager.Name)
	}

	return result, nil
}

func (service *invoiceOutputService) UniqueRecieved(projectID uint) ([]string, error) {
	ids, err := service.invoiceOutputRepo.UniqueWarehouseManager(projectID)
	if err != nil {
		return []string{}, err
	}

	result := []string{}
	for _, id := range ids {
		idconv, _ := strconv.ParseUint(id, 10, 32)
		recieved, err := service.workerRepo.GetByID(uint(idconv))
		if err != nil {
			return []string{}, err
		}

		result = append(result, recieved.Name)
	}

	return result, nil
}

func (service *invoiceOutputService) UniqueDistrict(projectID uint) ([]string, error) {
	ids, err := service.invoiceOutputRepo.UniqueWarehouseManager(projectID)
	if err != nil {
		return []string{}, err
	}

	result := []string{}
	for _, id := range ids {
		idconv, _ := strconv.ParseUint(id, 10, 32)
		district, err := service.districtRepo.GetByID(uint(idconv))
		if err != nil {
			return []string{}, err
		}

		result = append(result, district.Name)
	}

	return result, nil
}

func (service *invoiceOutputService) UniqueObject(projectID uint) ([]string, error) {
	ids, err := service.invoiceOutputRepo.UniqueWarehouseManager(projectID)
	if err != nil {
		return []string{}, err
	}

	result := []string{}
	for _, id := range ids {
		idconv, _ := strconv.ParseUint(id, 10, 32)
		object, err := service.objectRepo.GetByID(uint(idconv))
		if err != nil {
			return []string{}, err
		}

		result = append(result, object.Name)
	}

	return result, nil
}

func (service *invoiceOutputService) UniqueTeam(projectID uint) ([]string, error) {
	ids, err := service.invoiceOutputRepo.UniqueWarehouseManager(projectID)
	if err != nil {
		return []string{}, err
	}

	result := []string{}
	for _, id := range ids {
		idconv, _ := strconv.ParseUint(id, 10, 32)
		team, err := service.teamRepo.GetByID(uint(idconv))
		if err != nil {
			return []string{}, err
		}

		result = append(result, team.Number)
	}

	return result, nil
}

func (service *invoiceOutputService) Report(filter dto.InvoiceOutputReportFilterRequest, projectID uint) (string, error) {
	newFilter := dto.InvoiceOutputReportFilter{
		Code:     filter.Code,
		DateFrom: filter.DateFrom,
		DateTo:   filter.DateTo,
	}

	var err error
	if filter.WarehouseManager != "" {
		warehouseManager, err := service.workerRepo.GetByName(filter.WarehouseManager)
		if err != nil {
			return "", err
		}

		newFilter.WarehouseManagerID = warehouseManager.ID
	} else {
		newFilter.WarehouseManagerID = 0
	}

	if filter.Received != "" {
		released, err := service.workerRepo.GetByName(filter.Received)
		if err != nil {
			return "", err
		}

		newFilter.ReceivedID = released.ID
	} else {
		newFilter.ReceivedID = 0
	}

	if filter.District != "" {
		district, err := service.districtRepo.GetByName(filter.Received)
		if err != nil {
			return "", err
		}

		newFilter.DistrictID = district.ID
	} else {
		newFilter.DistrictID = 0
	}

	if filter.Team != "" {
		team, err := service.teamRepo.GetByNumber(filter.Team)
		if err != nil {
			return "", err
		}

		newFilter.TeamID = team.ID
	} else {
		newFilter.TeamID = 0
	}

	if filter.Object != "" {
		object, err := service.objectRepo.GetByName(filter.Object)
		if err != nil {
			return "", err
		}

		newFilter.ObjectID = object.ID
	} else {
		newFilter.ObjectID = 0
	}

	invoices, err := service.invoiceOutputRepo.ReportFilterData(newFilter, projectID)
	if err != nil {
		return "", err
	}

	templateFilePath := filepath.Join("./pkg/excels/templates/", "Invoice Output Report.xlsx")
	f, err := excelize.OpenFile(templateFilePath)
	if err != nil {
		return "", err
	}
	sheetName := "Sheet1"

	rowCount := 2
	for _, invoice := range invoices {
		invoiceMaterialRepo, err := service.invoiceMaterialRepo.GetByInvoice(projectID, invoice.ID, "output")
		if err != nil {
			return "", err
		}

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

			warehouseManager, err := service.workerRepo.GetByID(invoice.WarehouseManagerWorkerID)
			if err != nil {
				return "", err
			}
			f.SetCellValue(sheetName, "B"+fmt.Sprint(rowCount), warehouseManager.Name)

			recieved, err := service.workerRepo.GetByID(invoice.RecipientWorkerID)
			if err != nil {
				return "", err
			}
			f.SetCellValue(sheetName, "C"+fmt.Sprint(rowCount), recieved.Name)

			object, err := service.objectRepo.GetByID(invoice.ObjectID)
			if err != nil {
				return "", err
			}
			f.SetCellValue(sheetName, "D"+fmt.Sprint(rowCount), object.Name)

			team, err := service.teamRepo.GetByID(invoice.TeamID)
			if err != nil {
				return "", err
			}
			f.SetCellValue(sheetName, "E"+fmt.Sprint(rowCount), team.Number)

			dateOfInvoice := invoice.DateOfInvoice.String()
			dateOfInvoice = dateOfInvoice[:len(dateOfInvoice)-10]
			f.SetCellValue(sheetName, "F"+fmt.Sprint(rowCount), dateOfInvoice)

			f.SetCellValue(sheetName, "G"+fmt.Sprint(rowCount), material.Name)
			f.SetCellValue(sheetName, "H"+fmt.Sprint(rowCount), material.Unit)
			f.SetCellValue(sheetName, "I"+fmt.Sprint(rowCount), invoiceMaterial.Amount)
			f.SetCellValue(sheetName, "J"+fmt.Sprint(rowCount), materialCost.CostM19)
			f.SetCellValue(sheetName, "K"+fmt.Sprint(rowCount), invoiceMaterial.Notes)
			rowCount++
		}
	}

	currentTime := time.Now()
	fileName := fmt.Sprintf(
		"Отсчет накладной отпуск - %s.xlsx",
		currentTime.Format("02-01-2006"),
	)

	tempFilePath := filepath.Join("./pkg/excels/temp/", fileName)

	f.SaveAs(tempFilePath)

	if err := f.Close(); err != nil {
		fmt.Println(err)
	}

	return fileName, nil
}

func (service *invoiceOutputService) GetTotalMaterialAmount(projectID, materialID uint) (float64, error) {
	return service.materialLocationRepo.GetTotalAmountInWarehouse(projectID, materialID)
}

func (service *invoiceOutputService) GetSerialNumbersByMaterial(projectID, materialID uint) ([]string, error) {
	return service.serialNumberRepo.GetCodesByMaterialID(projectID, materialID, "warehouse")
}

func (service *invoiceOutputService) GetAvailableMaterialsInWarehouse(projectID uint) ([]dto.AvailableMaterialsInWarehouse, error) {
	data, err := service.invoiceOutputRepo.GetAvailableMaterialsInWarehouse(projectID)
	if err != nil {
		return []dto.AvailableMaterialsInWarehouse{}, err
	}

	result := []dto.AvailableMaterialsInWarehouse{}
	currentMaterial := dto.AvailableMaterialsInWarehouse{}
	for index, oneEntry := range data {
		if currentMaterial.ID == oneEntry.ID {
			currentMaterial.Amount += oneEntry.Amount
		} else {
			if index != 0 {
				result = append(result, currentMaterial)
			}
			currentMaterial = oneEntry
		}
	}

	if len(data) != 0 {
		result = append(result, currentMaterial)
	}

	return result, err

}
