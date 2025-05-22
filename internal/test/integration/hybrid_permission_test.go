package integration

import (
	"fmt"
	"gitee.com/flycash/permission-platform/internal/domain"
	"gitee.com/flycash/permission-platform/internal/repository"
	"gitee.com/flycash/permission-platform/internal/repository/dao"
	abacsvc "gitee.com/flycash/permission-platform/internal/service/abac"
	"gitee.com/flycash/permission-platform/internal/service/hybrid"
	"gitee.com/flycash/permission-platform/internal/service/rbac"
	"gitee.com/flycash/permission-platform/internal/test/integration/ioc/abac"
	rbacioc "gitee.com/flycash/permission-platform/internal/test/integration/ioc/rbac"
	testioc "gitee.com/flycash/permission-platform/internal/test/ioc"
	"github.com/ego-component/egorm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

type HybridPermissionSuite struct {
	suite.Suite
	db                *egorm.Component
	abacPermissionSvc abacsvc.PermissionSvc
	valRepo           repository.AttributeValueRepository
	definitionRepo    repository.AttributeDefinitionRepository
	permissionRepo    repository.PermissionRepository
	resourceRepo      repository.ResourceRepository
	policyRepo        repository.PolicyRepo
	svc               *rbacioc.Service
	rbacSvc           rbac.Service
	hybridSvc         hybrid.PermissionService
}

func (s *HybridPermissionSuite) SetupSuite() {
	s.db = testioc.InitDBAndTables()
	s.svc = rbacioc.Init()
	s.rbacSvc = s.svc.Svc
	svc := abac.Init(s.db)
	s.definitionRepo = svc.DefinitionRepo
	s.abacPermissionSvc = svc.PermissionSvc
	s.valRepo = svc.ValRepo
	s.policyRepo = svc.PolicyRepo
	s.permissionRepo = svc.PermissionRepo
	s.resourceRepo = svc.ResourceRepo
	s.hybridSvc = hybrid.NewRoleAsAttributePermissionService(s.rbacSvc, s.abacPermissionSvc)
}
func (s *HybridPermissionSuite) setupDefinitionV1(bizId int64) {
	// 初始化属性定义
	subjectAttrDef1 := domain.AttributeDefinition{
		ID:          20001,
		Name:        "role",
		DataType:    domain.DataTypeArray,
		Description: "角色",
		EntityType:  domain.EntityTypeSubject,
	}
	envAttrDef2 := domain.AttributeDefinition{
		ID:          20002,
		Name:        "time",
		DataType:    domain.DataTypeDatetime,
		Description: "时间",
		EntityType:  domain.EntityTypeEnvironment,
	}
	envAttrDef3 := domain.AttributeDefinition{
		ID:          20003,
		Name:        "location",
		DataType:    domain.DataTypeString,
		Description: "地点",
		EntityType:  domain.EntityTypeEnvironment,
	}
	resourceDef := domain.AttributeDefinition{
		ID:          20004,
		Name:        "codeRepo",
		DataType:    domain.DataTypeString,
		Description: "代码仓库",
		EntityType:  domain.EntityTypeResource,
	}
	defs := []domain.AttributeDefinition{
		subjectAttrDef1,
		envAttrDef2,
		envAttrDef3,
		resourceDef,
	}
	for idx := range defs {
		_, err := s.definitionRepo.Save(s.T().Context(), bizId, defs[idx])
		require.NoError(s.T(), err)
	}
}

func (s *HybridPermissionSuite) setupUserRole(userId, bizId int64) {
	now := time.Now()
	userRole1 := dao.UserRole{
		BizID:     bizId,
		UserID:    userId,
		RoleID:    1,
		RoleName:  "程序员",
		RoleType:  "user",
		StartTime: now.Add(-1 * time.Hour).UnixMilli(),
		EndTime:   now.Add(1 * time.Hour).UnixMilli(),
		Ctime:     now.UnixMilli(),
		Utime:     now.UnixMilli(),
	}
	userRole2 := dao.UserRole{
		BizID:     bizId,
		UserID:    userId,
		RoleID:    2,
		RoleName:  "cto",
		RoleType:  "user",
		StartTime: now.Add(-1 * time.Hour).UnixMilli(),
		EndTime:   now.Add(1 * time.Hour).UnixMilli(),
		Ctime:     now.UnixMilli(),
		Utime:     now.UnixMilli(),
	}
	userRole3 := dao.UserRole{
		BizID:     bizId,
		UserID:    userId,
		RoleID:    3,
		RoleName:  "经理",
		RoleType:  "user",
		StartTime: now.Add(-1 * time.Hour).UnixMilli(),
		EndTime:   now.Add(1 * time.Hour).UnixMilli(),
		Ctime:     now.UnixMilli(),
		Utime:     now.UnixMilli(),
	}
	roles := []dao.UserRole{
		userRole1,
		userRole2,
		userRole3,
	}
	for idx := range roles {
		err := s.db.WithContext(s.T().Context()).Create(&roles[idx]).Error
		require.NoError(s.T(), err)
	}
}

func (s *HybridPermissionSuite) TestCheck() {
	t := s.T()
	bizId := int64(20000)
	userId := int64(22)
	testcase := []struct {
		name       string
		before     func(t *testing.T)
		permission domain.Permission
		attrs      domain.Attributes
		uid        int64
		wantVal    bool
	}{
		{
			name: "cto,程序员在某个时间段，某个地点可以访问order.com代码仓库",
			uid:  userId,
			before: func(t *testing.T) {
				// 初始化属性
				s.setupDefinitionV1(bizId)
				// 初始化role
				s.setupUserRole(userId, bizId)

				// 创建资源属性值
				_, err := s.valRepo.SaveResourceValue(t.Context(), bizId, 33, domain.AttributeValue{
					Definition: domain.AttributeDefinition{
						ID: 20004,
					},
					Value: "order.com",
				})
				require.NoError(s.T(), err)
				// 创建权限
				permission := domain.Permission{
					BizID:       bizId,
					Name:        "代码仓库权限",
					Description: "代码仓库权限",
					Resource: domain.Resource{
						Type: "code",
						Key:  "order.com",
					},
					Action:   "push",
					Metadata: "[1]",
				}
				per, err := s.permissionRepo.Create(t.Context(), permission)
				require.NoError(t, err)
				// 创建资源
				res := domain.Resource{
					ID:          33,
					BizID:       bizId,
					Type:        "code",
					Name:        "order.com",
					Description: "用户",
					Key:         "order.com",
					Metadata:    "[1]",
				}
				_, err = s.resourceRepo.Create(t.Context(), res)
				require.NoError(t, err)
				startTime := time.Now().Add(-time.Hour)
				startStr := fmt.Sprintf("@day(%s)", startTime.Format("15:04"))
				endTime := time.Now().Add(1 * time.Hour)
				endStr := fmt.Sprintf("@day(%s)", endTime.Format("15:04"))
				policy := domain.Policy{
					BizID:       bizId,
					Name:        "代码仓库读取策略",
					Description: "允许用户读取代码仓库",
					Status:      domain.PolicyStatusActive,
				}
				id, err := s.policyRepo.Save(t.Context(), policy)
				require.NoError(t, err)
				rules := []domain.PolicyRule{
					{
						ID: 20010,
						AttrDef: domain.AttributeDefinition{
							ID: 20001,
						},
						Operator: domain.AnyMatch,
						Value:    "[\"程序员\",\"老板\"]",
					},
					{
						ID: 20011,
						AttrDef: domain.AttributeDefinition{
							ID: 20002,
						},
						Operator: domain.GreaterOrEqual,
						Value:    startStr,
					},
					{
						ID: 20012,
						AttrDef: domain.AttributeDefinition{
							ID: 20002,
						},
						Operator: domain.LessOrEqual,
						Value:    endStr,
					},
					{
						ID:       20013,
						Operator: domain.Equals,
						AttrDef: domain.AttributeDefinition{
							ID: 20003,
						},
						Value: "办公室",
					},
					{
						ID:       20014,
						Operator: domain.Equals,
						AttrDef: domain.AttributeDefinition{
							ID: 20004,
						},
						Value: "order.com",
					},
					{
						ID:       20015,
						Operator: domain.AND,
						LeftRule: &domain.PolicyRule{
							ID: 20010,
						},
						RightRule: &domain.PolicyRule{
							ID: 20011,
						},
					},
					{
						ID:       20016,
						Operator: domain.AND,
						LeftRule: &domain.PolicyRule{
							ID: 20015,
						},
						RightRule: &domain.PolicyRule{
							ID: 20012,
						},
					},
					{
						ID:       20017,
						Operator: domain.AND,
						LeftRule: &domain.PolicyRule{
							ID: 20016,
						},
						RightRule: &domain.PolicyRule{
							ID: 20013,
						},
					},
					{
						ID:       20018,
						Operator: domain.AND,
						LeftRule: &domain.PolicyRule{
							ID: 20017,
						},
						RightRule: &domain.PolicyRule{
							ID: 20014,
						},
					},
				}

				for idx := range rules {
					_, err = s.policyRepo.SaveRule(t.Context(), bizId, id, rules[idx])
					require.NoError(s.T(), err)
				}
				// 关联权限和策略
				err = s.policyRepo.SavePermissionPolicy(t.Context(), bizId, id, per.ID, domain.EffectAllow)
				require.NoError(t, err)
			},
			permission: domain.Permission{
				Resource: domain.Resource{
					Type: "code",
					Key:  "order.com",
				},
				Action: "push",
			},
			attrs: domain.Attributes{
				Environment: map[string]string{
					"time":     fmt.Sprintf("%d", time.Now().UnixMilli()),
					"location": "办公室",
				},
			},
			wantVal: true,
		},
	}
	for idx := range testcase {
		tc := testcase[idx]
		t.Run(tc.name, func(t *testing.T) {
			defer s.clearBizVal(bizId)
			tc.before(t)
			pass, err := s.hybridSvc.Check(t.Context(), bizId, tc.uid, tc.permission.Resource, []string{tc.permission.Action}, tc.attrs)
			require.NoError(t, err)
			assert.Equal(t, tc.wantVal, pass)
		})
	}

}

func (s *HybridPermissionSuite) clearBizVal(bizId int64) {
	t := s.T()
	t.Helper()
	s.db.WithContext(s.T().Context()).Where("biz_id = ?", bizId).Delete(&dao.Policy{})
	s.db.WithContext(s.T().Context()).Where("biz_id = ?", bizId).Delete(&dao.PolicyRule{})
	s.db.WithContext(s.T().Context()).Where("biz_id = ?", bizId).Delete(&dao.SubjectAttributeValue{})
	s.db.WithContext(s.T().Context()).Where("biz_id = ?", bizId).Delete(&dao.ResourceAttributeValue{})
	s.db.WithContext(s.T().Context()).Where("biz_id = ?", bizId).Delete(&dao.EnvironmentAttributeValue{})
	s.db.WithContext(s.T().Context()).Where("biz_id = ?", bizId).Delete(&dao.PermissionPolicy{})
	s.db.WithContext(s.T().Context()).Where("biz_id = ?", bizId).Delete(&dao.Permission{})
	s.db.WithContext(s.T().Context()).Where("biz_id = ?", bizId).Delete(&dao.AttributeDefinition{})
	s.db.WithContext(s.T().Context()).Where("biz_id = ?", bizId).Delete(&dao.Resource{})
	s.db.WithContext(s.T().Context()).Where("biz_id = ?", bizId).Delete(&dao.UserRole{})
}

func TestHybridPermissionSuite(t *testing.T) {
	suite.Run(t, new(HybridPermissionSuite))
}
