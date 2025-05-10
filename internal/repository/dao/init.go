package dao

import "github.com/ego-component/egorm"

func InitTables(db *egorm.Component) error {
	return db.AutoMigrate(
		&BusinessConfig{},
		&Resource{},
		&Permission{},
		&Role{},
		&RoleInclusion{},
		&RolePermission{},
		&UserRole{},
		&UserPermission{},
	)
}
