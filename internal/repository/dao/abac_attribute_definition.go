package dao

import (
	"context"
	"time"

	"gorm.io/gorm/clause"

	"github.com/ego-component/egorm"
)

// AttributeDefinition 属性定义表模型
type AttributeDefinition struct {
	ID             int64  `gorm:"column:id;primaryKey;;autoIncrement;"`
	BizID          int64  `gorm:"column:biz_id;uniqueIndex:idx_biz_id_name;comment:和name组成唯一索引，比如说代表订单组的biz_id"`
	Name           string `gorm:"column:name;size:100;not null;type:varchar(255);uniqueIndex:idx_biz_id_name;comment:属性名称"`
	Description    string `gorm:"column:description;type:text;comment:属性描述"`
	DataType       string `gorm:"column:data_type;type:varchar(255);not null;comment:属性数据类型"`
	EntityType     string `gorm:"column:entity_type;type:enum('subject','resource','environment');not null;comment:属性所属实体类型;index:idx_entity_type"`
	ValidationRule string `gorm:"column:validation_rule;comment:验证规则，正则表达式"`
	Ctime          int64  `gorm:"column:ctime;comment:创建时间"` // 使用毫秒级时间戳
	Utime          int64  `gorm:"column:utime;comment:更新时间"` // 使用毫秒级时间戳
}

// TableName 指定表名
func (AttributeDefinition) TableName() string {
	return "attribute_definitions"
}

type AttributeDefinitionDAO interface {
	// 返回属性定义id
	Save(ctx context.Context, definition AttributeDefinition) (int64, error)
	First(ctx context.Context, bizID int64, id int64) (AttributeDefinition, error)
	Del(ctx context.Context, bizID int64, id int64) error
	// 返回一个bizID所有的属性定义
	Find(ctx context.Context, bizID int64) ([]AttributeDefinition, error)
	FindByIDs(ctx context.Context, ids []int64) (map[int64]AttributeDefinition, error)
}

type attributeDefinitionDAO struct {
	db *egorm.Component
}

func NewAttributeDefinitionDAO(db *egorm.Component) AttributeDefinitionDAO {
	return &attributeDefinitionDAO{db: db}
}

func (a *attributeDefinitionDAO) FindByIDs(ctx context.Context, ids []int64) (map[int64]AttributeDefinition, error) {
	var definitions []AttributeDefinition
	err := a.db.WithContext(ctx).
		Where("id in ?", ids).Find(&definitions).Error
	if err != nil {
		return nil, err
	}
	res := make(map[int64]AttributeDefinition, len(ids))
	for idx := range definitions {
		definition := definitions[idx]
		res[definition.ID] = definition
	}
	return res, nil
}

func (a *attributeDefinitionDAO) Save(ctx context.Context, definition AttributeDefinition) (int64, error) {
	now := time.Now().UnixMilli()
	definition.Ctime = now
	definition.Utime = now
	err := a.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "biz_id"}, {Name: "name"}},
			DoUpdates: clause.AssignmentColumns([]string{"description", "data_type", "entity_type", "validation_rule"}),
		}).Create(&definition).Error
	return definition.ID, err
}

func (a *attributeDefinitionDAO) First(ctx context.Context, bizID, id int64) (AttributeDefinition, error) {
	var definition AttributeDefinition
	err := a.db.WithContext(ctx).Where("id = ? AND biz_id = ?", id, bizID).First(&definition).Error
	return definition, err
}

func (a *attributeDefinitionDAO) Del(ctx context.Context, bizID, id int64) error {
	return a.db.WithContext(ctx).Where("id = ? AND biz_id = ?", id, bizID).
		Delete(&AttributeDefinition{}).Error
}

func (a *attributeDefinitionDAO) Find(ctx context.Context, bizID int64) ([]AttributeDefinition, error) {
	var definitions []AttributeDefinition
	err := a.db.WithContext(ctx).Where("biz_id = ?", bizID).Find(&definitions).Error
	return definitions, err
}
