package internal

import (
	"context"
	"encoding/json"
	"time"

	permissionv1 "gitee.com/flycash/permission-platform/api/proto/gen/permission/v1"
	"github.com/ecodeclub/ekit/slice"
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
	// 过期时间
	expiration time.Duration
}

func NewRedisCachedClient(client permissionv1.PermissionServiceClient, rd redis.Cmdable, expiration time.Duration) *RedisCachedClient {
	return &RedisCachedClient{
		baseCachedClient: baseCachedClient{},
		client:           client,
		rd:               rd,
		expiration:       expiration,
	}
}

func (c *RedisCachedClient) Name() string {
	return "RedisCachedClient"
}

func (c *RedisCachedClient) CheckPermission(ctx context.Context, in *permissionv1.CheckPermissionRequest, opts ...grpc.CallOption) (*permissionv1.CheckPermissionResponse, error) {
	// 1. 从redis缓存取
	userPermission, err := c.getFromCache(ctx, in.GetPermission().GetBizId(), in.GetUid())
	if err == nil {
		resp, err1 := c.checkPermission(userPermission, in)
		if err1 == nil {
			return resp, nil
		}
	}
	// 2. 从 client 取
	resp, err := c.client.CheckPermission(ctx, in, opts...)
	if err != nil {
		return nil, err
	}
	if resp.Allowed {
		c.setToCache(ctx, in)
	}
	return resp, nil
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

func (c *RedisCachedClient) setToCache(ctx context.Context, in *permissionv1.CheckPermissionRequest) {
	perm := in.GetPermission()
	up := UserPermission{
		UserID: in.GetUid(),
		BizID:  perm.GetBizId(),
		Permissions: slice.Map(perm.GetActions(), func(_ int, src string) PermissionV1 {
			return PermissionV1{
				Resource: Resource{
					Key:  perm.GetResourceKey(),
					Type: perm.GetResourceType(),
				},
				Action: src,
				Effect: "allow",
			}
		}),
	}
	data, _ := json.Marshal(up)
	c.rd.Set(ctx, c.cacheKey(perm.GetBizId(), in.GetUid()), string(data), c.expiration)
}
