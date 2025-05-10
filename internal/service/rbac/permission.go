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
	return s.repo.CheckUserPermission(ctx, bizID, userID, permission)
}
