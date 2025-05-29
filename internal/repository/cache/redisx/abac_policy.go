package redisx

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"gitee.com/flycash/permission-platform/internal/domain"
	"gitee.com/flycash/permission-platform/internal/repository/cache"
	"github.com/ecodeclub/ekit/slice"
	"github.com/redis/go-redis/v9"
)

type abacPolicy struct {
	redisClient redis.Cmdable
}

func (a *abacPolicy) DelPolicy(ctx context.Context, bizID int64, policyID int64) error {
	return a.redisClient.HDel(ctx, a.tableKey(bizID), fmt.Sprintf("%d", policyID)).Err()
}

func (a *abacPolicy) GetPolicies(ctx context.Context, bizID int64, policyIDs []int64) (map[int64]domain.Policy, error) {
	permissionStrs := slice.Map(policyIDs, func(idx int, src int64) string {
		return fmt.Sprintf("%d", src)
	})
	policies := make(map[int64]domain.Policy, len(permissionStrs))
	vals, err := a.redisClient.HMGet(ctx, a.tableKey(bizID), permissionStrs...).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return policies, nil
		}
		return nil, err
	}
	for idx := range vals {
		val := vals[idx]
		if val != nil {
			policyStr, ok := val.(string)
			if ok {
				var policy domain.Policy
				err = json.Unmarshal([]byte(policyStr), &policy)
				if err != nil {
					return nil, fmt.Errorf("序列化失败 %w", err)
				}
				policies[policy.ID] = policy
			}
		}
	}
	return policies, nil
}

func (a *abacPolicy) SetPolicy(ctx context.Context, bizID int64, policies []domain.Policy) error {
	policyStrs := slice.Map(policies, func(idx int, src domain.Policy) string {
		v, _ := json.Marshal(src)
		return string(v)
	})
	hashmap := make(map[string]any, len(policies))
	for idx := range policies {
		policy := policies[idx]
		policyStr := policyStrs[idx]
		hashmap[fmt.Sprintf("%d", policy.ID)] = policyStr
	}
	return a.redisClient.HMSet(ctx, a.tableKey(bizID), hashmap).Err()
}

func (a *abacPolicy) GetPermissionPolicy(ctx context.Context, bizID int64, permissionIDs []int64) ([]domain.Policy, error) {
	return []domain.Policy{}, nil
}

func (a *abacPolicy) SetPermissionPolicy(ctx context.Context, bizID int64, reqs []cache.SetPermissionPolicyReq) error {
	return nil
}

func (a *abacPolicy) tableKey(bizID int64) string {
	return fmt.Sprintf("abac:policy:%d", bizID)
}
