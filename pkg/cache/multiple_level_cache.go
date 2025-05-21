package cache

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"gitee.com/flycash/permission-platform/pkg/bitring"
	"github.com/ecodeclub/ecache"
	eredis "github.com/ecodeclub/ecache/redis"
	"github.com/gotomicro/ego/core/elog"
	"github.com/redis/go-redis/v9"
)

type Entry struct {
	Key        string
	Val        any
	Expiration time.Duration
}

// DataLoader 从数据库加载数据的函数类型
type DataLoader func(ctx context.Context) ([]*Entry, error)

// MultiLevelCache 实现了多级缓存
type MultiLevelCache struct {
	redis                  ecache.Cache     // Redis缓存
	local                  ecache.Cache     // 本地缓存，仅Redis崩溃时使用
	dataLoader             DataLoader       // 从数据库加载数据的函数
	isRedisAvailable       atomic.Bool      // Redis是否可用
	redisCrashDetector     *bitring.BitRing // 错误检测器
	redisHealthCheckPeriod time.Duration
	redisPingTimeout       time.Duration
	mu                     sync.Mutex // 用于保护内部状态更新

	// 定时刷新相关
	refreshTicker            *time.Ticker
	stopRefreshCtx           context.Context
	stopRefreshCtxCancelFunc context.CancelFunc
	localCacheRefreshPeriod  time.Duration

	logger *elog.Component
}

// NewMultiLevelCache 创建一个新的多级缓存
func NewMultiLevelCache(
	rd redis.Cmdable,
	local ecache.Cache,
	dataLoader DataLoader,
	localCacheRefreshPeriod,
	redisPingTimeout,
	redisHealthCheckPeriod time.Duration,
	redisCrashDetector *bitring.BitRing,
) *MultiLevelCache {
	mlc := &MultiLevelCache{
		redis: &ecache.NamespaceCache{
			C:         eredis.NewCache(rd),
			Namespace: "permission-platform:multicluster:",
		},
		local:                   local,
		dataLoader:              dataLoader,
		redisCrashDetector:      redisCrashDetector,
		localCacheRefreshPeriod: localCacheRefreshPeriod,
		redisPingTimeout:        redisPingTimeout,
		redisHealthCheckPeriod:  redisHealthCheckPeriod,
		logger:                  elog.DefaultLogger,
	}

	mlc.stopRefreshCtx, mlc.stopRefreshCtxCancelFunc = context.WithCancel(context.Background())

	// 初始状态假设Redis可用
	mlc.isRedisAvailable.Store(true)

	// 启动Redis健康检查
	go mlc.redisHealthCheck(rd)

	return mlc
}

// redisHealthCheck 定期检查Redis健康状态
func (m *MultiLevelCache) redisHealthCheck(rd redis.Cmdable) {
	ticker := time.NewTicker(m.redisHealthCheckPeriod)
	defer ticker.Stop()
	for range ticker.C {
		if !m.isRedisAvailable.Load() {
			// Redis不可用状态下，检查Redis是否恢复
			ctx, cancel := context.WithTimeout(context.Background(), m.redisPingTimeout)
			// 尝试Ping Redis
			if err := rd.Ping(ctx); err == nil {
				m.handleRedisRecoveryEvent(context.Background())
			}
			cancel()
		}
	}
}

// handleRedisRecoveryEvent 处理Redis恢复事件
func (m *MultiLevelCache) handleRedisRecoveryEvent(ctx context.Context) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 已经恢复，避免重复处理
	if m.isRedisAvailable.Load() {
		return
	}

	// 标记Redis已恢复
	m.isRedisAvailable.Store(true)

	// 停止刷新本地缓存
	m.stopRefreshCtxCancelFunc()

	// 重置错误检测器
	m.redisCrashDetector.Reset()

	// 立即从数据库加载数据到Redis缓存
	if err := m.loadFromDBToCache(ctx, m.redis); err != nil {
		m.logger.Error("从数据库加载数据到Redis失败", elog.FieldErr(err))
	}
}

// loadFromDBToCache 从数据库加载数据到缓存
func (m *MultiLevelCache) loadFromDBToCache(ctx context.Context, c ecache.Cache) error {
	// 从数据库加载数据
	entries, err := m.dataLoader(ctx)
	if err != nil {
		return err
	}
	// 保存到缓存
	for i := range entries {
		err = c.Set(ctx, entries[i].Key, entries[i].Val, entries[i].Expiration)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *MultiLevelCache) Set(ctx context.Context, key string, val any, expiration time.Duration) error {
	if !m.isRedisAvailable.Load() {
		// Redis不可用，写入本地缓存
		return m.local.Set(ctx, key, val, expiration)
	}
	// Redis可用，只写入Redis
	err := m.redis.Set(ctx, key, val, expiration)
	m.redisCrashDetector.Add(err != nil)
	if err != nil && m.redisCrashDetector.IsConditionMet() {
		// Redis检测到崩溃，启动降级流程
		m.handleRedisCrashEvent(ctx)
	}
	return err
}

// handleRedisCrashEvent 处理Redis崩溃事件
func (m *MultiLevelCache) handleRedisCrashEvent(ctx context.Context) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 已经处于不可用状态，避免重复处理
	if !m.isRedisAvailable.Load() {
		return
	}

	// 标记Redis不可用
	m.isRedisAvailable.Store(false)

	// 立即从数据库加载数据到本地缓存
	if err := m.loadFromDBToCache(ctx, m.local); err != nil {
		m.logger.Error("从数据库加载数据到本地缓存失败", elog.FieldErr(err))
	}

	m.stopRefreshCtx, m.stopRefreshCtxCancelFunc = context.WithCancel(context.Background())

	// 启动定时从数据库刷新本地缓存的任务
	//nolint:contextcheck // 忽略
	go m.refreshLocalCache(m.stopRefreshCtx)
}

// refreshLocalCache 定期从数据库刷新本地缓存中的数据
func (m *MultiLevelCache) refreshLocalCache(ctx context.Context) {
	m.refreshTicker = time.NewTicker(m.localCacheRefreshPeriod)
	for {
		select {
		case <-m.refreshTicker.C:
			if err := m.loadFromDBToCache(ctx, m.local); err != nil {
				m.logger.Error("从数据库加载数据到本地缓存失败", elog.FieldErr(err))
			}
		case <-ctx.Done():
			m.refreshTicker.Stop()
			return
		}
	}
}

func (m *MultiLevelCache) Get(ctx context.Context, key string) Value {
	if !m.isRedisAvailable.Load() {
		// Redis不可用，查本地缓存
		return Value(m.local.Get(ctx, key))
	}
	// Redis可用，从Redis获取
	val := m.redis.Get(ctx, key)
	// 检查Redis是否出错（排除KeyNotFound）
	if val.Err != nil && !val.KeyNotFound() {
		m.redisCrashDetector.Add(true)
		if m.redisCrashDetector.IsConditionMet() {
			// Redis崩溃，切换到使用本地缓存
			m.handleRedisCrashEvent(ctx)
		}
	} else {
		// Redis正常响应
		m.redisCrashDetector.Add(false)
	}
	return Value(val)
}
