package redisx

import (
	"context"
	"encoding/json"
	"fmt"

	"gitee.com/flycash/permission-platform/internal/domain"
	"gitee.com/flycash/permission-platform/internal/repository/cache"
	"github.com/redis/go-redis/v9"
)

type abacPolicy struct {
	redisClient redis.Cmdable
}

func NewAbacPolicy(redisClient redis.Cmdable) cache.ABACPolicyCache {
	return &abacPolicy{
		redisClient: redisClient,
	}
}

func (a *abacPolicy) DelPolicy(ctx context.Context, bizID, policyID int64) error {
	return a.redisClient.HDel(ctx, a.tableKey(bizID), fmt.Sprintf("%d", policyID)).Err()
}

func (a *abacPolicy) GetPolicies(ctx context.Context, bizID int64) ([]domain.Policy, error) {
	v, err := a.redisClient.Get(ctx, a.tableKey(bizID)).Result()
	if err != nil {
		return nil, err
	}
	policies := make([]domain.Policy, 0)
	err = json.Unmarshal([]byte(v), &policies)
	return policies, err
}

func (a *abacPolicy) SetPolicy(ctx context.Context, bizID int64, policies []domain.Policy) error {
	policyByte, err := json.Marshal(policies)
	if err != nil {
		return err
	}
	return a.redisClient.Set(ctx, a.tableKey(bizID), string(policyByte), defaultTimeout).Err()
}

func (a *abacPolicy) tableKey(bizID int64) string {
	return fmt.Sprintf("abac:policy:%d", bizID)
}
