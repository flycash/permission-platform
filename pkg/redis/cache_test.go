//go:build e2e

package redis

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	permissionv1 "gitee.com/flycash/permission-platform/api/proto/gen/permission/v1"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
)

// MockPermissionServiceClient 是一个模拟的权限服务客户端
type MockPermissionServiceClient struct {
	mock.Mock
}

func (m *MockPermissionServiceClient) CheckPermission(ctx context.Context, req *permissionv1.CheckPermissionRequest, _ ...grpc.CallOption) (*permissionv1.CheckPermissionResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*permissionv1.CheckPermissionResponse), args.Error(1)
}

func TestAccessPlugin(t *testing.T) {
	// 创建模拟的权限服务客户端
	mockClient := new(MockPermissionServiceClient)

	// 创建插件
	plugin := NewAccessPlugin(mockClient)

	// 创建 Redis 客户端
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	client2 := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	client.AddHook(plugin)

	// 测试用例
	tests := []struct {
		name          string
		setupContext  func(ctx context.Context) context.Context
		setupMock     func()
		operation     func(ctx context.Context) error
		expectedError bool
	}{
		{
			name: "get operation-allowed",
			setupContext: func(ctx context.Context) context.Context {
				ctx = context.WithValue(ctx, bizIDKey, int64(1))
				ctx = context.WithValue(ctx, uidKey, int64(1))
				return ctx
			},
			setupMock: func() {
				_, err := client2.Set(t.Context(), "test-key", 1, 10*time.Second).Result()
				require.NoError(t, err)
				mockClient.On("CheckPermission", mock.Anything, &permissionv1.CheckPermissionRequest{
					Uid: 1,
					Permission: &permissionv1.Permission{
						ResourceKey:  "test-key",
						ResourceType: "redis_key",
						Actions:      []string{"read"},
					},
				}).Return(&permissionv1.CheckPermissionResponse{Allowed: true}, nil)
			},
			operation: func(ctx context.Context) error {
				_, err := client.Get(ctx, "test-key").Result()
				return err
			},
			expectedError: false,
		},
		{
			name: "set operation - denied",
			setupContext: func(ctx context.Context) context.Context {
				ctx = context.WithValue(ctx, bizIDKey, int64(1))
				ctx = context.WithValue(ctx, uidKey, int64(1))
				return ctx
			},
			setupMock: func() {
				mockClient.On("CheckPermission", mock.Anything, &permissionv1.CheckPermissionRequest{
					Uid: 1,
					Permission: &permissionv1.Permission{
						ResourceKey:  "test-key1",
						ResourceType: "redis_key",
						Actions:      []string{"write"},
					},
				}).Return(&permissionv1.CheckPermissionResponse{Allowed: false}, nil)
			},
			operation: func(ctx context.Context) error {
				return client.Set(ctx, "test-key1", "value", 0).Err()
			},
			expectedError: true,
		},
		{
			name: "del operation-allowed",
			setupContext: func(ctx context.Context) context.Context {
				ctx = context.WithValue(ctx, bizIDKey, int64(1))
				ctx = context.WithValue(ctx, uidKey, int64(1))
				return ctx
			},
			setupMock: func() {
				mockClient.On("CheckPermission", mock.Anything, &permissionv1.CheckPermissionRequest{
					Uid: 1,
					Permission: &permissionv1.Permission{
						ResourceKey:  "test-key2",
						ResourceType: "redis_key",
						Actions:      []string{"delete"},
					},
				}).Return(&permissionv1.CheckPermissionResponse{Allowed: true}, nil)
			},
			operation: func(ctx context.Context) error {
				return client.Del(ctx, "test-key2").Err()
			},
			expectedError: false,
		},
		{
			name: "with custom resource",
			setupContext: func(ctx context.Context) context.Context {
				_, err := client2.Set(t.Context(), "custom-key", 1, 10*time.Second).Result()
				require.NoError(t, err)
				ctx = context.WithValue(ctx, bizIDKey, int64(1))
				ctx = context.WithValue(ctx, uidKey, int64(1))
				ctx = context.WithValue(ctx, resourceKey, &permissionv1.Resource{
					Key:  "custom-key",
					Type: "custom_type",
				})
				return ctx
			},
			setupMock: func() {
				mockClient.On("CheckPermission", mock.Anything, &permissionv1.CheckPermissionRequest{
					Uid: 1,
					Permission: &permissionv1.Permission{
						ResourceKey:  "custom-key",
						ResourceType: "custom_type",
						Actions:      []string{"read"},
					},
				}).Return(&permissionv1.CheckPermissionResponse{Allowed: true}, nil)
			},
			operation: func(ctx context.Context) error {
				_, err := client.Get(ctx, "custom-key").Result()
				return err
			},
			expectedError: false,
		},
	}

	// 运行测试用例
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 设置 mock
			tt.setupMock()

			// 创建上下文
			ctx := tt.setupContext(t.Context())

			// 执行操作
			err := tt.operation(ctx)

			// 验证结果
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// 验证 mock 调用
			mockClient.AssertExpectations(t)
		})
	}
}
