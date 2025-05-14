package repository

import (
	"context"

	"gitee.com/flycash/permission-platform/internal/domain"
	"gitee.com/flycash/permission-platform/internal/repository/dao"
	"github.com/ecodeclub/ekit/slice"
)

// UserPermissionRepository 用户权限关系仓储接口
type UserPermissionRepository interface {
	Create(ctx context.Context, userPermission domain.UserPermission) (domain.UserPermission, error)
	FindByBizIDAndUserID(ctx context.Context, bizID, userID int64, offset, limit int) ([]domain.UserPermission, error)
	FindValidByBizIDAndUserID(ctx context.Context, bizID, userID, currentTime int64, offset, limit int) ([]domain.UserPermission, error)
	DeleteByBizIDAndID(ctx context.Context, bizID, id int64) error
	FindByBizID(ctx context.Context, bizID int64, offset, limit int) ([]domain.UserPermission, error)
}

// userPermissionRepository 用户权限关系仓储实现
type userPermissionRepository struct {
	userPermissionDAO dao.UserPermissionDAO
}

// NewUserPermissionRepository 创建用户权限关系仓储实例
func NewUserPermissionRepository(userPermissionDAO dao.UserPermissionDAO) UserPermissionRepository {
	return &userPermissionRepository{
		userPermissionDAO: userPermissionDAO,
	}
}

func (r *userPermissionRepository) Create(ctx context.Context, userPermission domain.UserPermission) (domain.UserPermission, error) {
	created, err := r.userPermissionDAO.Create(ctx, r.toEntity(userPermission))
	if err != nil {
		return domain.UserPermission{}, err
	}
	return r.toDomain(created), nil
}

func (r *userPermissionRepository) FindByBizIDAndUserID(ctx context.Context, bizID, userID int64, offset, limit int) ([]domain.UserPermission, error) {
	userPermissions, err := r.userPermissionDAO.FindByBizIDAndUserID(ctx, bizID, userID, offset, limit)
	if err != nil {
		return nil, err
	}

	return slice.Map(userPermissions, func(_ int, src dao.UserPermission) domain.UserPermission {
		return r.toDomain(src)
	}), nil
}

func (r *userPermissionRepository) FindValidByBizIDAndUserID(ctx context.Context, bizID, userID, currentTime int64, offset, limit int) ([]domain.UserPermission, error) {
	userPermissions, err := r.userPermissionDAO.FindValidPermissionsWithBizID(ctx, bizID, userID, currentTime, offset, limit)
	if err != nil {
		return nil, err
	}

	return slice.Map(userPermissions, func(_ int, src dao.UserPermission) domain.UserPermission {
		return r.toDomain(src)
	}), nil
}

func (r *userPermissionRepository) DeleteByBizIDAndID(ctx context.Context, bizID, id int64) error {
	return r.userPermissionDAO.DeleteByBizIDAndID(ctx, bizID, id)
}

func (r *userPermissionRepository) FindByBizID(ctx context.Context, bizID int64, offset, limit int) ([]domain.UserPermission, error) {
	userPermissions, err := r.userPermissionDAO.FindByBizID(ctx, bizID, offset, limit)
	if err != nil {
		return nil, err
	}

	return slice.Map(userPermissions, func(_ int, src dao.UserPermission) domain.UserPermission {
		return r.toDomain(src)
	}), nil
}

func (r *userPermissionRepository) toEntity(up domain.UserPermission) dao.UserPermission {
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

func (r *userPermissionRepository) toDomain(up dao.UserPermission) domain.UserPermission {
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
