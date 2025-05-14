package dao

import (
	"context"
	"time"

	"github.com/ego-component/egorm"
)

// BusinessConfig 业务配置表
type BusinessConfig struct {
	ID        int64  `gorm:"primaryKey;autoIncrement;comment:'业务ID'"`
	OwnerID   int64  `gorm:"type:BIGINT;comment:'业务方ID'"`
	OwnerType string `gorm:"type:ENUM('person', 'organization');comment:'业务方类型：person-个人,organization-组织'"`
	Name      string `gorm:"type:VARCHAR(255);NOT NULL;comment:'业务名称'"`
	RateLimit int    `gorm:"type:INT;DEFAULT:1000;comment:'每秒最大请求数'"`
	Token     string `gorm:"type:TXT;NOT NULL;comment:'业务方Token，内部包含bizID'"`
	Ctime     int64
	Utime     int64
}

// TableName 重命名表
func (BusinessConfig) TableName() string {
	return "business_configs"
}

type BusinessConfigDAO interface {
	Create(ctx context.Context, config BusinessConfig) (BusinessConfig, error)
	GetByIDs(ctx context.Context, id []int64) (map[int64]BusinessConfig, error)
	GetByID(ctx context.Context, id int64) (BusinessConfig, error)
	Find(ctx context.Context, offset int, limit int) ([]BusinessConfig, error)
	UpdateToken(ctx context.Context, id int64, token string) error
	Update(ctx context.Context, config BusinessConfig) error
	Delete(ctx context.Context, id int64) error
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

func (b *businessConfigDAO) Create(ctx context.Context, config BusinessConfig) (BusinessConfig, error) {
	now := time.Now().UnixMilli()
	config.Ctime = now
	config.Utime = now
	err := b.db.WithContext(ctx).Create(&config).Error
	return config, err
}

func (b *businessConfigDAO) Find(ctx context.Context, offset, limit int) ([]BusinessConfig, error) {
	var res []BusinessConfig
	err := b.db.WithContext(ctx).Limit(limit).Offset(offset).Find(&res).Error
	return res, err
}

func (b *businessConfigDAO) GetByID(ctx context.Context, id int64) (BusinessConfig, error) {
	var config BusinessConfig
	err := b.db.WithContext(ctx).Where("id = ?", id).First(&config).Error
	return config, err
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

func (b *businessConfigDAO) UpdateToken(ctx context.Context, id int64, token string) error {
	return b.db.WithContext(ctx).
		Model(&BusinessConfig{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"token": token,
			"utime": time.Now().UnixMilli(),
		}).Error
}

// Update 更新业务配置
func (b *businessConfigDAO) Update(ctx context.Context, config BusinessConfig) error {
	config.Utime = time.Now().UnixMilli()
	return b.db.WithContext(ctx).
		Model(&BusinessConfig{}).
		Where("id = ?", config.ID).
		Updates(map[string]any{
			"owner_id":   config.OwnerID,
			"owner_type": config.OwnerType,
			"name":       config.Name,
			"rate_limit": config.RateLimit,
			"utime":      config.Utime,
		}).Error
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
