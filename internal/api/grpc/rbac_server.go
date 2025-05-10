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

type RBACServer struct {
	permissionpb.UnimplementedRBACServiceServer
	rbacService rbac.RBACService
}

// NewRBACServer 创建RBAC服务器实例
func NewRBACServer(rbacService rbac.RBACService) *RBACServer {
	return &RBACServer{
		rbacService: rbacService,
	}
}

// 从gRPC上下文中获取调用者ID
func getUserIDFromContext(ctx context.Context) (int64, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return 0, status.Error(codes.Unauthenticated, "缺少认证信息")
	}

	userIDValues := md.Get("user-id")
	if len(userIDValues) == 0 {
		return 0, status.Error(codes.Unauthenticated, "缺少用户ID")
	}

	userID, err := strconv.ParseInt(userIDValues[0], 10, 64)
	if err != nil {
		return 0, status.Error(codes.InvalidArgument, "用户ID格式不正确")
	}

	return userID, nil
}

// 从proto消息转换为领域模型
func protoToDomainRole(protoRole *permissionpb.Role) domain.Role {
	var roleType domain.RoleType
	switch protoRole.Type {
	case "system":
		roleType = domain.RoleTypeSystem
	case "custom":
		roleType = domain.RoleTypeCustom
	case "temporary":
		roleType = domain.RoleTypeTemporary
	default:
		roleType = domain.RoleTypeCustom
	}

	return domain.Role{
		ID:          protoRole.Id,
		BizID:       protoRole.BizId,
		Type:        roleType,
		Name:        protoRole.Name,
		Description: protoRole.Description,
		StartTime:   protoRole.StartTime,
		EndTime:     protoRole.EndTime,
	}
}

// 从领域模型转换为proto消息
func domainToProtoRole(domainRole domain.Role) *permissionpb.Role {
	var roleType string
	switch domainRole.Type {
	case domain.RoleTypeSystem:
		roleType = "system"
	case domain.RoleTypeCustom:
		roleType = "custom"
	case domain.RoleTypeTemporary:
		roleType = "temporary"
	default:
		roleType = "custom"
	}

	return &permissionpb.Role{
		Id:          domainRole.ID,
		BizId:       domainRole.BizID,
		Type:        roleType,
		Name:        domainRole.Name,
		Description: domainRole.Description,
		StartTime:   domainRole.StartTime,
		EndTime:     domainRole.EndTime,
	}
}

// 从proto消息转换为领域模型
func protoToDomainResource(protoResource *permissionpb.Resource) domain.Resource {
	return domain.Resource{
		ID:          protoResource.Id,
		BizID:       protoResource.BizId,
		Type:        protoResource.Type,
		Key:         protoResource.Key,
		Name:        protoResource.Name,
		Description: protoResource.Description,
	}
}

// 从领域模型转换为proto消息
func domainToProtoResource(domainResource domain.Resource) *permissionpb.Resource {
	return &permissionpb.Resource{
		Id:          domainResource.ID,
		BizId:       domainResource.BizID,
		Type:        domainResource.Type,
		Key:         domainResource.Key,
		Name:        domainResource.Name,
		Description: domainResource.Description,
	}
}

// 从proto消息转换为领域模型
func protoToDomainPermission(protoPermission *permissionpb.Permission) domain.Permission {
	var action domain.ActionType
	// proto中Permission.actions是一个repeated ActionType
	if len(protoPermission.Actions) > 0 {
		switch protoPermission.Actions[0] {
		case permissionpb.ActionType_READ:
			action = domain.ActionTypeRead
		case permissionpb.ActionType_WRITE:
			action = domain.ActionTypeWrite
		default:
			action = domain.ActionTypeRead
		}
	} else {
		action = domain.ActionTypeRead
	}

	return domain.Permission{
		ID:           protoPermission.Id,
		BizID:        protoPermission.BizId,
		Name:         protoPermission.Name,
		Description:  protoPermission.Description,
		ResourceID:   protoPermission.ResourceId,
		ResourceType: protoPermission.ResourceType,
		ResourceKey:  protoPermission.ResourceKey,
		Action:       action,
	}
}

// 从领域模型转换为proto消息
func domainToProtoPermission(domainPermission domain.Permission) *permissionpb.Permission {
	var protoAction permissionpb.ActionType
	switch domainPermission.Action {
	case domain.ActionTypeRead:
		protoAction = permissionpb.ActionType_READ
	case domain.ActionTypeWrite:
		protoAction = permissionpb.ActionType_WRITE
	default:
		protoAction = permissionpb.ActionType_READ
	}

	return &permissionpb.Permission{
		Id:           domainPermission.ID,
		BizId:        domainPermission.BizID,
		Name:         domainPermission.Name,
		Description:  domainPermission.Description,
		ResourceId:   domainPermission.ResourceID,
		ResourceType: domainPermission.ResourceType,
		ResourceKey:  domainPermission.ResourceKey,
		Actions:      []permissionpb.ActionType{protoAction},
	}
}

// 从领域模型转换为proto消息
func domainToProtoUserRole(domainUserRole domain.UserRole) *permissionpb.UserRole {
	var roleType string
	switch domainUserRole.RoleType {
	case domain.RoleTypeSystem:
		roleType = "system"
	case domain.RoleTypeCustom:
		roleType = "custom"
	case domain.RoleTypeTemporary:
		roleType = "temporary"
	default:
		roleType = "custom"
	}

	return &permissionpb.UserRole{
		Id:        domainUserRole.ID,
		BizId:     domainUserRole.BizID,
		UserId:    domainUserRole.UserID,
		RoleId:    domainUserRole.RoleID,
		RoleName:  domainUserRole.RoleName,
		RoleType:  roleType,
		StartTime: domainUserRole.StartTime,
		EndTime:   domainUserRole.EndTime,
	}
}

// ===== Role相关接口实现 =====

// CreateRole 创建角色
func (s *RBACServer) CreateRole(ctx context.Context, req *permissionpb.CreateRoleRequest) (*permissionpb.CreateRoleResponse, error) {
	// 获取调用者ID
	_, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// 参数校验
	if req.Role == nil {
		return nil, status.Error(codes.InvalidArgument, "缺少角色信息")
	}

	// 转换领域模型
	domainRole := protoToDomainRole(req.Role)

	// 调用服务层
	createdRole, err := s.rbacService.CreateRole(ctx, domainRole)
	if err != nil {
		return nil, status.Error(codes.Internal, "创建角色失败: "+err.Error())
	}

	// 返回结果
	return &permissionpb.CreateRoleResponse{
		Role: domainToProtoRole(createdRole),
	}, nil
}

// GetRole 获取角色
func (s *RBACServer) GetRole(ctx context.Context, req *permissionpb.GetRoleRequest) (*permissionpb.GetRoleResponse, error) {
	// 获取调用者ID
	_, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// 参数校验
	if req.Id <= 0 {
		return nil, status.Error(codes.InvalidArgument, "角色ID无效")
	}

	// 调用服务层
	role, err := s.rbacService.GetRole(ctx, req.Id)
	if err != nil {
		return nil, status.Error(codes.Internal, "获取角色失败: "+err.Error())
	}

	// 返回结果
	return &permissionpb.GetRoleResponse{
		Role: domainToProtoRole(role),
	}, nil
}

// UpdateRole 更新角色
func (s *RBACServer) UpdateRole(ctx context.Context, req *permissionpb.UpdateRoleRequest) (*permissionpb.UpdateRoleResponse, error) {
	// 获取调用者ID
	_, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// 参数校验
	if req.Role == nil || req.Role.Id <= 0 {
		return nil, status.Error(codes.InvalidArgument, "角色ID无效")
	}

	// 转换领域模型
	domainRole := protoToDomainRole(req.Role)

	// 调用服务层
	updatedRole, err := s.rbacService.UpdateRole(ctx, domainRole)
	if err != nil {
		return nil, status.Error(codes.Internal, "更新角色失败: "+err.Error())
	}

	// 返回结果
	return &permissionpb.UpdateRoleResponse{
		Role: domainToProtoRole(updatedRole),
	}, nil
}

// DeleteRole 删除角色
func (s *RBACServer) DeleteRole(ctx context.Context, req *permissionpb.DeleteRoleRequest) (*permissionpb.DeleteRoleResponse, error) {
	// 获取调用者ID
	_, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// 参数校验
	if req.Id <= 0 {
		return nil, status.Error(codes.InvalidArgument, "角色ID无效")
	}

	// 调用服务层
	err = s.rbacService.DeleteRole(ctx, req.Id)
	if err != nil {
		return nil, status.Error(codes.Internal, "删除角色失败: "+err.Error())
	}

	// 返回结果
	return &permissionpb.DeleteRoleResponse{
		Success: true,
	}, nil
}

// ListRoles 获取角色列表
func (s *RBACServer) ListRoles(ctx context.Context, req *permissionpb.ListRolesRequest) (*permissionpb.ListRolesResponse, error) {
	// 获取调用者ID
	_, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// 参数校验
	if req.BizId <= 0 {
		return nil, status.Error(codes.InvalidArgument, "业务ID无效")
	}

	// 调用服务层
	roles, total, err := s.rbacService.ListRoles(ctx, req.BizId, int(req.Offset), int(req.Limit), req.Type)
	if err != nil {
		return nil, status.Error(codes.Internal, "获取角色列表失败: "+err.Error())
	}

	// 转换结果
	protoRoles := make([]*permissionpb.Role, 0, len(roles))
	for _, role := range roles {
		protoRoles = append(protoRoles, domainToProtoRole(role))
	}

	// 返回结果
	return &permissionpb.ListRolesResponse{
		Roles: protoRoles,
		Total: int32(total),
	}, nil
}

// ===== Resource相关接口实现 =====

// CreateResource 创建资源
func (s *RBACServer) CreateResource(ctx context.Context, req *permissionpb.CreateResourceRequest) (*permissionpb.CreateResourceResponse, error) {
	// 获取调用者ID
	_, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// 参数校验
	if req.Resource == nil {
		return nil, status.Error(codes.InvalidArgument, "缺少资源信息")
	}

	// 转换领域模型
	domainResource := protoToDomainResource(req.Resource)

	// 调用服务层
	createdResource, err := s.rbacService.CreateResource(ctx, domainResource)
	if err != nil {
		return nil, status.Error(codes.Internal, "创建资源失败: "+err.Error())
	}

	// 返回结果
	return &permissionpb.CreateResourceResponse{
		Resource: domainToProtoResource(createdResource),
	}, nil
}

// GetResource 获取资源
func (s *RBACServer) GetResource(ctx context.Context, req *permissionpb.GetResourceRequest) (*permissionpb.GetResourceResponse, error) {
	// 获取调用者ID
	_, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// 参数校验
	if req.Id <= 0 {
		return nil, status.Error(codes.InvalidArgument, "资源ID无效")
	}

	// 调用服务层
	resource, err := s.rbacService.GetResource(ctx, req.Id)
	if err != nil {
		return nil, status.Error(codes.Internal, "获取资源失败: "+err.Error())
	}

	// 返回结果
	return &permissionpb.GetResourceResponse{
		Resource: domainToProtoResource(resource),
	}, nil
}

// UpdateResource 更新资源
func (s *RBACServer) UpdateResource(ctx context.Context, req *permissionpb.UpdateResourceRequest) (*permissionpb.UpdateResourceResponse, error) {
	// 获取调用者ID
	_, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// 参数校验
	if req.Resource == nil || req.Resource.Id <= 0 {
		return nil, status.Error(codes.InvalidArgument, "资源ID无效")
	}

	// 转换领域模型
	domainResource := protoToDomainResource(req.Resource)

	// 调用服务层
	updatedResource, err := s.rbacService.UpdateResource(ctx, domainResource)
	if err != nil {
		return nil, status.Error(codes.Internal, "更新资源失败: "+err.Error())
	}

	// 返回结果
	return &permissionpb.UpdateResourceResponse{
		Resource: domainToProtoResource(updatedResource),
	}, nil
}

// DeleteResource 删除资源
func (s *RBACServer) DeleteResource(ctx context.Context, req *permissionpb.DeleteResourceRequest) (*permissionpb.DeleteResourceResponse, error) {
	// 获取调用者ID
	_, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// 参数校验
	if req.Id <= 0 {
		return nil, status.Error(codes.InvalidArgument, "资源ID无效")
	}

	// 调用服务层
	err = s.rbacService.DeleteResource(ctx, req.Id)
	if err != nil {
		return nil, status.Error(codes.Internal, "删除资源失败: "+err.Error())
	}

	// 返回结果
	return &permissionpb.DeleteResourceResponse{
		Success: true,
	}, nil
}

// ListResources 获取资源列表
func (s *RBACServer) ListResources(ctx context.Context, req *permissionpb.ListResourcesRequest) (*permissionpb.ListResourcesResponse, error) {
	// 获取调用者ID
	_, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// 参数校验
	if req.BizId <= 0 {
		return nil, status.Error(codes.InvalidArgument, "业务ID无效")
	}

	// 调用服务层
	resources, total, err := s.rbacService.ListResources(ctx, req.BizId, int(req.Offset), int(req.Limit), req.Type, req.Key)
	if err != nil {
		return nil, status.Error(codes.Internal, "获取资源列表失败: "+err.Error())
	}

	// 转换结果
	protoResources := make([]*permissionpb.Resource, 0, len(resources))
	for _, resource := range resources {
		protoResources = append(protoResources, domainToProtoResource(resource))
	}

	// 返回结果
	return &permissionpb.ListResourcesResponse{
		Resources: protoResources,
		Total:     int32(total),
	}, nil
}

// ===== Permission相关接口实现 =====

// CreatePermission 创建权限
func (s *RBACServer) CreatePermission(ctx context.Context, req *permissionpb.CreatePermissionRequest) (*permissionpb.CreatePermissionResponse, error) {
	// 获取调用者ID
	_, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// 参数校验
	if req.Permission == nil {
		return nil, status.Error(codes.InvalidArgument, "缺少权限信息")
	}

	// 转换领域模型
	domainPermission := protoToDomainPermission(req.Permission)

	// 调用服务层
	createdPermission, err := s.rbacService.CreatePermission(ctx, domainPermission)
	if err != nil {
		return nil, status.Error(codes.Internal, "创建权限失败: "+err.Error())
	}

	// 返回结果
	return &permissionpb.CreatePermissionResponse{
		Permission: domainToProtoPermission(createdPermission),
	}, nil
}

// GetPermission 获取权限
func (s *RBACServer) GetPermission(ctx context.Context, req *permissionpb.GetPermissionRequest) (*permissionpb.GetPermissionResponse, error) {
	// 获取调用者ID
	_, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// 参数校验
	if req.Id <= 0 {
		return nil, status.Error(codes.InvalidArgument, "权限ID无效")
	}

	// 调用服务层
	permission, err := s.rbacService.GetPermission(ctx, req.Id)
	if err != nil {
		return nil, status.Error(codes.Internal, "获取权限失败: "+err.Error())
	}

	// 返回结果
	return &permissionpb.GetPermissionResponse{
		Permission: domainToProtoPermission(permission),
	}, nil
}

// UpdatePermission 更新权限
func (s *RBACServer) UpdatePermission(ctx context.Context, req *permissionpb.UpdatePermissionRequest) (*permissionpb.UpdatePermissionResponse, error) {
	// 获取调用者ID
	_, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// 参数校验
	if req.Permission == nil || req.Permission.Id <= 0 {
		return nil, status.Error(codes.InvalidArgument, "权限ID无效")
	}

	// 转换领域模型
	domainPermission := protoToDomainPermission(req.Permission)

	// 调用服务层
	updatedPermission, err := s.rbacService.UpdatePermission(ctx, domainPermission)
	if err != nil {
		return nil, status.Error(codes.Internal, "更新权限失败: "+err.Error())
	}

	// 返回结果
	return &permissionpb.UpdatePermissionResponse{
		Permission: domainToProtoPermission(updatedPermission),
	}, nil
}

// DeletePermission 删除权限
func (s *RBACServer) DeletePermission(ctx context.Context, req *permissionpb.DeletePermissionRequest) (*permissionpb.DeletePermissionResponse, error) {
	// 获取调用者ID
	_, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// 参数校验
	if req.Id <= 0 {
		return nil, status.Error(codes.InvalidArgument, "权限ID无效")
	}

	// 调用服务层
	err = s.rbacService.DeletePermission(ctx, req.Id)
	if err != nil {
		return nil, status.Error(codes.Internal, "删除权限失败: "+err.Error())
	}

	// 返回结果
	return &permissionpb.DeletePermissionResponse{
		Success: true,
	}, nil
}

// ListPermissions 获取权限列表
func (s *RBACServer) ListPermissions(ctx context.Context, req *permissionpb.ListPermissionsRequest) (*permissionpb.ListPermissionsResponse, error) {
	// 获取调用者ID
	_, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// 参数校验
	if req.BizId <= 0 {
		return nil, status.Error(codes.InvalidArgument, "业务ID无效")
	}

	// 调用服务层
	permissions, total, err := s.rbacService.ListPermissions(ctx, req.BizId, int(req.Offset), int(req.Limit), req.ResourceType, req.ResourceKey, req.Action)
	if err != nil {
		return nil, status.Error(codes.Internal, "获取权限列表失败: "+err.Error())
	}

	// 转换结果
	protoPermissions := make([]*permissionpb.Permission, 0, len(permissions))
	for _, permission := range permissions {
		protoPermissions = append(protoPermissions, domainToProtoPermission(permission))
	}

	// 返回结果
	return &permissionpb.ListPermissionsResponse{
		Permissions: protoPermissions,
		Total:       int32(total),
	}, nil
}

// ===== UserRole相关接口实现 =====

// GrantUserRole 授予用户角色
func (s *RBACServer) GrantUserRole(ctx context.Context, req *permissionpb.GrantUserRoleRequest) (*permissionpb.GrantUserRoleResponse, error) {
	// 获取调用者ID
	_, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// 参数校验
	if req.BizId <= 0 || req.UserId <= 0 || req.RoleId <= 0 {
		return nil, status.Error(codes.InvalidArgument, "业务ID、用户ID或角色ID无效")
	}

	// 调用服务层
	userRole, err := s.rbacService.GrantUserRole(ctx, req.BizId, req.UserId, req.RoleId, req.StartTime, req.EndTime)
	if err != nil {
		return nil, status.Error(codes.Internal, "授予用户角色失败: "+err.Error())
	}

	// 返回结果
	return &permissionpb.GrantUserRoleResponse{
		UserRole: domainToProtoUserRole(userRole),
	}, nil
}

// RevokeUserRole 撤销用户角色
func (s *RBACServer) RevokeUserRole(ctx context.Context, req *permissionpb.RevokeUserRoleRequest) (*permissionpb.RevokeUserRoleResponse, error) {
	// 获取调用者ID
	_, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// 参数校验
	if req.BizId <= 0 || req.UserId <= 0 || req.RoleId <= 0 {
		return nil, status.Error(codes.InvalidArgument, "业务ID、用户ID或角色ID无效")
	}

	// 调用服务层
	err = s.rbacService.RevokeUserRole(ctx, req.BizId, req.UserId, req.RoleId)
	if err != nil {
		return nil, status.Error(codes.Internal, "撤销用户角色失败: "+err.Error())
	}

	// 返回结果
	return &permissionpb.RevokeUserRoleResponse{
		Success: true,
	}, nil
}

// ListUserRoles 获取用户角色列表
func (s *RBACServer) ListUserRoles(ctx context.Context, req *permissionpb.ListUserRolesRequest) (*permissionpb.ListUserRolesResponse, error) {
	// 获取调用者ID
	_, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// 参数校验
	if req.BizId <= 0 || req.UserId <= 0 {
		return nil, status.Error(codes.InvalidArgument, "业务ID或用户ID无效")
	}

	// 调用服务层
	userRoles, total, err := s.rbacService.ListUserRoles(ctx, req.BizId, req.UserId, int(req.Offset), int(req.Limit))
	if err != nil {
		return nil, status.Error(codes.Internal, "获取用户角色列表失败: "+err.Error())
	}

	// 转换结果
	protoUserRoles := make([]*permissionpb.UserRole, 0, len(userRoles))
	for _, userRole := range userRoles {
		protoUserRoles = append(protoUserRoles, domainToProtoUserRole(userRole))
	}

	// 返回结果
	return &permissionpb.ListUserRolesResponse{
		UserRoles: protoUserRoles,
		Total:     int32(total),
	}, nil
}
