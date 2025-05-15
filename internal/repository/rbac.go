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
	// 1. 获取用户直接分配的权限
	directPermissions, err := r.userPermissionRepo.FindByBizIDAndUserID(ctx, bizID, userID)
	if err != nil {
		return nil, err
	}

	// 2.获取用户通过角色获得的权限
	// 2.1 获取用户的所有角色ID（包括继承的角色）
	allRoleIDs, err := r.getAllRoleIDs(ctx, bizID, userID)
	if err != nil {
		return nil, err
	}

	// 2.2 如果用户没有任何角色，则只返回直接权限
	if len(allRoleIDs) == 0 {
		return directPermissions, nil
	}

	// 2.3 获取所有角色对应的权限
	rolePermissions, err := r.getRolePermissions(ctx, bizID, userID, allRoleIDs)
	if err != nil {
		return nil, err
	}

	// 3. 合并两条路径的权限
	return r.mergePermissions(directPermissions, rolePermissions), nil
}

// getAllRoleIDs 获取用户所有角色ID，包括继承的角色
func (r *rbacRepository) getAllRoleIDs(ctx context.Context, bizID, userID int64) ([]int64, error) {
	// 1. 获取用户直接角色
	directRoles, err := r.userRoleRepo.FindByBizIDAndUserID(ctx, bizID, userID)
	if err != nil {
		return nil, err
	}

	// 如果用户没有任何直接角色，返回空切片
	if len(directRoles) == 0 {
		return []int64{}, nil
	}

	// 2. 提取直接角色ID, 收集所有角色ID
	directRoleIDs := make([]int64, 0, len(directRoles))
	allRoleIDs := make(map[int64]struct{})
	for i := range directRoles {
		directRoleIDs = append(directRoleIDs, directRoles[i].Role.ID)
		allRoleIDs[directRoles[i].Role.ID] = struct{}{}
	}

	// 3. 递归获取包含的角色
	err = r.collectIncludedRoles(ctx, bizID, directRoleIDs, allRoleIDs)
	if err != nil {
		return nil, err
	}

	// 4. 将角色ID集合转为切片
	result := make([]int64, 0, len(allRoleIDs))
	for roleID := range allRoleIDs {
		result = append(result, roleID)
	}
	return result, nil
}

// collectIncludedRoles 递归收集角色包含关系
func (r *rbacRepository) collectIncludedRoles(ctx context.Context, bizID int64, roleIDs []int64, allRoleIDs map[int64]struct{}) error {
	if len(roleIDs) == 0 {
		return nil
	}
	for {
		// 获取当前这批角色包含的所有角色
		inclusions, err := r.roleInclusionRepo.FindByBizIDAndIncludingRoleIDs(ctx, bizID, roleIDs)
		if err != nil {
			return err
		}

		// 如果没有更多包含关系，跳出循环
		if len(inclusions) == 0 {
			break
		}

		// 收集新发现的角色ID
		newRoleIDs := make([]int64, len(inclusions))
		for i := range inclusions {
			includedRoleID := inclusions[i].IncludedRole.ID
			// 如果是新角色，加入待处理列表
			if _, exists := allRoleIDs[includedRoleID]; !exists {
				allRoleIDs[includedRoleID] = struct{}{}
				newRoleIDs = append(newRoleIDs, includedRoleID)
			}
		}

		// 递归处理新发现的角色
		if len(newRoleIDs) > 0 {
			if err1 := r.collectIncludedRoles(ctx, bizID, newRoleIDs, allRoleIDs); err1 != nil {
				return err1
			}
		}
	}

	return nil
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

// mergePermissions 合并两组权限
func (r *rbacRepository) mergePermissions(directPermissions, rolePermissions []domain.UserPermission) []domain.UserPermission {
	if len(directPermissions) == 0 {
		return rolePermissions
	}
	if len(rolePermissions) == 0 {
		return directPermissions
	}
	mergedMap := make(map[int64]domain.UserPermission)
	// 先处理角色权限
	for i := range rolePermissions {
		mergedMap[rolePermissions[i].Permission.ID] = rolePermissions[i]
	}
	// 再处理直接权限（直接权限优先级更高）
	for i := range directPermissions {
		mergedMap[directPermissions[i].Permission.ID] = directPermissions[i]
	}
	return mapx.Values(mergedMap)
}
