package dao

import (
	"context"
	"time"

	"github.com/ego-component/egorm"
)

// EffectType 权限效果枚举
type EffectType string

const (
	EffectTypeAllow EffectType = "allow" // 允许
	EffectTypeDeny  EffectType = "deny"  // 拒绝
)

// UserPermission 用户个人权限关联表
type UserPermission struct {
	ID             int64      `gorm:"primaryKey;autoIncrement;comment:'用户权限关联关系ID'"`
	BizID          int64      `gorm:"type:BIGINT;NOT NULL;uniqueIndex:uk_biz_user_permission,priority:1;index:idx_biz_user,priority:1;index:idx_biz_permission,priority:1;index:idx_biz_effect,priority:1;index:idx_biz_resource_type,priority:1;index:idx_biz_action,priority:1;index:idx_time_range,priority:1;index:idx_current_valid,priority:1;index:idx_biz_resource_key_action,priority:1;comment:'业务ID'"`
	UserID         int64      `gorm:"type:BIGINT;NOT NULL;uniqueIndex:uk_biz_user_permission,priority:2;index:idx_biz_user,priority:2;comment:'用户ID'"`
	PermissionID   int64      `gorm:"type:BIGINT;NOT NULL;uniqueIndex:uk_biz_user_permission,priority:3;index:idx_biz_permission,priority:2;comment:'权限ID'"`
	PermissionName string     `gorm:"type:VARCHAR(100);NOT NULL;comment:'权限名称（冗余字段，加速查询与展示）'"`
	ResourceType   string     `gorm:"type:VARCHAR(100);NOT NULL;index:idx_biz_resource_type,priority:2;index:idx_biz_resource_key_action,priority:2;comment:'资源类型（冗余字段，加速查询）'"`
	ResourceKey    string     `gorm:"type:VARCHAR(255);NOT NULL;index:idx_biz_resource_key_action,priority:3;comment:'资源标识符（冗余字段，加速查询）'"`
	ResourceName   string     `gorm:"type:VARCHAR(255);NOT NULL;comment:'资源名称（冗余字段，加速查询与展示）'"`
	Action         ActionType `gorm:"type:ENUM('create', 'read', 'update', 'delete', 'execute', 'export', 'import');NOT NULL;index:idx_biz_action,priority:2;index:idx_biz_resource_key_action,priority:4;comment:'操作类型（冗余字段，加速查询）'"`
	StartTime      int64      `gorm:"NULL;index:idx_time_range,priority:2;index:idx_current_valid,priority:3;comment:'权限生效时间，如果设置了表示临时权限'"`
	EndTime        int64      `gorm:"NULL;index:idx_time_range,priority:3;index:idx_current_valid,priority:4;comment:'权限失效时间，如果设置了表示临时权限'"`
	Effect         EffectType `gorm:"type:ENUM('allow', 'deny');NOT NULL;DEFAULT:'allow';index:idx_biz_effect,priority:2;index:idx_current_valid,priority:2;comment:'用于额外授予权限，或者取消权限，理论上不应该出现同时allow和deny，出现了就是deny优先于allow'"`
	Ctime          int64
	Utime          int64
}

func (UserPermission) TableName() string {
	return "user_permissions"
}

// UserPermissionDAO 用户权限关联数据访问接口
type UserPermissionDAO interface {
	// GetByID 根据ID获取用户权限关联
	GetByID(ctx context.Context, id int64) (UserPermission, error)
	// GetByIDs 根据多个ID批量获取用户权限关联
	GetByIDs(ctx context.Context, ids []int64) (map[int64]UserPermission, error)
	// FindByBizID 查找特定业务下的用户权限关联
	FindByBizID(ctx context.Context, bizID int64, offset, limit int) ([]UserPermission, error)
	// FindByUserID 查找特定用户的所有权限关联
	FindByUserID(ctx context.Context, userID int64) ([]UserPermission, error)
	// FindByPermissionID 查找拥有特定权限的所有用户关联
	FindByPermissionID(ctx context.Context, permissionID int64) ([]UserPermission, error)
	// FindByBizIDAndEffect 查找特定业务下指定效果的权限关联
	FindByBizIDAndEffect(ctx context.Context, bizID int64, effect EffectType, offset, limit int) ([]UserPermission, error)
	// FindByBizIDAndResourceType 查找特定业务下指定资源类型的权限关联
	FindByBizIDAndResourceType(ctx context.Context, bizID int64, resourceType string, offset, limit int) ([]UserPermission, error)
	// FindByBizIDAndAction 查找特定业务下指定操作类型的权限关联
	FindByBizIDAndAction(ctx context.Context, bizID int64, action ActionType, offset, limit int) ([]UserPermission, error)
	// FindByBizIDAndResourceKeyAction 查找特定业务下针对特定资源和操作的权限关联
	FindByBizIDAndResourceKeyAction(ctx context.Context, bizID int64, resourceKey string, action ActionType, offset, limit int) ([]UserPermission, error)
	// FindValidPermissions 查找当前有效的权限关联
	FindValidPermissions(ctx context.Context, bizID, userID, currentTime int64) ([]UserPermission, error)
	// ExistsByUserIDAndPermissionID 检查用户和权限关联是否存在
	ExistsByUserIDAndPermissionID(ctx context.Context, bizID, userID, permissionID int64) (bool, error)
	// Create 创建用户权限关联
	Create(ctx context.Context, userPermission UserPermission) (UserPermission, error)
	// BatchCreate 批量创建用户权限关联
	BatchCreate(ctx context.Context, userPermissions []UserPermission) error
	// Update 更新用户权限关联
	Update(ctx context.Context, userPermission UserPermission) error
	// Delete 删除用户权限关联
	Delete(ctx context.Context, id int64) error
	// DeleteByUserIDAndPermissionID 删除特定用户和权限的关联
	DeleteByUserIDAndPermissionID(ctx context.Context, bizID, userID, permissionID int64) error
	// DeleteByUserID 删除用户的所有权限关联
	DeleteByUserID(ctx context.Context, userID int64) error
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

func (u *userPermissionDAO) GetByID(ctx context.Context, id int64) (UserPermission, error) {
	var userPermission UserPermission
	err := u.db.WithContext(ctx).Where("id = ?", id).First(&userPermission).Error
	return userPermission, err
}

func (u *userPermissionDAO) GetByIDs(ctx context.Context, ids []int64) (map[int64]UserPermission, error) {
	var userPermissions []UserPermission
	err := u.db.WithContext(ctx).Where("id IN (?)", ids).Find(&userPermissions).Error
	if err != nil {
		return nil, err
	}

	result := make(map[int64]UserPermission, len(userPermissions))
	for i := range userPermissions {
		result[userPermissions[i].ID] = userPermissions[i]
	}
	return result, nil
}

func (u *userPermissionDAO) FindByBizID(ctx context.Context, bizID int64, offset, limit int) ([]UserPermission, error) {
	var userPermissions []UserPermission
	err := u.db.WithContext(ctx).Where("biz_id = ?", bizID).Offset(offset).Limit(limit).Find(&userPermissions).Error
	return userPermissions, err
}

func (u *userPermissionDAO) FindByUserID(ctx context.Context, userID int64) ([]UserPermission, error) {
	var userPermissions []UserPermission
	err := u.db.WithContext(ctx).Where("user_id = ?", userID).Find(&userPermissions).Error
	return userPermissions, err
}

func (u *userPermissionDAO) FindByPermissionID(ctx context.Context, permissionID int64) ([]UserPermission, error) {
	var userPermissions []UserPermission
	err := u.db.WithContext(ctx).Where("permission_id = ?", permissionID).Find(&userPermissions).Error
	return userPermissions, err
}

func (u *userPermissionDAO) FindByBizIDAndEffect(ctx context.Context, bizID int64, effect EffectType, offset, limit int) ([]UserPermission, error) {
	var userPermissions []UserPermission
	err := u.db.WithContext(ctx).
		Where("biz_id = ? AND effect = ?", bizID, effect).
		Offset(offset).
		Limit(limit).
		Find(&userPermissions).Error
	return userPermissions, err
}

func (u *userPermissionDAO) FindByBizIDAndResourceType(ctx context.Context, bizID int64, resourceType string, offset, limit int) ([]UserPermission, error) {
	var userPermissions []UserPermission
	err := u.db.WithContext(ctx).
		Where("biz_id = ? AND resource_type = ?", bizID, resourceType).
		Offset(offset).
		Limit(limit).
		Find(&userPermissions).Error
	return userPermissions, err
}

func (u *userPermissionDAO) FindByBizIDAndAction(ctx context.Context, bizID int64, action ActionType, offset, limit int) ([]UserPermission, error) {
	var userPermissions []UserPermission
	err := u.db.WithContext(ctx).
		Where("biz_id = ? AND action = ?", bizID, action).
		Offset(offset).
		Limit(limit).
		Find(&userPermissions).Error
	return userPermissions, err
}

func (u *userPermissionDAO) FindByBizIDAndResourceKeyAction(ctx context.Context, bizID int64, resourceKey string, action ActionType, offset, limit int) ([]UserPermission, error) {
	var userPermissions []UserPermission
	err := u.db.WithContext(ctx).
		Where("biz_id = ? AND resource_key = ? AND action = ?", bizID, resourceKey, action).
		Offset(offset).
		Limit(limit).
		Find(&userPermissions).Error
	return userPermissions, err
}

func (u *userPermissionDAO) FindValidPermissions(ctx context.Context, bizID, userID, currentTime int64) ([]UserPermission, error) {
	var userPermissions []UserPermission
	err := u.db.WithContext(ctx).
		Where("biz_id = ? AND user_id = ? AND (start_time IS NULL OR start_time <= ?) AND (end_time IS NULL OR end_time >= ?)",
			bizID, userID, currentTime, currentTime).
		Find(&userPermissions).Error
	return userPermissions, err
}

func (u *userPermissionDAO) ExistsByUserIDAndPermissionID(ctx context.Context, bizID, userID, permissionID int64) (bool, error) {
	var count int64
	err := u.db.WithContext(ctx).
		Model(&UserPermission{}).
		Where("biz_id = ? AND user_id = ? AND permission_id = ?", bizID, userID, permissionID).
		Count(&count).Error
	return count > 0, err
}

func (u *userPermissionDAO) Create(ctx context.Context, userPermission UserPermission) (UserPermission, error) {
	now := time.Now().UnixMilli()
	userPermission.Ctime = now
	userPermission.Utime = now
	err := u.db.WithContext(ctx).Create(&userPermission).Error
	return userPermission, err
}

func (u *userPermissionDAO) BatchCreate(ctx context.Context, userPermissions []UserPermission) error {
	if len(userPermissions) == 0 {
		return nil
	}

	now := time.Now().UnixMilli()
	for i := range userPermissions {
		userPermissions[i].Ctime = now
		userPermissions[i].Utime = now
	}

	return u.db.WithContext(ctx).CreateInBatches(userPermissions, batchSize).Error
}

func (u *userPermissionDAO) Update(ctx context.Context, userPermission UserPermission) error {
	userPermission.Utime = time.Now().UnixMilli()
	return u.db.WithContext(ctx).
		Model(&UserPermission{}).
		Where("id = ?", userPermission.ID).
		Updates(map[string]interface{}{
			"start_time": userPermission.StartTime,
			"end_time":   userPermission.EndTime,
			"effect":     userPermission.Effect,
			"utime":      userPermission.Utime,
		}).Error
}

func (u *userPermissionDAO) Delete(ctx context.Context, id int64) error {
	return u.db.WithContext(ctx).Where("id = ?", id).Delete(&UserPermission{}).Error
}

func (u *userPermissionDAO) DeleteByUserIDAndPermissionID(ctx context.Context, bizID, userID, permissionID int64) error {
	return u.db.WithContext(ctx).
		Where("biz_id = ? AND user_id = ? AND permission_id = ?", bizID, userID, permissionID).
		Delete(&UserPermission{}).Error
}

func (u *userPermissionDAO) DeleteByUserID(ctx context.Context, userID int64) error {
	return u.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&UserPermission{}).Error
}
