package audit

import (
	"context"
	"encoding/json"
	"fmt"

	"gitee.com/flycash/permission-platform/internal/api/grpc/interceptor/auth"
	"gitee.com/flycash/permission-platform/internal/repository/dao/audit"
	"github.com/gotomicro/ego/core/elog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type InterceptorBuilder struct {
	dao    audit.OperationLogDAO
	logger *elog.Component
}

func New(dao audit.OperationLogDAO) *InterceptorBuilder {
	return &InterceptorBuilder{
		dao:    dao,
		logger: elog.DefaultLogger.With(elog.FieldName("audit.OperationLog")),
	}
}

func (b *InterceptorBuilder) Build() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		var operationLog audit.OperationLog

		// 1. 从metadata中获取可选的操作者ID和请求唯一标识
		md, ok := metadata.FromIncomingContext(ctx)
		if ok {
			operator := md.Get("operator")
			if len(operator) != 0 {
				operationLog.Operator = operator[0]
			}
			key := md.Get("key")
			if len(key) != 0 {
				operationLog.Key = key[0]
			}
		}

		// 2.从Ctx中获取bizID，过了auth拦截器这里必然有
		operationLog.BizID, _ = auth.GetBizIDFromContext(ctx)

		// 3. 填充要调用的接口名称及参数
		operationLog.Method = info.FullMethod
		data, err := json.Marshal(req)
		if err != nil {
			b.logger.Error("JSON序列化请求失败",
				elog.FieldErr(err),
				elog.FieldKey("req"),
				elog.FieldValueAny(req))
			operationLog.Request = fmt.Sprintf("%#v", req)
		} else {
			operationLog.Request = string(data)
		}

		// 4. 持久化操作日志
		_, err = b.dao.Create(ctx, operationLog)
		if err != nil {
			b.logger.Error("存储操作日志失败",
				elog.FieldErr(err),
				elog.FieldKey("operationLog"),
				elog.FieldValueAny(operationLog))
		}

		return handler(ctx, req)
	}
}
