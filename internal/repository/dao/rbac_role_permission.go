package dao

import (
	"context"
	"fmt"
	"time"

	"gitee.com/flycash/permission-platform/internal/errs"
	"github.com/ego-component/egorm"
)

// RolePermission 角色权限关联表
type RolePermission struct {
	ID               int64  `gorm:"primaryKey;autoIncrement;comment:'角色权限关联关系ID'"`
	BizID            int64  `gorm:"type:BIGINT;NOT NULL;uniqueIndex:uk_biz_role_permission,priority:1;index:idx_biz_role,priority:1;index:idx_biz_permission,priority:1;index:idx_biz_role_type,priority:1;index:idx_biz_resource_type,priority:1;index:idx_biz_action,priority:1;index:idx_biz_resource_key_action,priority:1;comment:'业务ID'"`
	RoleID           int64  `gorm:"type:BIGINT;NOT NULL;uniqueIndex:uk_biz_role_permission,priority:2;index:idx_biz_role,priority:2;comment:'角色ID'"`
	PermissionID     int64  `gorm:"type:BIGINT;NOT NULL;uniqueIndex:uk_biz_role_permission,priority:3;index:idx_biz_permission,priority:2;comment:'权限ID'"`
	RoleName         string `gorm:"type:VARCHAR(255);NOT NULL;comment:'角色名称（冗余字段，加速查询）'"`
	RoleType         string `gorm:"type:VARCHAR(255);NOT NULL;index:idx_biz_role_type,priority:2;comment:'角色类型（冗余字段，加速查询）'"`
	ResourceType     string `gorm:"type:VARCHAR(255);NOT NULL;index:idx_biz_resource_type,priority:2;index:idx_biz_resource_key_action,priority:2;comment:'资源类型（冗余字段，加速查询）'"`
	ResourceKey      string `gorm:"type:VARCHAR(255);NOT NULL;index:idx_biz_resource_key_action,priority:3;comment:'资源标识符（冗余字段，加速查询）'"`
	PermissionAction string `gorm:"type:VARCHAR(255);NOT NULL;index:idx_biz_action,priority:2;index:idx_biz_resource_key_action,priority:4;comment:'操作类型（冗余字段，加速查询）'"`
	Ctime            int64
	Utime            int64
}

func (RolePermission) TableName() string {
	return "role_permissions"
}

// RolePermissionDAO 角色权限关联数据访问接口
type RolePermissionDAO interface {
	Create(ctx context.Context, rolePermission RolePermission) (RolePermission, error)

	FindByBizID(ctx context.Context, bizID int64, offset, limit int) ([]RolePermission, error)
	FindByBizIDAndID(ctx context.Context, bizID, id int64) (RolePermission, error)
	FindByBizIDAndRoleIDs(ctx context.Context, bizID int64, roleIDs []int64, offset, limit int) ([]RolePermission, error)
	FindByBizIDAndPermissionID(ctx context.Context, bizID, permissionID int64, offset, limit int) ([]RolePermission, error)

	DeleteByBizIDAndID(ctx context.Context, bizID, id int64) error
	DeleteByBizIDAndRoleIDAndPermissionID(ctx context.Context, bizID, roleID, permissionID int64) error
}

// rolePermissionDAO 角色权限关联数据访问实现
type rolePermissionDAO struct {
	db *egorm.Component
}

// NewRolePermissionDAO 创建角色权限关联数据访问对象
func NewRolePermissionDAO(db *egorm.Component) RolePermissionDAO {
	return &rolePermissionDAO{
		db: db,
	}
}

func (r *rolePermissionDAO) Create(ctx context.Context, rolePermission RolePermission) (RolePermission, error) {
	now := time.Now().UnixMilli()
	rolePermission.Ctime = now
	rolePermission.Utime = now
	err := r.db.WithContext(ctx).Create(&rolePermission).Error
	if err != nil {
		if isUniqueConstraintError(err) {
			return RolePermission{}, fmt.Errorf("%w", errs.ErrRolePermissionDuplicate)
		}
		return RolePermission{}, err
	}
	return rolePermission, nil
}

func (r *rolePermissionDAO) FindByBizID(ctx context.Context, bizID int64, offset, limit int) ([]RolePermission, error) {
	var rolePermissions []RolePermission
	err := r.db.WithContext(ctx).Where("biz_id = ?", bizID).Offset(offset).Limit(limit).Find(&rolePermissions).Error
	return rolePermissions, err
}

func (r *rolePermissionDAO) FindByBizIDAndID(ctx context.Context, bizID, id int64) (RolePermission, error) {
	var rolePermission RolePermission
	err := r.db.WithContext(ctx).Where("biz_id = ? AND id = ?", bizID, id).First(&rolePermission).Error
	return rolePermission, err
}

func (r *rolePermissionDAO) FindByBizIDAndPermissionID(ctx context.Context, bizID, permissionID int64, offset, limit int) ([]RolePermission, error) {
	var rolePermissions []RolePermission
	err := r.db.WithContext(ctx).Where("biz_id = ? AND permission_id = ?", bizID, permissionID).
		Offset(offset).Limit(limit).Find(&rolePermissions).Error
	return rolePermissions, err
}

func (r *rolePermissionDAO) FindByBizIDAndRoleIDs(ctx context.Context, bizID int64, roleIDs []int64, offset, limit int) ([]RolePermission, error) {
	var rolePermissions []RolePermission
	err := r.db.WithContext(ctx).Where("biz_id = ? AND role_id IN (?)", bizID, roleIDs).Offset(offset).Limit(limit).Find(&rolePermissions).Error
	return rolePermissions, err
}

func (r *rolePermissionDAO) FindByBizIDAndResourceType(ctx context.Context, bizID int64, resourceType string, offset, limit int) ([]RolePermission, error) {
	var rolePermissions []RolePermission
	err := r.db.WithContext(ctx).
		Where("biz_id = ? AND resource_type = ?", bizID, resourceType).
		Offset(offset).
		Limit(limit).
		Find(&rolePermissions).Error
	return rolePermissions, err
}

func (r *rolePermissionDAO) DeleteByBizIDAndID(ctx context.Context, bizID, id int64) error {
	return r.db.WithContext(ctx).Where("biz_id = ? AND id = ?", bizID, id).Delete(&RolePermission{}).Error
}

func (r *rolePermissionDAO) DeleteByBizIDAndRoleIDAndPermissionID(ctx context.Context, bizID, roleID, permissionID int64) error {
	return r.db.WithContext(ctx).Where("biz_id = ? AND role_id = ? AND permission_id = ?", bizID, roleID, permissionID).Delete(&RolePermission{}).Error
}
