package internal

import (
	"context"
	"encoding/json"
	permissionv1 "gitee.com/flycash/permission-platform/api/proto/gen/permission/v1"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
)

var _ permissionv1.PermissionServiceClient = (*RedisCachedClient)(nil)

type RedisCachedClient struct {
	baseCachedClient
	// 正常是 *AuthorizedClient
	client permissionv1.PermissionServiceClient
	// redis 缓存
	rd redis.Cmdable
}

func NewRedisCachedClient(client permissionv1.PermissionServiceClient, rd redis.Cmdable) *RedisCachedClient {
	return &RedisCachedClient{
		baseCachedClient: baseCachedClient{},
		client:           client,
		rd:               rd,
	}
}

func (c *RedisCachedClient) Name() string {
	return "RedisCachedClient"
}

func (c *RedisCachedClient) CheckPermission(ctx context.Context, in *permissionv1.CheckPermissionRequest, opts ...grpc.CallOption) (*permissionv1.CheckPermissionResponse, error) {
	// 1. 从redis缓存取，注意你不需要更新 Redis 的缓存
	userPermission, err := c.getFromCache(ctx, in.GetPermission().GetBizId(), in.GetUid())
	if err == nil {
		resp, err1 := c.checkPermission(userPermission, in)
		if err1 == nil {
			return resp, nil
		}
	}
	// 2. 从 client 取
	return c.client.CheckPermission(ctx, in, opts...)
}

func (c *RedisCachedClient) getFromCache(ctx context.Context, bizID, userID int64) (UserPermission, error) {
	val, err := c.rd.Get(ctx, c.cacheKey(bizID, userID)).Result()
	if err != nil {
		return UserPermission{}, err
	}
	var userPermission UserPermission
	err = json.Unmarshal([]byte(val), &userPermission)
	if err != nil {
		return UserPermission{}, err
	}
	return userPermission, nil
}
