package ioc

import (
	"context"
	"encoding/json"
	"fmt"
	"sync/atomic"
	"time"

	"gitee.com/flycash/permission-platform/internal/domain"
	"gitee.com/flycash/permission-platform/internal/repository"
	"gitee.com/flycash/permission-platform/pkg/bitring"
	"gitee.com/flycash/permission-platform/pkg/cache"
	"github.com/ecodeclub/ecache"
	"github.com/ego-component/eetcd"
	"github.com/gotomicro/ego/core/econf"
	"github.com/redis/go-redis/v9"
	clientv3 "go.etcd.io/etcd/client/v3"
)

func InitCacheKeyFunc() func(bizID, userID int64) string {
	return func(bizID, userID int64) string {
		return fmt.Sprintf("permissions:bizID:%d:userID:%d", bizID, userID)
	}
}

func InitMultipleLevelCache(
	r redis.Cmdable,
	local ecache.Cache,
	repo *repository.UserPermissionDefaultRepository,
	etcdClient *eetcd.Component,
	cacheKeyFunc func(bizID, userID int64) string,
) cache.Cache {
	type ErrorEventConfig struct {
		BitRingSize      int     `yaml:"bitRingSize"`
		RateThreshold    float64 `yaml:"rateThreshold"`
		ConsecutiveCount int     `yaml:"consecutiveCount"`
	}

	type Config struct {
		EtcdKey                 string           `yaml:"etcdKey"`
		LocalCacheRefreshPeriod time.Duration    `yaml:"localCacheRefreshPeriod"`
		RedisPingTimeout        time.Duration    `yaml:"redisPingTimeout"`
		RedisHealthCheckPeriod  time.Duration    `yaml:"redisHealthCheckPeriod"`
		ErrorEvents             ErrorEventConfig `yaml:"errorEvents"`
	}

	var cfg Config
	err := econf.UnmarshalKey("cache.multilevel", &cfg)
	if err != nil {
		panic(err)
	}

	// 函数式写法
	// var hotUsersPtr atomic.Pointer[[]domain.User]
	// hotUsersPtr.Store(&[]domain.User{})
	//
	// // 处理热点用户变更事件
	// go func() {
	// 	watchChan := etcdClient.Watch(context.Background(), cfg.EtcdKey)
	// 	for watchResp := range watchChan {
	// 		for _, event := range watchResp.Events {
	// 			if event.Type == clientv3.EventTypePut {
	// 				// 更新热点用户
	// 				var vals []domain.User
	// 				if err1 := json.Unmarshal(event.Kv.Value, &vals); err1 == nil {
	// 					hotUsersPtr.Store(&vals)
	// 				}
	// 			}
	// 		}
	// 	}
	// }()
	//
	// loadFunc := cache.DataLoader(func(ctx context.Context) ([]*cache.Entry, error) {
	// 	var entries []*cache.Entry
	// 	const day = 24 * time.Hour
	// 	const defaultExpiration = 36500 * day
	// 	// 遍历热点用户信息
	// 	hotUsers := *hotUsersPtr.Reload()
	// 	for i := range hotUsers {
	// 		// 直接从数据库中加载数据
	// 		perms, err2 := repo.GetAll(ctx, hotUsers[i].BizID, hotUsers[i].ID)
	// 		if err2 == nil {
	// 			// 封装为entry
	// 			val, _ := json.Marshal(perms)
	// 			entries = append(entries, &cache.Entry{
	// 				Key:        cacheKeyFunc(hotUsers[i].BizID, hotUsers[i].ID),
	// 				Val:        val,
	// 				Expiration: defaultExpiration,
	// 			})
	// 		}
	// 	}
	// 	return entries, nil
	// })

	// 结构体封装写法
	hotUserLoader := NewHotUserLoader([]domain.User{}, repo, cacheKeyFunc)

	// 处理热点用户变更事件
	go func() {
		watchChan := etcdClient.Watch(context.Background(), cfg.EtcdKey)
		for watchResp := range watchChan {
			for _, event := range watchResp.Events {
				if event.Type == clientv3.EventTypePut {
					// 更新热点用户
					_ = hotUserLoader.UpdateUsers(event.Kv.Value)
				}
			}
		}
	}()

	return cache.NewMultiLevelCacheV2(r,
		local,
		// loadFunc, // 函数式写法
		hotUserLoader.LoadUserPermissionsFromDB, // 结构体封装写法
		cfg.LocalCacheRefreshPeriod,
		cfg.RedisPingTimeout,
		cfg.RedisHealthCheckPeriod,
		bitring.NewBitRing(
			cfg.ErrorEvents.BitRingSize,
			cfg.ErrorEvents.RateThreshold,
			cfg.ErrorEvents.ConsecutiveCount,
		))
}

type HotUserLoader struct {
	hotUsersPtr  atomic.Pointer[[]domain.User]
	repo         *repository.UserPermissionDefaultRepository
	cacheKeyFunc func(bizID, userID int64) string
}

func NewHotUserLoader(hotUsers []domain.User, repo *repository.UserPermissionDefaultRepository, cacheKeyFunc func(bizID, userID int64) string) *HotUserLoader {
	h := &HotUserLoader{
		repo:         repo,
		cacheKeyFunc: cacheKeyFunc,
	}
	h.hotUsersPtr.Store(&hotUsers)
	return h
}

func (h *HotUserLoader) LoadUserPermissionsFromDB(ctx context.Context) ([]*cache.Entry, error) {
	const day = 24 * time.Hour
	const defaultExpiration = 36500 * day
	// 遍历热点用户信息
	hotUsers := *h.hotUsersPtr.Load()
	entries := make([]*cache.Entry, 0, len(hotUsers))
	for i := range hotUsers {
		// 直接从数据库中加载数据
		perms, err := h.repo.GetAll(ctx, hotUsers[i].BizID, hotUsers[i].ID)
		if err == nil {
			// 封装为entry
			val, _ := json.Marshal(perms)
			entries = append(entries, &cache.Entry{
				Key:        h.cacheKeyFunc(hotUsers[i].BizID, hotUsers[i].ID),
				Val:        val,
				Expiration: defaultExpiration,
			})
		}
	}
	return entries, nil
}

func (h *HotUserLoader) UpdateUsers(value []byte) error {
	var users []domain.User
	err := json.Unmarshal(value, &users)
	if err == nil {
		// 更新热点用户
		h.hotUsersPtr.Store(&users)
	}
	return err
}
