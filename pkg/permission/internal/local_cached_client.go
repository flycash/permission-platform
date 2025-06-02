package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync/atomic"
	"time"

	permissionv1 "gitee.com/flycash/permission-platform/api/proto/gen/permission/v1"
	"gitee.com/flycash/permission-platform/pkg/kafka"
	"github.com/patrickmn/go-cache"
	"google.golang.org/grpc"
)

var _ permissionv1.PermissionServiceClient = (*LocalCachedClient)(nil)

type LocalCachedClient struct {
	baseCachedClient
	// 正常是 *AuthorizedClient 或者 *RedisCachedClient
	client permissionv1.PermissionServiceClient
	// 本地缓存
	cache *cache.Cache
	// 监听用户权限事件
	consumer kafka.Consumer

	Timeout atomic.Value
	logger  *slog.Logger
}

func NewLocalCachedClient(
	client permissionv1.PermissionServiceClient,
	consumer kafka.Consumer,
	topic string,
	cacheDefaultExpiration, cacheCleanupInterval time.Duration,
	logger *slog.Logger,
) (*LocalCachedClient, error) {
	c := &LocalCachedClient{
		baseCachedClient: baseCachedClient{},
		client:           client,
		cache:            cache.New(cacheDefaultExpiration, cacheCleanupInterval),
		logger:           logger,
	}
	err := consumer.Subscribe(context.Background(), topic, nil)
	if err != nil {
		return nil, err
	}
	c.consumer = consumer
	go func() {
		for {
			er := c.Consume(context.Background())
			if er != nil {
				c.logger.Error("消费用户权限Binlog事件失败", slog.Any("err", err))
			}
		}
	}()
	return c, nil
}

func (c *LocalCachedClient) Consume(ctx context.Context) error {
	// 获取消息
	const defaultTimeout = time.Second * 15
	msg, err := c.consumer.ReadMessage(ctx, defaultTimeout)
	if err != nil {
		return fmt.Errorf("获取消息失败: %w", err)
	}

	// 解析消息
	var evt UserPermissionEvent
	err = json.Unmarshal(msg.Value, &evt)
	if err != nil {
		c.logger.Warn("解析消息失败",
			slog.Any("err", err),
			slog.Any("msg", msg))
		return err
	}

	// 更新本地缓存
	for uid := range evt.Permissions {
		c.cache.Set(c.cacheKey(evt.Permissions[uid].BizID, uid), evt.Permissions[uid], 0)
	}

	// 消费完成，提交消费进度
	_, err = c.consumer.CommitMessage(msg)
	if err != nil {
		c.logger.Warn("提交消息失败",
			slog.Any("err", err),
			slog.Any("msg", msg))
		return err
	}
	return nil
}

func (c *LocalCachedClient) Name() string {
	return "LocalCachedClient"
}

func (c *LocalCachedClient) CheckPermission(ctx context.Context, in *permissionv1.CheckPermissionRequest, opts ...grpc.CallOption) (*permissionv1.CheckPermissionResponse, error) {
	// 1. 从本地缓存取
	val, ok := c.cache.Get(c.cacheKey(in.GetPermission().GetBizId(), in.GetUid()))
	if ok {
		up, _ := val.(UserPermission)
		// 假定这里得到的up是与数据库中一致的，因为会有异步协程消费消息存入本地缓存，那么可以直接
		return c.checkPermission(up, in)
		// 如果假设不成立即up与数据库中不一致，那么需要，先走client再回填，注意看client_redis_cached.go中 CheckPermission 的实现
		// resp, err1 := c.checkPermission(up, in)
		// if err1 == nil {
		// 	return resp, nil
		// }
	}
	// 2. 从 client 取
	return c.client.CheckPermission(ctx, in, opts...)
}
