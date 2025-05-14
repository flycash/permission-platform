package rbac

import (
	"context"

	"gitee.com/flycash/permission-platform/internal/domain"
	"gitee.com/flycash/permission-platform/internal/repository"
)

// PermissionService RBAC模型下的权限服务接口
type PermissionService interface {
	// Check 检查用户是否有对特定权限
	Check(ctx context.Context, bizID, userID int64, permission domain.Permission) (bool, error)
}

type permissionService struct {
	repo repository.RBACRepository
}

// NewPermissionService 创建RBAC权限服务实例
func NewPermissionService(repo repository.RBACRepository) PermissionService {
	return &permissionService{
		repo: repo,
	}
}

// Check 检查用户权限
func (s *permissionService) Check(ctx context.Context, bizID, userID int64, permission domain.Permission) (bool, error) {
	// 获取用户所有权限（包括直接权限和通过角色获得的权限）
	permissions, err := s.repo.GetAllUserPermissions(ctx, bizID, userID)
	if err != nil {
		return false, err
	}
	// 如果没有任何权限，则无权限
	if len(permissions) == 0 {
		return false, nil
	}
	// 检查是否有匹配的权限
	for i := range permissions {
		// 匹配资源类型、资源键和操作
		if permissions[i].Permission.Resource.Type == permission.Resource.Type &&
			permissions[i].Permission.Resource.Key == permission.Resource.Key &&
			permissions[i].Permission.Action == permission.Action {

			// 如果有拒绝权限，直接返回拒绝
			if permissions[i].Effect.IsDeny() {
				return false, nil
			}
			// 找到匹配的允许权限
			return true, nil
		}
	}
	// 没有找到匹配的权限，返回无权限
	return false, nil
}
