//go:build e2e

package integration

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"gitee.com/flycash/permission-platform/internal/repository/cache/redisx"

	"github.com/ecodeclub/ecache/memory/lru"
	"github.com/ego-component/eetcd"
	"github.com/ego-component/egorm"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"gitee.com/flycash/permission-platform/internal/domain"
	"gitee.com/flycash/permission-platform/internal/repository"
	"gitee.com/flycash/permission-platform/internal/repository/cache"
	"gitee.com/flycash/permission-platform/internal/repository/cache/local"
	abacsvc "gitee.com/flycash/permission-platform/internal/service/abac"
	"gitee.com/flycash/permission-platform/internal/test/integration/ioc/abac"
	testioc "gitee.com/flycash/permission-platform/internal/test/ioc"
)

type AttributeValueTaskSuite struct {
	suite.Suite
	task        *abacsvc.AttributeValueTask
	etcdClient  *eetcd.Component
	valRepo     repository.AttributeValueRepository
	localCache  cache.ABACAttributeValCache
	redisCache  cache.ABACAttributeValCache
	db          *egorm.Component
	redisClient *redis.Client
	lruCache    *lru.Cache
}

func (s *AttributeValueTaskSuite) SetupSuite() {
	s.db = testioc.InitDBAndTables()
	s.redisClient = testioc.InitRedisClient()
	s.lruCache = lru.NewCache(10000)
	s.etcdClient = testioc.InitEtcdClient()

	svc := abac.Init(s.db, s.redisClient, s.lruCache)
	s.valRepo = svc.ValRepo

	// 创建本地缓存
	s.localCache = local.NewAbacAttributeValCache(s.lruCache)
	s.redisCache = redisx.NewAbacAttributeValCache(s.redisClient)

	s.task = abacsvc.NewAttributeValueTask(s.valRepo, s.localCache, s.redisCache, "resource", "subject", s.etcdClient)
}

func (s *AttributeValueTaskSuite) TestStartResLoop() {
	ctx := context.Background()
	bizID := int64(101)
	resourceID := int64(1001)

	// 创建测试数据
	obj := domain.ABACObject{
		BizID: bizID,
		ID:    resourceID,
	}
	value, err := json.Marshal(obj)
	require.NoError(s.T(), err)

	// 启动资源循环
	go func() {
		err := s.task.StartResLoop(ctx)
		require.NoError(s.T(), err)
	}()

	// 等待循环启动
	time.Sleep(time.Second)

	// 写入etcd
	_, err = s.etcdClient.Put(ctx, "resource", string(value))
	require.NoError(s.T(), err)

	// 等待处理
	time.Sleep(time.Second)

	// 验证缓存
	redisKey := "abac:attr:resource:101:1001"
	redisVal, err := s.redisClient.Get(ctx, redisKey).Result()
	require.NoError(s.T(), err)
	require.NotEmpty(s.T(), redisVal)

	localVal := s.lruCache.Get(ctx, redisKey)
	require.False(s.T(), localVal.KeyNotFound())
	require.NotNil(s.T(), localVal)
}

func (s *AttributeValueTaskSuite) TestStartSubjectLoop() {
	ctx := context.Background()
	bizID := int64(102)
	subjectID := int64(2001)

	// 创建测试数据
	obj := domain.ABACObject{
		BizID: bizID,
		ID:    subjectID,
	}
	value, err := json.Marshal(obj)
	require.NoError(s.T(), err)

	// 启动主体循环
	go func() {
		err := s.task.StartSubjectLoop(ctx)
		require.NoError(s.T(), err)
	}()

	// 等待循环启动
	time.Sleep(time.Second)

	// 写入etcd
	_, err = s.etcdClient.Put(ctx, "subject", string(value))
	require.NoError(s.T(), err)

	// 等待处理
	time.Sleep(time.Second)

	// 验证缓存
	redisKey := "abac:attr:subject:102:2001"
	redisVal, err := s.redisClient.Get(ctx, redisKey).Result()
	require.NoError(s.T(), err)
	require.NotEmpty(s.T(), redisVal)

	localVal := s.lruCache.Get(ctx, redisKey)
	require.False(s.T(), localVal.KeyNotFound())
	require.NotNil(s.T(), localVal)
}

func TestAttributeValueTaskSuite(t *testing.T) {
	suite.Run(t, new(AttributeValueTaskSuite))
}
