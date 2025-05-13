package domain

import "encoding/json"

// RoleMetadata 角色元数据
type RoleMetadata struct{}

func (m RoleMetadata) String() string {
	marshal, err := json.Marshal(m)
	if err != nil {
		return ""
	}
	return string(marshal)
}

// Role 角色
type Role struct {
	ID          int64
	BizID       int64
	Type        string
	Name        string
	Description string
	Metadata    RoleMetadata
	Ctime       int64
	Utime       int64
}
