//go:build wireinject

package ioc

import (
	rbacgrpc "gitee.com/flycash/permission-platform/internal/api/grpc/rbac"
	auditevt "gitee.com/flycash/permission-platform/internal/event/audit"
	"gitee.com/flycash/permission-platform/internal/ioc"
	"gitee.com/flycash/permission-platform/internal/repository"
	"gitee.com/flycash/permission-platform/internal/repository/dao"
	auditdao "gitee.com/flycash/permission-platform/internal/repository/dao/audit"
	rbacsvc "gitee.com/flycash/permission-platform/internal/service/rbac"
	"github.com/google/wire"
	"github.com/gotomicro/ego/core/econf"
)

var (
	baseSet = wire.NewSet(
		ioc.InitDB,
		// ioc.InitDistributedLock,
		ioc.InitEtcdClient,
		ioc.InitIDGenerator,
		ioc.InitRedisClient,
		ioc.InitLocalCache,
		ioc.InitRedisCmd,
		ioc.InitJWTToken,
		ioc.InitMultipleLevelCache,
		ioc.InitCacheKeyFunc,
		// local.NewLocalCache,
		// redis.NewCache,
	)
	rbacSvcSet = wire.NewSet(
		rbacsvc.NewService,
		rbacsvc.NewPermissionService,
		repository.NewDefaultRBACRepository,
		repository.NewCachedRBACRepository,
		convertRepository,

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

		auditdao.NewUserRoleLogDAO,
		auditdao.NewOperationLogDAO,

		initUserRoleBinlogEventConsumer,
	)
)

func convertRepository(repo *repository.CachedRBACRepository) repository.RBACRepository {
	return repo
}

func initUserRoleBinlogEventConsumer(dao auditdao.UserRoleLogDAO) *auditevt.UserRoleBinlogEventConsumer {
	type Consumer struct {
		GroupID string `yaml:"groupId"`
	}
	type Config struct {
		Topic    string   `yaml:"topic"`
		Consumer Consumer `yaml:"consumer"`
	}
	var cfg Config
	err := econf.UnmarshalKey("userRoleBinlogEvent", &cfg)
	if err != nil {
		panic(err)
	}
	eventConsumer, err := auditevt.NewUserRoleBinlogEventConsumer(
		ioc.InitKafkaConsumer(cfg.Consumer.GroupID),
		dao,
		cfg.Topic)
	if err != nil {
		panic(err)
	}
	return eventConsumer
}

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
		ioc.InitTasks,
		wire.Struct(new(ioc.App), "*"),
	)

	return new(ioc.App)
}
