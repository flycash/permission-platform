package dao

import (
	"gitee.com/flycash/permission-platform/internal/domain"
	"github.com/ecodeclub/ekit/sqlx"
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

// RoleInclusion 角色包含关系表
type RoleInclusion struct {
	ID                int64    `gorm:"primaryKey;autoIncrement;comment:角色包含关系ID'"`
	BizID             int64    `gorm:"type:BIGINT;NOT NULL;uniqueIndex:uk_biz_including_included,priority:1;index:idx_biz_including_role,priority:1;index:idx_biz_included_role,priority:1;comment:'业务ID'"`
	IncludingRoleID   int64    `gorm:"type:BIGINT;NOT NULL;uniqueIndex:uk_biz_including_included,priority:2;index:idx_biz_including_role,priority:2;comment:'包含者角色ID（拥有其他角色权限）'"`
	IncludingRoleType RoleType `gorm:"type:ENUM('system', 'custom', 'temporary');NOT NULL;comment:'包含者角色类型（冗余字段，加速查询）'"`
	IncludingRoleName string   `gorm:"type:VARCHAR(100);NOT NULL;comment:'包含者角色名称（冗余字段，加速查询）'"`
	IncludedRoleID    int64    `gorm:"type:BIGINT;NOT NULL;uniqueIndex:uk_biz_including_included,priority:3;index:idx_biz_included_role,priority:2;comment:'被包含角色ID（权限被包含）'"`
	IncludedRoleType  RoleType `gorm:"type:ENUM('system', 'custom', 'temporary');NOT NULL;comment:'被包含角色类型（冗余字段，加速查询）'"`
	IncludedRoleName  string   `gorm:"type:VARCHAR(100);NOT NULL;comment:'被包含角色名称（冗余字段，加速查询）'"`
	Ctime             int64
	Utime             int64
}

func (RoleInclusion) TableName() string {
	return "role_inclusions"
}

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
