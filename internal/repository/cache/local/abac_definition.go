package local

import (
	"context"
	"encoding/json"
	"gitee.com/flycash/permission-platform/internal/domain"
	repoCache "gitee.com/flycash/permission-platform/internal/repository/cache"
	"gitee.com/flycash/permission-platform/pkg/cache"
	"github.com/gotomicro/ego/core/elog"
	"github.com/redis/go-redis/v9"
	"strings"
	"time"
)

const (
	defaultExpiration = 24 * time.Hour
	defaultTimeout    = 5 * time.Second
)

type abacDefLocalCache struct {
	cache       cache.Cache
	redisClient *redis.Client
	logger      *elog.Component
}

func NewAbacDefLocalCache(cache cache.Cache, redisClient *redis.Client) repoCache.ABACDefinitionCache {
	localCache := &abacDefLocalCache{
		cache:       cache,
		redisClient: redisClient,
		logger:      elog.DefaultLogger,
	}
	go localCache.loop(context.Background())
	return localCache
}

func (a *abacDefLocalCache) GetDefinitions(ctx context.Context, bizID int64) (domain.BizAttrDefinition, error) {
	val := a.cache.Get(ctx, repoCache.DefKey(bizID))
	if val.Err != nil {
		return domain.BizAttrDefinition{}, val.Err
	}
	if val.KeyNotFound() {
		return domain.BizAttrDefinition{}, repoCache.ErrKeyNotFound
	}
	var res domain.BizAttrDefinition
	err := val.JSONScan(&res)
	return res, err
}

func (a *abacDefLocalCache) SetDefinitions(ctx context.Context, bizDef domain.BizAttrDefinition) error {
	key := repoCache.DefKey(bizDef.BizID)
	return a.cache.Set(ctx, key, bizDef, defaultExpiration)
}

func (a *abacDefLocalCache) loop(ctx context.Context) {
	// 就这个 channel 的表达式，你去问 deepseek
	pubsub := a.redisClient.PSubscribe(ctx, "__keyspace@*__:abac:def:*")
	defer pubsub.Close()
	ch := pubsub.Channel()
	for msg := range ch {
		// 在线上环境，小心别把敏感数据打出来了
		// 比如说你的 channel 里面包含了手机号码，你就别打了
		a.logger.Info("监控到 Redis 更新消息",
			elog.String("key", msg.Channel), elog.String("payload", msg.Payload))
		const channelMinLen = 2
		channel := msg.Channel
		channelStrList := strings.SplitN(channel, ":", channelMinLen)
		if len(channelStrList) < channelMinLen {
			a.logger.Error("监听到非法 Redis key", elog.String("channel", msg.Channel))
			continue
		}
		const keyIdx = 1
		key := channelStrList[keyIdx]
		ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
		eventType := msg.Payload
		a.handleConfigChange(ctx, key, eventType)
		cancel()
	}
}

func (a *abacDefLocalCache) handleConfigChange(ctx context.Context, key, event string) {
	// 自定义业务逻辑（如动态更新配置）
	switch event {
	case "set":
		res := a.redisClient.Get(ctx, key)
		if res.Err() != nil {
			a.logger.Error("订阅完获取键失败", elog.String("key", key))
		}
		var defs domain.BizAttrDefinition
		err := json.Unmarshal([]byte(res.Val()), &defs)
		if err != nil {
			a.logger.Error("序列化失败", elog.String("key", key), elog.String("val", res.Val()))
		}
		err = a.cache.Set(ctx, key, defs, defaultExpiration)
		if err != nil {
			a.logger.Error("本地缓存存储失败", elog.FieldErr(err), elog.String("key", key))
		}
	}
}
