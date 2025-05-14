package dao

import (
	"context"
	"time"

	"github.com/ego-component/egorm"
	"gorm.io/gorm/clause"
)

// ResourceAttributeValue 资源属性值表模型
type ResourceAttributeValue struct {
	ID          int64  `gorm:"column:id;primaryKey;autoIncrement;"`
	BizID       int64  `gorm:"column:biz_id;uniqueIndex:idx_biz_resource_attr;comment:biz_id + resource_key + attr_id 唯一索引"`
	ResourceID  int64  `gorm:"column:resource_id;not null;uniqueIndex:idx_biz_resource_attr;index:idx_resource_id;comment:资源ID"`
	AttributeID int64  `gorm:"column:attribute_id;not null;uniqueIndex:idx_biz_resource_attr;index:idx_attribute_id;comment:属性定义ID"`
	Value       string `gorm:"column:value;type:text;not null;comment:属性值，取决于 data_type"`
	Ctime       int64  `gorm:"column:ctime;"`
	Utime       int64  `gorm:"column:utime;"`
}

// TableName 指定表名
func (r ResourceAttributeValue) TableName() string {
	return "resource_attribute_values"
}

// ResourceAttributeValueDAO 资源属性值数据访问接口
type ResourceAttributeValueDAO interface {
	// 保存资源属性值，返回ID
	Save(ctx context.Context, value ResourceAttributeValue) (int64, error)
	// 查询单个资源属性值
	First(ctx context.Context, id int64) (ResourceAttributeValue, error)
	// 删除资源属性值
	Del(ctx context.Context, id int64) error
	// 查询资源的所有属性值
	FindByResource(ctx context.Context, bizID int64, resourceID int64) ([]ResourceAttributeValue, error)
	// 根据属性ID查询所有资源属性值
	FindByAttribute(ctx context.Context, bizID int64, attributeID int64) ([]ResourceAttributeValue, error)
}

type resourceAttributeValueDAO struct {
	db *egorm.Component
}

// NewResourceAttributeValueDAO 创建资源属性值数据访问对象
func NewResourceAttributeValueDAO(db *egorm.Component) ResourceAttributeValueDAO {
	return &resourceAttributeValueDAO{db: db}
}

func (r *resourceAttributeValueDAO) Save(ctx context.Context, value ResourceAttributeValue) (int64, error) {
	now := time.Now().UnixMilli()
	value.Ctime = now
	value.Utime = now
	err := r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "biz_id"}, {Name: "resource_id"}, {Name: "attribute_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"value", "utime"}),
		}).Create(&value).Error
	return value.ID, err
}

func (r *resourceAttributeValueDAO) First(ctx context.Context, id int64) (ResourceAttributeValue, error) {
	var value ResourceAttributeValue
	err := r.db.WithContext(ctx).
		Where("id = ?", id).
		First(&value).Error
	return value, err
}

func (r *resourceAttributeValueDAO) Del(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).
		Where("id = ?", id).
		Delete(&ResourceAttributeValue{}).Error
}

func (r *resourceAttributeValueDAO) FindByResource(ctx context.Context, bizID, resourceID int64) ([]ResourceAttributeValue, error) {
	var values []ResourceAttributeValue
	err := r.db.WithContext(ctx).
		Where("biz_id = ? AND resource_id = ?", bizID, resourceID).
		Find(&values).Error
	return values, err
}

func (r *resourceAttributeValueDAO) FindByAttribute(ctx context.Context, bizID, attributeID int64) ([]ResourceAttributeValue, error) {
	var values []ResourceAttributeValue
	err := r.db.WithContext(ctx).
		Where("biz_id = ? AND attribute_id = ?", bizID, attributeID).
		Find(&values).Error
	return values, err
}
