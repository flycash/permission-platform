package rbac

import (
	"context"
	"time"

	"gitee.com/flycash/permission-platform/internal/api/grpc/interceptor/auth"
	"gitee.com/flycash/permission-platform/internal/domain"
	"gitee.com/flycash/permission-platform/internal/pkg/jwt"
	"gitee.com/flycash/permission-platform/internal/repository"
)

// Service RBAC模型的管理接口
type Service interface {
	// 业务配置相关方法

	CreateBusinessConfig(ctx context.Context, config domain.BusinessConfig) (domain.BusinessConfig, error)
	GetBusinessConfigByID(ctx context.Context, id int64) (domain.BusinessConfig, error)
	UpdateBusinessConfig(ctx context.Context, config domain.BusinessConfig) (domain.BusinessConfig, error)
	DeleteBusinessConfigByID(ctx context.Context, id int64) error
	ListBusinessConfigs(ctx context.Context, offset, limit int) ([]domain.BusinessConfig, error)

	// 资源相关方法

	CreateResource(ctx context.Context, resource domain.Resource) (domain.Resource, error)
	GetResource(ctx context.Context, bizID, id int64) (domain.Resource, error)
	UpdateResource(ctx context.Context, resource domain.Resource) (domain.Resource, error)
	DeleteResource(ctx context.Context, bizID, id int64) error
	ListResources(ctx context.Context, bizID int64, offset, limit int) ([]domain.Resource, error)
	// 权限相关方法

	CreatePermission(ctx context.Context, permission domain.Permission) (domain.Permission, error)
	GetPermission(ctx context.Context, bizID, id int64) (domain.Permission, error)
	UpdatePermission(ctx context.Context, permission domain.Permission) (domain.Permission, error)
	DeletePermission(ctx context.Context, bizID, id int64) error
	ListPermissions(ctx context.Context, bizID int64, offset, limit int) ([]domain.Permission, error)

	// 角色相关方法

	CreateRole(ctx context.Context, role domain.Role) (domain.Role, error)
	GetRole(ctx context.Context, bizID, id int64) (domain.Role, error)
	UpdateRole(ctx context.Context, role domain.Role) (domain.Role, error)
	DeleteRole(ctx context.Context, bizID, id int64) error
	ListRolesByRoleType(ctx context.Context, bizID int64, roleType string, offset, limit int) ([]domain.Role, error)
	ListRoles(ctx context.Context, bizID int64, offset, limit int) ([]domain.Role, error)

	// 角色包含关系相关方法

	CreateRoleInclusion(ctx context.Context, roleInclusion domain.RoleInclusion) (domain.RoleInclusion, error)
	GetRoleInclusion(ctx context.Context, bizID, id int64) (domain.RoleInclusion, error)
	DeleteRoleInclusion(ctx context.Context, bizID, id int64) error
	ListRoleInclusionsByRoleID(ctx context.Context, bizID, roleID int64, isIncluding bool) ([]domain.RoleInclusion, error)
	ListRoleInclusions(ctx context.Context, bizID int64, offset, limit int) ([]domain.RoleInclusion, error)

	// 角色权限相关方法

	GrantRolePermission(ctx context.Context, rolePermission domain.RolePermission) (domain.RolePermission, error)
	RevokeRolePermission(ctx context.Context, bizID, id int64) error
	ListRolePermissionsByRoleID(ctx context.Context, bizID, roleID int64) ([]domain.RolePermission, error)
	ListRolePermissions(ctx context.Context, bizID int64) ([]domain.RolePermission, error)

	// 用户角色相关方法

	GrantUserRole(ctx context.Context, userRole domain.UserRole) (domain.UserRole, error)
	RevokeUserRole(ctx context.Context, bizID, id int64) error
	ListUserRolesByUserID(ctx context.Context, bizID, userID int64) ([]domain.UserRole, error)
	ListUserRoles(ctx context.Context, bizID int64) ([]domain.UserRole, error)

	// 用户权限相关方法

	GrantUserPermission(ctx context.Context, userPermission domain.UserPermission) (domain.UserPermission, error)
	RevokeUserPermission(ctx context.Context, bizID, id int64) error
	ListUserPermissionsByUserID(ctx context.Context, bizID, userID int64) ([]domain.UserPermission, error)
	ListUserPermissions(ctx context.Context, bizID int64, offset, limit int) ([]domain.UserPermission, error)
}

type rbacService struct {
	businessConfigRepo repository.BusinessConfigRepository
	resourceRepo       repository.ResourceRepository
	permissionRepo     repository.PermissionRepository
	roleRepo           repository.RoleRepository
	roleInclusionRepo  repository.RoleInclusionRepository
	rolePermissionRepo repository.RolePermissionRepository
	userRoleRepo       repository.UserRoleRepository
	userPermissionRepo repository.UserPermissionRepository
	jwtToken           *jwt.Token
}

// NewService 创建RBAC服务实例
func NewService(
	businessConfigRepo repository.BusinessConfigRepository,
	resourceRepo repository.ResourceRepository,
	permissionRepo repository.PermissionRepository,
	roleRepo repository.RoleRepository,
	roleInclusionRepo repository.RoleInclusionRepository,
	rolePermissionRepo repository.RolePermissionRepository,
	userRoleRepo repository.UserRoleRepository,
	userPermissionRepo repository.UserPermissionRepository,
	jwtToken *jwt.Token,
) Service {
	return &rbacService{
		businessConfigRepo: businessConfigRepo,
		resourceRepo:       resourceRepo,
		permissionRepo:     permissionRepo,
		roleRepo:           roleRepo,
		roleInclusionRepo:  roleInclusionRepo,
		rolePermissionRepo: rolePermissionRepo,
		userRoleRepo:       userRoleRepo,
		userPermissionRepo: userPermissionRepo,
		jwtToken:           jwtToken,
	}
}

// 业务配置相关方法实现

func (s *rbacService) CreateBusinessConfig(ctx context.Context, config domain.BusinessConfig) (domain.BusinessConfig, error) {
	businessConfig, err := s.businessConfigRepo.Create(ctx, config)
	if err != nil {
		return domain.BusinessConfig{}, err
	}
	const years = 10
	token, err := s.jwtToken.Encode(jwt.MapClaims{
		auth.BizIDName: businessConfig.ID,
		"exp":          time.Now().AddDate(years, 0, 0).Unix(),
	})
	if err != nil {
		return domain.BusinessConfig{}, err
	}
	businessConfig.Token = token
	err = s.businessConfigRepo.UpdateToken(ctx, businessConfig.ID, businessConfig.Token)
	if err != nil {
		return domain.BusinessConfig{}, err
	}
	return businessConfig, nil
}

func (s *rbacService) GetBusinessConfigByID(ctx context.Context, id int64) (domain.BusinessConfig, error) {
	return s.businessConfigRepo.FindByID(ctx, id)
}

func (s *rbacService) UpdateBusinessConfig(ctx context.Context, config domain.BusinessConfig) (domain.BusinessConfig, error) {
	return s.businessConfigRepo.Update(ctx, config)
}

func (s *rbacService) DeleteBusinessConfigByID(ctx context.Context, id int64) error {
	return s.businessConfigRepo.Delete(ctx, id)
}

func (s *rbacService) ListBusinessConfigs(ctx context.Context, offset, limit int) ([]domain.BusinessConfig, error) {
	return s.businessConfigRepo.Find(ctx, offset, limit)
}

// 资源相关方法实现

func (s *rbacService) CreateResource(ctx context.Context, resource domain.Resource) (domain.Resource, error) {
	return s.resourceRepo.Create(ctx, resource)
}

func (s *rbacService) GetResource(ctx context.Context, bizID, id int64) (domain.Resource, error) {
	return s.resourceRepo.FindByBizIDAndID(ctx, bizID, id)
}

func (s *rbacService) UpdateResource(ctx context.Context, resource domain.Resource) (domain.Resource, error) {
	return s.resourceRepo.UpdateByBizIDAndID(ctx, resource)
}

func (s *rbacService) DeleteResource(ctx context.Context, bizID, id int64) error {
	return s.resourceRepo.DeleteByBizIDAndID(ctx, bizID, id)
}

func (s *rbacService) ListResources(ctx context.Context, bizID int64, offset, limit int) ([]domain.Resource, error) {
	return s.resourceRepo.FindByBizID(ctx, bizID, offset, limit)
}

// 权限相关方法实现

func (s *rbacService) CreatePermission(ctx context.Context, permission domain.Permission) (domain.Permission, error) {
	return s.permissionRepo.Create(ctx, permission)
}

func (s *rbacService) GetPermission(ctx context.Context, bizID, id int64) (domain.Permission, error) {
	return s.permissionRepo.FindByBizIDAndID(ctx, bizID, id)
}

func (s *rbacService) UpdatePermission(ctx context.Context, permission domain.Permission) (domain.Permission, error) {
	return s.permissionRepo.UpdateByBizIDAndID(ctx, permission)
}

func (s *rbacService) DeletePermission(ctx context.Context, bizID, id int64) error {
	return s.permissionRepo.DeleteByBizIDAndID(ctx, bizID, id)
}

func (s *rbacService) ListPermissions(ctx context.Context, bizID int64, offset, limit int) ([]domain.Permission, error) {
	return s.permissionRepo.FindByBizID(ctx, bizID, offset, limit)
}

// 角色相关方法实现

func (s *rbacService) CreateRole(ctx context.Context, role domain.Role) (domain.Role, error) {
	return s.roleRepo.Create(ctx, role)
}

func (s *rbacService) GetRole(ctx context.Context, bizID, id int64) (domain.Role, error) {
	return s.roleRepo.FindByBizIDAndID(ctx, bizID, id)
}

func (s *rbacService) UpdateRole(ctx context.Context, role domain.Role) (domain.Role, error) {
	return s.roleRepo.UpdateByBizIDAndID(ctx, role)
}

func (s *rbacService) DeleteRole(ctx context.Context, bizID, id int64) error {
	return s.roleRepo.DeleteByBizIDAndID(ctx, bizID, id)
}

func (s *rbacService) ListRoles(ctx context.Context, bizID int64, offset, limit int) ([]domain.Role, error) {
	return s.roleRepo.FindByBizID(ctx, bizID, offset, limit)
}

func (s *rbacService) ListRolesByRoleType(ctx context.Context, bizID int64, roleType string, offset, limit int) ([]domain.Role, error) {
	return s.roleRepo.FindByBizIDAndType(ctx, bizID, roleType, offset, limit)
}

// 角色包含关系相关方法实现

func (s *rbacService) CreateRoleInclusion(ctx context.Context, roleInclusion domain.RoleInclusion) (domain.RoleInclusion, error) {
	return s.roleInclusionRepo.Create(ctx, roleInclusion)
}

func (s *rbacService) GetRoleInclusion(ctx context.Context, bizID, id int64) (domain.RoleInclusion, error) {
	return s.roleInclusionRepo.FindByBizIDAndID(ctx, bizID, id)
}

func (s *rbacService) DeleteRoleInclusion(ctx context.Context, bizID, id int64) error {
	return s.roleInclusionRepo.DeleteByBizIDAndID(ctx, bizID, id)
}

func (s *rbacService) ListRoleInclusions(ctx context.Context, bizID int64, offset, limit int) ([]domain.RoleInclusion, error) {
	return s.roleInclusionRepo.FindByBizID(ctx, bizID, offset, limit)
}

func (s *rbacService) ListRoleInclusionsByRoleID(ctx context.Context, bizID, roleID int64, isIncluding bool) ([]domain.RoleInclusion, error) {
	if isIncluding {
		return s.roleInclusionRepo.FindByBizIDAndIncludingRoleIDs(ctx, bizID, []int64{roleID})
	}
	return s.roleInclusionRepo.FindByBizIDAndIncludedRoleIDs(ctx, bizID, []int64{roleID})
}

// 角色权限相关方法实现

func (s *rbacService) GrantRolePermission(ctx context.Context, rolePermission domain.RolePermission) (domain.RolePermission, error) {
	return s.rolePermissionRepo.Create(ctx, rolePermission)
}

func (s *rbacService) RevokeRolePermission(ctx context.Context, bizID, id int64) error {
	return s.rolePermissionRepo.DeleteByBizIDAndID(ctx, bizID, id)
}

func (s *rbacService) ListRolePermissions(ctx context.Context, bizID int64) ([]domain.RolePermission, error) {
	return s.rolePermissionRepo.FindByBizID(ctx, bizID)
}

func (s *rbacService) ListRolePermissionsByRoleID(ctx context.Context, bizID, roleID int64) ([]domain.RolePermission, error) {
	return s.rolePermissionRepo.FindByBizIDAndRoleIDs(ctx, bizID, []int64{roleID})
}

// 用户角色相关方法实现

func (s *rbacService) GrantUserRole(ctx context.Context, userRole domain.UserRole) (domain.UserRole, error) {
	return s.userRoleRepo.Create(ctx, userRole)
}

func (s *rbacService) RevokeUserRole(ctx context.Context, bizID, id int64) error {
	return s.userRoleRepo.DeleteByBizIDAndID(ctx, bizID, id)
}

func (s *rbacService) ListUserRoles(ctx context.Context, bizID int64) ([]domain.UserRole, error) {
	return s.userRoleRepo.FindByBizID(ctx, bizID)
}

func (s *rbacService) ListUserRolesByUserID(ctx context.Context, bizID, userID int64) ([]domain.UserRole, error) {
	return s.userRoleRepo.FindByBizIDAndUserID(ctx, bizID, userID)
}

// 用户权限相关方法实现

func (s *rbacService) GrantUserPermission(ctx context.Context, userPermission domain.UserPermission) (domain.UserPermission, error) {
	return s.userPermissionRepo.Create(ctx, userPermission)
}

func (s *rbacService) RevokeUserPermission(ctx context.Context, bizID, id int64) error {
	return s.userPermissionRepo.DeleteByBizIDAndID(ctx, bizID, id)
}

func (s *rbacService) ListUserPermissions(ctx context.Context, bizID int64, offset, limit int) ([]domain.UserPermission, error) {
	return s.userPermissionRepo.FindByBizID(ctx, bizID, offset, limit)
}

func (s *rbacService) ListUserPermissionsByUserID(ctx context.Context, bizID, userID int64) ([]domain.UserPermission, error) {
	return s.userPermissionRepo.FindByBizIDAndUserID(ctx, bizID, userID)
}
