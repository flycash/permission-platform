package repository

import (
	"context"
	"time"

	"github.com/ecodeclub/ekit/slice"

	"gitee.com/flycash/permission-platform/internal/domain"
	"github.com/ecodeclub/ekit/mapx"
)

type RBACRepository interface {
	BusinessConfig() BusinessConfigRepository
	Resource() ResourceRepository
	Permission() PermissionRepository
	Role() RoleRepository
	RoleInclusion() RoleInclusionRepository
	RolePermission() RolePermissionRepository
	UserRole() UserRoleRepository
	UserPermission() UserPermissionRepository
	// GetAllUserPermissions 获取用户的所有权限，包括个人权限、个人拥有的角色（及包含的角色）对应的权限
	GetAllUserPermissions(ctx context.Context, bizID, userID int64) ([]domain.UserPermission, error)
}

const (
	oneHundredYears = 100
)

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

// NewRBACRepository 聚合根
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

func (r *rbacRepository) BusinessConfig() BusinessConfigRepository {
	return r.businessConfigRepo
}

func (r *rbacRepository) Resource() ResourceRepository {
	return r.resourceRepo
}

func (r *rbacRepository) Permission() PermissionRepository {
	return r.permissionRepo
}

func (r *rbacRepository) Role() RoleRepository {
	return r.roleRepo
}

func (r *rbacRepository) RoleInclusion() RoleInclusionRepository {
	return r.roleInclusionRepo
}

func (r *rbacRepository) RolePermission() RolePermissionRepository {
	return r.rolePermissionRepo
}

func (r *rbacRepository) UserRole() UserRoleRepository {
	return r.userRoleRepo
}

func (r *rbacRepository) UserPermission() UserPermissionRepository {
	return r.userPermissionRepo
}

func (r *rbacRepository) GetAllUserPermissions(ctx context.Context, bizID, userID int64) ([]domain.UserPermission, error) {
	perms, err := r.userPermissionRepo.FindByBizIDAndUserID(ctx, bizID, userID)
	if err != nil {
		return nil, err
	}

	roleIDs, err := r.getAllRoleIDs(ctx, bizID, userID)
	if err != nil {
		return nil, err
	}

	// 包含了所需的冗余字段，所以我们不再需要查 Permission 表和 Resource
	rolePerms, err := r.getRolePermissions(ctx, bizID, userID, roleIDs)
	if err != nil {
		return nil, err
	}
	return append(perms, rolePerms...), nil
}

// getAllRoleIDs 获取用户所有角色ID，包括继承的角色
func (r *rbacRepository) getAllRoleIDs(ctx context.Context, bizID, userID int64) ([]int64, error) {
	// 1. 先找到直接关联的角色
	directRoles, err := r.userRoleRepo.FindByBizIDAndUserID(ctx, bizID, userID)
	if err != nil {
		return nil, err
	}

	allRoleIDs := make(map[int64]interface{}, len(directRoles))

	directRoleIDs := slice.Map(directRoles, func(_ int, src domain.UserRole) int64 {
		allRoleIDs[src.Role.ID] = struct{}{}
		return src.Role.ID
	})

	includedIDs := directRoleIDs

	for {
		// 找到 directRoles 包含的角色
		inclusions, err := r.roleInclusionRepo.FindByBizIDAndIncludingRoleIDs(ctx, bizID, includedIDs)
		if err != nil {
			return nil, err
		}
		if len(inclusions) == 0 {
			break
		}
		includedIDs = slice.Map(inclusions, func(_ int, src domain.RoleInclusion) int64 {
			allRoleIDs[src.IncludedRole.ID] = struct{}{}
			return src.IncludedRole.ID
		})
	}

	// 你要根据 inclusions 里面的 IncludedRoleID 进步沿着包含链找下去
	return mapx.Keys(allRoleIDs), nil
}

// getRolePermissions 获取指定角色ID列表对应的所有权限
func (r *rbacRepository) getRolePermissions(ctx context.Context, bizID, userID int64, roleIDs []int64) ([]domain.UserPermission, error) {
	if len(roleIDs) == 0 {
		return []domain.UserPermission{}, nil
	}
	permissions, err := r.rolePermissionRepo.FindByBizIDAndRoleIDs(ctx, bizID, roleIDs)
	if err != nil {
		return nil, err
	}
	// 将RolePermission转换为UserPermission格式
	return slice.Map(permissions, func(_ int, src domain.RolePermission) domain.UserPermission {
		return domain.UserPermission{
			ID:         0,
			BizID:      bizID,
			UserID:     userID,
			Permission: src.Permission,
			StartTime:  time.Now().UnixMilli(),
			EndTime:    time.Now().AddDate(oneHundredYears, 0, 0).UnixMilli(),
			Effect:     domain.EffectAllow,
			Ctime:      src.Ctime,
			Utime:      src.Utime,
		}
	}), nil
}
