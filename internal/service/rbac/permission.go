package rbac

import (
	"context"
	"time"

	"gitee.com/flycash/permission-platform/internal/domain"
	"gitee.com/flycash/permission-platform/internal/repository"
	"github.com/ecodeclub/ekit/slice"
)

// PermissionService RBAC模型下的权限服务接口
type PermissionService interface {
	// Check 检查用户是否有对特定权限
	Check(ctx context.Context, bizID, userID int64, permission domain.Permission) (bool, error)
}

type permissionService struct {
	repo  repository.RBACRepository
	limit int
}

// NewPermissionService 创建RBAC权限服务实例
func NewPermissionService(repo repository.RBACRepository, limit int) PermissionService {
	return &permissionService{
		repo:  repo,
		limit: limit,
	}
}

// Check 检查用户权限
func (s *permissionService) Check(ctx context.Context, bizID, userID int64, permission domain.Permission) (bool, error) {
	currentTime := time.Now().UnixMilli()
	// 1. 检查用户直接权限
	hasMatch, isAllowed, err := s.checkUserDirectPermissions(ctx, bizID, userID, permission, currentTime)
	if err != nil {
		return false, err
	}
	// 如果有直接权限，根据是否允许返回结果
	if hasMatch {
		return isAllowed, nil
	}

	// 2. 获取用户所有角色ID（包括继承角色）
	allRoleIDs, err := s.collectAllRoleIDs(ctx, bizID, userID, currentTime)
	if err != nil {
		return false, err
	}
	// 如果用户没有任何角色，则无权限
	if len(allRoleIDs) == 0 {
		return false, nil
	}

	// 3. 检查角色权限
	hasRolePermission, err := s.checkRolePermissions(ctx, bizID, allRoleIDs, permission)
	if err != nil {
		return false, err
	}
	return hasRolePermission, nil
}

// checkUserDirectPermissions 检查用户直接分配的权限
// 返回值：
// - hasMatch: 是否找到匹配的权限
// - isAllowed: 若找到匹配的权限，是否允许访问
// - err: 可能的错误
func (s *permissionService) checkUserDirectPermissions(
	ctx context.Context,
	bizID,
	userID int64,
	permission domain.Permission,
	currentTime int64,
) (hasMatch, isAllowed bool, err error) {
	// 获取用户所有直接权限
	userPermissions, err := s.getUserPermissions(ctx, bizID, userID, currentTime)
	if err != nil {
		return false, false, err
	}
	// 检查是否有匹配的权限
	for i := range userPermissions {
		if userPermissions[i].Permission.Resource.Type == permission.Resource.Type &&
			userPermissions[i].Permission.Resource.Key == permission.Resource.Key &&
			userPermissions[i].Permission.Action == permission.Action {
			if userPermissions[i].Effect.IsDeny() {
				return true, false, nil // 有匹配的权限，但拒绝访问
			}
			return true, true, nil // 有匹配的权限，允许访问
		}
	}
	return false, false, nil
}

// getUserPermissions 获取用户所有直接分配的权限（处理分页）
func (s *permissionService) getUserPermissions(
	ctx context.Context,
	bizID,
	userID,
	currentTime int64,
) ([]domain.UserPermission, error) {
	offset := 0
	var allUserPermissions []domain.UserPermission

	for {
		userPermissions, total, err := s.repo.FindValidUserPermissionsByBizIDAndUserID(ctx, bizID, userID, currentTime, offset, s.limit)
		if err != nil {
			return nil, err
		}
		allUserPermissions = append(allUserPermissions, userPermissions...)

		// 优化分页边界条件判断
		if len(userPermissions) < s.limit || int64(offset+s.limit) >= int64(total) {
			break
		}
		offset += s.limit
	}

	return allUserPermissions, nil
}

// collectAllRoleIDs 获取用户所有角色ID，包括继承角色
func (s *permissionService) collectAllRoleIDs(
	ctx context.Context,
	bizID,
	userID,
	currentTime int64,
) ([]int64, error) {
	// 获取用户直接角色
	userRoles, err := s.getUserRoles(ctx, bizID, userID, currentTime)
	if err != nil {
		return nil, err
	}

	// 如果用户没有任何角色，返回空切片
	if len(userRoles) == 0 {
		return []int64{}, nil
	}

	// 收集用户的直接角色ID
	directRoleIDs := slice.Map(userRoles, func(_ int, src domain.UserRole) int64 {
		return src.Role.ID
	})

	// 收集所有角色ID，包括继承角色
	allRoleIDs := make(map[int64]struct{})

	// 先将直接角色加入集合
	for i := range directRoleIDs {
		allRoleIDs[directRoleIDs[i]] = struct{}{}
	}

	// 递归处理每个直接角色的包含关系
	for i := range directRoleIDs {
		if inclErr := s.collectIncludedRoles(ctx, bizID, directRoleIDs[i], allRoleIDs); inclErr != nil {
			return nil, inclErr
		}
	}

	// 将角色ID映射转换为切片
	result := make([]int64, 0, len(allRoleIDs))
	for roleID := range allRoleIDs {
		result = append(result, roleID)
	}
	return result, nil
}

// collectIncludedRoles 递归获取角色的包含关系
func (s *permissionService) collectIncludedRoles(
	ctx context.Context,
	bizID,
	includingRoleID int64,
	allRoleIDs map[int64]struct{},
) error {
	offset := 0
	for {
		inclusions, total, err := s.repo.FindRoleInclusionsByBizIDAndIncludingRoleID(ctx, bizID, includingRoleID, offset, s.limit)
		if err != nil {
			return err
		}
		for i := range inclusions {
			includedRoleID := inclusions[i].IncludedRole.ID
			if _, exists := allRoleIDs[includedRoleID]; !exists {
				allRoleIDs[includedRoleID] = struct{}{}
				// 递归处理包含的角色
				if inclErr := s.collectIncludedRoles(ctx, bizID, includedRoleID, allRoleIDs); inclErr != nil {
					return inclErr
				}
			}
		}
		// 优化分页边界条件判断
		if len(inclusions) < s.limit || int64(offset+s.limit) >= int64(total) {
			break
		}
		offset += s.limit
	}
	return nil
}

// getUserRoles 获取用户所有有效角色（处理分页）
func (s *permissionService) getUserRoles(
	ctx context.Context,
	bizID,
	userID,
	currentTime int64,
) ([]domain.UserRole, error) {
	offset := 0
	var allUserRoles []domain.UserRole

	for {
		userRoles, total, err := s.repo.FindValidUserRolesByBizIDAndUserID(ctx, bizID, userID, currentTime, offset, s.limit)
		if err != nil {
			return nil, err
		}
		allUserRoles = append(allUserRoles, userRoles...)

		// 优化分页边界条件判断
		if len(userRoles) < s.limit || int64(offset+s.limit) >= int64(total) {
			break
		}
		offset += s.limit
	}

	return allUserRoles, nil
}

// checkRolePermissions 检查是否有角色权限匹配目标权限
func (s *permissionService) checkRolePermissions(
	ctx context.Context,
	bizID int64,
	roleIDs []int64,
	permission domain.Permission,
) (bool, error) {
	// 获取所有角色权限
	rolePermissions, err := s.getRolePermissions(ctx, bizID, roleIDs)
	if err != nil {
		return false, err
	}
	// 检查是否有匹配的角色权限
	for i := range rolePermissions {
		if rolePermissions[i].Permission.Resource.Key == permission.Resource.Key &&
			rolePermissions[i].Permission.Action == permission.Action {
			return true, nil
		}
	}
	return false, nil
}

// getRolePermissions 获取所有角色的权限（处理分页）
func (s *permissionService) getRolePermissions(
	ctx context.Context,
	bizID int64,
	roleIDs []int64,
) ([]domain.RolePermission, error) {
	if len(roleIDs) == 0 {
		return []domain.RolePermission{}, nil
	}
	offset := 0
	var allRolePermissions []domain.RolePermission

	for {
		rolePermissions, total, err := s.repo.FindRolePermissionsByBizIDAndRoleIDs(ctx, bizID, roleIDs, offset, s.limit)
		if err != nil {
			return nil, err
		}
		allRolePermissions = append(allRolePermissions, rolePermissions...)

		// 优化分页边界条件判断
		if len(rolePermissions) < s.limit || int64(offset+s.limit) >= int64(total) {
			break
		}
		offset += s.limit
	}

	return allRolePermissions, nil
}
