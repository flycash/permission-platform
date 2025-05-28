package permission

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	permissionv1 "gitee.com/flycash/permission-platform/api/proto/gen/permission/v1"
	"gitee.com/flycash/permission-platform/pkg/kafka"
	"github.com/patrickmn/go-cache"
	"google.golang.org/grpc"
)

var _ permissionv1.PermissionServiceClient = (*LocalCachedClient)(nil)

type LocalCachedClient struct {
	baseCachedClient
	// 正常是 *Client 或者 *RedisCachedClient
	client permissionv1.PermissionServiceClient
	// 本地缓存
	cache *cache.Cache
	// 监听用户权限事件
	consumer kafka.Consumer
}

func NewLocalCachedClient(
	client permissionv1.PermissionServiceClient,
	consumer kafka.Consumer,
	topic string,
	cacheDefaultExpiration, cacheCleanupInterval time.Duration,
) (*LocalCachedClient, error) {
	c := &LocalCachedClient{
		baseCachedClient: baseCachedClient{},
		client:           client,
		cache:            cache.New(cacheDefaultExpiration, cacheCleanupInterval),
	}
	if topic == "" {
		topic = "user-permission-events"
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
				// c.logger.Error("消费用户权限Binlog事件失败", elog.FieldErr(er))
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
		// c.logger.Warn("解析消息失败",
		// 	elog.FieldErr(err),
		// 	elog.Any("msg", msg))
		return err
	}

	// 更新本地缓存
	for uid := range evt.Permissions {
		c.cache.Set(c.cacheKey(uid), evt.Permissions[uid], 0)
	}

	// 消费完成，提交消费进度
	_, err = c.consumer.CommitMessage(msg)
	if err != nil {
		// c.logger.Warn("提交消息失败",
		// 	elog.FieldErr(err),
		// 	elog.Any("msg", msg))
		return err
	}
	return nil
}

func (c *LocalCachedClient) CheckPermission(ctx context.Context, in *permissionv1.CheckPermissionRequest, opts ...grpc.CallOption) (*permissionv1.CheckPermissionResponse, error) {
	// 1. 从本地缓存取
	val, ok := c.cache.Get(c.cacheKey(in.GetUid()))
	if ok {
		up, _ := val.(UserPermission)
		// 假定这里的up就是与db一直的，因为会有异步携程消费消息存入本地缓存，那么可以直接
		return c.checkPermission(up, in)
		// 如果假设不成立，那么需要，再走client，再回填
		// resp, err1 := c.checkPermission(up, in)
		// if err1 == nil {
		// 	return resp, nil
		// }
	}
	// 2. 从 client 取
	return c.client.CheckPermission(ctx, in, opts...)
}
