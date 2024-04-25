package database

import (
	"fmt"
	"os"

	"gorm.io/gorm"
)

func InitialMigration(db *gorm.DB) {
  err := execSeedFile(db, "./pkg/database/seed/project_dev.sql")
  if err != nil {
    panic(err)
  } 

  err = execSeedFile(db, "./pkg/database/seed/resource.sql")
  if err != nil {
    panic(err)
  }

  err = execSeedFile(db, "./pkg/database/seed/superadmin.sql")
  if err != nil {
    panic(err)
  }
}
// Function for running SEED scripts
// filepath should related to the file main.go
func execSeedFile(db *gorm.DB, filepath string) error {

	file, err := os.ReadFile(filepath)
	if err != nil {
    return fmt.Errorf("не удалось найти файл в %v: %v", filepath, err)
	}

	sql := string(file)

	err = db.Exec(sql).Error
	if err != nil {
    return fmt.Errorf("не удалось запустить изначальный скрипт seed для доступов: %v", err)
	}

	return nil
}
