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

type PermissionServer struct {
	permissionpb.UnimplementedPermissionServiceServer
	rbacService rbac.PermissionService
}

// NewPermissionServer 创建权限服务器实例
func NewPermissionServer(rbacService rbac.PermissionService) *PermissionServer {
	return &PermissionServer{
		rbacService: rbacService,
	}
}

// CheckPermission 检查用户是否有对特定资源的特定操作权限
func (s *PermissionServer) CheckPermission(ctx context.Context, req *permissionpb.CheckPermissionRequest) (*permissionpb.CheckPermissionResponse, error) {
	// 参数校验
	if req.Uid <= 0 || req.Permission == nil || req.Permission.ResourceKey == "" || len(req.Permission.Actions) == 0 {
		return &permissionpb.CheckPermissionResponse{
			Allowed: false,
		}, nil
	}

	// 从metadata中获取bizID
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.InvalidArgument, "无法获取元数据")
	}

	bizIDValues := md.Get("biz-id")
	if len(bizIDValues) == 0 {
		return nil, status.Error(codes.InvalidArgument, "未提供业务ID")
	}

	bizID, err := strconv.ParseInt(bizIDValues[0], 10, 64)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "业务ID格式不正确")
	}

	// 检查所有action的权限
	allowed := true
	for _, protoAction := range req.Permission.Actions {
		// 将proto的action转换为domain模型的Permission
		domainPermission := domain.Permission{
			ResourceKey: req.Permission.ResourceKey,
			Action:      s.convertProtoActionToDomainAction(protoAction),
		}

		// 调用服务层检查权限
		hasPermission, err := s.rbacService.Check(ctx, bizID, req.Uid, domainPermission)
		if err != nil {
			return nil, status.Error(codes.Internal, "检查权限时发生错误")
		}

		if !hasPermission {
			allowed = false
			break
		}
	}

	// 返回响应
	return &permissionpb.CheckPermissionResponse{
		Allowed: allowed,
	}, nil
}

// convertProtoActionToDomainAction 将proto定义的操作类型转换为领域模型的操作类型
func (s *PermissionServer) convertProtoActionToDomainAction(action permissionpb.ActionType) domain.ActionType {
	switch action {
	case permissionpb.ActionType_READ:
		return domain.ActionTypeRead
	case permissionpb.ActionType_WRITE:
		return domain.ActionTypeWrite
	default:
		return domain.ActionTypeRead // 默认为读权限
	}
}
