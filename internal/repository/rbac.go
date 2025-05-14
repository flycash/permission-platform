package repository

import (
	"context"

	"gitee.com/flycash/permission-platform/internal/domain"
	"gitee.com/flycash/permission-platform/internal/repository/dao"
	"github.com/ecodeclub/ekit/slice"
)

type RBACRepository interface {
	// 业务配置相关方法

	CreateBusinessConfig(ctx context.Context, config domain.BusinessConfig) (domain.BusinessConfig, error)
	FindBusinessConfigs(ctx context.Context, offset, limit int) ([]domain.BusinessConfig, int, error)
	FindBusinessConfigByID(ctx context.Context, id int64) (domain.BusinessConfig, error)
	UpdateBusinessConfigToken(ctx context.Context, id int64, token string) error
	UpdateBusinessConfig(ctx context.Context, config domain.BusinessConfig) (domain.BusinessConfig, error)
	DeleteBusinessConfigByID(ctx context.Context, id int64) error

	// 资源相关方法

	CreateResource(ctx context.Context, resource domain.Resource) (domain.Resource, error)
	FindResourceByBizIDAndID(ctx context.Context, bizID, id int64) (domain.Resource, error)
	UpdateResourceByBizIDAndID(ctx context.Context, resource domain.Resource) (domain.Resource, error)
	DeleteResourceByBizIDAndID(ctx context.Context, bizID, id int64) error
	FindResourcesByBizIDAndTypeAndKey(ctx context.Context, bizID int64, resourceType, resourceKey string, offset, limit int) ([]domain.Resource, int, error)
	FindResourcesByBizID(ctx context.Context, bizID int64, offset, limit int) ([]domain.Resource, int, error)

	// 权限相关方法

	CreatePermission(ctx context.Context, permission domain.Permission) (domain.Permission, error)
	FindPermissionByBizIDAndID(ctx context.Context, bizID, id int64) (domain.Permission, error)
	FindPermissionsByBizIDAndResourceTypeAndKeyAndAction(ctx context.Context, bizID int64, resourceType, resourceKey, action string, offset, limit int) ([]domain.Permission, int, error)
	UpdatePermissionByBizIDAndID(ctx context.Context, permission domain.Permission) (domain.Permission, error)
	DeletePermissionByBizIDAndID(ctx context.Context, bizID, id int64) error
	FindPermissionsByBizID(ctx context.Context, bizID int64, offset, limit int) ([]domain.Permission, int, error)

	// 角色相关方法

	CreateRole(ctx context.Context, role domain.Role) (domain.Role, error)
	FindRoleByBizIDAndID(ctx context.Context, bizID, id int64) (domain.Role, error)
	FindRolesByBizIDAndType(ctx context.Context, bizID int64, roleType string, offset, limit int) ([]domain.Role, int, error)
	UpdateRoleByBizIDAndID(ctx context.Context, role domain.Role) (domain.Role, error)
	DeleteRoleByBizIDAndID(ctx context.Context, bizID, id int64) error
	FindRolesByBizID(ctx context.Context, bizID int64, offset, limit int) ([]domain.Role, int, error)

	// 角色包含关系相关方法

	CreateRoleInclusion(ctx context.Context, roleInclusion domain.RoleInclusion) (domain.RoleInclusion, error)
	FindRoleInclusionByBizIDAndID(ctx context.Context, bizID, id int64) (domain.RoleInclusion, error)
	FindRoleInclusionsByBizIDAndIncludingRoleID(ctx context.Context, bizID, includingRoleID int64, offset, limit int) ([]domain.RoleInclusion, int, error)
	FindRoleInclusionsByBizIDAndIncludedRoleID(ctx context.Context, bizID, includedRoleID int64, offset, limit int) ([]domain.RoleInclusion, int, error)
	DeleteRoleInclusionByBizIDAndID(ctx context.Context, bizID, id int64) error
	FindRoleInclusionsByBizID(ctx context.Context, bizID int64, offset, limit int) ([]domain.RoleInclusion, int, error)

	// 角色权限相关方法

	CreateRolePermission(ctx context.Context, rolePermission domain.RolePermission) (domain.RolePermission, error)
	FindRolePermissionsByBizIDAndRoleIDs(ctx context.Context, bizID int64, roleIDs []int64, offset, limit int) ([]domain.RolePermission, int, error)
	DeleteRolePermissionByBizIDAndID(ctx context.Context, bizID, id int64) error
	FindRolePermissionsByBizID(ctx context.Context, bizID int64, offset, limit int) ([]domain.RolePermission, int, error)

	// 用户角色相关方法

	CreateUserRole(ctx context.Context, userRole domain.UserRole) (domain.UserRole, error)
	FindUserRolesByBizIDAndUserID(ctx context.Context, bizID, userID int64, offset, limit int) ([]domain.UserRole, int, error)
	FindValidUserRolesByBizIDAndUserID(ctx context.Context, bizID, userID, currentTime int64, offset, limit int) ([]domain.UserRole, int, error)
	DeleteUserRoleByBizIDAndID(ctx context.Context, bizID, id int64) error
	FindUserRolesByBizID(ctx context.Context, bizID int64, offset, limit int) ([]domain.UserRole, int, error)

	// 用户权限相关方法

	CreateUserPermission(ctx context.Context, userPermission domain.UserPermission) (domain.UserPermission, error)
	FindUserPermissionsByBizIDAndUserID(ctx context.Context, bizID, userID int64, offset, limit int) ([]domain.UserPermission, int, error)
	FindValidUserPermissionsByBizIDAndUserID(ctx context.Context, bizID, userID, currentTime int64, offset, limit int) ([]domain.UserPermission, int, error)
	DeleteUserPermissionByBizIDAndID(ctx context.Context, bizID, id int64) error
	FindUserPermissionsByBizID(ctx context.Context, bizID int64, offset, limit int) ([]domain.UserPermission, int, error)
}

type rbacRepository struct {
	resourceDAO       dao.ResourceDAO
	permissionDAO     dao.PermissionDAO
	roleDAO           dao.RoleDAO
	rolePermissionDAO dao.RolePermissionDAO
	roleInclusionDAO  dao.RoleInclusionDAO
	userPermissionDAO dao.UserPermissionDAO
	userRoleDAO       dao.UserRoleDAO
	businessConfigDAO dao.BusinessConfigDAO
}

// NewRBACRepository 创建RBAC仓储的实例
func NewRBACRepository(
	resourceDAO dao.ResourceDAO,
	permissionDAO dao.PermissionDAO,
	roleDAO dao.RoleDAO,
	rolePermissionDAO dao.RolePermissionDAO,
	roleInclusionDAO dao.RoleInclusionDAO,
	userPermissionDAO dao.UserPermissionDAO,
	userRoleDAO dao.UserRoleDAO,
	businessConfigDAO dao.BusinessConfigDAO,
) RBACRepository {
	return &rbacRepository{
		resourceDAO:       resourceDAO,
		permissionDAO:     permissionDAO,
		roleDAO:           roleDAO,
		rolePermissionDAO: rolePermissionDAO,
		roleInclusionDAO:  roleInclusionDAO,
		userPermissionDAO: userPermissionDAO,
		userRoleDAO:       userRoleDAO,
		businessConfigDAO: businessConfigDAO,
	}
}

// ================== 业务配置相关方法 ==================

func (r *rbacRepository) CreateBusinessConfig(ctx context.Context, config domain.BusinessConfig) (domain.BusinessConfig, error) {
	created, err := r.businessConfigDAO.Create(ctx, r.toBusinessConfigEntity(config))
	if err != nil {
		return domain.BusinessConfig{}, err
	}
	return r.toBusinessConfigDomain(created), nil
}

func (r *rbacRepository) FindBusinessConfigs(ctx context.Context, offset, limit int) ([]domain.BusinessConfig, int, error) {
	list, err := r.businessConfigDAO.Find(ctx, offset, limit)
	if err != nil {
		return nil, 0, err
	}

	total, err := r.businessConfigDAO.Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	return slice.Map(list, func(_ int, src dao.BusinessConfig) domain.BusinessConfig {
		return r.toBusinessConfigDomain(src)
	}), int(total), nil
}

func (r *rbacRepository) FindBusinessConfigByID(ctx context.Context, id int64) (domain.BusinessConfig, error) {
	config, err := r.businessConfigDAO.GetByID(ctx, id)
	if err != nil {
		return domain.BusinessConfig{}, err
	}
	return r.toBusinessConfigDomain(config), nil
}

func (r *rbacRepository) UpdateBusinessConfig(ctx context.Context, config domain.BusinessConfig) (domain.BusinessConfig, error) {
	err := r.businessConfigDAO.Update(ctx, r.toBusinessConfigEntity(config))
	if err != nil {
		return domain.BusinessConfig{}, err
	}
	return config, nil
}

func (r *rbacRepository) UpdateBusinessConfigToken(ctx context.Context, id int64, token string) error {
	return r.businessConfigDAO.UpdateToken(ctx, id, token)
}

func (r *rbacRepository) DeleteBusinessConfigByID(ctx context.Context, id int64) error {
	return r.businessConfigDAO.Delete(ctx, id)
}

// ================== 资源相关方法 ==================

func (r *rbacRepository) CreateResource(ctx context.Context, resource domain.Resource) (domain.Resource, error) {
	created, err := r.resourceDAO.Create(ctx, r.toResourceEntity(resource))
	if err != nil {
		return domain.Resource{}, err
	}
	return r.toResourceDomain(created), nil
}

func (r *rbacRepository) FindResourceByBizIDAndID(ctx context.Context, bizID, id int64) (domain.Resource, error) {
	resource, err := r.resourceDAO.FindByBizIDAndID(ctx, bizID, id)
	if err != nil {
		return domain.Resource{}, err
	}
	return r.toResourceDomain(resource), nil
}

func (r *rbacRepository) UpdateResourceByBizIDAndID(ctx context.Context, resource domain.Resource) (domain.Resource, error) {
	err := r.resourceDAO.UpdateByBizIDAndID(ctx, r.toResourceEntity(resource))
	if err != nil {
		return domain.Resource{}, err
	}
	return resource, nil
}

func (r *rbacRepository) DeleteResourceByBizIDAndID(ctx context.Context, bizID, id int64) error {
	return r.resourceDAO.DeleteByBizIDAndID(ctx, bizID, id)
}

func (r *rbacRepository) FindResourcesByBizIDAndTypeAndKey(ctx context.Context, bizID int64, resourceType, resourceKey string, offset, limit int) ([]domain.Resource, int, error) {
	resources, err := r.resourceDAO.FindByBizIDAndTypeAndKey(ctx, bizID, resourceType, resourceKey, offset, limit)
	if err != nil {
		return nil, 0, err
	}

	total, err := r.resourceDAO.CountByBizIDAndTypeAndKey(ctx, bizID, resourceType, resourceKey)
	if err != nil {
		return nil, 0, err
	}

	return slice.Map(resources, func(_ int, src dao.Resource) domain.Resource {
		return r.toResourceDomain(src)
	}), int(total), nil
}

func (r *rbacRepository) FindResourcesByBizID(ctx context.Context, bizID int64, offset, limit int) ([]domain.Resource, int, error) {
	resources, err := r.resourceDAO.FindByBizID(ctx, bizID, offset, limit)
	if err != nil {
		return nil, 0, err
	}

	total, err := r.resourceDAO.CountByBizID(ctx, bizID)
	if err != nil {
		return nil, 0, err
	}

	return slice.Map(resources, func(_ int, src dao.Resource) domain.Resource {
		return r.toResourceDomain(src)
	}), int(total), nil
}

// ================== 权限相关方法 ==================

func (r *rbacRepository) CreatePermission(ctx context.Context, permission domain.Permission) (domain.Permission, error) {
	created, err := r.permissionDAO.Create(ctx, r.toPermissionEntity(permission))
	if err != nil {
		return domain.Permission{}, err
	}
	return r.toPermissionDomain(created), nil
}

func (r *rbacRepository) FindPermissionByBizIDAndID(ctx context.Context, bizID, id int64) (domain.Permission, error) {
	permission, err := r.permissionDAO.FindByBizIDAndID(ctx, bizID, id)
	if err != nil {
		return domain.Permission{}, err
	}
	return r.toPermissionDomain(permission), nil
}

func (r *rbacRepository) FindPermissionsByBizIDAndResourceTypeAndKeyAndAction(ctx context.Context, bizID int64, resourceType, resourceKey, action string, offset, limit int) ([]domain.Permission, int, error) {
	permissions, err := r.permissionDAO.FindByBizIDAndResourceTypeAndKeyAndAction(ctx, bizID, resourceType, resourceKey, action, offset, limit)
	if err != nil {
		return nil, 0, err
	}
	total, err := r.permissionDAO.CountByBizIDAndResourceTypeAndKeyAndAction(ctx, bizID, resourceType, resourceKey, action)
	if err != nil {
		return nil, 0, err
	}
	return slice.Map(permissions, func(_ int, src dao.Permission) domain.Permission {
		return r.toPermissionDomain(src)
	}), int(total), nil
}

func (r *rbacRepository) UpdatePermissionByBizIDAndID(ctx context.Context, permission domain.Permission) (domain.Permission, error) {
	err := r.permissionDAO.UpdateByBizIDAndID(ctx, r.toPermissionEntity(permission))
	if err != nil {
		return domain.Permission{}, err
	}
	return permission, nil
}

func (r *rbacRepository) DeletePermissionByBizIDAndID(ctx context.Context, bizID, id int64) error {
	return r.permissionDAO.DeleteByBizIDAndID(ctx, bizID, id)
}

func (r *rbacRepository) FindPermissionsByBizID(ctx context.Context, bizID int64, offset, limit int) ([]domain.Permission, int, error) {
	permissions, err := r.permissionDAO.FindByBizID(ctx, bizID, offset, limit)
	if err != nil {
		return nil, 0, err
	}

	total, err := r.permissionDAO.CountByBizID(ctx, bizID)
	if err != nil {
		return nil, 0, err
	}

	return slice.Map(permissions, func(_ int, src dao.Permission) domain.Permission {
		return r.toPermissionDomain(src)
	}), int(total), nil
}

// ================== 角色相关方法 ==================

func (r *rbacRepository) CreateRole(ctx context.Context, role domain.Role) (domain.Role, error) {
	created, err := r.roleDAO.Create(ctx, r.toRoleEntity(role))
	if err != nil {
		return domain.Role{}, err
	}
	return r.toRoleDomain(created), nil
}

func (r *rbacRepository) FindRoleByBizIDAndID(ctx context.Context, bizID, id int64) (domain.Role, error) {
	role, err := r.roleDAO.FindByBizIDAndID(ctx, bizID, id)
	if err != nil {
		return domain.Role{}, err
	}
	return r.toRoleDomain(role), nil
}

func (r *rbacRepository) FindRolesByBizIDAndType(ctx context.Context, bizID int64, roleType string, offset, limit int) ([]domain.Role, int, error) {
	roles, err := r.roleDAO.FindByBizIDAndType(ctx, bizID, roleType, offset, limit)
	if err != nil {
		return nil, 0, err
	}
	total, err := r.roleDAO.CountByBizIDAndType(ctx, bizID, roleType)
	if err != nil {
		return nil, 0, err
	}
	return slice.Map(roles, func(_ int, src dao.Role) domain.Role {
		return r.toRoleDomain(src)
	}), int(total), nil
}

func (r *rbacRepository) UpdateRoleByBizIDAndID(ctx context.Context, role domain.Role) (domain.Role, error) {
	err := r.roleDAO.UpdateByBizIDAndID(ctx, r.toRoleEntity(role))
	if err != nil {
		return domain.Role{}, err
	}
	return role, nil
}

func (r *rbacRepository) DeleteRoleByBizIDAndID(ctx context.Context, bizID, id int64) error {
	return r.roleDAO.DeleteByBizIDAndID(ctx, bizID, id)
}

func (r *rbacRepository) FindRolesByBizID(ctx context.Context, bizID int64, offset, limit int) ([]domain.Role, int, error) {
	roles, err := r.roleDAO.FindByBizID(ctx, bizID, offset, limit)
	if err != nil {
		return nil, 0, err
	}

	total, err := r.roleDAO.CountByBizID(ctx, bizID)
	if err != nil {
		return nil, 0, err
	}

	return slice.Map(roles, func(_ int, src dao.Role) domain.Role {
		return r.toRoleDomain(src)
	}), int(total), nil
}

// ================== 角色包含关系相关方法 ==================

func (r *rbacRepository) CreateRoleInclusion(ctx context.Context, roleInclusion domain.RoleInclusion) (domain.RoleInclusion, error) {
	created, err := r.roleInclusionDAO.Create(ctx, r.toRoleInclusionEntity(roleInclusion))
	if err != nil {
		return domain.RoleInclusion{}, err
	}
	return r.toRoleInclusionDomain(created), nil
}

func (r *rbacRepository) FindRoleInclusionByBizIDAndID(ctx context.Context, bizID, id int64) (domain.RoleInclusion, error) {
	roleInclusion, err := r.roleInclusionDAO.FindByBizIDAndID(ctx, bizID, id)
	if err != nil {
		return domain.RoleInclusion{}, err
	}
	return r.toRoleInclusionDomain(roleInclusion), nil
}

func (r *rbacRepository) FindRoleInclusionsByBizIDAndIncludingRoleID(ctx context.Context, bizID, includingRoleID int64, offset, limit int) ([]domain.RoleInclusion, int, error) {
	roleInclusions, err := r.roleInclusionDAO.FindByBizIDAndIncludingRoleID(ctx, bizID, includingRoleID, offset, limit)
	if err != nil {
		return nil, 0, err
	}
	total, err := r.roleInclusionDAO.CountByBizIDAndIncludingRoleID(ctx, bizID, includingRoleID)
	if err != nil {
		return nil, 0, err
	}
	return slice.Map(roleInclusions, func(_ int, src dao.RoleInclusion) domain.RoleInclusion {
		return r.toRoleInclusionDomain(src)
	}), int(total), nil
}

func (r *rbacRepository) FindRoleInclusionsByBizIDAndIncludedRoleID(ctx context.Context, bizID, includedRoleID int64, offset, limit int) ([]domain.RoleInclusion, int, error) {
	roleInclusions, err := r.roleInclusionDAO.FindByBizIDAndIncludedRoleID(ctx, bizID, includedRoleID, offset, limit)
	if err != nil {
		return nil, 0, err
	}
	total, err := r.roleInclusionDAO.CountByBizIDAndIncludedRoleID(ctx, bizID, includedRoleID)
	if err != nil {
		return nil, 0, err
	}
	return slice.Map(roleInclusions, func(_ int, src dao.RoleInclusion) domain.RoleInclusion {
		return r.toRoleInclusionDomain(src)
	}), int(total), nil
}

func (r *rbacRepository) DeleteRoleInclusionByBizIDAndID(ctx context.Context, bizID, id int64) error {
	return r.roleInclusionDAO.DeleteByBizIDAndID(ctx, bizID, id)
}

func (r *rbacRepository) FindRoleInclusionsByBizID(ctx context.Context, bizID int64, offset, limit int) ([]domain.RoleInclusion, int, error) {
	roleInclusions, err := r.roleInclusionDAO.FindByBizID(ctx, bizID, offset, limit)
	if err != nil {
		return nil, 0, err
	}

	total, err := r.roleInclusionDAO.CountByBizID(ctx, bizID)
	if err != nil {
		return nil, 0, err
	}

	return slice.Map(roleInclusions, func(_ int, src dao.RoleInclusion) domain.RoleInclusion {
		return r.toRoleInclusionDomain(src)
	}), int(total), nil
}

// ================== 角色权限相关方法 ==================

func (r *rbacRepository) CreateRolePermission(ctx context.Context, rolePermission domain.RolePermission) (domain.RolePermission, error) {
	created, err := r.rolePermissionDAO.Create(ctx, r.toRolePermissionEntity(rolePermission))
	if err != nil {
		return domain.RolePermission{}, err
	}
	return r.toRolePermissionDomain(created), nil
}

func (r *rbacRepository) FindRolePermissionsByBizIDAndRoleIDs(ctx context.Context, bizID int64, roleIDs []int64, offset, limit int) ([]domain.RolePermission, int, error) {
	rolePermissions, err := r.rolePermissionDAO.FindByBizIDAndRoleIDs(ctx, bizID, roleIDs, offset, limit)
	if err != nil {
		return nil, 0, err
	}
	total, err := r.rolePermissionDAO.CountByBizIDAndRoleIDs(ctx, bizID, roleIDs)
	if err != nil {
		return nil, 0, err
	}
	return slice.Map(rolePermissions, func(_ int, src dao.RolePermission) domain.RolePermission {
		return r.toRolePermissionDomain(src)
	}), int(total), nil
}

func (r *rbacRepository) DeleteRolePermissionByBizIDAndID(ctx context.Context, bizID, id int64) error {
	return r.rolePermissionDAO.DeleteByBizIDAndID(ctx, bizID, id)
}

func (r *rbacRepository) FindRolePermissionsByBizID(ctx context.Context, bizID int64, offset, limit int) ([]domain.RolePermission, int, error) {
	rolePermissions, err := r.rolePermissionDAO.FindByBizID(ctx, bizID, offset, limit)
	if err != nil {
		return nil, 0, err
	}

	total, err := r.rolePermissionDAO.CountByBizID(ctx, bizID)
	if err != nil {
		return nil, 0, err
	}

	return slice.Map(rolePermissions, func(_ int, src dao.RolePermission) domain.RolePermission {
		return r.toRolePermissionDomain(src)
	}), int(total), nil
}

// ================== 用户角色权限相关方法 ==================

func (r *rbacRepository) CreateUserRole(ctx context.Context, userRole domain.UserRole) (domain.UserRole, error) {
	created, err := r.userRoleDAO.Create(ctx, dao.UserRole{
		ID:        userRole.ID,
		BizID:     userRole.BizID,
		UserID:    userRole.UserID,
		RoleID:    userRole.Role.ID,
		RoleName:  userRole.Role.Name,
		RoleType:  userRole.Role.Type,
		StartTime: userRole.StartTime,
		EndTime:   userRole.EndTime,
		Ctime:     userRole.Ctime,
		Utime:     userRole.Utime,
	})
	if err != nil {
		return domain.UserRole{}, err
	}
	return r.toUserRoleDomain(created), nil
}

func (r *rbacRepository) FindUserRolesByBizIDAndUserID(ctx context.Context, bizID, userID int64, offset, limit int) ([]domain.UserRole, int, error) {
	userRoles, err := r.userRoleDAO.FindByBizIDAndUserID(ctx, bizID, userID, offset, limit)
	if err != nil {
		return nil, 0, err
	}
	total, err := r.userRoleDAO.CountByBizIDAndUserID(ctx, bizID, userID)
	if err != nil {
		return nil, 0, err
	}
	return slice.Map(userRoles, func(_ int, src dao.UserRole) domain.UserRole {
		return r.toUserRoleDomain(src)
	}), int(total), nil
}

func (r *rbacRepository) FindValidUserRolesByBizIDAndUserID(ctx context.Context, bizID, userID, currentTime int64, offset, limit int) ([]domain.UserRole, int, error) {
	userRoles, err := r.userRoleDAO.FindValidUserRolesWithBizID(ctx, bizID, userID, currentTime, offset, limit)
	if err != nil {
		return nil, 0, err
	}
	total, err := r.userRoleDAO.CountValidUserRolesWithBizID(ctx, bizID, userID, currentTime)
	if err != nil {
		return nil, 0, err
	}
	return slice.Map(userRoles, func(_ int, src dao.UserRole) domain.UserRole {
		return r.toUserRoleDomain(src)
	}), int(total), nil
}

func (r *rbacRepository) DeleteUserRoleByBizIDAndID(ctx context.Context, bizID, id int64) error {
	return r.userRoleDAO.DeleteByBizIDAndID(ctx, bizID, id)
}

func (r *rbacRepository) FindUserRolesByBizID(ctx context.Context, bizID int64, offset, limit int) ([]domain.UserRole, int, error) {
	userRoles, err := r.userRoleDAO.FindByBizID(ctx, bizID, offset, limit)
	if err != nil {
		return nil, 0, err
	}

	total, err := r.userRoleDAO.CountByBizID(ctx, bizID)
	if err != nil {
		return nil, 0, err
	}

	return slice.Map(userRoles, func(_ int, src dao.UserRole) domain.UserRole {
		return r.toUserRoleDomain(src)
	}), int(total), nil
}

// ================== 用户权限相关方法 ==================

func (r *rbacRepository) CreateUserPermission(ctx context.Context, userPermission domain.UserPermission) (domain.UserPermission, error) {
	created, err := r.userPermissionDAO.Create(ctx, dao.UserPermission{
		ID:               userPermission.ID,
		BizID:            userPermission.BizID,
		UserID:           userPermission.UserID,
		PermissionID:     userPermission.Permission.ID,
		PermissionName:   userPermission.Permission.Name,
		ResourceType:     userPermission.Permission.Resource.Type,
		ResourceKey:      userPermission.Permission.Resource.Key,
		ResourceName:     userPermission.Permission.Resource.Name,
		PermissionAction: userPermission.Permission.Action,
		StartTime:        userPermission.StartTime,
		EndTime:          userPermission.EndTime,
		Effect:           userPermission.Effect.String(),
		Ctime:            userPermission.Ctime,
		Utime:            userPermission.Utime,
	})
	if err != nil {
		return domain.UserPermission{}, err
	}
	return r.toUserPermissionDomain(created), nil
}

func (r *rbacRepository) FindUserPermissionsByBizIDAndUserID(ctx context.Context, bizID, userID int64, offset, limit int) ([]domain.UserPermission, int, error) {
	userPermissions, err := r.userPermissionDAO.FindByBizIDAndUserID(ctx, bizID, userID, offset, limit)
	if err != nil {
		return nil, 0, err
	}
	total, err := r.userPermissionDAO.CountByBizIDAndUserID(ctx, bizID, userID)
	if err != nil {
		return nil, 0, err
	}
	return slice.Map(userPermissions, func(_ int, src dao.UserPermission) domain.UserPermission {
		return r.toUserPermissionDomain(src)
	}), int(total), nil
}

func (r *rbacRepository) FindValidUserPermissionsByBizIDAndUserID(ctx context.Context, bizID, userID, currentTime int64, offset, limit int) ([]domain.UserPermission, int, error) {
	userPermissions, err := r.userPermissionDAO.FindValidPermissionsWithBizID(ctx, bizID, userID, currentTime, offset, limit)
	if err != nil {
		return nil, 0, err
	}
	total, err := r.userPermissionDAO.CountValidPermissionsWithBizID(ctx, bizID, userID, currentTime)
	if err != nil {
		return nil, 0, err
	}
	return slice.Map(userPermissions, func(_ int, src dao.UserPermission) domain.UserPermission {
		return r.toUserPermissionDomain(src)
	}), int(total), nil
}

func (r *rbacRepository) DeleteUserPermissionByBizIDAndID(ctx context.Context, bizID, id int64) error {
	return r.userPermissionDAO.DeleteByBizIDAndID(ctx, bizID, id)
}

func (r *rbacRepository) FindUserPermissionsByBizID(ctx context.Context, bizID int64, offset, limit int) ([]domain.UserPermission, int, error) {
	userPermissions, err := r.userPermissionDAO.FindByBizID(ctx, bizID, offset, limit)
	if err != nil {
		return nil, 0, err
	}

	total, err := r.userPermissionDAO.CountByBizID(ctx, bizID)
	if err != nil {
		return nil, 0, err
	}

	return slice.Map(userPermissions, func(_ int, src dao.UserPermission) domain.UserPermission {
		return r.toUserPermissionDomain(src)
	}), int(total), nil
}

func (r *rbacRepository) toBusinessConfigEntity(bc domain.BusinessConfig) dao.BusinessConfig {
	return dao.BusinessConfig{
		ID:        bc.ID,
		OwnerID:   bc.OwnerID,
		OwnerType: bc.OwnerType,
		Name:      bc.Name,
		RateLimit: bc.RateLimit,
		Token:     bc.Token,
		Ctime:     bc.Ctime,
		Utime:     bc.Utime,
	}
}

func (r *rbacRepository) toBusinessConfigDomain(bc dao.BusinessConfig) domain.BusinessConfig {
	return domain.BusinessConfig{
		ID:        bc.ID,
		OwnerID:   bc.OwnerID,
		OwnerType: bc.OwnerType,
		Name:      bc.Name,
		RateLimit: bc.RateLimit,
		Token:     bc.Token,
		Ctime:     bc.Ctime,
		Utime:     bc.Utime,
	}
}

func (r *rbacRepository) toResourceEntity(res domain.Resource) dao.Resource {
	return dao.Resource{
		ID:          res.ID,
		BizID:       res.BizID,
		Type:        res.Type,
		Key:         res.Key,
		Name:        res.Name,
		Description: res.Description,
		Metadata:    res.Metadata,
		Ctime:       res.Ctime,
		Utime:       res.Utime,
	}
}

func (r *rbacRepository) toResourceDomain(res dao.Resource) domain.Resource {
	return domain.Resource{
		ID:          res.ID,
		BizID:       res.BizID,
		Type:        res.Type,
		Key:         res.Key,
		Name:        res.Name,
		Description: res.Description,
		Metadata:    res.Metadata,
		Ctime:       res.Ctime,
		Utime:       res.Utime,
	}
}

func (r *rbacRepository) toPermissionEntity(p domain.Permission) dao.Permission {
	return dao.Permission{
		ID:           p.ID,
		BizID:        p.BizID,
		Name:         p.Name,
		Description:  p.Description,
		ResourceID:   p.Resource.ID,
		ResourceType: p.Resource.Type,
		ResourceKey:  p.Resource.Key,
		Action:       p.Action,
		Metadata:     p.Metadata,
		Ctime:        p.Ctime,
		Utime:        p.Utime,
	}
}

func (r *rbacRepository) toPermissionDomain(p dao.Permission) domain.Permission {
	return domain.Permission{
		ID:          p.ID,
		BizID:       p.BizID,
		Name:        p.Name,
		Description: p.Description,
		Resource: domain.Resource{
			ID:   p.ResourceID,
			Type: p.ResourceType,
			Key:  p.ResourceKey,
		},
		Action:   p.Action,
		Metadata: p.Metadata,
		Ctime:    p.Ctime,
		Utime:    p.Utime,
	}
}

func (r *rbacRepository) toRoleEntity(role domain.Role) dao.Role {
	return dao.Role{
		ID:          role.ID,
		BizID:       role.BizID,
		Type:        role.Type,
		Name:        role.Name,
		Description: role.Description,
		Metadata:    role.Metadata,
		Ctime:       role.Ctime,
		Utime:       role.Utime,
	}
}

func (r *rbacRepository) toRoleDomain(role dao.Role) domain.Role {
	return domain.Role{
		ID:          role.ID,
		BizID:       role.BizID,
		Type:        role.Type,
		Name:        role.Name,
		Description: role.Description,
		Metadata:    role.Metadata,
		Ctime:       role.Ctime,
		Utime:       role.Utime,
	}
}

func (r *rbacRepository) toRoleInclusionEntity(ri domain.RoleInclusion) dao.RoleInclusion {
	return dao.RoleInclusion{
		ID:                ri.ID,
		BizID:             ri.BizID,
		IncludingRoleID:   ri.IncludingRole.ID,
		IncludingRoleType: ri.IncludingRole.Type,
		IncludingRoleName: ri.IncludingRole.Name,
		IncludedRoleID:    ri.IncludedRole.ID,
		IncludedRoleType:  ri.IncludedRole.Type,
		IncludedRoleName:  ri.IncludedRole.Name,
		Ctime:             ri.Ctime,
		Utime:             ri.Utime,
	}
}

func (r *rbacRepository) toRoleInclusionDomain(ri dao.RoleInclusion) domain.RoleInclusion {
	return domain.RoleInclusion{
		ID:    ri.ID,
		BizID: ri.BizID,
		IncludingRole: domain.Role{
			ID:   ri.IncludingRoleID,
			Type: ri.IncludingRoleType,
			Name: ri.IncludingRoleName,
		},
		IncludedRole: domain.Role{
			ID:   ri.IncludedRoleID,
			Type: ri.IncludedRoleType,
			Name: ri.IncludedRoleName,
		},
		Ctime: ri.Ctime,
		Utime: ri.Utime,
	}
}

func (r *rbacRepository) toRolePermissionEntity(rp domain.RolePermission) dao.RolePermission {
	return dao.RolePermission{
		ID:               rp.ID,
		BizID:            rp.BizID,
		RoleID:           rp.Role.ID,
		RoleName:         rp.Role.Name,
		RoleType:         rp.Role.Type,
		PermissionID:     rp.Permission.ID,
		ResourceType:     rp.Permission.Resource.Type,
		ResourceKey:      rp.Permission.Resource.Key,
		PermissionAction: rp.Permission.Action,
		Ctime:            rp.Ctime,
		Utime:            rp.Utime,
	}
}

func (r *rbacRepository) toRolePermissionDomain(rp dao.RolePermission) domain.RolePermission {
	return domain.RolePermission{
		ID:    rp.ID,
		BizID: rp.BizID,
		Role: domain.Role{
			ID:   rp.RoleID,
			Name: rp.RoleName,
			Type: rp.RoleType,
		},
		Permission: domain.Permission{
			ID: rp.PermissionID,
			Resource: domain.Resource{
				Type: rp.ResourceType,
				Key:  rp.ResourceKey,
			},
			Action: rp.PermissionAction,
		},
		Ctime: rp.Ctime,
		Utime: rp.Utime,
	}
}

func (r *rbacRepository) toUserRoleDomain(created dao.UserRole) domain.UserRole {
	return domain.UserRole{
		ID:     created.ID,
		BizID:  created.BizID,
		UserID: created.UserID,
		Role: domain.Role{
			ID:   created.RoleID,
			Type: created.RoleType,
			Name: created.RoleName,
		},
		StartTime: created.StartTime,
		EndTime:   created.EndTime,
		Ctime:     created.Ctime,
		Utime:     created.Utime,
	}
}

func (r *rbacRepository) toUserPermissionDomain(created dao.UserPermission) domain.UserPermission {
	return domain.UserPermission{
		ID:     created.ID,
		BizID:  created.BizID,
		UserID: created.UserID,
		Permission: domain.Permission{
			ID:   created.PermissionID,
			Name: created.PermissionName,
			Resource: domain.Resource{
				Type: created.ResourceType,
				Key:  created.ResourceKey,
				Name: created.ResourceName,
			},
			Action: created.PermissionAction,
		},
		StartTime: created.StartTime,
		EndTime:   created.EndTime,
		Effect:    domain.Effect(created.Effect),
		Ctime:     created.Ctime,
		Utime:     created.Utime,
	}
}
