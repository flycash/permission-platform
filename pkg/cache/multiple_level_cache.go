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
	maxLoadRetries           int           // 加载数据的最大重试次数
	retryInterval            time.Duration // 重试间隔

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
		maxLoadRetries:          3,
		retryInterval:           time.Second,
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
	for {
		select {
		case <-ticker.C:
			ctx, cancel := context.WithTimeout(context.Background(), m.redisPingTimeout)
			err := rd.Ping(ctx).Err()
			cancel()

			if !m.isRedisAvailable.Load() && err == nil {
				// Redis恢复了
				m.handleRedisRecoveryEvent(context.Background())
			} else if m.isRedisAvailable.Load() && err != nil {
				// Redis可能崩溃了，记录错误
				m.redisCrashDetector.Add(true)
				if m.redisCrashDetector.IsConditionMet() {
					// 确认Redis崩溃
					m.handleRedisCrashEvent(context.Background())
				}
			} else if m.isRedisAvailable.Load() {
				// Redis正常，重置错误计数
				m.redisCrashDetector.Add(false)
			}
		case <-m.stopRefreshCtx.Done():
			return
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
	m.stopRefreshCtx, m.stopRefreshCtxCancelFunc = context.WithCancel(context.Background())

	// 重置错误检测器
	m.redisCrashDetector.Reset()

	// 立即从数据库加载数据到Redis缓存
	if err := m.loadFromDBToCacheWithRetry(ctx, m.redis); err != nil {
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
	var firstErr error
	for i := range entries {
		err = c.Set(ctx, entries[i].Key, entries[i].Val, entries[i].Expiration)
		if err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

// loadFromDBToCacheWithRetry 带重试的从数据库加载数据到缓存
func (m *MultiLevelCache) loadFromDBToCacheWithRetry(ctx context.Context, c ecache.Cache) error {
	var err error
	for i := 0; i < m.maxLoadRetries; i++ {
		err = m.loadFromDBToCache(ctx, c)
		if err == nil {
			return nil
		}

		// 最后一次尝试失败则直接返回错误
		if i == m.maxLoadRetries-1 {
			return err
		}

		// 等待一段时间后重试
		select {
		case <-time.After(m.retryInterval):
			// 继续下一次重试
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return err
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
	if err := m.loadFromDBToCacheWithRetry(ctx, m.local); err != nil {
		m.logger.Error("从数据库加载数据到本地缓存失败", elog.FieldErr(err))
	}

	// 创建新的上下文用于刷新任务
	m.stopRefreshCtxCancelFunc()
	m.stopRefreshCtx, m.stopRefreshCtxCancelFunc = context.WithCancel(context.Background())

	// 启动定时从数据库刷新本地缓存的任务
	go m.refreshLocalCache(m.stopRefreshCtx)
}

// refreshLocalCache 定期从数据库刷新本地缓存中的数据
func (m *MultiLevelCache) refreshLocalCache(ctx context.Context) {
	m.refreshTicker = time.NewTicker(m.localCacheRefreshPeriod)
	defer m.refreshTicker.Stop()

	for {
		select {
		case <-m.refreshTicker.C:
			if err := m.loadFromDBToCacheWithRetry(ctx, m.local); err != nil {
				m.logger.Error("从数据库加载数据到本地缓存失败", elog.FieldErr(err))
			}
		case <-ctx.Done():
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
			// 从本地缓存获取
			return Value(m.local.Get(ctx, key))
		}
	} else {
		// Redis正常响应
		m.redisCrashDetector.Add(false)
	}
	return Value(val)
}
