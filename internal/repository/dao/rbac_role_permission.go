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
	ID           int64      `gorm:"primaryKey;autoIncrement;comment:'角色权限关联关系ID'"`
	BizID        int64      `gorm:"type:BIGINT;NOT NULL;uniqueIndex:uk_biz_role_permission,priority:1;index:idx_biz_role,priority:1;index:idx_biz_permission,priority:1;index:idx_biz_role_type,priority:1;index:idx_biz_resource_type,priority:1;index:idx_biz_action,priority:1;index:idx_biz_resource_key_action,priority:1;comment:'业务ID'"`
	RoleID       int64      `gorm:"type:BIGINT;NOT NULL;uniqueIndex:uk_biz_role_permission,priority:2;index:idx_biz_role,priority:2;comment:'角色ID'"`
	PermissionID int64      `gorm:"type:BIGINT;NOT NULL;uniqueIndex:uk_biz_role_permission,priority:3;index:idx_biz_permission,priority:2;comment:'权限ID'"`
	RoleName     string     `gorm:"type:VARCHAR(100);NOT NULL;comment:'角色名称（冗余字段，加速查询）'"`
	RoleType     RoleType   `gorm:"type:ENUM('system', 'custom', 'temporary');NOT NULL;index:idx_biz_role_type,priority:2;comment:'角色类型（冗余字段，加速查询）'"`
	ResourceType string     `gorm:"type:VARCHAR(100);NOT NULL;index:idx_biz_resource_type,priority:2;index:idx_biz_resource_key_action,priority:2;comment:'资源类型（冗余字段，加速查询）'"`
	ResourceKey  string     `gorm:"type:VARCHAR(255);NOT NULL;index:idx_biz_resource_key_action,priority:3;comment:'资源标识符（冗余字段，加速查询）'"`
	Action       ActionType `gorm:"type:ENUM('create', 'read', 'update', 'delete', 'execute', 'export', 'import');NOT NULL;index:idx_biz_action,priority:2;index:idx_biz_resource_key_action,priority:4;comment:'操作类型（冗余字段，加速查询）'"`
	Ctime        int64
	Utime        int64
}

func (RolePermission) TableName() string {
	return "role_permissions"
}

// RolePermissionDAO 角色权限关联数据访问接口
type RolePermissionDAO interface {
	// FindByRoleID 查找特定角色的所有权限关联
	FindByRoleID(ctx context.Context, bizID int64, roleID int64) ([]RolePermission, error)
	// Create 创建角色权限关联
	Create(ctx context.Context, rolePermission RolePermission) (RolePermission, error)
	// DeleteByRoleIDAndPermissionID 删除特定角色和权限的关联
	DeleteByRoleIDAndPermissionID(ctx context.Context, bizID int64, roleID int64, permissionID int64) error
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

func (r *rolePermissionDAO) GetByID(ctx context.Context, id int64) (RolePermission, error) {
	var rolePermission RolePermission
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&rolePermission).Error
	return rolePermission, err
}

func (r *rolePermissionDAO) GetByIDs(ctx context.Context, ids []int64) (map[int64]RolePermission, error) {
	var rolePermissions []RolePermission
	err := r.db.WithContext(ctx).Where("id IN (?)", ids).Find(&rolePermissions).Error
	if err != nil {
		return nil, err
	}

	result := make(map[int64]RolePermission, len(rolePermissions))
	for i := range rolePermissions {
		result[rolePermissions[i].ID] = rolePermissions[i]
	}
	return result, nil
}

func (r *rolePermissionDAO) FindByBizID(ctx context.Context, bizID int64, offset, limit int) ([]RolePermission, error) {
	var rolePermissions []RolePermission
	err := r.db.WithContext(ctx).Where("biz_id = ?", bizID).Offset(offset).Limit(limit).Find(&rolePermissions).Error
	return rolePermissions, err
}

func (r *rolePermissionDAO) FindByPermissionID(ctx context.Context, permissionID int64) ([]RolePermission, error) {
	var rolePermissions []RolePermission
	err := r.db.WithContext(ctx).Where("permission_id = ?", permissionID).Find(&rolePermissions).Error
	return rolePermissions, err
}

func (r *rolePermissionDAO) FindByBizIDAndRoleType(ctx context.Context, bizID int64, roleType RoleType, offset, limit int) ([]RolePermission, error) {
	var rolePermissions []RolePermission
	err := r.db.WithContext(ctx).
		Where("biz_id = ? AND role_type = ?", bizID, roleType).
		Offset(offset).
		Limit(limit).
		Find(&rolePermissions).Error
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

func (r *rolePermissionDAO) FindByBizIDAndAction(ctx context.Context, bizID int64, action ActionType, offset, limit int) ([]RolePermission, error) {
	var rolePermissions []RolePermission
	err := r.db.WithContext(ctx).
		Where("biz_id = ? AND action = ?", bizID, action).
		Offset(offset).
		Limit(limit).
		Find(&rolePermissions).Error
	return rolePermissions, err
}

func (r *rolePermissionDAO) FindByBizIDAndResourceKeyAction(ctx context.Context, bizID int64, resourceKey string, action ActionType, offset, limit int) ([]RolePermission, error) {
	var rolePermissions []RolePermission
	err := r.db.WithContext(ctx).
		Where("biz_id = ? AND resource_key = ? AND action = ?", bizID, resourceKey, action).
		Offset(offset).
		Limit(limit).
		Find(&rolePermissions).Error
	return rolePermissions, err
}

func (r *rolePermissionDAO) ExistsByRoleIDAndPermissionID(ctx context.Context, bizID, roleID, permissionID int64) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&RolePermission{}).
		Where("biz_id = ? AND role_id = ? AND permission_id = ?", bizID, roleID, permissionID).
		Count(&count).Error
	return count > 0, err
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

func (r *rolePermissionDAO) BatchCreate(ctx context.Context, rolePermissions []RolePermission) error {
	if len(rolePermissions) == 0 {
		return nil
	}

	now := time.Now().UnixMilli()
	for i := range rolePermissions {
		rolePermissions[i].Ctime = now
		rolePermissions[i].Utime = now
	}

	return r.db.WithContext(ctx).CreateInBatches(rolePermissions, batchSize).Error
}

func (r *rolePermissionDAO) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&RolePermission{}).Error
}

func (r *rolePermissionDAO) DeleteByRoleID(ctx context.Context, roleID int64) error {
	return r.db.WithContext(ctx).Where("role_id = ?", roleID).Delete(&RolePermission{}).Error
}

func (r *rolePermissionDAO) FindByRoleID(ctx context.Context, bizID, roleID int64) ([]RolePermission, error) {
	var rolePermissions []RolePermission
	err := r.db.WithContext(ctx).Where("biz_id = ? AND role_id = ?", bizID, roleID).Find(&rolePermissions).Error
	return rolePermissions, err
}

func (r *rolePermissionDAO) DeleteByRoleIDAndPermissionID(ctx context.Context, bizID, roleID, permissionID int64) error {
	return r.db.WithContext(ctx).
		Where("biz_id = ? AND role_id = ? AND permission_id = ?", bizID, roleID, permissionID).
		Delete(&RolePermission{}).Error
}
