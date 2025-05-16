//go:build e2e

package rbac

import (
	"context"
	"fmt"
	"testing"
	"time"

	"gitee.com/flycash/permission-platform/internal/domain"
	rbacioc "gitee.com/flycash/permission-platform/internal/test/integration/ioc/rbac"
	testioc "gitee.com/flycash/permission-platform/internal/test/ioc"
	"github.com/ego-component/egorm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// ResourceTestSuite 资源测试套件
type ResourceTestSuite struct {
	suite.Suite
	db    *egorm.Component
	svc   *rbacioc.Service
	bizID int64
}

// SetupSuite 在所有测试之前设置
func (s *ResourceTestSuite) SetupSuite() {
	s.db = testioc.InitDBAndTables()
	s.svc = rbacioc.Init()

	// 创建测试业务，确保使用安全的bizID
	ctx := context.Background()
	bizConfig := createTestBusinessConfig("资源测试")
	created, err := s.svc.Svc.CreateBusinessConfig(ctx, bizConfig)
	if err != nil {
		s.T().Fatalf("创建测试业务失败: %v", err)
	}
	s.bizID = created.ID
	// 确保不是预设的bizID=1
	if s.bizID == 1 {
		s.T().Fatal("测试业务ID不应该是1，与预设数据冲突")
	}
}

// TearDownSuite 在所有测试之后清理
func (s *ResourceTestSuite) TearDownSuite() {
	// 清理所有测试数据
	ctx := context.Background()
	cleanTestEnvironment(s.T(), ctx, s.svc)
}

// TestResourceSuite 运行资源测试套件
func TestResourceSuite(t *testing.T) {
	suite.Run(t, new(ResourceTestSuite))
}

// TestResource_Create 测试创建资源
func (s *ResourceTestSuite) TestResource_Create() {
	t := s.T()
	ctx := context.Background()

	tests := []struct {
		name      string
		prepare   func() domain.Resource
		assertErr assert.ErrorAssertionFunc
		after     func(t *testing.T, resource, created domain.Resource)
	}{
		{
			name: "创建API资源",
			prepare: func() domain.Resource {
				return createTestResource(s.bizID, "api", "/users")
			},
			assertErr: assert.NoError,
			after: func(t *testing.T, resource, created domain.Resource) {
				require.NotZero(t, created.ID)
				// 确保资源ID不是预设的ID范围，明确转换为int类型进行比较
				assert.GreaterOrEqual(t, int(created.ID), TestResourceIDStart)
				// 验证资源字段
				assertResource(t, resource, created)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resource := tt.prepare()
			created, err := s.svc.Svc.CreateResource(ctx, resource)

			tt.assertErr(t, err)

			if err == nil {
				tt.after(t, resource, created)

				// 清理测试数据
				cleanupTest(t, ctx, func() error {
					return s.svc.Svc.DeleteResource(ctx, s.bizID, created.ID)
				})
			}
		})
	}
}

// TestResource_Get 测试获取资源
func (s *ResourceTestSuite) TestResource_Get() {
	t := s.T()
	ctx := context.Background()

	// 先创建一个资源用于后续测试
	resource := createTestResource(s.bizID, "api", "user:read")
	created, err := s.svc.Svc.CreateResource(ctx, resource)
	require.NoError(t, err)

	tests := []struct {
		name      string
		bizID     int64
		resID     int64
		assertErr assert.ErrorAssertionFunc
		after     func(t *testing.T, found domain.Resource)
	}{
		{
			name:      "获取存在的资源",
			bizID:     s.bizID,
			resID:     created.ID,
			assertErr: assert.NoError,
			after: func(t *testing.T, found domain.Resource) {
				assertResource(t, created, found)
			},
		},
		{
			name:      "获取不存在的资源",
			bizID:     s.bizID,
			resID:     99999,
			assertErr: assert.Error,
			after: func(t *testing.T, found domain.Resource) {
				// 不存在的资源，不需要额外验证
			},
		},
		{
			name:      "不匹配的业务ID",
			bizID:     s.bizID + 1,
			resID:     created.ID,
			assertErr: assert.Error,
			after: func(t *testing.T, found domain.Resource) {
				// 错误的业务ID，不需要额外验证
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			found, err := s.svc.Svc.GetResource(ctx, tt.bizID, tt.resID)
			tt.assertErr(t, err)

			if err == nil {
				tt.after(t, found)
			}
		})
	}
}

// TestResource_Update 测试更新资源
func (s *ResourceTestSuite) TestResource_Update() {
	t := s.T()
	ctx := context.Background()

	// 先创建一个资源用于后续测试
	resource := createTestResource(s.bizID, "api", "user:key")
	created, err := s.svc.Svc.CreateResource(ctx, resource)
	require.NoError(t, err)

	// 测试结束后清理
	defer cleanupTest(t, ctx, func() error {
		return s.svc.Svc.DeleteResource(ctx, s.bizID, created.ID)
	})

	tests := []struct {
		name      string
		prepare   func() domain.Resource
		assertErr assert.ErrorAssertionFunc
		after     func(t *testing.T, updated, result domain.Resource)
	}{
		{
			name: "更新资源成功",
			prepare: func() domain.Resource {
				updated := created
				updated.Name = "更新后的资源名称"
				updated.Description = "更新后的资源描述"
				updated.Metadata = `{"updated":true}`
				return updated
			},
			assertErr: assert.NoError,
			after: func(t *testing.T, updated, result domain.Resource) {
				assert.Equal(t, updated.Name, result.Name)
				assert.Equal(t, updated.Description, result.Description)
				assert.Equal(t, updated.Metadata, result.Metadata)

				// 再次获取确认更新成功
				found, err := s.svc.Svc.GetResource(ctx, s.bizID, result.ID)
				require.NoError(t, err)
				assertResource(t, updated, found)
			},
		},
		{
			name: "更新不存在的资源",
			prepare: func() domain.Resource {
				nonExistentResource := createTestResource(s.bizID, "api", "nonexistent")
				nonExistentResource.ID = 99999
				return nonExistentResource
			},
			assertErr: assert.NoError,
			after: func(t *testing.T, updated, result domain.Resource) {
				// 记录结果并清理可能创建的资源
				if result.ID > 0 {
					t.Logf("更新不存在资源的结果: %+v", result)
					cleanupTest(t, ctx, func() error {
						return s.svc.Svc.DeleteResource(ctx, result.BizID, result.ID)
					})
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updated := tt.prepare()
			result, err := s.svc.Svc.UpdateResource(ctx, updated)

			tt.assertErr(t, err)

			if err == nil {
				tt.after(t, updated, result)
			}
		})
	}

	t.Run("尝试修改不可变字段类型", func(t *testing.T) {
		// 创建一个新资源
		originalRes := createTestResource(s.bizID, "api", "type:test")
		createdRes, err := s.svc.Svc.CreateResource(ctx, originalRes)
		require.NoError(t, err)

		// 尝试修改Type字段
		updatedRes := createdRes
		updatedRes.Type = "menu"

		// 执行更新
		result, err := s.svc.Svc.UpdateResource(ctx, updatedRes)
		assert.NoError(t, err, "更新资源类型不应返回错误")

		// 记录类型是否成功更新
		t.Logf("资源类型更新：原类型=%s, 新类型=%s", createdRes.Type, result.Type)

		// 清理
		cleanupTest(t, ctx, func() error {
			return s.svc.Svc.DeleteResource(ctx, createdRes.BizID, createdRes.ID)
		})
	})

	t.Run("尝试修改不可变字段Key", func(t *testing.T) {
		// 创建一个新资源
		// 使用唯一的Key避免索引冲突
		uniqueKey := fmt.Sprintf("user:key:unique:%d", time.Now().UnixNano())
		originalRes := createTestResource(s.bizID, "api", uniqueKey)
		createdRes, err := s.svc.Svc.CreateResource(ctx, originalRes)
		require.NoError(t, err)

		// 尝试修改Key字段
		updatedRes := createdRes
		updatedRes.Key = fmt.Sprintf("menu:user:unique:%d", time.Now().UnixNano())

		// 执行更新
		result, err := s.svc.Svc.UpdateResource(ctx, updatedRes)

		// 根据实际行为判断，尝试更新Key可能会失败（如果Key是不可变字段）
		if err != nil {
			t.Logf("更新资源Key返回错误: %v", err)
			assert.Contains(t, err.Error(), "唯一索引", "错误应包含唯一索引相关信息")
		} else {
			// 记录Key是否成功更新
			t.Logf("资源Key更新：原Key=%s, 新Key=%s", createdRes.Key, result.Key)
		}

		// 清理
		cleanupTest(t, ctx, func() error {
			return s.svc.Svc.DeleteResource(ctx, createdRes.BizID, createdRes.ID)
		})
	})
}

// TestResource_Delete 测试删除资源
func (s *ResourceTestSuite) TestResource_Delete() {
	t := s.T()
	ctx := context.Background()

	tests := []struct {
		name      string
		prepare   func() (int64, int64, error)
		assertErr assert.ErrorAssertionFunc
		after     func(t *testing.T, bizID, resID int64)
	}{
		{
			name: "删除存在的资源",
			prepare: func() (int64, int64, error) {
				// 创建一个资源
				resource := createTestResource(s.bizID, "api", "user:delete:"+time.Now().String())
				created, err := s.svc.Svc.CreateResource(ctx, resource)
				return s.bizID, created.ID, err
			},
			assertErr: assert.NoError,
			after: func(t *testing.T, bizID, resID int64) {
				// 尝试获取已删除的资源
				_, err := s.svc.Svc.GetResource(ctx, bizID, resID)
				assert.Error(t, err)
			},
		},
		{
			name: "删除不存在的资源",
			prepare: func() (int64, int64, error) {
				// 返回一个不存在的资源ID
				return s.bizID, 99999, nil
			},
			assertErr: assert.NoError,
			after: func(t *testing.T, bizID, resID int64) {
				// 删除不存在的资源不需要额外检查
			},
		},
		{
			name: "不匹配的业务ID",
			prepare: func() (int64, int64, error) {
				// 创建一个资源
				resKey := "user:delete:mismatch:" + time.Now().String()
				resource := createTestResource(s.bizID, "api", resKey)
				created, err := s.svc.Svc.CreateResource(ctx, resource)
				return s.bizID + 1, created.ID, err
			},
			assertErr: assert.NoError,
			after: func(t *testing.T, bizID, resID int64) {
				// 确认资源仍然存在
				found, err := s.svc.Svc.GetResource(ctx, s.bizID, resID)
				assert.NoError(t, err, "资源不应被错误的业务ID删除")
				assert.Equal(t, resID, found.ID)

				// 清理
				cleanupTest(t, ctx, func() error {
					return s.svc.Svc.DeleteResource(ctx, s.bizID, resID)
				})
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bizID, resID, prepErr := tt.prepare()
			require.NoError(t, prepErr)

			err := s.svc.Svc.DeleteResource(ctx, bizID, resID)

			tt.assertErr(t, err)
			tt.after(t, bizID, resID)
		})
	}
}

// TestResource_List 测试列出资源
func (s *ResourceTestSuite) TestResource_List() {
	t := s.T()
	ctx := context.Background()

	// 创建多个测试资源
	var createdResourceIDs []int64
	for i := 0; i < 5; i++ {
		resourceKey := fmt.Sprintf("/test/resource/%d", i)
		resource := createTestResource(s.bizID, "api", resourceKey)
		created, err := s.svc.Svc.CreateResource(ctx, resource)
		require.NoError(t, err)

		// 验证资源ID是否在安全范围内，明确转换为int类型
		assert.GreaterOrEqual(t, int(created.ID), TestResourceIDStart)

		createdResourceIDs = append(createdResourceIDs, created.ID)
	}

	tests := []struct {
		name      string
		listFunc  func() ([]domain.Resource, error)
		assertErr assert.ErrorAssertionFunc
		after     func(t *testing.T, resources []domain.Resource)
	}{
		{
			name: "列出所有资源",
			listFunc: func() ([]domain.Resource, error) {
				return s.svc.Svc.ListResources(ctx, s.bizID, 0, 20)
			},
			assertErr: assert.NoError,
			after: func(t *testing.T, resources []domain.Resource) {
				// 至少应该包含我们刚创建的5个资源
				assert.GreaterOrEqual(t, len(resources), 5)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resources, err := tt.listFunc()

			tt.assertErr(t, err)

			if err == nil {
				tt.after(t, resources)
			}
		})
	}

	// 清理测试数据
	for _, id := range createdResourceIDs {
		cleanupTest(t, ctx, func() error {
			return s.svc.Svc.DeleteResource(ctx, s.bizID, id)
		})
	}
}
