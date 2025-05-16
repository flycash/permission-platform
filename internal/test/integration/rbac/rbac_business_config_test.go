//go:build e2e

package rbac

import (
	"context"
	"fmt"
	"testing"

	rbacioc "gitee.com/flycash/permission-platform/internal/test/integration/ioc/rbac"
	testioc "gitee.com/flycash/permission-platform/internal/test/ioc"
	"github.com/ego-component/egorm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// BusinessConfigTestSuite 业务配置测试套件
type BusinessConfigTestSuite struct {
	suite.Suite
	db  *egorm.Component
	svc *rbacioc.Service
}

// SetupSuite 在所有测试之前设置
func (s *BusinessConfigTestSuite) SetupSuite() {
	s.db = testioc.InitDBAndTables()
	s.svc = rbacioc.Init()
}

// TearDownSuite 在所有测试之后清理
func (s *BusinessConfigTestSuite) TearDownSuite() {
	// 清理所有测试数据，保留初始化脚本中的预设数据
	ctx := context.Background()
	cleanTestEnvironment(s.T(), ctx, s.svc)
}

// TestBusinessConfigSuite 运行业务配置测试套件
func TestBusinessConfigSuite(t *testing.T) {
	suite.Run(t, new(BusinessConfigTestSuite))
}

// TestBusinessConfig_Create 测试创建业务配置
func (s *BusinessConfigTestSuite) TestBusinessConfig_Create() {
	t := s.T()
	ctx := context.Background()

	// 创建业务配置
	bizConfig := createTestBusinessConfig("业务配置测试-创建")
	created, err := s.svc.Svc.CreateBusinessConfig(ctx, bizConfig)
	require.NoError(t, err)
	require.NotZero(t, created.ID)
	// 确保业务ID不是预设的ID=1
	assert.Greater(t, created.ID, int64(1))

	// 验证业务配置字段
	assertBusinessConfig(t, bizConfig, created)

	// 清理测试数据
	cleanupTest(t, ctx, func() error {
		return s.svc.Svc.DeleteBusinessConfigByID(ctx, created.ID)
	})
}

// TestBusinessConfig_Get 测试获取业务配置
func (s *BusinessConfigTestSuite) TestBusinessConfig_Get() {
	t := s.T()
	ctx := context.Background()

	// 创建业务配置
	bizConfig := createTestBusinessConfig("业务配置测试-获取")
	created, err := s.svc.Svc.CreateBusinessConfig(ctx, bizConfig)
	require.NoError(t, err)

	// 确保业务ID不是预设的ID=1
	assert.Greater(t, created.ID, int64(1))

	// 获取业务配置
	retrieved, err := s.svc.Svc.GetBusinessConfigByID(ctx, created.ID)
	require.NoError(t, err)
	assertBusinessConfig(t, created, retrieved)

	// 清理测试数据
	cleanupTest(t, ctx, func() error {
		return s.svc.Svc.DeleteBusinessConfigByID(ctx, created.ID)
	})
}

// TestBusinessConfig_Update 测试更新业务配置
func (s *BusinessConfigTestSuite) TestBusinessConfig_Update() {
	t := s.T()
	ctx := context.Background()

	// 先创建一个业务配置用于后续测试
	config := createTestBusinessConfig("更新测试")
	created, err := s.svc.Svc.CreateBusinessConfig(ctx, config)
	require.NoError(t, err)

	// 测试结束后清理
	defer cleanupTest(t, ctx, func() error {
		return s.svc.Svc.DeleteBusinessConfigByID(ctx, created.ID)
	})

	t.Run("更新业务配置成功", func(t *testing.T) {
		updated := created
		updated.Name = "更新后的业务名称"
		updated.RateLimit = 200

		result, err := s.svc.Svc.UpdateBusinessConfig(ctx, updated)
		assert.NoError(t, err)
		assert.Equal(t, "更新后的业务名称", result.Name)
		assert.Equal(t, 200, result.RateLimit)

		// 再次获取确认更新成功
		found, err := s.svc.Svc.GetBusinessConfigByID(ctx, result.ID)
		require.NoError(t, err)
		assertBusinessConfig(t, updated, found)
	})

	t.Run("更新不存在的业务配置", func(t *testing.T) {
		nonExistentConfig := createTestBusinessConfig("不存在")
		nonExistentConfig.ID = 99999
		_, err := s.svc.Svc.UpdateBusinessConfig(ctx, nonExistentConfig)
		assert.NoError(t, err)
	})
}

// TestBusinessConfig_Delete 测试删除业务配置
func (s *BusinessConfigTestSuite) TestBusinessConfig_Delete() {
	t := s.T()
	ctx := context.Background()

	t.Run("删除存在的业务配置", func(t *testing.T) {
		// 创建一个业务配置
		config := createTestBusinessConfig("删除测试")
		created, err := s.svc.Svc.CreateBusinessConfig(ctx, config)
		require.NoError(t, err)

		// 删除业务配置
		err = s.svc.Svc.DeleteBusinessConfigByID(ctx, created.ID)
		assert.NoError(t, err)

		// 尝试获取已删除的业务配置
		_, err = s.svc.Svc.GetBusinessConfigByID(ctx, created.ID)
		assert.Error(t, err)
	})
}

// TestBusinessConfig_List 测试列出业务配置
func (s *BusinessConfigTestSuite) TestBusinessConfig_List() {
	t := s.T()
	ctx := context.Background()

	// 创建多个测试业务配置
	var createdIDs []int64
	for i := 0; i < 5; i++ {
		config := createTestBusinessConfig(fmt.Sprintf("测试列表-%d", i))
		created, err := s.svc.Svc.CreateBusinessConfig(ctx, config)
		require.NoError(t, err)
		createdIDs = append(createdIDs, created.ID)
	}

	// 列出业务配置
	configs, err := s.svc.Svc.ListBusinessConfigs(ctx, 0, 10)
	require.NoError(t, err)

	// 至少应该有我们刚创建的配置数量
	assert.GreaterOrEqual(t, len(configs), 5)

	// 测试结束后清理
	for _, id := range createdIDs {
		cleanupTest(t, ctx, func() error {
			return s.svc.Svc.DeleteBusinessConfigByID(ctx, id)
		})
	}
}
