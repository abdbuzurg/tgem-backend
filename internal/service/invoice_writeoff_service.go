package service

import (
	"backend-v2/internal/dto"
	"backend-v2/internal/repository"
	"backend-v2/model"
	"backend-v2/pkg/utils"
	"fmt"

	"github.com/xuri/excelize/v2"
)

type invoiceWriteOffService struct {
	invoiceWriteOffRepo  repository.IInvoiceWriteOffRepository
	workerRepo           repository.IWorkerRepository
	objectRepo           repository.IObjectRepository
	teamRepo             repository.ITeamRepository
	materialLocationRepo repository.IMaterialLocationRepository
	invoiceMaterialsRepo repository.IInvoiceMaterialsRepository
	materialRepo         repository.IMaterialRepository
	materialCostRepo     repository.IMaterialCostRepository
}

func InitInvoiceWriteOffService(
	invoiceWriteOffRepo repository.IInvoiceWriteOffRepository,
	workerRepo repository.IWorkerRepository,
	objectRepo repository.IObjectRepository,
	teamRepo repository.ITeamRepository,
	materialLocationRepo repository.IMaterialLocationRepository,
	invoiceMaterialsRepo repository.IInvoiceMaterialsRepository,
	materialRepo repository.IMaterialRepository,
	materialCostRepo repository.IMaterialCostRepository,
) IInvoiceWriteOffService {
	return &invoiceWriteOffService{
		invoiceWriteOffRepo:  invoiceWriteOffRepo,
		workerRepo:           workerRepo,
		objectRepo:           objectRepo,
		teamRepo:             teamRepo,
		materialLocationRepo: materialLocationRepo,
		invoiceMaterialsRepo: invoiceMaterialsRepo,
		materialRepo:         materialRepo,
		materialCostRepo:     materialCostRepo,
	}
}

type IInvoiceWriteOffService interface {
	GetAll() ([]model.InvoiceWriteOff, error)
	GetPaginated(page, limit int, data model.InvoiceWriteOff) ([]dto.InvoiceWriteOffPaginated, error)
	Create(data dto.InvoiceWriteOff) (dto.InvoiceWriteOff, error)
	// Create( data dto.InvoiceWriteOff) (dto.InvoiceWriteOff, error)
	Delete(id uint) error
	Count() (int64, error)
}

func (service *invoiceWriteOffService) GetAll() ([]model.InvoiceWriteOff, error) {
	return service.invoiceWriteOffRepo.GetAll()
}

func (service *invoiceWriteOffService) GetPaginated(page, limit int, data model.InvoiceWriteOff) ([]dto.InvoiceWriteOffPaginated, error) {
	result := []dto.InvoiceWriteOffPaginated{}
	invoiceWriteOffs := []model.InvoiceWriteOff{}
	var err error
	if !utils.IsEmptyFields(data) {
		invoiceWriteOffs, err = service.invoiceWriteOffRepo.GetPaginatedFiltered(page, limit, data)
	} else {
		invoiceWriteOffs, err = service.invoiceWriteOffRepo.GetPaginated(page, limit)
	}

	if err != nil {
		return []dto.InvoiceWriteOffPaginated{}, err
	}

	for _, invoiceWriteOff := range invoiceWriteOffs {
		operatorAdd, err := service.workerRepo.GetByID(invoiceWriteOff.OperatorAddWorkerID)
		if err != nil {
			return []dto.InvoiceWriteOffPaginated{}, err
		}

		operatorEdit, err := service.workerRepo.GetByID(invoiceWriteOff.OperatorEditWorkerID)
		if err != nil {
			return []dto.InvoiceWriteOffPaginated{}, err
		}

		one := dto.InvoiceWriteOffPaginated{
			DateOfAdd:        invoiceWriteOff.DateOfAdd,
			DateOfEdit:       invoiceWriteOff.DateOfEdit,
			DateOfInvoice:    invoiceWriteOff.DateOfInvoice,
			ID:               invoiceWriteOff.ID,
			Notes:            invoiceWriteOff.Notes,
			OperatorAddName:  operatorAdd.Name,
			OperatorEditName: operatorEdit.Name,
			DeliveryCode:     invoiceWriteOff.DeliveryCode,
			WriteOffType:     invoiceWriteOff.WriteOffType,
		}

		result = append(result, one)
	}
	return result, nil
}

func (service *invoiceWriteOffService) Create(data dto.InvoiceWriteOff) (dto.InvoiceWriteOff, error) {
	invoiceWriteOff, err := service.invoiceWriteOffRepo.Create(data.Details)
	if err != nil {
		return dto.InvoiceWriteOff{}, err
	}

	invoiceWriteOff.DeliveryCode = utils.UniqueCodeGeneration("C", invoiceWriteOff.ID)
	invoiceWriteOff, err = service.invoiceWriteOffRepo.Update(invoiceWriteOff)
	if err != nil {
		return dto.InvoiceWriteOff{}, err
	}

	for _, invoiceMaterial := range data.Items {
		_, err := service.invoiceMaterialsRepo.Create(model.InvoiceMaterials{
			MaterialCostID: invoiceMaterial.MaterialCostID,
			InvoiceID:      invoiceWriteOff.ID,
			InvoiceType:    "writeoff",
			Amount:         invoiceMaterial.Amount,
			Notes:          "",
		})
		if err != nil {
			return dto.InvoiceWriteOff{}, err
		}

		_, err = service.materialLocationRepo.GetByMaterialCostIDOrCreate(invoiceMaterial.MaterialCostID, "writeoff-"+invoiceWriteOff.DeliveryCode, 0)
		if err != nil {
			return dto.InvoiceWriteOff{}, err
		}

	}

	switch data.Details.WriteOffType {
	case "акт склад 1":
		f, err := excelize.OpenFile("./pkg/excels/templates/act_warehouse_1.xlsx")
		if err != nil {
			return dto.InvoiceWriteOff{}, err
		}
		sheetName := "Sheet1"
		startingRow := 21
		currentInvoiceMaterails, err := service.invoiceMaterialsRepo.GetByInvoice(invoiceWriteOff.ID, "writeoff")
		f.InsertRows(sheetName, startingRow, len(currentInvoiceMaterails))
		defaultStyle, _ := f.NewStyle(&excelize.Style{
			Font: &excelize.Font{
				Size:      12,
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
				Size:      12,
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

		totalAmount := 0.0
		totalCost := 0.0
		for index, oneEntry := range currentInvoiceMaterails {
			materialCost, err := service.materialCostRepo.GetByID(oneEntry.MaterialCostID)
			if err != nil {
				return dto.InvoiceWriteOff{}, err
			}

			material, err := service.materialRepo.GetByID(materialCost.MaterialID)
			if err != nil {
				return dto.InvoiceWriteOff{}, err
			}
			f.SetCellStyle(sheetName, "A"+fmt.Sprint(startingRow+index), "G"+fmt.Sprint(startingRow+index), defaultStyle)
			f.SetCellStyle(sheetName, "B"+fmt.Sprint(startingRow+index), "B"+fmt.Sprint(startingRow+index), namingStyle)

			f.SetCellValue(sheetName, "A"+fmt.Sprint(startingRow+index), index+1)
			f.SetCellValue(sheetName, "B"+fmt.Sprint(startingRow+index), material.Name)
			f.SetCellValue(sheetName, "C"+fmt.Sprint(startingRow+index), invoiceWriteOff.DateOfInvoice.Format("02-01-2006"))
			f.SetCellValue(sheetName, "D"+fmt.Sprint(startingRow+index), material.Unit)
			f.SetCellValue(sheetName, "E"+fmt.Sprint(startingRow+index), oneEntry.Amount)
			f.SetCellValue(sheetName, "F"+fmt.Sprint(startingRow+index), materialCost.CostWithCustomer)
			price, _ := materialCost.CostWithCustomer.Float64()
			f.SetCellValue(sheetName, "G"+fmt.Sprint(startingRow+index), fmt.Sprintf("%.2f", price*oneEntry.Amount))

			totalAmount += oneEntry.Amount
			totalCost += price * oneEntry.Amount
		}

		f.SetCellValue(sheetName, "A"+fmt.Sprint(startingRow+len(currentInvoiceMaterails)), len(currentInvoiceMaterails))
		f.SetCellValue(sheetName, "E"+fmt.Sprint(startingRow+len(currentInvoiceMaterails)), totalAmount)
		f.SetCellValue(sheetName, "G"+fmt.Sprint(startingRow+len(currentInvoiceMaterails)), totalCost)

		f.SaveAs("./pkg/excels/writeoff/" + invoiceWriteOff.DeliveryCode + ".xlsx")
		if err := f.Close(); err != nil {
			fmt.Println(err)
		}

	case "акт склад 2":
		f, err := excelize.OpenFile("./pkg/excels/templates/act_warehouse_2.xlsx")
		if err != nil {
			return dto.InvoiceWriteOff{}, err
		}
		sheetName := "Sheet1"
		startingRow := 26
		currentInvoiceMaterails, err := service.invoiceMaterialsRepo.GetByInvoice(invoiceWriteOff.ID, "writeoff")
		f.InsertRows(sheetName, startingRow, len(currentInvoiceMaterails))
		defaultStyle, _ := f.NewStyle(&excelize.Style{
			Font: &excelize.Font{
				Size:      12,
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
				Size:      12,
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

		totalAmount := 0.0
		totalCost := 0.0
		for index, oneEntry := range currentInvoiceMaterails {
			materialCost, err := service.materialCostRepo.GetByID(oneEntry.MaterialCostID)
			if err != nil {
				return dto.InvoiceWriteOff{}, err
			}

			material, err := service.materialRepo.GetByID(materialCost.MaterialID)
			if err != nil {
				return dto.InvoiceWriteOff{}, err
			}
			f.SetCellStyle(sheetName, "A"+fmt.Sprint(startingRow+index), "G"+fmt.Sprint(startingRow+index), defaultStyle)
			f.SetCellStyle(sheetName, "B"+fmt.Sprint(startingRow+index), "B"+fmt.Sprint(startingRow+index), namingStyle)

			f.SetCellValue(sheetName, "A"+fmt.Sprint(startingRow+index), index+1)
			f.SetCellValue(sheetName, "B"+fmt.Sprint(startingRow+index), material.Name)
			f.SetCellValue(sheetName, "C"+fmt.Sprint(startingRow+index), material.Unit)
			f.SetCellValue(sheetName, "D"+fmt.Sprint(startingRow+index), materialCost.CostWithCustomer)
			f.SetCellValue(sheetName, "E"+fmt.Sprint(startingRow+index), oneEntry.Amount)
			price, _ := materialCost.CostWithCustomer.Float64()
			f.SetCellValue(sheetName, "F"+fmt.Sprint(startingRow+index), fmt.Sprintf("%.2f", price*oneEntry.Amount))

			totalAmount += oneEntry.Amount
			totalCost += price * oneEntry.Amount
		}

		f.SetCellValue(sheetName, "A"+fmt.Sprint(startingRow+len(currentInvoiceMaterails)), len(currentInvoiceMaterails))
		f.SetCellValue(sheetName, "E"+fmt.Sprint(startingRow+len(currentInvoiceMaterails)), totalAmount)
		f.SetCellValue(sheetName, "F"+fmt.Sprint(startingRow+len(currentInvoiceMaterails)), totalCost)

		f.SaveAs("./pkg/excels/writeoff/" + invoiceWriteOff.DeliveryCode + ".xlsx")
		if err := f.Close(); err != nil {
			fmt.Println(err)
		}
	case "акт склад 3":
		f, err := excelize.OpenFile("./pkg/excels/templates/act_warehouse_3.xlsx")
		if err != nil {
			return dto.InvoiceWriteOff{}, err
		}
		sheetName := "Sheet1"
		startingRow := 21
		currentInvoiceMaterails, err := service.invoiceMaterialsRepo.GetByInvoice(invoiceWriteOff.ID, "writeoff")
		f.InsertRows(sheetName, startingRow, len(currentInvoiceMaterails))
		defaultStyle, _ := f.NewStyle(&excelize.Style{
			Font: &excelize.Font{
				Size:      12,
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
				Size:      12,
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

		totalAmount := 0.0
		totalCost := 0.0
		for index, oneEntry := range currentInvoiceMaterails {
			materialCost, err := service.materialCostRepo.GetByID(oneEntry.MaterialCostID)
			if err != nil {
				return dto.InvoiceWriteOff{}, err
			}

			material, err := service.materialRepo.GetByID(materialCost.MaterialID)
			if err != nil {
				return dto.InvoiceWriteOff{}, err
			}
			f.SetCellStyle(sheetName, "A"+fmt.Sprint(startingRow+index), "F"+fmt.Sprint(startingRow+index), defaultStyle)
			f.SetCellStyle(sheetName, "B"+fmt.Sprint(startingRow+index), "B"+fmt.Sprint(startingRow+index), namingStyle)

			f.SetCellValue(sheetName, "A"+fmt.Sprint(startingRow+index), index+1)
			f.SetCellValue(sheetName, "B"+fmt.Sprint(startingRow+index), material.Name)
			f.SetCellValue(sheetName, "C"+fmt.Sprint(startingRow+index), material.Unit)
			f.SetCellValue(sheetName, "D"+fmt.Sprint(startingRow+index), oneEntry.Amount)
			f.SetCellValue(sheetName, "E"+fmt.Sprint(startingRow+index), materialCost.CostWithCustomer)
			price, _ := materialCost.CostWithCustomer.Float64()
			f.SetCellValue(sheetName, "F"+fmt.Sprint(startingRow+index), fmt.Sprintf("%.2f", price*oneEntry.Amount))

			totalAmount += oneEntry.Amount
			totalCost += price * oneEntry.Amount
		}

		f.SetCellValue(sheetName, "A"+fmt.Sprint(startingRow+len(currentInvoiceMaterails)), len(currentInvoiceMaterails))
		f.SetCellValue(sheetName, "D"+fmt.Sprint(startingRow+len(currentInvoiceMaterails)), totalAmount)
		f.SetCellValue(sheetName, "F"+fmt.Sprint(startingRow+len(currentInvoiceMaterails)), totalCost)

		f.SaveAs("./pkg/excels/writeoff/" + invoiceWriteOff.DeliveryCode + ".xlsx")
		if err := f.Close(); err != nil {
			fmt.Println(err)
		}
	case "акт ГСМ":
		f, err := excelize.OpenFile("./pkg/excels/templates/act_gsm.xlsx")
		if err != nil {
			return dto.InvoiceWriteOff{}, err
		}
		sheetName := "Sheet1"
		startingRow := 22
		currentInvoiceMaterails, err := service.invoiceMaterialsRepo.GetByInvoice(invoiceWriteOff.ID, "writeoff")
		f.InsertRows(sheetName, startingRow, len(currentInvoiceMaterails))
		defaultStyle, _ := f.NewStyle(&excelize.Style{
			Font: &excelize.Font{
				Size:      12,
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
				Size:      12,
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

		totalAmount := 0.0
		totalCost := 0.0
		for index, oneEntry := range currentInvoiceMaterails {
			materialCost, err := service.materialCostRepo.GetByID(oneEntry.MaterialCostID)
			if err != nil {
				return dto.InvoiceWriteOff{}, err
			}

			material, err := service.materialRepo.GetByID(materialCost.MaterialID)
			if err != nil {
				return dto.InvoiceWriteOff{}, err
			}

			f.MergeCell(sheetName, "B"+fmt.Sprint(startingRow+index), "C"+fmt.Sprint(startingRow+index))

			f.SetCellStyle(sheetName, "A"+fmt.Sprint(startingRow+index), "G"+fmt.Sprint(startingRow+index), defaultStyle)
			f.SetCellStyle(sheetName, "B"+fmt.Sprint(startingRow+index), "B"+fmt.Sprint(startingRow+index), namingStyle)

			f.SetCellValue(sheetName, "A"+fmt.Sprint(startingRow+index), index+1)
			f.SetCellValue(sheetName, "B"+fmt.Sprint(startingRow+index), material.Name)
			f.SetCellValue(sheetName, "D"+fmt.Sprint(startingRow+index), material.Unit)
			f.SetCellValue(sheetName, "E"+fmt.Sprint(startingRow+index), oneEntry.Amount)
			f.SetCellValue(sheetName, "F"+fmt.Sprint(startingRow+index), materialCost.CostWithCustomer)
			price, _ := materialCost.CostWithCustomer.Float64()
			f.SetCellValue(sheetName, "G"+fmt.Sprint(startingRow+index), fmt.Sprintf("%.2f", price*oneEntry.Amount))

			totalAmount += oneEntry.Amount
			totalCost += price * oneEntry.Amount
		}

		f.SetCellValue(sheetName, "A"+fmt.Sprint(startingRow+len(currentInvoiceMaterails)), len(currentInvoiceMaterails))
		f.SetCellValue(sheetName, "E"+fmt.Sprint(startingRow+len(currentInvoiceMaterails)), totalAmount)
		f.SetCellValue(sheetName, "G"+fmt.Sprint(startingRow+len(currentInvoiceMaterails)), totalCost)

		f.SaveAs("./pkg/excels/writeoff/" + invoiceWriteOff.DeliveryCode + ".xlsx")
		if err := f.Close(); err != nil {
			fmt.Println(err)
		}
	case "акт ПТО утерени брига":
		f, err := excelize.OpenFile("./pkg/excels/templates/act_pto_uteri_brigadi.xlsx")
		if err != nil {
			return dto.InvoiceWriteOff{}, err
		}
		sheetName := "Sheet1"
		startingRow := 12
		currentInvoiceMaterails, err := service.invoiceMaterialsRepo.GetByInvoice(invoiceWriteOff.ID, "writeoff")
		f.InsertRows(sheetName, startingRow, len(currentInvoiceMaterails))
		defaultStyle, _ := f.NewStyle(&excelize.Style{
			Font: &excelize.Font{
				Size:      12,
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
				Size:      12,
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
				return dto.InvoiceWriteOff{}, err
			}

			material, err := service.materialRepo.GetByID(materialCost.MaterialID)
			if err != nil {
				return dto.InvoiceWriteOff{}, err
			}

			f.SetCellStyle(sheetName, "A"+fmt.Sprint(startingRow+index), "D"+fmt.Sprint(startingRow+index), defaultStyle)
			f.SetCellStyle(sheetName, "B"+fmt.Sprint(startingRow+index), "B"+fmt.Sprint(startingRow+index), namingStyle)

			f.SetCellValue(sheetName, "A"+fmt.Sprint(startingRow+index), index+1)
			f.SetCellValue(sheetName, "B"+fmt.Sprint(startingRow+index), material.Name)
			f.SetCellValue(sheetName, "C"+fmt.Sprint(startingRow+index), material.Unit)
			f.SetCellValue(sheetName, "D"+fmt.Sprint(startingRow+index), oneEntry.Amount)

		}

		f.SaveAs("./pkg/excels/writeoff/" + invoiceWriteOff.DeliveryCode + ".xlsx")
		if err := f.Close(); err != nil {
			fmt.Println(err)
		}
	case "акт ПТО материал пас":
	case "акт ПТО услуго":
	case "Акт ПТО M19":
		f, err := excelize.OpenFile("./pkg/excels/templates/act_pto_m19.xlsx")
		if err != nil {
			return dto.InvoiceWriteOff{}, err
		}
		sheetName := "Sheet1"
		startingRow := 17
		currentInvoiceMaterails, err := service.invoiceMaterialsRepo.GetByInvoice(invoiceWriteOff.ID, "writeoff")
		f.InsertRows(sheetName, startingRow, len(currentInvoiceMaterails))
		defaultStyle, _ := f.NewStyle(&excelize.Style{
			Font: &excelize.Font{
				Size:      12,
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
				Size:      12,
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
				return dto.InvoiceWriteOff{}, err
			}

			material, err := service.materialRepo.GetByID(materialCost.MaterialID)
			if err != nil {
				return dto.InvoiceWriteOff{}, err
			}

			f.SetCellStyle(sheetName, "A"+fmt.Sprint(startingRow+index), "D"+fmt.Sprint(startingRow+index), defaultStyle)
			f.SetCellStyle(sheetName, "B"+fmt.Sprint(startingRow+index), "B"+fmt.Sprint(startingRow+index), namingStyle)

			f.SetCellValue(sheetName, "A"+fmt.Sprint(startingRow+index), index+1)
			f.SetCellValue(sheetName, "B"+fmt.Sprint(startingRow+index), material.Name)
			f.SetCellValue(sheetName, "C"+fmt.Sprint(startingRow+index), material.Unit)
			f.SetCellValue(sheetName, "D"+fmt.Sprint(startingRow+index), oneEntry.Amount)

		}

		f.SaveAs("./pkg/excels/writeoff/" + invoiceWriteOff.DeliveryCode + ".xlsx")
		if err := f.Close(); err != nil {
			fmt.Println(err)
		}

	default:
		return dto.InvoiceWriteOff{}, fmt.Errorf("unknown Write Off Type")
	}

	return data, nil
}

func (service *invoiceWriteOffService) Delete(id uint) error {
	return service.invoiceWriteOffRepo.Delete(id)
}

func (service *invoiceWriteOffService) Count() (int64, error) {
	return service.invoiceWriteOffRepo.Count()
}
