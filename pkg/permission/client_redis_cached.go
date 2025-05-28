package permission

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
	// 正常是 *Client
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

func (c *RedisCachedClient) CheckPermission(ctx context.Context, in *permissionv1.CheckPermissionRequest, opts ...grpc.CallOption) (*permissionv1.CheckPermissionResponse, error) {
	// 1. 从redis缓存取
	userPermission, err := c.getFromCache(ctx, in.GetUid())
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
		c.setToCache(ctx, userPermission, in)
	}
	return resp, nil
}

func (c *RedisCachedClient) getFromCache(ctx context.Context, uid int64) (UserPermission, error) {
	val, err := c.rd.Get(ctx, c.cacheKey(uid)).Result()
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

func (c *RedisCachedClient) setToCache(ctx context.Context, up UserPermission, in *permissionv1.CheckPermissionRequest) {
	uid := in.GetUid()
	permission := in.GetPermission()
	// 因为up可能是零值，也就是之前不再缓存中，所以需要设置一下bizID和userID
	up.BizID = permission.GetBizId()
	up.UserID = uid
	// 去掉与up中actions重合的部分
	existedActions := slice.Map(up.Permissions, func(_ int, p PermissionV1) string {
		return p.Action
	})
	actions := slice.FilterDelete(permission.GetActions(), func(_ int, action string) bool {
		return slice.Contains(existedActions, action)
	})
	// 添加新增的 PermissionV1
	up.Permissions = append(up.Permissions, slice.Map(actions, func(_ int, src string) PermissionV1 {
		return PermissionV1{
			Resource: Resource{
				Key:  permission.GetResourceKey(),
				Type: permission.GetResourceType(),
			},
			Action: src,
			Effect: "allow",
		}
	})...)
	data, _ := json.Marshal(up)
	c.rd.Set(ctx, c.cacheKey(uid), string(data), c.expiration)
}
