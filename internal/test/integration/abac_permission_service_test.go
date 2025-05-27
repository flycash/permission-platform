//go:build e2e

package integration

import (
	"fmt"
	"github.com/ecodeclub/ecache/memory/lru"
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

type ABACPermissionSuite struct {
	suite.Suite
	permissionSvc  abacsvc.PermissionSvc
	valRepo        repository.AttributeValueRepository
	definitionRepo repository.AttributeDefinitionRepository
	permissionRepo repository.PermissionRepository
	resourceRepo   repository.ResourceRepository
	policyRepo     repository.PolicyRepo
	db             *egorm.Component
}

func (s *ABACPermissionSuite) SetupSuite() {
	db := testioc.InitDBAndTables()
	redisClient := testioc.InitRedisClient()
	svc := abac.Init(db, redisClient, lru.NewCache(10000))
	s.definitionRepo = svc.DefinitionRepo
	s.permissionSvc = svc.PermissionSvc
	s.valRepo = svc.ValRepo
	s.policyRepo = svc.PolicyRepo
	s.permissionRepo = svc.PermissionRepo
	s.resourceRepo = svc.ResourceRepo
	s.db = db
}

func (s *ABACPermissionSuite) TestPermission() {
	bizId := int64(10000)

	testcase := []struct {
		name       string
		before     func(t *testing.T)
		permission domain.Permission
		attrs      domain.Attributes
		uid        int64
		wantVal    bool
	}{
		{
			name: "程序员在某个时间段，某个地点可以访问user.com代码仓库",
			uid:  22,
			permission: domain.Permission{
				Resource: domain.Resource{
					Key:  "user.com",
					Type: "code",
				},
				Action: "read",
			},
			attrs: domain.Attributes{
				Environment: map[string]string{
					"time":     fmt.Sprintf("%d", time.Now().UnixMilli()),
					"location": "办公室",
				},
			},
			before: func(t *testing.T) {
				s.setupDefinitionV1()
				// 创建主体属性值
				vals := []domain.AttributeValue{
					{
						Value: "程序员",
						Definition: domain.AttributeDefinition{
							ID: 10006,
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
						ID: 10009,
					},
					Value: "user.com",
				})
				require.NoError(s.T(), err)
				// 创建权限
				permission := domain.Permission{
					BizID:       bizId,
					Name:        "代码仓库权限",
					Description: "代码仓库权限",
					Resource: domain.Resource{
						Type: "code",
						Key:  "user.com",
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
					Type:        "code",
					Name:        "user.com",
					Description: "用户",
					Key:         "user.com",
					Metadata:    "[1]",
				}
				_, err = s.resourceRepo.Create(t.Context(), res)
				require.NoError(t, err)
				startTime := time.Now().Add(-time.Hour)
				startStr := fmt.Sprintf("@day(%s)", startTime.Format("15:04"))
				endTime := time.Now().Add(1 * time.Hour)
				endStr := fmt.Sprintf("@day(%s)", endTime.Format("15:04"))
				// 创建策略
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
						ID: 10010,
						AttrDef: domain.AttributeDefinition{
							ID: 10006,
						},
						Operator: domain.IN,
						Value:    "[\"程序员\",\"经理\"]",
					},
					{
						ID: 10011,
						AttrDef: domain.AttributeDefinition{
							ID: 10007,
						},
						Operator: domain.GreaterOrEqual,
						Value:    startStr,
					},
					{
						ID: 10012,
						AttrDef: domain.AttributeDefinition{
							ID: 10007,
						},
						Operator: domain.LessOrEqual,
						Value:    endStr,
					},
					{
						ID:       10013,
						Operator: domain.Equals,
						AttrDef: domain.AttributeDefinition{
							ID: 10008,
						},
						Value: "办公室",
					},
					{
						ID:       10014,
						Operator: domain.Equals,
						AttrDef: domain.AttributeDefinition{
							ID: 10009,
						},
						Value: "user.com",
					},
					{
						ID:       10015,
						Operator: domain.AND,
						LeftRule: &domain.PolicyRule{
							ID: 10010,
						},
						RightRule: &domain.PolicyRule{
							ID: 10011,
						},
					},
					{
						ID:       10016,
						Operator: domain.AND,
						LeftRule: &domain.PolicyRule{
							ID: 10015,
						},
						RightRule: &domain.PolicyRule{
							ID: 10012,
						},
					},
					{
						ID:       10017,
						Operator: domain.AND,
						LeftRule: &domain.PolicyRule{
							ID: 10016,
						},
						RightRule: &domain.PolicyRule{
							ID: 10013,
						},
					},
					{
						ID:       10018,
						Operator: domain.AND,
						LeftRule: &domain.PolicyRule{
							ID: 10017,
						},
						RightRule: &domain.PolicyRule{
							ID: 10014,
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
	}
	for idx := range testcase {
		tc := testcase[idx]
		s.T().Run(tc.name, func(t *testing.T) {
			defer s.clearBizVal(bizId)
			tc.before(t)
			pass, err := s.permissionSvc.Check(t.Context(), bizId, tc.uid, tc.permission.Resource, []string{tc.permission.Action}, tc.attrs)
			require.NoError(t, err)
			assert.Equal(t, tc.wantVal, pass)
		})
	}
}

func (s *ABACPermissionSuite) setupDefinitionV1() {
	// 初始化属性定义
	subjectAttrDef1 := domain.AttributeDefinition{
		ID:          10006,
		Name:        "role",
		DataType:    domain.DataTypeString,
		Description: "角色",
		EntityType:  domain.EntityTypeSubject,
	}
	envAttrDef2 := domain.AttributeDefinition{
		ID:          10007,
		Name:        "time",
		DataType:    domain.DataTypeDatetime,
		Description: "时间",
		EntityType:  domain.EntityTypeEnvironment,
	}
	envAttrDef3 := domain.AttributeDefinition{
		ID:          10008,
		Name:        "location",
		DataType:    domain.DataTypeString,
		Description: "地点",
		EntityType:  domain.EntityTypeEnvironment,
	}
	resourceDef := domain.AttributeDefinition{
		ID:          10009,
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
		_, err := s.definitionRepo.Save(s.T().Context(), 10000, defs[idx])
		require.NoError(s.T(), err)
	}
}

func (s *ABACPermissionSuite) clearBizVal(bizId int64) {
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
	suite.Run(t, new(ABACPermissionSuite))
}
