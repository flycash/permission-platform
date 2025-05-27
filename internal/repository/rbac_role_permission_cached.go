package repository

import (
	"context"

	"gitee.com/flycash/permission-platform/internal/domain"
	"gitee.com/flycash/permission-platform/internal/repository/cache"
	"github.com/gotomicro/ego/core/elog"
)

var _ RolePermissionRepository = (*RolePermissionCachedRepository)(nil)

// RolePermissionCachedRepository 角色权限关系仓储实现
type RolePermissionCachedRepository struct {
	repo   *RolePermissionDefaultRepository
	cache  cache.RolePermissionCache
	logger *elog.Component
}

// NewRolePermissionCachedRepository 创建角色权限关系仓储实例
func NewRolePermissionCachedRepository(
	repo *RolePermissionDefaultRepository,
	cache cache.RolePermissionCache,
) *RolePermissionCachedRepository {
	return &RolePermissionCachedRepository{
		repo:   repo,
		cache:  cache,
		logger: elog.DefaultLogger.With(elog.FieldName("RolePermissionCachedRepository")),
	}
}

func (r *RolePermissionCachedRepository) Create(ctx context.Context, rolePermission domain.RolePermission) (domain.RolePermission, error) {
	created, err := r.repo.Create(ctx, rolePermission)
	if err != nil {
		return domain.RolePermission{}, err
	}
	if err1 := r.cache.Add(ctx, []domain.RolePermission{created}); err1 != nil {
		r.logger.Warn("创建角色权限成功后，添加缓存失败",
			elog.FieldErr(err1),
			elog.Any("bizID", created.BizID),
			elog.Any("rolePermission", created),
		)
	}
	return created, nil
}

func (r *RolePermissionCachedRepository) FindByBizIDAndRoleIDs(ctx context.Context, bizID int64, roleIDs []int64) ([]domain.RolePermission, error) {
	perms, err := r.cache.Get(ctx, bizID, roleIDs...)
	if err == nil {
		return perms, nil
	}
	perms, err = r.repo.FindByBizIDAndRoleIDs(ctx, bizID, roleIDs)
	if err != nil {
		return nil, err
	}
	if err1 := r.cache.Set(ctx, perms); err1 != nil {
		r.logger.Warn("按BizID和角色ID集合查找角色权限成功后，设置缓存失败",
			elog.FieldErr(err1),
			elog.Any("bizID", bizID),
			elog.Any("rolePermissions", perms),
		)
	}
	return perms, nil
}

func (r *RolePermissionCachedRepository) DeleteByBizIDAndID(ctx context.Context, bizID, id int64) error {
	err := r.repo.DeleteByBizIDAndID(ctx, bizID, id)
	if err != nil {
		return err
	}
	if err1 := r.cache.Del(ctx, bizID, id); err1 != nil {
		r.logger.Warn("按BizID和ID删除角色权限成功后，删除相应缓存失败",
			elog.FieldErr(err1),
			elog.Any("bizID", bizID),
			elog.Any("id", id),
		)
	}
	return nil
}

func (r *RolePermissionCachedRepository) FindByBizID(ctx context.Context, bizID int64) ([]domain.RolePermission, error) {
	perms, err := r.cache.Get(ctx, bizID)
	if err == nil {
		return perms, nil
	}
	perms, err = r.repo.FindByBizID(ctx, bizID)
	if err != nil {
		return nil, err
	}
	if err1 := r.cache.Set(ctx, perms); err1 != nil {
		r.logger.Warn("按BizID查找角色权限成功后，设置缓存失败",
			elog.FieldErr(err1),
			elog.Any("bizID", bizID),
			elog.Any("rolePermissions", perms),
		)
	}
	return perms, nil
}
