package repository

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"gitee.com/flycash/permission-platform/internal/domain"
	"gitee.com/flycash/permission-platform/pkg/cache"
	"github.com/gotomicro/ego/core/elog"
)

type CachedRBACRepository struct {
	repo     *DefaultRBACRepository
	cache    cache.Cache
	cacheKey func(bizID, userID int64) string
	logger   *elog.Component
}

// NewCachedRBACRepository 添加了缓存的仓储
func NewCachedRBACRepository(
	repo *DefaultRBACRepository,
	cache cache.Cache,
	cacheKey func(bizID, userID int64) string,
) *CachedRBACRepository {
	return &CachedRBACRepository{
		repo:     repo,
		cache:    cache,
		cacheKey: cacheKey,
		logger:   elog.DefaultLogger,
	}
}

func (r *CachedRBACRepository) BusinessConfig() BusinessConfigRepository {
	return r.repo.businessConfigRepo
}

func (r *CachedRBACRepository) Resource() ResourceRepository {
	return r.repo.resourceRepo
}

func (r *CachedRBACRepository) Permission() PermissionRepository {
	return r.repo.permissionRepo
}

func (r *CachedRBACRepository) Role() RoleRepository {
	return r.repo.roleRepo
}

func (r *CachedRBACRepository) RoleInclusion() RoleInclusionRepository {
	return r.repo.roleInclusionRepo
}

func (r *CachedRBACRepository) RolePermission() RolePermissionRepository {
	return r.repo.rolePermissionRepo
}

func (r *CachedRBACRepository) UserRole() UserRoleRepository {
	return r.repo.userRoleRepo
}

func (r *CachedRBACRepository) UserPermission() UserPermissionRepository {
	return r.repo.userPermissionRepo
}

func (r *CachedRBACRepository) GetAllUserPermissions(ctx context.Context, bizID, userID int64) ([]domain.UserPermission, error) {
	res, err := r.getFromCache(ctx, bizID, userID)
	if err == nil {
		return res, nil
	}

	res, err = r.repo.GetAllUserPermissions(ctx, bizID, userID)
	if err != nil {
		r.logger.Error("从数据库中查找用户权限",
			elog.FieldErr(err),
			elog.FieldCustomKeyValue("bizID", strconv.FormatInt(bizID, 10)),
			elog.FieldCustomKeyValue("userID", strconv.FormatInt(userID, 10)),
		)
		return nil, err
	}

	if err1 := r.setToCache(ctx, bizID, userID, res); err1 != nil {
		r.logger.Error("存储用户权限到缓存失败",
			elog.FieldErr(err1),
			elog.FieldCustomKeyValue("bizID", strconv.FormatInt(bizID, 10)),
			elog.FieldCustomKeyValue("userID", strconv.FormatInt(userID, 10)),
		)
	}
	return res, err
}

func (r *CachedRBACRepository) getFromCache(ctx context.Context, bizID, userID int64) ([]domain.UserPermission, error) {
	val := r.cache.Get(ctx, r.cacheKey(bizID, userID))
	var res []domain.UserPermission
	err := val.JSONScan(&res)
	return res, err
}

func (r *CachedRBACRepository) setToCache(ctx context.Context, bizID, userID int64, permissions []domain.UserPermission) error {
	const day = 24 * time.Hour
	const defaultExpiration = 36500 * day
	value, err := json.Marshal(permissions)
	if err != nil {
		return err
	}
	return r.cache.Set(ctx, r.cacheKey(bizID, userID), value, defaultExpiration)
}
