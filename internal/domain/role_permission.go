package domain

// RolePermission 角色权限关联
type RolePermission struct {
	ID         int64
	BizID      int64
	Role       Role
	Permission Permission
	Ctime      int64
	Utime      int64
}
