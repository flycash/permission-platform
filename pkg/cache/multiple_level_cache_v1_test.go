//go:build unit

package cache

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"gitee.com/flycash/permission-platform/pkg/bitring"
	cachemocks "gitee.com/flycash/permission-platform/pkg/cache/mocks"
	"github.com/ecodeclub/ecache"
	"github.com/ecodeclub/ekit"
	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

// TestMultiLevelCache_Set_RedisAvailable 测试Redis可用时的Set操作
func TestMultiLevelCacheV2_Set_RedisAvailable(t *testing.T) {
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
				redisCmdMock.EXPECT().Set(gomock.Any(), "permission-platform:multiclusterv2:key1", "value1", time.Minute).Return(statusCmd)
				localCacheMock.EXPECT().Set(gomock.Any(), "key1", "value1", time.Minute).Return(nil)
				// 本地缓存不应该被调用
			},
			key:        "key1",
			val:        "value1",
			expiration: time.Minute,
			wantErr:    nil,
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

			// 创建MultiLevelCache实例
			mlc := NewMultiLevelCacheV2(
				localCacheMock,
				time.Hour,
				time.Second,
				crashDetector,
				redisCmdMock,
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

// TestMultiLevelCacheV2_Set_RedisUnavailable 测试Redis不可用时直接写入本地缓存的情况
func TestMultiLevelCacheV2_Set_RedisUnavailable(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	redisCmdMock := cachemocks.NewMockCmdable(ctrl)
	localCacheMock := cachemocks.NewMockCache(ctrl)

	// 设置本地缓存的预期行为 - 写入操作应该成功
	localCacheMock.EXPECT().Set(gomock.Any(), "key1", "value1", time.Minute).Return(nil)

	// 用于触发Redis崩溃的操作
	crashCmd := redis.NewStatusCmd(context.Background())
	crashCmd.SetErr(errors.New("redis connection error"))
	redisCmdMock.EXPECT().Set(gomock.Any(), "permission-platform:multiclusterv2:trigger", "crash", gomock.Any()).Return(crashCmd)

	// 允许数据加载到本地缓存
	localCacheMock.EXPECT().Set(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	// 设置错误检测器，1次错误就触发故障转移
	crashDetector := bitring.NewBitRing(1, 1.0, 1)

	// 创建MultiLevelCache实例
	mlc := NewMultiLevelCacheV2(
		localCacheMock,
		time.Hour,
		time.Second,
		crashDetector,
		redisCmdMock,
	)

	// 首先触发Redis故障
	_ = mlc.Set(context.Background(), "trigger", "crash", time.Minute)

	// 等待故障处理完成
	time.Sleep(time.Millisecond * 100)

	// 再次调用Set，这次应该直接写入本地缓存
	err := mlc.Set(context.Background(), "key1", "value1", time.Minute)
	assert.NoError(t, err) // 写入本地缓存应该成功
}

// TestMultiLevelCacheV2_Get_Redis 测试Get操作
func TestMultiLevelCacheV2_Get_Redis(t *testing.T) {
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
				redisCmdMock.EXPECT().Get(gomock.Any(), "permission-platform:multiclusterv2:key1").Return(stringCmd)
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
				redisCmdMock.EXPECT().Get(gomock.Any(), "permission-platform:multiclusterv2:key1").Return(stringCmd)
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
				redisCmdMock.EXPECT().Get(gomock.Any(), "permission-platform:multiclusterv2:key1").Return(stringCmd)
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
			// 创建MultiLevelCache实例
			mlc := NewMultiLevelCacheV2(
				localCacheMock,
				time.Hour,
				time.Second,
				bitring.NewBitRing(3, 0.5, 3),
				redisCmdMock,
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
func TestMultiLevelCacheV2_RedisCrashAndRecovery(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	redisCmdMock := cachemocks.NewMockCmdable(ctrl)
	localCacheMock := cachemocks.NewMockCache(ctrl)

	// 配置本地缓存
	localCacheMock.EXPECT().Get(gomock.Any(), gomock.Any()).Return(
		ecache.Value{AnyValue: ekit.AnyValue{Val: "local"}},
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

	mlc := NewMultiLevelCacheV2(
		localCacheMock,
		time.Millisecond*50, // 设置较短的刷新周期
		time.Millisecond*10, // 设置较短的ping超时
		bitring.NewBitRing(3, 0.5, 3),
		redisCmdMock,
	)

	// 1. 触发Redis故障 - 第一次Set操作失败
	err := mlc.Set(context.Background(), "key1", "value1", time.Minute)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "redis connection error")

	// 等待足够时间，确保故障处理完成
	time.Sleep(time.Millisecond * 100)

	// 2. 验证此时已从本地缓存获取数据
	res := mlc.Get(context.Background(), "key1")
	assert.Equal(t, "local", res.Val)

	// 3. 等待Redis恢复 - 让健康检查检测到Redis已恢复
	// 等待足够时间，让多次健康检查执行
	time.Sleep(time.Millisecond * 200)

	// 确保pingCounter >= 2，即Ping已经成功
	assert.True(t, pingCounter.Load() >= 2, "Redis健康检查应该执行多次")

	// 4. 验证Redis恢复后的操作
	// 执行一次新的Set操作，此时应该写入Redis成功
	err = mlc.Set(context.Background(), "key2", "value2", time.Minute)
	assert.NoError(t, err)
}
