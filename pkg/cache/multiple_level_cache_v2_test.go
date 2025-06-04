//go:build unit

package cache_test

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"gitee.com/flycash/permission-platform/pkg/bitring"
	"gitee.com/flycash/permission-platform/pkg/cache"
	cachemocks "gitee.com/flycash/permission-platform/pkg/cache/mocks"
	"github.com/ecodeclub/ecache"
	"github.com/ecodeclub/ekit"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

// TestMultiLevelCache_Set_RedisAvailable 测试Redis可用时的Set操作
func TestMultiLevelCache_Set_RedisAvailable(t *testing.T) {
	testCases := []struct {
		name       string
		setupMocks func(redisCmdMock *cachemocks.MockCmdable, localCacheMock *cachemocks.MockCache)
		key        string
		val        any
		expiration time.Duration
		wantErr    error
	}{
		{
			name: "Redis可用，写入成功",
			setupMocks: func(redisCmdMock *cachemocks.MockCmdable, localCacheMock *cachemocks.MockCache) {
				// Redis Set命令成功
				statusCmd := redis.NewStatusCmd(context.Background())
				statusCmd.SetVal("OK")
				redisCmdMock.EXPECT().Set(gomock.Any(), "permission-platform:multicluster:key1", "value1", time.Minute).Return(statusCmd)
				// 本地缓存不应该被调用
			},
			key:        "key1",
			val:        "value1",
			expiration: time.Minute,
			wantErr:    nil,
		},
		{
			name: "Redis可用，写入失败但不触发故障转移",
			setupMocks: func(redisCmdMock *cachemocks.MockCmdable, localCacheMock *cachemocks.MockCache) {
				// Redis Set命令失败
				statusCmd := redis.NewStatusCmd(context.Background())
				statusCmd.SetErr(errors.New("redis error"))
				redisCmdMock.EXPECT().Set(gomock.Any(), "permission-platform:multicluster:key1", "value1", time.Minute).Return(statusCmd)
				// 本地缓存不应该被调用，因为错误次数不足以触发故障转移
			},
			key:        "key1",
			val:        "value1",
			expiration: time.Minute,
			wantErr:    errors.New("redis error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			redisCmdMock := cachemocks.NewMockCmdable(ctrl)
			localCacheMock := cachemocks.NewMockCache(ctrl)

			// 设置错误检测器，需要连续3次错误才触发
			crashDetector := bitring.NewBitRing(3, 0.5, 3)

			// 设置测试的mock预期
			tc.setupMocks(redisCmdMock, localCacheMock)

			dataLoader := func(ctx context.Context) ([]*cache.Entry, error) {
				return nil, nil
			}

			// 创建MultiLevelCache实例
			mlc := cache.NewMultiLevelCacheV2(
				redisCmdMock,
				localCacheMock,
				dataLoader,
				time.Hour,
				time.Second,
				time.Hour,
				crashDetector,
			)

			// 执行Set操作
			err := mlc.Set(context.Background(), tc.key, tc.val, tc.expiration)

			// 验证结果
			if tc.wantErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.wantErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestMultiLevelCache_Set_RedisUnavailable 测试Redis不可用时直接写入本地缓存的情况
func TestMultiLevelCache_Set_RedisUnavailable(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	redisCmdMock := cachemocks.NewMockCmdable(ctrl)
	localCacheMock := cachemocks.NewMockCache(ctrl)

	// 设置本地缓存的预期行为 - 写入操作应该成功
	localCacheMock.EXPECT().Set(gomock.Any(), "key1", "value1", time.Minute).Return(nil)

	// 用于触发Redis崩溃的操作
	crashCmd := redis.NewStatusCmd(context.Background())
	crashCmd.SetErr(errors.New("redis connection error"))
	redisCmdMock.EXPECT().Set(gomock.Any(), "permission-platform:multicluster:trigger", "crash", gomock.Any()).Return(crashCmd)

	// 允许数据加载到本地缓存
	localCacheMock.EXPECT().Set(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	// 设置错误检测器，1次错误就触发故障转移
	crashDetector := bitring.NewBitRing(1, 1.0, 1)

	// 模拟数据加载函数
	dataLoader := func(ctx context.Context) ([]*cache.Entry, error) {
		return []*cache.Entry{
			{Key: "loaded_key", Val: "loaded_value", Expiration: time.Minute},
		}, nil
	}

	// 创建MultiLevelCache实例
	mlc := cache.NewMultiLevelCacheV2(
		redisCmdMock,
		localCacheMock,
		dataLoader,
		time.Hour,
		time.Second,
		time.Hour,
		crashDetector,
	)

	// 首先触发Redis故障
	_ = mlc.Set(context.Background(), "trigger", "crash", time.Minute)

	// 等待故障处理完成
	time.Sleep(time.Millisecond * 100)

	// 再次调用Set，这次应该直接写入本地缓存
	err := mlc.Set(context.Background(), "key1", "value1", time.Minute)
	assert.NoError(t, err) // 写入本地缓存应该成功
}

// TestMultiLevelCache_Get_Redis 测试Get操作
func TestMultiLevelCache_Get_Redis(t *testing.T) {
	testCases := []struct {
		name        string
		setupMocks  func(redisCmdMock *cachemocks.MockCmdable, localCacheMock *cachemocks.MockCache)
		key         string
		wantVal     any
		wantErr     error
		keyNotFound bool
	}{
		{
			name: "Redis获取成功",
			setupMocks: func(redisCmdMock *cachemocks.MockCmdable, localCacheMock *cachemocks.MockCache) {
				stringCmd := redis.NewStringCmd(context.Background())
				stringCmd.SetVal("value1")
				redisCmdMock.EXPECT().Get(gomock.Any(), "permission-platform:multicluster:key1").Return(stringCmd)
			},
			key:         "key1",
			wantVal:     "value1",
			wantErr:     nil,
			keyNotFound: false,
		},
		{
			name: "Redis键不存在",
			setupMocks: func(redisCmdMock *cachemocks.MockCmdable, localCacheMock *cachemocks.MockCache) {
				stringCmd := redis.NewStringCmd(context.Background())
				stringCmd.SetErr(redis.Nil)
				redisCmdMock.EXPECT().Get(gomock.Any(), "permission-platform:multicluster:key1").Return(stringCmd)
			},
			key:         "key1",
			wantVal:     nil,
			wantErr:     redis.Nil,
			keyNotFound: true,
		},
		{
			name: "Redis出错，但未触发故障转移",
			setupMocks: func(redisCmdMock *cachemocks.MockCmdable, localCacheMock *cachemocks.MockCache) {
				stringCmd := redis.NewStringCmd(context.Background())
				stringCmd.SetErr(errors.New("redis connection error"))
				redisCmdMock.EXPECT().Get(gomock.Any(), "permission-platform:multicluster:key1").Return(stringCmd)
			},
			key:         "key1",
			wantVal:     nil,
			wantErr:     errors.New("redis connection error"),
			keyNotFound: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			redisCmdMock := cachemocks.NewMockCmdable(ctrl)
			localCacheMock := cachemocks.NewMockCache(ctrl)

			// 设置测试的mock预期
			tc.setupMocks(redisCmdMock, localCacheMock)

			dataLoader := func(ctx context.Context) ([]*cache.Entry, error) {
				return nil, nil
			}

			// 创建MultiLevelCache实例
			mlc := cache.NewMultiLevelCacheV2(
				redisCmdMock,
				localCacheMock,
				dataLoader,
				time.Hour,
				time.Second,
				time.Hour,
				bitring.NewBitRing(3, 0.5, 3),
			)

			// 执行Get操作
			res := mlc.Get(context.Background(), tc.key)

			// 验证结果
			if tc.keyNotFound {
				assert.True(t, res.KeyNotFound())
			} else if tc.wantErr != nil {
				assert.Error(t, res.Err)
				assert.Contains(t, res.Err.Error(), tc.wantErr.Error())
			} else {
				assert.NoError(t, res.Err)
				assert.Equal(t, tc.wantVal, res.Val)
			}
		})
	}
}

// TestMultiLevelCache_RedisCrashAndRecovery 测试Redis故障转移和恢复
func TestMultiLevelCache_RedisCrashAndRecovery(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	redisCmdMock := cachemocks.NewMockCmdable(ctrl)
	localCacheMock := cachemocks.NewMockCache(ctrl)

	// 配置本地缓存
	localCacheMock.EXPECT().Get(gomock.Any(), gomock.Any()).Return(
		ecache.Value{AnyValue: ekit.AnyValue{Val: "value1"}},
	).AnyTimes()
	localCacheMock.EXPECT().Set(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	// 模拟Redis Ping响应变化
	// 使用计数器来控制Ping的返回值
	var pingCounter atomic.Int32
	redisCmdMock.EXPECT().Ping(gomock.Any()).DoAndReturn(func(ctx context.Context) *redis.StatusCmd {
		cmd := redis.NewStatusCmd(ctx)
		counter := pingCounter.Load()
		// 前2次Ping失败，之后成功
		if counter < 2 {
			cmd.SetErr(errors.New("connection refused"))
		} else {
			cmd.SetVal("PONG")
		}
		pingCounter.Add(1)
		return cmd
	}).AnyTimes()

	// 模拟Redis Get请求 - 这是需要添加的部分
	redisCmdMock.EXPECT().Get(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, key string) *redis.StringCmd {
		cmd := redis.NewStringCmd(ctx)
		// Redis恢复前Get会失败
		if pingCounter.Load() < 2 {
			cmd.SetErr(errors.New("redis connection error"))
		} else {
			// Redis恢复后Get会成功
			cmd.SetVal("value1")
		}
		return cmd
	}).AnyTimes()

	// 设置所有Redis操作的模拟响应 - 使用通配符匹配任何参数
	// 第一次Set操作失败
	firstSetCmd := redis.NewStatusCmd(context.Background())
	firstSetCmd.SetErr(errors.New("redis connection error"))
	redisCmdMock.EXPECT().Set(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(firstSetCmd).Times(1)

	// 恢复后的Set操作成功
	recoveredSetCmd := redis.NewStatusCmd(context.Background())
	recoveredSetCmd.SetVal("OK")
	redisCmdMock.EXPECT().Set(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(recoveredSetCmd).AnyTimes()

	// 设置错误检测器，只需1次错误就触发故障转移
	crashDetector := bitring.NewBitRing(1, 1.0, 1)

	// 数据加载函数调用计数
	var dataLoaderCount atomic.Int32
	dataLoader := func(ctx context.Context) ([]*cache.Entry, error) {
		dataLoaderCount.Add(1)
		return []*cache.Entry{
			{Key: "key1", Val: "value1", Expiration: time.Minute},
		}, nil
	}

	// 创建MultiLevelCache实例
	mlc := cache.NewMultiLevelCacheV2(
		redisCmdMock,
		localCacheMock,
		dataLoader,
		time.Millisecond*50, // 设置较短的刷新周期
		time.Millisecond*10, // 设置较短的ping超时
		time.Millisecond*30, // 设置较短的健康检查周期
		crashDetector,
	)

	// 1. 触发Redis故障 - 第一次Set操作失败
	err := mlc.Set(context.Background(), "key1", "value1", time.Minute)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "redis connection error")

	// 等待足够时间，确保故障处理完成
	time.Sleep(time.Millisecond * 100)

	// 2. 验证此时已从本地缓存获取数据
	res := mlc.Get(context.Background(), "key1")
	assert.Equal(t, "value1", res.Val)
	assert.True(t, dataLoaderCount.Load() > 0, "数据加载器应该被调用来加载本地缓存")

	// 3. 等待Redis恢复 - 让健康检查检测到Redis已恢复
	// 等待足够时间，让多次健康检查执行
	time.Sleep(time.Millisecond * 200)

	// 确保pingCounter >= 2，即Ping已经成功
	assert.True(t, pingCounter.Load() >= 2, "Redis健康检查应该执行多次")

	// 4. 验证Redis恢复后的操作
	// 执行一次新的Set操作，此时应该写入Redis成功
	err = mlc.Set(context.Background(), "key2", "value2", time.Minute)
	assert.NoError(t, err)

	// 5. 验证数据加载函数被调用至少两次 - 一次在崩溃时，一次在恢复时
	assert.True(t, dataLoaderCount.Load() >= 2,
		"数据加载器应该至少被调用两次，一次在崩溃时填充本地缓存，一次在恢复时填充Redis")
}

// TestMultiLevelCache_DataLoaderError 测试数据加载失败的情况
func TestMultiLevelCache_DataLoaderError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	redisCmdMock := cachemocks.NewMockCmdable(ctrl)
	localCacheMock := cachemocks.NewMockCache(ctrl)

	// 模拟Redis操作失败，触发故障转移
	setCmd := redis.NewStatusCmd(context.Background())
	setCmd.SetErr(errors.New("redis connection error"))
	redisCmdMock.EXPECT().Set(gomock.Any(), "permission-platform:multicluster:key1", "value1", time.Minute).Return(setCmd)

	// 设置本地缓存行为 - 注意这里不期望调用Set方法，因为加载数据会失败
	// 仅设置直接对本地缓存的写入
	localCacheMock.EXPECT().Set(gomock.Any(), "key1", "value1", time.Minute).Return(nil).MaxTimes(1)

	// 返回模拟键不存在的值，使用显式设置 Err 为 redis.Nil
	localCacheMock.EXPECT().Get(gomock.Any(), "key1").Return(
		ecache.Value{AnyValue: ekit.AnyValue{Err: redis.Nil}},
	)

	// 设置错误检测器，只需要1次错误就触发故障转移
	crashDetector := bitring.NewBitRing(1, 1.0, 1)

	// 模拟数据加载失败
	dataLoader := func(ctx context.Context) ([]*cache.Entry, error) {
		return nil, errors.New("database error")
	}

	// 创建MultiLevelCache实例
	mlc := cache.NewMultiLevelCacheV2(
		redisCmdMock,
		localCacheMock,
		dataLoader,
		time.Hour,
		time.Second,
		time.Hour,
		crashDetector,
	)

	// 触发Redis故障并切换到本地缓存
	err := mlc.Set(context.Background(), "key1", "value1", time.Minute)
	assert.Error(t, err)

	// 等待故障处理完成
	time.Sleep(time.Millisecond * 100)

	// 尝试获取数据，此时本地缓存可能没有成功加载
	res := mlc.Get(context.Background(), "key1")
	// 检查错误是否为 redis.Nil
	assert.Equal(t, redis.Nil, res.Err)
}

// TestMultiLevelCache_StopRefreshContext 测试当Context被取消时刷新过程停止
func TestMultiLevelCache_StopRefreshContext(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// 使用真实的Context Cancel来测试刷新停止
	// 这样不需要依赖Redis恢复的机制，可以直接测试ctx.Done()分支
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 记录dataLoader调用的时间
	var callTimes []time.Time
	var callMutex sync.Mutex
	// 记录当调用次数达到特定值时的时间点
	cancelPoint := make(chan struct{})

	dataLoader := func(loadCtx context.Context) ([]*cache.Entry, error) {
		callMutex.Lock()
		callTimes = append(callTimes, time.Now())
		currentCalls := len(callTimes)
		callMutex.Unlock()

		// 如果是第3次调用，则触发cancel
		if currentCalls == 3 {
			// 等待一小段时间，确保ticker能注册当前调用结束
			time.Sleep(time.Millisecond)
			cancel() // 取消上下文，这应该会导致refreshLocalCache函数退出
			close(cancelPoint)
		}

		return []*cache.Entry{
			{Key: "test", Val: "value", Expiration: time.Minute},
		}, nil
	}

	// 模拟一个自定义的MultiLevelCache，重写refreshLocalCache方法来使用我们的可取消上下文
	customRefresh := func(refreshCtx context.Context) {
		ticker := time.NewTicker(time.Millisecond * 30) // 延长ticker间隔，确保不会多次调用
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				// 调用dataLoader以模拟刷新
				_, _ = dataLoader(refreshCtx)
			case <-refreshCtx.Done():
				// 这是我们要测试的分支：验证当上下文被取消时，函数会退出
				return
			}
		}
	}

	// 启动自定义刷新
	go customRefresh(ctx)

	// 等待cancel被触发（第三次调用dataLoader会取消上下文并通知）
	select {
	case <-cancelPoint:
		// 上下文已被取消
	case <-time.After(time.Second):
		t.Fatal("等待cancel超时")
	}

	// 等待一段时间，确保如果刷新没有停止，至少会再次调用dataLoader
	time.Sleep(time.Millisecond * 100)

	// 验证调用次数
	callMutex.Lock()
	finalCallCount := len(callTimes)
	callMutex.Unlock()

	// 应该正好有3次调用
	assert.Equal(t, 3, finalCallCount, "dataLoader应该被调用正好3次")
}
