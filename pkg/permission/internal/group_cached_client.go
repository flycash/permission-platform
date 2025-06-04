package internal

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	permissionv1 "gitee.com/flycash/permission-platform/api/proto/gen/permission/v1"
	"github.com/ecodeclub/ekit/slice"
	"github.com/golang/groupcache"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var _ permissionv1.PermissionServiceClient = (*GroupCachedClient)(nil)

type GroupCachedClient struct {
	baseCachedClient
	// 正常是 *AuthorizedClient
	client     permissionv1.PermissionServiceClient
	rbacClient permissionv1.RBACServiceClient
	// 分布式本地缓存
	cache  *groupcache.Group
	logger *slog.Logger
}

// NewRBACGRPCClient 根据传入的地址创建GRPC客户端，你需要再调用下方的 NewAuthorizedClient
func NewRBACGRPCClient(addr string) (permissionv1.RBACServiceClient, error) {
	conn, err := grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		// 生产环境要使用tls，并要提供tls配置
		// grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{})),
	)
	if err != nil {
		return nil, err
	}
	return permissionv1.NewRBACServiceClient(conn), nil
}

//nolint:gosec // 忽略
func NewGroupCachedClient(
	client permissionv1.PermissionServiceClient,
	rbacClient permissionv1.RBACServiceClient,
	cacheSize int64,
	groupName string,
	addr string,
	peerAddrs []string,
	logger *slog.Logger,
) *GroupCachedClient {
	c := &GroupCachedClient{
		baseCachedClient: baseCachedClient{},
		client:           client,
		rbacClient:       rbacClient,
		logger:           logger,
	}
	// 初始化 HTTPPool
	pool := groupcache.NewHTTPPool(addr)
	pool.Set(peerAddrs...)

	// 注册 Group
	groupcache.NewGroup(groupName, cacheSize, groupcache.GetterFunc(func(ctx context.Context, key string, dest groupcache.Sink) error {
		// 需要从key反解析出bizID和userID，再调用rbacClient获取该用户的全部权限
		// 但是一旦拿到后后续不再可变，直到该用户的信息因为lru算法被剔除。
		bizID, userID, err := c.parseKey(key)
		if err != nil {
			return err
		}
		resp, err := c.rbacClient.GetAllPermissions(ctx, &permissionv1.GetAllPermissionsRequest{
			BizId:  bizID,
			UserId: userID,
		})
		if err != nil {
			return err
		}
		// 将 resp.GetUserPermissions() 转化为 UserPermission
		up := UserPermission{
			UserID: userID,
			BizID:  bizID,
			Permissions: slice.Map(resp.GetUserPermissions(), func(_ int, src *permissionv1.UserPermission) PermissionV1 {
				return PermissionV1{
					Resource: Resource{
						Key:  src.GetResourceKey(),
						Type: src.GetResourceType(),
					},
					Action: src.GetPermissionAction(),
					Effect: src.GetEffect(),
				}
			}),
		}
		data, err := json.Marshal(up)
		if err != nil {
			c.logger.Error("GetterFunc中反序列化失败",
				slog.String("key", key),
				slog.Any("err", err),
			)
			return err // 或者返回一个特定的错误类型
		}
		return dest.SetBytes(data)
	}))

	// 启动 HTTPServer
	go func() {
		if err := http.ListenAndServe(addr[len("http://"):], pool); err != nil {
			c.logger.Error("监听失败",
				slog.Any("addr", addr),
				slog.Any("err", err),
			)
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
