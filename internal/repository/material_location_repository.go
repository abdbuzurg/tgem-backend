package repository

import (
	"backend-v2/model"

	"gorm.io/gorm"
)

type materialLocationRepository struct {
	db *gorm.DB
}

func InitMaterialLocationRepository(db *gorm.DB) IMaterialLocationRepository {
	return &materialLocationRepository{
		db: db,
	}
}

type IMaterialLocationRepository interface {
	GetAll() ([]model.MaterialLocation, error)
	GetPaginated(page, limit int) ([]model.MaterialLocation, error)
	GetPaginatedFiltered(page, limit int, filter model.MaterialLocation) ([]model.MaterialLocation, error)
	GetByID(id uint) (model.MaterialLocation, error)
	GetByMaterialCostIDOrCreate(materialCostID uint, locationType string, locationTypeID uint) (model.MaterialLocation, error)
	Create(data model.MaterialLocation) (model.MaterialLocation, error)
	Update(data model.MaterialLocation) (model.MaterialLocation, error)
	Delete(id uint) error
	Count() (int64, error)
	GetUniqueMaterialCostsByLocation(locationType string, locationID uint) ([]uint, error)
	UniqueObjectIDs() ([]uint, error)
	UniqueTeamIDs() ([]uint, error)
	GetByLocationTypeAndID(locationType string, locationID uint) ([]model.MaterialLocation, error)
}

func (repo *materialLocationRepository) GetAll() ([]model.MaterialLocation, error) {
	data := []model.MaterialLocation{}
	err := repo.db.Order("id desc").Find(&data).Error
	return data, err
}

func (repo *materialLocationRepository) GetPaginated(page, limit int) ([]model.MaterialLocation, error) {
	data := []model.MaterialLocation{}
	err := repo.db.Order("id desc").Offset((page - 1) * limit).Limit(limit).Find(&data).Error
	return data, err
}

func (repo *materialLocationRepository) GetPaginatedFiltered(page, limit int, filter model.MaterialLocation) ([]model.MaterialLocation, error) {
	data := []model.MaterialLocation{}
	err := repo.db.
		Raw(`SELECT * FROM materials WHERE
			(nullif(?, '') IS NULL OR material_cost_id = ?) AND
			(nullif(?, '') IS NULL OR location_id = ?) AND
			(nullif(?, '') IS NULL OR location_type = ?) AND
			(nullif(?, '') IS NULL OR amount = ?) ORDER BY id DESC LIMIT ? OFFSET ?`,
			filter.MaterialCostID, filter.MaterialCostID,
			filter.LocationID, filter.LocationID,
			filter.LocationType, filter.LocationType,
			filter.Amount, filter.Amount,
			limit, (page-1)*limit,
		).
		Scan(&data).Error

	return data, err
}

func (repo *materialLocationRepository) GetByID(id uint) (model.MaterialLocation, error) {
	data := model.MaterialLocation{}
	err := repo.db.Find(&data, "id = ?", id).Error
	return data, err
}

func (repo *materialLocationRepository) GetByMaterialCostIDOrCreate(materialCostID uint, locationType string, locationID uint) (model.MaterialLocation, error) {
	data := model.MaterialLocation{}
	err := repo.db.FirstOrCreate(&data, model.MaterialLocation{LocationType: locationType, LocationID: locationID}).Error
	return data, err
}

func (repo *materialLocationRepository) Create(data model.MaterialLocation) (model.MaterialLocation, error) {
	err := repo.db.Create(&data).Error
	return data, err
}

func (repo *materialLocationRepository) Update(data model.MaterialLocation) (model.MaterialLocation, error) {
	err := repo.db.Model(&model.MaterialLocation{}).Select("*").Updates(&data).Where("id = ?", data.ID).Error
	return data, err
}

func (repo *materialLocationRepository) Delete(id uint) error {
	return repo.db.Delete(&model.MaterialLocation{}, "id = ?", id).Error
}

func (repo *materialLocationRepository) Count() (int64, error) {
	var count int64
	err := repo.db.Model(&model.MaterialLocation{}).Count(&count).Error
	return count, err
}

func (repo *materialLocationRepository) GetUniqueMaterialCostsByLocation(
	locationType string,
	locationID uint,
) ([]uint, error) {
	var data []uint
	err := repo.db.Raw(`
    SELECT DISTINCT(material_cost_id) FROM material_locations WHERE location_type = ? and location_id = ?
    `, locationType, locationID).Scan(&data).Error

	return data, err
}

func (repo *materialLocationRepository) UniqueObjectIDs() ([]uint, error) {
	var data []uint
	err := repo.db.Raw("SELECT DISTINCT(location_id) FROM material_locations WHERE location_type='objects'").Scan(&data).Error
	return data, err
}

func (repo *materialLocationRepository) UniqueTeamIDs() ([]uint, error) {
	var data []uint
	err := repo.db.Raw("SELECT DISTINCT(location_id) FROM material_locations WHERE location_type='teams'").Scan(&data).Error
	return data, err
}

func (repo *materialLocationRepository) GetByLocationTypeAndID(
	locationType string,
	locationID uint,
) ([]model.MaterialLocation, error) {

	var data []model.MaterialLocation
	err := repo.db.Raw(`
    SELECT * 
    FROM material_locations 
    WHERE 
      location_type = ? AND
      (nullif(?, 0) IS NULL OR location_id = ?)
  `, locationType, locationID, locationID).Scan(&data).Error

	return data, err
}
