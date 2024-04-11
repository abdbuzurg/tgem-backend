package database

import (
	"backend-v2/model"
	"backend-v2/pkg/security"
	"fmt"
	"os"

	"gorm.io/gorm"
)

func InitialMigration(db *gorm.DB) {

	superadmin_worker := model.Worker{
		Name:         "Суперадмин",
		JobTitle:     "Главный администратор сисетмы",
		MobileNumber: "9929999999",
	}

	err := db.Table("workers").FirstOrCreate(&superadmin_worker, model.Worker{Name: "Суперадмин"}).Error
	if err != nil {
		panic("Не удалось проверить наличие главного администратора как работника системы")
	}

	err = execSeedFile(db, "./pkg/database/seed/role.sql")
	if err != nil {
		panic(err)
	}

	superadmin_userPassword, err := security.Hash("password")
	if err != nil {
		panic("Не удалось создать пароль для главного администратора")
	}

	superadmin_user := model.User{
		WorkerID: superadmin_worker.ID,
		Username: "superadmin",
		RoleID:   1,
		Password: string(superadmin_userPassword),
	}

	err = db.Table("users").FirstOrCreate(&superadmin_user, model.User{Username: "superadmin"}).Error
	if err != nil {
		panic("Не удалось проверить наличие главного администратора как пользователя системы")
	}

  err = execSeedFile(db, "./pkg/database/seed/project_dev.sql")

	superAdminInProject := model.UserInProject{
		ProjectID: 1,
		UserID:    superadmin_user.ID,
	}

	err = db.Table("user_in_projects").FirstOrCreate(&superAdminInProject, model.UserInProject{UserID: 1, ProjectID: 1}).Error
	if err != nil {
		panic("Не удалось привязать администратора к проекту номер 1")
	}

	err = execSeedFile(db, "./pkg/database/seed/permission.sql")
	if err != nil {
		panic(err)
	}

}

func execSeedFile(db *gorm.DB, filepath string) error {

	file, err := os.ReadFile(filepath)
	if err != nil {
    return fmt.Errorf("не удалось найти файл permission.sql: %v", err)
	}

	sql := string(file)

	err = db.Exec(sql).Error
	if err != nil {
    return fmt.Errorf("не удалось запустить изначальный скрипт seed для доступов: %v", err)
	}

	return nil
}
