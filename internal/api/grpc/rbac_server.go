package grpc

import (
	"context"
	"encoding/json"
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
	rbacService rbac.Service
}

// NewRBACServer 创建RBAC服务器实例
func NewRBACServer(rbacService rbac.Service) *RBACServer {
	return &RBACServer{
		rbacService: rbacService,
	}
}

// 从gRPC上下文中获取调用者ID
func (s *RBACServer) getUserIDFromContext(ctx context.Context) (int64, error) {
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

// 从gRPC上下文中获取业务ID
func (s *RBACServer) getBizIDFromContext(ctx context.Context) (int64, error) {
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

// ==== 业务配置相关方法 ====

func (s *RBACServer) CreateBusinessConfig(ctx context.Context, req *permissionpb.CreateBusinessConfigRequest) (*permissionpb.CreateBusinessConfigResponse, error) {
	if req.Config == nil {
		return nil, status.Error(codes.InvalidArgument, "业务配置不能为空")
	}

	// 将proto中的业务配置转换为领域模型
	domainConfig := domain.BusinessConfig{
		ID:        req.Config.Id,
		OwnerID:   req.Config.OwnerId,
		OwnerType: req.Config.OwnerType,
		Name:      req.Config.Name,
		RateLimit: int(req.Config.RateLimit),
		Token:     req.Config.Token,
	}

	// 调用服务创建业务配置
	created, err := s.rbacService.CreateBusinessConfig(ctx, domainConfig)
	if err != nil {
		return nil, status.Error(codes.Internal, "创建业务配置失败: "+err.Error())
	}

	// 将领域模型转换回proto
	return &permissionpb.CreateBusinessConfigResponse{
		Config: &permissionpb.BusinessConfig{
			Id:        created.ID,
			OwnerId:   created.OwnerID,
			OwnerType: created.OwnerType,
			Name:      created.Name,
			RateLimit: int32(created.RateLimit),
			Token:     created.Token,
		},
	}, nil
}

func (s *RBACServer) GetBusinessConfig(ctx context.Context, req *permissionpb.GetBusinessConfigRequest) (*permissionpb.GetBusinessConfigResponse, error) {
	if req.Id <= 0 {
		return nil, status.Error(codes.InvalidArgument, "业务ID必须大于0")
	}

	// 调用服务获取业务配置
	config, err := s.rbacService.GetBusinessConfigByID(ctx, req.Id)
	if err != nil {
		return nil, status.Error(codes.Internal, "获取业务配置失败: "+err.Error())
	}

	// 将领域模型转换为proto响应
	return &permissionpb.GetBusinessConfigResponse{
		Config: &permissionpb.BusinessConfig{
			Id:        config.ID,
			OwnerId:   config.OwnerID,
			OwnerType: config.OwnerType,
			Name:      config.Name,
			RateLimit: int32(config.RateLimit),
			Token:     config.Token,
		},
	}, nil
}

func (s *RBACServer) UpdateBusinessConfig(ctx context.Context, req *permissionpb.UpdateBusinessConfigRequest) (*permissionpb.UpdateBusinessConfigResponse, error) {
	if req.Config == nil || req.Config.Id <= 0 {
		return nil, status.Error(codes.InvalidArgument, "业务配置不能为空且ID必须大于0")
	}

	// 将proto中的业务配置转换为领域模型
	domainConfig := domain.BusinessConfig{
		ID:        req.Config.Id,
		OwnerID:   req.Config.OwnerId,
		OwnerType: req.Config.OwnerType,
		Name:      req.Config.Name,
		RateLimit: int(req.Config.RateLimit),
		Token:     req.Config.Token,
	}

	// 调用服务更新业务配置
	_, err := s.rbacService.UpdateBusinessConfig(ctx, domainConfig)
	if err != nil {
		return nil, status.Error(codes.Internal, "更新业务配置失败: "+err.Error())
	}

	return &permissionpb.UpdateBusinessConfigResponse{
		Success: true,
	}, nil
}

func (s *RBACServer) DeleteBusinessConfig(ctx context.Context, req *permissionpb.DeleteBusinessConfigRequest) (*permissionpb.DeleteBusinessConfigResponse, error) {
	if req.Id <= 0 {
		return nil, status.Error(codes.InvalidArgument, "业务ID必须大于0")
	}

	// 调用服务删除业务配置
	err := s.rbacService.DeleteBusinessConfigByID(ctx, req.Id)
	if err != nil {
		return nil, status.Error(codes.Internal, "删除业务配置失败: "+err.Error())
	}

	return &permissionpb.DeleteBusinessConfigResponse{
		Success: true,
	}, nil
}

func (s *RBACServer) ListBusinessConfigs(ctx context.Context, req *permissionpb.ListBusinessConfigsRequest) (*permissionpb.ListBusinessConfigsResponse, error) {
	// 参数校验 - 设置默认分页
	offset := int(req.Offset)
	limit := int(req.Limit)
	if limit <= 0 {
		limit = 10 // 默认每页10条
	}

	// 调用服务获取业务配置列表
	configs, total, err := s.rbacService.ListBusinessConfigs(ctx, offset, limit)
	if err != nil {
		return nil, status.Error(codes.Internal, "获取业务配置列表失败: "+err.Error())
	}

	// 将领域模型列表转换为proto响应
	protoConfigs := make([]*permissionpb.BusinessConfig, 0, len(configs))
	for _, config := range configs {
		protoConfigs = append(protoConfigs, &permissionpb.BusinessConfig{
			Id:        config.ID,
			OwnerId:   config.OwnerID,
			OwnerType: config.OwnerType,
			Name:      config.Name,
			RateLimit: int32(config.RateLimit),
			Token:     config.Token,
		})
	}

	return &permissionpb.ListBusinessConfigsResponse{
		Configs: protoConfigs,
		Total:   int32(total),
	}, nil
}

// ==== 资源相关方法 ====

func (s *RBACServer) CreateResource(ctx context.Context, req *permissionpb.CreateResourceRequest) (*permissionpb.CreateResourceResponse, error) {
	if req.Resource == nil {
		return nil, status.Error(codes.InvalidArgument, "资源不能为空")
	}

	bizID, err := s.getBizIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// 将proto中的资源转换为领域模型
	var metadata domain.ResourceMetadata
	if req.Resource.Metadata != "" {
		// 解析元数据JSON
		if err := json.Unmarshal([]byte(req.Resource.Metadata), &metadata); err != nil {
			return nil, status.Error(codes.InvalidArgument, "资源元数据格式不正确: "+err.Error())
		}
	}

	domainResource := domain.Resource{
		BizID:       bizID,
		Type:        req.Resource.Type,
		Key:         req.Resource.Key,
		Name:        req.Resource.Name,
		Description: req.Resource.Description,
		Metadata:    metadata,
	}

	// 调用服务创建资源
	created, err := s.rbacService.CreateResource(ctx, domainResource)
	if err != nil {
		return nil, status.Error(codes.Internal, "创建资源失败: "+err.Error())
	}

	// 将领域模型转换回proto
	return &permissionpb.CreateResourceResponse{
		Resource: &permissionpb.Resource{
			Id:          created.ID,
			BizId:       created.BizID,
			Type:        created.Type,
			Key:         created.Key,
			Name:        created.Name,
			Description: created.Description,
			Metadata:    created.Metadata.String(),
		},
	}, nil
}

func (s *RBACServer) GetResource(ctx context.Context, req *permissionpb.GetResourceRequest) (*permissionpb.GetResourceResponse, error) {
	if req.Id <= 0 {
		return nil, status.Error(codes.InvalidArgument, "资源ID必须大于0")
	}

	bizID, err := s.getBizIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// 调用服务获取资源
	resource, err := s.rbacService.GetResource(ctx, bizID, req.Id)
	if err != nil {
		return nil, status.Error(codes.Internal, "获取资源失败: "+err.Error())
	}

	// 将领域模型转换为proto响应
	return &permissionpb.GetResourceResponse{
		Resource: &permissionpb.Resource{
			Id:          resource.ID,
			BizId:       resource.BizID,
			Type:        resource.Type,
			Key:         resource.Key,
			Name:        resource.Name,
			Description: resource.Description,
			Metadata:    resource.Metadata.String(),
		},
	}, nil
}

func (s *RBACServer) UpdateResource(ctx context.Context, req *permissionpb.UpdateResourceRequest) (*permissionpb.UpdateResourceResponse, error) {
	if req.Resource == nil || req.Resource.Id <= 0 {
		return nil, status.Error(codes.InvalidArgument, "资源不能为空且ID必须大于0")
	}

	bizID, err := s.getBizIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// 将proto中的资源转换为领域模型
	var metadata domain.ResourceMetadata
	if req.Resource.Metadata != "" {
		// 解析元数据JSON
		if err := json.Unmarshal([]byte(req.Resource.Metadata), &metadata); err != nil {
			return nil, status.Error(codes.InvalidArgument, "资源元数据格式不正确: "+err.Error())
		}
	}

	domainResource := domain.Resource{
		ID:          req.Resource.Id,
		BizID:       bizID,
		Type:        req.Resource.Type,
		Key:         req.Resource.Key,
		Name:        req.Resource.Name,
		Description: req.Resource.Description,
		Metadata:    metadata,
	}

	// 调用服务更新资源
	_, err = s.rbacService.UpdateResource(ctx, domainResource)
	if err != nil {
		return nil, status.Error(codes.Internal, "更新资源失败: "+err.Error())
	}

	return &permissionpb.UpdateResourceResponse{
		Success: true,
	}, nil
}

func (s *RBACServer) DeleteResource(ctx context.Context, req *permissionpb.DeleteResourceRequest) (*permissionpb.DeleteResourceResponse, error) {
	if req.Id <= 0 {
		return nil, status.Error(codes.InvalidArgument, "资源ID必须大于0")
	}

	bizID, err := s.getBizIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// 调用服务删除资源
	err = s.rbacService.DeleteResource(ctx, bizID, req.Id)
	if err != nil {
		return nil, status.Error(codes.Internal, "删除资源失败: "+err.Error())
	}

	return &permissionpb.DeleteResourceResponse{
		Success: true,
	}, nil
}

func (s *RBACServer) ListResources(ctx context.Context, req *permissionpb.ListResourcesRequest) (*permissionpb.ListResourcesResponse, error) {
	// 参数校验
	if req.BizId <= 0 {
		return nil, status.Error(codes.InvalidArgument, "业务ID必须大于0")
	}

	offset := int(req.Offset)
	limit := int(req.Limit)
	if limit <= 0 {
		limit = 10 // 默认每页10条
	}

	// 调用服务获取资源列表
	resources, total, err := s.rbacService.ListResources(ctx, req.BizId, "", "", offset, limit)
	if err != nil {
		return nil, status.Error(codes.Internal, "获取资源列表失败: "+err.Error())
	}

	// 将领域模型列表转换为proto响应
	protoResources := make([]*permissionpb.Resource, 0, len(resources))
	for _, resource := range resources {
		protoResources = append(protoResources, &permissionpb.Resource{
			Id:          resource.ID,
			BizId:       resource.BizID,
			Type:        resource.Type,
			Key:         resource.Key,
			Name:        resource.Name,
			Description: resource.Description,
			Metadata:    resource.Metadata.String(),
		})
	}

	return &permissionpb.ListResourcesResponse{
		Resources: protoResources,
		Total:     int32(total),
	}, nil
}

// ==== 权限相关方法 ====

func (s *RBACServer) CreatePermission(ctx context.Context, req *permissionpb.CreatePermissionRequest) (*permissionpb.CreatePermissionResponse, error) {
	if req.Permission == nil {
		return nil, status.Error(codes.InvalidArgument, "权限不能为空")
	}

	bizID, err := s.getBizIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// 处理Metadata
	var metadata domain.PermissionMetadata
	if req.Permission.Metadata != "" {
		// 解析元数据JSON
		if err := json.Unmarshal([]byte(req.Permission.Metadata), &metadata); err != nil {
			return nil, status.Error(codes.InvalidArgument, "权限元数据格式不正确: "+err.Error())
		}
	}

	// 构建domain层的Permission对象
	action := ""
	if len(req.Permission.Actions) > 0 {
		// 使用第一个action作为主要action，其他的可以在后续的实现中处理
		action = req.Permission.Actions[0]
	}

	domainPermission := domain.Permission{
		BizID:       bizID,
		Name:        req.Permission.Name,
		Description: req.Permission.Description,
		Resource: domain.Resource{
			ID:   req.Permission.ResourceId,
			Type: req.Permission.ResourceType,
			Key:  req.Permission.ResourceKey,
		},
		Action:   action,
		Metadata: metadata,
	}

	// 调用服务创建权限
	created, err := s.rbacService.CreatePermission(ctx, domainPermission)
	if err != nil {
		return nil, status.Error(codes.Internal, "创建权限失败: "+err.Error())
	}

	// 转换回proto
	actions := []string{}
	if created.Action != "" {
		actions = append(actions, created.Action)
	}

	return &permissionpb.CreatePermissionResponse{
		Permission: &permissionpb.Permission{
			Id:           created.ID,
			BizId:        created.BizID,
			Name:         created.Name,
			Description:  created.Description,
			ResourceId:   created.Resource.ID,
			ResourceType: created.Resource.Type,
			ResourceKey:  created.Resource.Key,
			Actions:      actions,
			Metadata:     created.Metadata.String(),
		},
	}, nil
}

func (s *RBACServer) GetPermission(ctx context.Context, req *permissionpb.GetPermissionRequest) (*permissionpb.GetPermissionResponse, error) {
	if req.Id <= 0 {
		return nil, status.Error(codes.InvalidArgument, "权限ID必须大于0")
	}

	bizID, err := s.getBizIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// 调用服务获取权限
	permission, err := s.rbacService.GetPermission(ctx, bizID, req.Id)
	if err != nil {
		return nil, status.Error(codes.Internal, "获取权限失败: "+err.Error())
	}

	// 转换为proto
	actions := []string{}
	if permission.Action != "" {
		actions = append(actions, permission.Action)
	}

	return &permissionpb.GetPermissionResponse{
		Permission: &permissionpb.Permission{
			Id:           permission.ID,
			BizId:        permission.BizID,
			Name:         permission.Name,
			Description:  permission.Description,
			ResourceId:   permission.Resource.ID,
			ResourceType: permission.Resource.Type,
			ResourceKey:  permission.Resource.Key,
			Actions:      actions,
			Metadata:     permission.Metadata.String(),
		},
	}, nil
}

func (s *RBACServer) UpdatePermission(ctx context.Context, req *permissionpb.UpdatePermissionRequest) (*permissionpb.UpdatePermissionResponse, error) {
	if req.Permission == nil || req.Permission.Id <= 0 {
		return nil, status.Error(codes.InvalidArgument, "权限不能为空且ID必须大于0")
	}

	bizID, err := s.getBizIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// 处理Metadata
	var metadata domain.PermissionMetadata
	if req.Permission.Metadata != "" {
		// 解析元数据JSON
		if err := json.Unmarshal([]byte(req.Permission.Metadata), &metadata); err != nil {
			return nil, status.Error(codes.InvalidArgument, "权限元数据格式不正确: "+err.Error())
		}
	}

	// 构建domain层的Permission对象
	action := ""
	if len(req.Permission.Actions) > 0 {
		// 使用第一个action作为主要action
		action = req.Permission.Actions[0]
	}

	domainPermission := domain.Permission{
		ID:          req.Permission.Id,
		BizID:       bizID,
		Name:        req.Permission.Name,
		Description: req.Permission.Description,
		Resource: domain.Resource{
			ID:   req.Permission.ResourceId,
			Type: req.Permission.ResourceType,
			Key:  req.Permission.ResourceKey,
		},
		Action:   action,
		Metadata: metadata,
	}

	// 调用服务更新权限
	_, err = s.rbacService.UpdatePermission(ctx, domainPermission)
	if err != nil {
		return nil, status.Error(codes.Internal, "更新权限失败: "+err.Error())
	}

	return &permissionpb.UpdatePermissionResponse{
		Success: true,
	}, nil
}

func (s *RBACServer) DeletePermission(ctx context.Context, req *permissionpb.DeletePermissionRequest) (*permissionpb.DeletePermissionResponse, error) {
	if req.Id <= 0 {
		return nil, status.Error(codes.InvalidArgument, "权限ID必须大于0")
	}

	bizID, err := s.getBizIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// 调用服务删除权限
	err = s.rbacService.DeletePermission(ctx, bizID, req.Id)
	if err != nil {
		return nil, status.Error(codes.Internal, "删除权限失败: "+err.Error())
	}

	return &permissionpb.DeletePermissionResponse{
		Success: true,
	}, nil
}

func (s *RBACServer) ListPermissions(ctx context.Context, req *permissionpb.ListPermissionsRequest) (*permissionpb.ListPermissionsResponse, error) {
	// 参数校验
	if req.BizId <= 0 {
		return nil, status.Error(codes.InvalidArgument, "业务ID必须大于0")
	}

	offset := int(req.Offset)
	limit := int(req.Limit)
	if limit <= 0 {
		limit = 10 // 默认每页10条
	}

	// 调用服务获取权限列表
	permissions, total, err := s.rbacService.ListPermissions(ctx, req.BizId, "", "", "", offset, limit)
	if err != nil {
		return nil, status.Error(codes.Internal, "获取权限列表失败: "+err.Error())
	}

	// 将领域模型列表转换为proto响应
	protoPermissions := make([]*permissionpb.Permission, 0, len(permissions))
	for _, perm := range permissions {
		actions := []string{}
		if perm.Action != "" {
			actions = append(actions, perm.Action)
		}

		protoPermissions = append(protoPermissions, &permissionpb.Permission{
			Id:           perm.ID,
			BizId:        perm.BizID,
			Name:         perm.Name,
			Description:  perm.Description,
			ResourceId:   perm.Resource.ID,
			ResourceType: perm.Resource.Type,
			ResourceKey:  perm.Resource.Key,
			Actions:      actions,
			Metadata:     perm.Metadata.String(),
		})
	}

	return &permissionpb.ListPermissionsResponse{
		Permissions: protoPermissions,
		Total:       int32(total),
	}, nil
}

// ==== 角色相关接口实现 ====

// 实现剩余的RBAC服务接口：
// CreateRole, GetRole, UpdateRole, DeleteRole, ListRoles
// CreateRoleInclusion, GetRoleInclusion, DeleteRoleInclusion, ListRoleInclusions
// GrantRolePermission, RevokeRolePermission, ListRolePermissions
// GrantUserRole, RevokeUserRole, ListUserRoles
// GrantUserPermission, RevokeUserPermission, ListUserPermissions

// 为简洁起见，下面是CreateRole方法的实现示例，其他方法也应该类似实现
func (s *RBACServer) CreateRole(ctx context.Context, req *permissionpb.CreateRoleRequest) (*permissionpb.CreateRoleResponse, error) {
	if req.Role == nil {
		return nil, status.Error(codes.InvalidArgument, "角色不能为空")
	}

	bizID, err := s.getBizIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// 处理Metadata
	var metadata domain.RoleMetadata
	if req.Role.Metadata != "" {
		// 解析元数据JSON
		if err := json.Unmarshal([]byte(req.Role.Metadata), &metadata); err != nil {
			return nil, status.Error(codes.InvalidArgument, "角色元数据格式不正确: "+err.Error())
		}
	}

	// 构建domain层的Role对象
	domainRole := domain.Role{
		BizID:       bizID,
		Type:        req.Role.Type,
		Name:        req.Role.Name,
		Description: req.Role.Description,
		Metadata:    metadata,
	}

	// 调用服务创建角色
	created, err := s.rbacService.CreateRole(ctx, domainRole)
	if err != nil {
		return nil, status.Error(codes.Internal, "创建角色失败: "+err.Error())
	}

	// 转换回proto
	return &permissionpb.CreateRoleResponse{
		Role: &permissionpb.Role{
			Id:          created.ID,
			BizId:       created.BizID,
			Type:        created.Type,
			Name:        created.Name,
			Description: created.Description,
			Metadata:    created.Metadata.String(),
		},
	}, nil
}

// GetRole 获取角色信息
func (s *RBACServer) GetRole(ctx context.Context, req *permissionpb.GetRoleRequest) (*permissionpb.GetRoleResponse, error) {
	if req.Id <= 0 {
		return nil, status.Error(codes.InvalidArgument, "角色ID必须大于0")
	}

	bizID, err := s.getBizIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// 调用服务获取角色信息
	role, err := s.rbacService.GetRole(ctx, bizID, req.Id)
	if err != nil {
		return nil, status.Error(codes.Internal, "获取角色失败: "+err.Error())
	}

	// 转换为proto响应
	return &permissionpb.GetRoleResponse{
		Role: &permissionpb.Role{
			Id:          role.ID,
			BizId:       role.BizID,
			Type:        role.Type,
			Name:        role.Name,
			Description: role.Description,
			Metadata:    role.Metadata.String(),
		},
	}, nil
}

// UpdateRole 更新角色信息
func (s *RBACServer) UpdateRole(ctx context.Context, req *permissionpb.UpdateRoleRequest) (*permissionpb.UpdateRoleResponse, error) {
	if req.Role == nil || req.Role.Id <= 0 {
		return nil, status.Error(codes.InvalidArgument, "角色不能为空且ID必须大于0")
	}

	bizID, err := s.getBizIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// 处理Metadata
	var metadata domain.RoleMetadata
	if req.Role.Metadata != "" {
		// 解析元数据JSON
		if err := json.Unmarshal([]byte(req.Role.Metadata), &metadata); err != nil {
			return nil, status.Error(codes.InvalidArgument, "角色元数据格式不正确: "+err.Error())
		}
	}

	// 构建domain层的Role对象
	domainRole := domain.Role{
		ID:          req.Role.Id,
		BizID:       bizID,
		Type:        req.Role.Type,
		Name:        req.Role.Name,
		Description: req.Role.Description,
		Metadata:    metadata,
	}

	// 调用服务更新角色
	_, err = s.rbacService.UpdateRole(ctx, domainRole)
	if err != nil {
		return nil, status.Error(codes.Internal, "更新角色失败: "+err.Error())
	}

	return &permissionpb.UpdateRoleResponse{
		Success: true,
	}, nil
}

// DeleteRole 删除角色
func (s *RBACServer) DeleteRole(ctx context.Context, req *permissionpb.DeleteRoleRequest) (*permissionpb.DeleteRoleResponse, error) {
	if req.Id <= 0 {
		return nil, status.Error(codes.InvalidArgument, "角色ID必须大于0")
	}

	bizID, err := s.getBizIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// 调用服务删除角色
	err = s.rbacService.DeleteRole(ctx, bizID, req.Id)
	if err != nil {
		return nil, status.Error(codes.Internal, "删除角色失败: "+err.Error())
	}

	return &permissionpb.DeleteRoleResponse{
		Success: true,
	}, nil
}

// ListRoles 获取角色列表
func (s *RBACServer) ListRoles(ctx context.Context, req *permissionpb.ListRolesRequest) (*permissionpb.ListRolesResponse, error) {
	// 参数校验
	if req.BizId <= 0 {
		return nil, status.Error(codes.InvalidArgument, "业务ID必须大于0")
	}

	offset := int(req.Offset)
	limit := int(req.Limit)
	if limit <= 0 {
		limit = 10 // 默认每页10条
	}

	// 调用服务获取角色列表
	roles, total, err := s.rbacService.ListRoles(ctx, req.BizId, "", offset, limit)
	if err != nil {
		return nil, status.Error(codes.Internal, "获取角色列表失败: "+err.Error())
	}

	// 将领域模型列表转换为proto响应
	protoRoles := make([]*permissionpb.Role, 0, len(roles))
	for _, role := range roles {
		protoRoles = append(protoRoles, &permissionpb.Role{
			Id:          role.ID,
			BizId:       role.BizID,
			Type:        role.Type,
			Name:        role.Name,
			Description: role.Description,
			Metadata:    role.Metadata.String(),
		})
	}

	return &permissionpb.ListRolesResponse{
		Roles: protoRoles,
		Total: int32(total),
	}, nil
}

// ==== 角色包含关系相关方法 ====

// CreateRoleInclusion 创建角色包含关系
func (s *RBACServer) CreateRoleInclusion(ctx context.Context, req *permissionpb.CreateRoleInclusionRequest) (*permissionpb.CreateRoleInclusionResponse, error) {
	if req.BizId <= 0 || req.IncludingRoleId <= 0 || req.IncludedRoleId <= 0 {
		return nil, status.Error(codes.InvalidArgument, "业务ID和角色ID必须大于0")
	}

	bizID, err := s.getBizIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// 构建domain层的RoleInclusion对象
	domainRoleInclusion := domain.RoleInclusion{
		BizID: bizID,
		IncludingRole: domain.Role{
			ID: req.IncludingRoleId,
		},
		IncludedRole: domain.Role{
			ID: req.IncludedRoleId,
		},
	}

	// 调用服务创建角色包含关系
	created, err := s.rbacService.CreateRoleInclusion(ctx, domainRoleInclusion)
	if err != nil {
		return nil, status.Error(codes.Internal, "创建角色包含关系失败: "+err.Error())
	}

	// 转换回proto
	return &permissionpb.CreateRoleInclusionResponse{
		RoleInclusion: &permissionpb.RoleInclusion{
			Id:              created.ID,
			IncludingRoleId: created.IncludingRole.ID,
			IncludedRoleId:  created.IncludedRole.ID,
		},
	}, nil
}

// GetRoleInclusion 获取角色包含关系
func (s *RBACServer) GetRoleInclusion(ctx context.Context, req *permissionpb.GetRoleInclusionRequest) (*permissionpb.GetRoleInclusionResponse, error) {
	if req.Id <= 0 {
		return nil, status.Error(codes.InvalidArgument, "角色包含关系ID必须大于0")
	}

	bizID, err := s.getBizIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// 调用服务获取角色包含关系
	roleInclusion, err := s.rbacService.GetRoleInclusion(ctx, bizID, req.Id)
	if err != nil {
		return nil, status.Error(codes.Internal, "获取角色包含关系失败: "+err.Error())
	}

	// 转换为proto响应
	return &permissionpb.GetRoleInclusionResponse{
		RoleInclusion: &permissionpb.RoleInclusion{
			Id:              roleInclusion.ID,
			IncludingRoleId: roleInclusion.IncludingRole.ID,
			IncludedRoleId:  roleInclusion.IncludedRole.ID,
		},
	}, nil
}

// DeleteRoleInclusion 删除角色包含关系
func (s *RBACServer) DeleteRoleInclusion(ctx context.Context, req *permissionpb.DeleteRoleInclusionRequest) (*permissionpb.DeleteRoleInclusionResponse, error) {
	if req.BizId <= 0 || req.IncludingRoleId <= 0 || req.IncludedRoleId <= 0 {
		return nil, status.Error(codes.InvalidArgument, "业务ID和角色ID必须大于0")
	}

	bizID, err := s.getBizIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// 由于接口中没有直接提供按角色ID删除的方法，我们需要先找到对应的inclusion
	// 这里可能需要添加新的方法到Service接口，或者直接提供角色ID到Repository层
	// 这里做一个简化处理：直接列出所有包含关系，然后找到匹配的删除
	inclusions, _, err := s.rbacService.ListRoleInclusions(ctx, bizID, req.IncludingRoleId, false, 0, 100)
	if err != nil {
		return nil, status.Error(codes.Internal, "查找角色包含关系失败: "+err.Error())
	}

	// 找到匹配的包含关系
	var inclusionID int64
	for _, inc := range inclusions {
		if inc.IncludingRole.ID == req.IncludingRoleId && inc.IncludedRole.ID == req.IncludedRoleId {
			inclusionID = inc.ID
			break
		}
	}

	if inclusionID == 0 {
		return nil, status.Error(codes.NotFound, "未找到匹配的角色包含关系")
	}

	// 调用服务删除角色包含关系
	err = s.rbacService.DeleteRoleInclusion(ctx, bizID, inclusionID)
	if err != nil {
		return nil, status.Error(codes.Internal, "删除角色包含关系失败: "+err.Error())
	}

	return &permissionpb.DeleteRoleInclusionResponse{
		Success: true,
	}, nil
}

// ListRoleInclusions 获取角色包含关系列表
func (s *RBACServer) ListRoleInclusions(ctx context.Context, req *permissionpb.ListRoleInclusionsRequest) (*permissionpb.ListRoleInclusionsResponse, error) {
	// 参数校验
	if req.BizId <= 0 {
		return nil, status.Error(codes.InvalidArgument, "业务ID必须大于0")
	}

	offset := int(req.Offset)
	limit := int(req.Limit)
	if limit <= 0 {
		limit = 10 // 默认每页10条
	}

	var includingRoleID int64
	var isIncluding bool
	if req.RoleId > 0 {
		includingRoleID = req.RoleId
		isIncluding = req.IsIncluding
	}

	// 调用服务获取角色包含关系列表
	roleInclusions, total, err := s.rbacService.ListRoleInclusions(ctx, req.BizId, includingRoleID, isIncluding, offset, limit)
	if err != nil {
		return nil, status.Error(codes.Internal, "获取角色包含关系列表失败: "+err.Error())
	}

	// 将领域模型列表转换为proto响应
	protoRoleInclusions := make([]*permissionpb.RoleInclusion, 0, len(roleInclusions))
	for _, inclusion := range roleInclusions {
		protoRoleInclusions = append(protoRoleInclusions, &permissionpb.RoleInclusion{
			Id:              inclusion.ID,
			BizId:           inclusion.BizID,
			IncludingRoleId: inclusion.IncludingRole.ID,
			IncludedRoleId:  inclusion.IncludedRole.ID,
			// 添加其他字段
		})
	}

	return &permissionpb.ListRoleInclusionsResponse{
		RoleInclusions: protoRoleInclusions,
		Total:          int32(total),
	}, nil
}

// ==== 角色权限相关方法 ====

// GrantRolePermission 授予角色权限
func (s *RBACServer) GrantRolePermission(ctx context.Context, req *permissionpb.GrantRolePermissionRequest) (*permissionpb.GrantRolePermissionResponse, error) {
	if req.BizId <= 0 || req.RoleId <= 0 || req.PermissionId <= 0 {
		return nil, status.Error(codes.InvalidArgument, "业务ID、角色ID和权限ID必须大于0")
	}

	bizID, err := s.getBizIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// 构建domain层的RolePermission对象
	domainRolePermission := domain.RolePermission{
		BizID: bizID,
		Role: domain.Role{
			ID: req.RoleId,
		},
		Permission: domain.Permission{
			ID: req.PermissionId,
		},
	}

	// 调用服务授予角色权限
	created, err := s.rbacService.GrantRolePermission(ctx, domainRolePermission)
	if err != nil {
		return nil, status.Error(codes.Internal, "授予角色权限失败: "+err.Error())
	}

	// 转换回proto
	return &permissionpb.GrantRolePermissionResponse{
		RolePermission: &permissionpb.RolePermission{
			Id:               created.ID,
			BizId:            created.BizID,
			RoleId:           created.Role.ID,
			PermissionId:     created.Permission.ID,
			RoleName:         created.Role.Name,
			RoleType:         created.Role.Type,
			ResourceType:     created.Permission.Resource.Type,
			ResourceKey:      created.Permission.Resource.Key,
			PermissionAction: created.Permission.Action,
		},
	}, nil
}

// RevokeRolePermission 撤销角色权限
func (s *RBACServer) RevokeRolePermission(ctx context.Context, req *permissionpb.RevokeRolePermissionRequest) (*permissionpb.RevokeRolePermissionResponse, error) {
	if req.BizId <= 0 || req.RoleId <= 0 || req.PermissionId <= 0 {
		return nil, status.Error(codes.InvalidArgument, "业务ID、角色ID和权限ID必须大于0")
	}

	bizID, err := s.getBizIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// 由于接口中没有直接提供按角色ID和权限ID撤销的方法，需要先找到对应关系ID
	// 与删除角色包含关系类似，这里做一个简化处理：
	// 获取角色的所有权限，找到匹配的权限ID，然后删除
	rolePermissions, _, err := s.rbacService.ListRolePermissions(ctx, bizID, req.RoleId, 0, 100)
	if err != nil {
		return nil, status.Error(codes.Internal, "查找角色权限关系失败: "+err.Error())
	}

	// 找到匹配的角色权限关系
	var rolePermissionID int64
	for _, rp := range rolePermissions {
		if rp.Role.ID == req.RoleId && rp.Permission.ID == req.PermissionId {
			rolePermissionID = rp.ID
			break
		}
	}

	if rolePermissionID == 0 {
		return nil, status.Error(codes.NotFound, "未找到匹配的角色权限关系")
	}

	// 调用服务撤销角色权限
	err = s.rbacService.RevokeRolePermission(ctx, bizID, rolePermissionID)
	if err != nil {
		return nil, status.Error(codes.Internal, "撤销角色权限失败: "+err.Error())
	}

	return &permissionpb.RevokeRolePermissionResponse{
		Success: true,
	}, nil
}

// ListRolePermissions 获取角色权限列表
func (s *RBACServer) ListRolePermissions(ctx context.Context, req *permissionpb.ListRolePermissionsRequest) (*permissionpb.ListRolePermissionsResponse, error) {
	if req.BizId <= 0 || req.RoleId <= 0 {
		return nil, status.Error(codes.InvalidArgument, "业务ID和角色ID必须大于0")
	}

	offset := int(req.Offset)
	limit := int(req.Limit)
	if limit <= 0 {
		limit = 10 // 默认每页10条
	}

	bizID, err := s.getBizIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// 调用服务获取角色权限列表
	rolePermissions, total, err := s.rbacService.ListRolePermissions(ctx, bizID, req.RoleId, offset, limit)
	if err != nil {
		return nil, status.Error(codes.Internal, "获取角色权限列表失败: "+err.Error())
	}

	// 将领域模型列表转换为proto响应
	protoRolePermissions := make([]*permissionpb.RolePermission, 0, len(rolePermissions))
	for _, rp := range rolePermissions {
		protoRolePermissions = append(protoRolePermissions, &permissionpb.RolePermission{
			Id:               rp.ID,
			BizId:            rp.BizID,
			RoleId:           rp.Role.ID,
			PermissionId:     rp.Permission.ID,
			RoleName:         rp.Role.Name,
			RoleType:         rp.Role.Type,
			ResourceType:     rp.Permission.Resource.Type,
			ResourceKey:      rp.Permission.Resource.Key,
			PermissionAction: rp.Permission.Action,
		})
	}

	return &permissionpb.ListRolePermissionsResponse{
		RolePermissions: protoRolePermissions,
		Total:           int32(total),
	}, nil
}

// ==== 用户角色相关方法 ====

// GrantUserRole 授予用户角色
func (s *RBACServer) GrantUserRole(ctx context.Context, req *permissionpb.GrantUserRoleRequest) (*permissionpb.GrantUserRoleResponse, error) {
	if req.BizId <= 0 || req.UserId <= 0 || req.RoleId <= 0 {
		return nil, status.Error(codes.InvalidArgument, "业务ID、用户ID和角色ID必须大于0")
	}

	bizID, err := s.getBizIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// 构建domain层的UserRole对象
	domainUserRole := domain.UserRole{
		BizID:  bizID,
		UserID: req.UserId,
		Role: domain.Role{
			ID: req.RoleId,
		},
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
	}

	// 调用服务授予用户角色
	created, err := s.rbacService.GrantUserRole(ctx, domainUserRole)
	if err != nil {
		return nil, status.Error(codes.Internal, "授予用户角色失败: "+err.Error())
	}

	// 转换回proto
	return &permissionpb.GrantUserRoleResponse{
		UserRole: &permissionpb.UserRole{
			Id:        created.ID,
			BizId:     created.BizID,
			UserId:    created.UserID,
			RoleId:    created.Role.ID,
			RoleName:  created.Role.Name,
			RoleType:  created.Role.Type,
			StartTime: created.StartTime,
			EndTime:   created.EndTime,
		},
	}, nil
}

// RevokeUserRole 撤销用户角色
func (s *RBACServer) RevokeUserRole(ctx context.Context, req *permissionpb.RevokeUserRoleRequest) (*permissionpb.RevokeUserRoleResponse, error) {
	if req.BizId <= 0 || req.UserId <= 0 || req.RoleId <= 0 {
		return nil, status.Error(codes.InvalidArgument, "业务ID、用户ID和角色ID必须大于0")
	}

	bizID, err := s.getBizIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// 由于接口中没有直接提供按用户ID和角色ID撤销的方法，需要先找到对应关系ID
	// 与删除角色权限关系类似，这里做一个简化处理：
	// 获取用户的所有角色，找到匹配的角色ID，然后删除
	userRoles, _, err := s.rbacService.ListUserRoles(ctx, bizID, req.UserId, 0, 100)
	if err != nil {
		return nil, status.Error(codes.Internal, "查找用户角色关系失败: "+err.Error())
	}

	// 找到匹配的用户角色关系
	var userRoleID int64
	for _, ur := range userRoles {
		if ur.UserID == req.UserId && ur.Role.ID == req.RoleId {
			userRoleID = ur.ID
			break
		}
	}

	if userRoleID == 0 {
		return nil, status.Error(codes.NotFound, "未找到匹配的用户角色关系")
	}

	// 调用服务撤销用户角色
	err = s.rbacService.RevokeUserRole(ctx, bizID, userRoleID)
	if err != nil {
		return nil, status.Error(codes.Internal, "撤销用户角色失败: "+err.Error())
	}

	return &permissionpb.RevokeUserRoleResponse{
		Success: true,
	}, nil
}

// ListUserRoles 获取用户角色列表
func (s *RBACServer) ListUserRoles(ctx context.Context, req *permissionpb.ListUserRolesRequest) (*permissionpb.ListUserRolesResponse, error) {
	if req.BizId <= 0 || req.UserId <= 0 {
		return nil, status.Error(codes.InvalidArgument, "业务ID和用户ID必须大于0")
	}

	offset := int(req.Offset)
	limit := int(req.Limit)
	if limit <= 0 {
		limit = 10 // 默认每页10条
	}

	bizID, err := s.getBizIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// 调用服务获取用户角色列表
	userRoles, total, err := s.rbacService.ListUserRoles(ctx, bizID, req.UserId, offset, limit)
	if err != nil {
		return nil, status.Error(codes.Internal, "获取用户角色列表失败: "+err.Error())
	}

	// 将领域模型列表转换为proto响应
	protoUserRoles := make([]*permissionpb.UserRole, 0, len(userRoles))
	for _, ur := range userRoles {
		protoUserRoles = append(protoUserRoles, &permissionpb.UserRole{
			Id:        ur.ID,
			BizId:     ur.BizID,
			UserId:    ur.UserID,
			RoleId:    ur.Role.ID,
			RoleName:  ur.Role.Name,
			RoleType:  ur.Role.Type,
			StartTime: ur.StartTime,
			EndTime:   ur.EndTime,
		})
	}

	return &permissionpb.ListUserRolesResponse{
		UserRoles: protoUserRoles,
		Total:     int32(total),
	}, nil
}

// ==== 用户权限相关方法 ====

// GrantUserPermission 授予用户权限
func (s *RBACServer) GrantUserPermission(ctx context.Context, req *permissionpb.GrantUserPermissionRequest) (*permissionpb.GrantUserPermissionResponse, error) {
	if req.BizId <= 0 || req.UserId <= 0 || req.PermissionId <= 0 {
		return nil, status.Error(codes.InvalidArgument, "业务ID、用户ID和权限ID必须大于0")
	}

	bizID, err := s.getBizIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// 确定权限效果
	var effect domain.Effect
	if req.Effect == "deny" {
		effect = domain.EffectDeny
	} else {
		effect = domain.EffectAllow // 默认为允许
	}

	// 构建domain层的UserPermission对象
	domainUserPermission := domain.UserPermission{
		BizID:  bizID,
		UserID: req.UserId,
		Permission: domain.Permission{
			ID: req.PermissionId,
		},
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
		Effect:    effect,
	}

	// 调用服务授予用户权限
	created, err := s.rbacService.GrantUserPermission(ctx, domainUserPermission)
	if err != nil {
		return nil, status.Error(codes.Internal, "授予用户权限失败: "+err.Error())
	}

	// 转换回proto
	return &permissionpb.GrantUserPermissionResponse{
		UserPermission: &permissionpb.UserPermission{
			Id:               created.ID,
			BizId:            created.BizID,
			UserId:           created.UserID,
			PermissionId:     created.Permission.ID,
			PermissionName:   created.Permission.Name,
			ResourceType:     created.Permission.Resource.Type,
			ResourceKey:      created.Permission.Resource.Key,
			ResourceName:     created.Permission.Resource.Name,
			PermissionAction: created.Permission.Action,
			StartTime:        created.StartTime,
			EndTime:          created.EndTime,
			Effect:           created.Effect.String(),
		},
	}, nil
}

// RevokeUserPermission 撤销用户权限
func (s *RBACServer) RevokeUserPermission(ctx context.Context, req *permissionpb.RevokeUserPermissionRequest) (*permissionpb.RevokeUserPermissionResponse, error) {
	if req.BizId <= 0 || req.UserId <= 0 || req.PermissionId <= 0 {
		return nil, status.Error(codes.InvalidArgument, "业务ID、用户ID和权限ID必须大于0")
	}

	bizID, err := s.getBizIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// 由于接口中没有直接提供按用户ID和权限ID撤销的方法，需要先找到对应关系ID
	// 与删除用户角色关系类似，这里做一个简化处理：
	// 获取用户的所有权限，找到匹配的权限ID，然后删除
	userPermissions, _, err := s.rbacService.ListUserPermissions(ctx, bizID, req.UserId, 0, 100)
	if err != nil {
		return nil, status.Error(codes.Internal, "查找用户权限关系失败: "+err.Error())
	}

	// 找到匹配的用户权限关系
	var userPermissionID int64
	for _, up := range userPermissions {
		if up.UserID == req.UserId && up.Permission.ID == req.PermissionId {
			userPermissionID = up.ID
			break
		}
	}

	if userPermissionID == 0 {
		return nil, status.Error(codes.NotFound, "未找到匹配的用户权限关系")
	}

	// 调用服务撤销用户权限
	err = s.rbacService.RevokeUserPermission(ctx, bizID, userPermissionID)
	if err != nil {
		return nil, status.Error(codes.Internal, "撤销用户权限失败: "+err.Error())
	}

	return &permissionpb.RevokeUserPermissionResponse{
		Success: true,
	}, nil
}

// ListUserPermissions 获取用户权限列表
func (s *RBACServer) ListUserPermissions(ctx context.Context, req *permissionpb.ListUserPermissionsRequest) (*permissionpb.ListUserPermissionsResponse, error) {
	if req.BizId <= 0 || req.UserId <= 0 {
		return nil, status.Error(codes.InvalidArgument, "业务ID和用户ID必须大于0")
	}

	offset := int(req.Offset)
	limit := int(req.Limit)
	if limit <= 0 {
		limit = 10 // 默认每页10条
	}

	bizID, err := s.getBizIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// 调用服务获取用户权限列表
	userPermissions, total, err := s.rbacService.ListUserPermissions(ctx, bizID, req.UserId, offset, limit)
	if err != nil {
		return nil, status.Error(codes.Internal, "获取用户权限列表失败: "+err.Error())
	}

	// 将领域模型列表转换为proto响应
	protoUserPermissions := make([]*permissionpb.UserPermission, 0, len(userPermissions))
	for _, up := range userPermissions {
		protoUserPermissions = append(protoUserPermissions, &permissionpb.UserPermission{
			Id:               up.ID,
			BizId:            up.BizID,
			UserId:           up.UserID,
			PermissionId:     up.Permission.ID,
			PermissionName:   up.Permission.Name,
			ResourceType:     up.Permission.Resource.Type,
			ResourceKey:      up.Permission.Resource.Key,
			ResourceName:     up.Permission.Resource.Name,
			PermissionAction: up.Permission.Action,
			StartTime:        up.StartTime,
			EndTime:          up.EndTime,
			Effect:           up.Effect.String(),
		})
	}

	return &permissionpb.ListUserPermissionsResponse{
		UserPermissions: protoUserPermissions,
		Total:           int32(total),
	}, nil
}
