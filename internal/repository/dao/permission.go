package dao

import (
	"gitee.com/flycash/permission-platform/internal/domain"
	"github.com/ecodeclub/ekit/sqlx"
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
