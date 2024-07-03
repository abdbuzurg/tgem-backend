package repository

import "gorm.io/gorm"

type tpNourashesObjectRepository struct {
	db *gorm.DB
}

func InitTPNourashesObjectsRepository(db *gorm.DB) ITPNourashesObjectsRepository {
	return &tpNourashesObjectRepository{
		db: db,
	}
}

type ITPNourashesObjectsRepository interface {
	GetTPObjectNames(targetID uint, targetType string) ([]string, error)
}

func (repo *tpNourashesObjectRepository) GetTPObjectNames(targetID uint, targetType string) ([]string, error) {
	result := []string{}
	err := repo.db.Raw(`
    SELECT objects.name
    FROM tp_nourashes_objects
    INNER JOIN objects ON objects.id = tp_nourashes_objects.tp_object_id
    WHERE 
      tp_nourashes_objects.target_id = ? AND
      tp_nourashes_objects.target_type = ?
  `, targetID, targetType).Scan(&result).Error

	return result, err
}
