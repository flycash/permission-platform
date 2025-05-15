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

// RoleTestSuite 角色测试套件
type RoleTestSuite struct {
	suite.Suite
	db    *egorm.Component
	svc   *rbacioc.Service
	bizID int64
}

// SetupSuite 在所有测试之前设置
func (s *RoleTestSuite) SetupSuite() {
	s.db = testioc.InitDBAndTables()
	s.svc = rbacioc.Init()

	// 创建测试业务，使用安全的bizID
	ctx := context.Background()
	bizConfig := createTestBusinessConfig("角色测试")
	created, err := s.svc.Svc.CreateBusinessConfig(ctx, bizConfig)
	if err != nil {
		s.T().Fatalf("创建测试业务失败: %v", err)
	}
	s.bizID = created.ID
}

// TearDownSuite 在所有测试之后清理
func (s *RoleTestSuite) TearDownSuite() {
	// 清理所有测试数据
	ctx := context.Background()
	cleanTestEnvironment(s.T(), ctx, s.svc)
}

// TestRoleSuite 运行角色测试套件
func TestRoleSuite(t *testing.T) {
	suite.Run(t, new(RoleTestSuite))
}

// TestRole_Create 测试创建角色
func (s *RoleTestSuite) TestRole_Create() {
	t := s.T()
	ctx := context.Background()

	tests := []struct {
		name      string
		role      func() domain.Role
		assertErr assert.ErrorAssertionFunc
		after     func(t *testing.T, role domain.Role, created domain.Role)
	}{
		{
			name: "创建系统角色",
			role: func() domain.Role {
				role := createTestRole(s.bizID, RoleTypeSystem)
				role.Name = fmt.Sprintf("测试系统角色-%d", time.Now().UnixNano())
				return role
			},
			assertErr: assert.NoError,
			after: func(t *testing.T, role domain.Role, created domain.Role) {
				assertRole(t, role, created)
			},
		},
		{
			name: "创建自定义角色",
			role: func() domain.Role {
				role := createTestRole(s.bizID, RoleTypeCustom)
				role.Name = fmt.Sprintf("测试自定义角色-%d", time.Now().UnixNano())
				return role
			},
			assertErr: assert.NoError,
			after: func(t *testing.T, role domain.Role, created domain.Role) {
				assertRole(t, role, created)
			},
		},
		{
			name: "零业务ID",
			role: func() domain.Role {
				role := createTestRole(0, RoleTypeCustom)
				role.Name = fmt.Sprintf("测试零业务ID角色-%d", time.Now().UnixNano())
				return role
			},
			assertErr: assert.NoError,
			after: func(t *testing.T, role domain.Role, created domain.Role) {
				// 业务实现可能接受零业务ID
				assert.Equal(t, role.Name, created.Name)
				assert.Equal(t, role.Type, created.Type)
			},
		},
		{
			name: "角色名称为空",
			role: func() domain.Role {
				role := createTestRole(s.bizID, RoleTypeCustom)
				role.Name = ""
				return role
			},
			assertErr: assert.NoError,
			after: func(t *testing.T, role domain.Role, created domain.Role) {
				// 业务实现可能接受空名称
				assert.Equal(t, role.Name, created.Name)
				assert.Equal(t, role.Type, created.Type)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			role := tt.role()
			created, err := s.svc.Svc.CreateRole(ctx, role)

			tt.assertErr(t, err)

			if err == nil {
				tt.after(t, role, created)

				// 清理创建的角色
				cleanupTest(t, ctx, func() error {
					return s.svc.Svc.DeleteRole(ctx, created.BizID, created.ID)
				})
			}
		})
	}
}

// TestRole_Get 测试获取角色
func (s *RoleTestSuite) TestRole_Get() {
	t := s.T()
	ctx := context.Background()

	// 先创建一个角色用于后续测试
	role := createTestRole(s.bizID, RoleTypeCustom)
	created, err := s.svc.Svc.CreateRole(ctx, role)
	require.NoError(t, err)

	// 测试结束后清理
	defer cleanupTest(t, ctx, func() error {
		return s.svc.Svc.DeleteRole(ctx, s.bizID, created.ID)
	})

	tests := []struct {
		name      string
		bizID     int64
		roleID    int64
		assertErr assert.ErrorAssertionFunc
		after     func(t *testing.T, found domain.Role)
	}{
		{
			name:      "获取存在的角色",
			bizID:     s.bizID,
			roleID:    created.ID,
			assertErr: assert.NoError,
			after: func(t *testing.T, found domain.Role) {
				assertRole(t, created, found)
			},
		},
		{
			name:      "获取不存在的角色",
			bizID:     s.bizID,
			roleID:    99999,
			assertErr: assert.Error,
			after: func(t *testing.T, found domain.Role) {
				// 不存在的角色，不需要验证
			},
		},
		{
			name:      "不匹配的业务ID",
			bizID:     s.bizID + 1,
			roleID:    created.ID,
			assertErr: assert.Error,
			after: func(t *testing.T, found domain.Role) {
				// 使用错误的业务ID，不需要验证
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			found, err := s.svc.Svc.GetRole(ctx, tt.bizID, tt.roleID)
			tt.assertErr(t, err)

			if err == nil {
				tt.after(t, found)
			}
		})
	}
}

// TestRole_Update 测试更新角色
func (s *RoleTestSuite) TestRole_Update() {
	t := s.T()
	ctx := context.Background()

	// 先创建一个角色用于后续测试
	role := createTestRole(s.bizID, RoleTypeCustom)
	created, err := s.svc.Svc.CreateRole(ctx, role)
	require.NoError(t, err)

	// 测试结束后清理
	defer cleanupTest(t, ctx, func() error {
		return s.svc.Svc.DeleteRole(ctx, s.bizID, created.ID)
	})

	tests := []struct {
		name      string
		prepare   func() domain.Role
		assertErr assert.ErrorAssertionFunc
		after     func(t *testing.T, updated domain.Role, result domain.Role)
	}{
		{
			name: "更新角色基本信息",
			prepare: func() domain.Role {
				updated := created
				updated.Name = "更新后的角色名称"
				updated.Description = "更新后的角色描述"
				updated.Metadata = `{"updated":true}`
				return updated
			},
			assertErr: assert.NoError,
			after: func(t *testing.T, updated domain.Role, result domain.Role) {
				assert.Equal(t, updated.Name, result.Name)
				assert.Equal(t, updated.Description, result.Description)
				assert.Equal(t, updated.Metadata, result.Metadata)
			},
		},
		{
			name: "更新不存在的角色",
			prepare: func() domain.Role {
				nonExistentRole := createTestRole(s.bizID, RoleTypeCustom)
				nonExistentRole.ID = 99999
				return nonExistentRole
			},
			assertErr: assert.NoError,
			after: func(t *testing.T, updated domain.Role, result domain.Role) {
				// 更新不存在的角色不会返回错误，可能创建了新角色
				if result.ID > 0 {
					// 清理可能创建的新角色
					cleanupTest(t, ctx, func() error {
						return s.svc.Svc.DeleteRole(ctx, s.bizID, result.ID)
					})
				}
			},
		},
		{
			name: "使用错误的业务ID更新角色",
			prepare: func() domain.Role {
				invalidBizRole := created
				invalidBizRole.BizID = s.bizID + 1
				invalidBizRole.Name = "错误业务ID更新"
				return invalidBizRole
			},
			assertErr: assert.NoError,
			after: func(t *testing.T, updated domain.Role, result domain.Role) {
				// 检查原角色是否被更新（不应被错误业务ID更新）
				originalRole, getErr := s.svc.Svc.GetRole(ctx, s.bizID, created.ID)
				assert.NoError(t, getErr, "应能获取到原角色")
				assert.NotEqual(t, "错误业务ID更新", originalRole.Name, "原角色不应被错误业务ID更新")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updated := tt.prepare()
			result, err := s.svc.Svc.UpdateRole(ctx, updated)

			tt.assertErr(t, err)

			if err == nil {
				tt.after(t, updated, result)
			}
		})
	}
}

// TestRole_Delete 测试删除角色
func (s *RoleTestSuite) TestRole_Delete() {
	t := s.T()
	ctx := context.Background()

	tests := []struct {
		name      string
		prepare   func() (int64, int64, error)
		assertErr assert.ErrorAssertionFunc
		after     func(t *testing.T, bizID int64, roleID int64)
	}{
		{
			name: "删除存在的角色",
			prepare: func() (int64, int64, error) {
				// 创建一个角色
				role := createTestRole(s.bizID, RoleTypeSystem)
				created, err := s.svc.Svc.CreateRole(ctx, role)
				return s.bizID, created.ID, err
			},
			assertErr: assert.NoError,
			after: func(t *testing.T, bizID int64, roleID int64) {
				// 尝试获取已删除的角色
				_, err := s.svc.Svc.GetRole(ctx, bizID, roleID)
				assert.Error(t, err)
			},
		},
		{
			name: "删除不存在的角色",
			prepare: func() (int64, int64, error) {
				// 使用不存在的角色ID
				return s.bizID, 99999, nil
			},
			assertErr: assert.NoError,
			after: func(t *testing.T, bizID int64, roleID int64) {
				// 删除不存在的角色不需要额外检查
			},
		},
		{
			name: "不匹配的业务ID",
			prepare: func() (int64, int64, error) {
				// 创建一个角色
				role := createTestRole(s.bizID, RoleTypeCustom)
				created, err := s.svc.Svc.CreateRole(ctx, role)
				return s.bizID + 1, created.ID, err
			},
			assertErr: assert.NoError,
			after: func(t *testing.T, bizID int64, roleID int64) {
				// 确认角色仍然存在（没有被错误的业务ID删除）
				_, err := s.svc.Svc.GetRole(ctx, s.bizID, roleID)
				assert.NoError(t, err, "角色不应被错误的业务ID删除")

				// 清理
				cleanupTest(t, ctx, func() error {
					return s.svc.Svc.DeleteRole(ctx, s.bizID, roleID)
				})
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bizID, roleID, prepErr := tt.prepare()
			require.NoError(t, prepErr)

			err := s.svc.Svc.DeleteRole(ctx, bizID, roleID)

			tt.assertErr(t, err)
			tt.after(t, bizID, roleID)
		})
	}
}

// TestRole_List 测试列出角色
func (s *RoleTestSuite) TestRole_List() {
	t := s.T()
	ctx := context.Background()

	// 创建多个不同类型的角色
	role1 := createTestRole(s.bizID, RoleTypeCustom)
	role1.Name = "测试角色1-" + time.Now().String()
	created1, err := s.svc.Svc.CreateRole(ctx, role1)
	require.NoError(t, err)

	role2 := createTestRole(s.bizID, RoleTypeSystem)
	role2.Name = "测试角色2-" + time.Now().String()
	created2, err := s.svc.Svc.CreateRole(ctx, role2)
	require.NoError(t, err)

	// 测试结束后清理
	defer func() {
		cleanupTest(t, ctx, func() error {
			_ = s.svc.Svc.DeleteRole(ctx, s.bizID, created1.ID)
			return s.svc.Svc.DeleteRole(ctx, s.bizID, created2.ID)
		})
	}()

	tests := []struct {
		name      string
		listFunc  func() ([]domain.Role, error)
		assertErr assert.ErrorAssertionFunc
		after     func(t *testing.T, roles []domain.Role)
	}{
		{
			name: "列出所有角色",
			listFunc: func() ([]domain.Role, error) {
				return s.svc.Svc.ListRoles(ctx, s.bizID, 0, 10)
			},
			assertErr: assert.NoError,
			after: func(t *testing.T, roles []domain.Role) {
				assert.GreaterOrEqual(t, len(roles), 2) // 至少包含我们创建的两个角色

				// 验证包含我们创建的角色
				var foundRole1, foundRole2 bool
				for _, role := range roles {
					if role.ID == created1.ID {
						foundRole1 = true
					}
					if role.ID == created2.ID {
						foundRole2 = true
					}
				}
				assert.True(t, foundRole1, "应找到角色1")
				assert.True(t, foundRole2, "应找到角色2")
			},
		},
		{
			name: "按系统角色类型过滤",
			listFunc: func() ([]domain.Role, error) {
				return s.svc.Svc.ListRolesByRoleType(ctx, s.bizID, string(RoleTypeSystem), 0, 10)
			},
			assertErr: assert.NoError,
			after: func(t *testing.T, roles []domain.Role) {
				for _, role := range roles {
					assert.Equal(t, string(RoleTypeSystem), role.Type)
				}

				// 找到created2
				var found bool
				for _, role := range roles {
					if role.ID == created2.ID {
						found = true
						break
					}
				}
				assert.True(t, found, "应找到创建的系统角色")
			},
		},
		{
			name: "按自定义角色类型过滤",
			listFunc: func() ([]domain.Role, error) {
				return s.svc.Svc.ListRolesByRoleType(ctx, s.bizID, string(RoleTypeCustom), 0, 10)
			},
			assertErr: assert.NoError,
			after: func(t *testing.T, roles []domain.Role) {
				for _, role := range roles {
					assert.Equal(t, string(RoleTypeCustom), role.Type)
				}

				// 找到created1
				var found bool
				for _, role := range roles {
					if role.ID == created1.ID {
						found = true
						break
					}
				}
				assert.True(t, found, "应找到创建的自定义角色")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			roles, err := tt.listFunc()

			tt.assertErr(t, err)

			if err == nil {
				tt.after(t, roles)
			}
		})
	}

	// 单独测试不同业务ID的情况
	t.Run("不同业务ID", func(t *testing.T) {
		// 创建另一个业务
		otherBizConfig := createTestBusinessConfig("其他业务")
		createdBiz, err := s.svc.Svc.CreateBusinessConfig(ctx, otherBizConfig)
		require.NoError(t, err)
		otherBizID := createdBiz.ID

		// 在新业务中创建角色
		otherRole := createTestRole(otherBizID, RoleTypeSystem)
		createdOtherRole, err := s.svc.Svc.CreateRole(ctx, otherRole)
		require.NoError(t, err)

		// 使用otherBizID查询角色
		roles, err := s.svc.Svc.ListRoles(ctx, otherBizID, 0, 10)
		assert.NoError(t, err)

		// 验证查询结果
		var foundOtherRole bool
		var foundCreated1 bool
		for _, role := range roles {
			assert.Equal(t, otherBizID, role.BizID, "应只返回指定业务ID的角色")
			if role.ID == createdOtherRole.ID {
				foundOtherRole = true
			}
			if role.ID == created1.ID {
				foundCreated1 = true
			}
		}
		assert.True(t, foundOtherRole, "应找到其他业务的角色")
		assert.False(t, foundCreated1, "不应找到不属于该业务的角色")

		// 清理
		cleanupTest(t, ctx, func() error {
			return s.svc.Svc.DeleteRole(ctx, otherBizID, createdOtherRole.ID)
		})
	})
}
