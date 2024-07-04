package service

import (
	"backend-v2/internal/dto"
	"backend-v2/internal/repository"
	"backend-v2/model"
	"backend-v2/pkg/utils"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
)

type materialLocationService struct {
	materialLocationRepo  repository.IMaterialLocationRepository
	materialCostRepo      repository.IMaterialCostRepository
	teamRepo              repository.ITeamRepository
	materialRepo          repository.IMaterialRepository
	objectRepo            repository.IObjectRepository
	materialDefectRepo    repository.IMaterialDefectRepository
	objectSupervisorsRepo repository.IObjectSupervisorsRepository
}

func InitMaterialLocationService(
	materialLocationRepo repository.IMaterialLocationRepository,
	materialCostRepo repository.IMaterialCostRepository,
	materialRepo repository.IMaterialRepository,
	teamRepo repository.ITeamRepository,
	objectRepo repository.IObjectRepository,
	materialDefectRepo repository.IMaterialDefectRepository,
	objectSupervisorsRepo repository.IObjectSupervisorsRepository,
) IMaterialLocationService {
	return &materialLocationService{
		materialLocationRepo:  materialLocationRepo,
		materialCostRepo:      materialCostRepo,
		materialRepo:          materialRepo,
		teamRepo:              teamRepo,
		objectRepo:            objectRepo,
		materialDefectRepo:    materialDefectRepo,
		objectSupervisorsRepo: objectSupervisorsRepo,
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
	UniqueObjects() ([]dto.ObjectDataForSelect, error)
	UniqueTeams() ([]dto.TeamDataForSelect, error)
	BalanceReport(projectID uint, data dto.ReportBalanceFilterRequest) (string, error)
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

func (service *materialLocationService) UniqueTeams() ([]dto.TeamDataForSelect, error) {
	return service.materialLocationRepo.UniqueTeams()
}

func (service *materialLocationService) UniqueObjects() ([]dto.ObjectDataForSelect, error) {
	return service.materialLocationRepo.UniqueObjects()
}

func (service *materialLocationService) BalanceReport(projectID uint, data dto.ReportBalanceFilterRequest) (string, error) {

	filter := dto.ReportBalanceFilter{
		LocationType: data.Type,
	}

	templateFilePath := filepath.Join("./pkg/excels/templates/", "Отчет Остатка.xlsx")
	f, err := excelize.OpenFile(templateFilePath)
	if err != nil {
		return "", err
	}

	sheetName := "Отчет"
	rowCount := 2

	switch data.Type {
	case "team":
		f.SetCellValue(sheetName, "I1", "№ Бригады")
		f.SetCellValue(sheetName, "J1", "Бригадир")
		if data.TeamID != 0 {
			filter.LocationID = data.TeamID
			break
		}

		filter.LocationID = 0
		break

	case "object":
		f.SetCellValue(sheetName, "I1", "Супервайзер")
		f.SetCellValue(sheetName, "J1", "Объект")
		f.SetCellValue(sheetName, "K1", "Тип Объекта")

		if data.ObjectID != 0 {
			filter.LocationID = data.ObjectID
			break
		}

		filter.LocationID = 0
		break

	case "warehouse":
		filter.LocationID = 0
		break

	default:
		return "", fmt.Errorf("incorrect type")
	}

	materialsData, err := service.materialLocationRepo.GetDataForBalanceReport(projectID, filter.LocationType, filter.LocationID)
	if err != nil {
		return "", err
	}

	locationInformation := struct {
		LocationID        uint
		LocationName      string
		LocationOwnerName string
		LocationType      string
	}{
		LocationID:        0,
		LocationName:      "",
		LocationOwnerName: "",
		LocationType:      "",
	}

	for _, entry := range materialsData {

		if entry.LocationID != locationInformation.LocationID {

			locationInformation.LocationID = entry.LocationID
			locationInformation.LocationOwnerName = ""

			if filter.LocationType == "team" {

				//teamData has TeamNumber and TeamLeaderName
				//the TeamNumber is repeated but TeamLeaderName is not
				teamData, err := service.teamRepo.GetTeamNumberAndTeamLeadersByID(projectID, entry.LocationID)
				if err != nil {
					return "", fmt.Errorf("Ошибка базы: %v", err)
				}
				locationInformation.LocationName = teamData[0].TeamNumber

				for index, entry := range teamData {
					if index == len(teamData)-1 {
						locationInformation.LocationOwnerName += entry.TeamLeaderName
						break
					}

					locationInformation.LocationOwnerName += entry.TeamLeaderName + ", "
				}

			}

			if filter.LocationType == "object" {
				// objectData has objectName and supervisorName
				// the objectName is repeated but supervisorName is not repeated
				objectData, err := service.objectSupervisorsRepo.GetSupervisorAndObjectNamesByObjectID(projectID, entry.LocationID)
				if err != nil {
					return "", fmt.Errorf("Ошибка базы: %v", err)
				}
				locationInformation.LocationName = objectData[0].ObjectName
        locationInformation.LocationType = utils.ObjectTypeConverter(objectData[0].ObjectType)

				for index, entry := range objectData {
					if index == len(objectData)-1 {
						locationInformation.LocationOwnerName += entry.SupervisorName
						break
					}

					locationInformation.LocationOwnerName += entry.SupervisorName + ", "

				}

			}
		}

		f.SetCellStr(sheetName, "A"+fmt.Sprint(rowCount), entry.MaterialCode)
		f.SetCellStr(sheetName, "B"+fmt.Sprint(rowCount), entry.MaterialName)
		f.SetCellStr(sheetName, "C"+fmt.Sprint(rowCount), entry.MaterialUnit)
		f.SetCellFloat(sheetName, "D"+fmt.Sprint(rowCount), entry.TotalAmount, 2, 64)
		f.SetCellFloat(sheetName, "E"+fmt.Sprint(rowCount), entry.DefectAmount, 2, 64)

		costM19, _ := entry.MaterialCostM19.Float64()
		totalCost, _ := entry.TotalCost.Float64()
		totalDefectCost, _ := entry.TotalDefectCost.Float64()
		f.SetCellFloat(sheetName, "F"+fmt.Sprint(rowCount), costM19, 2, 64)
		f.SetCellFloat(sheetName, "G"+fmt.Sprint(rowCount), totalCost, 2, 64)
		f.SetCellFloat(sheetName, "H"+fmt.Sprint(rowCount), totalDefectCost, 2, 64)

		f.SetCellStr(sheetName, "I"+fmt.Sprint(rowCount), locationInformation.LocationOwnerName)
		f.SetCellStr(sheetName, "J"+fmt.Sprint(rowCount), locationInformation.LocationName)
		f.SetCellStr(sheetName, "K"+fmt.Sprint(rowCount), locationInformation.LocationType)

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

	tempFilePath := filepath.Join("./pkg/excels/temp/", fileName)
	f.SaveAs(tempFilePath)
	if err := f.Close(); err != nil {
		fmt.Println(err)
	}

	return fileName, nil
}
