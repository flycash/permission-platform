package dao

import (
	"context"
	"time"

	"gitee.com/flycash/permission-platform/internal/domain"
	"github.com/ecodeclub/ekit/sqlx"
	"github.com/ego-component/egorm"
)

type ActionType string

const (
	ActionTypeCreate  ActionType = "create"
	ActionTypeRead    ActionType = "read"
	ActionTypeUpdate  ActionType = "update"
	ActionTypeDelete  ActionType = "delete"
	ActionTypeExecute ActionType = "execute"
	ActionTypeExport  ActionType = "export"
	ActionTypeImport  ActionType = "import"
)

// Permission 权限表 RBAC 与 ABAC 共享此表
type Permission struct {
	ID           int64                                      `gorm:"primaryKey;autoIncrement;comment:'权限ID'"`
	BizID        int64                                      `gorm:"type:BIGINT;NOT NULL;uniqueIndex:uk_biz_resource_action,priority:1;index:idx_biz_action,priority:1;index:idx_biz_resource_type,priority:1;index:idx_biz_resource_key,priority:1;comment:'业务ID'"`
	Name         string                                     `gorm:"type:VARCHAR(100);NOT NULL;comment:'权限名称'"`
	Description  string                                     `gorm:"type:TEXT;comment:'权限描述'"`
	ResourceID   int64                                      `gorm:"type:BIGINT;NOT NULL;uniqueIndex:uk_biz_resource_action,priority:2;index:idx_resource_id;comment:'关联的资源ID'"`
	ResourceType string                                     `gorm:"type:VARCHAR(100);NOT NULL;index:idx_biz_resource_type,priority:2;comment:'资源类型，冗余字段，加速查询'"`
	ResourceKey  string                                     `gorm:"type:VARCHAR(255);NOT NULL;index:idx_biz_resource_key,priority:2;comment:'资源业务标识符 (如 用户ID, 文档路径)，冗余字段，加速查询'"`
	Action       ActionType                                 `gorm:"type:ENUM('create', 'read', 'update', 'delete', 'execute', 'export', 'import');NOT NULL;uniqueIndex:uk_biz_resource_action,priority:3;index:idx_biz_action,priority:2;comment:'操作类型'"`
	Metadata     sqlx.JsonColumn[domain.PermissionMetadata] `gorm:"type:JSON;comment:'权限元数据，可扩展字段'"`
	Ctime        int64                                      `gorm:"-"`
	Utime        int64                                      `gorm:"-"`
}

func (Permission) TableName() string {
	return "permissions"
}

// PermissionDAO 权限数据访问接口
type PermissionDAO interface {
	// GetByID 根据ID获取权限
	GetByID(ctx context.Context, id int64) (Permission, error)
	// GetByIDs 根据多个ID批量获取权限
	GetByIDs(ctx context.Context, ids []int64) (map[int64]Permission, error)
	// FindByBizID 查找特定业务下的权限
	FindByBizID(ctx context.Context, bizID int64, offset, limit int) ([]Permission, error)
	// FindByResourceID 查找特定资源的权限
	FindByResourceID(ctx context.Context, resourceID int64) ([]Permission, error)
	// FindByBizIDAndResourceType 查找特定业务下指定资源类型的权限
	FindByBizIDAndResourceType(ctx context.Context, bizID int64, resourceType string, offset, limit int) ([]Permission, error)
	// FindByBizIDAndResourceKey 查找特定业务下指定资源Key的权限
	FindByBizIDAndResourceKey(ctx context.Context, bizID int64, resourceKey string, offset, limit int) ([]Permission, error)
	// FindByBizIDAndAction 查找特定业务下指定操作类型的权限
	FindByBizIDAndAction(ctx context.Context, bizID int64, action ActionType, offset, limit int) ([]Permission, error)
	// FindByBizIDResourceIDAndAction 根据业务ID、资源ID和操作类型查找权限
	FindByBizIDResourceIDAndAction(ctx context.Context, bizID int64, resourceID int64, action ActionType) (Permission, error)
	// Create 创建权限
	Create(ctx context.Context, permission Permission) (Permission, error)
	// Update 更新权限
	Update(ctx context.Context, permission Permission) error
	// Delete 删除权限
	Delete(ctx context.Context, id int64) error
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

func (p *permissionDAO) GetByID(ctx context.Context, id int64) (Permission, error) {
	var permission Permission
	err := p.db.WithContext(ctx).Where("id = ?", id).First(&permission).Error
	return permission, err
}

func (p *permissionDAO) GetByIDs(ctx context.Context, ids []int64) (map[int64]Permission, error) {
	var permissions []Permission
	err := p.db.WithContext(ctx).Where("id IN (?)", ids).Find(&permissions).Error
	if err != nil {
		return nil, err
	}

	result := make(map[int64]Permission, len(permissions))
	for _, permission := range permissions {
		result[permission.ID] = permission
	}
	return result, nil
}

func (p *permissionDAO) FindByBizID(ctx context.Context, bizID int64, offset, limit int) ([]Permission, error) {
	var permissions []Permission
	err := p.db.WithContext(ctx).Where("biz_id = ?", bizID).Offset(offset).Limit(limit).Find(&permissions).Error
	return permissions, err
}

func (p *permissionDAO) FindByResourceID(ctx context.Context, resourceID int64) ([]Permission, error) {
	var permissions []Permission
	err := p.db.WithContext(ctx).Where("resource_id = ?", resourceID).Find(&permissions).Error
	return permissions, err
}

func (p *permissionDAO) FindByBizIDAndResourceType(ctx context.Context, bizID int64, resourceType string, offset, limit int) ([]Permission, error) {
	var permissions []Permission
	err := p.db.WithContext(ctx).Where("biz_id = ? AND resource_type = ?", bizID, resourceType).Offset(offset).Limit(limit).Find(&permissions).Error
	return permissions, err
}

func (p *permissionDAO) FindByBizIDAndResourceKey(ctx context.Context, bizID int64, resourceKey string, offset, limit int) ([]Permission, error) {
	var permissions []Permission
	err := p.db.WithContext(ctx).Where("biz_id = ? AND resource_key = ?", bizID, resourceKey).Offset(offset).Limit(limit).Find(&permissions).Error
	return permissions, err
}

func (p *permissionDAO) FindByBizIDAndAction(ctx context.Context, bizID int64, action ActionType, offset, limit int) ([]Permission, error) {
	var permissions []Permission
	err := p.db.WithContext(ctx).Where("biz_id = ? AND action = ?", bizID, action).Offset(offset).Limit(limit).Find(&permissions).Error
	return permissions, err
}

func (p *permissionDAO) FindByBizIDResourceIDAndAction(ctx context.Context, bizID int64, resourceID int64, action ActionType) (Permission, error) {
	var permission Permission
	err := p.db.WithContext(ctx).Where("biz_id = ? AND resource_id = ? AND action = ?", bizID, resourceID, action).First(&permission).Error
	return permission, err
}

func (p *permissionDAO) Create(ctx context.Context, permission Permission) (Permission, error) {
	now := time.Now().UnixMilli()
	permission.Ctime = now
	permission.Utime = now
	err := p.db.WithContext(ctx).Create(&permission).Error
	return permission, err
}

func (p *permissionDAO) Update(ctx context.Context, permission Permission) error {
	permission.Utime = time.Now().UnixMilli()
	return p.db.WithContext(ctx).
		Model(&Permission{}).
		Where("id = ?", permission.ID).
		Updates(map[string]interface{}{
			"name":        permission.Name,
			"description": permission.Description,
			"metadata":    permission.Metadata,
			"utime":       permission.Utime,
		}).Error
}

func (p *permissionDAO) Delete(ctx context.Context, id int64) error {
	return p.db.WithContext(ctx).Where("id = ?", id).Delete(&Permission{}).Error
}
