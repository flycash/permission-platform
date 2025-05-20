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
	m.invokeFunc(ctx, func(ctx context.Context, c *Cluster) error {
		err1 := c.Set(ctx, key, val, expiration)
		if err1 != nil {
			mu.Lock()
			err = multierr.Append(err, fmt.Errorf("集群[%s]: %w", c.name, err1))
			mu.Unlock()
			return err1
		}
		return nil
	})
	return err
}

func (m *MultipleClusterCache) invokeFunc(ctx context.Context, fn func(ctx context.Context, c *Cluster) error) {
	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(len(m.clusters))
	for i := range m.clusters {
		g.Go(func() error {
			return fn(ctx, m.clusters[i])
		})
	}
	_ = g.Wait()
}

func (m *MultipleClusterCache) SetNX(ctx context.Context, key string, val any, expiration time.Duration) (bool, error) {
	var res atomic.Bool
	res.Store(true)
	var err error
	var mu sync.Mutex
	m.invokeFunc(ctx, func(ctx context.Context, c *Cluster) error {
		ok, err1 := c.SetNX(ctx, key, val, expiration)
		if err1 != nil {
			mu.Lock()
			err = multierr.Append(err, fmt.Errorf("集群[%s]: %w", c.name, err1))
			mu.Unlock()
			return err1
		}
		res.Store(ok && res.Load())
		return nil
	})
	return res.Load(), err
}

func (m *MultipleClusterCache) Get(ctx context.Context, key string) Value {
	var err error
	for i, j := 0, 1; i < len(m.clusters); i, j = i+2, j+2 {
		// 一次读取两个
		// 立即返回：找到key对应的Value或者明确地知道KeyNotFound
		// 否则继续向后找
		val := m.clusters[i].Get(ctx, key)
		if val.Err == nil || val.KeyNotFound() {
			return Value(val)
		} else {
			err = multierr.Append(err, fmt.Errorf("集群[%s]: %w", m.clusters[i].name, val.Err))
		}
		if j < len(m.clusters) {
			val = m.clusters[j].Get(ctx, key)
			if val.Err == nil || val.KeyNotFound() {
				return Value(val)
			} else {
				err = multierr.Append(err, fmt.Errorf("集群[%s]: %w", m.clusters[j].name, val.Err))
			}
		}
	}
	if err == nil {
		// 只有m.clusters的长度为0才可能走到这
		err = errs.ErrKeyNotExist
	}
	return Value{AnyValue: ekit.AnyValue{Err: err}}
}

func (m *MultipleClusterCache) Delete(ctx context.Context, key ...string) (int64, error) {
	var res atomic.Int64
	var err error
	var mu sync.Mutex
	m.invokeFunc(ctx, func(ctx context.Context, c *Cluster) error {
		n, err1 := c.Delete(ctx, key...)
		if err1 != nil {
			mu.Lock()
			err = multierr.Append(err, fmt.Errorf("集群[%s]: %w", c.name, err1))
			mu.Unlock()
			return err1
		}
		res.Add(n)
		return nil
	})
	if err != nil {
		return 0, err
	}
	return res.Load() / int64(len(m.clusters)), nil
}
