//go:build e2e

package integration

import (
	"fmt"
	"testing"
	"time"

	"gitee.com/flycash/permission-platform/internal/repository/dao"
	"github.com/ego-component/egorm"
	"github.com/stretchr/testify/assert"

	"gitee.com/flycash/permission-platform/internal/domain"
	"gitee.com/flycash/permission-platform/internal/repository"
	abacsvc "gitee.com/flycash/permission-platform/internal/service/abac"
	"gitee.com/flycash/permission-platform/internal/test/integration/ioc/abac"
	testioc "gitee.com/flycash/permission-platform/internal/test/ioc"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type PermissionSuite struct {
	suite.Suite
	permissionSvc  abacsvc.PermissionSvc
	valRepo        repository.AttributeValueRepository
	definitionRepo repository.AttributeDefinitionRepository
	permissionRepo repository.PermissionRepository
	resourceRepo   repository.ResourceRepository
	policyRepo     repository.PolicyRepo
	db             *egorm.Component
}

func (s *PermissionSuite) SetupSuite() {
	db := testioc.InitDBAndTables()
	svc := abac.Init(db)
	s.definitionRepo = svc.DefinitionRepo
	s.permissionSvc = svc.PermissionSvc
	s.valRepo = svc.ValRepo
	s.policyRepo = svc.PolicyRepo
	s.permissionRepo = svc.PermissionRepo
	s.resourceRepo = svc.ResourceRepo
	s.db = db
}

func (s *PermissionSuite) TestPermission() {
	bizId := int64(10000)

	testcase := []struct {
		name       string
		before     func(t *testing.T)
		permission domain.Permission
		attrs      domain.PermissionRequest
		uid        int64
		wantVal    bool
	}{
		{
			name: "policy是permit",
			uid:  22,
			permission: domain.Permission{
				Resource: domain.Resource{
					Key:  "/order/tab",
					Type: "table",
				},
				Action: "read",
			},
			attrs: domain.PermissionRequest{
				EnvironmentAttrValues: map[string]string{
					"time": fmt.Sprintf("%d", time.Now().UnixMilli()),
				},
			},
			before: func(t *testing.T) {
				s.setupDefinition()
				// 创建主体属性值
				vals := []domain.AttributeValue{
					{
						Value: "22",
						Definition: domain.AttributeDefinition{
							ID: 10001,
						},
					},
					{
						Value: "man",
						Definition: domain.AttributeDefinition{
							ID: 10002,
						},
					},
					{
						Value: "manager",
						Definition: domain.AttributeDefinition{
							ID: 10003,
						},
					},
				}
				for idx := range vals {
					_, err := s.valRepo.SaveSubjectValue(t.Context(), bizId, 22, vals[idx])
					require.NoError(s.T(), err)
				}
				// 创建资源属性值
				_, err := s.valRepo.SaveResourceValue(t.Context(), bizId, 33, domain.AttributeValue{
					Definition: domain.AttributeDefinition{
						ID: 10004,
					},
					Value: "order",
				})
				require.NoError(s.T(), err)
				// 创建权限
				permission := domain.Permission{
					BizID:       bizId,
					Name:        "读取订单权限",
					Description: "允许读取订单信息",
					Resource: domain.Resource{
						ID:   10001,
						Type: "order",
						Key:  "/order/tab",
					},
					Action:   "read",
					Metadata: "[1]",
				}
				per, err := s.permissionRepo.Create(t.Context(), permission)
				require.NoError(t, err)
				// 创建资源
				res := domain.Resource{
					ID:          33,
					BizID:       bizId,
					Type:        "table",
					Name:        "order_table",
					Description: "desc",
					Key:         "/order/tab",
					Metadata:    "[1]",
				}
				_, err = s.resourceRepo.Create(t.Context(), res)
				require.NoError(t, err)

				// 创建策略
				policy := domain.Policy{
					BizID:       bizId,
					Name:        "订单读取策略",
					Description: "允许用户读取订单信息",
					Status:      domain.PolicyStatusActive,
				}
				id, err := s.policyRepo.Save(t.Context(), policy)
				require.NoError(t, err)
				rules := []domain.PolicyRule{
					{
						ID: 10010,
						AttrDef: domain.AttributeDefinition{
							ID: 10001,
						},
						Operator: domain.GreaterOrEqual,
						Value:    "20",
					},
					{
						ID: 10011,
						AttrDef: domain.AttributeDefinition{
							ID: 10001,
						},
						Operator: domain.LessOrEqual,
						Value:    "30",
					},
					{
						ID: 10012,
						LeftRule: &domain.PolicyRule{
							ID: 10010,
						},
						RightRule: &domain.PolicyRule{
							ID: 10011,
						},
						Operator: domain.AND,
					},
					{
						ID:       10013,
						Operator: domain.IN,
						AttrDef: domain.AttributeDefinition{
							ID: 10003,
						},
						Value: "[\"admin\",\"manager\"]",
					},
					{
						ID:       10014,
						Operator: domain.AND,
						LeftRule: &domain.PolicyRule{
							ID: 10013,
						},
						RightRule: &domain.PolicyRule{
							ID: 10012,
						},
					},
					{
						ID:       10015,
						Operator: domain.GreaterOrEqual,
						AttrDef: domain.AttributeDefinition{
							ID: 10005,
						},
						Value: "@time(1)",
					},
					{
						ID:       10016,
						Operator: domain.AND,
						LeftRule: &domain.PolicyRule{
							ID: 10014,
						},
						RightRule: &domain.PolicyRule{
							ID: 10015,
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
			wantVal: true,
		},
		{
			name: "policy是有一个是deny有一个permit",
			uid:  23,
			permission: domain.Permission{
				Resource: domain.Resource{
					Key:  "/order/tab",
					Type: "order",
				},
				Action: "read",
			},
			attrs: domain.PermissionRequest{
				EnvironmentAttrValues: map[string]string{
					"time": fmt.Sprintf("%d", time.Now().UnixMilli()),
				},
			},
			before: func(t *testing.T) {
				s.setupDefinition()
				// 创建主体属性值，满足两个policy
				vals := []domain.AttributeValue{
					{
						Value: "25",
						Definition: domain.AttributeDefinition{
							ID: 10001, // age
						},
					},
					{
						Value: "man",
						Definition: domain.AttributeDefinition{
							ID: 10002, // gender
						},
					},
					{
						Value: "manager",
						Definition: domain.AttributeDefinition{
							ID: 10003, // position
						},
					},
				}
				for idx := range vals {
					_, err := s.valRepo.SaveSubjectValue(t.Context(), bizId, 23, vals[idx])
					require.NoError(s.T(), err)
				}
				// 创建资源属性值
				_, err := s.valRepo.SaveResourceValue(t.Context(), bizId, 33, domain.AttributeValue{
					Definition: domain.AttributeDefinition{
						ID: 10004,
					},
					Value: "order",
				})
				require.NoError(s.T(), err)
				// 创建权限
				permission := domain.Permission{
					BizID:       bizId,
					Name:        "读取订单权限",
					Description: "允许读取订单信息",
					Resource: domain.Resource{
						ID:   10001,
						Key:  "/order/tab",
						Type: "order",
					},
					Metadata: "[1]",

					Action: "read",
				}
				per, err := s.permissionRepo.Create(t.Context(), permission)
				require.NoError(t, err)
				// 创建资源
				res := domain.Resource{
					ID:          10020,
					BizID:       bizId,
					Type:        "order",
					Name:        "order_table",
					Description: "desc",
					Key:         "/order/tab",
					Metadata:    "[1]",
				}
				_, err = s.resourceRepo.Create(t.Context(), res)
				require.NoError(t, err)

				// 创建permit策略
				permitPolicy := domain.Policy{
					BizID:       bizId,
					Name:        "permit策略",
					Description: "允许用户读取订单信息",
					Status:      domain.PolicyStatusActive,
				}
				permitId, err := s.policyRepo.Save(t.Context(), permitPolicy)
				require.NoError(t, err)
				permitRules := []domain.PolicyRule{
					{
						ID: 20010,
						AttrDef: domain.AttributeDefinition{
							ID: 10001, // age
						},
						Operator: domain.GreaterOrEqual,
						Value:    "20",
					},
					{
						ID: 20011,
						AttrDef: domain.AttributeDefinition{
							ID: 10001,
						},
						Operator: domain.LessOrEqual,
						Value:    "30",
					},
					{
						ID:        20012,
						LeftRule:  &domain.PolicyRule{ID: 20010},
						RightRule: &domain.PolicyRule{ID: 20011},
						Operator:  domain.AND,
					},
					{
						ID:       20013,
						Operator: domain.IN,
						AttrDef: domain.AttributeDefinition{
							ID: 10003, // position
						},
						Value: "[\"admin\",\"manager\"]",
					},
					{
						ID:        20014,
						Operator:  domain.AND,
						LeftRule:  &domain.PolicyRule{ID: 20013},
						RightRule: &domain.PolicyRule{ID: 20012},
					},
				}
				for idx := range permitRules {
					_, err = s.policyRepo.SaveRule(t.Context(), bizId, permitId, permitRules[idx])
					require.NoError(s.T(), err)
				}
				// 关联permit策略
				err = s.policyRepo.SavePermissionPolicy(t.Context(), bizId, permitId, per.ID, domain.EffectAllow)
				require.NoError(t, err)

				// 创建deny策略，规则与permit类似但ID不同
				denyPolicy := domain.Policy{
					BizID:       bizId,
					Name:        "deny策略",
					Description: "拒绝用户读取订单信息",
					Status:      domain.PolicyStatusActive,
				}
				denyId, err := s.policyRepo.Save(t.Context(), denyPolicy)
				require.NoError(t, err)
				denyRules := []domain.PolicyRule{
					{
						ID: 30010,
						AttrDef: domain.AttributeDefinition{
							ID: 10001, // age
						},
						Operator: domain.GreaterOrEqual,
						Value:    "20",
					},
					{
						ID: 30011,
						AttrDef: domain.AttributeDefinition{
							ID: 10001,
						},
						Operator: domain.LessOrEqual,
						Value:    "30",
					},
					{
						ID:        30012,
						LeftRule:  &domain.PolicyRule{ID: 30010},
						RightRule: &domain.PolicyRule{ID: 30011},
						Operator:  domain.AND,
					},
					{
						ID:       30013,
						Operator: domain.IN,
						AttrDef: domain.AttributeDefinition{
							ID: 10003, // position
						},
						Value: "[\"admin\",\"manager\"]",
					},
					{
						ID:        30014,
						Operator:  domain.AND,
						LeftRule:  &domain.PolicyRule{ID: 30013},
						RightRule: &domain.PolicyRule{ID: 30012},
					},
				}
				for idx := range denyRules {
					_, err = s.policyRepo.SaveRule(t.Context(), bizId, denyId, denyRules[idx])
					require.NoError(s.T(), err)
				}
				// 关联deny策略
				err = s.policyRepo.SavePermissionPolicy(t.Context(), bizId, denyId, per.ID, domain.EffectDeny)
				require.NoError(t, err)
			},
			wantVal: false,
		},
	}
	for idx := range testcase {
		tc := testcase[idx]
		s.T().Run(tc.name, func(t *testing.T) {
			defer s.clearBizVal(bizId)
			tc.before(t)
			pass, err := s.permissionSvc.Check(t.Context(), bizId, tc.uid, tc.permission, tc.attrs)
			require.NoError(t, err)
			assert.Equal(t, tc.wantVal, pass)
		})
	}
}

func (s *PermissionSuite) setupDefinition() {
	// 初始化属性定义
	subjectAttrDef1 := domain.AttributeDefinition{
		ID:          10001,
		Name:        "age",
		DataType:    domain.DataTypeNumber,
		Description: "年龄",
		EntityType:  domain.EntityTypeSubject,
	}
	subjectAttrDef2 := domain.AttributeDefinition{
		ID:          10002,
		Name:        "gender",
		DataType:    domain.DataTypeString,
		Description: "性别",
		EntityType:  domain.EntityTypeSubject,
	}
	subjectAttrDef3 := domain.AttributeDefinition{
		ID:          10003,
		Name:        "position",
		DataType:    domain.DataTypeString,
		Description: "职位",
		EntityType:  domain.EntityTypeSubject,
	}
	resourceDef := domain.AttributeDefinition{
		ID:          10004,
		Name:        "table",
		DataType:    domain.DataTypeString,
		Description: "表",
		EntityType:  domain.EntityTypeResource,
	}
	envDef := domain.AttributeDefinition{
		ID:          10005,
		Name:        "time",
		DataType:    domain.DataTypeDatetime,
		Description: "访问时间",
		EntityType:  domain.EntityTypeEnvironment,
	}
	defs := []domain.AttributeDefinition{
		subjectAttrDef1,
		subjectAttrDef2,
		subjectAttrDef3,
		resourceDef,
		envDef,
	}
	for idx := range defs {
		_, err := s.definitionRepo.Save(s.T().Context(), 10000, defs[idx])
		require.NoError(s.T(), err)
	}
}

func (s *PermissionSuite) clearBizVal(bizId int64) {
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
}

func TestABACPermissionSuite(t *testing.T) {
	suite.Run(t, new(PermissionSuite))
}
