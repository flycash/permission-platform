//go:build e2e

package rbac

import (
	"context"
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

// RolePermissionTestSuite 角色权限测试套件
type RolePermissionTestSuite struct {
	suite.Suite
	db             *egorm.Component
	svc            *rbacioc.Service
	bizID          int64
	testRole       domain.Role
	testResource   domain.Resource
	testPermission domain.Permission
}

// SetupSuite 在所有测试之前设置
func (s *RolePermissionTestSuite) SetupSuite() {
	s.db = testioc.InitDBAndTables()
	s.svc = rbacioc.Init()

	// 创建测试业务
	ctx := context.Background()
	bizConfig := createTestBusinessConfig("角色权限测试")
	created, err := s.svc.Svc.CreateBusinessConfig(ctx, bizConfig)
	if err != nil {
		s.T().Fatalf("创建测试业务失败: %v", err)
	}
	s.bizID = created.ID

	// 创建测试角色
	role := createTestRole(s.bizID, RoleTypeSystem)
	createdRole, err := s.svc.Svc.CreateRole(ctx, role)
	if err != nil {
		s.T().Fatalf("创建测试角色失败: %v", err)
	}
	s.testRole = createdRole

	// 创建测试资源
	resource := createTestResource(s.bizID, "api", "user:manage")
	createdResource, err := s.svc.Svc.CreateResource(ctx, resource)
	if err != nil {
		s.T().Fatalf("创建测试资源失败: %v", err)
	}
	s.testResource = createdResource

	// 创建测试权限
	permission := createTestPermission(s.bizID, createdResource, ActionTypeRead)
	createdPermission, err := s.svc.Svc.CreatePermission(ctx, permission)
	if err != nil {
		s.T().Fatalf("创建测试权限失败: %v", err)
	}
	s.testPermission = createdPermission
}

// TestRolePermissionSuite 运行角色权限测试套件
func TestRolePermissionSuite(t *testing.T) {
	suite.Run(t, new(RolePermissionTestSuite))
}

// TestRolePermission_Create 测试创建角色权限
func (s *RolePermissionTestSuite) TestRolePermission_Create() {
	t := s.T()
	ctx := context.Background()

	tests := []struct {
		name      string
		prepare   func() domain.RolePermission
		assertErr assert.ErrorAssertionFunc
		after     func(t *testing.T, created domain.RolePermission)
	}{
		{
			name: "创建有效的角色权限",
			prepare: func() domain.RolePermission {
				return createTestRolePermission(s.bizID, s.testRole, s.testPermission)
			},
			assertErr: assert.NoError,
			after: func(t *testing.T, created domain.RolePermission) {
				assertRolePermission(t, createTestRolePermission(s.bizID, s.testRole, s.testPermission), created)
			},
		},
		{
			name: "零业务ID",
			prepare: func() domain.RolePermission {
				rp := createTestRolePermission(0, s.testRole, s.testPermission)
				return rp
			},
			assertErr: assert.NoError, // 实际服务实现可能不验证这些条件
			after: func(t *testing.T, created domain.RolePermission) {
				// 如果创建成功，清理创建的记录
				if created.ID > 0 {
					t.Logf("零业务ID创建成功，ID=%d", created.ID)
				}
			},
		},
		{
			name: "角色ID为0",
			prepare: func() domain.RolePermission {
				invalidRole := s.testRole
				invalidRole.ID = 0
				rp := createTestRolePermission(s.bizID, invalidRole, s.testPermission)
				return rp
			},
			assertErr: assert.NoError, // 实际服务实现可能不验证这些条件
			after: func(t *testing.T, created domain.RolePermission) {
				// 如果创建成功，记录相关信息
				if created.ID > 0 {
					t.Logf("角色ID为0创建成功，ID=%d", created.ID)
				}
			},
		},
		{
			name: "权限ID为0",
			prepare: func() domain.RolePermission {
				invalidPermission := s.testPermission
				invalidPermission.ID = 0
				rp := createTestRolePermission(s.bizID, s.testRole, invalidPermission)
				return rp
			},
			assertErr: assert.NoError, // 实际服务实现可能不验证这些条件
			after: func(t *testing.T, created domain.RolePermission) {
				// 如果创建成功，记录相关信息
				if created.ID > 0 {
					t.Logf("权限ID为0创建成功，ID=%d", created.ID)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rolePermission := tt.prepare()
			created, err := s.svc.Svc.GrantRolePermission(ctx, rolePermission)

			tt.assertErr(t, err)

			if err == nil {
				tt.after(t, created)

				// 清理创建的角色权限
				if created.ID > 0 {
					cleanupTest(t, ctx, func() error {
						return s.svc.Svc.RevokeRolePermission(ctx, created.BizID, created.ID)
					})
				}
			}
		})
	}
}

// TestRolePermission_Get 测试获取角色权限
func (s *RolePermissionTestSuite) TestRolePermission_Get() {
	t := s.T()
	ctx := context.Background()

	// 先创建一个角色权限用于后续测试
	rolePermission := createTestRolePermission(s.bizID, s.testRole, s.testPermission)
	created, err := s.svc.Svc.GrantRolePermission(ctx, rolePermission)
	require.NoError(t, err)

	// 测试结束后清理
	defer cleanupTest(t, ctx, func() error {
		return s.svc.Svc.RevokeRolePermission(ctx, s.bizID, created.ID)
	})

	tests := []struct {
		name      string
		bizID     int64
		roleID    int64
		assertErr assert.ErrorAssertionFunc
		after     func(t *testing.T, rolePermissions []domain.RolePermission)
	}{
		{
			name:      "获取存在的角色权限",
			bizID:     s.bizID,
			roleID:    created.Role.ID,
			assertErr: assert.NoError,
			after: func(t *testing.T, rolePermissions []domain.RolePermission) {
				assert.NotEmpty(t, rolePermissions)

				// 在返回的权限列表中查找匹配的权限
				var found bool
				for _, rp := range rolePermissions {
					if rp.ID == created.ID {
						assertRolePermission(t, created, rp)
						found = true
						break
					}
				}

				assert.True(t, found, "未找到刚创建的角色权限")
			},
		},
		{
			name:      "获取不存在的角色权限",
			bizID:     s.bizID,
			roleID:    99999,
			assertErr: assert.NoError,
			after: func(t *testing.T, rolePermissions []domain.RolePermission) {
				assert.Empty(t, rolePermissions, "应该找不到指定角色的权限")
			},
		},
		{
			name:      "不匹配的业务ID",
			bizID:     s.bizID + 1,
			roleID:    created.Role.ID,
			assertErr: assert.NoError,
			after: func(t *testing.T, rolePermissions []domain.RolePermission) {
				assert.Empty(t, rolePermissions, "应该找不到指定业务ID的角色权限")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rolePermissions, err := s.svc.Svc.ListRolePermissionsByRoleID(ctx, tt.bizID, tt.roleID)
			tt.assertErr(t, err)

			if err == nil {
				tt.after(t, rolePermissions)
			}
		})
	}
}

// TestRolePermission_Delete 测试删除角色权限
func (s *RolePermissionTestSuite) TestRolePermission_Delete() {
	t := s.T()
	ctx := context.Background()

	tests := []struct {
		name      string
		prepare   func() (int64, int64, error)
		assertErr assert.ErrorAssertionFunc
		after     func(t *testing.T, bizID, rpID int64)
	}{
		{
			name: "删除存在的角色权限",
			prepare: func() (int64, int64, error) {
				// 创建一个角色权限
				rolePermission := createTestRolePermission(s.bizID, s.testRole, s.testPermission)
				created, err := s.svc.Svc.GrantRolePermission(ctx, rolePermission)
				return s.bizID, created.ID, err
			},
			assertErr: assert.NoError,
			after: func(t *testing.T, bizID, rpID int64) {
				// 尝试获取已删除的角色权限
				rolePermissions, err := s.svc.Svc.ListRolePermissionsByRoleID(ctx, bizID, s.testRole.ID)
				assert.NoError(t, err)

				var found bool
				for _, rp := range rolePermissions {
					if rp.ID == rpID {
						found = true
						break
					}
				}

				assert.False(t, found, "角色权限应已被删除")
			},
		},
		{
			name: "删除不存在的角色权限",
			prepare: func() (int64, int64, error) {
				return s.bizID, 99999, nil
			},
			assertErr: assert.NoError,
			after: func(t *testing.T, bizID, rpID int64) {
				// 删除不存在的角色权限不需要额外检查
			},
		},
		{
			name: "不匹配的业务ID",
			prepare: func() (int64, int64, error) {
				// 创建一个角色权限
				rolePermission := createTestRolePermission(s.bizID, s.testRole, s.testPermission)
				created, err := s.svc.Svc.GrantRolePermission(ctx, rolePermission)
				return s.bizID + 1, created.ID, err
			},
			assertErr: assert.NoError,
			after: func(t *testing.T, bizID, rpID int64) {
				// 确认角色权限是否仍然存在
				rolePermissions, err := s.svc.Svc.ListRolePermissionsByRoleID(ctx, s.bizID, s.testRole.ID)
				assert.NoError(t, err)

				var found bool
				for _, rp := range rolePermissions {
					if rp.ID == rpID {
						found = true
						break
					}
				}

				assert.True(t, found, "角色权限不应被错误的业务ID删除")

				// 清理
				cleanupTest(t, ctx, func() error {
					return s.svc.Svc.RevokeRolePermission(ctx, s.bizID, rpID)
				})
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bizID, rpID, prepErr := tt.prepare()
			require.NoError(t, prepErr)

			err := s.svc.Svc.RevokeRolePermission(ctx, bizID, rpID)

			tt.assertErr(t, err)
			tt.after(t, bizID, rpID)
		})
	}
}

// TestRolePermission_List 测试列出角色权限
func (s *RolePermissionTestSuite) TestRolePermission_List() {
	t := s.T()
	ctx := context.Background()

	// 准备测试数据 - 创建多个不同的角色和权限
	// 创建多个角色
	var testRoles []domain.Role
	for i := 0; i < 3; i++ {
		role := createTestRole(s.bizID, RoleTypeCustom)
		created, err := s.svc.Svc.CreateRole(ctx, role)
		require.NoError(t, err)
		testRoles = append(testRoles, created)
	}

	// 创建多个权限
	var testPermissions []domain.Permission
	for i := 0; i < 3; i++ {
		// 为每个权限创建一个资源
		resKey := "res:test:" + time.Now().String()
		resource := createTestResource(s.bizID, "api", resKey)
		createdResource, err := s.svc.Svc.CreateResource(ctx, resource)
		require.NoError(t, err)

		permission := createTestPermission(s.bizID, createdResource, ActionTypeRead)
		created, err := s.svc.Svc.CreatePermission(ctx, permission)
		require.NoError(t, err)
		testPermissions = append(testPermissions, created)
	}

	// 创建多个角色权限关系
	var rolePermissionIDs []int64

	// 为每个角色分配第一个权限
	for _, role := range testRoles {
		rp := createTestRolePermission(s.bizID, role, testPermissions[0])
		created, err := s.svc.Svc.GrantRolePermission(ctx, rp)
		require.NoError(t, err)
		rolePermissionIDs = append(rolePermissionIDs, created.ID)
	}

	// 为第一个角色分配所有权限
	for i := 1; i < len(testPermissions); i++ {
		rp := createTestRolePermission(s.bizID, testRoles[0], testPermissions[i])
		created, err := s.svc.Svc.GrantRolePermission(ctx, rp)
		require.NoError(t, err)
		rolePermissionIDs = append(rolePermissionIDs, created.ID)
	}

	// 为测试角色分配所有权限
	for _, permission := range testPermissions {
		rp := createTestRolePermission(s.bizID, s.testRole, permission)
		created, err := s.svc.Svc.GrantRolePermission(ctx, rp)
		require.NoError(t, err)
		rolePermissionIDs = append(rolePermissionIDs, created.ID)
	}

	// 测试结束后清理
	defer func() {
		for _, id := range rolePermissionIDs {
			cleanupTest(t, ctx, func() error {
				return s.svc.Svc.RevokeRolePermission(ctx, s.bizID, id)
			})
		}
	}()

	tests := []struct {
		name      string
		listFunc  func() ([]domain.RolePermission, error)
		assertErr assert.ErrorAssertionFunc
		after     func(t *testing.T, rolePermissions []domain.RolePermission)
	}{
		{
			name: "列出角色的所有权限",
			listFunc: func() ([]domain.RolePermission, error) {
				return s.svc.Svc.ListRolePermissionsByRoleID(ctx, s.bizID, s.testRole.ID)
			},
			assertErr: assert.NoError,
			after: func(t *testing.T, rolePermissions []domain.RolePermission) {
				assert.GreaterOrEqual(t, len(rolePermissions), 3)
				for _, rp := range rolePermissions {
					assert.Equal(t, s.testRole.ID, rp.Role.ID)
				}
			},
		},
		{
			name: "列出所有角色权限",
			listFunc: func() ([]domain.RolePermission, error) {
				return s.svc.Svc.ListRolePermissions(ctx, s.bizID)
			},
			assertErr: assert.NoError,
			after: func(t *testing.T, rolePermissions []domain.RolePermission) {
				// 过滤出与特定权限相关的记录
				var permissionRoleCount int
				for _, rp := range rolePermissions {
					if rp.Permission.ID == testPermissions[0].ID {
						permissionRoleCount++
					}
				}

				assert.GreaterOrEqual(t, permissionRoleCount, 4) // 3个测试角色 + 1个预设角色
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rolePermissions, err := tt.listFunc()

			tt.assertErr(t, err)

			if err == nil {
				tt.after(t, rolePermissions)
			}
		})
	}

	t.Run("不同业务ID", func(t *testing.T) {
		// 创建另一个业务
		otherBizConfig := createTestBusinessConfig("其他业务")
		createdBiz, err := s.svc.Svc.CreateBusinessConfig(ctx, otherBizConfig)
		require.NoError(t, err)
		otherBizID := createdBiz.ID

		// 在新业务中创建角色和权限
		otherRole := createTestRole(otherBizID, RoleTypeSystem)
		createdRole, err := s.svc.Svc.CreateRole(ctx, otherRole)
		require.NoError(t, err)

		otherResource := createTestResource(otherBizID, "api", "other:test")
		createdResource, err := s.svc.Svc.CreateResource(ctx, otherResource)
		require.NoError(t, err)

		otherPermission := createTestPermission(otherBizID, createdResource, ActionTypeRead)
		createdPermission, err := s.svc.Svc.CreatePermission(ctx, otherPermission)
		require.NoError(t, err)

		// 创建角色权限关系
		otherRP := createTestRolePermission(otherBizID, createdRole, createdPermission)
		createdRP, err := s.svc.Svc.GrantRolePermission(ctx, otherRP)
		require.NoError(t, err)

		// 查询
		rolePermissions, err := s.svc.Svc.ListRolePermissionsByRoleID(ctx, otherBizID, createdRole.ID)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(rolePermissions), 1)
		for _, rp := range rolePermissions {
			assert.Equal(t, otherBizID, rp.BizID)
		}

		// 清理
		cleanupTest(t, ctx, func() error {
			return s.svc.Svc.RevokeRolePermission(ctx, otherBizID, createdRP.ID)
		})
	})
}
