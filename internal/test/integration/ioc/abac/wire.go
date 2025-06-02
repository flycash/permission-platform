//go:build wireinject

package abac

import (
	"gitee.com/flycash/permission-platform/internal/repository"
	"gitee.com/flycash/permission-platform/internal/repository/cache/local"
	"gitee.com/flycash/permission-platform/internal/repository/cache/redisx"
	"gitee.com/flycash/permission-platform/internal/repository/dao"
	abacsvc "gitee.com/flycash/permission-platform/internal/service/abac"
	"gitee.com/flycash/permission-platform/internal/service/abac/evaluator"
	"github.com/ecodeclub/ecache/memory/lru"
	"github.com/ego-component/egorm"
	"github.com/google/wire"
	"github.com/redis/go-redis/v9"
)

type Service struct {
	PermissionSvc  abacsvc.PermissionSvc
	ValRepo        repository.AttributeValueRepository
	DefinitionRepo repository.AttributeDefinitionRepository
	PermissionRepo repository.PermissionRepository
	ResourceRepo   repository.ResourceRepository
	PolicyRepo     repository.PolicyRepo
}

func Init(db *egorm.Component, redisClient *redis.Client, lruCache *lru.Cache) *Service {
	wire.Build(
		dao.NewSubjectAttributeValueDAO,
		dao.NewResourceAttributeValueDAO,
		dao.NewEnvironmentAttributeDAO,
		dao.NewPolicyDAO,
		dao.NewAttributeDefinitionDAO,
		dao.NewResourceDAO,
		dao.NewPermissionDAO,
		repository.NewPermissionRepository,
		repository.NewResourceRepository,
		initAbacDefinitionLocalCache,
		initAbacPolicyRepo,
		initAbacAttribueValRepo,
		evaluator.NewSelector,
		abacsvc.NewPolicyExecutor,
		abacsvc.NewPermissionSvc,
		wire.Struct(new(Service), "*"),
	)
	return nil
}

func initAbacDefinitionLocalCache(attrdao dao.AttributeDefinitionDAO, client *redis.Client, lruCache *lru.Cache) repository.AttributeDefinitionRepository {
	localCache := local.NewAbacDefLocalCache(lruCache, client)
	redisCache := redisx.NewAbacDefCache(client)
	return repository.NewAttributeDefinitionRepository(attrdao, localCache, redisCache)
}

func initAbacPolicyRepo(attrdao dao.PolicyDAO, client *redis.Client, lruCache *lru.Cache) repository.PolicyRepo {
	localCache := local.NewAbacPolicy(lruCache)
	redisCache := redisx.NewAbacPolicy(client)
	return repository.NewPolicyRepository(attrdao, localCache, redisCache)
}


func initAbacAttribueValRepo(envDao dao.EnvironmentAttributeDAO,
	resourceDao dao.ResourceAttributeValueDAO,
	subjectDao dao.SubjectAttributeValueDAO,
	definitionDao dao.AttributeDefinitionDAO,
	client *redis.Client, lruCache *lru.Cache)repository.AttributeValueRepository {
	localCache := local.NewAbacAttributeValCache(lruCache)
	redisCache := redisx.NewAbacAttributeValCache(client)
	return repository.NewAttributeValueRepository(envDao,resourceDao,subjectDao,definitionDao,redisCache,localCache)

}