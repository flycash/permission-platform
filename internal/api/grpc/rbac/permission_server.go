package rbac

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	permissionpb "gitee.com/flycash/permission-platform/api/proto/gen/permission/v1"
	"gitee.com/flycash/permission-platform/internal/domain"
	"gitee.com/flycash/permission-platform/internal/service/rbac"
)

type PermissionServiceServer struct {
	permissionpb.UnimplementedPermissionServiceServer
	baseServer
	rbacService rbac.PermissionService
}

// NewPermissionServiceServer 创建权限服务器实例
func NewPermissionServiceServer(rbacService rbac.PermissionService) *PermissionServiceServer {
	return &PermissionServiceServer{
		rbacService: rbacService,
	}
}

// CheckPermission 检查用户是否有对特定资源的特定操作权限
func (s *PermissionServiceServer) CheckPermission(ctx context.Context, req *permissionpb.CheckPermissionRequest) (*permissionpb.CheckPermissionResponse, error) {
	// 参数校验
	if req.Uid <= 0 || req.Permission == nil || req.Permission.ResourceKey == "" || len(req.Permission.Actions) == 0 {
		return &permissionpb.CheckPermissionResponse{
			Allowed: false,
		}, nil
	}

	// 从metadata中获取bizID
	bizID, err := s.getBizIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// 检查所有action的权限
	// 调用服务层检查权限
	hasPermission, err1 := s.rbacService.Check(ctx, bizID, req.Uid, domain.Resource{
		BizID: bizID,
		Type:  req.Permission.ResourceType,
		Key:   req.Permission.ResourceKey,
	}, req.Permission.Actions)
	if err1 != nil {
		return nil, status.Error(codes.Internal, "检查权限时发生错误")
	}

	// 所有action都有权限
	return &permissionpb.CheckPermissionResponse{
		Allowed: hasPermission,
	}, nil
}
