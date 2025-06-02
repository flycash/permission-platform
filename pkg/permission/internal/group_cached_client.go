package internal

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	permissionv1 "gitee.com/flycash/permission-platform/api/proto/gen/permission/v1"
	"github.com/ecodeclub/ekit/slice"
	"github.com/golang/groupcache"
	"google.golang.org/grpc"
)

var _ permissionv1.PermissionServiceClient = (*GroupCachedClient)(nil)

type GroupCachedClient struct {
	baseCachedClient
	// 正常是 *AuthorizedClient
	client     permissionv1.PermissionServiceClient
	rbacClient permissionv1.RBACServiceClient
	// 分布式本地缓存
	cache *groupcache.Group
}

//nolint:gosec // 忽略
func NewGroupCachedClient(
	client permissionv1.PermissionServiceClient,
	rbacClient permissionv1.RBACServiceClient,
	cacheSize int64,
	groupName string,
	addr string,
	peerAddrs []string,
) *GroupCachedClient {
	c := &GroupCachedClient{
		baseCachedClient: baseCachedClient{},
		client:           client,
		rbacClient:       rbacClient,
	}
	// 初始化 HTTPPool
	pool := groupcache.NewHTTPPool(addr)
	pool.Set(peerAddrs...)

	// 注册 Group
	groupcache.NewGroup(groupName, cacheSize, groupcache.GetterFunc(func(ctx context.Context, _ string, dest groupcache.Sink) error {
		// 如何从远程获取
		// 需要从key反解析出bizID和userID，再调用rbacClient获取该用户的全部权限，但是一旦拿到后后续不再可变，直到该用户的信息因为lru算法被剔除。
		bizID, userID := int64(0), int64(0)
		resp, err := c.rbacClient.GetAllPermissions(ctx, &permissionv1.GetAllPermissionsRequest{
			BizId:  bizID,
			UserId: userID,
		})
		if err != nil {
			return err
		}
		up := slice.Map(resp.GetUserPermissions(), func(_ int, _ *permissionv1.UserPermission) UserPermission {
			return UserPermission{}
		})
		data, _ := json.Marshal(up)
		return dest.SetBytes(data)
	}))

	// 启动 HTTPServer
	go func() {
		if err := http.ListenAndServe(addr[len("https://"):], pool); err != nil {
			log.Fatalf("listen %s: %v", addr, err)
		}
	}()
	c.cache = groupcache.GetGroup(groupName)
	return c
}

func (c *GroupCachedClient) Name() string {
	return "GroupCachedClient"
}

func (c *GroupCachedClient) CheckPermission(ctx context.Context, in *permissionv1.CheckPermissionRequest, opts ...grpc.CallOption) (*permissionv1.CheckPermissionResponse, error) {
	var val []byte
	err := c.cache.Get(ctx, c.cacheKey(in.GetPermission().GetBizId(), in.GetUid()), groupcache.AllocatingByteSliceSink(&val))
	if err == nil {
		var up UserPermission
		_ = json.Unmarshal(val, &up)
		return c.checkPermission(up, in)
	}
	return c.client.CheckPermission(ctx, in, opts...)
}
