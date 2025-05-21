package whitelist

import (
	"context"
	"strings"
	"sync"

	"gitee.com/flycash/permission-platform/internal/api/grpc/interceptor/auth"
	"github.com/ecodeclub/ekit/slice"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// InterceptorBuilder 它必须放在 JWT 之后
type InterceptorBuilder struct {
	// Biz ID  的白名单
	WhiteList []int64
	mutex     *sync.RWMutex
}

func NewInterceptorBuilder(whiteList []int64) *InterceptorBuilder {
	return &InterceptorBuilder{WhiteList: whiteList, mutex: &sync.RWMutex{}}
}

func (i *InterceptorBuilder) Build() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		i.mutex.RLock()
		defer i.mutex.RUnlock()
		if strings.Contains(info.FullMethod, "BusinessConfig") {
			bizID, err := auth.GetBizIDFromContext(ctx)
			if err != nil {
				return nil, status.Errorf(codes.PermissionDenied, "没有权限，原因: %s", err.Error())
			}
			if !slice.Contains(i.WhiteList, bizID) {
				return nil, status.Errorf(codes.PermissionDenied, "不在白名单内")
			}
		}
		return handler(ctx, req)
	}
}

func (i *InterceptorBuilder) UpdateWhiteList(list []int64) {
	// 并发安全
	i.mutex.Lock()
	defer i.mutex.Unlock()
	i.WhiteList = list
}
