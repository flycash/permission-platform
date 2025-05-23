//go:build wireinject

package ioc

import (
	"time"

	auditevt "gitee.com/flycash/permission-platform/internal/event/audit"
	auditdao "gitee.com/flycash/permission-platform/internal/repository/dao/audit"
	rbacsvc "gitee.com/flycash/permission-platform/internal/service/rbac"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/gotomicro/ego/core/econf"
	"github.com/withlin/canal-go/client"

	rbacgrpc "gitee.com/flycash/permission-platform/internal/api/grpc/rbac"
	"gitee.com/flycash/permission-platform/internal/ioc"
	"gitee.com/flycash/permission-platform/internal/repository"
	"gitee.com/flycash/permission-platform/internal/repository/dao"
	"github.com/google/wire"
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
		ioc.InitCanalConnector,
		ioc.InitKafkaProducer,

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

		initUserRoleBinlogEventProducer,
		initUserRoleBinlogEventConsumer,
	)
)

func convertRepository(repo *repository.CachedRBACRepository) repository.RBACRepository {
	return repo
}

func initUserRoleBinlogEventProducer(
	canalConn client.CanalConnector,
	kafkaProducer *kafka.Producer,
) *auditevt.CanalUserRoleBinlogEventProducer {

	type Producer struct {
		MinLoopDuration time.Duration `yaml:"minLoopDuration"`
		BatchSize       int32         `yaml:"batchSize"`
		Timeout         int64         `yaml:"timeout"`
		Units           int32         `yaml:"units"`
	}

	type Config struct {
		Topic    string   `yaml:"topic"`
		Producer Producer `yaml:"producer"`
	}
	var cfg Config
	err := econf.UnmarshalKey("userRoleBinlogEvent", &cfg)
	if err != nil {
		panic(err)
	}
	producer, err := auditevt.NewUserRoleBinlogEventProducer(kafkaProducer, cfg.Topic)
	if err != nil {
		panic(err)
	}
	eventProducer := auditevt.NewCanalUserRoleBinlogEventProducer(
		canalConn,
		producer,
		cfg.Producer.MinLoopDuration,
		cfg.Producer.BatchSize,
		cfg.Producer.Timeout,
		cfg.Producer.Units,
	)
	return eventProducer
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
