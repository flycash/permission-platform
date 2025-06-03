//go:build e2e

package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/ecodeclub/ecache/memory/lru"
	"github.com/ego-component/eetcd"
	"github.com/ego-component/egorm"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/suite"

	"gitee.com/flycash/permission-platform/internal/domain"
	"gitee.com/flycash/permission-platform/internal/repository"
	"gitee.com/flycash/permission-platform/internal/repository/cache"
	"gitee.com/flycash/permission-platform/internal/repository/cache/local"
	"gitee.com/flycash/permission-platform/internal/repository/dao"
	"gitee.com/flycash/permission-platform/internal/service/abac"
	abacioc "gitee.com/flycash/permission-platform/internal/test/integration/ioc/abac"
	testioc "gitee.com/flycash/permission-platform/internal/test/ioc"
)

type PolicyCronSuite struct {
	suite.Suite
	policyCron  *abac.PolicyCron
	etcdClient  *eetcd.Component
	policyRepo  repository.PolicyRepo
	policyCache cache.ABACPolicyCache
	db          *egorm.Component
	redisClient redis.Cmdable
	lruCache    *lru.Cache
}

func (s *PolicyCronSuite) SetupSuite() {
	db := testioc.InitDBAndTables()
	redisClient := testioc.InitRedisClient()
	lruCache := lru.NewCache(10000)
	etcdClient := testioc.InitEtcdClient()

	s.db = db
	s.redisClient = redisClient
	s.lruCache = lruCache
	s.etcdClient = etcdClient

	svc := abacioc.Init(db, redisClient, lruCache)
	s.policyRepo = svc.PolicyRepo

	// 创建本地缓存
	localCache := local.NewAbacPolicy(lruCache)

	s.policyCron = abac.NewPolicyCron(s.etcdClient, s.policyRepo, localCache)
}

func (s *PolicyCronSuite) clearBizVal(bizId int64) {
	t := s.T()
	t.Helper()
	s.db.WithContext(t.Context()).Where("biz_id = ?", bizId).Delete(&dao.Policy{})
	s.db.WithContext(t.Context()).Where("biz_id = ?", bizId).Delete(&dao.PolicyRule{})
	s.redisClient.Del(s.T().Context(), fmt.Sprintf("abac:policy:%d", bizId))
	s.lruCache.Delete(s.T().Context(), fmt.Sprintf("abac:policy:%d", bizId))
}

func (s *PolicyCronSuite) TestPolicyCron_Run() {
	ctx := context.Background()
	bizID := int64(26)
	defer s.clearBizVal(bizID)

	// 将业务ID写入etcd的hotkey中
	bizIDs := []int64{bizID}
	value, err := json.Marshal(bizIDs)
	s.NoError(err)
	_, err = s.etcdClient.Put(ctx, "hotPolicy", string(value))
	s.NoError(err)

	// 创建测试策略
	policy := domain.Policy{
		BizID:       bizID,
		Name:        "测试策略",
		Description: "这是一个测试策略",
	}
	policyID, err := s.policyRepo.Save(ctx, policy)
	s.NoError(err)
	s.Greater(policyID, int64(0))

	// 运行cron
	err = s.policyCron.Run(ctx)
	s.NoError(err)

	// 验证本地缓存已更新
	localCacheKey := fmt.Sprintf("abac:policy:%d", bizID)
	localVal := s.lruCache.Get(ctx, localCacheKey)
	s.False(localVal.KeyNotFound())
	s.NotNil(localVal)
}

func TestPolicyCronSuite(t *testing.T) {
	suite.Run(t, new(PolicyCronSuite))
}
