package dao

import (
	"context"
	"fmt"
	"time"

	"gitee.com/flycash/permission-platform/internal/errs"
	"github.com/ego-component/egorm"
	"gorm.io/gorm/clause"
)

// BusinessConfig 业务配置表
type BusinessConfig struct {
	ID        int64  `gorm:"primaryKey;autoIncrement;comment:'业务ID'"`
	OwnerID   int64  `gorm:"type:BIGINT;comment:'业务方ID'"`
	OwnerType string `gorm:"type:ENUM('person', 'organization');comment:'业务方类型：person-个人,organization-组织'"`
	Name      string `gorm:"type:VARCHAR(100);NOT NULL;comment:'业务名称'"`
	RateLimit int    `gorm:"type:INT;DEFAULT:1000;comment:'每秒最大请求数'"`
	Token     string `gorm:"type:TXT;NOT NULL;comment:'业务方Token，内部包含uid也就是上方的ownerID'"`
	Ctime     int64
	Utime     int64
}

// TableName 重命名表
func (BusinessConfig) TableName() string {
	return "business_configs"
}

type BusinessConfigDAO interface {
	GetByIDs(ctx context.Context, id []int64) (map[int64]BusinessConfig, error)
	GetByID(ctx context.Context, id int64) (BusinessConfig, error)
	Delete(ctx context.Context, id int64) error
	SaveConfig(ctx context.Context, config BusinessConfig) (BusinessConfig, error)
	Find(ctx context.Context, offset int, limit int) ([]BusinessConfig, error)
}

// Implementation of the BusinessConfigDAO interface
type businessConfigDAO struct {
	db *egorm.Component
}

// NewBusinessConfigDAO 创建一个新的BusinessConfigDAO实例
func NewBusinessConfigDAO(db *egorm.Component) BusinessConfigDAO {
	return &businessConfigDAO{
		db: db,
	}
}

func (b *businessConfigDAO) Find(ctx context.Context, offset, limit int) ([]BusinessConfig, error) {
	var res []BusinessConfig
	err := b.db.WithContext(ctx).Limit(limit).Offset(offset).Find(&res).Error
	return res, err
}

func (b *businessConfigDAO) GetByID(ctx context.Context, id int64) (BusinessConfig, error) {
	var config BusinessConfig

	// 根据ID查询业务配置
	err := b.db.WithContext(ctx).Where("id = ?", id).First(&config).Error
	if err != nil {
		return BusinessConfig{}, err
	}

	return config, nil
}

// GetByIDs 根据ID获取业务配置信息
func (b *businessConfigDAO) GetByIDs(ctx context.Context, ids []int64) (map[int64]BusinessConfig, error) {
	var configs []BusinessConfig
	// 根据ID查询业务配置
	err := b.db.WithContext(ctx).Where("id in (?)", ids).Find(&configs).Error
	if err != nil {
		return nil, err
	}
	configMap := make(map[int64]BusinessConfig, len(ids))
	for idx := range configs {
		config := configs[idx]
		configMap[config.ID] = config
	}
	return configMap, nil
}

// Delete 根据ID删除业务配置
func (b *businessConfigDAO) Delete(ctx context.Context, id int64) error {
	// 执行删除操作
	result := b.db.WithContext(ctx).
		Where("id = ?", id).
		Delete(&BusinessConfig{})
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// SaveConfig 保存业务配置
func (b *businessConfigDAO) SaveConfig(ctx context.Context, config BusinessConfig) (BusinessConfig, error) {
	now := time.Now().UnixMilli()
	config.Ctime = now
	config.Utime = now
	// 使用upsert语句，如果记录存在则更新，不存在则插入
	db := b.db.WithContext(ctx)
	result := db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "id"}}, // 根据ID判断冲突
		DoUpdates: clause.AssignmentColumns([]string{
			"owner_id",
			"owner_type",
			"rate_limit",
			"utime",
		}), // 只更新指定的非空列
	}).Create(&config)
	if result.Error != nil {
		if isUniqueConstraintError(result.Error) {
			return BusinessConfig{}, fmt.Errorf("%w", errs.ErrBusinessConfigDuplicate)
		}
		return BusinessConfig{}, result.Error
	}
	return config, nil
}
