package permission

import (
	"context"
	"log/slog"
	"time"

	permissionv1 "gitee.com/flycash/permission-platform/api/proto/gen/permission/v1"
	"gitee.com/flycash/permission-platform/pkg/kafka"
	"gitee.com/flycash/permission-platform/pkg/permission/internal"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
)

type (
	Client interface {
		Name() string
		CheckPermission(ctx context.Context, in *permissionv1.CheckPermissionRequest, opts ...grpc.CallOption) (*permissionv1.CheckPermissionResponse, error)
	}

	AuthorizedClient  = internal.AuthorizedClient
	AggregateClient   = internal.AggregatePermissionClient
	GroupCachedClient = internal.GroupCachedClient
	LocalCachedClient = internal.LocalCachedClient
	RedisCachedClient = internal.RedisCachedClient
)

// NewPermissionGRPCClient 根据传入的地址创建GRPC客户端，你需要再调用下方的 NewAuthorizedClient
func NewPermissionGRPCClient(addr string) (permissionv1.PermissionServiceClient, error) {
	return internal.NewPermissionGRPCClient(addr)
}

// NewRBACGRPCClient 根据传入的地址创建RBAC GRPC客户端
func NewRBACGRPCClient(addr string) (permissionv1.RBACServiceClient, error) {
	return internal.NewRBACGRPCClient(addr)
}

// NewAuthorizedClient 根据传入的GRPC客户端和认证token创建客户端
func NewAuthorizedClient(client permissionv1.PermissionServiceClient, token string) *AuthorizedClient {
	return internal.NewAuthorizedClient(client, token)
}

// NewGroupCachedClient 创建基于GroupCache的分布式缓存客户端
func NewGroupCachedClient(
	client permissionv1.PermissionServiceClient,
	rbacClient permissionv1.RBACServiceClient,
	cacheSize int64,
	groupName string,
	addr string,
	peerAddrs []string,
	logger *slog.Logger,
) *GroupCachedClient {
	return internal.NewGroupCachedClient(client, rbacClient, cacheSize, groupName, addr, peerAddrs, logger)
}

// NewLocalCachedClient 创建本地缓存客户端
func NewLocalCachedClient(
	client permissionv1.PermissionServiceClient,
	consumer kafka.Consumer,
	topic string,
	cacheDefaultExpiration, cacheCleanupInterval time.Duration,
	logger *slog.Logger,
) (*LocalCachedClient, error) {
	return internal.NewLocalCachedClient(client, consumer, topic, cacheDefaultExpiration, cacheCleanupInterval, logger)
}

// NewRedisCachedClient 创建Redis缓存客户端
func NewRedisCachedClient(client permissionv1.PermissionServiceClient, rd redis.Cmdable, expiration time.Duration) *RedisCachedClient {
	return internal.NewRedisCachedClient(client, rd, expiration)
}
