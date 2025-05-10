package rbac

import (
	"context"

	"gitee.com/flycash/permission-platform/internal/domain"
	"gitee.com/flycash/permission-platform/internal/repository"
)

// RBACService RBAC模型的管理接口
type RBACService interface {
	// Role相关方法
	CreateRole(ctx context.Context, role domain.Role) (domain.Role, error)
	GetRole(ctx context.Context, id int64) (domain.Role, error)
	UpdateRole(ctx context.Context, role domain.Role) (domain.Role, error)
	DeleteRole(ctx context.Context, id int64) error
	ListRoles(ctx context.Context, bizID int64, offset, limit int, roleType string) ([]domain.Role, int, error)

	// Resource相关方法
	CreateResource(ctx context.Context, resource domain.Resource) (domain.Resource, error)
	GetResource(ctx context.Context, id int64) (domain.Resource, error)
	UpdateResource(ctx context.Context, resource domain.Resource) (domain.Resource, error)
	DeleteResource(ctx context.Context, id int64) error
	ListResources(ctx context.Context, bizID int64, offset, limit int, resourceType, key string) ([]domain.Resource, int, error)

	// Permission相关方法
	CreatePermission(ctx context.Context, permission domain.Permission) (domain.Permission, error)
	GetPermission(ctx context.Context, id int64) (domain.Permission, error)
	UpdatePermission(ctx context.Context, permission domain.Permission) (domain.Permission, error)
	DeletePermission(ctx context.Context, id int64) error
	ListPermissions(ctx context.Context, bizID int64, offset, limit int, resourceType, resourceKey, action string) ([]domain.Permission, int, error)

	// 用户角色相关方法
	GrantUserRole(ctx context.Context, bizID, userID, roleID, startTime, endTime int64) (domain.UserRole, error)
	RevokeUserRole(ctx context.Context, bizID, userID, roleID int64) error
	ListUserRoles(ctx context.Context, bizID, userID int64, offset, limit int) ([]domain.UserRole, int, error)
}

type rbacService struct {
	repo repository.RBACRepository
}

// NewRBACService 创建RBAC服务实例
func NewRBACService(repo repository.RBACRepository) RBACService {
	return &rbacService{
		repo: repo,
	}
}

// Role相关方法实现
func (s *rbacService) CreateRole(ctx context.Context, role domain.Role) (domain.Role, error) {
	// 调用仓储层创建角色
	return s.repo.CreateRole(ctx, role)
}

func (s *rbacService) GetRole(ctx context.Context, id int64) (domain.Role, error) {
	// 调用仓储层获取角色
	return s.repo.GetRole(ctx, id)
}

func (s *rbacService) UpdateRole(ctx context.Context, role domain.Role) (domain.Role, error) {
	// 调用仓储层更新角色
	return s.repo.UpdateRole(ctx, role)
}

func (s *rbacService) DeleteRole(ctx context.Context, id int64) error {
	// 调用仓储层删除角色
	return s.repo.DeleteRole(ctx, id)
}

func (s *rbacService) ListRoles(ctx context.Context, bizID int64, offset, limit int, roleType string) ([]domain.Role, int, error) {
	// 调用仓储层获取角色列表
	return s.repo.ListRoles(ctx, bizID, offset, limit, roleType)
}

// Resource相关方法实现
func (s *rbacService) CreateResource(ctx context.Context, resource domain.Resource) (domain.Resource, error) {
	// 调用仓储层创建资源
	return s.repo.CreateResource(ctx, resource)
}

func (s *rbacService) GetResource(ctx context.Context, id int64) (domain.Resource, error) {
	// 调用仓储层获取资源
	return s.repo.GetResource(ctx, id)
}

func (s *rbacService) UpdateResource(ctx context.Context, resource domain.Resource) (domain.Resource, error) {
	// 调用仓储层更新资源
	return s.repo.UpdateResource(ctx, resource)
}

func (s *rbacService) DeleteResource(ctx context.Context, id int64) error {
	// 调用仓储层删除资源
	return s.repo.DeleteResource(ctx, id)
}

func (s *rbacService) ListResources(ctx context.Context, bizID int64, offset, limit int, resourceType, key string) ([]domain.Resource, int, error) {
	// 调用仓储层获取资源列表
	return s.repo.ListResources(ctx, bizID, offset, limit, resourceType, key)
}

// Permission相关方法实现
func (s *rbacService) CreatePermission(ctx context.Context, permission domain.Permission) (domain.Permission, error) {
	// 调用仓储层创建权限
	return s.repo.CreatePermission(ctx, permission)
}

func (s *rbacService) GetPermission(ctx context.Context, id int64) (domain.Permission, error) {
	return s.repo.GetPermission(ctx, id)
}

func (s *rbacService) UpdatePermission(ctx context.Context, permission domain.Permission) (domain.Permission, error) {
	// 调用仓储层更新权限
	return s.repo.UpdatePermission(ctx, permission)
}

func (s *rbacService) DeletePermission(ctx context.Context, id int64) error {
	// 调用仓储层删除权限
	return s.repo.DeletePermission(ctx, id)
}

func (s *rbacService) ListPermissions(ctx context.Context, bizID int64, offset, limit int, resourceType, resourceKey, action string) ([]domain.Permission, int, error) {
	// 调用仓储层获取权限列表
	return s.repo.ListPermissions(ctx, bizID, offset, limit, resourceType, resourceKey, action)
}

// 用户角色相关方法实现
func (s *rbacService) GrantUserRole(ctx context.Context, bizID, userID, roleID, startTime, endTime int64) (domain.UserRole, error) {
	// 调用仓储层授予用户角色
	return s.repo.GrantUserRole(ctx, bizID, userID, roleID, startTime, endTime)
}

func (s *rbacService) RevokeUserRole(ctx context.Context, bizID, userID, roleID int64) error {
	// 调用仓储层撤销用户角色
	return s.repo.RevokeUserRole(ctx, bizID, userID, roleID)
}

func (s *rbacService) ListUserRoles(ctx context.Context, bizID, userID int64, offset, limit int) ([]domain.UserRole, int, error) {
	// 调用仓储层获取用户角色列表
	return s.repo.ListUserRoles(ctx, bizID, userID, offset, limit)
}
