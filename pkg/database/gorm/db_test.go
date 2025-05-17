package gorm

import (
	"context"
	permissionv1 "gitee.com/flycash/permission-platform/api/proto/gen/permission/v1"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type User struct {
	ID   int `gorm:"primary_key"`
	Name string
	Age  int
}

// Implement AuthRequired interface for User
func (u User) ResourceKey(ctx context.Context) string {
	return "user"
}

func (u User) ResourceType(ctx context.Context) string {
	return "user"
}

// Mock permission service client
type mockPermissionServiceClient struct {
	permissionv1.PermissionServiceClient
}

func newMockPermissionServiceClient() *mockPermissionServiceClient {
	return &mockPermissionServiceClient{}
}

func (m *mockPermissionServiceClient) CheckPermission(ctx context.Context, req *permissionv1.CheckPermissionRequest, opts ...grpc.CallOption) (*permissionv1.CheckPermissionResponse, error) {
	// 检查请求中的action
	for _, action := range req.Permission.Actions {
		switch action {
		case "read", "create", "update":
			// 允许读取、创建和更新操作
			return &permissionv1.CheckPermissionResponse{
				Allowed: true,
			}, nil
		case "delete":
			// 拒绝删除操作
			return &permissionv1.CheckPermissionResponse{
				Allowed: false,
			}, nil
		default:
			// 其他操作默认拒绝
			return &permissionv1.CheckPermissionResponse{
				Allowed: false,
			}, nil
		}
	}
	return &permissionv1.CheckPermissionResponse{
		Allowed: false,
	}, nil
}

func TestGormAccessPlugin(t *testing.T) {
	// Create test cases
	tests := []struct {
		name          string
		setupContext  func(ctx context.Context) context.Context
		operation     func(db *gorm.DB) error
		expectedError bool
	}{
		{
			name: "read operation - allowed",
			setupContext: func(ctx context.Context) context.Context {
				ctx = context.WithValue(ctx, bizIDKey, int64(1))
				ctx = context.WithValue(ctx, uidKey, int64(1))
				return ctx
			},
			operation: func(db *gorm.DB) error {
				var user User
				return db.First(&user, 1).Error
			},
			expectedError: false,
		},
		{
			name: "create operation - allowed",
			setupContext: func(ctx context.Context) context.Context {
				ctx = context.WithValue(ctx, bizIDKey, int64(1))
				ctx = context.WithValue(ctx, uidKey, int64(1))
				return ctx
			},
			operation: func(db *gorm.DB) error {
				return db.Create(&User{Name: "test", Age: 20}).Error
			},
			expectedError: false,
		},
		{
			name: "update operation - allowed",
			setupContext: func(ctx context.Context) context.Context {
				ctx = context.WithValue(ctx, bizIDKey, int64(1))
				ctx = context.WithValue(ctx, uidKey, int64(1))
				return ctx
			},
			operation: func(db *gorm.DB) error {
				return db.Model(&User{}).Where("id = ?", 1).Update("name", "updated").Error
			},
			expectedError: false,
		},
		{
			name: "delete operation - denied",
			setupContext: func(ctx context.Context) context.Context {
				ctx = context.WithValue(ctx, bizIDKey, int64(1))
				ctx = context.WithValue(ctx, uidKey, int64(1))
				return ctx
			},
			operation: func(db *gorm.DB) error {
				return db.Delete(&User{}, 1).Error
			},
			expectedError: true,
		},
		{
			name: "missing biz_id",
			setupContext: func(ctx context.Context) context.Context {
				ctx = context.WithValue(ctx, uidKey, int64(1))
				return ctx
			},
			operation: func(db *gorm.DB) error {
				return db.Create(&User{Name: "test", Age: 20}).Error
			},
			expectedError: false,
		},
		{
			name: "missing uid",
			setupContext: func(ctx context.Context) context.Context {
				ctx = context.WithValue(ctx, bizIDKey, int64(1))
				return ctx
			},
			operation: func(db *gorm.DB) error {
				return db.Create(&User{Name: "test", Age: 20}).Error
			},
			expectedError: false,
		},
	}

	// Run tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test database connection
			dsn := "root:root@tcp(localhost:13316)/permission?charset=utf8mb4&collation=utf8mb4_general_ci&parseTime=True&loc=Local&timeout=1s&readTimeout=3s&writeTimeout=3s&multiStatements=true"
			db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
			require.NoError(t, err)

			// Create and initialize the plugin with mock client
			mockClient := newMockPermissionServiceClient()
			plugin := NewGormAccessPlugin(mockClient)
			err = plugin.Initialize(db)
			require.NoError(t, err)

			// Create context with test values
			ctx := tt.setupContext(context.Background())
			db = db.WithContext(ctx)

			// Run the operation
			err = tt.operation(db)

			// Check error expectation
			if tt.expectedError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
