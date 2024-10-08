package service

import (
	"backend-v2/internal/dto"
	"backend-v2/internal/repository"
	"backend-v2/model"
	"backend-v2/pkg/utils"
	"fmt"
	"path/filepath"
	"time"

	"github.com/xuri/excelize/v2"
	// "github.com/xuri/excelize/v2"
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
	invoiceCountRepo     repository.IInvoiceCountRepository
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
	invoiceCountRepo repository.IInvoiceCountRepository,
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
		invoiceCountRepo:     invoiceCountRepo,
	}
}

type IInvoiceWriteOffService interface {
	GetAll() ([]model.InvoiceWriteOff, error)
	GetPaginated(page, limit int, data dto.InvoiceWriteOffSearchParameters) ([]dto.InvoiceWriteOffPaginated, error)
	GetInvoiceMaterialsWithoutSerialNumbers(id uint) ([]dto.InvoiceMaterialsWithoutSerialNumberView, error)
	GetByID(id uint) (model.InvoiceWriteOff, error)
	Create(data dto.InvoiceWriteOff) (model.InvoiceWriteOff, error)
	Update(data dto.InvoiceWriteOff) (model.InvoiceWriteOff, error)
	Delete(id uint) error
	Count(filter dto.InvoiceWriteOffSearchParameters) (int64, error)
	GetMaterialsForEdit(id uint, locationType string, locationID uint) ([]dto.InvoiceWriteOffMaterialsForEdit, error)
	Confirmation(id, projectID uint) error
	Report(parameters dto.InvoiceWriteOffReportParameters) (string, error)
	GetMaterialsInLocation(projectID, locationID uint, locationType string) ([]dto.InvoiceReturnMaterialForSelect, error)
}

func (service *invoiceWriteOffService) GetAll() ([]model.InvoiceWriteOff, error) {
	return service.invoiceWriteOffRepo.GetAll()
}

func (service *invoiceWriteOffService) GetByID(id uint) (model.InvoiceWriteOff, error) {
	return service.invoiceWriteOffRepo.GetByID(id)
}

func (service *invoiceWriteOffService) GetPaginated(page, limit int, data dto.InvoiceWriteOffSearchParameters) ([]dto.InvoiceWriteOffPaginated, error) {
	invoiceWriteOffs, err := service.invoiceWriteOffRepo.GetPaginated(page, limit, data)
	if err != nil {
		return []dto.InvoiceWriteOffPaginated{}, err
	}

	for index, invoiceWriteOff := range invoiceWriteOffs {
		switch invoiceWriteOff.WriteOffType {
		case "writeoff-warehouse":
			break
		case "loss-warehouse":
			break
		case "loss-team":
			team, err := service.teamRepo.GetTeamNumberAndTeamLeadersByID(data.ProjectID, invoiceWriteOff.WriteOffLocationID)
			if err != nil {
				return []dto.InvoiceWriteOffPaginated{}, err
			}

			invoiceWriteOffs[index].WriteOffLocationName = team[0].TeamNumber + " (" + team[0].TeamLeaderName + ")"
			break

		case "writeoff-object":
			object, err := service.objectRepo.GetByID(invoiceWriteOff.WriteOffLocationID)
			if err != nil {
				return []dto.InvoiceWriteOffPaginated{}, err
			}

			invoiceWriteOffs[index].WriteOffLocationName = object.Name
			break

		case "loss-object":
			object, err := service.objectRepo.GetByID(invoiceWriteOff.WriteOffLocationID)
			if err != nil {
				return []dto.InvoiceWriteOffPaginated{}, err
			}

			invoiceWriteOffs[index].WriteOffLocationName = object.Name
			break
		default:
			return []dto.InvoiceWriteOffPaginated{}, fmt.Errorf("Обноружен неправильный тип списание %v", invoiceWriteOff.WriteOffType)
		}

	}

	return invoiceWriteOffs, nil
}

func (service *invoiceWriteOffService) Create(data dto.InvoiceWriteOff) (model.InvoiceWriteOff, error) {

	count, err := service.invoiceCountRepo.CountInvoice("writeoff", data.Details.ProjectID)
	if err != nil {
		return model.InvoiceWriteOff{}, err
	}

	data.Details.DeliveryCode = utils.UniqueCodeGeneration("С", int64(count+1), data.Details.ProjectID)

	invoiceMaterialForCreate := []model.InvoiceMaterials{}

	writeOffLocation := ""
	switch data.Details.WriteOffType {
	case "loss-warehouse":
		writeOffLocation = "warehouse"
		break
	case "writeoff-warehouse":
		writeOffLocation = "warehouse"
		break
	case "loss-team":
		writeOffLocation = "team"
		break
	case "loss-object":
		writeOffLocation = "object"
		break
	case "writeoff-object":
		writeOffLocation = "object"
		break
	default:
		return model.InvoiceWriteOff{}, fmt.Errorf("Неправильный вид списание обнаружен")
	}
	for _, invoiceMaterial := range data.Items {
		if len(invoiceMaterial.SerialNumbers) == 0 {
			materialInfoSorted, err := service.materialLocationRepo.GetMaterialAmountSortedByCostM19InLocation(data.Details.ProjectID, invoiceMaterial.MaterialID, writeOffLocation, data.Details.WriteOffLocationID)
			if err != nil {
				return model.InvoiceWriteOff{}, err
			}

			index := 0
			for invoiceMaterial.Amount > 0 {
				invoiceMaterialCreate := model.InvoiceMaterials{
					ProjectID:      data.Details.ProjectID,
					ID:             0,
					MaterialCostID: materialInfoSorted[index].MaterialCostID,
					InvoiceID:      0,
					InvoiceType:    "writeoff",
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
	}

	invoiceWriteOff, err := service.invoiceWriteOffRepo.Create(dto.InvoiceWriteOffMutationData{
		InvoiceWriteOff:  data.Details,
		InvoiceMaterials: invoiceMaterialForCreate,
	})
	if err != nil {
		return model.InvoiceWriteOff{}, err
	}

	return invoiceWriteOff, nil
}

func (service *invoiceWriteOffService) Update(data dto.InvoiceWriteOff) (model.InvoiceWriteOff, error) {
	writeOffLocation := ""
	switch data.Details.WriteOffType {
	case "loss-warehouse":
		writeOffLocation = "warehouse"
		break
	case "writeoff-warehouse":
		writeOffLocation = "warehouse"
		break
	case "loss-team":
		writeOffLocation = "team"
		break
	case "loss-object":
		writeOffLocation = "object"
		break
	case "writeoff-object":
		writeOffLocation = "object"
		break
	default:
		return model.InvoiceWriteOff{}, fmt.Errorf("Неправильный вид списание обнаружен")
	}
	invoiceMaterialForCreate := []model.InvoiceMaterials{}
	for _, invoiceMaterial := range data.Items {
		if len(invoiceMaterial.SerialNumbers) == 0 {
			materialInfoSorted, err := service.materialLocationRepo.GetMaterialAmountSortedByCostM19InLocation(data.Details.ProjectID, invoiceMaterial.MaterialID, writeOffLocation, data.Details.WriteOffLocationID)
			if err != nil {
				return model.InvoiceWriteOff{}, err
			}

			index := 0
			for invoiceMaterial.Amount > 0 {
				invoiceMaterialCreate := model.InvoiceMaterials{
					ProjectID:      data.Details.ProjectID,
					ID:             0,
					MaterialCostID: materialInfoSorted[index].MaterialCostID,
					InvoiceID:      0,
					InvoiceType:    "writeoff",
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
	}

	invoiceWriteOff, err := service.invoiceWriteOffRepo.Update(dto.InvoiceWriteOffMutationData{
		InvoiceWriteOff:  data.Details,
		InvoiceMaterials: invoiceMaterialForCreate,
	})
	if err != nil {
		return model.InvoiceWriteOff{}, err
	}

	return invoiceWriteOff, nil
}

func (service *invoiceWriteOffService) Delete(id uint) error {
	return service.invoiceWriteOffRepo.Delete(id)
}

func (service *invoiceWriteOffService) Count(filter dto.InvoiceWriteOffSearchParameters) (int64, error) {
	return service.invoiceWriteOffRepo.Count(filter)
}

func (service *invoiceWriteOffService) GetInvoiceMaterialsWithoutSerialNumbers(id uint) ([]dto.InvoiceMaterialsWithoutSerialNumberView, error) {
	return service.invoiceMaterialsRepo.GetInvoiceMaterialsWithoutSerialNumbers(id, "writeoff")
}

func (service *invoiceWriteOffService) GetMaterialsForEdit(id uint, locationType string, locationID uint) ([]dto.InvoiceWriteOffMaterialsForEdit, error) {

	data, err := service.invoiceWriteOffRepo.GetMaterialsForEdit(id, locationType, locationID)
	if err != nil {
		return []dto.InvoiceWriteOffMaterialsForEdit{}, nil
	}

	var result []dto.InvoiceWriteOffMaterialsForEdit
	for index, entry := range data {
		if index == 0 {
			result = append(result, entry)
			continue
		}

		lastItemIndex := len(result) - 1
		if result[lastItemIndex].MaterialID == entry.MaterialID {
			result[lastItemIndex].Amount += entry.Amount
			result[lastItemIndex].LocationAmount += entry.LocationAmount
		} else {
			result = append(result, entry)
		}
	}

	return result, nil
}

func (service *invoiceWriteOffService) Confirmation(id, projectID uint) error {
	invoiceWriteOff, err := service.invoiceWriteOffRepo.GetByID(id)
	if err != nil {
		return err
	}
	invoiceWriteOff.Confirmation = true

	invoiceMaterials, err := service.invoiceMaterialsRepo.GetByInvoice(invoiceWriteOff.ProjectID, invoiceWriteOff.ID, "writeoff")
	if err != nil {
		return err
	}

	materialsInTheLocation := []model.MaterialLocation{}

	switch invoiceWriteOff.WriteOffType {
	case "writeoff-warehouse":
		materialsInTheLocation, err = service.materialLocationRepo.GetMaterialsInLocationBasedOnInvoiceID(0, "warehouse", id, "writeoff")
		if err != nil {
			return err
		}

		break
	case "loss-warehouse":
		materialsInTheLocation, err = service.materialLocationRepo.GetMaterialsInLocationBasedOnInvoiceID(0, "warehouse", id, "writeoff")
		if err != nil {
			return err
		}

		break
	case "loss-team":
		materialsInTheLocation, err = service.materialLocationRepo.GetMaterialsInLocationBasedOnInvoiceID(0, "team", id, "writeoff")
		if err != nil {
			return err
		}

		break
	case "loss-object":
		materialsInTheLocation, err = service.materialLocationRepo.GetMaterialsInLocationBasedOnInvoiceID(0, "object", id, "writeoff")
		if err != nil {
			return err
		}

		break
	}

	materialsInWriteOffLocation, err := service.materialLocationRepo.GetByLocationType(invoiceWriteOff.WriteOffType)
	if err != nil {
		return err
	}

	for _, invoiceMaterial := range invoiceMaterials {
		indexOfExistingMaterialInLocation := -1
		for index, materialInTheLocation := range materialsInTheLocation {
			if materialInTheLocation.MaterialCostID == invoiceMaterial.MaterialCostID {
				indexOfExistingMaterialInLocation = index
				break
			}
		}

		if indexOfExistingMaterialInLocation == -1 {
			return fmt.Errorf("Ошибка, несанкционированный материал")
		}
		materialsInTheLocation[indexOfExistingMaterialInLocation].Amount -= invoiceMaterial.Amount

		indexOfExistingMaterialInWriteOffLocation := -1
		for index, materialInWriteOffLocation := range materialsInWriteOffLocation {
			if materialInWriteOffLocation.MaterialCostID == invoiceMaterial.MaterialCostID {
				indexOfExistingMaterialInWriteOffLocation = index
				break
			}
		}

		if indexOfExistingMaterialInWriteOffLocation != -1 {
			materialsInWriteOffLocation[indexOfExistingMaterialInWriteOffLocation].Amount += invoiceMaterial.Amount
		} else {
			materialsInWriteOffLocation = append(materialsInWriteOffLocation, model.MaterialLocation{
				ID:             0,
				MaterialCostID: invoiceMaterial.MaterialCostID,
				ProjectID:      invoiceWriteOff.ProjectID,
				LocationID:     0,
				LocationType:   invoiceWriteOff.WriteOffType,
				Amount:         invoiceMaterial.Amount,
			})
		}

		fmt.Println(materialsInTheLocation)
		fmt.Println(materialsInWriteOffLocation)
	}

	err = service.invoiceWriteOffRepo.Confirmation(dto.InvoiceWriteOffConfirmationData{
		InvoiceWriteOff:     invoiceWriteOff,
		MaterialsInLocation: materialsInTheLocation,
		MaterialsInWriteOff: materialsInWriteOffLocation,
	})

	return err
}

func (service *invoiceWriteOffService) Report(parameters dto.InvoiceWriteOffReportParameters) (string, error) {
	invoices, err := service.invoiceWriteOffRepo.ReportFilterData(parameters)
	if err != nil {
		return "", err
	}

	templateFilePath := filepath.Join("./pkg/excels/templates/", "Invoice Writeoff Report.xlsx")
	f, err := excelize.OpenFile(templateFilePath)
	if err != nil {
		return "", err
	}

	sheetName := "Sheet1"

	rowCount := 2
	for _, invoice := range invoices {
		invoiceMaterials, err := service.invoiceMaterialsRepo.GetDataForReport(invoice.ID, "writeoff")
		if err != nil {
			return "", err
		}

		for _, invoiceMaterial := range invoiceMaterials {
			f.SetCellStr(sheetName, "A"+fmt.Sprint(rowCount), invoice.DeliveryCode)
			f.SetCellStr(sheetName, "B"+fmt.Sprint(rowCount), invoice.ReleasedWorkerName)

			switch parameters.WriteOffType {
			case "writeoff-warehouse":
				f.SetCellStr(sheetName, "C"+fmt.Sprint(rowCount), "Склад")
				break

			case "loss-warehouse":
				f.SetCellStr(sheetName, "C"+fmt.Sprint(rowCount), "Склад")
				break

			case "loss-team":
				team, err := service.teamRepo.GetByID(parameters.WriteOffLocationID)
				if err != nil {
					return "", err
				}

				f.SetCellStr(sheetName, "C"+fmt.Sprint(rowCount), team.Number)
				break

			case "loss-object":
				object, err := service.objectRepo.GetByID(parameters.WriteOffLocationID)
				if err != nil {
					return "", err
				}

				f.SetCellStr(sheetName, "C"+fmt.Sprint(rowCount), object.Name)
				break

			case "writeoff-object":
				object, err := service.objectRepo.GetByID(parameters.WriteOffLocationID)
				if err != nil {
					return "", err
				}

				f.SetCellStr(sheetName, "C"+fmt.Sprint(rowCount), object.Name)
				break
			}

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
		"Отсчет накладной списание - %s.xlsx",
		currentTime.Format("02-01-2006"),
	)

	tempFilePath := filepath.Join("./pkg/excels/temp/", fileName)

	f.SaveAs(tempFilePath)

	if err := f.Close(); err != nil {
		fmt.Println(err)
	}

	return fileName, nil
}

func (service *invoiceWriteOffService) GetMaterialsInLocation(projectID, locationID uint, locationType string) ([]dto.InvoiceReturnMaterialForSelect, error) {

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
