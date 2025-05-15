//go:build e2e

package rbac

import (
	"context"
	"fmt"
	"testing"

	"gitee.com/flycash/permission-platform/internal/domain"
	"gitee.com/flycash/permission-platform/internal/service/rbac"
	rbacioc "gitee.com/flycash/permission-platform/internal/test/integration/ioc/rbac"
	testioc "gitee.com/flycash/permission-platform/internal/test/ioc"
	"github.com/ego-component/egorm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// TestPermissionCheckSuite 运行权限检查测试套件
func TestPermissionCheckSuite(t *testing.T) {
	suite.Run(t, new(PermissionCheckTestSuite))
}

// PermissionCheckTestSuite 权限检查测试套件
type PermissionCheckTestSuite struct {
	suite.Suite
	db            *egorm.Component
	svc           *rbacioc.Service
	permissionSvc rbac.PermissionService
	bizID         int64
	testUserID    int64
}

// SetupSuite 在所有测试之前设置基本环境
func (s *PermissionCheckTestSuite) SetupSuite() {
	s.db = testioc.InitDBAndTables()
	s.svc = rbacioc.Init()
	s.permissionSvc = s.svc.PermissionSvc

	// 创建测试业务
	ctx := context.Background()
	bizConfig := createTestBusinessConfig("权限检查测试")
	created, err := s.svc.Svc.CreateBusinessConfig(ctx, bizConfig)
	if err != nil {
		s.T().Fatalf("创建测试业务失败: %v", err)
	}
	s.bizID = created.ID
	s.testUserID = TestUserID
}

// TearDownSuite 在所有测试之后清理
func (s *PermissionCheckTestSuite) TearDownSuite() {
	// 清理所有测试数据
	ctx := context.Background()
	cleanTestEnvironment(s.T(), ctx, s.svc)
}

// 清理数据库中所有用户权限和角色
func cleanUserPermissionsAndRoles(t *testing.T, ctx context.Context, svc *rbacioc.Service, bizID, userID int64) {
	t.Logf("清理用户 %d 的所有权限和角色", userID)

	// 清理用户权限
	perms, err := svc.Svc.ListUserPermissionsByUserID(ctx, bizID, userID)
	require.NoError(t, err, "查询用户权限失败")

	t.Logf("找到用户 %d 的权限数量: %d", userID, len(perms))
	for _, perm := range perms {
		t.Logf("撤销用户权限: %+v", perm)
		err = svc.Svc.RevokeUserPermission(ctx, bizID, perm.ID)
		require.NoError(t, err, "撤销用户权限失败")
	}

	// 清理用户角色
	roles, err := svc.Svc.ListUserRolesByUserID(ctx, bizID, userID)
	require.NoError(t, err, "查询用户角色失败")

	t.Logf("找到用户 %d 的角色数量: %d", userID, len(roles))
	for _, role := range roles {
		t.Logf("撤销用户角色: %+v", role)
		err = svc.Svc.RevokeUserRole(ctx, bizID, role.ID)
		require.NoError(t, err, "撤销用户角色失败")
	}
}

// 清理角色嵌套
func cleanRoleInclusions(t *testing.T, ctx context.Context, svc *rbacioc.Service, bizID int64) {
	t.Logf("清理所有角色嵌套")

	// 查找所有角色嵌套
	inclusions, err := svc.Svc.ListRoleInclusions(ctx, bizID, 0, 1000)
	require.NoError(t, err, "查询角色嵌套失败")

	t.Logf("找到角色嵌套数量: %d", len(inclusions))
	for _, inclusion := range inclusions {
		t.Logf("删除角色嵌套: %+v", inclusion)
		err = svc.Svc.DeleteRoleInclusion(ctx, bizID, inclusion.ID)
		require.NoError(t, err, "删除角色嵌套失败")
	}
}

// TestCheck 测试权限检查功能
func (s *PermissionCheckTestSuite) TestCheck() {
	t := s.T()
	ctx := context.Background()

	// 首先清理所有可能存在的角色嵌套关系，防止循环依赖
	cleanRoleInclusions(t, ctx, s.svc, s.bizID)

	// 清理所有测试用户的权限和角色
	cleanUserPermissionsAndRoles(t, ctx, s.svc, s.bizID, s.testUserID)
	cleanUserPermissionsAndRoles(t, ctx, s.svc, s.bizID, s.testUserID+1)
	cleanUserPermissionsAndRoles(t, ctx, s.svc, s.bizID, s.testUserID+2)

	// 测试数据初始化函数
	setup := func(t *testing.T) (domain.Resource, domain.Permission, domain.Role, domain.Role, domain.Role) {
		// 创建测试资源 - API类型
		apiRes := createTestResource(s.bizID, "api", "/api/users")
		createdAPI, err := s.svc.Svc.CreateResource(ctx, apiRes)
		require.NoError(t, err, "创建API资源失败")
		t.Logf("创建资源成功: %+v", createdAPI)

		// 创建权限 - 读操作
		readPerm := createTestPermission(s.bizID, createdAPI, ActionTypeRead)
		createdReadPerm, err := s.svc.Svc.CreatePermission(ctx, readPerm)
		require.NoError(t, err, "创建读权限失败")
		t.Logf("创建读权限成功: %+v", createdReadPerm)

		// 创建权限 - 写操作
		writePerm := createTestPermission(s.bizID, createdAPI, ActionTypeWrite)
		createdWritePerm, err := s.svc.Svc.CreatePermission(ctx, writePerm)
		require.NoError(t, err, "创建写权限失败")
		t.Logf("创建写权限成功: %+v", createdWritePerm)

		// 创建权限 - 删除操作
		deletePerm := createTestPermission(s.bizID, createdAPI, "delete")
		createdDeletePerm, err := s.svc.Svc.CreatePermission(ctx, deletePerm)
		require.NoError(t, err, "创建删除权限失败")
		t.Logf("创建删除权限成功: %+v", createdDeletePerm)

		// 创建角色 - 管理员
		adminRole := createTestRole(s.bizID, RoleTypeSystem)
		adminRole.Name = "admin"
		createdAdminRole, err := s.svc.Svc.CreateRole(ctx, adminRole)
		require.NoError(t, err, "创建管理员角色失败")
		t.Logf("创建管理员角色成功: %+v", createdAdminRole)

		// 创建角色 - 编辑者
		editorRole := createTestRole(s.bizID, RoleTypeSystem)
		editorRole.Name = "editor"
		createdEditorRole, err := s.svc.Svc.CreateRole(ctx, editorRole)
		require.NoError(t, err, "创建编辑者角色失败")
		t.Logf("创建编辑者角色成功: %+v", createdEditorRole)

		// 创建角色 - 查看者
		viewerRole := createTestRole(s.bizID, RoleTypeCustom)
		viewerRole.Name = "viewer"
		createdViewerRole, err := s.svc.Svc.CreateRole(ctx, viewerRole)
		require.NoError(t, err, "创建查看者角色失败")
		t.Logf("创建查看者角色成功: %+v", createdViewerRole)

		// 分配权限给角色
		// 管理员拥有所有权限
		adminReadPerm := createTestRolePermission(s.bizID, createdAdminRole, createdReadPerm)
		_, err = s.svc.Svc.GrantRolePermission(ctx, adminReadPerm)
		require.NoError(t, err, "授予管理员读权限失败")

		adminWritePerm := createTestRolePermission(s.bizID, createdAdminRole, createdWritePerm)
		_, err = s.svc.Svc.GrantRolePermission(ctx, adminWritePerm)
		require.NoError(t, err, "授予管理员写权限失败")

		adminDeletePerm := createTestRolePermission(s.bizID, createdAdminRole, createdDeletePerm)
		_, err = s.svc.Svc.GrantRolePermission(ctx, adminDeletePerm)
		require.NoError(t, err, "授予管理员删除权限失败")

		// 编辑者拥有读写权限
		editorReadPerm := createTestRolePermission(s.bizID, createdEditorRole, createdReadPerm)
		_, err = s.svc.Svc.GrantRolePermission(ctx, editorReadPerm)
		require.NoError(t, err, "授予编辑者读权限失败")

		editorWritePerm := createTestRolePermission(s.bizID, createdEditorRole, createdWritePerm)
		_, err = s.svc.Svc.GrantRolePermission(ctx, editorWritePerm)
		require.NoError(t, err, "授予编辑者写权限失败")

		// 查看者只有读权限
		viewerReadPerm := createTestRolePermission(s.bizID, createdViewerRole, createdReadPerm)
		_, err = s.svc.Svc.GrantRolePermission(ctx, viewerReadPerm)
		require.NoError(t, err, "授予查看者读权限失败")

		return createdAPI, createdReadPerm, createdAdminRole, createdEditorRole, createdViewerRole
	}

	// 初始化测试数据
	apiResource, readPerm, adminRole, editorRole, viewerRole := setup(t)

	// 准备额外的测试用户
	testUserID2 := s.testUserID + 1
	testUserID3 := s.testUserID + 2

	tests := []struct {
		name      string
		before    func(t *testing.T) (int64, domain.Resource, []string)
		expected  bool
		assertErr assert.ErrorAssertionFunc
	}{
		{
			name: "用户没有任何权限时返回false",
			before: func(t *testing.T) (int64, domain.Resource, []string) {
				// 用户没有任何权限，直接返回测试用户ID
				return s.testUserID, apiResource, []string{string(ActionTypeRead)}
			},
			expected:  false,
			assertErr: assert.NoError,
		},
		{
			name: "用户有直接允许权限返回true",
			before: func(t *testing.T) (int64, domain.Resource, []string) {
				cleanUserPermissionsAndRoles(t, ctx, s.svc, s.bizID, s.testUserID)

				// 为用户直接授予读权限
				userPerm := createTestUserPermission(s.bizID, s.testUserID, readPerm, domain.EffectAllow)
				created, err := s.svc.Svc.GrantUserPermission(ctx, userPerm)
				require.NoError(t, err, "授予用户直接权限失败")
				t.Logf("授予用户权限成功: %+v", created)

				return s.testUserID, apiResource, []string{string(ActionTypeRead)}
			},
			expected:  true,
			assertErr: assert.NoError,
		},
		{
			name: "用户有直接拒绝权限返回false",
			before: func(t *testing.T) (int64, domain.Resource, []string) {
				cleanUserPermissionsAndRoles(t, ctx, s.svc, s.bizID, testUserID2)

				// 为用户直接授予读权限但效果是拒绝
				userPerm := createTestUserPermission(s.bizID, testUserID2, readPerm, domain.EffectDeny)
				created, err := s.svc.Svc.GrantUserPermission(ctx, userPerm)
				require.NoError(t, err, "授予用户拒绝权限失败")
				t.Logf("授予用户拒绝权限成功: %+v", created)

				return testUserID2, apiResource, []string{string(ActionTypeRead)}
			},
			expected:  false,
			assertErr: assert.NoError,
		},
		{
			name: "通过角色获得允许权限返回true",
			before: func(t *testing.T) (int64, domain.Resource, []string) {
				cleanUserPermissionsAndRoles(t, ctx, s.svc, s.bizID, s.testUserID)

				// 为用户分配查看者角色，从而获得读权限
				userRole := createTestUserRole(s.bizID, s.testUserID, viewerRole)
				created, err := s.svc.Svc.GrantUserRole(ctx, userRole)
				require.NoError(t, err, "授予用户角色失败")
				t.Logf("授予用户角色成功: %+v", created)

				return s.testUserID, apiResource, []string{string(ActionTypeRead)}
			},
			expected:  true,
			assertErr: assert.NoError,
		},
		{
			name: "操作不匹配返回false",
			before: func(t *testing.T) (int64, domain.Resource, []string) {
				// 用户只有读权限，但尝试写操作
				return s.testUserID, apiResource, []string{string(ActionTypeWrite)}
			},
			expected:  false,
			assertErr: assert.NoError,
		},
		{
			name: "资源类型不匹配返回false",
			before: func(t *testing.T) (int64, domain.Resource, []string) {
				// 创建一个不同类型的资源
				diffTypeRes := apiResource
				diffTypeRes.Type = "different-type"
				return s.testUserID, diffTypeRes, []string{string(ActionTypeRead)}
			},
			expected:  false,
			assertErr: assert.NoError,
		},
		{
			name: "资源键不匹配返回false",
			before: func(t *testing.T) (int64, domain.Resource, []string) {
				// 创建一个键不同的资源
				diffKeyRes := apiResource
				diffKeyRes.Key = "different-key"
				return s.testUserID, diffKeyRes, []string{string(ActionTypeRead)}
			},
			expected:  false,
			assertErr: assert.NoError,
		},
		{
			name: "同时拥有允许和拒绝权限时拒绝优先",
			before: func(t *testing.T) (int64, domain.Resource, []string) {
				cleanUserPermissionsAndRoles(t, ctx, s.svc, s.bizID, testUserID2)

				// 为用户分配管理员角色，获得允许权限
				userRole := createTestUserRole(s.bizID, testUserID2, adminRole)
				created, err := s.svc.Svc.GrantUserRole(ctx, userRole)
				require.NoError(t, err, "授予用户管理员角色失败")
				t.Logf("授予用户管理员角色成功: %+v", created)

				// 然后添加拒绝权限
				userPerm := createTestUserPermission(s.bizID, testUserID2, readPerm, domain.EffectDeny)
				created2, err := s.svc.Svc.GrantUserPermission(ctx, userPerm)
				require.NoError(t, err, "授予用户拒绝权限失败")
				t.Logf("授予用户拒绝权限成功: %+v", created2)

				return testUserID2, apiResource, []string{string(ActionTypeRead)}
			},
			expected:  false,
			assertErr: assert.NoError,
		},
		{
			name: "支持多个操作中的一个",
			before: func(t *testing.T) (int64, domain.Resource, []string) {
				// 用户只有读权限，传入多个操作，包括读
				return s.testUserID, apiResource, []string{string(ActionTypeWrite), string(ActionTypeRead)}
			},
			expected:  true,
			assertErr: assert.NoError,
		},
		{
			name: "通过角色嵌套获得权限",
			before: func(t *testing.T) (int64, domain.Resource, []string) {
				cleanUserPermissionsAndRoles(t, ctx, s.svc, s.bizID, testUserID3)
				cleanRoleInclusions(t, ctx, s.svc, s.bizID)

				// 创建角色嵌套：管理员包含编辑者
				roleInclusion := createTestRoleInclusion(s.bizID, adminRole, editorRole)
				created, err := s.svc.Svc.CreateRoleInclusion(ctx, roleInclusion)
				require.NoError(t, err, "创建角色嵌套失败")
				t.Logf("创建角色嵌套成功: %+v", created)

				// 用户分配管理员角色
				userRole := createTestUserRole(s.bizID, testUserID3, adminRole)
				created2, err := s.svc.Svc.GrantUserRole(ctx, userRole)
				require.NoError(t, err, "授予用户管理员角色失败")
				t.Logf("授予用户管理员角色成功: %+v", created2)

				// 此时用户应该通过角色嵌套获得写权限
				return testUserID3, apiResource, []string{string(ActionTypeWrite)}
			},
			expected:  true,
			assertErr: assert.NoError,
		},
	}

	// 执行测试
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("==== 开始执行测试: %s ====", tt.name)
			userID, resource, actions := tt.before(t)
			t.Logf("使用参数: userID=%d, resource=%+v, actions=%v", userID, resource, actions)

			// 执行权限检查
			t.Logf("执行权限检查...")
			hasPermission, err := s.permissionSvc.Check(ctx, s.bizID, userID, resource, actions)
			t.Logf("权限检查结果: %v, 错误: %v", hasPermission, err)

			// 验证结果
			tt.assertErr(t, err)
			assert.Equal(t, tt.expected, hasPermission, fmt.Sprintf("期望结果 %v，实际结果 %v", tt.expected, hasPermission))
			t.Logf("==== 测试完成: %s ====", tt.name)
		})
	}
}
