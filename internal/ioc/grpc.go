package ioc

import (
	permissionv1 "gitee.com/flycash/permission-platform/api/proto/gen/permission/v1"
	"gitee.com/flycash/permission-platform/internal/api/grpc/interceptor/auth"
	"gitee.com/flycash/permission-platform/internal/api/grpc/rbac"
	"gitee.com/flycash/permission-platform/internal/pkg/jwt"
	"github.com/gotomicro/ego/server/egrpc"
)

func InitGRPC(
	crudServer *rbac.Server,
	permServer *rbac.PermissionServiceServer,
	token *jwt.Token,
) []*egrpc.Component {
	authInterceptor := auth.New(token).Build()

	rbacServer := egrpc.Load("server.grpc.rbac").Build(
		egrpc.WithUnaryInterceptor(authInterceptor),
	)
	permissionv1.RegisterRBACServiceServer(rbacServer.Server, crudServer)
	permissionv1.RegisterPermissionServiceServer(rbacServer.Server, permServer)

	return []*egrpc.Component{rbacServer}
}
