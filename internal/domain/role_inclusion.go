package domain

// RoleInclusion 角色包含关系
type RoleInclusion struct {
	ID            int64
	BizID         int64
	IncludingRole Role // 包含者角色
	IncludedRole  Role // 被包含角色
	Ctime         int64
	Utime         int64
}
