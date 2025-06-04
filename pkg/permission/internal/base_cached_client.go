package internal

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	permissionv1 "gitee.com/flycash/permission-platform/api/proto/gen/permission/v1"
)

const (
	keyPrefix = "client:userpermissions"
	bizIDKey  = "bizId"
	userIDKey = "userId"
)

type baseCachedClient struct{}

func (c *baseCachedClient) cacheKey(bizID, userID int64) string {
	return fmt.Sprintf("%s:%s:%d:%s:%d", keyPrefix, bizIDKey, bizID, userIDKey, userID)
}

// parseKey 解析缓存key，提取出bizID和userID
func (c *baseCachedClient) parseKey(key string) (bizID, userID int64, err error) {
	// 按冒号分割字符串
	parts := strings.Split(key, ":")

	// 检查格式是否正确：应该有6个部分
	const six = 6
	if len(parts) != six {
		return 0, 0, fmt.Errorf("invalid cache key format: expected 6 parts, got %d in key %s", len(parts), key)
	}

	// 重构后的前缀检查
	expectedPrefix := strings.Split(keyPrefix, ":")
	if len(expectedPrefix) != 2 || parts[0] != expectedPrefix[0] || parts[1] != expectedPrefix[1] {
		return 0, 0, fmt.Errorf("invalid cache key prefix: expected %s, got %s:%s in key %s", keyPrefix, parts[0], parts[1], key)
	}

	// 检查bizID和userID标识符
	if parts[2] != bizIDKey || parts[4] != userIDKey {
		return 0, 0, fmt.Errorf("invalid cache key format: expected %s:%%d:%s:%%d after prefix, got %s:%%s:%s:%%s in key %s",
			bizIDKey, userIDKey, parts[2], parts[4], key)
	}

	// 解析bizID
	bizID, err = strconv.ParseInt(parts[3], 10, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid bizID in cache key %s: %w", key, err)
	}

	// 解析userID
	userID, err = strconv.ParseInt(parts[5], 10, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid userID in cache key %s: %w", key, err)
	}

	return bizID, userID, nil
}

func (c *baseCachedClient) checkPermission(userPermission UserPermission, in *permissionv1.CheckPermissionRequest) (*permissionv1.CheckPermissionResponse, error) {
	bizID := in.GetPermission().GetBizId()
	resourceType := in.GetPermission().GetResourceType()
	resourceKey := in.GetPermission().GetResourceKey()
	actions := in.GetPermission().GetActions()
	allowedCounter := 0
	for i := range userPermission.Permissions {

		permission := userPermission.Permissions[i]

		if userPermission.BizID == bizID &&
			permission.Resource.Type == resourceType &&
			permission.Resource.Key == resourceKey &&
			slices.Contains(actions, permission.Action) {

			if permission.Effect == "deny" {
				return &permissionv1.CheckPermissionResponse{Allowed: false}, nil
			}

			allowedCounter++
		}
	}
	if len(actions) != allowedCounter {
		return nil, fmt.Errorf("%w, actions: %v", ErrUnknownPermissionAction, actions)
	}
	return &permissionv1.CheckPermissionResponse{Allowed: true}, nil
	// return &permissionv1.CheckPermissionResponse{Allowed: len(actions) == allowedCounter}, nil
}
