package abac

import (
	"context"
	"encoding/json"
	"fmt"
	"gitee.com/flycash/permission-platform/internal/repository"
	"gitee.com/flycash/permission-platform/internal/repository/cache"
	"github.com/ego-component/eetcd"
	"time"
)

type PolicyCron struct {
	client *eetcd.Component
	repo   repository.PolicyRepo
	cache  cache.ABACPolicyCache
}

const (
	hotPolicyName  = "hotPolicy"
	defaultTimeout = 5 * time.Second
)
func NewPolicyCron(client *eetcd.Component,repo repository.PolicyRepo,ca cache.ABACPolicyCache) *PolicyCron {
	return &PolicyCron{
		client: client,
		repo:   repo,
		cache:  ca,
	}
}

// 定时任务
func (p *PolicyCron) Run(ctx context.Context) error {
	nctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	resp, err := p.client.Get(nctx, hotPolicyName)
	cancel()
	if err != nil {
		return fmt.Errorf("从etcd中获取hotkey失败 %w", err)
	}
	if len(resp.Kvs) == 0 {
		return fmt.Errorf("热点key键不存在")
	}
	kvs := resp.Kvs[0]
	var bizIDs []int64
	err = json.Unmarshal(kvs.Value, &bizIDs)
	if err != nil {
		return fmt.Errorf("序列化失败 %w", err)
	}
	for _, bizID := range bizIDs {
		loopctx,loopcancel := context.WithTimeout(ctx,defaultTimeout)
		err = p.oneloop(loopctx, bizID)
		loopcancel()
		if err != nil {
			return fmt.Errorf("设置本地缓存失败 %w", err)
		}
	}
	return nil
}


func (p *PolicyCron) oneloop(ctx context.Context, bizID int64) error {
	policies,err := p.repo.FindBizPolicies(ctx, bizID)
	if err != nil {
		return err
	}
	return p.cache.SetPolicy(ctx, bizID, policies)
}
