package internal

type UserPermissionEvent struct {
	// uid => 全部权限
	Permissions map[int64]UserPermission `json:"permissions"`
}

type UserPermission struct {
	UserID      int64          `json:"userId"`
	BizID       int64          `json:"bizId"`
	Permissions []PermissionV1 `json:"permissions"`
}

type PermissionV1 struct {
	Resource Resource `json:"resource"`
	Action   string   `json:"action"`
	Effect   string   `json:"effect"`
}

type Resource struct {
	Key  string `json:"key"`
	Type string `json:"type"`
}
