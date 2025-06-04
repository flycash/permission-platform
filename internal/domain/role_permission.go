package domain

// RolePermission 角色权限关联
type RolePermission struct {
	ID         int64      `json:"id,omitzero"`
	BizID      int64      `json:"bizId,omitzero"`
	Role       Role       `json:"role,omitzero"`
	Permission Permission `json:"permission,omitzero"`
	Ctime      int64      `json:"ctime,omitzero"`
	Utime      int64      `json:"utime,omitzero"`
}
