//go:build wireinject

package rbac

import (
	testioc "gitee.com/flycash/permission-platform/internal/test/ioc"
	"github.com/google/wire"

	"gitee.com/flycash/permission-platform/internal/repository"
	"gitee.com/flycash/permission-platform/internal/repository/dao"
	rbacsvc "gitee.com/flycash/permission-platform/internal/service/rbac"
)

type Service struct {
	Svc                rbacsvc.Service
	PermissionSvc      rbacsvc.PermissionService
	BusinessConfigRepo repository.BusinessConfigRepository
	ResourceRepo       repository.ResourceRepository
	PermissionRepo     repository.PermissionRepository
	RoleRepo           repository.RoleRepository
	RolePermissionRepo repository.RolePermissionRepository
	UserRoleRepo       repository.UserRoleRepository
}

func Init() *Service {
	wire.Build(
		testioc.BaseSet,

		rbacsvc.NewPermissionService,

		rbacsvc.NewService,

		dao.NewBusinessConfigDAO,
		repository.NewBusinessConfigRepository,

		dao.NewResourceDAO,
		repository.NewResourceRepository,
		dao.NewPermissionDAO,
		repository.NewPermissionRepository,
		dao.NewRoleDAO,
		repository.NewRoleRepository,
		dao.NewRoleInclusionDAO,
		repository.NewRoleInclusionDefaultRepository,
		dao.NewRolePermissionDAO,
		repository.NewRolePermissionDefaultRepository,
		dao.NewUserRoleDAO,
		repository.NewUserRoleDefaultRepository,
		dao.NewUserPermissionDAO,
		repository.NewUserPermissionDefaultRepository,
		convertRepository,

		wire.Struct(new(Service), "*"),
	)
	return nil
}

func convertRepository(repo *repository.UserPermissionDefaultRepository) repository.UserPermissionRepository {
	return repo
}
