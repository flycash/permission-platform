package hybrid

import (
	"context"
	"gitee.com/flycash/permission-platform/internal/domain"
	"gitee.com/flycash/permission-platform/internal/service/abac"
	"gitee.com/flycash/permission-platform/internal/service/abac/converter"
	"gitee.com/flycash/permission-platform/internal/service/rbac"
	"github.com/ecodeclub/ekit/slice"
)

const (
	defaultRoleName = "role"
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
	converter         converter.Converter[[]string]
	abacPermissionSvc abac.PermissionSvc
}

func NewRoleAsAttributePermissionService(rbacSvc rbac.Service, abacPermissionSvc abac.PermissionSvc) PermissionService {
	return &roleAsAttributePermissionService{
		rbacSvc:           rbacSvc,
		abacPermissionSvc: abacPermissionSvc,
		converter:         converter.NewArrayConverter(),
	}
}

func (p *roleAsAttributePermissionService) Check(ctx context.Context, bizID, userID int64, resource domain.Resource, actions []string, attrs domain.Attributes) (bool, error) {
	userRoles, err := p.rbacSvc.ListUserRolesByUserID(ctx, bizID, userID)
	if err != nil {
		return false, err
	}
	nameList := slice.Map(userRoles, func(_ int, src domain.UserRole) string {
		return src.Role.Name
	})
	val, _ := p.converter.Encode(nameList)
	attrs.Subject[defaultRoleName] = val
	return p.abacPermissionSvc.Check(ctx, bizID, userID, resource, actions, attrs)
}
