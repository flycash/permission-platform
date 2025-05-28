package repository

import (
	"context"
	"errors"

	"gitee.com/flycash/permission-platform/internal/repository/cache"
	"github.com/gotomicro/ego/core/elog"

	"gitee.com/flycash/permission-platform/internal/domain"
	"gitee.com/flycash/permission-platform/internal/repository/dao"
)

type AttributeDefinitionRepository interface {
	// 返回属性定义id
	Save(ctx context.Context, bizID int64, definition domain.AttributeDefinition) (int64, error)
	First(ctx context.Context, bizID int64, id int64) (domain.AttributeDefinition, error)
	Del(ctx context.Context, bizID int64, id int64) error
	// 返回一个bizID所有的属性定义
	Find(ctx context.Context, bizID int64) (domain.BizAttrDefinition, error)
}

type attributeDefinitionRepository struct {
	definitionDao dao.AttributeDefinitionDAO
	localCache    cache.ABACDefinitionCache
	redisCache    cache.ABACDefinitionCache
	logger        *elog.Component
}

func NewAttributeDefinitionRepository(
	definitionDao dao.AttributeDefinitionDAO,
	localCache cache.ABACDefinitionCache,
	redisCache cache.ABACDefinitionCache,
) AttributeDefinitionRepository {
	return &attributeDefinitionRepository{
		definitionDao: definitionDao,
		localCache:    localCache,
		redisCache:    redisCache,
		logger:        elog.DefaultLogger,
	}
}

// 将domain.AttributeDefinition转换为dao.Definition
func toDaoDefinition(bizID int64, def domain.AttributeDefinition) dao.AttributeDefinition {
	return dao.AttributeDefinition{
		ID:             def.ID,
		BizID:          bizID,
		Name:           def.Name,
		Description:    def.Description,
		DataType:       def.DataType.String(),
		EntityType:     def.EntityType.String(),
		ValidationRule: def.ValidationRule,
		Ctime:          def.Ctime,
		Utime:          def.Utime,
	}
}

// 将dao.AttributeDefinition转换为domain.Definition
func toDomainDefinition(daoDef dao.AttributeDefinition) domain.AttributeDefinition {
	return domain.AttributeDefinition{
		ID:             daoDef.ID,
		Name:           daoDef.Name,
		Description:    daoDef.Description,
		DataType:       domain.DataType(daoDef.DataType),
		EntityType:     domain.EntityType(daoDef.EntityType),
		ValidationRule: daoDef.ValidationRule,
		Ctime:          daoDef.Ctime,
		Utime:          daoDef.Utime,
	}
}

func (a *attributeDefinitionRepository) Save(ctx context.Context, bizID int64, definition domain.AttributeDefinition) (int64, error) {
	daoDef := toDaoDefinition(bizID, definition)
	id, err := a.definitionDao.Save(ctx, daoDef)
	if err != nil {
		return 0, err
	}
	err = a.setCache(ctx, bizID)
	if err != nil {
		a.logger.Error("保存到缓存失败",
			elog.FieldErr(err),
			elog.Int64("bizID", bizID),
		)
	}
	return id, nil
}

func (a *attributeDefinitionRepository) First(ctx context.Context, bizID, id int64) (domain.AttributeDefinition, error) {
	daoDef, err := a.definitionDao.First(ctx, bizID, id)
	if err != nil {
		return domain.AttributeDefinition{}, err
	}
	return toDomainDefinition(daoDef), nil
}

func (a *attributeDefinitionRepository) Del(ctx context.Context, bizID, id int64) error {
	return a.definitionDao.Del(ctx, bizID, id)
}

func (a *attributeDefinitionRepository) Find(ctx context.Context, bizID int64) (domain.BizAttrDefinition, error) {
	// Try local cache first
	defs, localCacheErr := a.localCache.GetDefinitions(ctx, bizID)
	if localCacheErr == nil {
		return defs, nil
	}
	if !errors.Is(localCacheErr, cache.ErrKeyNotFound) {
		a.logger.Error("从本地缓存获取数据失败",
			elog.FieldErr(localCacheErr), elog.Int64("bizID", bizID))
	}

	// Try Redis cache if local cache misses
	defs, redisCacheErr := a.redisCache.GetDefinitions(ctx, bizID)
	if redisCacheErr == nil {
		// Update local cache with Redis data
		if err := a.localCache.SetDefinitions(ctx, defs); err != nil {
			a.logger.Error("更新本地缓存失败",
				elog.FieldErr(err), elog.Int64("bizID", bizID))
		}
		return defs, nil
	}
	if !errors.Is(redisCacheErr, cache.ErrKeyNotFound) {
		a.logger.Error("从redis获取数据失败",
			elog.FieldErr(redisCacheErr), elog.Int64("bizID", bizID))
	}

	// Get from database if both caches miss
	defs, err := a.findByDB(ctx, bizID)
	if err != nil {
		return domain.BizAttrDefinition{}, err
	}

	// Update both caches
	if err := a.redisCache.SetDefinitions(ctx, defs); err != nil {
		a.logger.Error("更新redis缓存失败",
			elog.FieldErr(err), elog.Int64("bizID", bizID))
	}

	if err := a.localCache.SetDefinitions(ctx, defs); err != nil {
		a.logger.Error("更新本地缓存失败",
			elog.FieldErr(err), elog.Int64("bizID", bizID))
	}

	return defs, nil
}

func (a *attributeDefinitionRepository) findByDB(ctx context.Context, bizID int64) (domain.BizAttrDefinition, error) {
	daos, err := a.definitionDao.Find(ctx, bizID)
	if err != nil {
		return domain.BizAttrDefinition{}, err
	}
	bizDef := domain.BizAttrDefinition{
		BizID:   bizID,
		AllDefs: make(map[int64]domain.AttributeDefinition, len(daos)),
	}
	for _, daoDef := range daos {
		def := toDomainDefinition(daoDef)
		switch daoDef.EntityType {
		case domain.SubjectType.String():
			bizDef.SubjectAttrDefs = append(bizDef.SubjectAttrDefs, def)
		case domain.ResourceType.String():
			bizDef.ResourceAttrDefs = append(bizDef.ResourceAttrDefs, def)
		case domain.EnvironmentType.String():
			bizDef.EnvironmentAttrDefs = append(bizDef.EnvironmentAttrDefs, def)
		}
		bizDef.AllDefs[def.ID] = def
	}
	return bizDef, nil
}

func (a *attributeDefinitionRepository) setCache(ctx context.Context, bizID int64) error {
	bizAttrDef, err := a.findByDB(ctx, bizID)
	if err != nil {
		return err
	}
	return a.redisCache.SetDefinitions(ctx, bizAttrDef)
}
