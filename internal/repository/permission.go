package repository

import (
	"context"

	"gitee.com/flycash/permission-platform/internal/domain"
	"gitee.com/flycash/permission-platform/internal/repository/dao"
	"github.com/ecodeclub/ekit/slice"
)

// PermissionRepository 权限仓储接口
type PermissionRepository interface {
	Create(ctx context.Context, permission domain.Permission) (domain.Permission, error)
	FindPermissions(ctx context.Context, bizID int64, resourceType, resourceKey, action string) ([]domain.Permission, error)
	FindByBizID(ctx context.Context, bizID int64, offset, limit int) ([]domain.Permission, error)
	FindByBizIDAndID(ctx context.Context, bizID, id int64) (domain.Permission, error)

	UpdateByBizIDAndID(ctx context.Context, permission domain.Permission) (domain.Permission, error)

	DeleteByBizIDAndID(ctx context.Context, bizID, id int64) error
}

// permissionRepository 权限仓储实现
type permissionRepository struct {
	permissionDAO dao.PermissionDAO
}

func (r *permissionRepository) FindPermissions(ctx context.Context, bizID int64, resourceType, resourceKey, action string) ([]domain.Permission, error) {
	permissions, err := r.permissionDAO.FindPermissions(ctx, bizID, resourceType, resourceKey, action)
	if err != nil {
		return nil, err
	}
	list := slice.Map(permissions, func(_ int, src dao.Permission) domain.Permission {
		return r.toDomain(src)
	})
	return list, nil
}

// NewPermissionRepository 创建权限仓储实例
func NewPermissionRepository(permissionDAO dao.PermissionDAO) PermissionRepository {
	return &permissionRepository{
		permissionDAO: permissionDAO,
	}
}

func (r *permissionRepository) Create(ctx context.Context, permission domain.Permission) (domain.Permission, error) {
	created, err := r.permissionDAO.Create(ctx, r.toEntity(permission))
	if err != nil {
		return domain.Permission{}, err
	}
	return r.toDomain(created), nil
}

func (r *permissionRepository) FindByBizIDAndID(ctx context.Context, bizID, id int64) (domain.Permission, error) {
	permission, err := r.permissionDAO.FindByBizIDAndID(ctx, bizID, id)
	if err != nil {
		return domain.Permission{}, err
	}
	return r.toDomain(permission), nil
}

func (r *permissionRepository) UpdateByBizIDAndID(ctx context.Context, permission domain.Permission) (domain.Permission, error) {
	err := r.permissionDAO.UpdateByBizIDAndID(ctx, r.toEntity(permission))
	if err != nil {
		return domain.Permission{}, err
	}
	return permission, nil
}

func (r *permissionRepository) DeleteByBizIDAndID(ctx context.Context, bizID, id int64) error {
	return r.permissionDAO.DeleteByBizIDAndID(ctx, bizID, id)
}

func (r *permissionRepository) FindByBizID(ctx context.Context, bizID int64, offset, limit int) ([]domain.Permission, error) {
	permissions, err := r.permissionDAO.FindByBizID(ctx, bizID, offset, limit)
	if err != nil {
		return nil, err
	}

	return slice.Map(permissions, func(_ int, src dao.Permission) domain.Permission {
		return r.toDomain(src)
	}), nil
}

func (r *permissionRepository) toEntity(p domain.Permission) dao.Permission {
	return dao.Permission{
		ID:           p.ID,
		BizID:        p.BizID,
		Name:         p.Name,
		Description:  p.Description,
		ResourceID:   p.Resource.ID,
		ResourceType: p.Resource.Type,
		ResourceKey:  p.Resource.Key,
		Action:       p.Action,
		Metadata:     p.Metadata,
		Ctime:        p.Ctime,
		Utime:        p.Utime,
	}
}

func (r *permissionRepository) toDomain(p dao.Permission) domain.Permission {
	return domain.Permission{
		ID:          p.ID,
		BizID:       p.BizID,
		Name:        p.Name,
		Description: p.Description,
		Resource: domain.Resource{
			ID:   p.ResourceID,
			Type: p.ResourceType,
			Key:  p.ResourceKey,
		},
		Action:   p.Action,
		Metadata: p.Metadata,
		Ctime:    p.Ctime,
		Utime:    p.Utime,
	}
}
