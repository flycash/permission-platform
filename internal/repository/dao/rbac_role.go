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
