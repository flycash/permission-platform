package permission

import (
	"fmt"
	"slices"

	permissionv1 "gitee.com/flycash/permission-platform/api/proto/gen/permission/v1"
)

type baseCachedClient struct{}

func (c *baseCachedClient) cacheKey(userID int64) string {
	return fmt.Sprintf("client:userpermissions:userId:%d", userID)
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
