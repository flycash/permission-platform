package cache

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ecodeclub/ecache"
	eredis "github.com/ecodeclub/ecache/redis"
	"github.com/ecodeclub/ekit"
	"github.com/redis/go-redis/v9"
	"go.uber.org/multierr"
)

type Cluster struct {
	name     string
	instance ecache.Cache
}

func NewCluster(name string, cmd redis.Cmdable) *Cluster {
	return &Cluster{
		name: name,
		instance: &ecache.NamespaceCache{
			C:         eredis.NewCache(cmd),
			Namespace: "permission-platform:multicluster:",
		},
	}
}

type MultipleClusterCache struct {
	clusters []*Cluster
}

func NewMultipleClusterCache(clusters []*Cluster) *MultipleClusterCache {
	return &MultipleClusterCache{
		clusters: clusters,
	}
}

func (m *MultipleClusterCache) Set(ctx context.Context, key string, val any, expiration time.Duration) error {
	var err error
	var mu sync.Mutex
	var wg sync.WaitGroup
	for i := range m.clusters {
		c := m.clusters[i]
		wg.Add(1)
		go func() {
			defer wg.Done()
			err1 := c.instance.Set(ctx, key, val, expiration)
			if err1 != nil {
				mu.Lock()
				err = multierr.Append(err, fmt.Errorf("集群[%s]: %w", c.name, err1))
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	return err
}

func (m *MultipleClusterCache) Get(ctx context.Context, key string) Value {
	var err error
	var mu sync.Mutex
	var needReturn atomic.Bool
	var valuePtr atomic.Pointer[ecache.Value]

	// 集群查询函数
	queryCluster := func(idx int) {
		val := m.clusters[idx].instance.Get(ctx, key)
		if val.Err == nil || val.KeyNotFound() {
			// 明确地 "已找到" 和 "没找到" 需要直接返回
			needReturn.Store(true)
			valuePtr.Store(&val)
		} else {
			mu.Lock()
			err = multierr.Append(err, fmt.Errorf("集群[%s]: %w", m.clusters[idx].name, val.Err))
			mu.Unlock()
		}
	}
	// 遍历集群
	const step = 2
	for i, j := 0, 1; i < len(m.clusters); i, j = i+step, j+step {
		// 一次读取两个
		// 立即返回：找到key对应的Value或者明确地知道KeyNotFound
		// 否则继续向后找
		var wg sync.WaitGroup

		// 第一个集群
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			queryCluster(idx)
		}(i)

		// 第二个集群（如果存在）
		if j < len(m.clusters) {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				queryCluster(idx)
			}(j)
		}

		// 等待当前批次查找结束
		wg.Wait()
		if needReturn.Load() {
			return *valuePtr.Load()
		}
	}
	return Value{AnyValue: ekit.AnyValue{Err: err}}
}
