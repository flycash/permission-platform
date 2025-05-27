package redisx

import (
	"context"
	"encoding/json"
	"fmt"
	"gitee.com/flycash/permission-platform/internal/domain"
	repoCache "gitee.com/flycash/permission-platform/internal/repository/cache"
	"gitee.com/flycash/permission-platform/pkg/cache"
	"github.com/ecodeclub/ecache"
	"github.com/ecodeclub/ecache/redis"
	goredis "github.com/redis/go-redis/v9"
	"time"
)

const defaultTimeout = 5 * time.Second

type abacDefCache struct {
	ca cache.Cache
}

func NewAbacDefCache(client goredis.Cmdable) repoCache.ABACDefinitionCache {
	return &abacDefCache{
		ca: &ecache.NamespaceCache{
			C:         redis.NewCache(client),
			Namespace: "abac:def:",
		},
	}
}

func (a *abacDefCache) GetDefinitions(ctx context.Context, bizID int64) (domain.BizAttrDefinition, error) {
	val := a.ca.Get(ctx, a.key(bizID))
	if val.Err != nil {
		return domain.BizAttrDefinition{}, val.Err
	}
	if val.KeyNotFound() {
		return domain.BizAttrDefinition{}, repoCache.ErrKeyNotFound
	}
	var res domain.BizAttrDefinition
	err := val.JSONScan(&res)
	return res, err
}

func (a *abacDefCache) SetDefinitions(ctx context.Context, bizDef domain.BizAttrDefinition) error {
	vByte, err := json.Marshal(bizDef)
	if err != nil {
		return fmt.Errorf("序列化失败 %w", err)
	}
	return a.ca.Set(ctx, a.key(bizDef.BizID), string(vByte), defaultTimeout)
}

func (a *abacDefCache) key(bizID int64) string {
	return fmt.Sprintf("%d", bizID)
}
