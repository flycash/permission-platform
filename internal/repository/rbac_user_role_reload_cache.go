package repository

import (
	"context"

	"gitee.com/flycash/permission-platform/internal/domain"
	"github.com/gotomicro/ego/core/elog"
)

var _ UserRoleRepository = (*UserRoleReloadCacheRepository)(nil)

// UserRoleReloadCacheRepository 用户角色关系仓储实现
type UserRoleReloadCacheRepository struct {
	repo          *UserRoleDefaultRepository
	cacheReloader UserPermissionCacheReloader
	logger        *elog.Component
}

// NewUserRoleReloadCacheRepository 创建可以重载缓存的用户角色关系仓储实例
func NewUserRoleReloadCacheRepository(
	repo *UserRoleDefaultRepository,
	cacheReloader UserPermissionCacheReloader,
) *UserRoleReloadCacheRepository {
	return &UserRoleReloadCacheRepository{
		repo:          repo,
		cacheReloader: cacheReloader,
		logger:        elog.DefaultLogger.With(elog.FieldName("UserRoleReloadCacheRepository")),
	}
}

func (r *UserRoleReloadCacheRepository) Create(ctx context.Context, userRole domain.UserRole) (domain.UserRole, error) {
	created, err := r.repo.Create(ctx, userRole)
	if err != nil {
		return domain.UserRole{}, err
	}
	if err1 := r.cacheReloader.Reload(ctx, []domain.User{{ID: userRole.UserID, BizID: userRole.BizID}}); err1 != nil {
		r.logger.Warn("创建用户角色成功后，重新加载受影响用户的缓存失败",
			elog.FieldErr(err1),
			elog.Any("bizID", created.BizID),
			elog.Any("userID", created.UserID),
		)
	}
	return created, nil
}

func (r *UserRoleReloadCacheRepository) FindByBizIDAndUserID(ctx context.Context, bizID, userID int64) ([]domain.UserRole, error) {
	return r.repo.FindByBizIDAndUserID(ctx, bizID, userID)
}

func (r *UserRoleReloadCacheRepository) DeleteByBizIDAndID(ctx context.Context, bizID, id int64) error {
	deleted, err := r.repo.FindByBizIDAndID(ctx, bizID, id)
	if err != nil {
		return err
	}
	err = r.repo.DeleteByBizIDAndID(ctx, bizID, id)
	if err != nil {
		return err
	}
	if err1 := r.cacheReloader.Reload(ctx, []domain.User{{ID: deleted.UserID, BizID: deleted.BizID}}); err1 != nil {
		r.logger.Warn("删除用户角色成功后，重新加载受影响用户的缓存失败",
			elog.FieldErr(err1),
			elog.Any("bizID", bizID),
			elog.Any("userID", deleted.UserID),
		)
	}
	return nil
}

func (r *UserRoleReloadCacheRepository) FindByBizID(ctx context.Context, bizID int64) ([]domain.UserRole, error) {
	return r.repo.FindByBizID(ctx, bizID)
}
