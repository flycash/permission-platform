package auth

import (
	"context"
	"errors"
	"gitee.com/flycash/permission-platform/internal/errs"
	"gitee.com/flycash/permission-platform/internal/pkg/jwt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const BizIDName = "biz_id"

type InterceptorBuilder struct {
	token *jwt.Token
}

func New(token *jwt.Token) *InterceptorBuilder {
	return &InterceptorBuilder{
		token: token,
	}
}

func (b *InterceptorBuilder) Build() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// 1. 提取metadata
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "missing metadata")
		}

		// 2. 获取Authorization头
		authHeaders := md.Get("Authorization")
		if len(authHeaders) == 0 {
			return nil, status.Error(codes.Unauthenticated, "authorization token is required")
		}

		// 3. 处理Bearer Token格式
		tokenStr := authHeaders[0]
		// 4. 使用现有JwtAuth解码验证
		val, err := b.token.Decode(tokenStr)
		if err != nil {
			// 细化错误类型处理
			if errors.Is(err, jwt.ErrTokenExpired) {
				return nil, status.Error(codes.Unauthenticated, "token expired")
			}
			if errors.Is(err, jwt.ErrTokenSignatureInvalid) {
				return nil, status.Error(codes.Unauthenticated, "invalid signature")
			}
			return nil, status.Error(codes.Unauthenticated, "invalid token: "+err.Error())
		}
		v, ok := val[BizIDName]
		if ok {
			bizId := v.(float64)
			ctx = context.WithValue(ctx, BizIDName, int64(bizId))
		}
		return handler(ctx, req)
	}
}

func GetBizIDFromContext(ctx context.Context) (int64, error) {
	val := ctx.Value(BizIDName)
	if val == nil {
		return 0, errs.ErrBizIDNotFound
	}
	v, ok := val.(int64)
	if !ok {
		return 0, errs.ErrBizIDNotFound
	}
	return v, nil
}
