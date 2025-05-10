package domain

// UserPermission 用户权限关联
type UserPermission struct {
	ID             int64      `json:"id"`
	BizID          int64      `json:"bizId"`
	UserID         int64      `json:"userId"`
	PermissionID   int64      `json:"permissionId"`
	PermissionName string     `json:"permissionName"`
	ResourceType   string     `json:"resourceType"`
	ResourceKey    string     `json:"resourceKey"`
	ResourceName   string     `json:"resourceName"`
	Action         ActionType `json:"action"`
	StartTime      int64      `json:"startTime"` // 临时权限生效时间
	EndTime        int64      `json:"endTime"`   // 临时权限失效时间
	Effect         string     `json:"effect"`    // allow, deny
	OnlyValid      bool       `json:"onlyValid"` // 是否只返回有效的权限
	Ctime          int64      `json:"ctime"`
	Utime          int64      `json:"utime"`
}
