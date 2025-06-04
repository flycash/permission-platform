package cache

import (
	"context"
	"sync/atomic"
	"time"

	eredis "github.com/ecodeclub/ecache/redis"

	"gitee.com/flycash/permission-platform/pkg/bitring"
	"github.com/ecodeclub/ecache"
	"github.com/redis/go-redis/v9"
)

type MultiLevelCacheV1 struct {
	redis                  ecache.Cache // Redis缓存
	local                  ecache.Cache // 本地缓存，仅Redis崩溃时使用
	isRedisAvailable       atomic.Bool  // Redis是否可用
	redisHealthCheckPeriod time.Duration
	redisPingTimeout       time.Duration
	redisCrashDetector     *bitring.BitRing // 错误检测器
}

func NewMultiLevelCacheV1(local ecache.Cache,
	redisHealthCheckPeriod time.Duration,
	redisPingTimeout time.Duration,
	redisCrashDetector *bitring.BitRing,
	redisClient redis.Cmdable,
) *MultiLevelCacheV1 {
	cachev2 := &MultiLevelCacheV1{
		redis: &ecache.NamespaceCache{
			C:         eredis.NewCache(redisClient),
			Namespace: "permission-platform:multiclusterv2:",
		},
		redisHealthCheckPeriod: redisHealthCheckPeriod,
		redisPingTimeout:       redisPingTimeout,
		redisCrashDetector:     redisCrashDetector,
		local:                  local,
	}
	cachev2.isRedisAvailable.Store(true)
	go cachev2.redisHealthCheck(redisClient)
	return cachev2
}

// redisHealthCheck 定期检查Redis健康状态
func (m *MultiLevelCacheV1) redisHealthCheck(rd redis.Cmdable) {
	ticker := time.NewTicker(m.redisHealthCheckPeriod)

	defer ticker.Stop()
	for range ticker.C {
		if !m.isRedisAvailable.Load() {
			// Redis不可用状态下，检查Redis是否恢复
			ctx, cancel := context.WithTimeout(context.Background(), m.redisPingTimeout)
			// 尝试Ping Redis
			if err := rd.Ping(ctx); err == nil {
				m.handleRedisRecoveryEvent()
			}
			cancel()
		}
	}
}

// handleRedisRecoveryEvent 处理Redis恢复事件
func (m *MultiLevelCacheV1) handleRedisRecoveryEvent() {
	// 标记Redis已恢复
	m.isRedisAvailable.CompareAndSwap(false, true)
	m.redisCrashDetector.Reset()
}

func (m *MultiLevelCacheV1) Set(ctx context.Context, key string, val any, expiration time.Duration) error {
	err := m.local.Set(ctx, key, val, expiration)
	if err != nil {
		return err
	}
	// redis不可用就返回
	if !m.isRedisAvailable.Load() {
		return nil
	}
	err = m.redis.Set(ctx, key, val, expiration)
	m.redisCrashDetector.Add(err != nil)
	if err != nil && m.redisCrashDetector.IsConditionMet() {
		// Redis检测到崩溃，启动降级流程
		m.handleRedisCrashEvent()
	}
	return err
}

// handleRedisCrashEvent 处理Redis崩溃事件
func (m *MultiLevelCacheV1) handleRedisCrashEvent() {
	// 标记Redis不可用
	m.isRedisAvailable.CompareAndSwap(true, false)
}

func (m *MultiLevelCacheV1) Get(ctx context.Context, key string) Value {
	if !m.isRedisAvailable.Load() {
		// Redis不可用，查本地缓存
		return m.local.Get(ctx, key)
	}
	// Redis可用，从Redis获取
	val := m.redis.Get(ctx, key)
	// 检查Redis是否出错（排除KeyNotFound）
	if val.Err != nil && !val.KeyNotFound() {
		m.redisCrashDetector.Add(true)
		if m.redisCrashDetector.IsConditionMet() {
			// Redis崩溃，切换到使用本地缓存
			m.handleRedisCrashEvent()
		}
	} else {
		// Redis正常响应
		m.redisCrashDetector.Add(false)
	}
	return val
}
