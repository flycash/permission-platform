package rbac

import (
	"context"

	"gitee.com/flycash/permission-platform/internal/api/grpc/interceptor/auth"
)

type baseServer struct{}

// 从gRPC上下文中获取业务ID
func (s *baseServer) getBizIDFromContext(ctx context.Context) (int64, error) {
	return auth.GetBizIDFromContext(ctx)
}
