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

// RoleType 角色类型枚举
type RoleType string

const (
	RoleTypeSystem    RoleType = "system"    // 系统角色
	RoleTypeCustom    RoleType = "custom"    // 自定义角色
	RoleTypeTemporary RoleType = "temporary" // 临时角色
)

// Role 角色记录表
type Role struct {
	ID          int64                                `gorm:"primaryKey;autoIncrement;comment:角色ID'"`
	BizID       int64                                `gorm:"type:BIGINT;NOT NULL;index:idx_biz_id;uniqueIndex:uk_biz_type_name,priority:1;comment:'业务ID'"`
	Type        RoleType                             `gorm:"type:ENUM('system', 'custom', 'temporary');NOT NULL;DEFAULT:'custom';index:idx_type;uniqueIndex:uk_biz_type_name,priority:2;index:idx_temporary_validity,priority:1;comment:'角色类型：system(系统角色)、custom(自定义角色)、temporary(临时角色)'"`
	Name        string                               `gorm:"type:VARCHAR(100);NOT NULL;uniqueIndex:uk_biz_type_name,priority:3;comment:'角色名称'"`
	Description string                               `gorm:"type:TEXT;comment:'角色描述'"`
	Metadata    sqlx.JsonColumn[domain.RoleMetadata] `gorm:"type:JSON;comment:'角色元数据，可扩展字段'"`
	StartTime   int64                                `gorm:"NULL;index:idx_temporary_validity,priority:2;comment:'临时角色生效时间'"`
	EndTime     int64                                `gorm:"NULL;index:idx_temporary_validity,priority:3;comment:'临时角色失效时间'"`
	Ctime       int64
	Utime       int64
}

func (Role) TableName() string {
	return "roles"
}

// RoleDAO 角色数据访问接口
type RoleDAO interface {
	// GetByID 根据ID获取角色
	GetByID(ctx context.Context, id int64) (Role, error)
	// FindByBizID 查找特定业务下的角色
	FindByBizID(ctx context.Context, bizID int64, offset, limit int) ([]Role, error)
	// FindByBizIDAndType 查找特定业务下指定类型的角色
	FindByBizIDAndType(ctx context.Context, bizID int64, roleType RoleType, offset, limit int) ([]Role, error)
	// Create 创建角色
	Create(ctx context.Context, role Role) (Role, error)
	// Update 更新角色
	Update(ctx context.Context, role Role) error
	// Delete 删除角色
	Delete(ctx context.Context, id int64) error
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

func (r *roleDAO) GetByID(ctx context.Context, id int64) (Role, error) {
	var role Role
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&role).Error
	return role, err
}

func (r *roleDAO) FindByBizID(ctx context.Context, bizID int64, offset, limit int) ([]Role, error) {
	var roles []Role
	err := r.db.WithContext(ctx).Where("biz_id = ?", bizID).Offset(offset).Limit(limit).Find(&roles).Error
	return roles, err
}

func (r *roleDAO) FindByBizIDAndType(ctx context.Context, bizID int64, roleType RoleType, offset, limit int) ([]Role, error) {
	var roles []Role
	err := r.db.WithContext(ctx).Where("biz_id = ? AND type = ?", bizID, roleType).Offset(offset).Limit(limit).Find(&roles).Error
	return roles, err
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

func (r *roleDAO) Update(ctx context.Context, role Role) error {
	role.Utime = time.Now().UnixMilli()
	return r.db.WithContext(ctx).
		Model(&Role{}).
		Where("id = ?", role.ID).
		Updates(map[string]interface{}{
			"description": role.Description,
			"metadata":    role.Metadata,
			"start_time":  role.StartTime,
			"end_time":    role.EndTime,
			"utime":       role.Utime,
		}).Error
}

func (r *roleDAO) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&Role{}).Error
}
