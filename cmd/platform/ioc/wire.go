//go:build wireinject

package ioc

import (
	rbacgrpc "gitee.com/flycash/permission-platform/internal/api/grpc/rbac"
	auditevt "gitee.com/flycash/permission-platform/internal/event/audit"
	permissionevt "gitee.com/flycash/permission-platform/internal/event/permission"
	"gitee.com/flycash/permission-platform/internal/ioc"
	"gitee.com/flycash/permission-platform/internal/repository"
	"gitee.com/flycash/permission-platform/internal/repository/cache"
	"gitee.com/flycash/permission-platform/internal/repository/dao"
	auditdao "gitee.com/flycash/permission-platform/internal/repository/dao/audit"
	rbacsvc "gitee.com/flycash/permission-platform/internal/service/rbac"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
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
		ioc.InitKafkaProducer,
		// local.NewLocalCache,
		// redis.NewCache,
	)
	rbacSvcSet = wire.NewSet(
		rbacsvc.NewService,
		rbacsvc.NewPermissionService,

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
		repository.NewRoleInclusionReloadCacheRepository,
		wire.Bind(new(repository.RoleInclusionRepository), new(*repository.RoleInclusionReloadCacheRepository)),

		dao.NewRolePermissionDAO,
		repository.NewRolePermissionDefaultRepository,
		repository.NewRolePermissionReloadCacheRepository,
		wire.Bind(new(repository.RolePermissionRepository), new(*repository.RolePermissionReloadCacheRepository)),

		dao.NewUserRoleDAO,
		repository.NewUserRoleDefaultRepository,
		repository.NewUserRoleReloadCacheRepository,
		wire.Bind(new(repository.UserRoleRepository), new(*repository.UserRoleReloadCacheRepository)),

		dao.NewUserPermissionDAO,
		repository.NewUserPermissionDefaultRepository,

		cache.NewUserPermissionCache,
		repository.NewUserPermissionCachedRepository,
		wire.Bind(new(repository.UserPermissionRepository), new(*repository.UserPermissionCachedRepository)),
		wire.Bind(new(repository.UserPermissionCacheReloader), new(*repository.UserPermissionCachedRepository)),

		auditdao.NewUserRoleLogDAO,
		auditdao.NewOperationLogDAO,

		initUserRoleBinlogEventConsumer,
		initUserPermissionEventProducer,
	)
)

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

func initUserPermissionEventProducer(producer *kafka.Producer) permissionevt.UserPermissionEventProducer {
	type Config struct {
		Topic string `yaml:"topic"`
	}
	var cfg Config
	err := econf.UnmarshalKey("userPermissionEvent", &cfg)
	if err != nil {
		panic(err)
	}
	p, err := permissionevt.NewUserPermissionEventProducer(producer, cfg.Topic)
	if err != nil {
		panic(err)
	}
	return p
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
