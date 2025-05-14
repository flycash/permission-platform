package repository

import (
	"context"

	"gitee.com/flycash/permission-platform/internal/domain"
	"gitee.com/flycash/permission-platform/internal/repository/dao"
)

type RBACRepository interface {
	// 业务配置相关方法

	CreateBusinessConfig(ctx context.Context, config domain.BusinessConfig) (domain.BusinessConfig, error)
	FindBusinessConfigs(ctx context.Context, offset, limit int) ([]domain.BusinessConfig, error)
	FindBusinessConfigByID(ctx context.Context, id int64) (domain.BusinessConfig, error)
	UpdateBusinessConfigToken(ctx context.Context, id int64, token string) error
	UpdateBusinessConfig(ctx context.Context, config domain.BusinessConfig) (domain.BusinessConfig, error)
	DeleteBusinessConfigByID(ctx context.Context, id int64) error

	// 资源相关方法

	CreateResource(ctx context.Context, resource domain.Resource) (domain.Resource, error)
	FindResourceByBizIDAndID(ctx context.Context, bizID, id int64) (domain.Resource, error)
	UpdateResourceByBizIDAndID(ctx context.Context, resource domain.Resource) (domain.Resource, error)
	DeleteResourceByBizIDAndID(ctx context.Context, bizID, id int64) error
	FindResourcesByBizIDAndTypeAndKey(ctx context.Context, bizID int64, resourceType, resourceKey string, offset, limit int) ([]domain.Resource, error)
	FindResourcesByBizID(ctx context.Context, bizID int64, offset, limit int) ([]domain.Resource, error)

	// 权限相关方法

	CreatePermission(ctx context.Context, permission domain.Permission) (domain.Permission, error)
	FindPermissionByBizIDAndID(ctx context.Context, bizID, id int64) (domain.Permission, error)
	FindPermissionsByBizIDAndResourceTypeAndKeyAndAction(ctx context.Context, bizID int64, resourceType, resourceKey, action string, offset, limit int) ([]domain.Permission, error)
	UpdatePermissionByBizIDAndID(ctx context.Context, permission domain.Permission) (domain.Permission, error)
	DeletePermissionByBizIDAndID(ctx context.Context, bizID, id int64) error
	FindPermissionsByBizID(ctx context.Context, bizID int64, offset, limit int) ([]domain.Permission, error)

	// 角色相关方法

	CreateRole(ctx context.Context, role domain.Role) (domain.Role, error)
	FindRoleByBizIDAndID(ctx context.Context, bizID, id int64) (domain.Role, error)
	FindRolesByBizIDAndType(ctx context.Context, bizID int64, roleType string, offset, limit int) ([]domain.Role, error)
	UpdateRoleByBizIDAndID(ctx context.Context, role domain.Role) (domain.Role, error)
	DeleteRoleByBizIDAndID(ctx context.Context, bizID, id int64) error
	FindRolesByBizID(ctx context.Context, bizID int64, offset, limit int) ([]domain.Role, error)

	// 角色包含关系相关方法

	CreateRoleInclusion(ctx context.Context, roleInclusion domain.RoleInclusion) (domain.RoleInclusion, error)
	FindRoleInclusionByBizIDAndID(ctx context.Context, bizID, id int64) (domain.RoleInclusion, error)
	FindRoleInclusionsByBizIDAndIncludingRoleID(ctx context.Context, bizID, includingRoleID int64, offset, limit int) ([]domain.RoleInclusion, error)
	FindRoleInclusionsByBizIDAndIncludedRoleID(ctx context.Context, bizID, includedRoleID int64, offset, limit int) ([]domain.RoleInclusion, error)
	DeleteRoleInclusionByBizIDAndID(ctx context.Context, bizID, id int64) error
	FindRoleInclusionsByBizID(ctx context.Context, bizID int64, offset, limit int) ([]domain.RoleInclusion, error)

	// 角色权限相关方法

	CreateRolePermission(ctx context.Context, rolePermission domain.RolePermission) (domain.RolePermission, error)
	FindRolePermissionsByBizIDAndRoleIDs(ctx context.Context, bizID int64, roleIDs []int64, offset, limit int) ([]domain.RolePermission, error)
	DeleteRolePermissionByBizIDAndID(ctx context.Context, bizID, id int64) error
	FindRolePermissionsByBizID(ctx context.Context, bizID int64, offset, limit int) ([]domain.RolePermission, error)

	// 用户角色相关方法

	CreateUserRole(ctx context.Context, userRole domain.UserRole) (domain.UserRole, error)
	FindUserRolesByBizIDAndUserID(ctx context.Context, bizID, userID int64, offset, limit int) ([]domain.UserRole, error)
	FindValidUserRolesByBizIDAndUserID(ctx context.Context, bizID, userID, currentTime int64, offset, limit int) ([]domain.UserRole, error)
	DeleteUserRoleByBizIDAndID(ctx context.Context, bizID, id int64) error
	FindUserRolesByBizID(ctx context.Context, bizID int64, offset, limit int) ([]domain.UserRole, error)

	// 用户权限相关方法

	CreateUserPermission(ctx context.Context, userPermission domain.UserPermission) (domain.UserPermission, error)
	FindUserPermissionsByBizIDAndUserID(ctx context.Context, bizID, userID int64, offset, limit int) ([]domain.UserPermission, error)
	FindValidUserPermissionsByBizIDAndUserID(ctx context.Context, bizID, userID, currentTime int64, offset, limit int) ([]domain.UserPermission, error)
	DeleteUserPermissionByBizIDAndID(ctx context.Context, bizID, id int64) error
	FindUserPermissionsByBizID(ctx context.Context, bizID int64, offset, limit int) ([]domain.UserPermission, error)
}

type rbacRepository struct {
	businessConfigRepo BusinessConfigRepository
	resourceRepo       ResourceRepository
	permissionRepo     PermissionRepository
	roleRepo           RoleRepository
	roleInclusionRepo  RoleInclusionRepository
	rolePermissionRepo RolePermissionRepository
	userRoleRepo       UserRoleRepository
	userPermissionRepo UserPermissionRepository
}

// NewRBACRepository 接收8个子Repository
func NewRBACRepository(
	businessConfigRepo BusinessConfigRepository,
	resourceRepo ResourceRepository,
	permissionRepo PermissionRepository,
	roleRepo RoleRepository,
	roleInclusionRepo RoleInclusionRepository,
	rolePermissionRepo RolePermissionRepository,
	userRoleRepo UserRoleRepository,
	userPermissionRepo UserPermissionRepository,
) RBACRepository {
	return &rbacRepository{
		businessConfigRepo: businessConfigRepo,
		resourceRepo:       resourceRepo,
		permissionRepo:     permissionRepo,
		roleRepo:           roleRepo,
		roleInclusionRepo:  roleInclusionRepo,
		rolePermissionRepo: rolePermissionRepo,
		userRoleRepo:       userRoleRepo,
		userPermissionRepo: userPermissionRepo,
	}
}

// NewRBACRepositoryOld 原有工厂函数的兼容包装
func NewRBACRepositoryOld(
	resourceDAO dao.ResourceDAO,
	permissionDAO dao.PermissionDAO,
	roleDAO dao.RoleDAO,
	rolePermissionDAO dao.RolePermissionDAO,
	roleInclusionDAO dao.RoleInclusionDAO,
	userPermissionDAO dao.UserPermissionDAO,
	userRoleDAO dao.UserRoleDAO,
	businessConfigDAO dao.BusinessConfigDAO,
) RBACRepository {
	businessConfigRepo := NewBusinessConfigRepository(businessConfigDAO)
	resourceRepo := NewResourceRepository(resourceDAO)
	permissionRepo := NewPermissionRepository(permissionDAO)
	roleRepo := NewRoleRepository(roleDAO)
	roleInclusionRepo := NewRoleInclusionRepository(roleInclusionDAO)
	rolePermissionRepo := NewRolePermissionRepository(rolePermissionDAO)
	userRoleRepo := NewUserRoleRepository(userRoleDAO)
	userPermissionRepo := NewUserPermissionRepository(userPermissionDAO)

	return NewRBACRepository(
		businessConfigRepo,
		resourceRepo,
		permissionRepo,
		roleRepo,
		roleInclusionRepo,
		rolePermissionRepo,
		userRoleRepo,
		userPermissionRepo,
	)
}

// ========== 业务配置相关方法实现 ==========

func (r *rbacRepository) CreateBusinessConfig(ctx context.Context, config domain.BusinessConfig) (domain.BusinessConfig, error) {
	return r.businessConfigRepo.Create(ctx, config)
}

func (r *rbacRepository) FindBusinessConfigs(ctx context.Context, offset, limit int) ([]domain.BusinessConfig, error) {
	return r.businessConfigRepo.Find(ctx, offset, limit)
}

func (r *rbacRepository) FindBusinessConfigByID(ctx context.Context, id int64) (domain.BusinessConfig, error) {
	return r.businessConfigRepo.FindByID(ctx, id)
}

func (r *rbacRepository) UpdateBusinessConfig(ctx context.Context, config domain.BusinessConfig) (domain.BusinessConfig, error) {
	return r.businessConfigRepo.Update(ctx, config)
}

func (r *rbacRepository) UpdateBusinessConfigToken(ctx context.Context, id int64, token string) error {
	return r.businessConfigRepo.UpdateToken(ctx, id, token)
}

func (r *rbacRepository) DeleteBusinessConfigByID(ctx context.Context, id int64) error {
	return r.businessConfigRepo.Delete(ctx, id)
}

// ========== 资源相关方法实现 ==========

func (r *rbacRepository) CreateResource(ctx context.Context, resource domain.Resource) (domain.Resource, error) {
	return r.resourceRepo.Create(ctx, resource)
}

func (r *rbacRepository) FindResourceByBizIDAndID(ctx context.Context, bizID, id int64) (domain.Resource, error) {
	return r.resourceRepo.FindByBizIDAndID(ctx, bizID, id)
}

func (r *rbacRepository) UpdateResourceByBizIDAndID(ctx context.Context, resource domain.Resource) (domain.Resource, error) {
	return r.resourceRepo.UpdateByBizIDAndID(ctx, resource)
}

func (r *rbacRepository) DeleteResourceByBizIDAndID(ctx context.Context, bizID, id int64) error {
	return r.resourceRepo.DeleteByBizIDAndID(ctx, bizID, id)
}

func (r *rbacRepository) FindResourcesByBizIDAndTypeAndKey(ctx context.Context, bizID int64, resourceType, resourceKey string, offset, limit int) ([]domain.Resource, error) {
	return r.resourceRepo.FindByBizIDAndTypeAndKey(ctx, bizID, resourceType, resourceKey, offset, limit)
}

func (r *rbacRepository) FindResourcesByBizID(ctx context.Context, bizID int64, offset, limit int) ([]domain.Resource, error) {
	return r.resourceRepo.FindByBizID(ctx, bizID, offset, limit)
}

// ========== 权限相关方法实现 ==========

func (r *rbacRepository) CreatePermission(ctx context.Context, permission domain.Permission) (domain.Permission, error) {
	return r.permissionRepo.Create(ctx, permission)
}

func (r *rbacRepository) FindPermissionByBizIDAndID(ctx context.Context, bizID, id int64) (domain.Permission, error) {
	return r.permissionRepo.FindByBizIDAndID(ctx, bizID, id)
}

func (r *rbacRepository) FindPermissionsByBizIDAndResourceTypeAndKeyAndAction(ctx context.Context, bizID int64, resourceType, resourceKey, action string, offset, limit int) ([]domain.Permission, error) {
	return r.permissionRepo.FindByBizIDAndResourceTypeAndKeyAndAction(ctx, bizID, resourceType, resourceKey, action, offset, limit)
}

func (r *rbacRepository) UpdatePermissionByBizIDAndID(ctx context.Context, permission domain.Permission) (domain.Permission, error) {
	return r.permissionRepo.UpdateByBizIDAndID(ctx, permission)
}

func (r *rbacRepository) DeletePermissionByBizIDAndID(ctx context.Context, bizID, id int64) error {
	return r.permissionRepo.DeleteByBizIDAndID(ctx, bizID, id)
}

func (r *rbacRepository) FindPermissionsByBizID(ctx context.Context, bizID int64, offset, limit int) ([]domain.Permission, error) {
	return r.permissionRepo.FindByBizID(ctx, bizID, offset, limit)
}

// ========== 角色相关方法实现 ==========

func (r *rbacRepository) CreateRole(ctx context.Context, role domain.Role) (domain.Role, error) {
	return r.roleRepo.Create(ctx, role)
}

func (r *rbacRepository) FindRoleByBizIDAndID(ctx context.Context, bizID, id int64) (domain.Role, error) {
	return r.roleRepo.FindByBizIDAndID(ctx, bizID, id)
}

func (r *rbacRepository) FindRolesByBizIDAndType(ctx context.Context, bizID int64, roleType string, offset, limit int) ([]domain.Role, error) {
	return r.roleRepo.FindByBizIDAndType(ctx, bizID, roleType, offset, limit)
}

func (r *rbacRepository) UpdateRoleByBizIDAndID(ctx context.Context, role domain.Role) (domain.Role, error) {
	return r.roleRepo.UpdateByBizIDAndID(ctx, role)
}

func (r *rbacRepository) DeleteRoleByBizIDAndID(ctx context.Context, bizID, id int64) error {
	return r.roleRepo.DeleteByBizIDAndID(ctx, bizID, id)
}

func (r *rbacRepository) FindRolesByBizID(ctx context.Context, bizID int64, offset, limit int) ([]domain.Role, error) {
	return r.roleRepo.FindByBizID(ctx, bizID, offset, limit)
}

// ========== 角色包含关系相关方法实现 ==========

func (r *rbacRepository) CreateRoleInclusion(ctx context.Context, roleInclusion domain.RoleInclusion) (domain.RoleInclusion, error) {
	return r.roleInclusionRepo.Create(ctx, roleInclusion)
}

func (r *rbacRepository) FindRoleInclusionByBizIDAndID(ctx context.Context, bizID, id int64) (domain.RoleInclusion, error) {
	return r.roleInclusionRepo.FindByBizIDAndID(ctx, bizID, id)
}

func (r *rbacRepository) FindRoleInclusionsByBizIDAndIncludingRoleID(ctx context.Context, bizID, includingRoleID int64, offset, limit int) ([]domain.RoleInclusion, error) {
	return r.roleInclusionRepo.FindByBizIDAndIncludingRoleID(ctx, bizID, includingRoleID, offset, limit)
}

func (r *rbacRepository) FindRoleInclusionsByBizIDAndIncludedRoleID(ctx context.Context, bizID, includedRoleID int64, offset, limit int) ([]domain.RoleInclusion, error) {
	return r.roleInclusionRepo.FindByBizIDAndIncludedRoleID(ctx, bizID, includedRoleID, offset, limit)
}

func (r *rbacRepository) DeleteRoleInclusionByBizIDAndID(ctx context.Context, bizID, id int64) error {
	return r.roleInclusionRepo.DeleteByBizIDAndID(ctx, bizID, id)
}

func (r *rbacRepository) FindRoleInclusionsByBizID(ctx context.Context, bizID int64, offset, limit int) ([]domain.RoleInclusion, error) {
	return r.roleInclusionRepo.FindByBizID(ctx, bizID, offset, limit)
}

// ========== 角色权限相关方法实现 ==========

func (r *rbacRepository) CreateRolePermission(ctx context.Context, rolePermission domain.RolePermission) (domain.RolePermission, error) {
	return r.rolePermissionRepo.Create(ctx, rolePermission)
}

func (r *rbacRepository) FindRolePermissionsByBizIDAndRoleIDs(ctx context.Context, bizID int64, roleIDs []int64, offset, limit int) ([]domain.RolePermission, error) {
	return r.rolePermissionRepo.FindByBizIDAndRoleIDs(ctx, bizID, roleIDs, offset, limit)
}

func (r *rbacRepository) DeleteRolePermissionByBizIDAndID(ctx context.Context, bizID, id int64) error {
	return r.rolePermissionRepo.DeleteByBizIDAndID(ctx, bizID, id)
}

func (r *rbacRepository) FindRolePermissionsByBizID(ctx context.Context, bizID int64, offset, limit int) ([]domain.RolePermission, error) {
	return r.rolePermissionRepo.FindByBizID(ctx, bizID, offset, limit)
}

// ========== 用户角色相关方法实现 ==========

func (r *rbacRepository) CreateUserRole(ctx context.Context, userRole domain.UserRole) (domain.UserRole, error) {
	return r.userRoleRepo.Create(ctx, userRole)
}

func (r *rbacRepository) FindUserRolesByBizIDAndUserID(ctx context.Context, bizID, userID int64, offset, limit int) ([]domain.UserRole, error) {
	return r.userRoleRepo.FindByBizIDAndUserID(ctx, bizID, userID, offset, limit)
}

func (r *rbacRepository) FindValidUserRolesByBizIDAndUserID(ctx context.Context, bizID, userID, currentTime int64, offset, limit int) ([]domain.UserRole, error) {
	return r.userRoleRepo.FindValidByBizIDAndUserID(ctx, bizID, userID, currentTime, offset, limit)
}

func (r *rbacRepository) DeleteUserRoleByBizIDAndID(ctx context.Context, bizID, id int64) error {
	return r.userRoleRepo.DeleteByBizIDAndID(ctx, bizID, id)
}

func (r *rbacRepository) FindUserRolesByBizID(ctx context.Context, bizID int64, offset, limit int) ([]domain.UserRole, error) {
	return r.userRoleRepo.FindByBizID(ctx, bizID, offset, limit)
}

// ========== 用户权限相关方法实现 ==========

func (r *rbacRepository) CreateUserPermission(ctx context.Context, userPermission domain.UserPermission) (domain.UserPermission, error) {
	return r.userPermissionRepo.Create(ctx, userPermission)
}

func (r *rbacRepository) FindUserPermissionsByBizIDAndUserID(ctx context.Context, bizID, userID int64, offset, limit int) ([]domain.UserPermission, error) {
	return r.userPermissionRepo.FindByBizIDAndUserID(ctx, bizID, userID, offset, limit)
}

func (r *rbacRepository) FindValidUserPermissionsByBizIDAndUserID(ctx context.Context, bizID, userID, currentTime int64, offset, limit int) ([]domain.UserPermission, error) {
	return r.userPermissionRepo.FindValidByBizIDAndUserID(ctx, bizID, userID, currentTime, offset, limit)
}

func (r *rbacRepository) DeleteUserPermissionByBizIDAndID(ctx context.Context, bizID, id int64) error {
	return r.userPermissionRepo.DeleteByBizIDAndID(ctx, bizID, id)
}

func (r *rbacRepository) FindUserPermissionsByBizID(ctx context.Context, bizID int64, offset, limit int) ([]domain.UserPermission, error) {
	return r.userPermissionRepo.FindByBizID(ctx, bizID, offset, limit)
}
