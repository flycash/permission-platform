package dao

import (
	"context"
	"time"

	"github.com/ego-component/egorm"
	"gorm.io/gorm/clause"
)

// EnvironmentAttributeValue 环境属性表模型
type EnvironmentAttributeValue struct {
	ID        int64  `gorm:"column:id;primaryKey;autoIncrement;"`
	BizID     int64  `gorm:"column:biz_id;uniqueIndex:idx_biz_attribute;comment:业务ID"`
	AttrDefID int64  `gorm:"column:attr_def_id;not null;uniqueIndex:idx_biz_attribute;comment:属性定义ID"`
	Value     string `gorm:"column:value;type:text;comment:属性值，取决于 data_type"`
	Ctime     int64  `gorm:"column:ctime;comment:创建时间"`
	Utime     int64  `gorm:"column:utime;comment:更新时间"`
}

// TableName 指定表名
func (e EnvironmentAttributeValue) TableName() string {
	return "environment_attribute_values"
}

// EnvironmentAttributeDAO 环境属性数据访问接口
type EnvironmentAttributeDAO interface {
	// 保存环境属性，返回ID
	Save(ctx context.Context, attr EnvironmentAttributeValue) (int64, error)
	// 查询单个环境属性
	First(ctx context.Context, id int64) (EnvironmentAttributeValue, error)
	// 通过属性ID和业务ID查询环境属性
	FirstByAttribute(ctx context.Context, bizID int64, attributeID int64) (EnvironmentAttributeValue, error)
	// 删除环境属性
	Del(ctx context.Context, id int64) error
	// 查询业务的所有环境属性
	FindByBiz(ctx context.Context, bizID int64) ([]EnvironmentAttributeValue, error)
}

type environmentAttributeDAO struct {
	db *egorm.Component
}

// NewEnvironmentAttributeDAO 创建环境属性数据访问对象
func NewEnvironmentAttributeDAO(db *egorm.Component) EnvironmentAttributeDAO {
	return &environmentAttributeDAO{db: db}
}

func (e *environmentAttributeDAO) Save(ctx context.Context, attr EnvironmentAttributeValue) (int64, error) {
	now := time.Now().UnixMilli()
	attr.Ctime = now
	attr.Utime = now
	err := e.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "biz_id"}, {Name: "attr_def_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"value", "utime"}),
		}).Create(&attr).Error
	return attr.ID, err
}

func (e *environmentAttributeDAO) First(ctx context.Context, id int64) (EnvironmentAttributeValue, error) {
	var attr EnvironmentAttributeValue
	err := e.db.WithContext(ctx).
		Where("id = ?", id).
		First(&attr).Error
	return attr, err
}

func (e *environmentAttributeDAO) FirstByAttribute(ctx context.Context, bizID, attributeID int64) (EnvironmentAttributeValue, error) {
	var attr EnvironmentAttributeValue
	err := e.db.WithContext(ctx).
		Where("biz_id = ? AND attr_def_id = ?", bizID, attributeID).
		First(&attr).Error
	return attr, err
}

func (e *environmentAttributeDAO) Del(ctx context.Context, id int64) error {
	return e.db.WithContext(ctx).
		Where("id = ?", id).
		Delete(&EnvironmentAttributeValue{}).Error
}

func (e *environmentAttributeDAO) FindByBiz(ctx context.Context, bizID int64) ([]EnvironmentAttributeValue, error) {
	var attrs []EnvironmentAttributeValue
	err := e.db.WithContext(ctx).
		Where("biz_id = ?", bizID).
		Find(&attrs).Error
	return attrs, err
}
