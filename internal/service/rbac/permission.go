package rbac

import (
	"context"
	"github.com/ecodeclub/ekit/slice"

	"gitee.com/flycash/permission-platform/internal/domain"
	"gitee.com/flycash/permission-platform/internal/repository"
)

// PermissionService RBAC模型下的权限服务接口
type PermissionService interface {
	// Check 检查用户是否有对特定权限
	Check(ctx context.Context, bizID, userID int64, resource domain.Resource, actions []string) (bool, error)
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
func (s *permissionService) Check(ctx context.Context, bizID, userID int64, resource domain.Resource, actions []string) (bool, error) {
	// 拿到用户的所有权限，看一下有没有我需要的权限
	// allow or deny
	permissions, err := s.repo.GetAllUserPermissions(ctx, bizID, userID)
	if err != nil {
		return false, err
	}
	var res bool
	for i := range permissions {
		p := permissions[i]
		pr := p.Permission.Resource
		if pr.Key == resource.Key && pr.Type == resource.Type &&
			slice.Contains(actions, p.Permission.Action) {
			// 找到了 resource，找到了 action
			// 负权限
			if p.Effect.IsDeny() {
				return false, nil
			}
			res = true
		}
	}
	return res, nil
}
