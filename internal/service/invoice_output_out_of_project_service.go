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

type invoiceOutputOutOfProjectService struct {
	invoiceOutputOutOfProjectRepo repository.IInvoiceOutputOutOfProjectRepository
	invoiceOutputRepo             repository.IInvoiceOutputRepository
	materialLocationRepo          repository.IMaterialLocationRepository
	invoiceCountRepo              repository.IInvoiceCountRepository
	invoiceMaterialsRepo          repository.IInvoiceMaterialsRepository
	materialsRepo                 repository.IMaterialRepository
	workerRepo                    repository.IWorkerRepository
	materialCostRepo              repository.IMaterialCostRepository
}

func InitInvoiceOutputOutOfProjectService(
	invoiceOutputOutOfProjectRepo repository.IInvoiceOutputOutOfProjectRepository,
	invoiceOutputRepo repository.IInvoiceOutputRepository,
	materialLocationRepo repository.IMaterialLocationRepository,
	invoiceCountRepo repository.IInvoiceCountRepository,
	invoiceMaterialsRepo repository.IInvoiceMaterialsRepository,
	materialsRepo repository.IMaterialRepository,
	workerRepo repository.IWorkerRepository,
	materialCostRepo repository.IMaterialCostRepository,
) IInvoiceOutputOutOfProjectService {
	return &invoiceOutputOutOfProjectService{
		invoiceOutputOutOfProjectRepo: invoiceOutputOutOfProjectRepo,
		invoiceOutputRepo:             invoiceOutputRepo,
		materialLocationRepo:          materialLocationRepo,
		invoiceCountRepo:              invoiceCountRepo,
		invoiceMaterialsRepo:          invoiceMaterialsRepo,
		materialsRepo:                 materialsRepo,
		workerRepo:                    workerRepo,
		materialCostRepo:              materialCostRepo,
	}
}

type IInvoiceOutputOutOfProjectService interface {
	GetPaginated(page, limit int, filter dto.InvoiceOutputOutOfProjectSearchParameters) ([]dto.InvoiceOutputOutOfProjectPaginated, error)
	GetByID(id uint) (model.InvoiceOutputOutOfProject, error)
	Count(data dto.InvoiceOutputOutOfProjectSearchParameters) (int64, error)
	Create(data dto.InvoiceOutputOutOfProject) (model.InvoiceOutputOutOfProject, error)
	Delete(id uint) error
	GetInvoiceMaterialsWithoutSerialNumbers(id uint) ([]dto.InvoiceMaterialsWithoutSerialNumberView, error)
	GetInvoiceMaterialsWithSerialNumbers(id uint) ([]dto.InvoiceMaterialsWithSerialNumberView, error)
	Update(data dto.InvoiceOutputOutOfProject) (model.InvoiceOutputOutOfProject, error)
	Confirmation(id uint) error
	GetMaterialsForEdit(id uint) ([]dto.InvoiceOutputMaterialsForEdit, error)
	GetUniqueNameOfProjects(projectID uint) ([]string, error)
	Report(filter dto.InvoiceOutputOutOfProjectReportFilter) (string, error)
	GetDocument(deliveryCode string) (string, error)
}

func (service *invoiceOutputOutOfProjectService) GetPaginated(page, limit int, filter dto.InvoiceOutputOutOfProjectSearchParameters) ([]dto.InvoiceOutputOutOfProjectPaginated, error) {
	return service.invoiceOutputOutOfProjectRepo.GetPaginated(page, limit, filter)
}

func (service *invoiceOutputOutOfProjectService) Count(filter dto.InvoiceOutputOutOfProjectSearchParameters) (int64, error) {
	return service.invoiceOutputOutOfProjectRepo.Count(filter)
}

func (service *invoiceOutputOutOfProjectService) Create(data dto.InvoiceOutputOutOfProject) (model.InvoiceOutputOutOfProject, error) {

	count, err := service.invoiceCountRepo.CountInvoice("output", data.Details.ProjectID)
	if err != nil {
		return model.InvoiceOutputOutOfProject{}, err
	}

	data.Details.DeliveryCode = utils.UniqueCodeGeneration("О", int64(count+1), data.Details.ProjectID)

	invoiceMaterialForCreate := []model.InvoiceMaterials{}
	for _, invoiceMaterial := range data.Items {
		if len(invoiceMaterial.SerialNumbers) == 0 {
			materialInfoSorted, err := service.materialLocationRepo.GetMaterialAmountSortedByCostM19InLocation(data.Details.ProjectID, invoiceMaterial.MaterialID, "warehouse", 0)
			if err != nil {
				return model.InvoiceOutputOutOfProject{}, err
			}

			index := 0
			for invoiceMaterial.Amount > 0 {
				invoiceMaterialCreate := model.InvoiceMaterials{
					ProjectID:      data.Details.ProjectID,
					ID:             0,
					MaterialCostID: materialInfoSorted[index].MaterialCostID,
					InvoiceID:      0,
					InvoiceType:    "output-out-of-project",
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
		}
	}

	if err := service.GenerateExcelFile(data.Details, invoiceMaterialForCreate); err != nil {
		return model.InvoiceOutputOutOfProject{}, err
	}

	invoiceOutput, err := service.invoiceOutputOutOfProjectRepo.Create(dto.InvoiceOutputOutOfProjectCreateQueryData{
		Invoice:          data.Details,
		InvoiceMaterials: invoiceMaterialForCreate,
	})
	if err != nil {
		return model.InvoiceOutputOutOfProject{}, err
	}

	return invoiceOutput, nil
}

func (service *invoiceOutputOutOfProjectService) Delete(id uint) error {
	return service.invoiceOutputOutOfProjectRepo.Delete(id)
}

func (service *invoiceOutputOutOfProjectService) GetInvoiceMaterialsWithoutSerialNumbers(id uint) ([]dto.InvoiceMaterialsWithoutSerialNumberView, error) {
	return service.invoiceMaterialsRepo.GetInvoiceMaterialsWithoutSerialNumbers(id, "output-out-of-project")
}

func (service *invoiceOutputOutOfProjectService) GetInvoiceMaterialsWithSerialNumbers(id uint) ([]dto.InvoiceMaterialsWithSerialNumberView, error) {
	queryData, err := service.invoiceMaterialsRepo.GetInvoiceMaterialsWithSerialNumbers(id, "output")
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

func (service *invoiceOutputOutOfProjectService) Update(data dto.InvoiceOutputOutOfProject) (model.InvoiceOutputOutOfProject, error) {
	invoiceMaterialForCreate := []model.InvoiceMaterials{}
	for _, invoiceMaterial := range data.Items {
		if len(invoiceMaterial.SerialNumbers) == 0 {
			materialInfoSorted, err := service.materialLocationRepo.GetMaterialAmountSortedByCostM19InLocation(data.Details.ProjectID, invoiceMaterial.MaterialID, "warehouse", 0)
			if err != nil {
				return model.InvoiceOutputOutOfProject{}, err
			}

			index := 0
			for invoiceMaterial.Amount > 0 {
				invoiceMaterialCreate := model.InvoiceMaterials{
					ProjectID:      data.Details.ProjectID,
					ID:             0,
					MaterialCostID: materialInfoSorted[index].MaterialCostID,
					InvoiceID:      0,
					InvoiceType:    "output-out-of-project",
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
		}
	}

	excelFilePath := filepath.Join("./pkg/excels/output/", data.Details.DeliveryCode+".xlsx")
	if err := os.Remove(excelFilePath); err != nil {
		return model.InvoiceOutputOutOfProject{}, err
	}

	if err := service.GenerateExcelFile(data.Details, invoiceMaterialForCreate); err != nil {
		return model.InvoiceOutputOutOfProject{}, err
	}

	invoiceOutput, err := service.invoiceOutputOutOfProjectRepo.Update(dto.InvoiceOutputOutOfProjectCreateQueryData{
		Invoice:          data.Details,
		InvoiceMaterials: invoiceMaterialForCreate,
	})
	if err != nil {
		return model.InvoiceOutputOutOfProject{}, err
	}

	return invoiceOutput, nil
}

func (service *invoiceOutputOutOfProjectService) GetByID(id uint) (model.InvoiceOutputOutOfProject, error) {
	return service.invoiceOutputOutOfProjectRepo.GetByID(id)
}

func (service *invoiceOutputOutOfProjectService) Confirmation(id uint) error {
	invoiceOutputOutOfProject, err := service.invoiceOutputOutOfProjectRepo.GetByID(id)
	if err != nil {
		return err
	}
	invoiceOutputOutOfProject.Confirmation = true

	invoiceMaterials, err := service.invoiceMaterialsRepo.GetByInvoice(invoiceOutputOutOfProject.ProjectID, invoiceOutputOutOfProject.ID, "output-out-of-project")
	if err != nil {
		return err
	}

	materialsInWarehouse, err := service.materialLocationRepo.GetMaterialsInLocationBasedOnInvoiceID(0, "warehouse", id, "output-out-of-project")
	if err != nil {
		return err
	}

	materialsOutOfProject, err := service.materialLocationRepo.GetMaterialsInLocationBasedOnInvoiceID(0, "out-of-project", id, "output-out-of-project")
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

		if materialsInWarehouse[materialInWarehouseIndex].Amount < invoiceMaterial.Amount {
			material, err := service.materialsRepo.GetByMaterialCostID(invoiceMaterial.MaterialCostID)
			if err != nil {
				return fmt.Errorf("Недостаточно материалов в складе и данные про материал не были получены: %v", err)
			}

			return fmt.Errorf("Недостаточно материала %v на складе: в накладной указано - %v а на складе - %v", material.Name, invoiceMaterial.Amount, materialsInWarehouse[materialInWarehouseIndex].Amount)
		}
		materialsInWarehouse[materialInWarehouseIndex].Amount -= invoiceMaterial.Amount

		materialOutOfProjectIndex := -1
		for index, materialOutOfProject := range materialsOutOfProject {
			if materialOutOfProject.MaterialCostID == invoiceMaterial.MaterialCostID {
				materialInWarehouseIndex = index
				break
			}
		}

		if materialOutOfProjectIndex != -1 {
		} else {
			materialsOutOfProject = append(materialsOutOfProject, model.MaterialLocation{
				ID:             0,
				ProjectID:      invoiceOutputOutOfProject.ProjectID,
				MaterialCostID: invoiceMaterial.MaterialCostID,
				LocationID:     0,
				LocationType:   "out-of-project",
				Amount:         invoiceMaterial.Amount,
			})
		}
	}

	err = service.invoiceOutputOutOfProjectRepo.Confirmation(dto.InvoiceOutputOutOfProjectConfirmationQueryData{
		InvoiceData:           invoiceOutputOutOfProject,
		WarehouseMaterials:    materialsInWarehouse,
		OutOfProjectMaterials: materialsOutOfProject,
	})

	return nil
}

func (service *invoiceOutputOutOfProjectService) GetMaterialsForEdit(id uint) ([]dto.InvoiceOutputMaterialsForEdit, error) {
	data, err := service.invoiceOutputOutOfProjectRepo.GetMaterialsForEdit(id)
	if err != nil {
		return []dto.InvoiceOutputMaterialsForEdit{}, nil
	}

	var result []dto.InvoiceOutputMaterialsForEdit
	for index, entry := range data {
		if index == 0 {
			result = append(result, entry)
			continue
		}

		lastItemIndex := len(result) - 1
		if result[lastItemIndex].MaterialID == entry.MaterialID {
			result[lastItemIndex].Amount += entry.Amount
			result[lastItemIndex].WarehouseAmount += entry.WarehouseAmount
		} else {
			result = append(result, entry)
		}
	}

	return result, nil
}

func (service *invoiceOutputOutOfProjectService) GetUniqueNameOfProjects(projectID uint) ([]string, error) {
	return service.invoiceOutputOutOfProjectRepo.GetUniqueNameOfProjects(projectID)
}

func (service *invoiceOutputOutOfProjectService) Report(filter dto.InvoiceOutputOutOfProjectReportFilter) (string, error) {
	invoices, err := service.invoiceOutputOutOfProjectRepo.ReportFilterData(filter)
	if err != nil {
		return "", err
	}

	templateFilePath := filepath.Join("./pkg/excels/templates/", "Invoice Output Out Of Project.xlsx")
	f, err := excelize.OpenFile(templateFilePath)
	if err != nil {
		return "", err
	}
	sheetName := "Sheet1"

	rowCount := 2
	for _, invoice := range invoices {
		invoiceMaterials, err := service.invoiceMaterialsRepo.GetDataForReport(invoice.ID, "output-out-of-project")
		if err != nil {
			return "", err
		}

		for _, invoiceMaterial := range invoiceMaterials {
			f.SetCellStr(sheetName, "A"+fmt.Sprint(rowCount), invoice.DeliveryCode)
			f.SetCellStr(sheetName, "B"+fmt.Sprint(rowCount), invoice.ReleasedWorkerName)
			f.SetCellStr(sheetName, "C"+fmt.Sprint(rowCount), invoice.NameOfProject)

			dateOfInvoice := invoice.DateOfInvoice.String()
			dateOfInvoice = dateOfInvoice[:len(dateOfInvoice)-10]
			f.SetCellStr(sheetName, "D"+fmt.Sprint(rowCount), dateOfInvoice)

			f.SetCellValue(sheetName, "E"+fmt.Sprint(rowCount), invoiceMaterial.MaterialName)
			f.SetCellValue(sheetName, "F"+fmt.Sprint(rowCount), invoiceMaterial.MaterialUnit)
			f.SetCellFloat(sheetName, "G"+fmt.Sprint(rowCount), invoiceMaterial.InvoiceMaterialAmount, 2, 64)

			costM19, _ := invoiceMaterial.MaterialCostM19.Float64()
			f.SetCellFloat(sheetName, "H"+fmt.Sprint(rowCount), costM19, 2, 64)
			f.SetCellValue(sheetName, "I"+fmt.Sprint(rowCount), invoiceMaterial.InvoiceMaterialNotes)
			rowCount++
		}
	}

	currentTime := time.Now()
	fileName := fmt.Sprintf(
		"Отсчет накладной отпуск вне проекта - %s.xlsx",
		currentTime.Format("02-01-2006"),
	)

	tempFilePath := filepath.Join("./pkg/excels/temp/", fileName)

	f.SaveAs(tempFilePath)

	if err := f.Close(); err != nil {
		fmt.Println(err)
	}

	return fileName, nil
}

func (service *invoiceOutputOutOfProjectService) GenerateExcelFile(details model.InvoiceOutputOutOfProject, items []model.InvoiceMaterials) error {

	templateFilePath := filepath.Join("./pkg/excels/templates/output out of project.xlsx")
	f, err := excelize.OpenFile(templateFilePath)
	if err != nil {
		return err
	}

	sheetName := "Отпуск"
	startingRow := 5
	f.InsertRows(sheetName, startingRow, len(items))

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

	materialNamingStyle, _ := f.NewStyle(&excelize.Style{
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

	workerNamingStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Size:      8,
			VertAlign: "center",
		},
		Alignment: &excelize.Alignment{
			Horizontal: "left",
			WrapText:   true,
			Vertical:   "center",
		},
	})

	f.SetCellValue(sheetName, "C1", fmt.Sprintf(`НАКЛАДНАЯ 
№ %s
от %s года       
на отпуск материала 
`, details.DeliveryCode, utils.DateConverter(details.DateOfInvoice)))

	for index, oneEntry := range items {
		material, err := service.materialsRepo.GetByMaterialCostID(oneEntry.MaterialCostID)
		if err != nil {
			return err
		}

		materialCost, err := service.materialCostRepo.GetByID(oneEntry.MaterialCostID)
		if err != nil {
			return err
		}
		f.SetCellStyle(sheetName, "A"+fmt.Sprint(startingRow+index), "G"+fmt.Sprint(startingRow+index), defaultStyle)
		f.SetCellStyle(sheetName, "B"+fmt.Sprint(startingRow+index), "B"+fmt.Sprint(startingRow+index), materialNamingStyle)

		f.SetCellInt(sheetName, "A"+fmt.Sprint(startingRow+index), index+1)
		f.SetCellStr(sheetName, "B"+fmt.Sprint(startingRow+index), material.Code)
		f.SetCellStr(sheetName, "C"+fmt.Sprint(startingRow+index), material.Name)
		f.SetCellStr(sheetName, "D"+fmt.Sprint(startingRow+index), material.Unit)
		f.SetCellFloat(sheetName, "E"+fmt.Sprint(startingRow+index), oneEntry.Amount, 3, 64)

		materialCostM19, _ := materialCost.CostM19.Float64()
		f.SetCellFloat(sheetName, "F"+fmt.Sprint(startingRow+index), materialCostM19, 3, 64)
		f.SetCellStr(sheetName, "G"+fmt.Sprint(startingRow+index), oneEntry.Notes)
	}

	released, err := service.workerRepo.GetByID(details.ReleasedWorkerID)
	if err != nil {
		return err
	}
	f.SetCellStyle(sheetName, "C"+fmt.Sprint(6+len(items)), "C"+fmt.Sprint(6+len(items)), workerNamingStyle)
	f.SetCellStr(sheetName, "C"+fmt.Sprint(6+len(items)), released.Name)

	excelFilePath := filepath.Join("./pkg/excels/output/", details.DeliveryCode+".xlsx")
	if err := f.SaveAs(excelFilePath); err != nil {
		return err
	}

	// if err := f.Close(); err != nil {
	// 	fmt.Println(err)
	// 	return err
	// }

	// path, err := os.Getwd()
	// if err != nil {
	// 	return err
	// }

	// fullExcelFilePath := filepath.Join(path, "pkg/excels/output/"+data.Details.DeliveryCode+".xlsx")
	// fullPythonScriptPath := filepath.Join(path, "pkg/scripts/excelToPdf.py")
	//  fmt.Println(fullPythonScriptPath)
	//  fmt.Println(fullExcelFilePath)
	// if runtime.GOOS == "windows" {
	//    cmd := exec.Command("python", fullPythonScriptPath, fullExcelFilePath)
	//    out, err := cmd.Output()
	//    fmt.Println(string(out))
	// 	if err != nil {
	// 		fmt.Println("Error: ", err)
	// 		return err
	// 	}
	// }

	return nil
}

func (service *invoiceOutputOutOfProjectService) GetDocument(deliveryCode string) (string, error) {
	invoiceOutput, err := service.invoiceOutputOutOfProjectRepo.GetByDeliveryCode(deliveryCode)
	if err != nil {
		return "", err
	}

	if invoiceOutput.Confirmation {
		return ".pdf", nil
	} else {
		return ".xlsx", nil
	}
}
