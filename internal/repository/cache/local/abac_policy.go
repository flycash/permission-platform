package local

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"gitee.com/flycash/permission-platform/internal/repository/cache"
	"github.com/ecodeclub/ecache"

	"gitee.com/flycash/permission-platform/internal/domain"
)

var defaultTimeOut = 48 * time.Hour

type abacPolicyLocalCache struct {
	cache ecache.Cache
}

func NewAbacPolicy(ca ecache.Cache) cache.ABACPolicyCache {
	return &abacPolicyLocalCache{
		cache: ca,
	}
}

func (a *abacPolicyLocalCache) GetPolicies(ctx context.Context, bizID int64) ([]domain.Policy, error) {
	key := a.tableKey(bizID)
	val := a.cache.Get(ctx, key)
	if val.Err != nil {
		return nil, val.Err
	}
	if val.KeyNotFound() {
		return nil, cache.ErrKeyNotFound
	}
	var policies []domain.Policy
	err := val.JSONScan(&policies)
	if err != nil {
		return nil, err
	}
	return policies, nil
}

// SetPolicy 本地缓存默认全部替换
func (a *abacPolicyLocalCache) SetPolicy(ctx context.Context, bizID int64, policies []domain.Policy) error {
	key := a.tableKey(bizID)
	policyByte, err := json.Marshal(policies)
	if err != nil {
		return err
	}
	err = a.cache.Set(ctx, key, string(policyByte), defaultTimeOut)
	if err != nil {
		return err
	}
	return nil
}

// DelPolicy 不会使用，删除功能直接使用上面的全量替换即可
func (a *abacPolicyLocalCache) DelPolicy(_ context.Context, _, _ int64) error {
	return nil
}

func (a *abacPolicyLocalCache) tableKey(bizID int64) string {
	return fmt.Sprintf("abac:policy:%d", bizID)
}
