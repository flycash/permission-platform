package grpc

import (
	"context"
	"strconv"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	permissionpb "gitee.com/flycash/permission-platform/api/proto/gen/permission/v1"
	"gitee.com/flycash/permission-platform/internal/domain"
	"gitee.com/flycash/permission-platform/internal/service/rbac"
)

type PermissionServiceServer struct {
	permissionpb.UnimplementedPermissionServiceServer
	rbacService rbac.PermissionService
}

// NewPermissionServiceServer 创建权限服务器实例
func NewPermissionServiceServer(rbacService rbac.PermissionService) *PermissionServiceServer {
	return &PermissionServiceServer{
		rbacService: rbacService,
	}
}

// 从gRPC上下文中获取业务ID
func (s *PermissionServiceServer) getBizIDFromContext(ctx context.Context) (int64, error) {
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
	for i := range req.Permission.Actions {
		// 将proto的权限转换为domain模型的Permission
		domainPermission := domain.Permission{
			BizID: bizID,
			Resource: domain.Resource{
				BizID: bizID,
				Key:   req.Permission.ResourceKey,
				Type:  req.Permission.ResourceType,
			},
			Action: req.Permission.Actions[i],
		}

		// 调用服务层检查权限
		hasPermission, err := s.rbacService.Check(ctx, bizID, req.Uid, domainPermission)
		if err != nil {
			return nil, status.Error(codes.Internal, "检查权限时发生错误")
		}

		if !hasPermission {
			return &permissionpb.CheckPermissionResponse{
				Allowed: false,
			}, nil
		}
	}

	// 所有action都有权限
	return &permissionpb.CheckPermissionResponse{
		Allowed: true,
	}, nil
}
