package domain

// RoleType 角色类型
type RoleType string

const (
	RoleTypeSystem    RoleType = "system"
	RoleTypeCustom    RoleType = "custom"
	RoleTypeTemporary RoleType = "temporary"
)

// RoleMetadata 角色元数据
type RoleMetadata struct {
	IsAdmin    bool   `json:"isAdmin,omitempty"`    // 是否为管理员角色
	AdminLevel string `json:"adminLevel,omitempty"` // 管理级别: platform, business, service
}

type Role struct {
	ID          int64    `json:"id"`
	BizID       int64    `json:"bizId"`
	Type        RoleType `json:"type"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	StartTime   int64    `json:"startTime"`
	EndTime     int64    `json:"endTime"`
	Ctime       int64
	Utime       int64
}
