package redisx

import (
	"context"
	"encoding/json"
	"fmt"
	"gitee.com/flycash/permission-platform/internal/domain"
	repoCache "gitee.com/flycash/permission-platform/internal/repository/cache"
	goredis "github.com/redis/go-redis/v9"
)

const (
	resName = "resource"
	subName = "subject"
	envName = "env"
	number0 = 0
)

type abacAttributeValCache struct {
	client goredis.Cmdable
}

func (a *abacAttributeValCache) GetAttrResObj(ctx context.Context, bizID int64, id int64) (domain.ABACObject, error) {
	return a.getAttrObj(ctx,a.key(resName, bizID, id))
}

func (a *abacAttributeValCache) SetAttrResObj(ctx context.Context, objs []domain.ABACObject) error {
	return a.setAttrObj(ctx,resName,objs)
}

func (a *abacAttributeValCache) GetAttrSubObj(ctx context.Context, bizID int64, id int64) (domain.ABACObject, error) {
	return a.getAttrObj(ctx,a.key(subName, bizID, id))
}

func (a *abacAttributeValCache) SetAttrSubObj(ctx context.Context, objs []domain.ABACObject) error {
	return a.setAttrObj(ctx,subName,objs)
}

func (a *abacAttributeValCache) GetAttrEnvObj(ctx context.Context, bizID int64) (domain.ABACObject, error) {
	return a.getAttrObj(ctx,a.key(envName,bizID,number0))
}

func (a *abacAttributeValCache) SetAttrEnvObj(ctx context.Context, objs []domain.ABACObject) error {
	return a.setAttrObj(ctx,envName,objs)
}

func NewAbacAttributeValCache(client goredis.Cmdable) repoCache.ABACAttributeValCache {
	return &abacAttributeValCache{
		client: client,
	}
}

func (a *abacAttributeValCache) getAttrObj(ctx context.Context, key string) (domain.ABACObject, error) {
	val, err := a.client.Get(ctx,key).Result()
	if err != nil {
		return domain.ABACObject{}, fmt.Errorf("批量获取属性值失败: %w", err)
	}
	var obj domain.ABACObject
	if err := json.Unmarshal([]byte(val), &obj); err != nil {
		return domain.ABACObject{}, fmt.Errorf("解析属性值失败: %w", err)
	}

	return obj, nil
}

func (a *abacAttributeValCache) setAttrObj(ctx context.Context, typName string , objs []domain.ABACObject) error {
	pipe := a.client.Pipeline()
	for idx := range objs {
		obj := objs[idx]
		vByte, err := json.Marshal(obj)
		if err != nil {
			return fmt.Errorf("序列化失败 %w", err)
		}
		key := a.key(typName,obj.BizID,obj.ID)
		pipe.Set(ctx, key, string(vByte), defaultTimeout)
	}
	_, err := pipe.Exec(ctx)
	return err
}

func (a *abacAttributeValCache) key(typName string, bizID int64,id int64) string {
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