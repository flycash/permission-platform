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
		wire.Bind(new(repository.RoleInclusionRepository), new(*repository.RoleInclusionDefaultRepository)),
		dao.NewRolePermissionDAO,
		repository.NewRolePermissionDefaultRepository,
		wire.Bind(new(repository.RolePermissionRepository), new(*repository.RolePermissionDefaultRepository)),
		dao.NewUserRoleDAO,
		repository.NewUserRoleDefaultRepository,
		wire.Bind(new(repository.UserRoleRepository), new(*repository.UserRoleDefaultRepository)),
		dao.NewUserPermissionDAO,
		repository.NewUserPermissionDefaultRepository,
		wire.Bind(new(repository.UserPermissionRepository), new(*repository.UserPermissionDefaultRepository)),

		wire.Struct(new(Service), "*"),
	)
	return nil
}
