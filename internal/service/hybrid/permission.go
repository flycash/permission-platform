package hybrid

import (
	"context"
	"strconv"

	"gitee.com/flycash/permission-platform/internal/domain"
	"gitee.com/flycash/permission-platform/internal/service/abac"
	"gitee.com/flycash/permission-platform/internal/service/rbac"
)

type PermissionService interface {
	Check(ctx context.Context, bizID, userID int64, resource domain.Resource, actions []string, attrs domain.Attributes) (bool, error)
}

type permissionService struct {
	rbacSvc rbac.PermissionService
	abacSvc abac.PermissionSvc
}

func NewPermissionService(
	rbacSvc rbac.PermissionService,
	abacSvc abac.PermissionSvc,
) PermissionService {
	return &permissionService{
		rbacSvc: rbacSvc,
		abacSvc: abacSvc,
	}
}

func (p *permissionService) Check(ctx context.Context, bizID, userID int64, resource domain.Resource, actions []string, attrs domain.Attributes) (bool, error) {
	ok, err := p.rbacSvc.Check(ctx, bizID, userID, resource, actions)
	if err != nil || !ok {
		return false, err
	}
	return p.abacSvc.Check(ctx, bizID, userID, resource, actions, attrs)
}

type roleAsAttributePermissionService struct {
	rbacSvc           rbac.Service
	abacPermissionSvc abac.PermissionSvc
}

func NewRoleAsAttributePermissionService(rbacSvc rbac.Service, abacPermissionSvc abac.PermissionSvc) PermissionService {
	return &roleAsAttributePermissionService{rbacSvc: rbacSvc, abacPermissionSvc: abacPermissionSvc}
}

func (p *roleAsAttributePermissionService) Check(ctx context.Context, bizID, userID int64, resource domain.Resource, actions []string, attrs domain.Attributes) (bool, error) {
	userRoles, err := p.rbacSvc.ListUserRolesByUserID(ctx, bizID, userID)
	if err != nil {
		return false, err
	}
	// 将用户角色作为普通主体属性
	for i := range userRoles {
		attrs.Subject[userRoles[i].Role.Name] = strconv.FormatInt(userRoles[i].Role.ID, 10)
	}
	return p.abacPermissionSvc.Check(ctx, bizID, userID, resource, actions, attrs)
}
