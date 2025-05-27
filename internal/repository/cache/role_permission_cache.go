package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"gitee.com/flycash/permission-platform/internal/domain"
	"github.com/ecodeclub/ecache"
	"github.com/ecodeclub/ekit/slice"
)

var ErrKeyNotFound = errors.New("key不存在")

type RolePermissionCache interface {
	// Get 获取某个业务下的用户的全部权限
	Get(ctx context.Context, bizID int64, roleIDs ...int64) ([]domain.RolePermission, error)
	// Add 添加角色权限关系到某个业务下，假定 permissions 中的bizID相同，即属于同一个业务下的角色权限
	Add(ctx context.Context, permissions []domain.RolePermission) error
	// Set 设置某个业务下的全部角色权限，假定 permissions 中的bizID相同，即属于同一个业务下的角色权限
	Set(ctx context.Context, permissions []domain.RolePermission) error
	// Del 删除某个业务下的具体角色权限
	Del(ctx context.Context, bizID, id int64) error
}

type rolePermissionCache struct {
	c ecache.Cache
}

func NewRolePermissionCache(c ecache.Cache) RolePermissionCache {
	return &rolePermissionCache{c: c}
}

func (r *rolePermissionCache) Get(ctx context.Context, bizID int64, roleIDs ...int64) ([]domain.RolePermission, error) {
	val := r.c.Get(ctx, r.cacheKey(bizID))
	if val.Err != nil {
		if val.KeyNotFound() {
			return nil, fmt.Errorf("%w", ErrKeyNotFound)
		}
		return nil, val.Err
	}
	var permissions []domain.RolePermission
	err := val.JSONScan(&permissions)
	if err == nil && len(roleIDs) > 0 {
		permissions = slice.FilterDelete(permissions, func(_ int, src domain.RolePermission) bool {
			return !slice.Contains(roleIDs, src.Role.ID)
		})
	}
	return permissions, err
}

func (r *rolePermissionCache) cacheKey(bizID int64) string {
	return fmt.Sprintf("rolePermissions:bizID:%d", bizID)
}

func (r *rolePermissionCache) Add(ctx context.Context, permissions []domain.RolePermission) error {
	if len(permissions) == 0 {
		return nil
	}
	// 取出全部
	const first = 0
	perms, err := r.Get(ctx, permissions[first].BizID)
	if err != nil && !errors.Is(err, ErrKeyNotFound) {
		return err
	}
	// 添加新权限并重置缓存，可以达成如果之前没有任何权限就Set
	return r.Set(ctx, append(perms, permissions...))
}

func (r *rolePermissionCache) Set(ctx context.Context, permissions []domain.RolePermission) error {
	if len(permissions) == 0 {
		return nil
	}
	value, err := json.Marshal(permissions)
	if err != nil {
		return err
	}
	const first = 0
	return r.c.Set(ctx, r.cacheKey(permissions[first].BizID), value, defaultExpiration)
}

func (r *rolePermissionCache) Del(ctx context.Context, bizID, id int64) error {
	perms, err := r.Get(ctx, bizID)
	if err != nil && !errors.Is(err, ErrKeyNotFound) {
		return err
	}
	perms = slice.FilterDelete(perms, func(_ int, src domain.RolePermission) bool {
		return src.ID == id
	})
	return r.Set(ctx, perms)
}
