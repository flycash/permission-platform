//go:build e2e

package rbac

import (
	"context"
	"testing"

	"gitee.com/flycash/permission-platform/internal/domain"
	rbacioc "gitee.com/flycash/permission-platform/internal/test/integration/ioc/rbac"
	testioc "gitee.com/flycash/permission-platform/internal/test/ioc"
	"github.com/ego-component/egorm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// UserRoleTestSuite 用户角色测试套件
type UserRoleTestSuite struct {
	suite.Suite
	db         *egorm.Component
	svc        *rbacioc.Service
	bizID      int64
	testUserID int64
	testRole   domain.Role
}

// SetupSuite 在所有测试之前设置
func (s *UserRoleTestSuite) SetupSuite() {
	s.db = testioc.InitDBAndTables()
	s.svc = rbacioc.Init()

	// 创建测试业务
	ctx := context.Background()
	bizConfig := createTestBusinessConfig("用户角色测试")
	created, err := s.svc.Svc.CreateBusinessConfig(ctx, bizConfig)
	if err != nil {
		s.T().Fatalf("创建测试业务失败: %v", err)
	}
	s.bizID = created.ID
	s.testUserID = TestUserID

	// 创建测试角色
	role := createTestRole(s.bizID, RoleTypeSystem)
	createdRole, err := s.svc.Svc.CreateRole(ctx, role)
	if err != nil {
		s.T().Fatalf("创建测试角色失败: %v", err)
	}
	s.testRole = createdRole
}

// TearDownSuite 在所有测试之后清理
func (s *UserRoleTestSuite) TearDownSuite() {
	// 清理所有测试数据
	ctx := context.Background()
	cleanTestEnvironment(s.T(), ctx, s.svc)
}

// TestUserRoleSuite 运行用户角色测试套件
func TestUserRoleSuite(t *testing.T) {
	suite.Run(t, new(UserRoleTestSuite))
}

// TestUserRole_Create 测试创建用户角色
func (s *UserRoleTestSuite) TestUserRole_Create() {
	t := s.T()
	ctx := context.Background()

	tests := []struct {
		name     string
		userRole func() domain.UserRole
		wantErr  bool
		errorMsg string
	}{
		{
			name: "创建有效的用户角色",
			userRole: func() domain.UserRole {
				return createTestUserRole(s.bizID, s.testUserID, s.testRole)
			},
			wantErr: false,
		},
		{
			name: "零业务ID",
			userRole: func() domain.UserRole {
				ur := createTestUserRole(0, s.testUserID, s.testRole)
				return ur
			},
			wantErr:  false,
			errorMsg: "",
		},
		{
			name: "用户ID为0",
			userRole: func() domain.UserRole {
				ur := createTestUserRole(s.bizID, 0, s.testRole)
				return ur
			},
			wantErr:  false,
			errorMsg: "",
		},
		{
			name: "角色ID为0",
			userRole: func() domain.UserRole {
				invalidRole := s.testRole
				invalidRole.ID = 0
				ur := createTestUserRole(s.bizID, s.testUserID, invalidRole)
				return ur
			},
			wantErr:  false,
			errorMsg: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRole := tt.userRole()
			created, err := s.svc.Svc.GrantUserRole(ctx, userRole)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
				assertUserRole(t, userRole, created)

				// 清理创建的用户角色
				cleanupTest(t, ctx, func() error {
					return s.svc.Svc.RevokeUserRole(ctx, s.bizID, created.ID)
				})
			}
		})
	}
}

// TestUserRole_Get 测试获取用户角色
func (s *UserRoleTestSuite) TestUserRole_Get() {
	t := s.T()
	ctx := context.Background()

	// 先创建一个用户角色用于后续测试
	userRole := createTestUserRole(s.bizID, s.testUserID, s.testRole)
	created, err := s.svc.Svc.GrantUserRole(ctx, userRole)
	require.NoError(t, err)

	// 测试结束后清理
	defer cleanupTest(t, ctx, func() error {
		return s.svc.Svc.RevokeUserRole(ctx, s.bizID, created.ID)
	})

	t.Run("获取存在的用户角色", func(t *testing.T) {
		// 使用ListUserRolesByUserID查找用户角色
		roles, err := s.svc.Svc.ListUserRolesByUserID(ctx, s.bizID, s.testUserID)
		assert.NoError(t, err)
		assert.NotEmpty(t, roles)

		// 在返回的角色列表中查找匹配的角色
		var found bool
		for _, role := range roles {
			if role.ID == created.ID {
				assertUserRole(t, created, role)
				found = true
				break
			}
		}

		assert.True(t, found, "未找到刚创建的用户角色")
	})

	t.Run("获取不存在的用户角色", func(t *testing.T) {
		// 使用不存在的用户ID查询
		nonExistentUserID := s.testUserID + 1000
		roles, err := s.svc.Svc.ListUserRolesByUserID(ctx, s.bizID, nonExistentUserID)
		assert.NoError(t, err)
		assert.Empty(t, roles, "不应找到不存在用户的角色")
	})

	t.Run("不匹配的业务ID", func(t *testing.T) {
		// 使用错误的业务ID查询
		roles, err := s.svc.Svc.ListUserRolesByUserID(ctx, s.bizID+1, s.testUserID)
		assert.NoError(t, err)
		assert.Empty(t, roles, "不应找到不匹配业务ID的角色")
	})
}

// TestUserRole_Delete 测试删除用户角色
func (s *UserRoleTestSuite) TestUserRole_Delete() {
	t := s.T()
	ctx := context.Background()

	t.Run("删除存在的用户角色", func(t *testing.T) {
		// 创建一个用户角色
		userRole := createTestUserRole(s.bizID, s.testUserID, s.testRole)
		created, err := s.svc.Svc.GrantUserRole(ctx, userRole)
		require.NoError(t, err)

		// 删除用户角色
		err = s.svc.Svc.RevokeUserRole(ctx, s.bizID, created.ID)
		assert.NoError(t, err)

		// 尝试获取已删除的用户角色
		roles, err := s.svc.Svc.ListUserRolesByUserID(ctx, s.bizID, s.testUserID)
		assert.NoError(t, err)

		// 验证角色已被删除
		var found bool
		for _, role := range roles {
			if role.ID == created.ID {
				found = true
				break
			}
		}

		assert.False(t, found, "用户角色应已被删除")
	})

	t.Run("删除不存在的用户角色", func(t *testing.T) {
		err := s.svc.Svc.RevokeUserRole(ctx, s.bizID, 99999)
		assert.NoError(t, err)
	})

	t.Run("不匹配的业务ID", func(t *testing.T) {
		// 创建一个用户角色
		userRole := createTestUserRole(s.bizID, s.testUserID, s.testRole)
		created, err := s.svc.Svc.GrantUserRole(ctx, userRole)
		require.NoError(t, err)

		// 使用错误的业务ID尝试删除
		err = s.svc.Svc.RevokeUserRole(ctx, s.bizID+1, created.ID)
		assert.NoError(t, err)

		// 确认用户角色是否仍然存在
		roles, err := s.svc.Svc.ListUserRolesByUserID(ctx, s.bizID, s.testUserID)
		assert.NoError(t, err)

		// 验证角色是否还存在
		var found bool
		for _, role := range roles {
			if role.ID == created.ID {
				found = true
				break
			}
		}

		// 记录是否仍然存在
		if found {
			t.Logf("错误的业务ID未能删除角色，角色ID=%d 仍然存在", created.ID)
		} else {
			t.Logf("错误的业务ID成功删除了角色，角色ID=%d", created.ID)
		}

		// 如果仍然存在则清理
		if found {
			cleanupTest(t, ctx, func() error {
				return s.svc.Svc.RevokeUserRole(ctx, s.bizID, created.ID)
			})
		}
	})
}

// TestUserRole_List 测试列出用户角色
func (s *UserRoleTestSuite) TestUserRole_List() {
	t := s.T()
	ctx := context.Background()

	// 创建多个角色
	var testRoles []domain.Role
	for i := 0; i < 3; i++ {
		role := createTestRole(s.bizID, RoleTypeCustom)
		created, err := s.svc.Svc.CreateRole(ctx, role)
		require.NoError(t, err)
		testRoles = append(testRoles, created)
	}

	// 创建多个用户
	var testUserIDs []int64
	for i := 0; i < 3; i++ {
		testUserIDs = append(testUserIDs, s.testUserID+int64(i))
	}

	// 创建多个用户角色关系
	var userRoleIDs []int64

	// 为第一个用户分配所有角色
	for _, role := range testRoles {
		ur := createTestUserRole(s.bizID, testUserIDs[0], role)
		created, err := s.svc.Svc.GrantUserRole(ctx, ur)
		require.NoError(t, err)
		userRoleIDs = append(userRoleIDs, created.ID)
	}

	// 为所有用户分配第一个角色
	for i := 1; i < len(testUserIDs); i++ {
		ur := createTestUserRole(s.bizID, testUserIDs[i], testRoles[0])
		created, err := s.svc.Svc.GrantUserRole(ctx, ur)
		require.NoError(t, err)
		userRoleIDs = append(userRoleIDs, created.ID)
	}

	// 测试结束后清理
	defer func() {
		for _, id := range userRoleIDs {
			cleanupTest(t, ctx, func() error {
				return s.svc.Svc.RevokeUserRole(ctx, s.bizID, id)
			})
		}
	}()

	t.Run("列出用户的所有角色", func(t *testing.T) {
		userRoles, err := s.svc.Svc.ListUserRolesByUserID(ctx, s.bizID, testUserIDs[0])
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(userRoles), 3)
		for _, ur := range userRoles {
			assert.Equal(t, testUserIDs[0], ur.UserID)
		}
	})

	t.Run("列出所有用户角色", func(t *testing.T) {
		userRoles, err := s.svc.Svc.ListUserRoles(ctx, s.bizID)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(userRoles), len(userRoleIDs))

		// 检查用户角色记录是否完整
		roleMap := make(map[int64]bool)
		for _, ur := range userRoles {
			for _, id := range userRoleIDs {
				if ur.ID == id {
					roleMap[id] = true
				}
			}
		}
		assert.Equal(t, len(userRoleIDs), len(roleMap), "应找到所有创建的用户角色")
	})

	t.Run("不同业务ID", func(t *testing.T) {
		// 创建另一个业务
		otherBizConfig := createTestBusinessConfig("其他业务")
		createdBiz, err := s.svc.Svc.CreateBusinessConfig(ctx, otherBizConfig)
		require.NoError(t, err)
		otherBizID := createdBiz.ID

		// 在新业务中创建角色
		otherRole := createTestRole(otherBizID, RoleTypeSystem)
		createdRole, err := s.svc.Svc.CreateRole(ctx, otherRole)
		require.NoError(t, err)

		// 创建用户角色关系
		otherUR := createTestUserRole(otherBizID, testUserIDs[0], createdRole)
		createdUR, err := s.svc.Svc.GrantUserRole(ctx, otherUR)
		require.NoError(t, err)

		// 查询
		userRoles, err := s.svc.Svc.ListUserRolesByUserID(ctx, otherBizID, testUserIDs[0])
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(userRoles), 1)
		for _, ur := range userRoles {
			assert.Equal(t, otherBizID, ur.BizID)
		}

		// 清理
		cleanupTest(t, ctx, func() error {
			return s.svc.Svc.RevokeUserRole(ctx, otherBizID, createdUR.ID)
		})
	})
}
