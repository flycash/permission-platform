//go:build e2e

package gorm

import (
	"context"
	"testing"

	"gitee.com/flycash/permission-platform/internal/test/ioc"

	"github.com/stretchr/testify/require"

	permissionv1 "gitee.com/flycash/permission-platform/api/proto/gen/permission/v1"
	"github.com/stretchr/testify/suite"
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
func (u User) ResourceKey(_ context.Context) string {
	return "user"
}

func (u User) ResourceType(_ context.Context) string {
	return "user"
}

// SimpleModel 一个没有实现 AuthRequired 接口的模型
type SimpleModel struct {
	ID   int `gorm:"primary_key"`
	Name string
}

// Mock permission service client
type mockPermissionServiceClient struct {
	permissionv1.PermissionServiceClient
}

func newMockPermissionServiceClient() *mockPermissionServiceClient {
	return &mockPermissionServiceClient{}
}

func (*mockPermissionServiceClient) CheckPermission(_ context.Context, req *permissionv1.CheckPermissionRequest, _ ...grpc.CallOption) (*permissionv1.CheckPermissionResponse, error) {
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

// GormAccessPluginTestSuite 测试套件
type GormAccessPluginTestSuite struct {
	suite.Suite
	db     *gorm.DB
	plugin *GormAccessPlugin
}

func (s *GormAccessPluginTestSuite) SetupSuite() {
	// 创建数据库连接
	dsn := "root:root@tcp(localhost:13316)/permission?charset=utf8mb4&collation=utf8mb4_general_ci&parseTime=True&loc=Local&timeout=1s&readTimeout=3s&writeTimeout=3s&multiStatements=true"
	ioc.WaitForDBSetup(dsn)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	s.Require().NoError(err)
	s.db = db
	err = db.AutoMigrate(&User{}, &SimpleModel{})
	require.NoError(s.T(), err)
	// 创建并初始化插件
	mockClient := newMockPermissionServiceClient()
	s.plugin = NewGormAccessPlugin(mockClient)
	err = s.plugin.Initialize(db)
	s.Require().NoError(err)
}

func (s *GormAccessPluginTestSuite) TestReadOperation() {
	ctx := context.Background()
	ctx = context.WithValue(ctx, bizIDKey, int64(1))
	ctx = context.WithValue(ctx, uidKey, int64(1))
	db := s.db.WithContext(ctx)

	var user User
	err := db.First(&user, 1).Error
	s.NoError(err)
}

func (s *GormAccessPluginTestSuite) TestCreateOperation() {
	ctx := context.Background()
	ctx = context.WithValue(ctx, bizIDKey, int64(1))
	ctx = context.WithValue(ctx, uidKey, int64(1))
	db := s.db.WithContext(ctx)

	err := db.Create(&User{Name: "test", Age: 20}).Error
	s.NoError(err)
}

func (s *GormAccessPluginTestSuite) TestUpdateOperation() {
	ctx := context.Background()
	ctx = context.WithValue(ctx, bizIDKey, int64(1))
	ctx = context.WithValue(ctx, uidKey, int64(1))
	db := s.db.WithContext(ctx)

	err := db.Model(&User{}).Where("id = ?", 1).Update("name", "updated").Error
	s.NoError(err)
}

func (s *GormAccessPluginTestSuite) TestDeleteOperation() {
	ctx := context.Background()
	ctx = context.WithValue(ctx, bizIDKey, int64(1))
	ctx = context.WithValue(ctx, uidKey, int64(1))
	db := s.db.WithContext(ctx)

	err := db.Delete(&User{}, 1).Error
	s.Error(err)
}

func (s *GormAccessPluginTestSuite) TestMissingBizID() {
	ctx := context.Background()
	ctx = context.WithValue(ctx, uidKey, int64(1))
	db := s.db.WithContext(ctx)

	err := db.Create(&User{Name: "test", Age: 20}).Error
	s.NoError(err)
}

func (s *GormAccessPluginTestSuite) TestMissingUID() {
	ctx := context.Background()
	ctx = context.WithValue(ctx, bizIDKey, int64(1))
	db := s.db.WithContext(ctx)

	err := db.Create(&User{Name: "test", Age: 20}).Error
	s.NoError(err)
}

func (s *GormAccessPluginTestSuite) TestSimpleModelWithResourceInContext() {
	ctx := context.Background()
	ctx = context.WithValue(ctx, bizIDKey, int64(1))
	ctx = context.WithValue(ctx, uidKey, int64(1))
	ctx = context.WithValue(ctx, resourceKey, &permissionv1.Resource{
		Key:  "simple",
		Type: "simple",
	})
	db := s.db.WithContext(ctx)
	err := db.Create(&SimpleModel{ID: 1, Name: "test"}).Error
	s.NoError(err)
	// 测试读取操作
	var model SimpleModel
	err = db.First(&model, 1).Error
	s.NoError(err)

	// 测试删除操作
	err = db.Delete(&SimpleModel{}, 1).Error
	s.Error(err)
}

func TestGormAccessPluginSuite(t *testing.T) {
	suite.Run(t, new(GormAccessPluginTestSuite))
}
