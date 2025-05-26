package repository

import (
	"context"

	"gitee.com/flycash/permission-platform/internal/domain"
	"gitee.com/flycash/permission-platform/internal/repository/cache"
	"github.com/gotomicro/ego/core/elog"
)

var (
	_ UserPermissionRepository    = (*UserPermissionCachedRepository)(nil)
	_ UserPermissionCacheReloader = (*UserPermissionCachedRepository)(nil)
)

type UserPermissionCacheReloader interface {
	Reload(ctx context.Context, users []domain.User) error
}

type UserPermissionCachedRepository struct {
	repo   *UserPermissionDefaultRepository
	cache  cache.UserPermissionCache
	logger *elog.Component
}

// NewUserPermissionCachedRepository 添加了缓存的仓储
func NewUserPermissionCachedRepository(
	repo *UserPermissionDefaultRepository,
	cache cache.UserPermissionCache,
) *UserPermissionCachedRepository {
	return &UserPermissionCachedRepository{
		repo:   repo,
		cache:  cache,
		logger: elog.DefaultLogger.With(elog.FieldName("UserPermissionCachedRepository")),
	}
}

func (r *UserPermissionCachedRepository) Create(ctx context.Context, userPermission domain.UserPermission) (domain.UserPermission, error) {
	created, err := r.repo.Create(ctx, userPermission)
	if err != nil {
		return domain.UserPermission{}, err
	}
	if err1 := r.Reload(ctx, []domain.User{{ID: created.UserID, BizID: created.BizID}}); err1 != nil {
		r.logger.Warn("创建用户权限成功后重新加载缓存失败",
			elog.FieldErr(err1),
			elog.Any("bizID", created.BizID),
			elog.Any("userID", created.UserID),
		)
	}
	return created, err
}

func (r *UserPermissionCachedRepository) Reload(ctx context.Context, users []domain.User) error {
	for i := range users {
		perms, err := r.repo.GetAllUserPermissions(ctx, users[i].BizID, users[i].ID)
		if err != nil {
			return err
		}
		err = r.cache.Set(ctx, perms)
		if err != nil {
			r.logger.Warn("重新加载用户全部权限到缓存失败",
				elog.FieldErr(err),
				elog.Any("bizID", users[i].BizID),
				elog.Any("userID", users[i].ID),
			)
		}
	}
	return nil
}

func (r *UserPermissionCachedRepository) FindByBizID(ctx context.Context, bizID int64, offset, limit int) ([]domain.UserPermission, error) {
	return r.repo.FindByBizID(ctx, bizID, offset, limit)
}

func (r *UserPermissionCachedRepository) FindByBizIDAndUserID(ctx context.Context, bizID, userID int64) ([]domain.UserPermission, error) {
	perms, err := r.cache.Get(ctx, bizID, userID)
	if err == nil {
		return perms, nil
	}
	perms, err = r.repo.FindByBizIDAndUserID(ctx, bizID, userID)
	if err != nil {
		return nil, err
	}
	if err1 := r.cache.Set(ctx, perms); err1 != nil {
		r.logger.Warn("查找用户权限成功后重新设置缓存失败",
			elog.FieldErr(err1),
			elog.Any("bizID", bizID),
			elog.Any("userID", userID),
		)
	}
	return perms, err
}

func (r *UserPermissionCachedRepository) DeleteByBizIDAndID(ctx context.Context, bizID, id int64) error {
	deleted, err := r.repo.FindByBizIDAndID(ctx, bizID, id)
	if err != nil {
		return err
	}
	err = r.repo.DeleteByBizIDAndID(ctx, bizID, id)
	if err != nil {
		return err
	}
	if err1 := r.Reload(ctx, []domain.User{{ID: deleted.UserID, BizID: deleted.BizID}}); err1 != nil {
		r.logger.Warn("删除用户权限成功后重新加载缓存失败",
			elog.FieldErr(err1),
			elog.Any("bizID", bizID),
			elog.Any("userID", deleted.UserID),
		)
	}
	return nil
}

func (r *UserPermissionCachedRepository) GetAllUserPermissions(ctx context.Context, bizID, userID int64) ([]domain.UserPermission, error) {
	perms, err := r.cache.Get(ctx, bizID, userID)
	if err == nil {
		return perms, nil
	}

	perms, err = r.repo.GetAllUserPermissions(ctx, bizID, userID)
	if err != nil {
		r.logger.Error("从数据库中查找用户全部权限失败",
			elog.FieldErr(err),
			elog.Any("bizID", bizID),
			elog.Any("userID", userID),
		)
		return nil, err
	}

	if err1 := r.cache.Set(ctx, perms); err1 != nil {
		r.logger.Warn("存储用户全部权限到缓存失败",
			elog.FieldErr(err1),
			elog.Any("bizID", bizID),
			elog.Any("userID", userID),
		)
	}
	return perms, err
}
