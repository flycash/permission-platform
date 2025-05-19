//go:build wireinject

package ioc

import (
	rbacsvc "gitee.com/flycash/permission-platform/internal/service/rbac"

	rbacgrpc "gitee.com/flycash/permission-platform/internal/api/grpc/rbac"
	"gitee.com/flycash/permission-platform/internal/ioc"
	"gitee.com/flycash/permission-platform/internal/repository"
	"gitee.com/flycash/permission-platform/internal/repository/dao"
	"github.com/google/wire"
)

var (
	baseSet = wire.NewSet(
		ioc.InitDB,
		ioc.InitDistributedLock,
		ioc.InitEtcdClient,
		ioc.InitIDGenerator,
		ioc.InitRedisClient,
		ioc.InitGoCache,
		ioc.InitRedisCmd,
		ioc.InitJWTToken,
		// local.NewLocalCache,
		// redis.NewCache,
	)
	rbacSvcSet = wire.NewSet(
		rbacsvc.NewService,
		rbacsvc.NewPermissionService,
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
	)
)

func InitApp() *ioc.App {
	wire.Build(
		// 基础设施
		baseSet,

		// RBAC 服务
		rbacSvcSet,

		// GRPC服务器
		rbacgrpc.NewServer,
		rbacgrpc.NewPermissionServiceServer,
		ioc.InitGRPC,
		wire.Struct(new(ioc.App), "*"),
	)

	return new(ioc.App)
}
