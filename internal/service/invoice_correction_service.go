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
)

type invoiceCorrectionService struct {
	invoiceCorrectionRepo repository.IInvoiceCorrectionRepository
	invoiceObjectRepo     repository.IInvoiceObjectRepository
	invoiceMaterialsRepo  repository.IInvoiceMaterialsRepository
	materialLocationRepo  repository.IMaterialLocationRepository
}

func InitInvoiceCorrectionService(
	invoiceCorrection repository.IInvoiceCorrectionRepository,
	invoiceObjectRepo repository.IInvoiceObjectRepository,
	invoiceMaterialsRepo repository.IInvoiceMaterialsRepository,
	materialLocationRepo repository.IMaterialLocationRepository,
) IInvoiceCorrectionService {
	return &invoiceCorrectionService{
		invoiceCorrectionRepo: invoiceCorrection,
		invoiceObjectRepo:     invoiceObjectRepo,
		invoiceMaterialsRepo:  invoiceMaterialsRepo,
		materialLocationRepo:  materialLocationRepo,
	}
}

type IInvoiceCorrectionService interface {
	GetPaginated(page, limit int, projectID uint) ([]dto.InvoiceCorrectionPaginated, error)
	GetAll(projectID uint) ([]dto.InvoiceCorrectionPaginated, error)
	GetTotalAmounInLocationByTeamName(projectID, materialID uint, teamNumber string) (float64, error)
	GetInvoiceMaterialsByInvoiceObjectID(id uint) ([]dto.InvoiceCorrectionMaterialsData, error)
	GetSerialNumberOfMaterialInTeam(projectID uint, materialID uint, teamID uint) ([]string, error)
	Create(data dto.InvoiceCorrectionCreate) (model.InvoiceObject, error)
	UniqueObject(projectID uint) ([]dto.ObjectDataForSelect, error)
	UniqueTeam(projectID uint) ([]dto.DataForSelect[uint], error)
	Report(filter dto.InvoiceCorrectionReportFilter) (string, error)
	Count(projectID uint) (int64, error)
	GetOperationsByInvoiceObjectID(id uint) ([]dto.InvoiceCorrectionOperationsData, error)
}

func (service *invoiceCorrectionService) GetPaginated(page, limit int, projectID uint) ([]dto.InvoiceCorrectionPaginated, error) {
	return service.invoiceCorrectionRepo.GetPaginated(page, limit, projectID)
}

func (service *invoiceCorrectionService) GetAll(projectID uint) ([]dto.InvoiceCorrectionPaginated, error) {
	return service.invoiceObjectRepo.GetForCorrection(projectID)
}

func (service *invoiceCorrectionService) GetTotalAmounInLocationByTeamName(projectID, materialID uint, teamNumber string) (float64, error) {
	return service.materialLocationRepo.GetTotalAmountInTeamsByTeamNumber(projectID, materialID, teamNumber)
}

func (service *invoiceCorrectionService) GetInvoiceMaterialsByInvoiceObjectID(id uint) ([]dto.InvoiceCorrectionMaterialsData, error) {
	data, err := service.invoiceCorrectionRepo.GetInvoiceMaterialsDataByInvoiceObjectID(id)
	if err != nil {
		return []dto.InvoiceCorrectionMaterialsData{}, err
	}

	result := []dto.InvoiceCorrectionMaterialsData{}
	resultIndex := 0
	for index, entry := range data {
		if index == 0 {
			result = append(result, entry)
			continue
		}

		if entry.MaterialID == result[resultIndex].MaterialID {
			result[resultIndex].MaterialAmount += entry.MaterialAmount
		} else {
			result = append(result, entry)
			resultIndex++
		}
	}

	return result, nil
}

func (service *invoiceCorrectionService) GetSerialNumberOfMaterialInTeam(projectID uint, materialID uint, teamID uint) ([]string, error) {
	return service.invoiceCorrectionRepo.GetSerialNumberOfMaterialInTeam(projectID, materialID, teamID)
}

func (service *invoiceCorrectionService) Count(projectID uint) (int64, error) {
	return service.invoiceCorrectionRepo.Count(projectID)
}

func (service *invoiceCorrectionService) Create(data dto.InvoiceCorrectionCreate) (model.InvoiceObject, error) {

	invoiceObject, err := service.invoiceObjectRepo.GetByID(data.Details.InvoiceObjectID)
	if err != nil {
		return model.InvoiceObject{}, err
	}

	invoiceObject.ConfirmedByOperator = true
	invoiceObject.DateOfCorrection = data.Details.DateOfCorrection

	invoiceMaterialForCreate := []model.InvoiceMaterials{}
	for _, invoiceMaterial := range data.Items {
		materialInfoSorted, err := service.materialLocationRepo.GetMaterialAmountSortedByCostM19InLocation(invoiceObject.ProjectID, invoiceMaterial.MaterialID, "team", invoiceObject.TeamID)
		if err != nil {
			return model.InvoiceObject{}, err
		}

		fmt.Println(materialInfoSorted)

		index := 0
		for invoiceMaterial.MaterialAmount > 0 {
			if len(materialInfoSorted) == 0 {
				fmt.Println(len(materialInfoSorted))
				return model.InvoiceObject{}, fmt.Errorf("Ошибка корректировки: количество материала внутри корректировки превышает количество материала у бригады")
			}
			invoiceMaterialCreate := model.InvoiceMaterials{
				ProjectID:      invoiceObject.ProjectID,
				ID:             0,
				MaterialCostID: materialInfoSorted[index].MaterialCostID,
				InvoiceID:      invoiceObject.ID,
				InvoiceType:    "object-correction",
				IsDefected:     false,
				Amount:         0,
				Notes:          invoiceMaterial.Notes,
			}

			if materialInfoSorted[index].MaterialAmount <= invoiceMaterial.MaterialAmount {
				invoiceMaterialCreate.Amount = materialInfoSorted[index].MaterialAmount
				invoiceMaterial.MaterialAmount -= materialInfoSorted[index].MaterialAmount
			} else {
				invoiceMaterialCreate.Amount = invoiceMaterial.MaterialAmount
				invoiceMaterial.MaterialAmount = 0
			}

			invoiceMaterialForCreate = append(invoiceMaterialForCreate, invoiceMaterialCreate)
			index++
		}
	}

	invoiceOperationsForCreate := []model.InvoiceOperations{}
	for _, invoiceOperation := range data.Operations {
		invoiceOperationsForCreate = append(invoiceOperationsForCreate, model.InvoiceOperations{
			ID:          0,
			ProjectID:   invoiceObject.ProjectID,
			OperationID: invoiceOperation.OperationID,
			InvoiceID:   invoiceObject.ID,
			InvoiceType: "object-correction",
			Amount:      invoiceOperation.Amount,
			Notes:       "",
		})
	}

	toBeUpdatedTeamLocations := []model.MaterialLocation{}
	toBeUpdatedObjectLocations := []model.MaterialLocation{}
	for _, invoiceMaterial := range invoiceMaterialForCreate {
		materialInTeamLocation, err := service.materialLocationRepo.GetByMaterialCostIDOrCreate(
			invoiceObject.ProjectID,
			invoiceMaterial.MaterialCostID,
			"team",
			invoiceObject.TeamID,
		)
		if err != nil {
			return model.InvoiceObject{}, err
		}

		materialInObjectLocation, err := service.materialLocationRepo.GetByMaterialCostIDOrCreate(
			invoiceObject.ProjectID,
			invoiceMaterial.MaterialCostID,
			"object",
			invoiceObject.ObjectID,
		)
		if err != nil {
			return model.InvoiceObject{}, err
		}

		materialInTeamLocation.Amount -= invoiceMaterial.Amount
		materialInObjectLocation.Amount += invoiceMaterial.Amount

		toBeUpdatedTeamLocations = append(toBeUpdatedTeamLocations, materialInTeamLocation)
		toBeUpdatedObjectLocations = append(toBeUpdatedObjectLocations, materialInObjectLocation)
	}

	result, err := service.invoiceCorrectionRepo.Create(dto.InvoiceCorrectionCreateQuery{
		Details:           invoiceObject,
		InvoiceMaterials:  invoiceMaterialForCreate,
		InvoiceOperations: invoiceOperationsForCreate,
		TeamLocation:      toBeUpdatedTeamLocations,
		ObjectLocation:    toBeUpdatedObjectLocations,
		OperatorDetails: model.InvoiceObjectOperator{
			OperatorWorkerID: data.Details.OperatorWorkerID,
			InvoiceObjectID:  invoiceObject.ID,
		},
	})

	return result, nil
}

func (service *invoiceCorrectionService) UniqueObject(projectID uint) ([]dto.ObjectDataForSelect, error) {
	return service.invoiceCorrectionRepo.UniqueObject(projectID)
}

func (service *invoiceCorrectionService) UniqueTeam(projectID uint) ([]dto.DataForSelect[uint], error) {
	return service.invoiceCorrectionRepo.UniqueTeam(projectID)
}

func (service *invoiceCorrectionService) Report(filter dto.InvoiceCorrectionReportFilter) (string, error) {
	invoices, err := service.invoiceCorrectionRepo.ReportFilterData(filter)
	if err != nil {
		return "", err
	}

	fmt.Println(invoices)

	templateFilePath := filepath.Join("./pkg/excels/templates/", "Object Spenditure Report.xlsx")
	f, err := excelize.OpenFile(templateFilePath)
	if err != nil {
		return "", err
	}
	sheetName := "Sheet1"

	rowCount := 2
	for _, invoice := range invoices {
		invoiceMaterials, err := service.invoiceMaterialsRepo.GetDataForReport(invoice.ID, "object-correction")
		if err != nil {
			return "", err
		}

		for index, invoiceMaterial := range invoiceMaterials {
			f.SetCellStr(sheetName, "A"+fmt.Sprint(rowCount+index), invoice.DeliveryCode)
			f.SetCellStr(sheetName, "B"+fmt.Sprint(rowCount+index), invoice.ObjectName)
			f.SetCellStr(sheetName, "C"+fmt.Sprint(rowCount+index), utils.ObjectTypeConverter(invoice.ObjectType))
			f.SetCellStr(sheetName, "D"+fmt.Sprint(rowCount+index), invoice.TeamNumber)
			f.SetCellStr(sheetName, "E"+fmt.Sprint(rowCount+index), invoice.TeamLeaderName)
			dateOfInvoice := invoice.DateOfInvoice.String()
			dateOfInvoice = dateOfInvoice[:len(dateOfInvoice)-10]
			f.SetCellStr(sheetName, "F"+fmt.Sprint(rowCount+index), dateOfInvoice)
			f.SetCellStr(sheetName, "G"+fmt.Sprint(rowCount+index), invoice.OperatorName)
			dateOfCorrection := invoice.DateOfInvoice.String()
			dateOfCorrection = dateOfCorrection[:len(dateOfCorrection)-10]
			f.SetCellStr(sheetName, "H"+fmt.Sprint(rowCount+index), dateOfCorrection)

			f.SetCellStr(sheetName, "I"+fmt.Sprint(rowCount+index), invoiceMaterial.MaterialName)
			f.SetCellStr(sheetName, "J"+fmt.Sprint(rowCount+index), invoiceMaterial.MaterialUnit)
			f.SetCellFloat(sheetName, "K"+fmt.Sprint(rowCount+index), invoiceMaterial.InvoiceMaterialAmount, 2, 64)
			f.SetCellStr(sheetName, "L"+fmt.Sprint(rowCount+index), invoiceMaterial.InvoiceMaterialNotes)
		}

		rowCount += len(invoiceMaterials)
	}

	currentTime := time.Now()
	fileName := fmt.Sprintf(
		"Отсчет Расхода - %s.xlsx",
		currentTime.Format("02-01-2006"),
	)

	tempFilePath := filepath.Join("./pkg/excels/temp/", fileName)

	f.SaveAs(tempFilePath)

	if err := f.Close(); err != nil {
		fmt.Println(err)
	}

	return fileName, nil
}

func (service *invoiceCorrectionService) GetOperationsByInvoiceObjectID(id uint) ([]dto.InvoiceCorrectionOperationsData, error) {
	return service.invoiceCorrectionRepo.GetOperationsByInvoiceObjectID(id)
}
