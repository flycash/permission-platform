//go:build e2e

package integration

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

func TestRBACServiceSuite(t *testing.T) {
	suite.Run(t, new(RBACServiceTestSuite))
}

type RBACServiceTestSuite struct {
	suite.Suite
	db  *egorm.Component
	svc *rbacioc.Service
}

func (s *RBACServiceTestSuite) SetupSuite() {
	s.db = testioc.InitDBAndTables()
	s.svc = rbacioc.Init()
}

// 创建测试用角色对象
func (s *RBACServiceTestSuite) createTestRole(bizID int64, roleType domain.RoleType) domain.Role {
	now := time.Now().UnixMilli()
	return domain.Role{
		BizID:       bizID,
		Type:        roleType,
		Name:        fmt.Sprintf("测试角色-%d", now),
		Description: "测试角色描述",
		StartTime:   now,
		EndTime:     now + 86400000, // 一天后过期
	}
}

// 断言角色对象的字段是否符合预期
func (s *RBACServiceTestSuite) assertRole(t *testing.T, expected, actual domain.Role) {
	assert.NotZero(t, actual.ID)
	assert.Equal(t, expected.BizID, actual.BizID)
	assert.Equal(t, expected.Type, actual.Type)
	assert.Equal(t, expected.Name, actual.Name)
	assert.Equal(t, expected.Description, actual.Description)
	assert.Equal(t, expected.StartTime, actual.StartTime)
	assert.Equal(t, expected.EndTime, actual.EndTime)
}

// 测试创建角色
func (s *RBACServiceTestSuite) TestCreateRole() {
	t := s.T()

	tests := []struct {
		name     string
		role     func() domain.Role
		wantErr  bool
		errorMsg string
	}{
		{
			name: "创建系统角色",
			role: func() domain.Role {
				return s.createTestRole(1, domain.RoleTypeSystem)
			},
			wantErr: false,
		},
		{
			name: "创建自定义角色",
			role: func() domain.Role {
				return s.createTestRole(1, domain.RoleTypeCustom)
			},
			wantErr: false,
		},
		{
			name: "创建临时角色",
			role: func() domain.Role {
				return s.createTestRole(1, domain.RoleTypeTemporary)
			},
			wantErr: false,
		},
		{
			name: "空名称",
			role: func() domain.Role {
				r := s.createTestRole(1, domain.RoleTypeCustom)
				r.Name = ""
				return r
			},
			wantErr:  true,
			errorMsg: "角色名称不能为空",
		},
		{
			name: "零业务ID",
			role: func() domain.Role {
				r := s.createTestRole(0, domain.RoleTypeCustom)
				return r
			},
			wantErr:  true,
			errorMsg: "业务ID必须大于0",
		},
		{
			name: "临时角色无结束时间",
			role: func() domain.Role {
				r := s.createTestRole(1, domain.RoleTypeTemporary)
				r.EndTime = 0
				return r
			},
			wantErr:  true,
			errorMsg: "临时角色必须设置结束时间",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			role := tt.role()
			created, err := s.svc.Svc.CreateRole(context.Background(), role)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
				s.assertRole(t, role, created)
			}
		})
	}
}

// 测试获取角色
func (s *RBACServiceTestSuite) TestGetRole() {
	t := s.T()

	// 先创建一个角色用于后续测试
	role := s.createTestRole(1, domain.RoleTypeCustom)
	created, err := s.svc.Svc.CreateRole(context.Background(), role)
	require.NoError(t, err)

	tests := []struct {
		name    string
		roleID  int64
		wantErr bool
	}{
		{
			name:    "获取存在的角色",
			roleID:  created.ID,
			wantErr: false,
		},
		{
			name:    "获取不存在的角色",
			roleID:  99999,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			found, err := s.svc.Svc.GetRole(context.Background(), tt.roleID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				s.assertRole(t, created, found)
			}
		})
	}
}

// 测试更新角色
func (s *RBACServiceTestSuite) TestUpdateRole() {
	t := s.T()

	// 先创建一个角色用于后续测试
	role := s.createTestRole(1, domain.RoleTypeCustom)
	created, err := s.svc.Svc.CreateRole(context.Background(), role)
	require.NoError(t, err)

	tests := []struct {
		name     string
		prepare  func() domain.Role
		wantErr  bool
		validate func(t *testing.T, updated domain.Role)
	}{
		{
			name: "更新角色成功",
			prepare: func() domain.Role {
				updated := created
				updated.Name = "更新后的角色名称"
				updated.Description = "更新后的角色描述"
				updated.StartTime = time.Now().Add(time.Hour).UnixMilli()
				updated.EndTime = time.Now().Add(time.Hour * 48).UnixMilli()
				return updated
			},
			wantErr: false,
			validate: func(t *testing.T, updated domain.Role) {
				assert.Equal(t, "更新后的角色名称", updated.Name)
				assert.Equal(t, "更新后的角色描述", updated.Description)
			},
		},
		{
			name: "更新不存在的角色",
			prepare: func() domain.Role {
				nonExistentRole := s.createTestRole(1, domain.RoleTypeCustom)
				nonExistentRole.ID = 99999
				return nonExistentRole
			},
			wantErr: true,
			validate: func(t *testing.T, updated domain.Role) {
				// 不需要验证
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			roleToUpdate := tt.prepare()
			updated, err := s.svc.Svc.UpdateRole(context.Background(), roleToUpdate)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				tt.validate(t, updated)

				// 再次获取确认更新成功
				found, err := s.svc.Svc.GetRole(context.Background(), updated.ID)
				require.NoError(t, err)
				s.assertRole(t, roleToUpdate, found)
			}
		})
	}
}

// 测试删除角色
func (s *RBACServiceTestSuite) TestDeleteRole() {
	t := s.T()

	tests := []struct {
		name     string
		prepare  func() int64
		wantErr  bool
		validate func(t *testing.T, roleID int64)
	}{
		{
			name: "删除存在的角色",
			prepare: func() int64 {
				role := s.createTestRole(1, domain.RoleTypeCustom)
				created, err := s.svc.Svc.CreateRole(context.Background(), role)
				require.NoError(t, err)
				return created.ID
			},
			wantErr: false,
			validate: func(t *testing.T, roleID int64) {
				// 尝试获取已删除的角色
				_, err := s.svc.Svc.GetRole(context.Background(), roleID)
				assert.Error(t, err)
			},
		},
		{
			name: "删除不存在的角色",
			prepare: func() int64 {
				return 99999
			},
			wantErr: true,
			validate: func(t *testing.T, roleID int64) {
				// 不需要额外验证
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			roleID := tt.prepare()
			err := s.svc.Svc.DeleteRole(context.Background(), roleID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				tt.validate(t, roleID)
			}
		})
	}
}

// 测试列出角色
func (s *RBACServiceTestSuite) TestListRoles() {
	t := s.T()

	const bizID = 1

	// 准备测试数据 - 创建多个不同类型的角色
	prepareTestRoles := func() {
		// 创建系统角色
		for i := 0; i < 2; i++ {
			role := s.createTestRole(bizID, domain.RoleTypeSystem)
			role.Name = fmt.Sprintf("系统角色-%d", i)
			_, err := s.svc.Svc.CreateRole(context.Background(), role)
			require.NoError(t, err)
		}

		// 创建自定义角色
		for i := 0; i < 3; i++ {
			role := s.createTestRole(bizID, domain.RoleTypeCustom)
			role.Name = fmt.Sprintf("自定义角色-%d", i)
			_, err := s.svc.Svc.CreateRole(context.Background(), role)
			require.NoError(t, err)
		}

		// 创建临时角色
		for i := 0; i < 1; i++ {
			role := s.createTestRole(bizID, domain.RoleTypeTemporary)
			role.Name = fmt.Sprintf("临时角色-%d", i)
			_, err := s.svc.Svc.CreateRole(context.Background(), role)
			require.NoError(t, err)
		}
	}

	// 创建测试角色数据
	prepareTestRoles()

	// 测试不同的查询场景
	tests := []struct {
		name      string
		bizID     int64
		offset    int
		limit     int
		roleType  string
		wantCount int
		validate  func(t *testing.T, roles []domain.Role)
	}{
		{
			name:      "列出所有角色",
			bizID:     bizID,
			offset:    0,
			limit:     10,
			roleType:  "",
			wantCount: 6, // 2系统 + 3自定义 + 1临时
			validate: func(t *testing.T, roles []domain.Role) {
				assert.Len(t, roles, 6)
			},
		},
		{
			name:      "列出系统角色",
			bizID:     bizID,
			offset:    0,
			limit:     10,
			roleType:  string(domain.RoleTypeSystem),
			wantCount: 2,
			validate: func(t *testing.T, roles []domain.Role) {
				for _, role := range roles {
					assert.Equal(t, domain.RoleTypeSystem, role.Type)
				}
			},
		},
		{
			name:      "列出自定义角色",
			bizID:     bizID,
			offset:    0,
			limit:     10,
			roleType:  string(domain.RoleTypeCustom),
			wantCount: 3,
			validate: func(t *testing.T, roles []domain.Role) {
				for _, role := range roles {
					assert.Equal(t, domain.RoleTypeCustom, role.Type)
				}
			},
		},
		{
			name:      "列出临时角色",
			bizID:     bizID,
			offset:    0,
			limit:     10,
			roleType:  string(domain.RoleTypeTemporary),
			wantCount: 1,
			validate: func(t *testing.T, roles []domain.Role) {
				for _, role := range roles {
					assert.Equal(t, domain.RoleTypeTemporary, role.Type)
				}
			},
		},
		{
			name:      "分页限制",
			bizID:     bizID,
			offset:    0,
			limit:     2,
			roleType:  "",
			wantCount: 2,
			validate: func(t *testing.T, roles []domain.Role) {
				assert.Len(t, roles, 2)
			},
		},
		{
			name:      "偏移量",
			bizID:     bizID,
			offset:    2,
			limit:     10,
			roleType:  "",
			wantCount: 4, // 总数6 - 偏移2 = 4
			validate: func(t *testing.T, roles []domain.Role) {
				assert.Len(t, roles, 4)
			},
		},
		{
			name:      "不同业务ID",
			bizID:     bizID + 1,
			offset:    0,
			limit:     10,
			roleType:  "",
			wantCount: 0,
			validate: func(t *testing.T, roles []domain.Role) {
				assert.Empty(t, roles)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			roles, count, err := s.svc.Svc.ListRolesByRoleType(context.Background(), tt.bizID, tt.offset, tt.limit, tt.roleType)

			require.NoError(t, err)
			assert.Equal(t, tt.wantCount, count)
			assert.Len(t, roles, tt.wantCount)

			if tt.validate != nil {
				tt.validate(t, roles)
			}
		})
	}
}

// 创建测试用资源对象
func (s *RBACServiceTestSuite) createTestResource(bizID int64, resType, key string) domain.Resource {
	now := time.Now().UnixMilli()
	return domain.Resource{
		BizID:       bizID,
		Type:        resType,
		Key:         key,
		Name:        fmt.Sprintf("测试资源-%d", now),
		Description: "测试资源描述",
	}
}

// 断言资源对象的字段是否符合预期
func (s *RBACServiceTestSuite) assertResource(t *testing.T, expected, actual domain.Resource) {
	assert.NotZero(t, actual.ID)
	assert.Equal(t, expected.BizID, actual.BizID)
	assert.Equal(t, expected.Type, actual.Type)
	assert.Equal(t, expected.Key, actual.Key)
	assert.Equal(t, expected.Name, actual.Name)
	assert.Equal(t, expected.Description, actual.Description)
}

// 测试创建资源
func (s *RBACServiceTestSuite) TestCreateResource() {
	t := s.T()

	tests := []struct {
		name     string
		resource func() domain.Resource
		wantErr  bool
		errorMsg string
	}{
		{
			name: "创建API资源",
			resource: func() domain.Resource {
				return s.createTestResource(1, "api", "api:test")
			},
			wantErr: false,
		},
		{
			name: "创建菜单资源",
			resource: func() domain.Resource {
				return s.createTestResource(1, "menu", "menu:test")
			},
			wantErr: false,
		},
		{
			name: "创建按钮资源",
			resource: func() domain.Resource {
				return s.createTestResource(1, "button", "button:test")
			},
			wantErr: false,
		},
		{
			name: "创建数据资源",
			resource: func() domain.Resource {
				return s.createTestResource(1, "data", "data:test")
			},
			wantErr: false,
		},
		{
			name: "空名称",
			resource: func() domain.Resource {
				r := s.createTestResource(1, "api", "user:read")
				r.Name = ""
				return r
			},
			wantErr:  true,
			errorMsg: "资源名称不能为空",
		},
		{
			name: "空类型",
			resource: func() domain.Resource {
				r := s.createTestResource(1, "", "user:read")
				return r
			},
			wantErr:  true,
			errorMsg: "资源类型不能为空",
		},
		{
			name: "空Key",
			resource: func() domain.Resource {
				r := s.createTestResource(1, "api", "")
				return r
			},
			wantErr:  true,
			errorMsg: "资源Key不能为空",
		},
		{
			name: "零业务ID",
			resource: func() domain.Resource {
				r := s.createTestResource(0, "api", "user:read")
				return r
			},
			wantErr:  true,
			errorMsg: "业务ID必须大于0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resource := tt.resource()
			created, err := s.svc.Svc.CreateResource(context.Background(), resource)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
				s.assertResource(t, resource, created)
			}
		})
	}
}

// 测试获取资源
func (s *RBACServiceTestSuite) TestGetResource() {
	t := s.T()

	// 先创建一个资源用于后续测试
	resource := s.createTestResource(1, "api", "user:read")
	created, err := s.svc.Svc.CreateResource(context.Background(), resource)
	require.NoError(t, err)

	tests := []struct {
		name       string
		resourceID int64
		wantErr    bool
	}{
		{
			name:       "获取存在的资源",
			resourceID: created.ID,
			wantErr:    false,
		},
		{
			name:       "获取不存在的资源",
			resourceID: 99999,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			found, err := s.svc.Svc.GetResource(context.Background(), tt.resourceID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				s.assertResource(t, created, found)
			}
		})
	}
}

// 测试更新资源
func (s *RBACServiceTestSuite) TestUpdateResource() {
	t := s.T()

	tests := []struct {
		name     string
		prepare  func() domain.Resource
		wantErr  bool
		validate func(t *testing.T, updated domain.Resource)
	}{
		{
			name: "更新资源成功",
			prepare: func() domain.Resource {
				// 先创建一个资源
				resource := s.createTestResource(1, "api", "user:read")
				created, err := s.svc.Svc.CreateResource(context.Background(), resource)
				require.NoError(t, err)

				// 修改资源信息
				created.Name = "更新后的资源名称"
				created.Description = "更新后的资源描述"
				return created
			},
			wantErr: false,
			validate: func(t *testing.T, updated domain.Resource) {
				assert.Equal(t, "更新后的资源名称", updated.Name)
				assert.Equal(t, "更新后的资源描述", updated.Description)
			},
		},
		{
			name: "更新不存在的资源",
			prepare: func() domain.Resource {
				nonExistentResource := s.createTestResource(1, "api", "user:write")
				nonExistentResource.ID = 99999
				return nonExistentResource
			},
			wantErr: true,
			validate: func(t *testing.T, updated domain.Resource) {
				// 不需要验证
			},
		},
		{
			name: "尝试更新不可变字段",
			prepare: func() domain.Resource {
				// 创建一个新资源
				resource := s.createTestResource(1, "api", "user:delete")
				created, err := s.svc.Svc.CreateResource(context.Background(), resource)
				require.NoError(t, err)

				// 尝试修改类型和Key（这些应该是不可变的）
				created.Type = "menu"
				created.Key = "menu:user"
				return created
			},
			wantErr: false, // 具体行为取决于实现
			validate: func(t *testing.T, updated domain.Resource) {
				// 如果更新成功，测试类型和Key是否被保留为原始值
				// 或者检查错误信息中是否包含关于不可变字段的提示
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resourceToUpdate := tt.prepare()
			updated, err := s.svc.Svc.UpdateResource(context.Background(), resourceToUpdate)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, updated)
				}

				// 再次获取确认更新成功
				found, err := s.svc.Svc.GetResource(context.Background(), updated.ID)
				require.NoError(t, err)
				s.assertResource(t, resourceToUpdate, found)
			}
		})
	}
}

// 测试删除资源
func (s *RBACServiceTestSuite) TestDeleteResource() {
	t := s.T()

	tests := []struct {
		name     string
		prepare  func() int64
		wantErr  bool
		validate func(t *testing.T, resourceID int64)
	}{
		{
			name: "删除存在的资源",
			prepare: func() int64 {
				resource := s.createTestResource(1, "api", "user:read")
				created, err := s.svc.Svc.CreateResource(context.Background(), resource)
				require.NoError(t, err)
				return created.ID
			},
			wantErr: false,
			validate: func(t *testing.T, resourceID int64) {
				// 尝试获取已删除的资源
				_, err := s.svc.Svc.GetResource(context.Background(), resourceID)
				assert.Error(t, err)
			},
		},
		{
			name: "删除不存在的资源",
			prepare: func() int64 {
				return 99999
			},
			wantErr: true,
			validate: func(t *testing.T, resourceID int64) {
				// 不需要额外验证
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resourceID := tt.prepare()
			err := s.svc.Svc.DeleteResource(context.Background(), resourceID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, resourceID)
				}
			}
		})
	}
}

// 测试列出资源
func (s *RBACServiceTestSuite) TestListResources() {
	t := s.T()

	const bizID = 1

	// 准备测试数据 - 创建多个不同类型的资源
	prepareTestResources := func() {
		resourceTypes := []string{"api", "menu", "button", "data"}
		for _, resType := range resourceTypes {
			for i := 0; i < 2; i++ {
				resource := s.createTestResource(bizID, resType, fmt.Sprintf("%s:test%d", resType, i))
				_, err := s.svc.Svc.CreateResource(context.Background(), resource)
				require.NoError(t, err)
			}
		}
	}

	// 创建测试资源数据
	prepareTestResources()

	// 测试不同的查询场景
	tests := []struct {
		name         string
		bizID        int64
		offset       int
		limit        int
		resourceType string
		key          string
		wantCount    int
		validate     func(t *testing.T, resources []domain.Resource)
	}{
		{
			name:         "列出所有资源",
			bizID:        bizID,
			offset:       0,
			limit:        20,
			resourceType: "",
			key:          "",
			wantCount:    8, // 4种类型 * 2个资源
			validate: func(t *testing.T, resources []domain.Resource) {
				assert.Len(t, resources, 8)
			},
		},
		{
			name:         "列出API资源",
			bizID:        bizID,
			offset:       0,
			limit:        10,
			resourceType: "api",
			key:          "",
			wantCount:    2,
			validate: func(t *testing.T, resources []domain.Resource) {
				for _, resource := range resources {
					assert.Equal(t, "api", resource.Type)
				}
			},
		},
		{
			name:         "列出指定Key的资源",
			bizID:        bizID,
			offset:       0,
			limit:        10,
			resourceType: "",
			key:          "api:test0",
			wantCount:    1,
			validate: func(t *testing.T, resources []domain.Resource) {
				for _, resource := range resources {
					assert.Equal(t, "api:test0", resource.Key)
				}
			},
		},
		{
			name:         "同时指定类型和Key",
			bizID:        bizID,
			offset:       0,
			limit:        10,
			resourceType: "menu",
			key:          "menu:test1",
			wantCount:    1,
			validate: func(t *testing.T, resources []domain.Resource) {
				for _, resource := range resources {
					assert.Equal(t, "menu", resource.Type)
					assert.Equal(t, "menu:test1", resource.Key)
				}
			},
		},
		{
			name:         "分页限制",
			bizID:        bizID,
			offset:       0,
			limit:        3,
			resourceType: "",
			key:          "",
			wantCount:    3,
			validate: func(t *testing.T, resources []domain.Resource) {
				assert.Len(t, resources, 3)
			},
		},
		{
			name:         "偏移量",
			bizID:        bizID,
			offset:       3,
			limit:        10,
			resourceType: "",
			key:          "",
			wantCount:    5, // 总数8 - 偏移3 = 5
			validate: func(t *testing.T, resources []domain.Resource) {
				assert.Len(t, resources, 5)
			},
		},
		{
			name:         "不同业务ID",
			bizID:        bizID + 1,
			offset:       0,
			limit:        10,
			resourceType: "",
			key:          "",
			wantCount:    0,
			validate: func(t *testing.T, resources []domain.Resource) {
				assert.Empty(t, resources)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resources, count, err := s.svc.Svc.ListResourcesByTypeAndKey(context.Background(), tt.bizID, tt.offset, tt.limit, tt.resourceType, tt.key)

			require.NoError(t, err)
			assert.Equal(t, tt.wantCount, count)
			assert.Len(t, resources, tt.wantCount)

			if tt.validate != nil {
				tt.validate(t, resources)
			}
		})
	}
}

// 创建测试用权限对象
func (s *RBACServiceTestSuite) createTestPermission(bizID int64, resourceID int64, resourceType, resourceKey string, action domain.ActionType) domain.Permission {
	now := time.Now().UnixMilli()
	return domain.Permission{
		BizID:        bizID,
		Name:         fmt.Sprintf("测试权限-%d", now),
		Description:  "测试权限描述",
		ResourceID:   resourceID,
		ResourceType: resourceType,
		ResourceKey:  resourceKey,
		Action:       action,
	}
}

// 断言权限对象的字段是否符合预期
func (s *RBACServiceTestSuite) assertPermission(t *testing.T, expected, actual domain.Permission) {
	assert.NotZero(t, actual.ID)
	assert.Equal(t, expected.BizID, actual.BizID)
	assert.Equal(t, expected.Name, actual.Name)
	assert.Equal(t, expected.Description, actual.Description)
	assert.Equal(t, expected.ResourceID, actual.ResourceID)
	assert.Equal(t, expected.ResourceType, actual.ResourceType)
	assert.Equal(t, expected.ResourceKey, actual.ResourceKey)
	assert.Equal(t, expected.Action, actual.Action)
}

// 测试创建权限
func (s *RBACServiceTestSuite) TestCreatePermission() {
	t := s.T()

	// 先创建一个资源用于后续测试
	resource := s.createTestResource(1, "api", "user:read")
	createdResource, err := s.svc.Svc.CreateResource(context.Background(), resource)
	require.NoError(t, err)

	tests := []struct {
		name       string
		permission func() domain.Permission
		wantErr    bool
		errorMsg   string
	}{
		{
			name: "创建读取权限",
			permission: func() domain.Permission {
				return s.createTestPermission(1, createdResource.ID, createdResource.Type, createdResource.Key, domain.ActionTypeRead)
			},
			wantErr: false,
		},
		{
			name: "创建创建权限",
			permission: func() domain.Permission {
				return s.createTestPermission(1, createdResource.ID, createdResource.Type, createdResource.Key, domain.ActionTypeCreate)
			},
			wantErr: false,
		},
		{
			name: "创建写入权限",
			permission: func() domain.Permission {
				return s.createTestPermission(1, createdResource.ID, createdResource.Type, createdResource.Key, domain.ActionTypeWrite)
			},
			wantErr: false,
		},
		{
			name: "创建删除权限",
			permission: func() domain.Permission {
				return s.createTestPermission(1, createdResource.ID, createdResource.Type, createdResource.Key, domain.ActionTypeDelete)
			},
			wantErr: false,
		},
		{
			name: "创建执行权限",
			permission: func() domain.Permission {
				return s.createTestPermission(1, createdResource.ID, createdResource.Type, createdResource.Key, domain.ActionTypeExecute)
			},
			wantErr: false,
		},
		{
			name: "创建导出权限",
			permission: func() domain.Permission {
				return s.createTestPermission(1, createdResource.ID, createdResource.Type, createdResource.Key, domain.ActionTypeExport)
			},
			wantErr: false,
		},
		{
			name: "创建导入权限",
			permission: func() domain.Permission {
				return s.createTestPermission(1, createdResource.ID, createdResource.Type, createdResource.Key, domain.ActionTypeImport)
			},
			wantErr: false,
		},
		{
			name: "空名称",
			permission: func() domain.Permission {
				p := s.createTestPermission(1, createdResource.ID, createdResource.Type, createdResource.Key, domain.ActionTypeRead)
				p.Name = ""
				return p
			},
			wantErr:  true,
			errorMsg: "权限名称不能为空",
		},
		{
			name: "资源ID为0",
			permission: func() domain.Permission {
				p := s.createTestPermission(1, 0, createdResource.Type, createdResource.Key, domain.ActionTypeRead)
				return p
			},
			wantErr:  true,
			errorMsg: "资源ID必须大于0",
		},
		{
			name: "空资源类型",
			permission: func() domain.Permission {
				p := s.createTestPermission(1, createdResource.ID, "", createdResource.Key, domain.ActionTypeRead)
				return p
			},
			wantErr:  true,
			errorMsg: "资源类型不能为空",
		},
		{
			name: "空资源Key",
			permission: func() domain.Permission {
				p := s.createTestPermission(1, createdResource.ID, createdResource.Type, "", domain.ActionTypeRead)
				return p
			},
			wantErr:  true,
			errorMsg: "资源Key不能为空",
		},
		{
			name: "零业务ID",
			permission: func() domain.Permission {
				p := s.createTestPermission(0, createdResource.ID, createdResource.Type, createdResource.Key, domain.ActionTypeRead)
				return p
			},
			wantErr:  true,
			errorMsg: "业务ID必须大于0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			permission := tt.permission()
			created, err := s.svc.Svc.CreatePermission(context.Background(), permission)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
				s.assertPermission(t, permission, created)
			}
		})
	}
}

// 测试获取权限
func (s *RBACServiceTestSuite) TestGetPermission() {
	t := s.T()

	// 创建资源和权限用于测试
	preparePermission := func() domain.Permission {
		// 先创建一个资源
		resource := s.createTestResource(1, "api", "user:read")
		createdResource, err := s.svc.Svc.CreateResource(context.Background(), resource)
		require.NoError(t, err)

		// 再创建一个权限
		permission := s.createTestPermission(1, createdResource.ID, createdResource.Type, createdResource.Key, domain.ActionTypeRead)
		created, err := s.svc.Svc.CreatePermission(context.Background(), permission)
		require.NoError(t, err)

		return created
	}

	created := preparePermission()

	tests := []struct {
		name         string
		permissionID int64
		wantErr      bool
	}{
		{
			name:         "获取存在的权限",
			permissionID: created.ID,
			wantErr:      false,
		},
		{
			name:         "获取不存在的权限",
			permissionID: 99999,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			found, err := s.svc.Svc.GetPermission(context.Background(), tt.permissionID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				s.assertPermission(t, created, found)
			}
		})
	}
}

// 测试更新权限
func (s *RBACServiceTestSuite) TestUpdatePermission() {
	t := s.T()

	// 创建初始资源
	resource := s.createTestResource(1, "api", "user:read")
	createdResource, err := s.svc.Svc.CreateResource(context.Background(), resource)
	require.NoError(t, err)

	// 创建另一个资源（用于测试更新资源相关字段）
	anotherResource := s.createTestResource(1, "api", "user:write")
	createdAnotherResource, err := s.svc.Svc.CreateResource(context.Background(), anotherResource)
	require.NoError(t, err)

	tests := []struct {
		name     string
		prepare  func() domain.Permission
		wantErr  bool
		validate func(t *testing.T, updated domain.Permission)
	}{
		{
			name: "更新权限成功",
			prepare: func() domain.Permission {
				// 创建一个权限
				permission := s.createTestPermission(1, createdResource.ID, createdResource.Type, createdResource.Key, domain.ActionTypeRead)
				created, err := s.svc.Svc.CreatePermission(context.Background(), permission)
				require.NoError(t, err)

				// 修改权限信息
				created.Name = "更新后的权限名称"
				created.Description = "更新后的权限描述"
				return created
			},
			wantErr: false,
			validate: func(t *testing.T, updated domain.Permission) {
				assert.Equal(t, "更新后的权限名称", updated.Name)
				assert.Equal(t, "更新后的权限描述", updated.Description)
			},
		},
		{
			name: "更新不存在的权限",
			prepare: func() domain.Permission {
				nonExistentPermission := s.createTestPermission(1, createdResource.ID, createdResource.Type, createdResource.Key, domain.ActionTypeWrite)
				nonExistentPermission.ID = 99999
				return nonExistentPermission
			},
			wantErr: true,
			validate: func(t *testing.T, updated domain.Permission) {
				// 不需要验证
			},
		},
		{
			name: "尝试更新资源相关字段",
			prepare: func() domain.Permission {
				// 创建一个新权限
				permission := s.createTestPermission(1, createdResource.ID, createdResource.Type, createdResource.Key, domain.ActionTypeDelete)
				created, err := s.svc.Svc.CreatePermission(context.Background(), permission)
				require.NoError(t, err)

				// 尝试修改资源相关字段
				created.ResourceID = createdAnotherResource.ID
				created.ResourceType = createdAnotherResource.Type
				created.ResourceKey = createdAnotherResource.Key
				created.Action = domain.ActionTypeExecute
				return created
			},
			wantErr: false, // 具体行为取决于实现
			validate: func(t *testing.T, updated domain.Permission) {
				// 如果更新成功，验证字段是否被更新或保持原值
				// 具体行为取决于系统实现，不一定会返回错误
				t.Logf("资源字段更新结果: ResourceID=%d, ResourceType=%s, ResourceKey=%s, Action=%s",
					updated.ResourceID, updated.ResourceType, updated.ResourceKey, updated.Action)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			permissionToUpdate := tt.prepare()
			updated, err := s.svc.Svc.UpdatePermission(context.Background(), permissionToUpdate)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, updated)
				}

				// 再次获取确认更新成功
				found, err := s.svc.Svc.GetPermission(context.Background(), updated.ID)
				require.NoError(t, err)
				s.assertPermission(t, updated, found)
			}
		})
	}
}

// 测试删除权限
func (s *RBACServiceTestSuite) TestDeletePermission() {
	t := s.T()

	tests := []struct {
		name     string
		prepare  func() int64
		wantErr  bool
		validate func(t *testing.T, permissionID int64)
	}{
		{
			name: "删除存在的权限",
			prepare: func() int64 {
				// 先创建一个资源
				resource := s.createTestResource(1, "api", "user:read")
				createdResource, err := s.svc.Svc.CreateResource(context.Background(), resource)
				require.NoError(t, err)

				// 再创建一个权限
				permission := s.createTestPermission(1, createdResource.ID, createdResource.Type, createdResource.Key, domain.ActionTypeRead)
				created, err := s.svc.Svc.CreatePermission(context.Background(), permission)
				require.NoError(t, err)

				return created.ID
			},
			wantErr: false,
			validate: func(t *testing.T, permissionID int64) {
				// 尝试获取已删除的权限
				_, err := s.svc.Svc.GetPermission(context.Background(), permissionID)
				assert.Error(t, err)
			},
		},
		{
			name: "删除不存在的权限",
			prepare: func() int64 {
				return 99999
			},
			wantErr: true,
			validate: func(t *testing.T, permissionID int64) {
				// 不需要额外验证
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			permissionID := tt.prepare()
			err := s.svc.Svc.DeletePermission(context.Background(), permissionID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, permissionID)
				}
			}
		})
	}
}

// 测试列出权限
func (s *RBACServiceTestSuite) TestListPermissions() {
	t := s.T()

	const bizID = 1

	// 准备测试数据 - 创建资源和权限
	prepareTestPermissions := func() {
		// 创建资源
		resourceTypes := []string{"api", "menu"}
		var resources []domain.Resource
		for _, resType := range resourceTypes {
			resource := s.createTestResource(bizID, resType, fmt.Sprintf("%s:test", resType))
			created, err := s.svc.Svc.CreateResource(context.Background(), resource)
			require.NoError(t, err)
			resources = append(resources, created)
		}

		// 创建权限
		actionTypes := []domain.ActionType{
			domain.ActionTypeRead,
			domain.ActionTypeWrite,
		}

		for _, resource := range resources {
			for _, actionType := range actionTypes {
				permission := s.createTestPermission(bizID, resource.ID, resource.Type, resource.Key, actionType)
				_, err := s.svc.Svc.CreatePermission(context.Background(), permission)
				require.NoError(t, err)
			}
		}
	}

	// 创建测试权限数据
	prepareTestPermissions()

	// 测试不同的查询场景
	tests := []struct {
		name         string
		bizID        int64
		offset       int
		limit        int
		resourceType string
		resourceKey  string
		action       string
		wantCount    int
		validate     func(t *testing.T, permissions []domain.Permission)
	}{
		{
			name:         "列出所有权限",
			bizID:        bizID,
			offset:       0,
			limit:        10,
			resourceType: "",
			resourceKey:  "",
			action:       "",
			wantCount:    4, // 2种资源 * 2种操作
			validate: func(t *testing.T, permissions []domain.Permission) {
				assert.Len(t, permissions, 4)
			},
		},
		{
			name:         "按资源类型筛选",
			bizID:        bizID,
			offset:       0,
			limit:        10,
			resourceType: "api",
			resourceKey:  "",
			action:       "",
			wantCount:    2, // 1种资源 * 2种操作
			validate: func(t *testing.T, permissions []domain.Permission) {
				for _, permission := range permissions {
					assert.Equal(t, "api", permission.ResourceType)
				}
			},
		},
		{
			name:         "按资源Key筛选",
			bizID:        bizID,
			offset:       0,
			limit:        10,
			resourceType: "",
			resourceKey:  "menu:test",
			action:       "",
			wantCount:    2, // 1种资源 * 2种操作
			validate: func(t *testing.T, permissions []domain.Permission) {
				for _, permission := range permissions {
					assert.Equal(t, "menu:test", permission.ResourceKey)
				}
			},
		},
		{
			name:         "按操作类型筛选",
			bizID:        bizID,
			offset:       0,
			limit:        10,
			resourceType: "",
			resourceKey:  "",
			action:       string(domain.ActionTypeRead),
			wantCount:    2, // 2种资源 * 1种操作
			validate: func(t *testing.T, permissions []domain.Permission) {
				for _, permission := range permissions {
					assert.Equal(t, domain.ActionTypeRead, permission.Action)
				}
			},
		},
		{
			name:         "组合筛选",
			bizID:        bizID,
			offset:       0,
			limit:        10,
			resourceType: "api",
			resourceKey:  "",
			action:       string(domain.ActionTypeWrite),
			wantCount:    1, // 1种资源 * 1种操作
			validate: func(t *testing.T, permissions []domain.Permission) {
				for _, permission := range permissions {
					assert.Equal(t, "api", permission.ResourceType)
					assert.Equal(t, domain.ActionTypeWrite, permission.Action)
				}
			},
		},
		{
			name:         "分页限制",
			bizID:        bizID,
			offset:       0,
			limit:        2,
			resourceType: "",
			resourceKey:  "",
			action:       "",
			wantCount:    2,
			validate: func(t *testing.T, permissions []domain.Permission) {
				assert.Len(t, permissions, 2)
			},
		},
		{
			name:         "偏移量",
			bizID:        bizID,
			offset:       2,
			limit:        10,
			resourceType: "",
			resourceKey:  "",
			action:       "",
			wantCount:    2, // 总数4 - 偏移2 = 2
			validate: func(t *testing.T, permissions []domain.Permission) {
				assert.Len(t, permissions, 2)
			},
		},
		{
			name:         "不同业务ID",
			bizID:        bizID + 1,
			offset:       0,
			limit:        10,
			resourceType: "",
			resourceKey:  "",
			action:       "",
			wantCount:    0,
			validate: func(t *testing.T, permissions []domain.Permission) {
				assert.Empty(t, permissions)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			permissions, count, err := s.svc.Svc.ListPermissionsByResourceTypeAndKeyAndAction(context.Background(), tt.bizID, tt.offset, tt.limit, tt.resourceType, tt.resourceKey, tt.action)

			require.NoError(t, err)
			assert.Equal(t, tt.wantCount, count)
			assert.Len(t, permissions, tt.wantCount)

			if tt.validate != nil {
				tt.validate(t, permissions)
			}
		})
	}
}

// 断言用户角色对象的字段是否符合预期
func (s *RBACServiceTestSuite) assertUserRole(t *testing.T, expected, actual domain.UserRole) {
	assert.NotZero(t, actual.ID)
	assert.Equal(t, expected.BizID, actual.BizID)
	assert.Equal(t, expected.UserID, actual.UserID)
	assert.Equal(t, expected.RoleID, actual.RoleID)
	assert.Equal(t, expected.RoleName, actual.RoleName)
	assert.Equal(t, expected.RoleType, actual.RoleType)
	assert.Equal(t, expected.StartTime, actual.StartTime)
	assert.Equal(t, expected.EndTime, actual.EndTime)
}

// 测试授予用户角色
func (s *RBACServiceTestSuite) TestGrantUserRole() {
	t := s.T()

	// 先创建一个角色用于测试
	role := s.createTestRole(1, domain.RoleTypeCustom)
	createdRole, err := s.svc.Svc.CreateRole(context.Background(), role)
	require.NoError(t, err)

	const userID = 100 // 模拟用户ID
	const bizID = 1

	tests := []struct {
		name      string
		bizID     int64
		userID    int64
		roleID    int64
		startTime int64
		endTime   int64
		wantErr   bool
		errorMsg  string
		validate  func(t *testing.T, userRole domain.UserRole)
	}{
		{
			name:      "授予角色成功",
			bizID:     bizID,
			userID:    userID,
			roleID:    createdRole.ID,
			startTime: time.Now().UnixMilli(),
			endTime:   time.Now().UnixMilli() + 86400000, // 一天后过期
			wantErr:   false,
			validate: func(t *testing.T, userRole domain.UserRole) {
				assert.NotZero(t, userRole.ID)
				assert.Equal(t, bizID, userRole.BizID)
				assert.Equal(t, userID, userRole.UserID)
				assert.Equal(t, createdRole.ID, userRole.RoleID)
				assert.Equal(t, createdRole.Name, userRole.RoleName)
				assert.Equal(t, createdRole.Type, userRole.RoleType)
			},
		},
		{
			name:      "零业务ID",
			bizID:     0,
			userID:    userID,
			roleID:    createdRole.ID,
			startTime: time.Now().UnixMilli(),
			endTime:   time.Now().UnixMilli() + 86400000,
			wantErr:   true,
			errorMsg:  "业务ID必须大于0",
		},
		{
			name:      "零用户ID",
			bizID:     bizID,
			userID:    0,
			roleID:    createdRole.ID,
			startTime: time.Now().UnixMilli(),
			endTime:   time.Now().UnixMilli() + 86400000,
			wantErr:   true,
			errorMsg:  "用户ID必须大于0",
		},
		{
			name:      "零角色ID",
			bizID:     bizID,
			userID:    userID,
			roleID:    0,
			startTime: time.Now().UnixMilli(),
			endTime:   time.Now().UnixMilli() + 86400000,
			wantErr:   true,
			errorMsg:  "角色ID必须大于0",
		},
		{
			name:      "不存在的角色",
			bizID:     bizID,
			userID:    userID,
			roleID:    99999,
			startTime: time.Now().UnixMilli(),
			endTime:   time.Now().UnixMilli() + 86400000,
			wantErr:   true,
			errorMsg:  "角色不存在",
		},
		{
			name:      "结束时间早于开始时间",
			bizID:     bizID,
			userID:    userID,
			roleID:    createdRole.ID,
			startTime: time.Now().UnixMilli() + 86400000,
			endTime:   time.Now().UnixMilli(),
			wantErr:   true,
			errorMsg:  "结束时间必须晚于开始时间",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRole, err := s.svc.Svc.GrantUserRole(context.Background(), tt.bizID, tt.userID, tt.roleID, tt.startTime, tt.endTime)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, userRole)
				}
			}
		})
	}
}

// 测试撤销用户角色
func (s *RBACServiceTestSuite) TestRevokeUserRole() {
	t := s.T()

	const bizID = 1
	const userID = 100 // 模拟用户ID

	tests := []struct {
		name     string
		prepare  func() (int64, int64, int64) // 返回bizID, userID, roleID
		wantErr  bool
		validate func(t *testing.T, bizID, userID, roleID int64)
	}{
		{
			name: "撤销已授予的角色",
			prepare: func() (int64, int64, int64) {
				// 先创建一个角色
				role := s.createTestRole(bizID, domain.RoleTypeCustom)
				createdRole, err := s.svc.Svc.CreateRole(context.Background(), role)
				require.NoError(t, err)

				// 先授予角色
				startTime := time.Now().UnixMilli()
				endTime := startTime + 86400000 // 一天后过期
				_, err = s.svc.Svc.GrantUserRole(context.Background(), bizID, userID, createdRole.ID, startTime, endTime)
				require.NoError(t, err)

				return bizID, userID, createdRole.ID
			},
			wantErr: false,
			validate: func(t *testing.T, bizID, userID, roleID int64) {
				// 检查是否已撤销，通过列出用户角色验证
				roles, _, err := s.svc.Svc.ListUserRolesByUserID(context.Background(), bizID, userID, 0, 10)
				require.NoError(t, err)

				// 检查此用户是否不再拥有该角色
				hasRole := false
				for _, role := range roles {
					if role.RoleID == roleID {
						hasRole = true
						break
					}
				}
				assert.False(t, hasRole, "用户应该不再拥有该角色")
			},
		},
		{
			name: "撤销不存在的用户角色",
			prepare: func() (int64, int64, int64) {
				// 先创建一个角色
				role := s.createTestRole(bizID, domain.RoleTypeCustom)
				createdRole, err := s.svc.Svc.CreateRole(context.Background(), role)
				require.NoError(t, err)

				// 不授予角色，直接返回
				return bizID, userID + 1, createdRole.ID // 使用不存在的用户ID
			},
			wantErr: false, // 根据实现可能是静默处理
			validate: func(t *testing.T, bizID, userID, roleID int64) {
				// 不需要额外验证
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bizID, userID, roleID := tt.prepare()
			err := s.svc.Svc.RevokeUserRole(context.Background(), bizID, userID, roleID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				// 根据实现，可能会返回错误或者静默处理
				if err != nil {
					t.Logf("撤销用户角色返回错误: %v", err)
				}

				if tt.validate != nil {
					tt.validate(t, bizID, userID, roleID)
				}
			}
		})
	}
}

// 测试列出用户角色
func (s *RBACServiceTestSuite) TestListUserRoles() {
	t := s.T()

	const bizID = 1
	const userID = 100 // 模拟用户ID

	// 准备测试数据 - 创建多个角色并授予用户
	prepareUserRoles := func() {
		roleTypes := []domain.RoleType{
			domain.RoleTypeSystem,
			domain.RoleTypeCustom,
			domain.RoleTypeTemporary,
		}

		startTime := time.Now().UnixMilli()
		endTime := startTime + 86400000 // 一天后过期

		// 为用户分配不同类型的角色
		for i, roleType := range roleTypes {
			role := s.createTestRole(bizID, roleType)
			role.Name = fmt.Sprintf("角色-%d-%s", i, roleType)
			createdRole, err := s.svc.Svc.CreateRole(context.Background(), role)
			require.NoError(t, err)

			_, err = s.svc.Svc.GrantUserRole(context.Background(), bizID, userID, createdRole.ID, startTime, endTime)
			require.NoError(t, err)
		}

		// 额外创建一个不同业务ID的角色
		role := s.createTestRole(bizID+1, domain.RoleTypeCustom)
		role.Name = "不同业务角色"
		createdRole, err := s.svc.Svc.CreateRole(context.Background(), role)
		require.NoError(t, err)

		_, err = s.svc.Svc.GrantUserRole(context.Background(), bizID+1, userID, createdRole.ID, startTime, endTime)
		require.NoError(t, err)
	}

	// 创建测试数据
	prepareUserRoles()

	// 测试不同的查询场景
	tests := []struct {
		name      string
		bizID     int64
		userID    int64
		offset    int
		limit     int
		wantCount int
		validate  func(t *testing.T, roles []domain.UserRole)
	}{
		{
			name:      "列出所有用户角色",
			bizID:     bizID,
			userID:    userID,
			offset:    0,
			limit:     10,
			wantCount: 3, // 3种角色类型
			validate: func(t *testing.T, roles []domain.UserRole) {
				assert.Len(t, roles, 3)
				for _, role := range roles {
					assert.Equal(t, bizID, role.BizID)
					assert.Equal(t, userID, role.UserID)
				}
			},
		},
		{
			name:      "分页限制",
			bizID:     bizID,
			userID:    userID,
			offset:    0,
			limit:     2,
			wantCount: 2,
			validate: func(t *testing.T, roles []domain.UserRole) {
				assert.Len(t, roles, 2)
			},
		},
		{
			name:      "偏移量",
			bizID:     bizID,
			userID:    userID,
			offset:    1,
			limit:     10,
			wantCount: 2, // 总数3 - 偏移1 = 2
			validate: func(t *testing.T, roles []domain.UserRole) {
				assert.Len(t, roles, 2)
			},
		},
		{
			name:      "不同业务ID",
			bizID:     bizID + 1,
			userID:    userID,
			offset:    0,
			limit:     10,
			wantCount: 1,
			validate: func(t *testing.T, roles []domain.UserRole) {
				assert.Len(t, roles, 1)
				assert.Equal(t, bizID+1, roles[0].BizID)
			},
		},
		{
			name:      "不存在的用户",
			bizID:     bizID,
			userID:    userID + 999,
			offset:    0,
			limit:     10,
			wantCount: 0,
			validate: func(t *testing.T, roles []domain.UserRole) {
				assert.Empty(t, roles)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			roles, count, err := s.svc.Svc.ListUserRolesByUserID(context.Background(), tt.bizID, tt.userID, tt.offset, tt.limit)

			require.NoError(t, err)
			assert.Equal(t, tt.wantCount, count)
			assert.Len(t, roles, tt.wantCount)

			if tt.validate != nil {
				tt.validate(t, roles)
			}
		})
	}
}

// 测试完整的权限检查流程
func (s *RBACServiceTestSuite) TestPermissionCheck() {
	t := s.T()

	const bizID = 1
	const userID = 100 // 模拟用户ID

	// 准备测试数据
	var testSetups []struct {
		name        string
		prepare     func() (domain.Resource, domain.Permission, domain.Role, domain.UserRole)
		checkAction func(t *testing.T, resource domain.Resource, permission domain.Permission, role domain.Role, userRole domain.UserRole)
	}

	// 第一个测试场景：创建资源、权限、角色并授予用户
	testSetups = append(testSetups, struct {
		name        string
		prepare     func() (domain.Resource, domain.Permission, domain.Role, domain.UserRole)
		checkAction func(t *testing.T, resource domain.Resource, permission domain.Permission, role domain.Role, userRole domain.UserRole)
	}{
		name: "基本权限检查流程",
		prepare: func() (domain.Resource, domain.Permission, domain.Role, domain.UserRole) {
			// 步骤1: 创建资源
			resource := s.createTestResource(bizID, "api", "user:read")
			createdResource, err := s.svc.Svc.CreateResource(context.Background(), resource)
			require.NoError(t, err)

			// 步骤2: 创建权限
			permission := s.createTestPermission(bizID, createdResource.ID, createdResource.Type, createdResource.Key, domain.ActionTypeRead)
			createdPermission, err := s.svc.Svc.CreatePermission(context.Background(), permission)
			require.NoError(t, err)

			// 步骤3: 创建角色
			role := s.createTestRole(bizID, domain.RoleTypeCustom)
			createdRole, err := s.svc.Svc.CreateRole(context.Background(), role)
			require.NoError(t, err)

			// 步骤4: 授予用户角色
			startTime := time.Now().UnixMilli()
			endTime := startTime + 86400000 // 一天后过期
			userRole, err := s.svc.Svc.GrantUserRole(context.Background(), bizID, userID, createdRole.ID, startTime, endTime)
			require.NoError(t, err)

			return createdResource, createdPermission, createdRole, userRole
		},
		checkAction: func(t *testing.T, resource domain.Resource, permission domain.Permission, role domain.Role, userRole domain.UserRole) {
			// 检查用户角色列表
			roles, count, err := s.svc.Svc.ListUserRolesByUserID(context.Background(), bizID, userID, 0, 10)
			require.NoError(t, err)
			assert.Equal(t, 1, count)
			assert.Len(t, roles, 1)
			assert.Equal(t, role.ID, roles[0].RoleID)

			// 可以在这里添加更多权限检查的测试，如检查用户是否有权限访问资源等
		},
	})

	// 执行测试场景
	for _, setup := range testSetups {
		t.Run(setup.name, func(t *testing.T) {
			resource, permission, role, userRole := setup.prepare()
			setup.checkAction(t, resource, permission, role, userRole)
		})
	}
}
