package dao

import (
	"context"
	"time"

	"github.com/ego-component/egorm"
)

// UserPermission 用户个人权限关联表
type UserPermission struct {
	ID               int64  `gorm:"primaryKey;autoIncrement;comment:'用户权限关联关系ID'"`
	BizID            int64  `gorm:"type:BIGINT;NOT NULL;uniqueIndex:uk_biz_user_permission,priority:1;index:idx_biz_user,priority:1;index:idx_biz_permission,priority:1;index:idx_biz_effect,priority:1;index:idx_biz_resource_type,priority:1;index:idx_biz_action,priority:1;index:idx_time_range,priority:1;index:idx_current_valid,priority:1;index:idx_biz_resource_key_action,priority:1;comment:'业务ID'"`
	UserID           int64  `gorm:"type:BIGINT;NOT NULL;uniqueIndex:uk_biz_user_permission,priority:2;index:idx_biz_user,priority:2;comment:'用户ID'"`
	PermissionID     int64  `gorm:"type:BIGINT;NOT NULL;uniqueIndex:uk_biz_user_permission,priority:3;index:idx_biz_permission,priority:2;comment:'权限ID'"`
	PermissionName   string `gorm:"type:VARCHAR(255);NOT NULL;comment:'权限名称（冗余字段，加速查询与展示）'"`
	ResourceType     string `gorm:"type:VARCHAR(255);NOT NULL;index:idx_biz_resource_type,priority:2;index:idx_biz_resource_key_action,priority:2;comment:'资源类型（冗余字段，加速查询）'"`
	ResourceKey      string `gorm:"type:VARCHAR(255);NOT NULL;index:idx_biz_resource_key_action,priority:3;comment:'资源标识符（冗余字段，加速查询）'"`
	PermissionAction string `gorm:"type:VARCHAR(255);NOT NULL;index:idx_biz_action,priority:2;index:idx_biz_resource_key_action,priority:4;comment:'操作类型（冗余字段，加速查询）'"`
	StartTime        int64  `gorm:"NOT NULL;index:idx_time_range,priority:2;index:idx_current_valid,priority:3;comment:'权限生效时间'"`
	EndTime          int64  `gorm:"NOT NULL;index:idx_time_range,priority:3;index:idx_current_valid,priority:4;comment:'权限失效时间'"`
	Effect           string `gorm:"type:ENUM('allow', 'deny');NOT NULL;DEFAULT:'allow';index:idx_biz_effect,priority:2;index:idx_current_valid,priority:2;comment:'用于额外授予权限，或者取消权限，理论上不应该出现同时allow和deny，出现了就是deny优先于allow'"`
	Ctime            int64
	Utime            int64
}

func (UserPermission) TableName() string {
	return "user_permissions"
}

// UserPermissionDAO 用户权限关联数据访问接口
type UserPermissionDAO interface {
	Create(ctx context.Context, userPermission UserPermission) (UserPermission, error)

	FindByBizID(ctx context.Context, bizID int64, offset, limit int) ([]UserPermission, error)
	FindByBizIDAndUserID(ctx context.Context, bizID, userID int64) ([]UserPermission, error)

	DeleteByBizIDAndID(ctx context.Context, bizID, id int64) error
	DeleteByBizIDAndUserIDAndPermissionID(ctx context.Context, bizID, userID, permissionID int64) error
}

// userPermissionDAO 用户权限关联数据访问实现
type userPermissionDAO struct {
	db *egorm.Component
}

// NewUserPermissionDAO 创建用户权限关联数据访问对象
func NewUserPermissionDAO(db *egorm.Component) UserPermissionDAO {
	return &userPermissionDAO{
		db: db,
	}
}

func (u *userPermissionDAO) Create(ctx context.Context, userPermission UserPermission) (UserPermission, error) {
	now := time.Now().UnixMilli()
	userPermission.Ctime = now
	userPermission.Utime = now
	err := u.db.WithContext(ctx).Create(&userPermission).Error
	return userPermission, err
}

func (u *userPermissionDAO) FindByBizID(ctx context.Context, bizID int64, offset, limit int) ([]UserPermission, error) {
	var userPermissions []UserPermission
	err := u.db.WithContext(ctx).Where("biz_id = ?", bizID).Offset(offset).Limit(limit).Find(&userPermissions).Error
	return userPermissions, err
}

func (u *userPermissionDAO) FindByBizIDAndUserID(ctx context.Context, bizID, userID int64) ([]UserPermission, error) {
	now := time.Now().UnixMilli()
	var userPermissions []UserPermission
	err := u.db.WithContext(ctx).
		Where("biz_id = ? AND user_id = ? AND start_time <= ? AND end_time >= ?", bizID, userID, now, now).
		Find(&userPermissions).Error
	return userPermissions, err
}

func (u *userPermissionDAO) DeleteByBizIDAndID(ctx context.Context, bizID, id int64) error {
	return u.db.WithContext(ctx).Where("biz_id = ? AND id = ?", bizID, id).Delete(&UserPermission{}).Error
}

func (u *userPermissionDAO) DeleteByBizIDAndUserIDAndPermissionID(ctx context.Context, bizID, userID, permissionID int64) error {
	return u.db.WithContext(ctx).Where("biz_id = ? AND user_id = ? AND permission_id = ?", bizID, userID, permissionID).Delete(&UserPermission{}).Error
}
