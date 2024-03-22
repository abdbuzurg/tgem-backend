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
	}
}

type IInvoiceReturnService interface {
	GetAll() ([]model.InvoiceReturn, error)
	GetByID(id uint) (model.InvoiceReturn, error)
	GetPaginated(page, limit int, data model.InvoiceReturn) ([]dto.InvoiceReturnPaginated, error)
	Create(data dto.InvoiceReturn) (dto.InvoiceReturn, error)
	Delete(id uint) error
	Count(projectID uint) (int64, error)
	Confirmation(id uint) error
	UniqueCode(projectID uint) ([]string, error)
	UniqueTeam(projectID uint) ([]string, error)
	UniqueObject(projectID uint) ([]string, error)
	Report(filter dto.InvoiceReturnReportFilterRequest, projectID uint) (string, error)
}

func (service *invoiceReturnService) GetAll() ([]model.InvoiceReturn, error) {
	return service.invoiceReturnRepo.GetAll()
}

func (service *invoiceReturnService) GetByID(id uint) (model.InvoiceReturn, error) {
	return service.invoiceReturnRepo.GetByID(id)
}

func (service *invoiceReturnService) GetPaginated(page, limit int, data model.InvoiceReturn) ([]dto.InvoiceReturnPaginated, error) {
	result := []dto.InvoiceReturnPaginated{}
	invoiceReturns := []model.InvoiceReturn{}
	var err error
	if !utils.IsEmptyFields(data) {
		invoiceReturns, err = service.invoiceReturnRepo.GetPaginatedFiltered(page, limit, data)
	} else {
		invoiceReturns, err = service.invoiceReturnRepo.GetPaginated(page, limit)
		fmt.Println(invoiceReturns)
	}

	if err != nil {
		return []dto.InvoiceReturnPaginated{}, err
	}

	for _, invoiceReturn := range invoiceReturns {
		operatorAdd, err := service.workerRepo.GetByID(invoiceReturn.OperatorAddWorkerID)
		if err != nil {
			return []dto.InvoiceReturnPaginated{}, err
		}

		operatorEdit, err := service.workerRepo.GetByID(invoiceReturn.OperatorEditWorkerID)
		if err != nil {
			return []dto.InvoiceReturnPaginated{}, err
		}

		one := dto.InvoiceReturnPaginated{
			DateOfAdd:        invoiceReturn.DateOfAdd,
			DateOfEdit:       invoiceReturn.DateOfEdit,
			DateOfInvoice:    invoiceReturn.DateOfInvoice,
			ID:               invoiceReturn.ID,
			Notes:            invoiceReturn.Notes,
			OperatorAddName:  operatorAdd.Name,
			OperatorEditName: operatorEdit.Name,
			DeliveryCode:     invoiceReturn.DeliveryCode,
			ProjectName:      "",
			Confirmation:     invoiceReturn.Confirmation,
		}

		switch invoiceReturn.ReturnerType {
		case "teams":
			team, err := service.teamRepo.GetByID(invoiceReturn.ReturnerID)
			if err != nil {
				return []dto.InvoiceReturnPaginated{}, err
			}
			one.ReturnerName = team.Number
			one.ReturnerType = "Бригада"
		case "objects":
			object, err := service.objectRepo.GetByID(invoiceReturn.ReturnerID)
			if err != nil {
				return []dto.InvoiceReturnPaginated{}, err
			}
			one.ReturnerName = object.Name
			one.ReturnerType = "Объект"
		}

		result = append(result, one)
	}
	return result, nil
}

func (service *invoiceReturnService) Create(data dto.InvoiceReturn) (dto.InvoiceReturn, error) {
	invoiceReturn, err := service.invoiceReturnRepo.Create(data.Details)
	if err != nil {
		return dto.InvoiceReturn{}, err
	}
	deliveryCode := utils.UniqueCodeGeneration("В", invoiceReturn.ID)
	invoiceReturn.DeliveryCode = deliveryCode
	invoiceReturn, err = service.invoiceReturnRepo.Update(invoiceReturn)
	if err != nil {
		return data, nil
	}

	for _, invoiceMaterial := range data.Items {
		_, err = service.invoiceMaterialsRepo.Create(model.InvoiceMaterials{
			MaterialCostID: invoiceMaterial.MaterialCostID,
			InvoiceID:      invoiceReturn.ID,
			InvoiceType:    "return",
			Amount:         invoiceMaterial.Amount,
			IsDefected:     invoiceMaterial.IsDefected,
			Notes:          invoiceReturn.Notes,
		})
		if err != nil {
			return dto.InvoiceReturn{}, err
		}
	}

	f, err := excelize.OpenFile("./pkg/excels/templates/return.xlsx")
	if err != nil {
		return dto.InvoiceReturn{}, err
	}
	sheetName := "Возврат"
	startingRow := 5
	currentInvoiceMaterails, err := service.invoiceMaterialsRepo.GetByInvoice(invoiceReturn.ID, "return")
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
			return dto.InvoiceReturn{}, err
		}

		material, err := service.materialRepo.GetByID(materialCost.MaterialID)
		if err != nil {
			return dto.InvoiceReturn{}, err
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

	f.SaveAs("./pkg/excels/return/" + invoiceReturn.DeliveryCode + ".xlsx")
	if err := f.Close(); err != nil {
		fmt.Println(err)
	}

	return data, nil
}

func (service *invoiceReturnService) Delete(id uint) error {
	return service.invoiceReturnRepo.Delete(id)
}

func (service *invoiceReturnService) Count(projectID uint) (int64, error) {
	return service.invoiceReturnRepo.Count(projectID)
}

func (service *invoiceReturnService) Confirmation(id uint) error {
	invoiceReturn, err := service.invoiceReturnRepo.GetByID(id)
	if err != nil {
		return err
	}
	invoiceReturn.Confirmation = true
	invoiceReturn, err = service.invoiceReturnRepo.Update(invoiceReturn)
	if err != nil {
		return err
	}

	invoiceMaterials, err := service.invoiceMaterialsRepo.GetByInvoice(invoiceReturn.ID, "return")
	if err != nil {
		return err
	}

	for _, invoiceMaterial := range invoiceMaterials {
		oldLocation, err := service.materialLocationRepo.GetByMaterialCostIDOrCreate(invoiceMaterial.MaterialCostID, invoiceReturn.ReturnerType, invoiceReturn.ReturnerID)
		if err != nil {
			return err
		}

		newLocation, err := service.materialLocationRepo.GetByMaterialCostIDOrCreate(invoiceMaterial.MaterialCostID, "warehouse", 0)
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
	}

	return nil
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
		invoiceMaterialRepo, err := service.invoiceMaterialsRepo.GetByInvoice(invoice.ID, "output")
		if err != nil {
			return "", err
		}

		fmt.Println(invoiceMaterialRepo)
		for index, invoiceMaterial := range invoiceMaterialRepo {
			materialCost, err := service.materialCostRepo.GetByID(invoiceMaterial.MaterialCostID)
			if err != nil {
				return "", nil
			}

			material, err := service.materialRepo.GetByID(materialCost.MaterialID)
			if err != nil {
				return "", nil
			}

			fmt.Println(materialCost, material)
			if index == 0 {
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
			} else {
				f.SetCellValue(sheetName, "A"+fmt.Sprint(rowCount), "-")
				f.SetCellValue(sheetName, "B"+fmt.Sprint(rowCount), "-")
				f.SetCellValue(sheetName, "C"+fmt.Sprint(rowCount), "-")
				f.SetCellValue(sheetName, "D"+fmt.Sprint(rowCount), "-")
			}

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
