package dao

import (
	"context"
	"fmt"
	"time"

	"gitee.com/flycash/permission-platform/internal/errs"
	"github.com/ego-component/egorm"
)

// Permission 权限表 RBAC 与 ABAC 共享此表
type Permission struct {
	ID           int64  `gorm:"primaryKey;autoIncrement;comment:'权限ID'"`
	BizID        int64  `gorm:"type:BIGINT;NOT NULL;uniqueIndex:uk_biz_resource_action,priority:1;index:idx_biz_action,priority:1;index:idx_biz_resource_type,priority:1;index:idx_biz_resource_key,priority:1;comment:'业务ID'"`
	Name         string `gorm:"type:VARCHAR(255);NOT NULL;comment:'权限名称'"`
	Description  string `gorm:"type:TEXT;comment:'权限描述'"`
	ResourceID   int64  `gorm:"type:BIGINT;NOT NULL;uniqueIndex:uk_biz_resource_action,priority:2;index:idx_resource_id;comment:'关联的资源ID，创建后不可修改'"`
	ResourceType string `gorm:"type:VARCHAR(255);NOT NULL;index:idx_biz_resource_type,priority:2;comment:'资源类型，冗余字段，加速查询'"`
	ResourceKey  string `gorm:"type:VARCHAR(255);NOT NULL;index:idx_biz_resource_key,priority:2;comment:'资源业务标识符 (如 用户ID, 文档路径)，冗余字段，加速查询'"`
	Action       string `gorm:"type:VARCHAR(255);NOT NULL;NOT NULL;uniqueIndex:uk_biz_resource_action,priority:3;index:idx_biz_action,priority:2;comment:'操作类型'"`
	Metadata     string `gorm:"type:TEXT;comment:'权限元数据，可扩展字段'"`
	Ctime        int64
	Utime        int64
}

func (Permission) TableName() string {
	return "permissions"
}

// PermissionDAO 权限数据访问接口
type PermissionDAO interface {
	Create(ctx context.Context, permission Permission) (Permission, error)

	FindByBizID(ctx context.Context, bizID int64, offset, limit int) ([]Permission, error)
	FindByBizIDAndID(ctx context.Context, bizID, id int64) (Permission, error)

	UpdateByBizIDAndID(ctx context.Context, permission Permission) error

	DeleteByBizIDAndID(ctx context.Context, bizID, id int64) error

	FindPermissions(ctx context.Context, bizID int64, resourceType, resourceKey string, actions []string) ([]Permission, error)
}

// permissionDAO 权限数据访问实现
type permissionDAO struct {
	db *egorm.Component
}

// NewPermissionDAO 创建权限数据访问对象
func NewPermissionDAO(db *egorm.Component) PermissionDAO {
	return &permissionDAO{
		db: db,
	}
}

func (p *permissionDAO) FindPermissions(ctx context.Context, bizID int64, resourceType, resourceKey string, actions []string) ([]Permission, error) {
	var permissions []Permission
	err := p.db.WithContext(ctx).
		Where("biz_id = ? AND resource_key = ? AND resource_type = ? AND action in ?", bizID, resourceKey, resourceType, actions).Find(&permissions).Error
	return permissions, err
}

func (p *permissionDAO) Create(ctx context.Context, permission Permission) (Permission, error) {
	now := time.Now().UnixMilli()
	permission.Ctime = now
	permission.Utime = now
	err := p.db.WithContext(ctx).Create(&permission).Error
	if err != nil {
		if isUniqueConstraintError(err) {
			return Permission{}, fmt.Errorf("%w", errs.ErrPermissionDuplicate)
		}
		return Permission{}, err
	}
	return permission, nil
}

func (p *permissionDAO) FindByBizID(ctx context.Context, bizID int64, offset, limit int) ([]Permission, error) {
	var permissions []Permission
	err := p.db.WithContext(ctx).Where("biz_id = ?", bizID).Offset(offset).Limit(limit).Find(&permissions).Error
	return permissions, err
}

func (p *permissionDAO) FindByBizIDAndID(ctx context.Context, bizID, id int64) (Permission, error) {
	var permission Permission
	err := p.db.WithContext(ctx).Where("biz_id = ? AND id = ?", bizID, id).First(&permission).Error
	return permission, err
}

func (p *permissionDAO) UpdateByBizIDAndID(ctx context.Context, permission Permission) error {
	permission.Utime = time.Now().UnixMilli()
	return p.db.WithContext(ctx).
		Model(&Permission{}).
		Where("biz_id = ? AND id = ?", permission.BizID, permission.ID).
		Updates(map[string]any{
			"name":        permission.Name,
			"description": permission.Description,
			"action":      permission.Action,
			"metadata":    permission.Metadata,
			"utime":       permission.Utime,
		}).Error
}

func (p *permissionDAO) DeleteByBizIDAndID(ctx context.Context, bizID, id int64) error {
	return p.db.WithContext(ctx).Where("biz_id = ? AND id = ?", bizID, id).Delete(&Permission{}).Error
}
