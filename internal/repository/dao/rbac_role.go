package dao

import (
	"context"
	"fmt"
	"time"

	"gitee.com/flycash/permission-platform/internal/domain"
	"gitee.com/flycash/permission-platform/internal/errs"
	"github.com/ecodeclub/ekit/sqlx"
	"github.com/ego-component/egorm"
)

// Role 角色记录表
type Role struct {
	ID          int64                                `gorm:"primaryKey;autoIncrement;comment:角色ID'"`
	BizID       int64                                `gorm:"type:BIGINT;NOT NULL;index:idx_biz_id;uniqueIndex:uk_biz_type_name,priority:1;comment:'业务ID'"`
	Type        string                               `gorm:"type:VARCHAR(255);NOT NULL;index:idx_role_type;uniqueIndex:uk_biz_type_name,priority:2;comment:'角色类（被冗余，创建后不可修改）'"`
	Name        string                               `gorm:"type:VARCHAR(255);NOT NULL;uniqueIndex:uk_biz_type_name,priority:3;comment:'角色名称（被冗余，创建后不可修改）'"`
	Description string                               `gorm:"type:TEXT;comment:'角色描述'"`
	Metadata    sqlx.JsonColumn[domain.RoleMetadata] `gorm:"type:JSON;comment:'角色元数据，可扩展字段'"`
	Ctime       int64
	Utime       int64
}

func (Role) TableName() string {
	return "roles"
}

// RoleDAO 角色数据访问接口
type RoleDAO interface {
	Create(ctx context.Context, role Role) (Role, error)

	FindByBizID(ctx context.Context, bizID int64, offset, limit int) ([]Role, error)
	CountByBizID(ctx context.Context, bizID int64) (int64, error)

	FindByBizIDAndID(ctx context.Context, bizID, id int64) (Role, error)

	FindByBizIDAndType(ctx context.Context, bizID int64, roleType string, offset, limit int) ([]Role, error)
	CountByBizIDAndType(ctx context.Context, bizID int64, roleType string) (int64, error)

	UpdateByBizIDAndID(ctx context.Context, role Role) error

	DeleteByBizIDAndID(ctx context.Context, bizID, id int64) error
}

// roleDAO 角色数据访问实现
type roleDAO struct {
	db *egorm.Component
}

// NewRoleDAO 创建角色数据访问对象
func NewRoleDAO(db *egorm.Component) RoleDAO {
	return &roleDAO{
		db: db,
	}
}

func (r *roleDAO) Create(ctx context.Context, role Role) (Role, error) {
	now := time.Now().UnixMilli()
	role.Ctime = now
	role.Utime = now
	err := r.db.WithContext(ctx).Create(&role).Error
	if isUniqueConstraintError(err) {
		return Role{}, fmt.Errorf("%w", errs.ErrRoleDuplicate)
	}
	return role, err
}

func (r *roleDAO) FindByBizID(ctx context.Context, bizID int64, offset, limit int) ([]Role, error) {
	var roles []Role
	err := r.db.WithContext(ctx).Where("biz_id = ?", bizID).Offset(offset).Limit(limit).Find(&roles).Error
	return roles, err
}

func (r *roleDAO) FindByBizIDAndID(ctx context.Context, bizID, id int64) (Role, error) {
	var role Role
	err := r.db.WithContext(ctx).Where("biz_id = ? AND id = ?", bizID, id).First(&role).Error
	return role, err
}

func (r *roleDAO) FindByBizIDAndType(ctx context.Context, bizID int64, roleType string, offset, limit int) ([]Role, error) {
	var roles []Role
	err := r.db.WithContext(ctx).Where("biz_id = ? AND type = ?", bizID, roleType).Offset(offset).Limit(limit).Find(&roles).Error
	return roles, err
}

func (r *roleDAO) UpdateByBizIDAndID(ctx context.Context, role Role) error {
	role.Utime = time.Now().UnixMilli()
	return r.db.WithContext(ctx).
		Model(&Role{}).
		Where("biz_id = ? AND id = ?", role.BizID, role.ID).
		Updates(map[string]interface{}{
			"description": role.Description,
			"metadata":    role.Metadata,
			"utime":       role.Utime,
		}).Error
}

func (r *roleDAO) DeleteByBizIDAndID(ctx context.Context, bizID, id int64) error {
	return r.db.WithContext(ctx).Where("biz_id = ? AND id = ?", bizID, id).Delete(&Role{}).Error
}

func (r *roleDAO) CountByBizID(ctx context.Context, bizID int64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&Role{}).Where("biz_id = ?", bizID).Count(&count).Error
	return count, err
}

func (r *roleDAO) CountByBizIDAndType(ctx context.Context, bizID int64, roleType string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&Role{}).Where("biz_id = ? AND type = ?", bizID, roleType).Count(&count).Error
	return count, err
}
