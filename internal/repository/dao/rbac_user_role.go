package dao

import (
	"context"
	"time"

	"github.com/ego-component/egorm"
)

// UserRole 用户角色关联关系表
type UserRole struct {
	ID        int64  `gorm:"primaryKey;autoIncrement;comment:用户角色关联关系主键'"`
	BizID     int64  `gorm:"type:BIGINT;NOT NULL;uniqueIndex:uk_biz_user_role,priority:1;index:idx_biz_user,priority:1;index:idx_biz_role,priority:1;index:idx_biz_user_role_validity,priority:1;comment:'业务ID'"`
	UserID    int64  `gorm:"type:BIGINT;NOT NULL;uniqueIndex:uk_biz_user_role,priority:2;index:idx_biz_user,priority:2;index:idx_biz_user_role_validity,priority:2;comment:'用户ID'"`
	RoleID    int64  `gorm:"type:BIGINT;NOT NULL;uniqueIndex:uk_biz_user_role,priority:3;index:idx_biz_role,priority:2;comment:'角色ID'"`
	RoleName  string `gorm:"type:VARCHAR(255);NOT NULL;comment:'角色名称（冗余字段，加速查询）'"`
	RoleType  string `gorm:"type:VARCHAR(255);NOT NULL;index:idx_biz_user_role_validity,priority:3;comment:'角色类型（冗余字段，加速查询）'"`
	StartTime int64  `gorm:"NOT NULL;index:idx_biz_user_role_validity,priority:4;comment:'授予角色生效时间'"`
	EndTime   int64  `gorm:"NOT NULL;index:idx_biz_user_role_validity,priority:5;comment:'授予角色失效时间'"`
	Ctime     int64
	Utime     int64
}

func (UserRole) TableName() string {
	return "user_roles"
}

// UserRoleDAO 用户角色关联数据访问接口
type UserRoleDAO interface {
	Create(ctx context.Context, userRole UserRole) (UserRole, error)

	FindByBizID(ctx context.Context, bizID int64, offset, limit int) ([]UserRole, error)
	CountByBizID(ctx context.Context, bizID int64) (int64, error)

	FindByBizIDAndID(ctx context.Context, bizID, id int64) (UserRole, error)

	FindByBizIDAndUserID(ctx context.Context, bizID int64, userID int64, offset, limit int) ([]UserRole, error)
	CountByBizIDAndUserID(ctx context.Context, bizID int64, userID int64) (int64, error)

	FindByBizIDAndRoleID(ctx context.Context, bizID int64, roleID int64, offset, limit int) ([]UserRole, error)
	CountByBizIDAndRoleID(ctx context.Context, bizID int64, roleID int64) (int64, error)

	FindValidUserRolesWithBizID(ctx context.Context, bizID, userID, currentTime int64, offset, limit int) ([]UserRole, error)
	CountValidUserRolesWithBizID(ctx context.Context, bizID, userID, currentTime int64) (int64, error)

	DeleteByBizIDAndID(ctx context.Context, bizID, id int64) error
	DeleteByBizIDAndUserIDAndRoleID(ctx context.Context, bizID, userID, roleID int64) error
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

func (u *userRoleDAO) Create(ctx context.Context, userRole UserRole) (UserRole, error) {
	now := time.Now().UnixMilli()
	userRole.Ctime = now
	userRole.Utime = now
	err := u.db.WithContext(ctx).Create(&userRole).Error
	return userRole, err
}

func (u *userRoleDAO) FindByBizIDAndUserID(ctx context.Context, bizID, userID int64, offset, limit int) ([]UserRole, error) {
	var userRoles []UserRole
	err := u.db.WithContext(ctx).Where("biz_id = ? AND user_id = ?", bizID, userID).Offset(offset).Limit(limit).Find(&userRoles).Error
	return userRoles, err
}

func (u *userRoleDAO) FindValidUserRolesWithBizID(ctx context.Context, bizID, userID, currentTime int64, offset, limit int) ([]UserRole, error) {
	var userRoles []UserRole
	err := u.db.WithContext(ctx).
		Where("biz_id = ? AND user_id = ? AND start_time <= ? AND end_time >= ?",
			bizID, userID, currentTime, currentTime).
		Offset(offset).
		Limit(limit).
		Find(&userRoles).Error
	return userRoles, err
}

func (u *userRoleDAO) FindByBizID(ctx context.Context, bizID int64, offset, limit int) ([]UserRole, error) {
	var userRoles []UserRole
	err := u.db.WithContext(ctx).Where("biz_id = ?", bizID).Offset(offset).Limit(limit).Find(&userRoles).Error
	return userRoles, err
}

func (u *userRoleDAO) DeleteByUserIDAndRoleID(ctx context.Context, bizID, userID, roleID int64) error {
	return u.db.WithContext(ctx).
		Where("biz_id = ? AND user_id = ? AND role_id = ?", bizID, userID, roleID).
		Delete(&UserRole{}).Error
}

func (u *userRoleDAO) CountByBizID(ctx context.Context, bizID int64) (int64, error) {
	var count int64
	err := u.db.WithContext(ctx).Model(&UserRole{}).Where("biz_id = ?", bizID).Count(&count).Error
	return count, err
}

func (u *userRoleDAO) CountByBizIDAndUserID(ctx context.Context, bizID, userID int64) (int64, error) {
	var count int64
	err := u.db.WithContext(ctx).Model(&UserRole{}).Where("biz_id = ? AND user_id = ?", bizID, userID).Count(&count).Error
	return count, err
}

func (u *userRoleDAO) CountByBizIDAndRoleID(ctx context.Context, bizID, roleID int64) (int64, error) {
	var count int64
	err := u.db.WithContext(ctx).Model(&UserRole{}).Where("biz_id = ? AND role_id = ?", bizID, roleID).Count(&count).Error
	return count, err
}

func (u *userRoleDAO) CountValidUserRolesWithBizID(ctx context.Context, bizID, userID, currentTime int64) (int64, error) {
	var count int64
	err := u.db.WithContext(ctx).Model(&UserRole{}).
		Where("biz_id = ? AND user_id = ? AND start_time <= ? AND end_time >= ? ",
			bizID, userID, currentTime, currentTime).
		Count(&count).Error
	return count, err
}

func (u *userRoleDAO) DeleteByBizIDAndID(ctx context.Context, bizID, id int64) error {
	return u.db.WithContext(ctx).
		Where("biz_id = ? AND id = ?", bizID, id).
		Delete(&UserRole{}).Error
}

func (u *userRoleDAO) DeleteByBizIDAndUserIDAndRoleID(ctx context.Context, bizID, userID, roleID int64) error {
	return u.db.WithContext(ctx).
		Where("biz_id = ? AND user_id = ? AND role_id = ?", bizID, userID, roleID).
		Delete(&UserRole{}).Error
}

func (u *userRoleDAO) FindByBizIDAndID(ctx context.Context, bizID, id int64) (UserRole, error) {
	var userRole UserRole
	err := u.db.WithContext(ctx).
		Where("biz_id = ? AND id = ?", bizID, id).
		First(&userRole).Error
	return userRole, err
}

func (u *userRoleDAO) FindByBizIDAndRoleID(ctx context.Context, bizID, roleID int64, offset, limit int) ([]UserRole, error) {
	var userRoles []UserRole
	err := u.db.WithContext(ctx).
		Where("biz_id = ? AND role_id = ?", bizID, roleID).
		Offset(offset).
		Limit(limit).
		Find(&userRoles).Error
	return userRoles, err
}
