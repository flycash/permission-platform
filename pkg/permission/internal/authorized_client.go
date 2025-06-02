package internal

import (
	"context"

	permissionv1 "gitee.com/flycash/permission-platform/api/proto/gen/permission/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

var _ permissionv1.PermissionServiceClient = (*AuthorizedClient)(nil)

// AuthorizedClient
type AuthorizedClient struct {
	client permissionv1.PermissionServiceClient
	token  string
}

// NewGRPCClient 根据传入的地址创建GRPC客户端，你需要再调用下方的 NewClient
func NewGRPCClient(addr string) (permissionv1.PermissionServiceClient, error) {
	conn, err := grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		// 生产环境要使用tls，并要提供tls配置
		// grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{})),
	)
	if err != nil {
		return nil, err
	}
	return permissionv1.NewPermissionServiceClient(conn), nil
}

// NewClient 根据传入的GRPC客户端和认证token创建客户端
func NewClient(client permissionv1.PermissionServiceClient, token string) *AuthorizedClient {
	return &AuthorizedClient{
		client: client,
		token:  token,
	}
}

func (c *AuthorizedClient) Name() string {
	return "AuthorizedClient"
}

func (c *AuthorizedClient) CheckPermission(ctx context.Context, in *permissionv1.CheckPermissionRequest, opts ...grpc.CallOption) (*permissionv1.CheckPermissionResponse, error) {
	ctx = metadata.AppendToOutgoingContext(ctx, "Authorization", c.token)
	return c.client.CheckPermission(ctx, in, opts...)
}
