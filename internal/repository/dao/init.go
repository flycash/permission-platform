package dao

import (
	"errors"

	auditdao "gitee.com/flycash/permission-platform/internal/repository/dao/audit"
	"github.com/ego-component/egorm"
	"github.com/go-sql-driver/mysql"
)

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
		&AttributeDefinition{},
		&SubjectAttributeValue{},
		&ResourceAttributeValue{},
		&EnvironmentAttributeValue{},
		&Policy{},
		&PolicyRule{},
		&PermissionPolicy{},
		&auditdao.OperationLog{},
		&auditdao.UserRoleLog{},
	)
}

// isUniqueConstraintError 检查是否是唯一索引冲突错误
func isUniqueConstraintError(err error) bool {
	if err == nil {
		return false
	}
	me := new(mysql.MySQLError)
	if ok := errors.As(err, &me); ok {
		const uniqueIndexErrNo uint16 = 1062
		return me.Number == uniqueIndexErrNo
	}
	return false
}
