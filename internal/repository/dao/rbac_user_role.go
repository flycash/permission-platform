package dao

import (
	"context"
	"time"

	"github.com/ego-component/egorm"
)

// UserRole 用户角色关联关系表
type UserRole struct {
	ID        int64    `gorm:"primaryKey;autoIncrement;comment:用户角色关联关系主键'"`
	BizID     int64    `gorm:"type:BIGINT;NOT NULL;uniqueIndex:uk_biz_user_role,priority:1;index:idx_biz_user,priority:1;index:idx_biz_role,priority:1;index:idx_biz_user_role_validity,priority:1;comment:'业务ID'"`
	UserID    int64    `gorm:"type:BIGINT;NOT NULL;uniqueIndex:uk_biz_user_role,priority:2;index:idx_biz_user,priority:2;index:idx_biz_user_role_validity,priority:2;comment:'用户ID'"`
	RoleID    int64    `gorm:"type:BIGINT;NOT NULL;uniqueIndex:uk_biz_user_role,priority:3;index:idx_biz_role,priority:2;comment:'角色ID'"`
	RoleName  string   `gorm:"type:VARCHAR(100);NOT NULL;comment:'角色名称（冗余字段，加速查询）'"`
	RoleType  RoleType `gorm:"type:ENUM('system', 'custom', 'temporary');NOT NULL;index:idx_biz_user_role_validity,priority:3;comment:'角色类型（冗余字段，加速查询）'"`
	StartTime int64    `gorm:"NULL;index:idx_biz_user_role_validity,priority:4;comment:'临时角色生效时间（冗余字段，加速查询）'"`
	EndTime   int64    `gorm:"NULL;index:idx_biz_user_role_validity,priority:5;comment:'临时角色失效时间（冗余字段，加速查询）'"`
	Ctime     int64
	Utime     int64
}

func (UserRole) TableName() string {
	return "user_roles"
}

// UserRoleDAO 用户角色关联数据访问接口
type UserRoleDAO interface {
	// FindByUserID 查找特定用户的所有角色关联
	FindByUserID(ctx context.Context, userID int64) ([]UserRole, error)
	// FindValidRoles 查找当前有效的角色关联
	FindValidRoles(ctx context.Context, bizID, userID int64, currentTime int64) ([]UserRole, error)
	// Create 创建用户角色关联
	Create(ctx context.Context, userRole UserRole) (UserRole, error)
	// DeleteByUserIDAndRoleID 删除特定用户和角色的关联
	DeleteByUserIDAndRoleID(ctx context.Context, bizID int64, userID int64, roleID int64) error
}

// userRoleDAO 用户角色关联数据访问实现
type userRoleDAO struct {
	db *egorm.Component
}

// NewUserRoleDAO 创建用户角色关联数据访问对象
func NewUserRoleDAO(db *egorm.Component) UserRoleDAO {
	return &userRoleDAO{
		db: db,
	}
}

func (u *userRoleDAO) GetByID(ctx context.Context, id int64) (UserRole, error) {
	var userRole UserRole
	err := u.db.WithContext(ctx).Where("id = ?", id).First(&userRole).Error
	return userRole, err
}

func (u *userRoleDAO) GetByIDs(ctx context.Context, ids []int64) (map[int64]UserRole, error) {
	var userRoles []UserRole
	err := u.db.WithContext(ctx).Where("id IN (?)", ids).Find(&userRoles).Error
	if err != nil {
		return nil, err
	}

	result := make(map[int64]UserRole, len(userRoles))
	for i := range userRoles {
		result[userRoles[i].ID] = userRoles[i]
	}
	return result, nil
}

func (u *userRoleDAO) FindByBizID(ctx context.Context, bizID int64, offset, limit int) ([]UserRole, error) {
	var userRoles []UserRole
	err := u.db.WithContext(ctx).Where("biz_id = ?", bizID).Offset(offset).Limit(limit).Find(&userRoles).Error
	return userRoles, err
}

func (u *userRoleDAO) FindByRoleID(ctx context.Context, roleID int64) ([]UserRole, error) {
	var userRoles []UserRole
	err := u.db.WithContext(ctx).Where("role_id = ?", roleID).Find(&userRoles).Error
	return userRoles, err
}

func (u *userRoleDAO) FindByBizIDAndRoleType(ctx context.Context, bizID int64, roleType RoleType, offset, limit int) ([]UserRole, error) {
	var userRoles []UserRole
	err := u.db.WithContext(ctx).
		Where("biz_id = ? AND role_type = ?", bizID, roleType).
		Offset(offset).
		Limit(limit).
		Find(&userRoles).Error
	return userRoles, err
}

func (u *userRoleDAO) ExistsByUserIDAndRoleID(ctx context.Context, bizID, userID, roleID int64) (bool, error) {
	var count int64
	err := u.db.WithContext(ctx).
		Model(&UserRole{}).
		Where("biz_id = ? AND user_id = ? AND role_id = ?", bizID, userID, roleID).
		Count(&count).Error
	return count > 0, err
}

func (u *userRoleDAO) BatchCreate(ctx context.Context, userRoles []UserRole) error {
	if len(userRoles) == 0 {
		return nil
	}

	now := time.Now().UnixMilli()
	for i := range userRoles {
		userRoles[i].Ctime = now
		userRoles[i].Utime = now
	}

	return u.db.WithContext(ctx).CreateInBatches(userRoles, batchSize).Error
}

func (u *userRoleDAO) Update(ctx context.Context, userRole UserRole) error {
	userRole.Utime = time.Now().UnixMilli()
	return u.db.WithContext(ctx).
		Model(&UserRole{}).
		Where("id = ?", userRole.ID).
		Updates(map[string]interface{}{
			"start_time": userRole.StartTime,
			"end_time":   userRole.EndTime,
			"utime":      userRole.Utime,
		}).Error
}

func (u *userRoleDAO) Delete(ctx context.Context, id int64) error {
	return u.db.WithContext(ctx).Where("id = ?", id).Delete(&UserRole{}).Error
}

func (u *userRoleDAO) DeleteByUserID(ctx context.Context, userID int64) error {
	return u.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&UserRole{}).Error
}

func (u *userRoleDAO) DeleteByRoleID(ctx context.Context, roleID int64) error {
	return u.db.WithContext(ctx).Where("role_id = ?", roleID).Delete(&UserRole{}).Error
}

func (u *userRoleDAO) FindByUserID(ctx context.Context, userID int64) ([]UserRole, error) {
	var userRoles []UserRole
	err := u.db.WithContext(ctx).Where("user_id = ?", userID).Find(&userRoles).Error
	return userRoles, err
}

func (u *userRoleDAO) FindValidRoles(ctx context.Context, bizID, userID, currentTime int64) ([]UserRole, error) {
	var userRoles []UserRole
	err := u.db.WithContext(ctx).
		Where("biz_id = ? AND user_id = ? AND (start_time IS NULL OR start_time <= ?) AND (end_time IS NULL OR end_time >= ?)",
			bizID, userID, currentTime, currentTime).
		Find(&userRoles).Error
	return userRoles, err
}

func (u *userRoleDAO) Create(ctx context.Context, userRole UserRole) (UserRole, error) {
	now := time.Now().UnixMilli()
	userRole.Ctime = now
	userRole.Utime = now
	err := u.db.WithContext(ctx).Create(&userRole).Error
	return userRole, err
}

func (u *userRoleDAO) DeleteByUserIDAndRoleID(ctx context.Context, bizID, userID, roleID int64) error {
	return u.db.WithContext(ctx).
		Where("biz_id = ? AND user_id = ? AND role_id = ?", bizID, userID, roleID).
		Delete(&UserRole{}).Error
}
