package rbac

import (
	"context"
	"time"

	"github.com/ecodeclub/ekit/slice"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	permissionpb "gitee.com/flycash/permission-platform/api/proto/gen/permission/v1"
	"gitee.com/flycash/permission-platform/internal/domain"
	"gitee.com/flycash/permission-platform/internal/service/rbac"
)

type Server struct {
	permissionpb.UnimplementedRBACServiceServer
	baseServer
	rbacService rbac.Service
}

// NewServer 创建RBAC服务器实例
func NewServer(rbacService rbac.Service) *Server {
	return &Server{
		rbacService: rbacService,
	}
}

// ==== 业务配置相关方法 ====

func (s *Server) CreateBusinessConfig(ctx context.Context, req *permissionpb.CreateBusinessConfigRequest) (*permissionpb.CreateBusinessConfigResponse, error) {
	if req.Config == nil {
		return nil, status.Error(codes.InvalidArgument, "业务配置不能为空")
	}

	// 将proto中的业务配置转换为领域模型
	req.Config.Id = 0
	domainConfig := s.toBusinessConfigDomain(req.Config)

	// 调用服务创建业务配置
	created, err := s.rbacService.CreateBusinessConfig(ctx, domainConfig)
	if err != nil {
		return nil, status.Error(codes.Internal, "创建业务配置失败: "+err.Error())
	}

	// 将领域模型转换回proto
	return &permissionpb.CreateBusinessConfigResponse{
		Config: s.toBusinessConfigProto(created),
	}, nil
}

func (s *Server) toBusinessConfigDomain(req *permissionpb.BusinessConfig) domain.BusinessConfig {
	return domain.BusinessConfig{
		ID:        req.Id,
		OwnerID:   req.OwnerId,
		OwnerType: req.OwnerType,
		Name:      req.Name,
		RateLimit: int(req.RateLimit),
		Token:     req.Token,
	}
}

func (s *Server) toBusinessConfigProto(created domain.BusinessConfig) *permissionpb.BusinessConfig {
	return &permissionpb.BusinessConfig{
		Id:        created.ID,
		OwnerId:   created.OwnerID,
		OwnerType: created.OwnerType,
		Name:      created.Name,
		RateLimit: int32(created.RateLimit),
		Token:     created.Token,
	}
}

func (s *Server) GetBusinessConfig(ctx context.Context, req *permissionpb.GetBusinessConfigRequest) (*permissionpb.GetBusinessConfigResponse, error) {
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
		Config: s.toBusinessConfigProto(config),
	}, nil
}

func (s *Server) UpdateBusinessConfig(ctx context.Context, req *permissionpb.UpdateBusinessConfigRequest) (*permissionpb.UpdateBusinessConfigResponse, error) {
	if req.Config == nil || req.Config.Id <= 0 {
		return nil, status.Error(codes.InvalidArgument, "业务配置不能为空且ID必须大于0")
	}

	// 将proto中的业务配置转换为领域模型
	domainConfig := s.toBusinessConfigDomain(req.Config)

	// 调用服务更新业务配置
	_, err := s.rbacService.UpdateBusinessConfig(ctx, domainConfig)
	if err != nil {
		return nil, status.Error(codes.Internal, "更新业务配置失败: "+err.Error())
	}

	return &permissionpb.UpdateBusinessConfigResponse{
		Success: true,
	}, nil
}

func (s *Server) DeleteBusinessConfig(ctx context.Context, req *permissionpb.DeleteBusinessConfigRequest) (*permissionpb.DeleteBusinessConfigResponse, error) {
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

func (s *Server) ListBusinessConfigs(ctx context.Context, req *permissionpb.ListBusinessConfigsRequest) (*permissionpb.ListBusinessConfigsResponse, error) {
	// 参数校验 - 设置默认分页
	offset := int(req.Offset)
	limit := int(req.Limit)
	if limit <= 0 {
		limit = 10 // 默认每页10条
	}
	// 调用服务获取业务配置列表
	configs, err := s.rbacService.ListBusinessConfigs(ctx, offset, limit)
	if err != nil {
		return nil, status.Error(codes.Internal, "获取业务配置列表失败: "+err.Error())
	}
	return &permissionpb.ListBusinessConfigsResponse{
		Configs: slice.Map(configs, func(_ int, src domain.BusinessConfig) *permissionpb.BusinessConfig {
			return s.toBusinessConfigProto(src)
		}),
	}, nil
}

// ==== 资源相关方法 ====

func (s *Server) CreateResource(ctx context.Context, req *permissionpb.CreateResourceRequest) (*permissionpb.CreateResourceResponse, error) {
	if req.Resource == nil {
		return nil, status.Error(codes.InvalidArgument, "资源不能为空")
	}

	bizID, err := s.getBizIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	req.Resource.Id = 0
	req.Resource.BizId = bizID
	domainResource := s.toResourceDomain(req.Resource)

	// 调用服务创建资源
	created, err := s.rbacService.CreateResource(ctx, domainResource)
	if err != nil {
		return nil, status.Error(codes.Internal, "创建资源失败: "+err.Error())
	}

	// 将领域模型转换回proto
	return &permissionpb.CreateResourceResponse{
		Resource: s.toResourceProto(created),
	}, nil
}

func (s *Server) toResourceDomain(req *permissionpb.Resource) domain.Resource {
	var md string
	if req.Metadata != "" {
		md = req.Metadata
	}

	return domain.Resource{
		ID:          req.Id,
		BizID:       req.BizId,
		Type:        req.Type,
		Key:         req.Key,
		Name:        req.Name,
		Description: req.Description,
		Metadata:    md,
	}
}

func (s *Server) toResourceProto(created domain.Resource) *permissionpb.Resource {
	return &permissionpb.Resource{
		Id:          created.ID,
		BizId:       created.BizID,
		Type:        created.Type,
		Key:         created.Key,
		Name:        created.Name,
		Description: created.Description,
		Metadata:    created.Metadata,
	}
}

func (s *Server) GetResource(ctx context.Context, req *permissionpb.GetResourceRequest) (*permissionpb.GetResourceResponse, error) {
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
		Resource: s.toResourceProto(resource),
	}, nil
}

func (s *Server) UpdateResource(ctx context.Context, req *permissionpb.UpdateResourceRequest) (*permissionpb.UpdateResourceResponse, error) {
	if req.Resource == nil || req.Resource.Id <= 0 {
		return nil, status.Error(codes.InvalidArgument, "资源不能为空且ID必须大于0")
	}

	bizID, err := s.getBizIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// 将proto中的资源转换为领域模型
	req.Resource.BizId = bizID
	domainResource := s.toResourceDomain(req.Resource)

	// 调用服务更新资源
	_, err = s.rbacService.UpdateResource(ctx, domainResource)
	if err != nil {
		return nil, status.Error(codes.Internal, "更新资源失败: "+err.Error())
	}

	return &permissionpb.UpdateResourceResponse{
		Success: true,
	}, nil
}

func (s *Server) DeleteResource(ctx context.Context, req *permissionpb.DeleteResourceRequest) (*permissionpb.DeleteResourceResponse, error) {
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

func (s *Server) ListResources(ctx context.Context, req *permissionpb.ListResourcesRequest) (*permissionpb.ListResourcesResponse, error) {
	// 参数校验
	if req.BizId <= 0 {
		return nil, status.Error(codes.InvalidArgument, "业务ID必须大于0")
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

	// 调用服务获取资源列表
	resources, err := s.rbacService.ListResources(ctx, bizID, offset, limit)
	if err != nil {
		return nil, status.Error(codes.Internal, "获取资源列表失败: "+err.Error())
	}

	return &permissionpb.ListResourcesResponse{
		Resources: slice.Map(resources, func(_ int, src domain.Resource) *permissionpb.Resource {
			return s.toResourceProto(src)
		}),
	}, nil
}

// ==== 权限相关方法 ====

func (s *Server) CreatePermission(ctx context.Context, req *permissionpb.CreatePermissionRequest) (*permissionpb.CreatePermissionResponse, error) {
	if req.Permission == nil {
		return nil, status.Error(codes.InvalidArgument, "权限不能为空")
	}

	bizID, err := s.getBizIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// 构建domain层的Permission对象
	req.Permission.Id = 0
	req.Permission.BizId = bizID
	domainPermission := s.toPermissionDomain(req.Permission)

	// 调用服务创建权限
	created, err := s.rbacService.CreatePermission(ctx, domainPermission)
	if err != nil {
		return nil, status.Error(codes.Internal, "创建权限失败: "+err.Error())
	}

	// 转换回proto
	return &permissionpb.CreatePermissionResponse{
		Permission: s.toPermissionProto(created),
	}, nil
}

func (s *Server) toPermissionDomain(req *permissionpb.Permission) domain.Permission {
	// 处理Actions字段，proto是repeated字段，domain中是单个Action
	// 因目前领域模型只支持一个Action，所以暂时只取第一个
	var action string
	if len(req.Actions) > 0 {
		action = req.Actions[0]
	}

	// 处理Metadata字段
	var md string
	if req.Metadata != "" {
		md = req.Metadata
	}

	return domain.Permission{
		ID:          req.Id,
		BizID:       req.BizId,
		Name:        req.Name,
		Description: req.Description,
		Resource: domain.Resource{
			ID:   req.ResourceId,
			Type: req.ResourceType,
			Key:  req.ResourceKey,
		},
		Action:   action,
		Metadata: md,
	}
}

func (s *Server) toPermissionProto(created domain.Permission) *permissionpb.Permission {
	// 将domain的单个Action转为proto的actions数组
	var actions []string
	if created.Action != "" {
		actions = append(actions, created.Action)
	}

	return &permissionpb.Permission{
		Id:           created.ID,
		BizId:        created.BizID,
		Name:         created.Name,
		Description:  created.Description,
		ResourceId:   created.Resource.ID,
		ResourceType: created.Resource.Type,
		ResourceKey:  created.Resource.Key,
		Actions:      actions,
		Metadata:     created.Metadata,
	}
}

func (s *Server) GetPermission(ctx context.Context, req *permissionpb.GetPermissionRequest) (*permissionpb.GetPermissionResponse, error) {
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
	return &permissionpb.GetPermissionResponse{
		Permission: s.toPermissionProto(permission),
	}, nil
}

func (s *Server) UpdatePermission(ctx context.Context, req *permissionpb.UpdatePermissionRequest) (*permissionpb.UpdatePermissionResponse, error) {
	if req.Permission == nil || req.Permission.Id <= 0 {
		return nil, status.Error(codes.InvalidArgument, "权限不能为空且ID必须大于0")
	}

	bizID, err := s.getBizIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// 构建domain层的Permission对象
	req.Permission.BizId = bizID
	domainPermission := s.toPermissionDomain(req.Permission)

	// 调用服务更新权限
	_, err = s.rbacService.UpdatePermission(ctx, domainPermission)
	if err != nil {
		return nil, status.Error(codes.Internal, "更新权限失败: "+err.Error())
	}

	return &permissionpb.UpdatePermissionResponse{
		Success: true,
	}, nil
}

func (s *Server) DeletePermission(ctx context.Context, req *permissionpb.DeletePermissionRequest) (*permissionpb.DeletePermissionResponse, error) {
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

func (s *Server) ListPermissions(ctx context.Context, req *permissionpb.ListPermissionsRequest) (*permissionpb.ListPermissionsResponse, error) {
	// 参数校验
	if req.BizId <= 0 {
		return nil, status.Error(codes.InvalidArgument, "业务ID必须大于0")
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

	// 调用服务获取权限列表
	permissions, err := s.rbacService.ListPermissions(ctx, bizID, offset, limit)
	if err != nil {
		return nil, status.Error(codes.Internal, "获取权限列表失败: "+err.Error())
	}

	return &permissionpb.ListPermissionsResponse{
		Permissions: slice.Map(permissions, func(_ int, src domain.Permission) *permissionpb.Permission {
			return s.toPermissionProto(src)
		}),
	}, nil
}

// ==== 角色相关接口实现 ====

func (s *Server) CreateRole(ctx context.Context, req *permissionpb.CreateRoleRequest) (*permissionpb.CreateRoleResponse, error) {
	if req.Role == nil {
		return nil, status.Error(codes.InvalidArgument, "角色不能为空")
	}

	bizID, err := s.getBizIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// 构建domain层的Role对象
	req.Role.Id = 0
	req.Role.BizId = bizID
	domainRole := s.toRoleDomain(req.Role)

	// 调用服务创建角色
	created, err := s.rbacService.CreateRole(ctx, domainRole)
	if err != nil {
		return nil, status.Error(codes.Internal, "创建角色失败: "+err.Error())
	}

	// 转换回proto
	return &permissionpb.CreateRoleResponse{
		Role: s.toRoleProto(created),
	}, nil
}

func (s *Server) toRoleDomain(req *permissionpb.Role) domain.Role {
	var md string
	if req.Metadata != "" {
		md = req.Metadata
	}

	return domain.Role{
		ID:          req.Id,
		BizID:       req.BizId,
		Type:        req.Type,
		Name:        req.Name,
		Description: req.Description,
		Metadata:    md,
	}
}

func (s *Server) toRoleProto(created domain.Role) *permissionpb.Role {
	return &permissionpb.Role{
		Id:          created.ID,
		BizId:       created.BizID,
		Type:        created.Type,
		Name:        created.Name,
		Description: created.Description,
		Metadata:    created.Metadata,
	}
}

// GetRole 获取角色信息
func (s *Server) GetRole(ctx context.Context, req *permissionpb.GetRoleRequest) (*permissionpb.GetRoleResponse, error) {
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
		Role: s.toRoleProto(role),
	}, nil
}

// UpdateRole 更新角色信息
func (s *Server) UpdateRole(ctx context.Context, req *permissionpb.UpdateRoleRequest) (*permissionpb.UpdateRoleResponse, error) {
	if req.Role == nil || req.Role.Id <= 0 {
		return nil, status.Error(codes.InvalidArgument, "角色不能为空且ID必须大于0")
	}

	bizID, err := s.getBizIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// 构建domain层的Role对象
	req.Role.BizId = bizID
	domainRole := s.toRoleDomain(req.Role)

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
func (s *Server) DeleteRole(ctx context.Context, req *permissionpb.DeleteRoleRequest) (*permissionpb.DeleteRoleResponse, error) {
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
func (s *Server) ListRoles(ctx context.Context, req *permissionpb.ListRolesRequest) (*permissionpb.ListRolesResponse, error) {
	// 参数校验
	if req.BizId <= 0 {
		return nil, status.Error(codes.InvalidArgument, "业务ID必须大于0")
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

	// 调用服务获取角色列表
	roles, err := s.rbacService.ListRolesByRoleType(ctx, bizID, req.Type, offset, limit)
	if err != nil {
		return nil, status.Error(codes.Internal, "获取角色列表失败: "+err.Error())
	}

	return &permissionpb.ListRolesResponse{
		Roles: slice.Map(roles, func(_ int, src domain.Role) *permissionpb.Role {
			return s.toRoleProto(src)
		}),
	}, nil
}

// ==== 角色包含关系相关方法 ====

// CreateRoleInclusion 创建角色包含关系
func (s *Server) CreateRoleInclusion(ctx context.Context, req *permissionpb.CreateRoleInclusionRequest) (*permissionpb.CreateRoleInclusionResponse, error) {
	if req.RoleInclusion == nil {
		return nil, status.Error(codes.InvalidArgument, "角色包含关系不能为空")
	}

	bizID, err := s.getBizIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// 构建domain层的RoleInclusion对象
	req.RoleInclusion.Id = 0
	req.RoleInclusion.BizId = bizID
	domainRoleInclusion := s.toRoleInclusionDomain(req.RoleInclusion)

	// 调用服务创建角色包含关系
	created, err := s.rbacService.CreateRoleInclusion(ctx, domainRoleInclusion)
	if err != nil {
		return nil, status.Error(codes.Internal, "创建角色包含关系失败: "+err.Error())
	}

	// 转换回proto
	return &permissionpb.CreateRoleInclusionResponse{
		RoleInclusion: s.toRoleInclusionProto(created),
	}, nil
}

func (s *Server) toRoleInclusionDomain(ri *permissionpb.RoleInclusion) domain.RoleInclusion {
	return domain.RoleInclusion{
		ID:    ri.Id,
		BizID: ri.BizId,
		IncludingRole: domain.Role{
			ID:   ri.IncludingRoleId,
			Type: ri.IncludingRoleType,
			Name: ri.IncludingRoleName,
		},
		IncludedRole: domain.Role{
			ID:   ri.IncludedRoleId,
			Type: ri.IncludedRoleType,
			Name: ri.IncludedRoleName,
		},
	}
}

func (s *Server) toRoleInclusionProto(created domain.RoleInclusion) *permissionpb.RoleInclusion {
	return &permissionpb.RoleInclusion{
		Id:                created.ID,
		BizId:             created.BizID,
		IncludingRoleId:   created.IncludingRole.ID,
		IncludingRoleType: created.IncludingRole.Type,
		IncludingRoleName: created.IncludingRole.Name,
		IncludedRoleId:    created.IncludedRole.ID,
		IncludedRoleType:  created.IncludedRole.Type,
		IncludedRoleName:  created.IncludedRole.Name,
	}
}

// GetRoleInclusion 获取角色包含关系
func (s *Server) GetRoleInclusion(ctx context.Context, req *permissionpb.GetRoleInclusionRequest) (*permissionpb.GetRoleInclusionResponse, error) {
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
		RoleInclusion: s.toRoleInclusionProto(roleInclusion),
	}, nil
}

// DeleteRoleInclusion 删除角色包含关系
func (s *Server) DeleteRoleInclusion(ctx context.Context, req *permissionpb.DeleteRoleInclusionRequest) (*permissionpb.DeleteRoleInclusionResponse, error) {
	if req.Id <= 0 {
		return nil, status.Error(codes.InvalidArgument, "角色包含关系ID必须大于0")
	}

	bizID, err := s.getBizIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// 调用服务删除角色包含关系
	err = s.rbacService.DeleteRoleInclusion(ctx, bizID, req.Id)
	if err != nil {
		return nil, status.Error(codes.Internal, "删除角色包含关系失败: "+err.Error())
	}

	return &permissionpb.DeleteRoleInclusionResponse{
		Success: true,
	}, nil
}

// ListRoleInclusions 获取角色包含关系列表
func (s *Server) ListRoleInclusions(ctx context.Context, req *permissionpb.ListRoleInclusionsRequest) (*permissionpb.ListRoleInclusionsResponse, error) {
	// 参数校验
	if req.BizId <= 0 {
		return nil, status.Error(codes.InvalidArgument, "业务ID必须大于0")
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

	// 调用服务获取角色包含关系列表
	roleInclusions, err := s.rbacService.ListRoleInclusions(ctx, bizID, offset, limit)
	if err != nil {
		return nil, status.Error(codes.Internal, "获取角色包含关系列表失败: "+err.Error())
	}

	// 将领域模型列表转换为proto响应
	protoRoleInclusions := slice.Map(roleInclusions, func(_ int, src domain.RoleInclusion) *permissionpb.RoleInclusion {
		return s.toRoleInclusionProto(src)
	})

	return &permissionpb.ListRoleInclusionsResponse{
		RoleInclusions: protoRoleInclusions,
	}, nil
}

// ==== 角色权限相关方法 ====

// GrantRolePermission 授予角色权限
func (s *Server) GrantRolePermission(ctx context.Context, req *permissionpb.GrantRolePermissionRequest) (*permissionpb.GrantRolePermissionResponse, error) {
	if req.RolePermission == nil {
		return nil, status.Error(codes.InvalidArgument, "角色权限关系不能为空")
	}

	bizID, err := s.getBizIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// 构建domain层的RolePermission对象
	req.RolePermission.Id = 0
	req.RolePermission.BizId = bizID
	domainRolePermission := s.toRolePermissionDomain(req.RolePermission)

	// 调用服务授予角色权限
	created, err := s.rbacService.GrantRolePermission(ctx, domainRolePermission)
	if err != nil {
		return nil, status.Error(codes.Internal, "授予角色权限失败: "+err.Error())
	}

	// 转换回proto
	return &permissionpb.GrantRolePermissionResponse{
		RolePermission: s.toRolePermissionProto(created),
	}, nil
}

func (s *Server) toRolePermissionDomain(req *permissionpb.RolePermission) domain.RolePermission {
	return domain.RolePermission{
		ID:    req.Id,
		BizID: req.BizId,
		Role: domain.Role{
			ID:   req.RoleId,
			Name: req.RoleName,
			Type: req.RoleType,
		},
		Permission: domain.Permission{
			ID: req.PermissionId,
			Resource: domain.Resource{
				Type: req.ResourceType,
				Key:  req.ResourceKey,
			},
			Action: req.PermissionAction,
		},
	}
}

func (s *Server) toRolePermissionProto(created domain.RolePermission) *permissionpb.RolePermission {
	return &permissionpb.RolePermission{
		Id:               created.ID,
		BizId:            created.BizID,
		RoleId:           created.Role.ID,
		PermissionId:     created.Permission.ID,
		RoleName:         created.Role.Name,
		RoleType:         created.Role.Type,
		ResourceType:     created.Permission.Resource.Type,
		ResourceKey:      created.Permission.Resource.Key,
		PermissionAction: created.Permission.Action,
	}
}

// RevokeRolePermission 撤销角色权限
func (s *Server) RevokeRolePermission(ctx context.Context, req *permissionpb.RevokeRolePermissionRequest) (*permissionpb.RevokeRolePermissionResponse, error) {
	if req.Id <= 0 {
		return nil, status.Error(codes.InvalidArgument, "角色权限关系ID必须大于0")
	}

	bizID, err := s.getBizIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// 调用服务撤销角色权限
	err = s.rbacService.RevokeRolePermission(ctx, bizID, req.Id)
	if err != nil {
		return nil, status.Error(codes.Internal, "撤销角色权限失败: "+err.Error())
	}

	return &permissionpb.RevokeRolePermissionResponse{
		Success: true,
	}, nil
}

// ListRolePermissions 获取角色权限列表
func (s *Server) ListRolePermissions(ctx context.Context, req *permissionpb.ListRolePermissionsRequest) (*permissionpb.ListRolePermissionsResponse, error) {
	if req.BizId <= 0 {
		return nil, status.Error(codes.InvalidArgument, "业务ID必须大于0")
	}

	bizID, err := s.getBizIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// 调用服务获取角色权限列表
	rolePermissions, err := s.rbacService.ListRolePermissions(ctx, bizID)
	if err != nil {
		return nil, status.Error(codes.Internal, "获取角色权限列表失败: "+err.Error())
	}

	return &permissionpb.ListRolePermissionsResponse{
		RolePermissions: slice.Map(rolePermissions, func(_ int, src domain.RolePermission) *permissionpb.RolePermission {
			return s.toRolePermissionProto(src)
		}),
	}, nil
}

// ==== 用户角色相关方法 ====

// GrantUserRole 授予用户角色
func (s *Server) GrantUserRole(ctx context.Context, req *permissionpb.GrantUserRoleRequest) (*permissionpb.GrantUserRoleResponse, error) {
	if req.UserRole == nil {
		return nil, status.Error(codes.InvalidArgument, "用户角色关系不能为空")
	}

	bizID, err := s.getBizIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// 构建domain层的UserRole对象
	req.UserRole.Id = 0
	req.UserRole.BizId = bizID

	domainUserRole := s.toUserRoleDomain(req.UserRole)

	// 调用服务授予用户角色
	created, err := s.rbacService.GrantUserRole(ctx, domainUserRole)
	if err != nil {
		return nil, status.Error(codes.Internal, "授予用户角色失败: "+err.Error())
	}

	// 转换回proto
	return &permissionpb.GrantUserRoleResponse{
		UserRole: s.toUserRoleProto(created),
	}, nil
}

// UserRole 转换
func (s *Server) toUserRoleDomain(ur *permissionpb.UserRole) domain.UserRole {
	startTime := ur.StartTime
	endTime := ur.EndTime
	if startTime == 0 {
		startTime = time.Now().UnixMilli()
	}
	if endTime == 0 {
		endTime = time.Now().AddDate(100, 0, 0).UnixMilli()
	}
	return domain.UserRole{
		ID:     ur.Id,
		BizID:  ur.BizId,
		UserID: ur.UserId,
		Role: domain.Role{
			ID:   ur.RoleId,
			Name: ur.RoleName,
			Type: ur.RoleType,
		},
		StartTime: startTime,
		EndTime:   endTime,
	}
}

func (s *Server) toUserRoleProto(ur domain.UserRole) *permissionpb.UserRole {
	return &permissionpb.UserRole{
		Id:        ur.ID,
		BizId:     ur.BizID,
		UserId:    ur.UserID,
		RoleId:    ur.Role.ID,
		RoleName:  ur.Role.Name,
		RoleType:  ur.Role.Type,
		StartTime: ur.StartTime,
		EndTime:   ur.EndTime,
	}
}

// RevokeUserRole 撤销用户角色
func (s *Server) RevokeUserRole(ctx context.Context, req *permissionpb.RevokeUserRoleRequest) (*permissionpb.RevokeUserRoleResponse, error) {
	if req.Id <= 0 {
		return nil, status.Error(codes.InvalidArgument, "用户角色关系ID必须大于0")
	}

	bizID, err := s.getBizIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// 调用服务撤销用户角色
	err = s.rbacService.RevokeUserRole(ctx, bizID, req.Id)
	if err != nil {
		return nil, status.Error(codes.Internal, "撤销用户角色失败: "+err.Error())
	}

	return &permissionpb.RevokeUserRoleResponse{
		Success: true,
	}, nil
}

// ListUserRoles 获取用户角色列表
func (s *Server) ListUserRoles(ctx context.Context, req *permissionpb.ListUserRolesRequest) (*permissionpb.ListUserRolesResponse, error) {
	if req.BizId <= 0 {
		return nil, status.Error(codes.InvalidArgument, "业务ID必须大于0")
	}

	bizID, err := s.getBizIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// 调用服务获取用户角色列表
	userRoles, err := s.rbacService.ListUserRoles(ctx, bizID)
	if err != nil {
		return nil, status.Error(codes.Internal, "获取用户角色列表失败: "+err.Error())
	}

	// 将领域模型列表转换为proto响应
	return &permissionpb.ListUserRolesResponse{
		UserRoles: slice.Map(userRoles, func(_ int, src domain.UserRole) *permissionpb.UserRole {
			return s.toUserRoleProto(src)
		}),
	}, nil
}

// ==== 用户权限相关方法 ====

// GrantUserPermission 授予用户权限
func (s *Server) GrantUserPermission(ctx context.Context, req *permissionpb.GrantUserPermissionRequest) (*permissionpb.GrantUserPermissionResponse, error) {
	if req.UserPermission == nil {
		return nil, status.Error(codes.InvalidArgument, "用户权限关系不能为空")
	}

	bizID, err := s.getBizIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// 设置业务ID
	req.UserPermission.Id = 0
	req.UserPermission.BizId = bizID

	// 构建domain层的UserPermission对象
	domainUserPermission := s.toUserPermissionDomain(req.UserPermission)

	// 调用服务授予用户权限
	created, err := s.rbacService.GrantUserPermission(ctx, domainUserPermission)
	if err != nil {
		return nil, status.Error(codes.Internal, "授予用户权限失败: "+err.Error())
	}

	// 转换回proto
	return &permissionpb.GrantUserPermissionResponse{
		UserPermission: s.toUserPermissionProto(created),
	}, nil
}

func (s *Server) toUserPermissionDomain(up *permissionpb.UserPermission) domain.UserPermission {
	var effect domain.Effect
	if up.Effect == "deny" {
		effect = domain.EffectDeny
	} else {
		effect = domain.EffectAllow // 默认为允许
	}

	startTime := up.StartTime
	endTime := up.EndTime
	if startTime == 0 {
		startTime = time.Now().UnixMilli()
	}
	if endTime == 0 {
		endTime = time.Now().AddDate(100, 0, 0).UnixMilli()
	}

	return domain.UserPermission{
		ID:     up.Id,
		BizID:  up.BizId,
		UserID: up.UserId,
		Permission: domain.Permission{
			ID:   up.PermissionId,
			Name: up.PermissionName,
			Resource: domain.Resource{
				Type: up.ResourceType,
				Key:  up.ResourceKey,
			},
			Action: up.PermissionAction,
		},
		StartTime: startTime,
		EndTime:   endTime,
		Effect:    effect,
	}
}

func (s *Server) toUserPermissionProto(up domain.UserPermission) *permissionpb.UserPermission {
	return &permissionpb.UserPermission{
		Id:               up.ID,
		BizId:            up.BizID,
		UserId:           up.UserID,
		PermissionId:     up.Permission.ID,
		PermissionName:   up.Permission.Name,
		ResourceType:     up.Permission.Resource.Type,
		ResourceKey:      up.Permission.Resource.Key,
		PermissionAction: up.Permission.Action,
		StartTime:        up.StartTime,
		EndTime:          up.EndTime,
		Effect:           up.Effect.String(),
	}
}

// RevokeUserPermission 撤销用户权限
func (s *Server) RevokeUserPermission(ctx context.Context, req *permissionpb.RevokeUserPermissionRequest) (*permissionpb.RevokeUserPermissionResponse, error) {
	if req.Id <= 0 {
		return nil, status.Error(codes.InvalidArgument, "用户权限关系ID必须大于0")
	}

	bizID, err := s.getBizIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// 调用服务撤销用户权限
	err = s.rbacService.RevokeUserPermission(ctx, bizID, req.Id)
	if err != nil {
		return nil, status.Error(codes.Internal, "撤销用户权限失败: "+err.Error())
	}

	return &permissionpb.RevokeUserPermissionResponse{
		Success: true,
	}, nil
}

// ListUserPermissions 获取用户权限列表
func (s *Server) ListUserPermissions(ctx context.Context, req *permissionpb.ListUserPermissionsRequest) (*permissionpb.ListUserPermissionsResponse, error) {
	if req.BizId <= 0 {
		return nil, status.Error(codes.InvalidArgument, "业务ID必须大于0")
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
	userPermissions, err := s.rbacService.ListUserPermissions(ctx, bizID, offset, limit)
	if err != nil {
		return nil, status.Error(codes.Internal, "获取用户权限列表失败: "+err.Error())
	}

	// 将领域模型列表转换为proto响应
	return &permissionpb.ListUserPermissionsResponse{
		UserPermissions: slice.Map(userPermissions, func(_ int, src domain.UserPermission) *permissionpb.UserPermission {
			return s.toUserPermissionProto(src)
		}),
	}, nil
}
