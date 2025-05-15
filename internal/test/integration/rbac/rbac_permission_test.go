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

// PermissionTestSuite 权限测试套件
type PermissionTestSuite struct {
	suite.Suite
	db           *egorm.Component
	svc          *rbacioc.Service
	bizID        int64
	testResource domain.Resource
}

// SetupSuite 在所有测试之前设置
func (s *PermissionTestSuite) SetupSuite() {
	s.db = testioc.InitDBAndTables()
	s.svc = rbacioc.Init()

	// 创建测试业务，确保使用安全的bizID
	ctx := context.Background()
	bizConfig := createTestBusinessConfig("权限测试")
	created, err := s.svc.Svc.CreateBusinessConfig(ctx, bizConfig)
	if err != nil {
		s.T().Fatalf("创建测试业务失败: %v", err)
	}
	s.bizID = created.ID
	// 确保不是预设的bizID=1
	if s.bizID == 1 {
		s.T().Fatal("测试业务ID不应该是1，与预设数据冲突")
	}

	// 创建测试资源
	resource := createTestResource(s.bizID, "api", "/users")
	createdResource, err := s.svc.Svc.CreateResource(ctx, resource)
	if err != nil {
		s.T().Fatalf("创建测试资源失败: %v", err)
	}
	s.testResource = createdResource
	// 确保资源ID不在预设范围内
	if s.testResource.ID < TestResourceIDStart {
		s.T().Fatalf("测试资源ID应该大于等于%d，当前为%d", TestResourceIDStart, s.testResource.ID)
	}
}

// TearDownSuite 在所有测试之后清理
func (s *PermissionTestSuite) TearDownSuite() {
	// 清理所有测试数据
	ctx := context.Background()
	cleanTestEnvironment(s.T(), ctx, s.svc)
}

// TestPermissionSuite 运行权限测试套件
func TestPermissionSuite(t *testing.T) {
	suite.Run(t, new(PermissionTestSuite))
}

// TestPermission_Create 测试创建权限
func (s *PermissionTestSuite) TestPermission_Create() {
	t := s.T()
	ctx := context.Background()

	tests := []struct {
		name      string
		prepare   func() domain.Permission
		assertErr assert.ErrorAssertionFunc
		after     func(t *testing.T, permission, created domain.Permission)
	}{
		{
			name: "创建有效权限",
			prepare: func() domain.Permission {
				// 创建权限，添加时间戳确保唯一性
				permission := createTestPermission(s.bizID, s.testResource, ActionTypeRead)
				permission.Name = fmt.Sprintf("测试创建权限-%d", time.Now().UnixNano())
				return permission
			},
			assertErr: assert.NoError,
			after: func(t *testing.T, permission, created domain.Permission) {
				require.NotZero(t, created.ID)
				// 确保权限ID不是预设的ID范围
				assert.GreaterOrEqual(t, created.ID, TestPermissionIDStart)
				// 验证权限字段
				assertPermission(t, permission, created)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			permission := tt.prepare()
			created, err := s.svc.Svc.CreatePermission(ctx, permission)

			tt.assertErr(t, err)

			if err == nil {
				tt.after(t, permission, created)

				// 清理测试数据
				cleanupTest(t, ctx, func() error {
					return s.svc.Svc.DeletePermission(ctx, s.bizID, created.ID)
				})
			}
		})
	}
}

// TestPermission_Get 测试获取权限
func (s *PermissionTestSuite) TestPermission_Get() {
	t := s.T()
	ctx := context.Background()

	// 先删除之前可能留下的测试数据
	permissions, err := s.svc.Svc.ListPermissions(ctx, s.bizID, 0, 100)
	if err == nil {
		for _, perm := range permissions {
			_ = s.svc.Svc.DeletePermission(ctx, perm.BizID, perm.ID)
		}
	}

	// 创建一个新的权限用于测试
	permission := createTestPermission(s.bizID, s.testResource, ActionTypeRead)
	// 确保权限名称唯一
	permission.Name = fmt.Sprintf("测试获取权限-%d", time.Now().UnixNano())
	created, err := s.svc.Svc.CreatePermission(ctx, permission)
	require.NoError(t, err)

	tests := []struct {
		name      string
		bizID     int64
		permID    int64
		assertErr assert.ErrorAssertionFunc
		after     func(t *testing.T, found domain.Permission)
	}{
		{
			name:      "获取存在的权限",
			bizID:     s.bizID,
			permID:    created.ID,
			assertErr: assert.NoError,
			after: func(t *testing.T, found domain.Permission) {
				assertPermission(t, created, found)
			},
		},
		{
			name:      "获取不存在的权限",
			bizID:     s.bizID,
			permID:    99999,
			assertErr: assert.Error,
			after: func(t *testing.T, found domain.Permission) {
				// 不存在的权限，不需要验证
			},
		},
		{
			name:      "不匹配的业务ID",
			bizID:     s.bizID + 1,
			permID:    created.ID,
			assertErr: assert.Error,
			after: func(t *testing.T, found domain.Permission) {
				// 使用错误的业务ID，不需要验证
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			found, err := s.svc.Svc.GetPermission(ctx, tt.bizID, tt.permID)
			tt.assertErr(t, err)

			if err == nil {
				tt.after(t, found)
			}
		})
	}

	// 测试完成后删除权限
	err = s.svc.Svc.DeletePermission(ctx, s.bizID, created.ID)
	assert.NoError(t, err)
}

// TestPermission_Update 测试更新权限
func (s *PermissionTestSuite) TestPermission_Update() {
	t := s.T()
	ctx := context.Background()

	// 先确保清除之前的测试数据
	permissions, err := s.svc.Svc.ListPermissions(ctx, s.bizID, 0, 100)
	if err == nil {
		for _, perm := range permissions {
			// 避免删除测试资源的其他权限
			if perm.Resource.ID != s.testResource.ID {
				_ = s.svc.Svc.DeletePermission(ctx, perm.BizID, perm.ID)
			}
		}
	}

	// 为测试创建一个新的权限
	permission := createTestPermission(s.bizID, s.testResource, ActionTypeRead)
	// 确保权限名称唯一
	permission.Name = fmt.Sprintf("测试更新权限-%d", time.Now().UnixNano())
	created, err := s.svc.Svc.CreatePermission(ctx, permission)
	require.NoError(t, err)

	tests := []struct {
		name      string
		prepare   func() domain.Permission
		assertErr assert.ErrorAssertionFunc
		after     func(t *testing.T, updated, result domain.Permission)
	}{
		{
			name: "更新权限成功",
			prepare: func() domain.Permission {
				updated := created
				updated.Name = fmt.Sprintf("更新后的权限名称-%d", time.Now().UnixNano())
				updated.Description = "更新后的权限描述"
				return updated
			},
			assertErr: assert.NoError,
			after: func(t *testing.T, updated, result domain.Permission) {
				assert.Contains(t, result.Name, "更新后的权限名称")
				assert.Equal(t, "更新后的权限描述", result.Description)

				// 再次获取确认更新成功
				found, err := s.svc.Svc.GetPermission(ctx, s.bizID, result.ID)
				require.NoError(t, err)
				assertPermission(t, updated, found)

				// 清理
				err = s.svc.Svc.DeletePermission(ctx, s.bizID, result.ID)
				assert.NoError(t, err)
			},
		},
		{
			name: "更新不存在的权限",
			prepare: func() domain.Permission {
				nonExistentPermission := createTestPermission(s.bizID, s.testResource, ActionTypeWrite)
				nonExistentPermission.ID = 99999
				nonExistentPermission.Name = fmt.Sprintf("不存在的权限-%d", time.Now().UnixNano())
				return nonExistentPermission
			},
			assertErr: assert.NoError,
			after: func(t *testing.T, updated, result domain.Permission) {
				// 检查结果并清理可能创建的记录
				if result.ID > 0 {
					cleanupTest(t, ctx, func() error {
						return s.svc.Svc.DeletePermission(ctx, result.BizID, result.ID)
					})
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updated := tt.prepare()
			result, err := s.svc.Svc.UpdatePermission(ctx, updated)

			tt.assertErr(t, err)

			if err == nil {
				tt.after(t, updated, result)
			}
		})
	}

	t.Run("尝试更新资源和操作类型", func(t *testing.T) {
		// 创建一个新权限
		originalPerm := createTestPermission(s.bizID, s.testResource, ActionTypeRead)
		originalPerm.Name = fmt.Sprintf("测试资源操作更新-%d", time.Now().UnixNano())
		created, err := s.svc.Svc.CreatePermission(ctx, originalPerm)
		require.NoError(t, err)

		// 创建另一个资源
		anotherResource := createTestResource(s.bizID, "api", fmt.Sprintf("user:write:%d", time.Now().UnixNano()))
		createdResource, err := s.svc.Svc.CreateResource(ctx, anotherResource)
		require.NoError(t, err)

		// 尝试修改资源和操作类型
		updatePerm := created
		updatePerm.Resource = createdResource
		updatePerm.Action = string(ActionTypeWrite)
		updatePerm.Name = fmt.Sprintf("更新资源操作后-%d", time.Now().UnixNano())

		// 执行更新并验证结果
		result, err := s.svc.Svc.UpdatePermission(ctx, updatePerm)
		assert.NoError(t, err, "更新权限资源/操作类型不应返回错误")

		// 记录资源和操作类型是否成功更新
		t.Logf("权限资源/操作更新：原资源ID=%d, 新资源ID=%d, 原操作=%s, 新操作=%s",
			created.Resource.ID, result.Resource.ID, created.Action, result.Action)

		// 清理
		err = s.svc.Svc.DeletePermission(ctx, result.BizID, result.ID)
		assert.NoError(t, err)

		// 清理资源
		err = s.svc.Svc.DeleteResource(ctx, createdResource.BizID, createdResource.ID)
		assert.NoError(t, err)
	})
}

// TestPermission_Delete 测试删除权限
func (s *PermissionTestSuite) TestPermission_Delete() {
	t := s.T()
	ctx := context.Background()

	tests := []struct {
		name      string
		prepare   func() (int64, int64, error)
		assertErr assert.ErrorAssertionFunc
		after     func(t *testing.T, bizID, permID int64)
	}{
		{
			name: "删除存在的权限",
			prepare: func() (int64, int64, error) {
				// 创建一个权限
				permission := createTestPermission(s.bizID, s.testResource, ActionTypeRead)
				permission.Name = fmt.Sprintf("测试删除权限-%d", time.Now().UnixNano())
				created, err := s.svc.Svc.CreatePermission(ctx, permission)
				return s.bizID, created.ID, err
			},
			assertErr: assert.NoError,
			after: func(t *testing.T, bizID, permID int64) {
				// 尝试获取已删除的权限
				_, err := s.svc.Svc.GetPermission(ctx, bizID, permID)
				assert.Error(t, err)
			},
		},
		{
			name: "删除不存在的权限",
			prepare: func() (int64, int64, error) {
				return s.bizID, 99999, nil
			},
			assertErr: assert.NoError,
			after: func(t *testing.T, bizID, permID int64) {
				// 删除不存在的权限不需要额外检查
			},
		},
		{
			name: "不匹配的业务ID",
			prepare: func() (int64, int64, error) {
				// 创建一个权限
				permission := createTestPermission(s.bizID, s.testResource, ActionTypeRead)
				permission.Name = fmt.Sprintf("测试业务ID不匹配-%d", time.Now().UnixNano())
				created, err := s.svc.Svc.CreatePermission(ctx, permission)
				return s.bizID + 1, created.ID, err
			},
			assertErr: assert.NoError,
			after: func(t *testing.T, bizID, permID int64) {
				// 确认权限仍然存在
				found, err := s.svc.Svc.GetPermission(ctx, s.bizID, permID)
				assert.NoError(t, err, "权限不应被错误的业务ID删除")
				assert.Equal(t, permID, found.ID)

				// 清理
				cleanupTest(t, ctx, func() error {
					return s.svc.Svc.DeletePermission(ctx, s.bizID, permID)
				})
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bizID, permID, prepErr := tt.prepare()
			require.NoError(t, prepErr)

			err := s.svc.Svc.DeletePermission(ctx, bizID, permID)

			tt.assertErr(t, err)
			tt.after(t, bizID, permID)
		})
	}
}

// TestPermission_List 测试列出权限
func (s *PermissionTestSuite) TestPermission_List() {
	t := s.T()
	ctx := context.Background()

	// 创建的测试资源和权限
	var testResources []domain.Resource
	var testPermissions []domain.Permission

	// 准备测试数据 - 创建多个不同资源类型和操作的权限
	prepareTestPermissions := func() {
		// 删除之前的数据
		permissions, err := s.svc.Svc.ListPermissions(ctx, s.bizID, 0, 1000)
		if err == nil {
			for _, perm := range permissions {
				_ = s.svc.Svc.DeletePermission(ctx, perm.BizID, perm.ID)
			}
		}

		resources, err := s.svc.Svc.ListResources(ctx, s.bizID, 0, 1000)
		if err == nil {
			for _, res := range resources {
				_ = s.svc.Svc.DeleteResource(ctx, res.BizID, res.ID)
			}
		}

		// 创建不同类型的资源
		resourceTypes := []string{"api", "menu"}

		for _, resType := range resourceTypes {
			for i := 0; i < 2; i++ {
				resource := createTestResource(s.bizID, resType, fmt.Sprintf("%s:test%d:%d", resType, i, time.Now().UnixNano()))
				created, err := s.svc.Svc.CreateResource(ctx, resource)
				require.NoError(t, err)
				testResources = append(testResources, created)
			}
		}

		// 对每个资源创建不同操作的权限
		actionTypes := []ActionType{
			ActionTypeRead,
			ActionTypeWrite,
		}

		for _, resource := range testResources {
			for _, actionType := range actionTypes {
				permission := createTestPermission(s.bizID, resource, actionType)
				// 添加时间戳确保唯一性
				permission.Name = fmt.Sprintf("列表测试权限-%s-%s-%d", resource.Type, actionType, time.Now().UnixNano())
				created, err := s.svc.Svc.CreatePermission(ctx, permission)
				require.NoError(t, err)
				testPermissions = append(testPermissions, created)
			}
		}

		// 为不同业务ID创建权限
		otherBizResource := createTestResource(s.bizID+1, "api", fmt.Sprintf("api:other-biz:%d", time.Now().UnixNano()))
		createdResource, err := s.svc.Svc.CreateResource(ctx, otherBizResource)
		require.NoError(t, err)
		testResources = append(testResources, createdResource)

		otherBizPermission := createTestPermission(s.bizID+1, createdResource, ActionTypeRead)
		otherBizPermission.Name = fmt.Sprintf("其他业务权限-%d", time.Now().UnixNano())
		created, err := s.svc.Svc.CreatePermission(ctx, otherBizPermission)
		require.NoError(t, err)
		testPermissions = append(testPermissions, created)
	}

	// 创建测试权限数据
	prepareTestPermissions()

	// 确保创建了足够的测试数据
	t.Logf("创建了 %d 个测试资源和 %d 个测试权限", len(testResources), len(testPermissions))

	// 计算每种类型的权限数量
	apiPermCount := 0
	menuPermCount := 0
	readPermCount := 0
	for _, perm := range testPermissions {
		if perm.Resource.Type == "api" && perm.BizID == s.bizID {
			apiPermCount++
		}
		if perm.Resource.Type == "menu" && perm.BizID == s.bizID {
			menuPermCount++
		}
		if perm.Action == string(ActionTypeRead) && perm.BizID == s.bizID {
			readPermCount++
		}
	}

	t.Logf("API资源权限: %d, 菜单资源权限: %d, 读操作权限: %d", apiPermCount, menuPermCount, readPermCount)

	tests := []struct {
		name      string
		listFunc  func() ([]domain.Permission, error)
		assertErr assert.ErrorAssertionFunc
		after     func(t *testing.T, permissions []domain.Permission)
	}{
		{
			name: "列出所有权限",
			listFunc: func() ([]domain.Permission, error) {
				return s.svc.Svc.ListPermissions(ctx, s.bizID, 0, 20)
			},
			assertErr: assert.NoError,
			after: func(t *testing.T, permissions []domain.Permission) {
				assert.GreaterOrEqual(t, len(permissions), 8) // 至少8个权限(4个资源×2种操作)
				t.Logf("查询到 %d 个权限", len(permissions))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			permissions, err := tt.listFunc()

			tt.assertErr(t, err)

			if err == nil {
				tt.after(t, permissions)
			}
		})
	}

	// 测试结束后清理所有创建的测试数据
	for _, perm := range testPermissions {
		_ = s.svc.Svc.DeletePermission(ctx, perm.BizID, perm.ID)
	}

	for _, res := range testResources {
		_ = s.svc.Svc.DeleteResource(ctx, res.BizID, res.ID)
	}
}
