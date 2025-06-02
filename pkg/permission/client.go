package permission

import (
	"context"

	permissionv1 "gitee.com/flycash/permission-platform/api/proto/gen/permission/v1"
	"google.golang.org/grpc"
)

type Client interface {
	Name() string
	CheckPermission(ctx context.Context, in *permissionv1.CheckPermissionRequest, opts ...grpc.CallOption) (*permissionv1.CheckPermissionResponse, error)
}
