package domain

// RolePermission 角色权限关联
type RolePermission struct {
	ID             int64  `json:"id"`
	BizID          int64  `json:"bizId"`
	RoleID         int64  `json:"roleId"`
	PermissionID   int64  `json:"permissionId"`
	PermissionName string `json:"permissionName"`
	PermissionType string `json:"permissionType"` // 对应权限的资源类型
	StartTime      int64  `json:"startTime"`      // 临时角色有效期开始时间
	EndTime        int64  `json:"endTime"`        // 临时角色有效期结束时间
}
