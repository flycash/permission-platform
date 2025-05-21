package rbac

import (
	"context"
	"strconv"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type baseServer struct{}

// 从gRPC上下文中获取业务ID
func (s *baseServer) getBizIDFromContext(ctx context.Context) (int64, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return 0, status.Error(codes.InvalidArgument, "无法获取元数据")
	}

	bizIDValues := md.Get("biz-id")
	if len(bizIDValues) == 0 {
		return 0, status.Error(codes.InvalidArgument, "未提供业务ID")
	}

	bizID, err := strconv.ParseInt(bizIDValues[0], 10, 64)
	if err != nil {
		return 0, status.Error(codes.InvalidArgument, "业务ID格式不正确")
	}

	return bizID, nil
}
