package rbac

import (
	"context"
	"time"

	"gitee.com/flycash/permission-platform/internal/domain"
	"gitee.com/flycash/permission-platform/internal/repository"
	"github.com/golang-jwt/jwt/v4"
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
	ListResourcesByTypeAndKey(ctx context.Context, bizID int64, resourceType, resourceKey string, offset, limit int) ([]domain.Resource, error)
	ListResources(ctx context.Context, bizID int64, offset, limit int) ([]domain.Resource, error)
	// 权限相关方法

	CreatePermission(ctx context.Context, permission domain.Permission) (domain.Permission, error)
	GetPermission(ctx context.Context, bizID, id int64) (domain.Permission, error)
	UpdatePermission(ctx context.Context, permission domain.Permission) (domain.Permission, error)
	DeletePermission(ctx context.Context, bizID, id int64) error
	ListPermissionsByResourceTypeAndKeyAndAction(ctx context.Context, bizID int64, resourceType, resourceKey, action string) ([]domain.Permission, error)
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
	repo repository.RBACRepository
}

// NewService 创建RBAC服务实例
func NewService(repo repository.RBACRepository) Service {
	return &rbacService{
		repo: repo,
	}
}

// 业务配置相关方法实现

func (s *rbacService) CreateBusinessConfig(ctx context.Context, config domain.BusinessConfig) (domain.BusinessConfig, error) {
	businessConfig, err := s.repo.BusinessConfig().Create(ctx, config)
	if err != nil {
		return domain.BusinessConfig{}, err
	}
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"iat":    time.Now().Unix(),
		"iss":    "permission-platform",
		"biz_id": businessConfig.ID,
	})
	token, err := jwtToken.SignedString([]byte(businessConfig.Name))
	if err != nil {
		return domain.BusinessConfig{}, err
	}
	businessConfig.Token = token
	err = s.repo.BusinessConfig().UpdateToken(ctx, businessConfig.ID, businessConfig.Token)
	if err != nil {
		return domain.BusinessConfig{}, err
	}
	return businessConfig, nil
}

func (s *rbacService) GetBusinessConfigByID(ctx context.Context, id int64) (domain.BusinessConfig, error) {
	return s.repo.BusinessConfig().FindByID(ctx, id)
}

func (s *rbacService) UpdateBusinessConfig(ctx context.Context, config domain.BusinessConfig) (domain.BusinessConfig, error) {
	return s.repo.BusinessConfig().Update(ctx, config)
}

func (s *rbacService) DeleteBusinessConfigByID(ctx context.Context, id int64) error {
	return s.repo.BusinessConfig().Delete(ctx, id)
}

func (s *rbacService) ListBusinessConfigs(ctx context.Context, offset, limit int) ([]domain.BusinessConfig, error) {
	return s.repo.BusinessConfig().Find(ctx, offset, limit)
}

// 资源相关方法实现

func (s *rbacService) CreateResource(ctx context.Context, resource domain.Resource) (domain.Resource, error) {
	return s.repo.Resource().Create(ctx, resource)
}

func (s *rbacService) GetResource(ctx context.Context, bizID, id int64) (domain.Resource, error) {
	return s.repo.Resource().FindByBizIDAndID(ctx, bizID, id)
}

func (s *rbacService) UpdateResource(ctx context.Context, resource domain.Resource) (domain.Resource, error) {
	return s.repo.Resource().UpdateByBizIDAndID(ctx, resource)
}

func (s *rbacService) DeleteResource(ctx context.Context, bizID, id int64) error {
	return s.repo.Resource().DeleteByBizIDAndID(ctx, bizID, id)
}

func (s *rbacService) ListResources(ctx context.Context, bizID int64, offset, limit int) ([]domain.Resource, error) {
	return s.repo.Resource().FindByBizID(ctx, bizID, offset, limit)
}

func (s *rbacService) ListResourcesByTypeAndKey(ctx context.Context, bizID int64, resourceType, resourceKey string, offset, limit int) ([]domain.Resource, error) {
	return s.repo.Resource().FindByBizIDAndTypeAndKey(ctx, bizID, resourceType, resourceKey, offset, limit)
}

// 权限相关方法实现

func (s *rbacService) CreatePermission(ctx context.Context, permission domain.Permission) (domain.Permission, error) {
	return s.repo.Permission().Create(ctx, permission)
}

func (s *rbacService) GetPermission(ctx context.Context, bizID, id int64) (domain.Permission, error) {
	return s.repo.Permission().FindByBizIDAndID(ctx, bizID, id)
}

func (s *rbacService) UpdatePermission(ctx context.Context, permission domain.Permission) (domain.Permission, error) {
	return s.repo.Permission().UpdateByBizIDAndID(ctx, permission)
}

func (s *rbacService) DeletePermission(ctx context.Context, bizID, id int64) error {
	return s.repo.Permission().DeleteByBizIDAndID(ctx, bizID, id)
}

func (s *rbacService) ListPermissions(ctx context.Context, bizID int64, offset, limit int) ([]domain.Permission, error) {
	return s.repo.Permission().FindByBizID(ctx, bizID, offset, limit)
}

func (s *rbacService) ListPermissionsByResourceTypeAndKeyAndAction(ctx context.Context, bizID int64, resourceType, resourceKey, action string) ([]domain.Permission, error) {
	return s.repo.Permission().FindByBizIDAndResourceTypeAndKeyAndAction(ctx, bizID, resourceType, resourceKey, action)
}

// 角色相关方法实现

func (s *rbacService) CreateRole(ctx context.Context, role domain.Role) (domain.Role, error) {
	return s.repo.Role().Create(ctx, role)
}

func (s *rbacService) GetRole(ctx context.Context, bizID, id int64) (domain.Role, error) {
	return s.repo.Role().FindByBizIDAndID(ctx, bizID, id)
}

func (s *rbacService) UpdateRole(ctx context.Context, role domain.Role) (domain.Role, error) {
	return s.repo.Role().UpdateByBizIDAndID(ctx, role)
}

func (s *rbacService) DeleteRole(ctx context.Context, bizID, id int64) error {
	return s.repo.Role().DeleteByBizIDAndID(ctx, bizID, id)
}

func (s *rbacService) ListRoles(ctx context.Context, bizID int64, offset, limit int) ([]domain.Role, error) {
	return s.repo.Role().FindByBizID(ctx, bizID, offset, limit)
}

func (s *rbacService) ListRolesByRoleType(ctx context.Context, bizID int64, roleType string, offset, limit int) ([]domain.Role, error) {
	return s.repo.Role().FindByBizIDAndType(ctx, bizID, roleType, offset, limit)
}

// 角色包含关系相关方法实现

func (s *rbacService) CreateRoleInclusion(ctx context.Context, roleInclusion domain.RoleInclusion) (domain.RoleInclusion, error) {
	return s.repo.RoleInclusion().Create(ctx, roleInclusion)
}

func (s *rbacService) GetRoleInclusion(ctx context.Context, bizID, id int64) (domain.RoleInclusion, error) {
	return s.repo.RoleInclusion().FindByBizIDAndID(ctx, bizID, id)
}

func (s *rbacService) DeleteRoleInclusion(ctx context.Context, bizID, id int64) error {
	return s.repo.RoleInclusion().DeleteByBizIDAndID(ctx, bizID, id)
}

func (s *rbacService) ListRoleInclusions(ctx context.Context, bizID int64, offset, limit int) ([]domain.RoleInclusion, error) {
	return s.repo.RoleInclusion().FindByBizID(ctx, bizID, offset, limit)
}

func (s *rbacService) ListRoleInclusionsByRoleID(ctx context.Context, bizID, roleID int64, isIncluding bool) ([]domain.RoleInclusion, error) {
	if isIncluding {
		return s.repo.RoleInclusion().FindByBizIDAndIncludingRoleIDs(ctx, bizID, []int64{roleID})
	}
	return s.repo.RoleInclusion().FindByBizIDAndIncludedRoleIDs(ctx, bizID, []int64{roleID})
}

// 角色权限相关方法实现

func (s *rbacService) GrantRolePermission(ctx context.Context, rolePermission domain.RolePermission) (domain.RolePermission, error) {
	return s.repo.RolePermission().Create(ctx, rolePermission)
}

func (s *rbacService) RevokeRolePermission(ctx context.Context, bizID, id int64) error {
	return s.repo.RolePermission().DeleteByBizIDAndID(ctx, bizID, id)
}

func (s *rbacService) ListRolePermissions(ctx context.Context, bizID int64) ([]domain.RolePermission, error) {
	return s.repo.RolePermission().FindByBizID(ctx, bizID)
}

func (s *rbacService) ListRolePermissionsByRoleID(ctx context.Context, bizID, roleID int64) ([]domain.RolePermission, error) {
	return s.repo.RolePermission().FindByBizIDAndRoleIDs(ctx, bizID, []int64{roleID})
}

// 用户角色相关方法实现

func (s *rbacService) GrantUserRole(ctx context.Context, userRole domain.UserRole) (domain.UserRole, error) {
	return s.repo.UserRole().Create(ctx, userRole)
}

func (s *rbacService) RevokeUserRole(ctx context.Context, bizID, id int64) error {
	return s.repo.UserRole().DeleteByBizIDAndID(ctx, bizID, id)
}

func (s *rbacService) ListUserRoles(ctx context.Context, bizID int64) ([]domain.UserRole, error) {
	return s.repo.UserRole().FindByBizID(ctx, bizID)
}

func (s *rbacService) ListUserRolesByUserID(ctx context.Context, bizID, userID int64) ([]domain.UserRole, error) {
	return s.repo.UserRole().FindByBizIDAndUserID(ctx, bizID, userID)
}

// 用户权限相关方法实现

func (s *rbacService) GrantUserPermission(ctx context.Context, userPermission domain.UserPermission) (domain.UserPermission, error) {
	return s.repo.UserPermission().Create(ctx, userPermission)
}

func (s *rbacService) RevokeUserPermission(ctx context.Context, bizID, id int64) error {
	return s.repo.UserPermission().DeleteByBizIDAndID(ctx, bizID, id)
}

func (s *rbacService) ListUserPermissions(ctx context.Context, bizID int64, offset, limit int) ([]domain.UserPermission, error) {
	return s.repo.UserPermission().FindByBizID(ctx, bizID, offset, limit)
}

func (s *rbacService) ListUserPermissionsByUserID(ctx context.Context, bizID, userID int64) ([]domain.UserPermission, error) {
	return s.repo.UserPermission().FindByBizIDAndUserID(ctx, bizID, userID)
}
