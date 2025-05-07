package dao

import (
	"gitee.com/flycash/permission-platform/internal/domain"
	"github.com/ecodeclub/ekit/sqlx"
)

// Resource 资源表 RBAC 与 ABAC 共享此表
type Resource struct {
	ID          int64                                    `gorm:"primaryKey;autoIncrement;comment:资源ID'"`
	BizID       int64                                    `gorm:"type:BIGINT;NOT NULL;uniqueIndex:uk_biz_type_key,priority:1;index:idx_biz_type,priority:1;index:idx_biz_key,priority:1;comment:'业务ID'"`
	Type        string                                   `gorm:"type:VARCHAR(100);NOT NULL;uniqueIndex:uk_biz_type_key,priority:2;index:idx_biz_type,priority:2;comment:'资源类型，创建后不允许修改'"`
	Key         string                                   `gorm:"type:VARCHAR(255);NOT NULL;uniqueIndex:uk_biz_type_key,priority:3;index:idx_biz_key,priority:2;comment:'资源业务标识符 (如 用户ID, 文档路径)，创建后不允许修改'"`
	Name        string                                   `gorm:"type:VARCHAR(255);NOT NULL;comment:'资源名称'"`
	Description string                                   `gorm:"type:TEXT;comment:'资源描述'"`
	Metadata    sqlx.JsonColumn[domain.ResourceMetadata] `gorm:"type:JSON;comment:'资源元数据'"`
	Ctime       int64
	Utime       int64
}

func (Resource) TableName() string {
	return "resources"
}
