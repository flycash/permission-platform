package cache

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"gitee.com/flycash/permission-platform/internal/errs"
	"github.com/ecodeclub/ecache"
	eredis "github.com/ecodeclub/ecache/redis"
	"github.com/ecodeclub/ekit"
	"github.com/redis/go-redis/v9"
	"go.uber.org/multierr"
	"golang.org/x/sync/errgroup"
)

type Cluster struct {
	name string
	ecache.Cache
}

func NewCluster(name string, cmd redis.Cmdable) *Cluster {
	return &Cluster{
		name: name,
		Cache: &ecache.NamespaceCache{
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
	g, ctx := errgroup.WithContext(ctx)
	for i := range m.clusters {
		c := m.clusters[i]
		g.Go(func() error {
			err1 := c.Set(ctx, key, val, expiration)
			if err1 != nil {
				mu.Lock()
				err = multierr.Append(err, fmt.Errorf("集群[%s]: %w", c.name, err1))
				mu.Unlock()
				return err1
			}
			return nil
		})
	}
	_ = g.Wait()
	return err
}

func (m *MultipleClusterCache) Get(ctx context.Context, key string) Value {
	var err error
	var valuePtr atomic.Pointer[ecache.Value]
	var needReturn atomic.Bool
	const step = 2
	for i, j := 0, 1; i < len(m.clusters); i, j = i+step, j+step {
		// 一次读取两个
		// 立即返回：找到key对应的Value或者明确地知道KeyNotFound
		// 否则继续向后找
		g, gCtx := errgroup.WithContext(ctx)
		g.Go(func() error {
			val := m.clusters[i].Get(gCtx, key)
			if val.Err == nil || val.KeyNotFound() {
				// 明确地 “已找到” 和 “没找到” 需要直接返回
				needReturn.Store(true)
				valuePtr.Store(&val)
				return nil
			} else {
				err = multierr.Append(err, fmt.Errorf("集群[%s]: %w", m.clusters[i].name, val.Err))
				return val.Err
			}
		})
		if j < len(m.clusters) {
			g.Go(func() error {
				val := m.clusters[j].Get(gCtx, key)
				if val.Err == nil || val.KeyNotFound() {
					// 明确地 “已找到” 和 “没找到” 需要直接返回
					needReturn.Store(true)
					valuePtr.Store(&val)
					return nil
				} else {
					err = multierr.Append(err, fmt.Errorf("集群[%s]: %w", m.clusters[j].name, val.Err))
					return val.Err
				}
			})
		}
		// 等待当前批次查找结束
		_ = g.Wait()
		if needReturn.Load() {
			return Value(*valuePtr.Load())
		}
	}
	if err == nil {
		// 只有m.clusters的长度为0才能走到这
		err = errs.ErrKeyNotExist
	}
	return Value{AnyValue: ekit.AnyValue{Err: err}}
}
