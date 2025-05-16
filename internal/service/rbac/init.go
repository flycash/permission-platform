package rbac

import (
	"context"
	"fmt"
	"time"

	jwtauth "gitee.com/flycash/permission-platform/internal/api/grpc/interceptor/jwt"
	"gitee.com/flycash/permission-platform/internal/domain"
	"gitee.com/flycash/permission-platform/internal/repository"
	"github.com/golang-jwt/jwt/v4"
)

type SystemTableResource string

const (
	BusinessConfigTable SystemTableResource = "business_configs"
	ResourceTable       SystemTableResource = "resources"
	PermissionTable     SystemTableResource = "permissions"
	RoleTable           SystemTableResource = "roles"
	RoleInclusionTable  SystemTableResource = "role_inclusions"
	RolePermissionTable SystemTableResource = "role_permissions"
	UserRoleTable       SystemTableResource = "user_roles"
	UserPermissionTable SystemTableResource = "user_permissions"
)

func (rk SystemTableResource) Type() string {
	return "system_table"
}

func (rk SystemTableResource) Key() string {
	return fmt.Sprintf("/admin/%s", rk)
}

func (rk SystemTableResource) String() string {
	return string(rk)
}

type AccountResource string

const (
	ManagerAccountResource AccountResource = "account"
)

func (a AccountResource) String() string {
	return string(a)
}

func (a AccountResource) Type() string {
	return "admin_account"
}

func (a AccountResource) Key() string {
	return "/admin/account"
}

type PermissionActionType string

const (
	PermissionActionWrite PermissionActionType = "write"
	PermissionActionRead  PermissionActionType = "read"
)

func (a PermissionActionType) String() string {
	return string(a)
}

const DefaultAccountRoleType = "admin_account"

type InitService struct {
	bizID      int64
	userID     int64
	rateLimit  int
	jwtAuthKey string
	repo       repository.RBACRepository
}

func NewInitService(bizID, userID int64, rateLimit int, jwtAuthKey string, repo repository.RBACRepository) *InitService {
	return &InitService{
		bizID:      bizID,
		userID:     userID,
		rateLimit:  rateLimit,
		jwtAuthKey: jwtAuthKey,
		repo:       repo,
	}
}

func (s *InitService) Init(ctx context.Context) error {
	if err := s.createSystemBusinessConfig(ctx); err != nil {
		return err
	}

	resources, err := s.createSystemResources(ctx)
	if err != nil {
		return err
	}

	permissions, err := s.createPermissionsForSystemResources(ctx, resources)
	if err != nil {
		return err
	}

	role, err := s.createSystemAdminRole(ctx)
	if err != nil {
		return err
	}

	err = s.grantRolePermissions(ctx, role, permissions)
	if err != nil {
		return err
	}

	return s.grantUserRole(ctx, role)
}

func (s *InitService) createSystemBusinessConfig(ctx context.Context) error {
	// 生成Token
	const years = 100
	auth := jwtauth.NewJwtAuth(s.jwtAuthKey)
	token, err := auth.Encode(jwt.MapClaims{
		jwtauth.BizIDName: s.bizID,
		"exp":             time.Now().AddDate(years, 0, 0).Unix(),
	})
	if err != nil {
		return fmt.Errorf("生成Token失败: %w", err)
	}
	// 创建业务配置
	bizConfig, err := s.repo.BusinessConfig().Create(ctx, domain.BusinessConfig{
		ID:        s.bizID,
		OwnerID:   s.userID,
		OwnerType: "organization",
		Name:      "权限平台管理后台",
		RateLimit: s.rateLimit,
		Token:     token,
	})
	if err != nil {
		return fmt.Errorf("创建业务配置失败：%w", err)
	}
	if bizConfig.ID != s.bizID {
		return fmt.Errorf("业务配置ID不符合预期：id = %d", bizConfig.ID)
	}
	return nil
}

func (s *InitService) createSystemResources(ctx context.Context) ([]domain.Resource, error) {
	// 将管理平台的8张表，作为业务内部资源初始化，但使用预定义的Type、Key和Name
	systemResources := []SystemTableResource{
		BusinessConfigTable,
		ResourceTable,
		PermissionTable,
		RoleTable,
		RoleInclusionTable,
		RolePermissionTable,
		UserRoleTable,
		UserPermissionTable,
	}
	resources := make([]domain.Resource, 0, len(systemResources)+1)
	for i := range systemResources {
		res, err := s.repo.Resource().Create(ctx, domain.Resource{
			BizID: s.bizID,
			Type:  systemResources[i].Type(),
			Key:   systemResources[i].Key(),
			Name:  systemResources[i].String(),
		})
		if err != nil {
			return nil, err
		}
		resources = append(resources, res)
	}

	// ”账号管理“资源也作为业务内部资源初始化，但使用预定义的Type、Key和Name
	res, err := s.repo.Resource().Create(ctx, domain.Resource{
		BizID: s.bizID,
		Type:  ManagerAccountResource.Type(),
		Key:   ManagerAccountResource.Key(),
		Name:  ManagerAccountResource.String(),
	})
	if err != nil {
		return nil, err
	}
	resources = append(resources, res)

	return resources, nil
}

func (s *InitService) createPermissionsForSystemResources(ctx context.Context, resources []domain.Resource) ([]domain.Permission, error) {
	systemResourcePermissions := []PermissionActionType{
		PermissionActionRead,
		PermissionActionWrite,
	}
	permissions := make([]domain.Permission, 0, len(resources)*len(systemResourcePermissions))
	for i := range resources {
		// 为每个资源赋予预定义的权限
		for j := range systemResourcePermissions {
			res, err := s.repo.Permission().Create(ctx, domain.Permission{
				BizID:       s.bizID,
				Name:        fmt.Sprintf("%s-%s", resources[i].Name, systemResourcePermissions[j].String()),
				Description: fmt.Sprintf("%s-%s", resources[i].Name, systemResourcePermissions[j].String()),
				Resource:    resources[i],
				Action:      systemResourcePermissions[j].String(),
			})
			if err != nil {
				return nil, err
			}
			permissions = append(permissions, res)
		}
	}
	return permissions, nil
}

func (s *InitService) createSystemAdminRole(ctx context.Context) (domain.Role, error) {
	return s.repo.Role().Create(ctx, domain.Role{
		BizID:       s.bizID,
		Type:        DefaultAccountRoleType,
		Name:        "权限平台管理后台系统管理员",
		Description: "具有权限平台管理后台内最高管理权限",
	})
}

func (s *InitService) grantRolePermissions(ctx context.Context, role domain.Role, permissions []domain.Permission) error {
	for i := range permissions {
		_, err := s.repo.RolePermission().Create(ctx, domain.RolePermission{
			BizID:      s.bizID,
			Role:       role,
			Permission: permissions[i],
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *InitService) grantUserRole(ctx context.Context, role domain.Role) error {
	const years = 100
	_, err := s.repo.UserRole().Create(ctx, domain.UserRole{
		BizID:     s.bizID,
		UserID:    s.userID,
		Role:      role,
		StartTime: time.Now().UnixMilli(),
		EndTime:   time.Now().AddDate(years, 0, 0).UnixMilli(),
	})
	return err
}
