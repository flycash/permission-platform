package local

import (
	"context"
	"encoding/json"
	"fmt"

	"gitee.com/flycash/permission-platform/internal/domain"
	repoCache "gitee.com/flycash/permission-platform/internal/repository/cache"
	"github.com/ecodeclub/ecache"
)

const (
	resName = "resource"
	subName = "subject"
	envName = "env"
	number0 = 0
)

type abacAttributeValCache struct {
	cache ecache.Cache
}

func NewAbacAttributeValCache(cache ecache.Cache) repoCache.ABACAttributeValCache {
	return &abacAttributeValCache{
		cache: cache,
	}
}

func (a *abacAttributeValCache) GetAttrResObj(ctx context.Context, bizID, id int64) (domain.ABACObject, error) {
	return a.getAttrObj(ctx, a.key(resName, bizID, id))
}

func (a *abacAttributeValCache) SetAttrResObj(ctx context.Context, objs []domain.ABACObject) error {
	return a.setAttrObj(ctx, resName, objs)
}

func (a *abacAttributeValCache) GetAttrSubObj(ctx context.Context, bizID, id int64) (domain.ABACObject, error) {
	return a.getAttrObj(ctx, a.key(subName, bizID, id))
}

func (a *abacAttributeValCache) SetAttrSubObj(ctx context.Context, objs []domain.ABACObject) error {
	return a.setAttrObj(ctx, subName, objs)
}

func (a *abacAttributeValCache) GetAttrEnvObj(ctx context.Context, bizID int64) (domain.ABACObject, error) {
	return a.getAttrObj(ctx, a.key(envName, bizID, number0))
}

func (a *abacAttributeValCache) SetAttrEnvObj(ctx context.Context, objs []domain.ABACObject) error {
	return a.setAttrObj(ctx, envName, objs)
}

func (a *abacAttributeValCache) getAttrObj(ctx context.Context, key string) (domain.ABACObject, error) {
	val := a.cache.Get(ctx, key)
	if val.Err != nil {
		return domain.ABACObject{}, fmt.Errorf("获取属性值失败: %w", val.Err)
	}
	if val.KeyNotFound() {
		return domain.ABACObject{}, repoCache.ErrKeyNotFound
	}

	var obj domain.ABACObject
	if err := val.JSONScan(&obj); err != nil {
		return domain.ABACObject{}, fmt.Errorf("解析属性值失败: %w", err)
	}
	return obj, nil
}

func (a *abacAttributeValCache) setAttrObj(ctx context.Context, typName string, objs []domain.ABACObject) error {
	for _, obj := range objs {
		vByte, err := json.Marshal(obj)
		if err != nil {
			return fmt.Errorf("序列化失败 %w", err)
		}
		key := a.key(typName, obj.BizID, obj.ID)
		if err := a.cache.Set(ctx, key, string(vByte), defaultTimeout); err != nil {
			return fmt.Errorf("设置属性值失败: %w", err)
		}
	}
	return nil
}

func (a *abacAttributeValCache) key(typName string, bizID, id int64) string {
	switch typName {
	case resName:
		return fmt.Sprintf("abac:attr:%s:%d:%d", resName, bizID, id)
	case subName:
		return fmt.Sprintf("abac:attr:%s:%d:%d", subName, bizID, id)
	case envName:
		return fmt.Sprintf("abac:attr:%s:%d", envName, bizID)
	default:
		return ""
	}
}
