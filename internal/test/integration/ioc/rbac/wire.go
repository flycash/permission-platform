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
		repository.NewRBACRepositoryOld,

		dao.NewBusinessConfigDAO,
		dao.NewResourceDAO,
		dao.NewPermissionDAO,
		dao.NewRoleDAO,
		dao.NewRoleInclusionDAO,
		dao.NewRolePermissionDAO,
		dao.NewUserRoleDAO,
		dao.NewUserPermissionDAO,

		wire.Struct(new(Service), "*"),
	)
	return nil
}
