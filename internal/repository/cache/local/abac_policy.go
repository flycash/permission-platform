package local

import (
	"context"
	"encoding/json"
	"fmt"
	"gitee.com/flycash/permission-platform/internal/domain"
	"gitee.com/flycash/permission-platform/pkg/cache"
	"time"
)

var defaultTimeOut = 48 * time.Hour

type abacPolicyLocalCache struct {
	cache cache.Cache
}

func (a *abacPolicyLocalCache) GetPolicies(ctx context.Context, bizID int64, permissionIDs []int64) (map[int64]domain.Policy, error) {
	val := a.cache.Get(ctx, a.tableKey(bizID))
	if val.KeyNotFound() {
		return map[int64]domain.Policy{}, nil
	}
	var bizPolicyMap map[int64]domain.Policy
	err := val.JSONScan(&bizPolicyMap)
	if err != nil {
		return nil, err
	}
	resPolicyMap := make(map[int64]domain.Policy,len(permissionIDs))
	for idx := range permissionIDs {
		permissionID := permissionIDs[idx]
		if v,ok := bizPolicyMap[permissionID]; ok {
			resPolicyMap[permissionID] = v
		}
	}
	return resPolicyMap, nil
}

// SetPolicy 本地缓存默认全部替换
func (a *abacPolicyLocalCache) SetPolicy(ctx context.Context, bizID int64, policies []domain.Policy) error {
	policyMap := make(map[int64]domain.Policy)
	for idx := range policies {
		policy := policies[idx]
		policyMap[policy.ID] = policy
	}
	policyByte, err := json.Marshal(policyMap)
	if err != nil {
		return fmt.Errorf("序列化失败: %w", err)
	}
	err = a.cache.Set(ctx, a.tableKey(bizID), string(policyByte), defaultTimeOut)
	return err
}

// DelPolicy 不会使用，删除功能直接使用上面的全量替换即可
func (a *abacPolicyLocalCache) DelPolicy(ctx context.Context, bizID int64, permissionID int64) error {
	return nil
}

func (a *abacPolicyLocalCache) tableKey(bizID int64) string {
	return fmt.Sprintf("abac:policy:%d", bizID)
}
