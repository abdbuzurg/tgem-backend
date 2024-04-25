package service

import (
	"backend-v2/internal/dto"
	"backend-v2/internal/repository"
	"backend-v2/model"
	"backend-v2/pkg/utils"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

type materialLocationService struct {
	materialLocationRepo repository.IMaterialLocationRepository
	materialCostRepo     repository.IMaterialCostRepository
	materialRepo         repository.IMaterialRepository
	teamRepo             repository.ITeamRepository
	objectRepo           repository.IObjectRepository
	materialDefectRepo   repository.IMaterialDefectRepository
}

func InitMaterialLocationService(
	materialLocationRepo repository.IMaterialLocationRepository,
	materialCostRepo repository.IMaterialCostRepository,
	materialRepo repository.IMaterialRepository,
	teamRepo repository.ITeamRepository,
	objectRepo repository.IObjectRepository,
	materialDefectRepo repository.IMaterialDefectRepository,
) IMaterialLocationService {
	return &materialLocationService{
		materialLocationRepo: materialLocationRepo,
		materialCostRepo:     materialCostRepo,
		materialRepo:         materialRepo,
		teamRepo:             teamRepo,
		objectRepo:           objectRepo,
		materialDefectRepo:   materialDefectRepo,
	}
}

type IMaterialLocationService interface {
	GetAll() ([]model.MaterialLocation, error)
	GetPaginated(page, limit int, data model.MaterialLocation) ([]model.MaterialLocation, error)
	GetByID(id uint) (model.MaterialLocation, error)
	Create(data model.MaterialLocation) (model.MaterialLocation, error)
	Update(data model.MaterialLocation) (model.MaterialLocation, error)
	Delete(id uint) error
	Count() (int64, error)
	GetMaterialsInLocation(locationType string, locationID uint) ([]model.Material, error)
	UniqueObjects() ([]string, error)
	UniqueTeams() ([]string, error)
  BalanceReport(data dto.ReportBalanceFilterRequest) (string, error)
}

func (service *materialLocationService) GetAll() ([]model.MaterialLocation, error) {
	return service.materialLocationRepo.GetAll()
}

func (service *materialLocationService) GetPaginated(page, limit int, data model.MaterialLocation) ([]model.MaterialLocation, error) {
	if !(utils.IsEmptyFields(data)) {
		return service.materialLocationRepo.GetPaginatedFiltered(page, limit, data)
	}

	return service.materialLocationRepo.GetPaginated(page, limit)
}

func (service *materialLocationService) GetByID(id uint) (model.MaterialLocation, error) {
	return service.materialLocationRepo.GetByID(id)
}

func (service *materialLocationService) Create(data model.MaterialLocation) (model.MaterialLocation, error) {
	return service.materialLocationRepo.Create(data)
}

func (service *materialLocationService) Update(data model.MaterialLocation) (model.MaterialLocation, error) {
	return service.materialLocationRepo.Update(data)
}

func (service *materialLocationService) Delete(id uint) error {
	return service.materialLocationRepo.Delete(id)
}

func (service *materialLocationService) Count() (int64, error) {
	return service.materialLocationRepo.Count()
}

func (service *materialLocationService) GetMaterialsInLocation(
	locationType string,
	locationID uint,
) ([]model.Material, error) {
	materialCostIDs, err := service.materialLocationRepo.GetUniqueMaterialCostsByLocation(locationType, locationID)
	if err != nil {
		return []model.Material{}, err
	}

	var result []model.Material
	for _, materialCostID := range materialCostIDs {
		materialCost, err := service.materialCostRepo.GetByID(materialCostID)
		if err != nil {
			return []model.Material{}, err
		}

		material, err := service.materialRepo.GetByID(materialCost.MaterialID)
		if err != nil {
			return []model.Material{}, err
		}

		exist := false
		for _, alreadyIn := range result {
			if alreadyIn.ID == material.ID {
				exist = true
				break
			}
		}

		if !exist {
			result = append(result, material)
		}
	}

	return result, nil
}

func (service *materialLocationService) UniqueTeams() ([]string, error) {

	teamIDs, err := service.materialLocationRepo.UniqueTeamIDs()
	if err != nil {
		return []string{}, err
	}

	teams, err := service.teamRepo.GetByRangeOfIDs(teamIDs)
	if err != nil {
		return []string{}, err
	}

	result := []string{}
	for _, team := range teams {
		result = append(result, team.Number)
	}

	return result, nil
}

func (service *materialLocationService) UniqueObjects() ([]string, error) {

	objectIDs, err := service.materialLocationRepo.UniqueObjectIDs()
	if err != nil {
		return []string{}, err
	}

	objects, err := service.objectRepo.GetByRangeOfIDs(objectIDs)
	if err != nil {
		return []string{}, err
	}

	result := []string{}
	for _, object := range objects {
		result = append(result, object.Name)
	}

	return result, err
}

func (service *materialLocationService) BalanceReport(data dto.ReportBalanceFilterRequest) (string, error) {

	filter := dto.ReportBalanceFilter{
		LocationType: data.Type,
	}

	switch data.Type {
	case "teams":
		if data.Team != "" {
			team, err := service.teamRepo.GetByNumber(data.Team)
			if err != nil {
				return "", err
			}

			filter.LocationID = team.ID
			break
		}

		filter.LocationID = 0
		break
	case "objects":
		if data.Object != "" {
			object, err := service.objectRepo.GetByName(data.Object)
			if err != nil {
				return "", err
			}

			filter.LocationID = object.ID
			break
		}

		filter.LocationID = 0
		break
	case "warehouse":
		filter.LocationID = 0
	default:
		return "", fmt.Errorf("incorrect type")
	}

	materialLocations, err := service.materialLocationRepo.GetByLocationTypeAndID(filter.LocationType, filter.LocationID)
	if err != nil {
		return "", err
	}

	f, err := excelize.OpenFile("./pkg/excels/report/Balance Report.xlsx")
  defer f.Close()
	if err != nil {
		return "", err
	}

	sheetName := "Sheet1"
	rowCount := 2

	for _, materialLocation := range materialLocations {
    if materialLocation.Amount == 0 {
      continue
    }

		materialCost, err := service.materialCostRepo.GetByID(materialLocation.MaterialCostID)
		if err != nil {
			return "", err
		}

		material, err := service.materialRepo.GetByID(materialCost.MaterialID)
		if err != nil {
			return "", err
		}

		materialDefect, err := service.materialDefectRepo.GetByMaterialLocationID(materialLocation.ID)
    if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
      return "", err
    }

    f.SetCellValue(sheetName, "A"+fmt.Sprint(rowCount), material.Code)
    f.SetCellValue(sheetName, "B"+fmt.Sprint(rowCount), material.Name)
    f.SetCellValue(sheetName, "C"+fmt.Sprint(rowCount), material.Unit)
    f.SetCellValue(sheetName, "D"+fmt.Sprint(rowCount), materialLocation.Amount)
    f.SetCellValue(sheetName, "E"+fmt.Sprint(rowCount), materialDefect.Amount)
    costM19, _ := materialCost.CostM19.Float64()
    f.SetCellValue(sheetName, "F"+fmt.Sprint(rowCount), costM19)
    f.SetCellValue(sheetName, "G"+fmt.Sprint(rowCount), materialLocation.Amount * costM19)
    f.SetCellValue(sheetName, "H"+fmt.Sprint(rowCount), materialDefect.Amount * costM19)

    rowCount++
	}

  currentTime := time.Now()
  var fileName string
  if filter.LocationID == 0 {

    fileName = fmt.Sprintf(
      "Report Balance %s %s.xlsx", 
      strings.ToUpper(filter.LocationType),
      currentTime.Format("02-01-2006"),
    )

  } else {

    fileName = fmt.Sprintf(
      "Report Balance %s-%d %s.xlsx", 
      strings.ToUpper(filter.LocationType),
      filter.LocationID,
      currentTime.Format("02-01-2006"),
    )

  }

  f.SaveAs("./pkg/excels/report/" + fileName)
	if err := f.Close(); err != nil {
		fmt.Println(err)
	}

	return fileName, nil
}
