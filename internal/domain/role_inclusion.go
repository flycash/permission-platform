package domain

// RoleInclusion 角色包含关系
type RoleInclusion struct {
	ID                int64    `json:"id"`
	BizID             int64    `json:"bizId"`
	IncludingRoleID   int64    `json:"includingRoleId"`   // 包含者角色ID
	IncludingRoleType RoleType `json:"includingRoleType"` // 包含者角色类型
	IncludingRoleName string   `json:"includingRoleName"` // 包含者角色名称
	IncludedRoleID    int64    `json:"includedRoleId"`    // 被包含角色ID
	IncludedRoleType  RoleType `json:"includedRoleType"`  // 被包含角色类型
	IncludedRoleName  string   `json:"includedRoleName"`  // 被包含角色名称
	IsIncluding       bool     `json:"isIncluding"`       // 是否为包含关系
	Ctime             int64    `json:"ctime"`
	Utime             int64    `json:"utime"`
}
