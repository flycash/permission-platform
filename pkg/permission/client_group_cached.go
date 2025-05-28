package permission

import (
	"context"
	"encoding/json"

	permissionv1 "gitee.com/flycash/permission-platform/api/proto/gen/permission/v1"
	"gitee.com/flycash/permission-platform/internal/domain"
	"github.com/golang/groupcache"
	"google.golang.org/grpc"
)

var _ permissionv1.PermissionServiceClient = (*GroupCachedClient)(nil)

type GroupCachedClient struct {
	baseCachedClient
	// 正常是 *Client
	client     permissionv1.PermissionServiceClient
	rbacClient permissionv1.RBACServiceClient
	// 分布式本地缓存
	cache *groupcache.Group
}

func NewGroupCachedClient(client permissionv1.PermissionServiceClient, cacheSize int64) *GroupCachedClient {
	c := &GroupCachedClient{
		baseCachedClient: baseCachedClient{},
		client:           client,
	}
	c.cache = groupcache.NewGroup("client", cacheSize, groupcache.GetterFunc(func(ctx context.Context, key string, dest groupcache.Sink) error {
		// 如何从远程获取
		bizID, userID := int64(0), int64(0)
		_, err := c.rbacClient.GetAllPermissions(ctx, &permissionv1.GetAllPermissionsRequest{
			BizId:  bizID,
			UserId: userID,
		})
		if err != nil {
			return err
		}
		return nil
	}))
	return c
}

func (c *GroupCachedClient) CheckPermission(ctx context.Context, in *permissionv1.CheckPermissionRequest, opts ...grpc.CallOption) (*permissionv1.CheckPermissionResponse, error) {
	var val []byte
	err := c.cache.Get(ctx, c.cacheKey(in.GetUid()), groupcache.AllocatingByteSliceSink(&val))
	if err == nil {
		var up []domain.UserPermission
		_ = json.Unmarshal(val, &up)
		return c.CheckPermission(ctx, in)
	}
	return c.client.CheckPermission(ctx, in, opts...)
}
