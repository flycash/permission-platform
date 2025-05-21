package dao

import (
	"context"
	"time"

	"github.com/ego-component/egorm"
	"gorm.io/gorm/clause"
)

// SubjectAttributeValue 主体属性值表模型
type SubjectAttributeValue struct {
	ID        int64  `gorm:"column:id;primaryKey;autoIncrement;"`
	BizID     int64  `gorm:"column:biz_id;uniqueIndex:idx_biz_subject_attr;comment:biz_id + subject_id + attr_id 唯一索引"`
	SubjectID int64  `gorm:"column:subject_id;not null;uniqueIndex:idx_biz_subject_attr;index:idx_subject_id;comment:主体ID，通常是用户ID"`
	AttrDefID int64  `gorm:"column:attr_def_id;not null;uniqueIndex:idx_biz_subject_attr;index:idx_attribute_id;comment:属性定义ID"`
	Value     string `gorm:"column:value;type:text;not null;comment:属性值，取决于 data_type"`
	Ctime     int64  `gorm:"column:ctime;"`
	Utime     int64  `gorm:"column:utime;"`
}

// TableName 指定表名
func (s SubjectAttributeValue) TableName() string {
	return "subject_attribute_values"
}

// SubjectAttributeValueDAO 主体属性值数据访问接口
type SubjectAttributeValueDAO interface {
	// 保存主体属性值，返回ID
	Save(ctx context.Context, value SubjectAttributeValue) (int64, error)
	// 查询单个主体属性值
	First(ctx context.Context, id int64) (SubjectAttributeValue, error)
	// 删除主体属性值
	Del(ctx context.Context, id int64) error
	// 查询主体的所有属性值
	FindBySubject(ctx context.Context, bizID int64, subjectID int64) ([]SubjectAttributeValue, error)
}

type abacSubjectAttributeValueDAO struct {
	db *egorm.Component
}

// NewSubjectAttributeValueDAO 创建主体属性值数据访问对象
func NewSubjectAttributeValueDAO(db *egorm.Component) SubjectAttributeValueDAO {
	return &abacSubjectAttributeValueDAO{db: db}
}

func (s *abacSubjectAttributeValueDAO) Save(ctx context.Context, value SubjectAttributeValue) (int64, error) {
	now := time.Now().UnixMilli()
	value.Ctime = now
	value.Utime = now
	err := s.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "biz_id"}, {Name: "subject_id"}, {Name: "attr_def_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"value", "utime"}),
		}).Create(&value).Error
	return value.ID, err
}

func (s *abacSubjectAttributeValueDAO) First(ctx context.Context, id int64) (SubjectAttributeValue, error) {
	var value SubjectAttributeValue
	err := s.db.WithContext(ctx).
		Where("id = ?", id).
		First(&value).Error
	return value, err
}

func (s *abacSubjectAttributeValueDAO) Del(ctx context.Context, id int64) error {
	return s.db.WithContext(ctx).
		Where("id = ?", id).
		Delete(&SubjectAttributeValue{}).Error
}

func (s *abacSubjectAttributeValueDAO) FindBySubject(ctx context.Context, bizID, subjectID int64) ([]SubjectAttributeValue, error) {
	var values []SubjectAttributeValue
	err := s.db.WithContext(ctx).
		Where("biz_id = ? AND subject_id = ?", bizID, subjectID).
		Find(&values).Error
	return values, err
}
