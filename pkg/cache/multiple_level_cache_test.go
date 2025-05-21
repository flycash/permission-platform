package cache_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"gitee.com/flycash/permission-platform/internal/errs"
	"gitee.com/flycash/permission-platform/pkg/bitring"
	"gitee.com/flycash/permission-platform/pkg/cache"
	cachemocks "gitee.com/flycash/permission-platform/pkg/cache/mocks"
	"github.com/ecodeclub/ecache"
	"github.com/ecodeclub/ekit"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestMultiLevelCache_Set_RedisAvailable(t *testing.T) {
	testCases := []struct {
		name       string
		mock       func(ctrl *gomock.Controller) (redis.Cmdable, ecache.Cache, ecache.Cache, cache.DataLoader)
		key        string
		val        any
		expiration time.Duration
		wantErr    error
	}{
		{
			name: "Redis可用，写入成功",
			mock: func(ctrl *gomock.Controller) (redis.Cmdable, ecache.Cache, ecache.Cache, cache.DataLoader) {
				redisCmdable := cachemocks.NewMockCmdable(ctrl)

				redisCache := cachemocks.NewMockCache(ctrl)
				redisCache.EXPECT().Set(gomock.Any(), "key1", "val1", time.Minute).Return(nil)

				localCache := cachemocks.NewMockCache(ctrl)
				// 本地缓存不应该被调用

				dataLoader := func(ctx context.Context) ([]*cache.Entry, error) {
					return nil, nil
				}

				return redisCmdable, redisCache, localCache, dataLoader
			},
			key:        "key1",
			val:        "val1",
			expiration: time.Minute,
			wantErr:    nil,
		},
		{
			name: "Redis可用，写入失败但不足以触发故障",
			mock: func(ctrl *gomock.Controller) (redis.Cmdable, ecache.Cache, ecache.Cache, cache.DataLoader) {
				redisCmdable := cachemocks.NewMockCmdable(ctrl)

				redisErr := errors.New("redis error")
				redisCache := cachemocks.NewMockCache(ctrl)
				redisCache.EXPECT().Set(gomock.Any(), "key1", "val1", time.Minute).Return(redisErr)

				localCache := cachemocks.NewMockCache(ctrl)
				// 本地缓存不应该被调用

				dataLoader := func(ctx context.Context) ([]*cache.Entry, error) {
					return nil, nil
				}

				return redisCmdable, redisCache, localCache, dataLoader
			},
			key:        "key1",
			val:        "val1",
			expiration: time.Minute,
			wantErr:    errors.New("redis error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			_, redisCache, localCache, dataLoader := tc.mock(ctrl)

			// 创建一个不会触发的错误检测器 (1/3 的错误率触发)
			crashDetector := bitring.New(3, 3)

			// 使用 NewMultiLevelCache 函数创建实例
			c := cache.NewMultiLevelCache(
				cachemocks.NewMockCmdable(ctrl),
				localCache,
				dataLoader,
				time.Hour,   // localCacheRefreshPeriod
				time.Second, // redisPingTimeout
				time.Hour,   // redisHealthCheckPeriod
				crashDetector,
			)

			// 替换 redis 字段为测试用的 mock
			c.SetRedisCache(redisCache)

			err := c.Set(context.Background(), tc.key, tc.val, tc.expiration)

			if tc.wantErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.wantErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMultiLevelCache_Set_RedisUnavailable(t *testing.T) {
	testCases := []struct {
		name       string
		mock       func(ctrl *gomock.Controller) (redis.Cmdable, ecache.Cache, ecache.Cache, cache.DataLoader)
		key        string
		val        any
		expiration time.Duration
		wantErr    error
	}{
		{
			name: "Redis不可用，写入本地缓存",
			mock: func(ctrl *gomock.Controller) (redis.Cmdable, ecache.Cache, ecache.Cache, cache.DataLoader) {
				redisCmdable := cachemocks.NewMockCmdable(ctrl)

				redisCache := cachemocks.NewMockCache(ctrl)
				// Redis缓存不应该被调用

				localCache := cachemocks.NewMockCache(ctrl)
				localCache.EXPECT().Set(gomock.Any(), "key1", "val1", time.Minute).Return(nil)

				dataLoader := func(ctx context.Context) ([]*cache.Entry, error) {
					return nil, nil
				}

				return redisCmdable, redisCache, localCache, dataLoader
			},
			key:        "key1",
			val:        "val1",
			expiration: time.Minute,
			wantErr:    nil,
		},
		{
			name: "Redis不可用，写入本地缓存失败",
			mock: func(ctrl *gomock.Controller) (redis.Cmdable, ecache.Cache, ecache.Cache, cache.DataLoader) {
				redisCmdable := cachemocks.NewMockCmdable(ctrl)

				redisCache := cachemocks.NewMockCache(ctrl)
				// Redis缓存不应该被调用

				localErr := errors.New("local cache error")
				localCache := cachemocks.NewMockCache(ctrl)
				localCache.EXPECT().Set(gomock.Any(), "key1", "val1", time.Minute).Return(localErr)

				dataLoader := func(ctx context.Context) ([]*cache.Entry, error) {
					return nil, nil
				}

				return redisCmdable, redisCache, localCache, dataLoader
			},
			key:        "key1",
			val:        "val1",
			expiration: time.Minute,
			wantErr:    errors.New("local cache error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			redisCmdable, redisCache, localCache, dataLoader := tc.mock(ctrl)

			cache := &cache.MultiLevelCache{
				Redis:                   redisCache,
				Local:                   localCache,
				DataLoader:              dataLoader,
				RedisCrashDetector:      bitring.New(3, 1), // 不重要，因为Redis已标记为不可用
				RedisHealthCheckPeriod:  time.Hour,
				RedisPingTimeout:        time.Second,
				LocalCacheRefreshPeriod: time.Hour,
			}

			// 标记Redis不可用
			cache.IsRedisAvailable.Store(false)

			err := cache.Set(context.Background(), tc.key, tc.val, tc.expiration)

			if tc.wantErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.wantErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMultiLevelCache_Get_RedisAvailable(t *testing.T) {
	testCases := []struct {
		name      string
		mock      func(ctrl *gomock.Controller) (redis.Cmdable, ecache.Cache, ecache.Cache, cache.DataLoader)
		key       string
		wantValue cache.Value
		wantErr   error
	}{
		{
			name: "Redis可用，成功获取值",
			mock: func(ctrl *gomock.Controller) (redis.Cmdable, ecache.Cache, ecache.Cache, cache.DataLoader) {
				redisCmdable := cachemocks.NewMockCmdable(ctrl)

				redisCache := cachemocks.NewMockCache(ctrl)
				redisCache.EXPECT().Get(gomock.Any(), "key1").Return(ecache.Value{AnyValue: ekit.AnyValue{Val: "value1"}})

				localCache := cachemocks.NewMockCache(ctrl)
				// 本地缓存不应该被调用

				dataLoader := func(ctx context.Context) ([]*cache.Entry, error) {
					return nil, nil
				}

				return redisCmdable, redisCache, localCache, dataLoader
			},
			key:       "key1",
			wantValue: cache.Value{AnyValue: ekit.AnyValue{Val: "value1"}},
			wantErr:   nil,
		},
		{
			name: "Redis可用，键不存在",
			mock: func(ctrl *gomock.Controller) (redis.Cmdable, ecache.Cache, ecache.Cache, cache.DataLoader) {
				redisCmdable := cachemocks.NewMockCmdable(ctrl)

				redisCache := cachemocks.NewMockCache(ctrl)
				redisCache.EXPECT().Get(gomock.Any(), "key1").Return(ecache.Value{AnyValue: ekit.AnyValue{Err: errs.ErrKeyNotExist}})

				localCache := cachemocks.NewMockCache(ctrl)
				// 本地缓存不应该被调用

				dataLoader := func(ctx context.Context) ([]*cache.Entry, error) {
					return nil, nil
				}

				return redisCmdable, redisCache, localCache, dataLoader
			},
			key:       "key1",
			wantValue: cache.Value{AnyValue: ekit.AnyValue{Err: errs.ErrKeyNotExist}},
			wantErr:   errs.ErrKeyNotExist,
		},
		{
			name: "Redis可用，返回错误但不足以触发故障",
			mock: func(ctrl *gomock.Controller) (redis.Cmdable, ecache.Cache, ecache.Cache, cache.DataLoader) {
				redisCmdable := cachemocks.NewMockCmdable(ctrl)

				redisErr := errors.New("redis error")
				redisCache := cachemocks.NewMockCache(ctrl)
				redisCache.EXPECT().Get(gomock.Any(), "key1").Return(ecache.Value{AnyValue: ekit.AnyValue{Err: redisErr}})

				localCache := cachemocks.NewMockCache(ctrl)
				// 本地缓存不应该被调用

				dataLoader := func(ctx context.Context) ([]*cache.Entry, error) {
					return nil, nil
				}

				return redisCmdable, redisCache, localCache, dataLoader
			},
			key:       "key1",
			wantValue: cache.Value{AnyValue: ekit.AnyValue{Err: errors.New("redis error")}},
			wantErr:   errors.New("redis error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			redisCmdable, redisCache, localCache, dataLoader := tc.mock(ctrl)

			// 创建一个不会触发的错误检测器 (1/3 的错误率触发)
			crashDetector := bitring.New(3, 3)

			cache := &cache.MultiLevelCache{
				Redis:                   redisCache,
				Local:                   localCache,
				DataLoader:              dataLoader,
				RedisCrashDetector:      crashDetector,
				RedisHealthCheckPeriod:  time.Hour, // 设置较长时间，避免测试期间触发健康检查
				RedisPingTimeout:        time.Second,
				LocalCacheRefreshPeriod: time.Hour,
			}

			// 标记Redis可用
			cache.IsRedisAvailable.Store(true)

			res := cache.Get(context.Background(), tc.key)

			if tc.wantErr != nil {
				assert.Error(t, res.Err)
				if tc.wantErr == errs.ErrKeyNotExist {
					assert.True(t, res.KeyNotFound())
				} else {
					assert.Contains(t, res.Err.Error(), tc.wantErr.Error())
				}
			} else {
				assert.NoError(t, res.Err)
				assert.Equal(t, tc.wantValue.Val, res.Val)
			}
		})
	}
}

func TestMultiLevelCache_Get_RedisUnavailable(t *testing.T) {
	testCases := []struct {
		name      string
		mock      func(ctrl *gomock.Controller) (redis.Cmdable, ecache.Cache, ecache.Cache, cache.DataLoader)
		key       string
		wantValue cache.Value
		wantErr   error
	}{
		{
			name: "Redis不可用，从本地缓存获取值",
			mock: func(ctrl *gomock.Controller) (redis.Cmdable, ecache.Cache, ecache.Cache, cache.DataLoader) {
				redisCmdable := cachemocks.NewMockCmdable(ctrl)

				redisCache := cachemocks.NewMockCache(ctrl)
				// Redis缓存不应该被调用

				localCache := cachemocks.NewMockCache(ctrl)
				localCache.EXPECT().Get(gomock.Any(), "key1").Return(ecache.Value{AnyValue: ekit.AnyValue{Val: "local_value"}})

				dataLoader := func(ctx context.Context) ([]*cache.Entry, error) {
					return nil, nil
				}

				return redisCmdable, redisCache, localCache, dataLoader
			},
			key:       "key1",
			wantValue: cache.Value{AnyValue: ekit.AnyValue{Val: "local_value"}},
			wantErr:   nil,
		},
		{
			name: "Redis不可用，本地缓存不存在键",
			mock: func(ctrl *gomock.Controller) (redis.Cmdable, ecache.Cache, ecache.Cache, cache.DataLoader) {
				redisCmdable := cachemocks.NewMockCmdable(ctrl)

				redisCache := cachemocks.NewMockCache(ctrl)
				// Redis缓存不应该被调用

				localCache := cachemocks.NewMockCache(ctrl)
				localCache.EXPECT().Get(gomock.Any(), "key1").Return(ecache.Value{AnyValue: ekit.AnyValue{Err: errs.ErrKeyNotExist}})

				dataLoader := func(ctx context.Context) ([]*cache.Entry, error) {
					return nil, nil
				}

				return redisCmdable, redisCache, localCache, dataLoader
			},
			key:       "key1",
			wantValue: cache.Value{AnyValue: ekit.AnyValue{Err: errs.ErrKeyNotExist}},
			wantErr:   errs.ErrKeyNotExist,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			redisCmdable, redisCache, localCache, dataLoader := tc.mock(ctrl)

			cache := &cache.MultiLevelCache{
				Redis:                   redisCache,
				Local:                   localCache,
				DataLoader:              dataLoader,
				RedisCrashDetector:      bitring.New(3, 1), // 不重要，因为Redis已标记为不可用
				RedisHealthCheckPeriod:  time.Hour,
				RedisPingTimeout:        time.Second,
				LocalCacheRefreshPeriod: time.Hour,
			}

			// 标记Redis不可用
			cache.IsRedisAvailable.Store(false)

			res := cache.Get(context.Background(), tc.key)

			if tc.wantErr != nil {
				assert.Error(t, res.Err)
				if tc.wantErr == errs.ErrKeyNotExist {
					assert.True(t, res.KeyNotFound())
				} else {
					assert.Contains(t, res.Err.Error(), tc.wantErr.Error())
				}
			} else {
				assert.NoError(t, res.Err)
				assert.Equal(t, tc.wantValue.Val, res.Val)
			}
		})
	}
}

func TestMultiLevelCache_Redis_Failover(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// 模拟Redis命令接口
	redisCmdable := cachemocks.NewMockCmdable(ctrl)

	// 设置Redis缓存的mock
	redisCache := cachemocks.NewMockCache(ctrl)
	// Set操作第一次成功，第二次失败
	redisCache.EXPECT().Set(gomock.Any(), "key1", "val1", time.Minute).Return(nil)
	redisError := errors.New("redis connection lost")
	redisCache.EXPECT().Set(gomock.Any(), "key2", "val2", time.Minute).Return(redisError)

	// Get操作将失败
	redisCache.EXPECT().Get(gomock.Any(), "key2").Return(ecache.Value{AnyValue: ekit.AnyValue{Err: redisError}})

	// 模拟本地缓存
	localCache := cachemocks.NewMockCache(ctrl)

	// 预期故障转移后的操作
	// 以下操作将在Redis崩溃检测器触发后执行:

	// 1. 加载数据到本地缓存
	entries := []*cache.Entry{
		{Key: "key1", Val: "val1", Expiration: time.Minute},
		{Key: "key2", Val: "val2", Expiration: time.Minute},
	}

	dataLoaderCalled := false
	dataLoader := func(ctx context.Context) ([]*cache.Entry, error) {
		dataLoaderCalled = true
		return entries, nil
	}

	// 2. 数据加载到本地缓存
	localCache.EXPECT().Set(gomock.Any(), "key1", "val1", time.Minute).Return(nil)
	localCache.EXPECT().Set(gomock.Any(), "key2", "val2", time.Minute).Return(nil)

	// 3. 从本地缓存获取
	localCache.EXPECT().Get(gomock.Any(), "key2").Return(ecache.Value{AnyValue: ekit.AnyValue{Val: "val2"}})

	// 创建一个BitRing，设置为连续 1 个错误就触发
	crashDetector := bitring.New(1, 1)

	// 创建缓存实例
	cache := &cache.MultiLevelCache{
		Redis:                    redisCache,
		Local:                    localCache,
		DataLoader:               dataLoader,
		RedisCrashDetector:       crashDetector,
		RedisHealthCheckPeriod:   time.Hour,
		RedisPingTimeout:         time.Second,
		LocalCacheRefreshPeriod:  time.Hour,
		MaxLoadRetries:           3,
		RetryInterval:            time.Millisecond,
		StopRefreshCtx:           context.Background(),
		StopRefreshCtxCancelFunc: func() {},
	}

	// 标记Redis初始可用
	cache.IsRedisAvailable.Store(true)

	// 第一次Set操作正常
	err := cache.Set(context.Background(), "key1", "val1", time.Minute)
	assert.NoError(t, err)

	// 第二次Set操作失败，触发故障转移
	err = cache.Set(context.Background(), "key2", "val2", time.Minute)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "redis connection lost")

	// 验证数据加载被调用
	assert.True(t, dataLoaderCalled)

	// Get操作应该从本地缓存获取
	res := cache.Get(context.Background(), "key2")
	assert.NoError(t, res.Err)
	assert.Equal(t, "val2", res.Val)

	// 验证Redis已被标记为不可用
	assert.False(t, cache.IsRedisAvailable.Load())
}

func TestMultiLevelCache_LoadFromDBWithRetry(t *testing.T) {
	testCases := []struct {
		name          string
		mockLoader    func(ctx context.Context) ([]*cache.Entry, error)
		mockCache     func(ctrl *gomock.Controller) ecache.Cache
		maxRetries    int
		expectedErr   error
		expectedCalls int
	}{
		{
			name: "第一次尝试成功",
			mockLoader: func(ctx context.Context) ([]*cache.Entry, error) {
				return []*cache.Entry{
					{Key: "key1", Val: "val1", Expiration: time.Minute},
				}, nil
			},
			mockCache: func(ctrl *gomock.Controller) ecache.Cache {
				cache := cachemocks.NewMockCache(ctrl)
				cache.EXPECT().Set(gomock.Any(), "key1", "val1", time.Minute).Return(nil)
				return cache
			},
			maxRetries:    3,
			expectedErr:   nil,
			expectedCalls: 1,
		},
		{
			name: "三次尝试都失败",
			mockLoader: func(ctx context.Context) ([]*cache.Entry, error) {
				return nil, errors.New("database error")
			},
			mockCache: func(ctrl *gomock.Controller) ecache.Cache {
				return cachemocks.NewMockCache(ctrl)
			},
			maxRetries:    3,
			expectedErr:   errors.New("database error"),
			expectedCalls: 3,
		},
		{
			name: "第二次尝试成功",
			mockLoader: func(calls *int) func(ctx context.Context) ([]*cache.Entry, error) {
				return func(ctx context.Context) ([]*cache.Entry, error) {
					*calls++
					if *calls == 1 {
						return nil, errors.New("first attempt fails")
					}
					return []*cache.Entry{
						{Key: "key1", Val: "val1", Expiration: time.Minute},
					}, nil
				}
			}(new(int)),
			mockCache: func(ctrl *gomock.Controller) ecache.Cache {
				cache := cachemocks.NewMockCache(ctrl)
				cache.EXPECT().Set(gomock.Any(), "key1", "val1", time.Minute).Return(nil)
				return cache
			},
			maxRetries:    3,
			expectedErr:   nil,
			expectedCalls: 2,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			cache := tc.mockCache(ctrl)
			calls := 0

			dataLoader := func(ctx context.Context) ([]*cache.Entry, error) {
				calls++
				return tc.mockLoader(ctx)
			}

			mlc := &cache.MultiLevelCache{
				DataLoader:     dataLoader,
				MaxLoadRetries: tc.maxRetries,
				RetryInterval:  time.Millisecond, // 使用短间隔加速测试
			}

			err := mlc.loadFromDBToCacheWithRetry(context.Background(), cache)

			if tc.expectedErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}

			// 验证尝试次数
			assert.Equal(t, tc.expectedCalls, calls)
		})
	}
}
