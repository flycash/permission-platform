package cache

import (
	"context"
	"encoding/json"
	"time"

	"gitee.com/flycash/permission-platform/internal/domain"
	"gitee.com/flycash/permission-platform/pkg/cache"
)

const (
	day               = 24 * time.Hour
	defaultExpiration = 36500 * day
)

type UserPermissionCache interface {
	// Get 获取某个业务下的用户的全部权限
	Get(ctx context.Context, bizID, userID int64) ([]domain.UserPermission, error)
	// Set 设置某个业务下的用户的全部权限，假定 permissions 中的bizID和UserID分别相同，即属于同一个业务下的同一个用户
	Set(ctx context.Context, permissions []domain.UserPermission) error
}

type userPermissionCache struct {
	c            cache.Cache
	cacheKeyFunc func(bizID, userID int64) string
}

func NewUserPermissionCache(c cache.Cache, cacheKeyFunc func(bizID, userID int64) string) UserPermissionCache {
	return &userPermissionCache{
		c:            c,
		cacheKeyFunc: cacheKeyFunc,
	}
}

func (r *userPermissionCache) Get(ctx context.Context, bizID, userID int64) ([]domain.UserPermission, error) {
	val := r.c.Get(ctx, r.cacheKeyFunc(bizID, userID))
	if val.Err != nil {
		return nil, val.Err
	}
	var res []domain.UserPermission
	err := val.JSONScan(&res)
	return res, err
}

func (r *userPermissionCache) Set(ctx context.Context, permissions []domain.UserPermission) error {
	if len(permissions) == 0 {
		return nil
	}
	value, err := json.Marshal(permissions)
	if err != nil {
		return err
	}
	const first = 0
	bizID, userID := permissions[first].BizID, permissions[first].UserID
	return r.c.Set(ctx, r.cacheKeyFunc(bizID, userID), value, defaultExpiration)
}
