package repository

import (
	"context"
	"time"

	"gitee.com/flycash/permission-platform/internal/domain"
	"gitee.com/flycash/permission-platform/internal/repository/dao"
	"github.com/ecodeclub/ekit/mapx"
	"github.com/ecodeclub/ekit/slice"
)

const (
	oneHundredYears = 100
)

var _ UserPermissionRepository = (*UserPermissionDefaultRepository)(nil)

// UserPermissionRepository 用户权限关系仓储接口
type UserPermissionRepository interface {
	Create(ctx context.Context, userPermission domain.UserPermission) (domain.UserPermission, error)

	FindByBizID(ctx context.Context, bizID int64, offset, limit int) ([]domain.UserPermission, error)
	FindByBizIDAndUserID(ctx context.Context, bizID, userID int64) ([]domain.UserPermission, error)

	DeleteByBizIDAndID(ctx context.Context, bizID, id int64) error

	// GetAll 获取用户的所有权限，包括个人权限、个人拥有的角色（及包含的角色）对应的权限
	GetAll(ctx context.Context, bizID, userID int64) ([]domain.UserPermission, error)
}

// UserPermissionDefaultRepository 用户权限关系仓储实现
type UserPermissionDefaultRepository struct {
	roleInclusionDAO  dao.RoleInclusionDAO
	rolePermissionDAO dao.RolePermissionDAO
	userRoleDAO       dao.UserRoleDAO
	userPermissionDAO dao.UserPermissionDAO
}

// NewUserPermissionDefaultRepository 创建用户权限关系仓储实例
func NewUserPermissionDefaultRepository(
	roleInclusionDAO dao.RoleInclusionDAO,
	rolePermissionDAO dao.RolePermissionDAO,
	userRoleDAO dao.UserRoleDAO,
	userPermissionDAO dao.UserPermissionDAO,
) *UserPermissionDefaultRepository {
	return &UserPermissionDefaultRepository{
		roleInclusionDAO:  roleInclusionDAO,
		rolePermissionDAO: rolePermissionDAO,
		userRoleDAO:       userRoleDAO,
		userPermissionDAO: userPermissionDAO,
	}
}

func (r *UserPermissionDefaultRepository) Create(ctx context.Context, userPermission domain.UserPermission) (domain.UserPermission, error) {
	created, err := r.userPermissionDAO.Create(ctx, r.toEntity(userPermission))
	if err != nil {
		return domain.UserPermission{}, err
	}
	return r.toDomain(created), nil
}

func (r *UserPermissionDefaultRepository) FindByBizIDAndUserID(ctx context.Context, bizID, userID int64) ([]domain.UserPermission, error) {
	userPermissions, err := r.userPermissionDAO.FindByBizIDAndUserID(ctx, bizID, userID)
	if err != nil {
		return nil, err
	}

	return slice.Map(userPermissions, func(_ int, src dao.UserPermission) domain.UserPermission {
		return r.toDomain(src)
	}), nil
}

func (r *UserPermissionDefaultRepository) FindByBizIDAndID(ctx context.Context, bizID, id int64) (domain.UserPermission, error) {
	up, err := r.userPermissionDAO.FindByBizIDANDID(ctx, bizID, id)
	if err != nil {
		return domain.UserPermission{}, err
	}
	return r.toDomain(up), nil
}

func (r *UserPermissionDefaultRepository) DeleteByBizIDAndID(ctx context.Context, bizID, id int64) error {
	return r.userPermissionDAO.DeleteByBizIDAndID(ctx, bizID, id)
}

func (r *UserPermissionDefaultRepository) FindByBizID(ctx context.Context, bizID int64, offset, limit int) ([]domain.UserPermission, error) {
	userPermissions, err := r.userPermissionDAO.FindByBizID(ctx, bizID, offset, limit)
	if err != nil {
		return nil, err
	}

	return slice.Map(userPermissions, func(_ int, src dao.UserPermission) domain.UserPermission {
		return r.toDomain(src)
	}), nil
}

func (r *UserPermissionDefaultRepository) toEntity(up domain.UserPermission) dao.UserPermission {
	return dao.UserPermission{
		ID:               up.ID,
		BizID:            up.BizID,
		UserID:           up.UserID,
		PermissionID:     up.Permission.ID,
		PermissionName:   up.Permission.Name,
		ResourceType:     up.Permission.Resource.Type,
		ResourceKey:      up.Permission.Resource.Key,
		PermissionAction: up.Permission.Action,
		StartTime:        up.StartTime,
		EndTime:          up.EndTime,
		Effect:           up.Effect.String(),
		Ctime:            up.Ctime,
		Utime:            up.Utime,
	}
}

func (r *UserPermissionDefaultRepository) toDomain(up dao.UserPermission) domain.UserPermission {
	return domain.UserPermission{
		ID:     up.ID,
		BizID:  up.BizID,
		UserID: up.UserID,
		Permission: domain.Permission{
			ID:   up.PermissionID,
			Name: up.PermissionName,
			Resource: domain.Resource{
				Type: up.ResourceType,
				Key:  up.ResourceKey,
			},
			Action: up.PermissionAction,
		},
		StartTime: up.StartTime,
		EndTime:   up.EndTime,
		Effect:    domain.Effect(up.Effect),
		Ctime:     up.Ctime,
		Utime:     up.Utime,
	}
}

func (r *UserPermissionDefaultRepository) GetAll(ctx context.Context, bizID, userID int64) ([]domain.UserPermission, error) {
	permissions, err := r.userPermissionDAO.FindByBizIDAndUserID(ctx, bizID, userID)
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

	perms := slice.Map(permissions, func(_ int, src dao.UserPermission) domain.UserPermission {
		return r.toDomain(src)
	})
	return append(perms, rolePerms...), nil
}

// getAllRoleIDs 获取用户所有角色ID，包括继承的角色
func (r *UserPermissionDefaultRepository) getAllRoleIDs(ctx context.Context, bizID, userID int64) ([]int64, error) {
	// 1. 先找到直接关联的角色
	directRoles, err := r.userRoleDAO.FindByBizIDAndUserID(ctx, bizID, userID)
	if err != nil {
		return nil, err
	}

	allRoleIDs := make(map[int64]any, len(directRoles))

	directRoleIDs := slice.Map(directRoles, func(_ int, src dao.UserRole) int64 {
		allRoleIDs[src.RoleID] = struct{}{}
		return src.RoleID
	})

	includedIDs := directRoleIDs

	for {
		// 找到 directRoles 包含的角色
		inclusions, err := r.roleInclusionDAO.FindByBizIDAndIncludingRoleIDs(ctx, bizID, includedIDs)
		if err != nil {
			return nil, err
		}
		if len(inclusions) == 0 {
			break
		}
		includedIDs = slice.Map(inclusions, func(_ int, src dao.RoleInclusion) int64 {
			allRoleIDs[src.IncludedRoleID] = struct{}{}
			return src.IncludedRoleID
		})
	}

	// 你要根据 inclusions 里面的 IncludedRoleID 进步沿着包含链找下去
	return mapx.Keys(allRoleIDs), nil
}

// getRolePermissions 获取指定角色ID列表对应的所有权限
func (r *UserPermissionDefaultRepository) getRolePermissions(ctx context.Context, bizID, userID int64, roleIDs []int64) ([]domain.UserPermission, error) {
	if len(roleIDs) == 0 {
		return []domain.UserPermission{}, nil
	}
	permissions, err := r.rolePermissionDAO.FindByBizIDAndRoleIDs(ctx, bizID, roleIDs)
	if err != nil {
		return nil, err
	}
	// 将RolePermission转换为UserPermission格式
	return slice.Map(permissions, func(_ int, src dao.RolePermission) domain.UserPermission {
		return domain.UserPermission{
			ID:     0,
			BizID:  bizID,
			UserID: userID,
			Permission: domain.Permission{
				ID:    src.PermissionID,
				BizID: src.BizID,
				Resource: domain.Resource{
					BizID: src.BizID,
					Type:  src.ResourceType,
					Key:   src.ResourceKey,
				},
				Action: src.PermissionAction,
			},
			StartTime: time.Now().UnixMilli(),
			EndTime:   time.Now().AddDate(oneHundredYears, 0, 0).UnixMilli(),
			Effect:    domain.EffectAllow,
			Ctime:     src.Ctime,
			Utime:     src.Utime,
		}
	}), nil
}
