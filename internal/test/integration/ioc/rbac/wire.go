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
	Svc  rbacsvc.Service
	Repo repository.RBACRepository
}

func Init() *Service {
	wire.Build(
		testioc.BaseSet,

		rbacsvc.NewService,
		repository.NewRBACRepository,

		dao.NewBusinessConfigDAO,
		repository.NewBusinessConfigRepository,
		dao.NewResourceDAO,
		repository.NewResourceRepository,
		dao.NewPermissionDAO,
		repository.NewPermissionRepository,
		dao.NewRoleDAO,
		repository.NewRoleRepository,
		dao.NewRoleInclusionDAO,
		repository.NewRoleInclusionRepository,
		dao.NewRolePermissionDAO,
		repository.NewRolePermissionRepository,
		dao.NewUserRoleDAO,
		repository.NewUserRoleRepository,
		dao.NewUserPermissionDAO,
		repository.NewUserPermissionRepository,

		wire.Struct(new(Service), "*"),
	)
	return nil
}
