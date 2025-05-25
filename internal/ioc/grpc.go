package ioc

import (
	permissionv1 "gitee.com/flycash/permission-platform/api/proto/gen/permission/v1"
	"gitee.com/flycash/permission-platform/internal/api/grpc/interceptor/audit"
	"gitee.com/flycash/permission-platform/internal/api/grpc/interceptor/auth"
	"gitee.com/flycash/permission-platform/internal/api/grpc/rbac"
	"gitee.com/flycash/permission-platform/internal/pkg/jwt"
	auditdao "gitee.com/flycash/permission-platform/internal/repository/dao/audit"
	"github.com/gotomicro/ego/server/egrpc"
)

func InitGRPC(
	crudServer *rbac.Server,
	permServer *rbac.PermissionServiceServer,
	token *jwt.Token,
	auditDAO auditdao.OperationLogDAO,
) []*egrpc.Component {
	authInterceptor := auth.New(token).Build()
	auditInterceptor := audit.New(auditDAO).Build()

	rbacServer := egrpc.Load("server.grpc.rbac").Build(
		egrpc.WithUnaryInterceptor(authInterceptor, auditInterceptor),
	)
	permissionv1.RegisterRBACServiceServer(rbacServer.Server, crudServer)
	permissionv1.RegisterPermissionServiceServer(rbacServer.Server, permServer)

	return []*egrpc.Component{rbacServer}
}
