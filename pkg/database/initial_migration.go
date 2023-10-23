package database

import (
	"backend-v2/model"
	"backend-v2/pkg/security"

	"github.com/casbin/casbin/v2"
	gormadapter "github.com/casbin/gorm-adapter/v3"
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

	superadmin_userPassword, err := security.Hash("password")
	if err != nil {
		panic("Не удалось создать пароль для главного администратора")
	}
	superadmin_user := model.User{
		WorkerID: superadmin_worker.ID,
		Username: "superadmin",
		Password: string(superadmin_userPassword),
	}

	err = db.Table("users").FirstOrCreate(&superadmin_user, model.User{Username: "superadmin"}).Error
	if err != nil {
		panic("Не удалось проверить наличие главного администратора как пользователя системы")
	}

	err = db.Raw("DELETE FROM casbin_ruler WHERE v0 = ?", superadmin_user.ID).Error
	if err != nil {
		panic("Не удалось попрабить доступы администратора так как они не совпадают с преждевнесенными доступами")
	}

	a, err := gormadapter.NewAdapterByDB(db)
	if err != nil {
		panic(err)
	}
	gormadapter.TurnOffAutoMigrate(db)

	e, err := casbin.NewEnforcer("./pkg/database/acl_model.conf", a)
	if err != nil {
		panic(err)
	}

	allPolicies := e.GetPolicy()
	if len(allPolicies) < len(SUPERADMIN_ACL_POLICIES) {
		for _, policy := range SUPERADMIN_ACL_POLICIES {
			_, err := e.AddPolicy(policy[0], policy[1], policy[2])
			if err != nil {
				panic("Не удалось добавить доступ к администартору")
			}
		}
	}

}
