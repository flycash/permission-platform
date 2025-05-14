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

// ResourceDAO 资源数据访问接口
type ResourceDAO interface {
	// GetByID 根据ID获取资源
	GetByID(ctx context.Context, id int64) (Resource, error)
	// FindByBizID 查找特定业务下的资源
	FindByBizID(ctx context.Context, bizID int64, offset, limit int) ([]Resource, error)
	// FindByBizIDAndType 查找特定业务下指定类型的资源
	FindByBizIDAndType(ctx context.Context, bizID int64, resourceType string, offset, limit int) ([]Resource, error)
	// FindByBizIDAndKey 查找特定业务下指定Key的资源
	FindByBizIDAndKey(ctx context.Context, bizID int64, key string) (Resource, error)
	// Create 创建资源
	Create(ctx context.Context, resource Resource) (Resource, error)
	// Update 更新资源
	Update(ctx context.Context, resource Resource) error
	// Delete 删除资源
	Delete(ctx context.Context, id int64) error
}

// resourceDAO 资源数据访问实现
type resourceDAO struct {
	db *egorm.Component
}

// NewResourceDAO 创建资源数据访问对象
func NewResourceDAO(db *egorm.Component) ResourceDAO {
	return &resourceDAO{
		db: db,
	}
}

func (r *resourceDAO) GetByID(ctx context.Context, id int64) (Resource, error) {
	var resource Resource
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&resource).Error
	return resource, err
}

func (r *resourceDAO) FindByBizID(ctx context.Context, bizID int64, offset, limit int) ([]Resource, error) {
	var resources []Resource
	err := r.db.WithContext(ctx).Where("biz_id = ?", bizID).Offset(offset).Limit(limit).Find(&resources).Error
	return resources, err
}

func (r *resourceDAO) FindByBizIDAndType(ctx context.Context, bizID int64, resourceType string, offset, limit int) ([]Resource, error) {
	var resources []Resource
	err := r.db.WithContext(ctx).Where("biz_id = ? AND type = ?", bizID, resourceType).Offset(offset).Limit(limit).Find(&resources).Error
	return resources, err
}

func (r *resourceDAO) FindByBizIDAndKey(ctx context.Context, bizID int64, key string) (Resource, error) {
	var resource Resource
	err := r.db.WithContext(ctx).Where("biz_id = ? AND `key` = ?", bizID, key).First(&resource).Error
	return resource, err
}

func (r *resourceDAO) Create(ctx context.Context, resource Resource) (Resource, error) {
	now := time.Now().UnixMilli()
	resource.Ctime = now
	resource.Utime = now
	err := r.db.WithContext(ctx).Create(&resource).Error
	if err != nil {
		if isUniqueConstraintError(err) {
			return Resource{}, fmt.Errorf("%w", errs.ErrResourceDuplicate)
		}
		return Resource{}, err
	}
	return resource, nil
}

func (r *resourceDAO) Update(ctx context.Context, resource Resource) error {
	resource.Utime = time.Now().UnixMilli()
	return r.db.WithContext(ctx).
		Model(&Resource{}).
		Where("id = ?", resource.ID).
		Updates(map[string]interface{}{
			"name":        resource.Name,
			"description": resource.Description,
			"metadata":    resource.Metadata,
			"utime":       resource.Utime,
		}).Error
}

func (r *resourceDAO) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&Resource{}).Error
}
