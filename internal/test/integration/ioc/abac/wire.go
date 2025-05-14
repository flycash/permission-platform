//go:build wireinject

package abac

import (
	"gitee.com/flycash/permission-platform/internal/pkg/checker"
	"gitee.com/flycash/permission-platform/internal/repository"
	"gitee.com/flycash/permission-platform/internal/repository/dao"
	abacsvc "gitee.com/flycash/permission-platform/internal/service/abac"
	"github.com/ego-component/egorm"
	"github.com/google/wire"
)

type Service struct {
	PermissionSvc  abacsvc.PermissionSvc
	ValRepo        repository.AttributeValueRepository
	DefinitionRepo repository.AttributeDefinitionRepository
	PermissionRepo repository.RBACRepository
	PolicyRepo     repository.PolicyRepo
}

func Init(db *egorm.Component) *Service {
	wire.Build(
		dao.NewSubjectAttributeValueDAO,
		dao.NewResourceAttributeValueDAO,
		dao.NewEnvironmentAttributeDAO,
		dao.NewPolicyDAO,
		dao.NewAttributeDefinitionDAO,
		dao.NewResourceDAO,
		dao.NewPermissionDAO,
		dao.NewRoleDAO,
		dao.NewRoleInclusionDAO,
		dao.NewRolePermissionDAO,
		dao.NewUserRoleDAO,
		dao.NewUserPermissionDAO,
		dao.NewBusinessConfigDAO,
		repository.NewPolicyRepository,
		repository.NewAttributeDefinitionRepository,
		repository.NewAttributeValueRepository,
		repository.NewRBACRepository,
		checker.NewCheckerBuilder,
		abacsvc.NewRuleParser,
		abacsvc.NewPermissionSvc,
		wire.Struct(new(Service), "*"),
	)
	return nil
}
