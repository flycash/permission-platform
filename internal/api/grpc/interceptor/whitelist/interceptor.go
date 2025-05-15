package whitelist

import (
	"context"
	"gitee.com/flycash/permission-platform/internal/api/grpc/interceptor/jwt"
	"github.com/ecodeclub/ekit/slice"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"strings"
)

type InterceptorBuilder struct {
	// Biz ID  的白名单
	WhiteList []int64
}

func (i *InterceptorBuilder) Build() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		// 操作 BusinessConfig 的代码
		if strings.Contains(info.FullMethod, "BusinessConfig") {
			bizID, err := jwt.GetBizIDFromContext(ctx)
			if err != nil {
				// 没有 Biz ID 肯定不能操作
				return nil, err
			}
			if !slice.Contains(i.WhiteList, bizID) {
				return nil, status.Errorf(codes.PermissionDenied, "没有权限")
			}
		}
		return handler(ctx, req)
	}
}
