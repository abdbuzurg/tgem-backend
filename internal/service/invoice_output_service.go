package service

import (
	"backend-v2/internal/dto"
	"backend-v2/internal/repository"
	"backend-v2/model"
	"backend-v2/pkg/utils"
	"fmt"
	"strconv"

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
	GetUnconfirmedByObjectInvoices() ([]dto.InvoiceObject, error)
	Create(data dto.InvoiceOutput) (dto.InvoiceOutput, error)
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

func (service *invoiceOutputService) Create(data dto.InvoiceOutput) (dto.InvoiceOutput, error) {
	count, err := service.invoiceOutputRepo.Count(data.Details.ProjectID)
	if err != nil {
		return dto.InvoiceOutput{}, err
	}

	data.Details.DeliveryCode = utils.UniqueCodeGeneration("О", count+1, data.Details.ProjectID)
	invoiceOutput, err := service.invoiceOutputRepo.Create(data.Details)
	if err != nil {
		return dto.InvoiceOutput{}, err
	}

	data.Details = invoiceOutput

	for _, invoiceMaterial := range data.Items {
		materialCosts, err := service.materialCostRepo.GetByMaterialIDSorted(invoiceMaterial.MaterialID)
		if err != nil {
			return dto.InvoiceOutput{}, err
		}

		materialLocations := []model.MaterialLocation{}

		for _, materialCost := range materialCosts {
			materialLocation, err := service.materialLocationRepo.GetByMaterialCostIDOrCreate(invoiceOutput.ProjectID, materialCost.ID, "warehouse", 0)
			if err != nil {
				return dto.InvoiceOutput{}, err
			}

			materialLocations = append(materialLocations, materialLocation)
		}

		index := 0
		for invoiceMaterial.Amount > 0 {
			invoiceMaterialCreate := model.InvoiceMaterials{
				ProjectID:      invoiceOutput.ProjectID,
				ID:             0,
				MaterialCostID: materialCosts[index].ID,
				InvoiceID:      invoiceOutput.ID,
				InvoiceType:    "output",
				IsDefected:     false,
				Amount:         0,
				Notes:          "",
			}
			if materialLocations[index].Amount <= invoiceMaterial.Amount {
				invoiceMaterialCreate.Amount = materialLocations[index].Amount
				invoiceMaterial.Amount -= materialLocations[index].Amount
				materialLocations[index].Amount = 0
			} else {
				materialLocations[index].Amount -= invoiceMaterial.Amount
				invoiceMaterialCreate.Amount = invoiceMaterial.Amount
				invoiceMaterial.Amount = 0
			}

			invoiceMaterialCreate, err = service.invoiceMaterialRepo.Create(invoiceMaterialCreate)
			if err != nil {
				return dto.InvoiceOutput{}, err
			}

			index++
		}

		if len(invoiceMaterial.SerialNumbers) == 0 {
			continue
		}

		for _, code := range invoiceMaterial.SerialNumbers {
			serialNumber, err := service.serialNumberRepo.GetByCode(code)
			if err != nil {
				return dto.InvoiceOutput{}, err
			}

			invoiceMaterialSaved, err := service.invoiceMaterialRepo.GetByMaterialCostID(
				serialNumber.MaterialCostID,
				"output",
				invoiceOutput.ID,
			)
			if err != nil {
				return dto.InvoiceOutput{}, err
			}

			serialNumber.Status = "pending"
			serialNumber.StatusID = invoiceMaterialSaved.ID
			_, err = service.serialNumberRepo.Update(serialNumber)
			if err != nil {
				return dto.InvoiceOutput{}, err
			}
		}
	}

	f, err := excelize.OpenFile("./pkg/excels/templates/output.xlsx")
	if err != nil {
		return dto.InvoiceOutput{}, err
	}
	sheetName := "Отпуск"
	startingRow := 5
	currentInvoiceMaterails, err := service.invoiceMaterialRepo.GetByInvoice(invoiceOutput.ProjectID, invoiceOutput.ID, "output")
	f.InsertRows(sheetName, startingRow, len(currentInvoiceMaterails))

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

	for index, oneEntry := range currentInvoiceMaterails {
		materialCost, err := service.materialCostRepo.GetByID(oneEntry.MaterialCostID)
		if err != nil {
			return dto.InvoiceOutput{}, err
		}

		material, err := service.materialRepo.GetByID(materialCost.MaterialID)
		if err != nil {
			return dto.InvoiceOutput{}, err
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

	f.SaveAs("./pkg/excels/output/" + invoiceOutput.DeliveryCode + ".xlsx")
	if err := f.Close(); err != nil {
		fmt.Println(err)
	}

	return data, nil
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
	invoiceOutput, err = service.invoiceOutputRepo.Update(invoiceOutput)
	if err != nil {
		return err
	}

	invoiceMaterails, err := service.invoiceMaterialRepo.GetByInvoice(invoiceOutput.ProjectID, invoiceOutput.ID, "output")
	if err != nil {
		return err
	}

	for _, invoiceMaterial := range invoiceMaterails {
		oldLocation, err := service.materialLocationRepo.GetByMaterialCostIDOrCreate(invoiceOutput.ProjectID, invoiceMaterial.MaterialCostID, "warehouse", 0)
		if err != nil {
			return err
		}

		newLocation, err := service.materialLocationRepo.GetByMaterialCostIDOrCreate(invoiceOutput.ProjectID, invoiceMaterial.MaterialCostID, "teams", invoiceOutput.TeamID)
		if err != nil {
			return err
		}

		oldLocation.Amount -= invoiceMaterial.Amount
		newLocation.Amount += invoiceMaterial.Amount

		_, err = service.materialLocationRepo.Update(oldLocation)
		if err != nil {
			return err
		}

		_, err = service.materialLocationRepo.Update(newLocation)
		if err != nil {
			return err
		}

		serialNumbers, err := service.serialNumberRepo.GetByStatus("pending", invoiceMaterial.ID)
		if err != nil {
			return err
		}

		for _, serialNumber := range serialNumbers {
			serialNumber.Status = "teams"
			serialNumber.StatusID = newLocation.ID

			_, err = service.serialNumberRepo.Update(serialNumber)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (service *invoiceOutputService) GetUnconfirmedByObjectInvoices() ([]dto.InvoiceObject, error) {
	invoiceOutputs, err := service.invoiceOutputRepo.GetUnconfirmedByObjectInvoices()
	if err != nil {
		return []dto.InvoiceObject{}, err
	}

	result := []dto.InvoiceObject{}
	for _, invoiceOutput := range invoiceOutputs {
		team, err := service.teamRepo.GetByID(invoiceOutput.TeamID)
		if err != nil {
			return []dto.InvoiceObject{}, err
		}

		teamLeader, err := service.workerRepo.GetByID(team.LeaderWorkerID)
		if err != nil {
			return []dto.InvoiceObject{}, err
		}

		object, err := service.objectRepo.GetByID(invoiceOutput.ObjectID)
		if err != nil {
			return []dto.InvoiceObject{}, err
		}

		result = append(result, dto.InvoiceObject{
			ID:             invoiceOutput.ID,
			TeamLeaderName: teamLeader.Name,
			TeamNumber:     team.Number,
			ObjectName:     object.Name,
		})
	}

	return result, nil
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

	f, err := excelize.OpenFile("./pkg/excels/report/Invoice Output Report.xlsx")
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

	fileName := "Invoice Output Report " + fmt.Sprint(rowCount) + ".xlsx"
	f.SaveAs("./pkg/excels/report/" + fileName)
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
