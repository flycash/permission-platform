//go:build unit

package cache_test

import (
	"context"
	"errors"
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
			mlc := cache.NewMultiLevelCache(
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
			mlc := cache.NewMultiLevelCache(
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

	// 配置本地缓存返回值
	localCacheMock.EXPECT().Get(gomock.Any(), "key2").Return(
		ecache.Value{AnyValue: ekit.AnyValue{Val: "value2"}},
	).AnyTimes()

	// 配置本地缓存可接收任何数据写入
	localCacheMock.EXPECT().Set(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	// 第1步：设置成功的Redis Ping，确保初始状态是可用的
	pingCmd := redis.NewStatusCmd(context.Background())
	pingCmd.SetVal("PONG")
	redisCmdMock.EXPECT().Ping(gomock.Any()).Return(pingCmd).AnyTimes()

	// 第2步：第一次Set操作成功
	setCmd1 := redis.NewStatusCmd(context.Background())
	setCmd1.SetVal("OK")
	redisCmdMock.EXPECT().Set(gomock.Any(), "permission-platform:multicluster:key1", "value1", time.Minute).Return(setCmd1)

	// 第3步：第二次Set操作时Redis失败，触发故障转移
	setCmd2 := redis.NewStatusCmd(context.Background())
	setCmd2.SetErr(errors.New("redis connection error"))
	redisCmdMock.EXPECT().Set(gomock.Any(), "permission-platform:multicluster:key2", "value2", time.Minute).Return(setCmd2)

	// 设置错误检测器，只需要1次错误就触发故障转移
	crashDetector := bitring.NewBitRing(1, 1.0, 1)

	// 模拟数据加载函数
	dataLoaderCalled := false
	dataLoader := func(ctx context.Context) ([]*cache.Entry, error) {
		dataLoaderCalled = true
		return []*cache.Entry{
			{Key: "key2", Val: "value2", Expiration: time.Minute},
		}, nil
	}

	// 创建MultiLevelCache实例
	mlc := cache.NewMultiLevelCache(
		redisCmdMock,
		localCacheMock,
		dataLoader,
		time.Millisecond*100, // 设置较短的刷新周期，以便测试
		time.Millisecond*50,  // 设置较短的超时，以便测试
		time.Millisecond*50,  // 设置较短的健康检查周期，以便测试
		crashDetector,
	)

	// 1. 第一次Set操作成功
	err := mlc.Set(context.Background(), "key1", "value1", time.Minute)
	assert.NoError(t, err)

	// 2. 第二次Set操作失败，应触发故障转移
	err = mlc.Set(context.Background(), "key2", "value2", time.Minute)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "redis connection error")

	// 等待一段时间，确保故障处理完成
	time.Sleep(time.Millisecond * 200)

	// 3. 验证dataLoader被调用
	assert.True(t, dataLoaderCalled)

	// 4. 现在Get操作应该从本地缓存获取
	res := mlc.Get(context.Background(), "key2")
	assert.Equal(t, "value2", res.Val)
}

// TestMultiLevelCache_RedisUnavailable 测试Redis不可用时的本地缓存操作
func TestMultiLevelCache_RedisUnavailable(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	redisCmdMock := cachemocks.NewMockCmdable(ctrl)
	localCacheMock := cachemocks.NewMockCache(ctrl)

	// 模拟Redis不可用状态
	// 初始Set操作失败，并设置为触发故障转移
	setCmd := redis.NewStatusCmd(context.Background())
	setCmd.SetErr(errors.New("redis connection error"))
	redisCmdMock.EXPECT().Set(gomock.Any(), "permission-platform:multicluster:key1", "value1", time.Minute).Return(setCmd)

	// 设置本地缓存的行为
	localCacheMock.EXPECT().Set(gomock.Any(), "key1", "value1", time.Minute).Return(nil)
	localCacheMock.EXPECT().Get(gomock.Any(), "key1").Return(
		ecache.Value{AnyValue: ekit.AnyValue{Val: "value1"}},
	)

	// 允许数据加载到本地缓存
	localCacheMock.EXPECT().Set(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	// 设置错误检测器，只需要1次错误就触发故障转移
	crashDetector := bitring.NewBitRing(1, 1.0, 1)

	// 模拟数据加载函数
	dataLoader := func(ctx context.Context) ([]*cache.Entry, error) {
		return []*cache.Entry{
			{Key: "key1", Val: "value1", Expiration: time.Minute},
		}, nil
	}

	// 创建MultiLevelCache实例
	mlc := cache.NewMultiLevelCache(
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
	assert.Contains(t, err.Error(), "redis connection error")

	// 等待故障处理完成
	time.Sleep(time.Millisecond * 100)

	// 现在从本地缓存获取数据
	res := mlc.Get(context.Background(), "key1")
	assert.NoError(t, res.Err)
	assert.Equal(t, "value1", res.Val)
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
	mlc := cache.NewMultiLevelCache(
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
